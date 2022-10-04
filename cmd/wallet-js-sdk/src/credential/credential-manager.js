/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { contentTypes, UniversalWallet } from "..";
import jp from "jsonpath";

const JSONLD_CTX_BLINDED_ROUTING_MANIFEST_MAPPING = [
  "https://w3id.org/wallet/v1",
  "https://trustbloc.github.io/context/wallet/manifest-mapping-v1.jsonld",
];

const JSONLD_CREDENTIAL_METADATA_MODEL = [
  "https://w3id.org/wallet/v1",
  "https://www.w3.org/2018/credentials/v1",
  "https://trustbloc.github.io/context/wallet/credential-metadata-v1.jsonld",
];

const MANIFEST_MAPPING_METADATA_TYPE = "ManifestMapping";
const CREDENTIAL_METADATA_MODEL_TYPE = "CredentialMetadata";
const SUPPORTED_VC_FORMAT = ["ldp_vc", "jwt_vc"];

/**
 *  credential module provides wallet credential handling features,
 *
 *  @module credential
 *
 */

/**
 *  - Save, get, remove & get all.
 *  - Issue, prove, verify & derive.
 *  - query (QueryByExample, QueryByFrame, PresentationExchange & DIDAuth)
 *
 * @alias module:credential
 */
export class CredentialManager {
  /**
   *
   * @param {string} agent - aries agent.
   * @param {string} user -  unique wallet user identifier, the one used to create wallet profile.
   *
   */
  constructor({ agent, user } = {}) {
    this.agent = agent;
    this.wallet = new UniversalWallet({ agent: this.agent, user });
  }

