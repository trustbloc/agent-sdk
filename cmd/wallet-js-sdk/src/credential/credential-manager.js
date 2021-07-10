/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {contentTypes, UniversalWallet} from "..";

const JSONLD_CTX_MANIFEST_MAPPING = ['https://w3id.org/wallet/v1', 'https://trustbloc.github.io/context/wallet/manifest-mapping-v1.jsonld']
const MANIFEST_MAPPING_METADATA_TYPE = 'ManifestMapping'

/**
 *  CredentialManager provides wallet credential features,
 *
 *  - Save, get, remove & get all.
 *  - Issue, prove, verify & derive.
 *  - query (QueryByExample, QueryByFrame, PresentationExchange & DIDAuth)
 *
 */
export class CredentialManager {
    /**
     *
     * @class CredentialManager
     *
     * @param {string} agent - aries agent.
     * @param {string} user -  unique wallet user identifier, the one used to create wallet profile.
     *
     */
    constructor({agent, user} = {}) {
        this.agent = agent
        this.wallet = new UniversalWallet({agent: this.agent, user})
    }

    /**
     * Saves given credential into wallet content store.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *  @param {Object} contents - credential(s) to be saved in wallet content store.
     *  @param {Object} contents.credential - credential to be saved in wallet content store.
     *  @param {Array<Object>} contents.credentials - array of credentials to be saved in wallet content store.
     *  @param {Object} contents.presentation - presentation from which all the credentials to be saved in wallet content store.
     *  @param {Object} options - options for saving credential.
     *  @param {boolean} options.verify - (optional) to verify credential before save.
     *  @param {String} options.collection - (optional) ID of the wallet collection to which the credential should belong.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async save(auth, {credential = null, credentials = [], presentation = {verifiableCredential: []}} = {}, {verify = false, collection = ''} = {}) {
        let contents = (credential ? [credential] : []).concat(credentials, presentation.verifiableCredential)

        // verify all credentials
        if (verify) {
            const _doVerify = async (rawCredential) => {
                let {verified, error} = await this.wallet.verify(auth, {rawCredential})

                if (!verified) {
                    console.error(`verification failed for ${rawCredential.id}`)
                    throw `credential verification failed`
                }
            };

            await Promise.all(contents.map(_doVerify))
        }

        const _save = async (credential) => {
            await this.wallet.add({
                auth,
                contentType: contentTypes.CREDENTIAL,
                content: credential,
                collectionID: collection
            })
        };

        await Promise.all(contents.map(_save))
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
        return await this.wallet.get({auth, contentType: contentTypes.CREDENTIAL, contentID})
    }

    /**
     * Gets All credentials from wallet.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *
     *  @returns {Promise<Object>} result.contents - promise containing results or error if operation fails.
     */
    async getAll(auth) {
        return await this.wallet.getAll({auth, contentType: contentTypes.CREDENTIAL})
    }