  /**
   * Saves given credential into wallet content store along with credential metadata & manifest details along with saved credential.
   *
   *  @param {string} auth - authorization token for wallet operations.
   *  @param {Object} contents - credential(s) to be saved in wallet content store.
   *  @param {Array<Object>} contents.credentials - array of credentials to be saved in wallet content store.
   *  @param {Object} contents.presentation - presentation from which all the credentials to be saved in wallet content store.
   *  If credential response presentation is provided then no need to supply descriptor map along with manifest.
   *  Refer @see {@link https://identity.foundation/credential-manifest/#credential-response|Credential Response} for more details.
   *  @param {Object} options - options for saving credential.
   *  @param {boolean} options.verify - (optional) to verify credential before save.
   *  @param {String} options.collection - (optional) ID of the wallet collection to which the credential should belong.
   *  @param {String} options.manifest - (required) credential manifest of the credential being saved.
   *  Refer @see {@link https://identity.foundation/credential-manifest/#credential-manifest-2|Credential Manifest} for more details.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async save(
    auth,
    { credentials = [], presentation = { verifiableCredential: [] } } = {},
    { verify = false, collection = "", manifest, descriptorMap } = {}
  ) {
    if (!manifest) {
      throw "credential manifest information is required";
    }

    // credential array takes precedence over presentation.
    let contents =
      credentials.length > 0 ? credentials : presentation.verifiableCredential;

    // verify all credentials if required.
    if (verify) {
      const _doVerify = async (rawCredential) => {
        let { verified, error } = await this.wallet.verify(auth, {
          rawCredential,
        });

        if (!verified) {
          // TODO: error message won't have an ID when rawCredential is a raw JWT VC
          console.error(`verification failed for ${rawCredential.id}`);
          throw `credential verification failed`;
        }
      };

      await Promise.all(contents.map(_doVerify));
    }

    // prepare save credential
    const _saveCredential = async (credential) => {
      await this.wallet.add({
        auth,
        contentType: contentTypes.CREDENTIAL,
        content: credential,
        collectionID: collection,
      });
    };

    // prepare save metadata
    const _saveMetadata = async (descriptor) => {
      const { id, format, path } = descriptor;
      if (!SUPPORTED_VC_FORMAT.includes(format)) {
        console.warn(
          `unsupported credential format '${format}', supporting only '${SUPPORTED_VC_FORMAT}' for now`
        );

        return;
      }

      const credentialMatch = jp.query(
        credentials.length > 0 ? credentials : presentation,
        path
      );

      if (credentialMatch.length != 1) {
        throw "credential match from descriptor should resolve to a single credential";
      }

      await this.saveCredentialMetadata(auth, {
        credential: credentialMatch[0],
        descriptorID: id,
        manifest,
        collection,
      });
    };

    // validate response and collection descriptor map.
    let descriptors;
    if (descriptorMap) {
      descriptors = descriptorMap;
    } else if (presentation["credential_response"]) {
      // validate response against manifest provided
      if (presentation["credential_response"].manifest_id != manifest.id) {
        throw "credential response not matching with manifest provided";
      }

      descriptors = presentation["credential_response"].descriptor_map;
    } else {
      throw "descriptor map is required to save mapping between credential being saved and manifest";
    }

    // save all credentials & metadata first, to catch any duplicate entry issues before saving manifests.
    await Promise.all(
      contents.map(_saveCredential).concat(descriptors.map(_saveMetadata))
    );
  }

  /**
   * Reads credential metadata and saves credential metadata data model into wallet content store.
   *
   *  @param {string} auth - authorization token for wallet operations.
   *  @param {Object} options - subjects from which credential metadata will be extracted.
   *  @param {Object} options.credential - credential data model from which basic credential attributes like type, issuer, expiration etc will be read.
   *  @param {String} options.manifestID - ID of the credential manifest of the given credential
   *  @param {String} options.descriptorID - ID of the credential manifest output descriptor of the given credential.
   *  @param {String} options.collection - (optional) ID of the collection to which this credential belongs.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async saveCredentialMetadata(
    auth,
    { credential, manifest, descriptorID, collection }
  ) {
    const {
      id,
      type,
      issuer,
      name,
      description,
      issuanceDate,
      expirationDate,
    } = credential;

    const resolved = await this.resolveManifest(auth, {
      credential,
      descriptorID,
      manifest,
    });

    await this.wallet.add({
      auth,
      contentType: contentTypes.METADATA,
      collectionID: collection,
      content: {
        "@context": JSONLD_CREDENTIAL_METADATA_MODEL,
        context: credential["@context"],
        type: CREDENTIAL_METADATA_MODEL_TYPE,
        credentialType: type,
        id,
        issuer,
        name,
        description,
        issuanceDate,
        expirationDate,
        resolved,
        collection,
        issuerStyle: manifest.issuer,
      },
    });
  }

  /**
   * Gets credential metadata from wallet content store and also optionally resolves credential data using credential manifest.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {String} id - credential ID.
   *
   *  @returns {Promise<Object>} - promise containing credential metadata or error if operation fails.
   */
  async getCredentialMetadata(auth, id) {
    let { content } = await this.wallet.get({
      auth,
      contentType: contentTypes.METADATA,
      contentID: id,
    });

    return content;
  }

  /**
   * Gets all credential metadata models from wallet content store.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {Object} options - options to get all credential metadata.
   *  @param {Bool} options.credentialIDs - (optional) filters credential metadata by given credential IDs.
   *  @param {String} options.collection - (optional) filters credential metadata by given collection ID.
   *
   *  @returns {Promise<Object>} - promise containing list of credential metadata or error if operation fails.
   */
  async getAllCredentialMetadata(
    auth,
    { credentialIDs = [], collection = "" } = {}
  ) {
    const { contents } = await this.wallet.getAll({
      auth,
      contentType: contentTypes.METADATA,
      collectionID: collection,
    });

    const _filterCred = (id) => {
      if (credentialIDs.length == 0) {
        return true;
      }

      return credentialIDs.includes(id);
    };

    const metadataList = Object.values(contents).filter(
      (metadata) =>
        metadata.type == CREDENTIAL_METADATA_MODEL_TYPE &&
        _filterCred(metadata.id)
    );

    return metadataList;
  }

  /**
   * Updates credential metadata. Currently supporting updating only credential name and description fields.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {String} id - ID of the credential metadata to be updated.
   *  @param {Object} options - options to update credential metadata.
   *  @param {String} options.name - (optional) name attribute of the credential metadata to be updated.
   *  @param {String} options.description - (optional) description attribute of the credential metadata to be updated.
   *  @param {String} options.collection - (optional) ID  of the collection to which this credential metadata to be updated.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async updateCredentialMetadata(auth, id, { name, description, collection }) {
    let { content } = await this.wallet.get({
      auth,
      contentType: contentTypes.METADATA,
      contentID: id,
    });

    if (name) {
      content.name = name;
    }

    if (description) {
      content.description = description;
    }

    // remove
    await this.wallet.remove({
      auth,
      contentType: contentTypes.METADATA,
      contentID: id,
    });

    // add again
    await this.wallet.add({
      auth,
      contentType: contentTypes.METADATA,
      collectionID: collection,
      content,
    });
  }

  /**
   * Resolves credential by credential manifest, descriptor or response.
   *
   * Given credential can be resolved by raw credential, ID of the credential saved in wallet, credential response,
   * ID of the manifest saved in wallet, raw credential manifest, output descriptor of the manifest etc
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {Object} options - options to resolve credential from wallet.
   *  @param {String} options.credentialID - (optional) ID of the credential to be resolved from wallet content store.
   *  @param {String} options.credential - (optional) raw credential data model to be resolved.
   *  @param {String} options.response - (optional) credential response using which given raw credential or credential ID to be resolved.
   *  @param {String} options.manifestID - (optional) ID of the manifest from wallet content store.
   *  @param {String} options.manifest - (optional) raw manifest to be used for resolving credential.
   *  @param {String} options.descriptorID - (optional) if response not provided then this descriptor ID can be used to resolve credential.
   *
   * Refer @see {@link https://identity.foundation/credential-manifest/|Credential Manifest Specifications} for more details.
   *
   *  @returns {Promise<Object>} - promise containing resolved results or error if operation fails.
   */
  async resolveManifest(
    auth,
    {
      credentialID,
      credential,
      response,
      manifestID,
      manifest,
      descriptorID,
    }
  ) {
    if (!manifest) {
      const { content } = await this.wallet.get({
        auth,
        contentType: contentTypes.METADATA,
        contentID: manifestID,
      });

      manifest = content;
    }

    let result = await this.wallet.resolveCredential(auth, manifest, {
      credentialID,
      descriptorID,
      response,
      credential,
    });

    return result.resolved;
  }

  /**
   * Gets credential from wallet
   *
   *  @param {string} auth - authorization token for wallet operations.
   *  @param {string} contentID - ID of the credential to be read from wallet content store.
   *
   *  @returns {Promise<Object>} result.content -- promise containing credential or error if operation fails.
   */
  async get(auth, contentID) {
    return await this.wallet.get({
      auth,
      contentType: contentTypes.CREDENTIAL,
      contentID,
    });
  }

  /**
   * Gets All credentials from wallet.
   *
   *  @param {string} auth - authorization token for wallet operations.
   *
   *  @returns {Promise<Object>} result.contents - promise containing results or error if operation fails.
   */
  async getAll(auth, { collectionID } = {}) {
    return await this.wallet.getAll({
      auth,
      contentType: contentTypes.CREDENTIAL,
      collectionID,
    });
  }

  /**
   * Removes credential and its metadata from wallet.
   *
   *  Doesn't delete respective credential manifest since one credential manifest can be referred by many other credentials too.
   *
   *  @param {string} auth - authorization token for wallet operations.
   *  @param {string} contentID - ID of the credential to be removed from wallet content store.
   *
   *  @returns {Promise<Object>} - empty promise or an error if operation fails.
   */
  async remove(auth, contentID) {
    return await Promise.all([
      this.wallet.remove({
        auth,
        contentType: contentTypes.CREDENTIAL,
        contentID,
      }),
      this.wallet.remove({
        auth,
        contentType: contentTypes.METADATA,
        contentID,
      }),
    ]);
  }