    /**
     * Removes credential from wallet
     *
     *  @param {string} auth - authorization token for wallet operations.
     *  @param {string} contentID - ID of the credential to be removed from wallet content store.
     *
     *  @returns {Promise<Object>} - empty promise or an error if operation fails.
     */
    async remove(auth, contentID) {
        return await this.wallet.remove({auth, contentType: contentTypes.CREDENTIAL, contentID})
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
     *  @param {string} proofOptions.proofRepresentation - (optional) type of proof data expected ( "proofValue" or "jws").
     *  By default, 'proofValue' will be used.
     *
     *  @returns {Promise<Object>} - promise containing issued credential or an error if operation fails.
     */
    async issue(auth, credential, {controller, verificationMethod, created, domain, challenge, proofType, proofRepresentation} = {}) {
        return await this.wallet.issue(auth, credential, {
            controller,
            verificationMethod,
            created,
            domain,
            challenge,
            proofType,
            proofRepresentation
        })
    }

    /**
     * Prepares verifiable presentation of given credential(s).
     *
     *  @param {String} auth -  authorization token for performing this wallet operation.
     *  @param {Object} credentialOptions - credential/presentations to verify..
     *  @param {Array<string>} credentialOptions.storedCredentials - (optional) ids of the credentials already saved in wallet content store.
     *  @param {Array<Object>} credentialOptions.rawCredentials - (optional) list of raw credentials to be presented.
     *  @param {Object} credentialOptions.presentation - (optional) presentation to be proved.
     *  @param {Object} proofOptions - proof options for issuing credential.
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
     *  @param {String} proofOptions.proofRepresentation - (optional) type of proof data expected ( "proofValue" or "jws").
     *  By default, 'proofValue' will be used.
     *
     * @returns {Promise<Object>} - promise of signed presentation or an error if operation fails.
     */
    async present(auth, {storedCredentials = [], rawCredentials = [], presentation = {}},
                  {controller, verificationMethod, created, domain, challenge, proofType, proofRepresentation} = {}) {
        return await this.wallet.prove(auth,
            {
                storedCredentials,
                rawCredentials,
                presentation
            },
            {
                controller,
                verificationMethod,
                created,
                domain,
                challenge,
                proofType,
                proofRepresentation
            })
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
    async verify(auth, {storedCredentialID = '', rawCredential = {}, presentation = {}}) {
        return await this.wallet.verify(auth, {storedCredentialID, rawCredential, presentation})
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
    async derive(auth = '', {storedCredentialID = '', rawCredential = {}}, {frame, nonce} = {}) {
        return await this.wallet.derive(auth, {
            storedCredentialID,
            rawCredential
        }, {frame, nonce})
    }

    /**
     *  runs credential queries in wallet.
     *
     *  @param {String} auth -  authorization token for performing this wallet operation.
     *  @param {Object} query - list of credential queries, any types of supported query types can be mixed.
     *
     * @returns {Promise<Object>} - promise of presentation(s) containing credential results or an error if operation fails.
     */
    async query(auth = '', query = []) {
        return await this.wallet.query(auth, query)
    }

    /**
     *  saves manifest credential along with its mapping to given connection ID.
     *
     *  @param {String} auth -  authorization token for performing this wallet operation.
     *  @param {Object} manifest - manifest credential (can be of any type).
     *  @param {String} connectionID - connection ID to which manifest credential to be mapped.
     *
     * @returns {Promise<Object>} - empty promise or an error if operation fails.
     */
    async saveManifestCredential(auth = '', manifest = {}, connectionID = '') {
        if (!manifest.id) {
            throw "invalid manifest credential, credential id is required."
        }

        await this.save(auth, {credential: manifest})

        let content = {
            "@context": JSONLD_CTX_MANIFEST_MAPPING,
            id: manifest.id,
            type: MANIFEST_MAPPING_METADATA_TYPE,
            connectionID
        }

        await this.wallet.add({auth, contentType: contentTypes.METADATA, content})
    }

    /**
     *  Returns connection ID mapped to given manifest credential ID.
     *
     *  @param {String} auth -  authorization token for performing this wallet operation.
     *  @param {String} manifestCredID - ID of manifest credential.
     *
     * @returns {Promise<String>} - promise containing connection ID or an error if operation fails.
     */
    async getManifestConnection(auth = '', manifestCredID=''){
        let {content} = await this.wallet.get({auth, contentType: contentTypes.METADATA, contentID:manifestCredID})

        return content.connectionID
    }

    /**
     *  Gets all manifest credentials saved in wallet.
     *
     *  @param {String} auth -  authorization token for performing this wallet operation.
     *
     * @returns {Promise<Object>} - promise containing manifest credential search results or an error if operation fails.
     */
    async getAllManifests(auth = ''){
       return await this.query(auth, [{
           type: "QueryByExample",
           credentialQuery: [{
               example: {
                   "@context": ["https://www.w3.org/2018/credentials/v1"],
                   type: ["IssuerManifestCredential"]
               }
           }]
       }])
    }
}