  /**
   * Issues credential from wallet
   *
   *  @param {string} auth - authorization token for wallet operations.
   *  @param {Object} credential - credential to be signed from wallet.
   *  @param {Object} proofOptions - credential to be signed from wallet.
   *  @param {string} proofOptions.controller -  DID to be used for signing.
   *  @param {string} proofOptions.verificationMethod - (optional) VerificationMethod is the URI of the verificationMethod used for the proof.
   *  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions.
   *  @param {string} proofOptions.created - (optional) Created date of the proof.
   *  By default, current system time will be used.
   *  @param {string} proofOptions.domain - (optional) operational domain of a digital proof.
   *  By default, domain will not be part of proof.
   *  @param {string} proofOptions.challenge - (optional) random or pseudo-random value option authentication.
   *  By default, challenge will not be part of proof.
   *  @param {string} proofOptions.proofType - (optional) signature type used for signing.
   *  By default, proof will be generated in Ed25519Signature2018 format.
   *  @param {String} proofOptions.proofFormat - (optional) representational format for the credential.
   *  Valid values are "ExternalJWTProofFormat" and "EmbeddedLDProofFormat".
   *  By default, credential will be JSON-LD with embedded proof.
   *  @param {string} proofOptions.proofRepresentation - (optional) type of proof data expected ( "proofValue" or "jws").
   *  By default, 'proofValue' will be used.
   *
   *  @returns {Promise<Object>} - promise containing issued credential or an error if operation fails.
   */
  async issue(
    auth,
    credential,
    {
      controller,
      verificationMethod,
      created,
      domain,
      challenge,
      proofType,
      proofFormat,
      proofRepresentation,
    } = {}
  ) {
    return await this.wallet.issue(auth, credential, {
      controller,
      verificationMethod,
      created,
      domain,
      challenge,
      proofType,
      proofFormat,
      proofRepresentation,
    });
  }

  /**
   * Prepares verifiable presentation of given credential(s).
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} credentialOptions - credential/presentations to verify..
   *  @param {Array<string>} credentialOptions.storedCredentials - (optional) ids of the credentials already saved in wallet content store.
   *  @param {Array<Object>} credentialOptions.rawCredentials - (optional) list of raw credentials to be presented.
   *  @param {Object} credentialOptions.presentation - (optional) presentation to be proved.
   *  @param {Object} proofOptions - proof options for signing presentation.
   *  @param {String} proofOptions.controller -  DID to be used for signing.
   *  @param {String} proofOptions.verificationMethod - (optional) VerificationMethod is the URI of the verificationMethod used for the proof.
   *  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions.
   *  @param {String} proofOptions.created - (optional) Created date of the proof.
   *  By default, current system time will be used.
   *  @param {String} proofOptions.domain - (optional) operational domain of a digital proof.
   *  By default, domain will not be part of proof.
   *  @param {String} proofOptions.challenge - (optional) random or pseudo-random value option authentication.
   *  By default, challenge will not be part of proof.
   *  @param {String} proofOptions.proofType - (optional) signature type used for signing.
   *  By default, proof will be generated in Ed25519Signature2018 format.
   *  @param {String} proofOptions.proofFormat - (optional) representational format for the presentation.
   *  Valid values are "ExternalJWTProofFormat" and "EmbeddedLDProofFormat".
   *  By default, presentation will be JSON-LD with embedded proof.
   *  @param {String} proofOptions.proofRepresentation - (optional) type of proof data expected ( "proofValue" or "jws").
   *  By default, 'proofValue' will be used.
   *
   * @returns {Promise<Object>} - promise of signed presentation or an error if operation fails.
   */
  async present(
    auth,
    { storedCredentials = [], rawCredentials = [], presentation = {} },
    {
      controller,
      verificationMethod,
      created,
      domain,
      challenge,
      proofType,
      proofFormat,
      proofRepresentation,
    } = {}
  ) {
    return await this.wallet.prove(
      auth,
      {
        storedCredentials,
        rawCredentials,
        presentation,
      },
      {
        controller,
        verificationMethod,
        created,
        domain,
        challenge,
        proofType,
        proofFormat,
        proofRepresentation,
      }
    );
  }

  /**
   *  Verifies credential/presentation from wallet.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {String} verificationOption - credential/presentation to be verified.
   *  @param {String} verificationOption.storedCredentialID - (optional) id of the credential already saved in wallet content store.
   *  @param {Object} verificationOption.rawCredential - (optional) credential to be verified.
   *  @param {Object} verificationOption.presentation - (optional) presentation to be verified.
   *
   * @returns {Promise<Object>} - promise of verification result(bool) and error containing cause if verification fails.
   */
  async verify(
    auth,
    { storedCredentialID = "", rawCredential = {}, presentation = {} }
  ) {
    return await this.wallet.verify(auth, {
      storedCredentialID,
      rawCredential,
      presentation,
    });
  }

  /**
   *  Derives a credential from wallet.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *
   *  @param {String} credentialOption - credential to be dervied.
   *  @param {String} credentialOption.storedCredentialID - (optional) id of the credential already saved in wallet content store.
   *  @param {Object} credentialOption.rawCredential - (optional) credential to be derived.
   *
   *  @param {Object} deriveOption - derive options.
   *  @param {Object} deriveOption.frame -  JSON-LD frame used for derivation.
   *  @param {String} deriveOption.nonce - (optional) to prove uniqueness or freshness of the proof..
   *
   * @returns {Promise<Object>} - promise of derived credential or error if operation fails.
   */
  async derive(
    auth = "",
    { storedCredentialID = "", rawCredential = {} },
    { frame, nonce } = {}
  ) {
    return await this.wallet.derive(
      auth,
      {
        storedCredentialID,
        rawCredential,
      },
      { frame, nonce }
    );
  }

  /**
   *  runs credential queries in wallet.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} query - list of credential queries, any types of supported query types can be mixed.
   *
   * @returns {Promise<Object>} - promise of presentation(s) containing credential results or an error if operation fails.
   */
  async query(auth = "", query = []) {
    return await this.wallet.query(auth, query);
  }

  /**
   *  saves manifest credential along with its mapping to given connection ID.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} manifest - manifest credential (can be of any type).
   *  @param {String} connectionID - connection ID to which manifest credential to be mapped.
   *
   *  @deprecated, to be used for DIDComm blinded routing flow only
   *
   * @returns {Promise<Object>} - empty promise or an error if operation fails.
   */
  async saveManifestVC(auth = "", manifest = {}, connectionID = "") {
    if (!manifest.id) {
      throw "invalid manifest credential, credential id is required.";
    }

    await this.wallet.add({
      auth,
      contentType: contentTypes.CREDENTIAL,
      content: manifest,
    });

    let content = {
      "@context": JSONLD_CTX_BLINDED_ROUTING_MANIFEST_MAPPING,
      id: manifest.id,
      type: MANIFEST_MAPPING_METADATA_TYPE,
      connectionID,
    };

    await this.wallet.add({
      auth,
      contentType: contentTypes.METADATA,
      content,
    });
  }

  /**
   *  Returns connection ID mapped to given manifest credential ID.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {String} manifestCredID - ID of manifest credential.
   *
   * @returns {Promise<String>} - promise containing connection ID or an error if operation fails.
   */
  async getManifestConnection(auth = "", manifestCredID = "") {
    let { content } = await this.wallet.get({
      auth,
      contentType: contentTypes.METADATA,
      contentID: manifestCredID,
    });

    return content.connectionID;
  }

  /**
   *  Gets all manifest credentials saved in wallet.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *
   * @returns {Promise<Object>} - promise containing manifest credential search results or an error if operation fails.
   */
  async getAllManifestVCs(auth = "") {
    return await this.query(auth, [
      {
        type: "QueryByExample",
        credentialQuery: [
          {
            example: {
              "@context": ["https://www.w3.org/2018/credentials/v1"],
              type: ["IssuerManifestCredential"],
            },
          },
        ],
      },
    ]);
  }
}
