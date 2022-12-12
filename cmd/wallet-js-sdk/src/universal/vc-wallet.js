/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/**
 *  vcwallet module provides verifiable credential wallet SDK for aries universal wallet implementation.
 *
 *  @module vcwallet
 *
 */

/**
 * Supported content type from this wallet.
 * @see {@link https://github.com/hyperledger/aries-framework-go/blob/main/docs/vc_wallet.md#supported-data-models|Aries VC Wallet Data Models}
 * @enum {string}
 */
export const contentTypes = {
  COLLECTION: "collection",
  CREDENTIAL: "credential",
  DID_RESOLUTION_RESPONSE: "didResolutionResponse",
  METADATA: "metadata",
  CONNECTION: "connection",
  KEY: "key",
};

/**
 * UniversalWallet is universal wallet SDK built on top aries verifiable credential wallet controller (vcwallet).
 *
 * https://w3c-ccg.github.io/universal-wallet-interop-spec/
 *
 * Aries JS Controller: https://github.com/hyperledger/aries-framework-go/blob/main/docs/vc_wallet.md#javascript
 *
 * Refer Agent SDK Open API spec for detailed vcwallet request response models.
 *
 * @alias module:vcwallet
 */
export class UniversalWallet {
  /**
   *
   * @param agent - aries agent.
   * @param user -  unique wallet user identifier, the one used to create wallet profile.
   *
   */
  constructor({ agent, user } = {}) {
    this.agent = agent;
    this.user = user;
  }

  /**
   * Unlocks given wallet's key manager instance & content store and
   * returns a authorization token to be used for performing wallet operations.
   *
   *  @param {Object} options
   *  @param {String} options.localKMSPassphrase - (optional) passphrase for local kms for key operations.
   *  @param {Object} options.webKMSAuth - (optional) WebKMSAuth for authorizing access to web/remote kms.
   *  @param {String} options.webKMSAuth.authToken - (optional) Http header 'authorization' bearer token to be used, i.e access token.
   *  @param {String} options.webKMSAuth.capability - (optional) Capability if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.webKMSAuth.authzKeyStoreURL - (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.webKMSAuth.secretShare - (optional) secret share if ZCAP sign header feature to be used for authorizing access.
   *  @param {Object} options.edvUnlocks - (optional) for authorizing access to wallet's EDV content store.
   *  @param {String} options.edvUnlocks.authToken - (optional) Http header 'authorization' bearer token to be used, i.e access token.
   *  @param {String} options.edvUnlocks.capability - (optional) Capability if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.edvUnlocks.authzKeyStoreURL - (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.edvUnlocks.secretShare - (optional) secret share if ZCAP sign header feature to be used for authorizing access.
   *  @param {Time} options.expiry - (optional) time duration in milliseconds for which this profile will be unlocked.
   *
   * @returns {Promise<Object>} - 'object.token' - auth token subsequent use of wallet features.
   */
  async open({ localKMSPassphrase, webKMSAuth, edvUnlocks, expiry } = {}) {
    return await this.agent.vcwallet.open({
      userID: this.user,
      localKMSPassphrase,
      webKMSAuth,
      edvUnlocks,
      expiry,
    });
  }

  /**
   * Expires token issued to this VC wallet, removes wallet's key manager instance and closes wallet content store.
   *
   * @returns {Promise<Object>} - 'object.closed' -  bool flag false if token is not found or already expired for this wallet user.
   */
  async close() {
    return await this.agent.vcwallet.close({ userID: this.user });
  }

  /**
   * Adds given content to wallet content store.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {Object} request.contentType - type of the content to be added to the wallet, refer aries vc wallet for supported types.
   *  @param {String} request.content - content to be added wallet store.
   *  @param {String} request.collectionID - (optional) ID of the wallet collection to which the content should belong.
   *
   * @returns {Promise<Object>} - empty promise or an error if adding content to wallet store fails.
   */
  async add({ auth, contentType, content = {}, collectionID } = {}) {
    return await this.agent.vcwallet.add({
      userID: this.user,
      auth,
      contentType,
      collectionID,
      content,
    });
  }

  /**
   * remove given content from wallet content store.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {Object} request.contentType - type of the content to be removed from the wallet.
   *  @param {String} request.contentID - id of the content to be removed from wallet.
   *
   * @returns {Promise<Object>} - empty promise or an error if operation fails.
   */
  async remove({ auth = "", contentType = "", contentID = "" } = {}) {
    return await this.agent.vcwallet.remove({
      userID: this.user,
      auth,
      contentType,
      contentID,
    });
  }

  /**
   *  gets wallet content by ID from wallet content store.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {Object} request.contentType - type of the content to be removed from the wallet.
   *  @param {String} request.contentID - id of the content to be returned from wallet.
   *
   * @returns {Promise<Object>} - promise containing content or an error if operation fails.
   */
  async get({ auth = "", contentType = "", contentID = "" } = {}) {
    return await this.agent.vcwallet.get({
      userID: this.user,
      auth,
      contentType,
      contentID,
    });
  }

  /**
   *  gets all wallet contents from wallet content store for given type.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {Object} request.contentType - type of the contents to be returned from wallet.
   *  @param {String} request.collectionID - id of the collection on which the response contents to be filtered.
   *
   * @returns {Promise<Object>} - promise containing response contents or an error if operation fails.
   */
  async getAll({ auth, contentType, collectionID } = {}) {
    return await this.agent.vcwallet.getAll({
      userID: this.user,
      auth,
      contentType,
      collectionID,
    });
  }

  /**
   *  runs credential queries against wallet credential contents.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} query - credential query, refer: https://w3c-ccg.github.io/vp-request-spec/#format
   *
   * @returns {Promise<Object>} - promise of presentation(s) containing credential results or an error if operation fails.
   */
  async query(auth = "", query = []) {
    return await this.agent.vcwallet.query({ userID: this.user, auth, query });
  }

  /**
   *  runs credential queries against wallet credential contents.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} credential -  credential to be signed from wallet.
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
   * @returns {Promise<Object>} - promise of credential issued or an error if operation fails.
   */
  async issue(auth = "", credential = {}, proofOptions = {}) {
    return await this.agent.vcwallet.issue({
      userID: this.user,
      auth,
      credential,
      proofOptions,
    });
  }

  /**
   *  produces a Verifiable Presentation from wallet.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} credentialOptions - credential/presentations to verify..
   *  @param {Array<string>} credentialOptions.storedCredentials - (optional) ids of the credentials already saved in wallet content store.
   *  @param {Array<Object>} credentialOptions.rawCredentials - (optional) list of raw credentials to be presented.
   *  @param {Object} credentialOptions.presentation - (optional) presentation to be proved.
   *  @param {Object} proofOptions - proof options for signing.
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
  async prove(
    auth = "",
    { storedCredentials = [], rawCredentials = [], presentation = {} } = {},
    proofOptions = {}
  ) {
    return await this.agent.vcwallet.prove({
      userID: this.user,
      auth,
      storedCredentials,
      rawCredentials,
      presentation,
      proofOptions,
    });
  }

  /**
   *  verifies credential/presentation from wallet.
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
    auth = "",
    { storedCredentialID = "", rawCredential = {}, presentation = {} } = {}
  ) {
    return await this.agent.vcwallet.verify({
      userID: this.user,
      auth,
      storedCredentialID,
      rawCredential,
      presentation,
    });
  }

  /**
   *  derives a credential from wallet.
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
    deriveOption = {}
  ) {
    return await this.agent.vcwallet.derive({
      userID: this.user,
      auth,
      storedCredentialID,
      rawCredential,
      deriveOption,
    });
  }

  /**
   *  creates a key pair from wallet.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {String} request.keyType - type of the key to be created, refer aries kms for supported key types.
   *
   * @returns {Promise<Object>} - promise of derived credential or error if operation fails.
   */
  async createKeyPair(auth, { keyType } = {}) {
    return await this.agent.vcwallet.createKeyPair({
      userID: this.user,
      auth,
      keyType,
    });
  }

  /**
   *  accepts an out of band invitation and performs did-exchange.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} invitation - out of band invitation.
   *
   *  @param {Object} options - (optional) for accepting incoming out-of-band invitation and connecting to inviter.
   *  @param {String} options.myLabel - (optional) for providing label to be shared with the other agent during the subsequent did-exchange.
   *  @param {Array<string>} options.routerConnections - (optional) to provide router connection to be used.
   *  @param {String} options.reuseConnection - (optional) to provide DID to be used when reusing a connection.
   *  @param {Bool} options.reuseAnyConnection=false - (optional) to use any recognized DID in the services array for a reusable connection.
   *  @param {Time} options.timeout - (optional) to wait for connection status to be 'completed'.
   *
   * @returns {Promise<Object>} - promise of object containing connection ID or error if operation fails.
   */
  async connect(
    auth,
    invitation = {},
    {
      myLabel,
      routerConnections = [],
      reuseConnection,
      reuseAnyConnection = false,
      timeout,
    } = {}
  ) {
    return await this.agent.vcwallet.connect({
      userID: this.user,
      auth,
      invitation,
      myLabel,
      routerConnections,
      reuseConnection,
      reuseAnyConnection,
      timeout,
    });
  }

  /**
   *  accepts an out of band invitation, sends propose presentation message to inviter to initiate credential share interaction
   *  and waits for request presentation message from inviter as a response.
   *
   *  @see {@link https://w3c-ccg.github.io/universal-wallet-interop-spec/#proposepresentation|WACI Propose Presentation }
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} invitation - out of band invitation.
   *
   *  @param {Object} connectOptions - (optional) for accepting incoming out-of-band invitation and connecting to inviter.
   *  @param {String} connectOptions.myLabel - (optional) for providing label to be shared with the other agent during the subsequent did-exchange.
   *  @param {Array<string>} connectOptions.routerConnections - (optional) to provide router connection to be used.
   *  @param {String} connectOptions.reuseConnection - (optional) to provide DID to be used when reusing a connection.
   *  @param {Bool} connectOptions.reuseAnyConnection=false - (optional) to use any recognized DID in the services array for a reusable connection.
   *  @param {timeout} connectOptions.connectionTimeout - (optional) to wait for connection status to be 'completed'.
   *
   *  @param {Object} proposeOptions - (optional) for sending message proposing presentation.
   *  @param {String} proposeOptions.from - (optional) option from DID option to customize sender DID..
   *  @param {Time} proposeOptions.timeout - (optional) to wait for request presentation message from relying party.
   *
   * @returns {Promise<Object>} - promise of object containing presentation request message from relying party or error if operation fails.
   */
  async proposePresentation(
    auth,
    invitation = {},
    {
      myLabel,
      routerConnections,
      reuseConnection,
      reuseAnyConnection = false,
      connectionTimeout,
    },
    { from, timeout } = {}
  ) {
    return await this.agent.vcwallet.proposePresentation({
      userID: this.user,
      auth,
      invitation,
      from,
      timeout,
      connectOptions: {
        myLabel,
        routerConnections,
        reuseConnection,
        reuseAnyConnection,
        timeout: connectionTimeout,
      },
    });
  }

  /**
   *  sends present proof message from wallet to relying party as part of ongoing credential share interaction.
   *
   *  @see {@link https://w3c-ccg.github.io/universal-wallet-interop-spec/#presentproof|WACI Present Proof }
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {String} threadID - threadID of credential interaction.
   *  @param {Object} presentation - to be sent as part of present proof message..
   *
   *  @param {Object} options - (optional) for sending present proof message.
   *  @param {Bool} options.waitForDone - (optional) If true then wallet will wait for present proof protocol status to be done or abandoned .
   *  @param {Time} options.WaitForDoneTimeout - (optional) timeout to wait for present proof operation to be done.
   *
   * @returns {Promise<Object>} - promise of object containing present proof status & redirect info or error if operation fails.
   */
  async presentProof(
    auth,
    threadID,
    presentation,
    { waitForDone, WaitForDoneTimeout } = {}
  ) {
    return await this.agent.vcwallet.presentProof({
      userID: this.user,
      auth,
      threadID,
      presentation,
      waitForDone,
      WaitForDoneTimeout,
    });
  }

  /**
   *  accepts an out of band invitation, sends propose credential message to issuer to initiate credential issuance interaction
   *  and waits for offer credential message from inviter as a response.
   *
   *  @see {@link https://w3c-ccg.github.io/universal-wallet-interop-spec/#proposecredential|WACI Propose Credential }
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} invitation - out of band invitation.
   *
   *  @param {Object} connectOptions - (optional) for accepting incoming out-of-band invitation and connecting to inviter.
   *  @param {String} connectOptions.myLabel - (optional) for providing label to be shared with the other agent during the subsequent did-exchange.
   *  @param {Array<string>} connectOptions.routerConnections - (optional) to provide router connection to be used.
   *  @param {String} connectOptions.reuseConnection - (optional) to provide DID to be used when reusing a connection.
   *  @param {Bool} connectOptions.reuseAnyConnection=false - (optional) to use any recognized DID in the services array for a reusable connection.
   *  @param {timeout} connectOptions.connectionTimeout - (optional) to wait for connection status to be 'completed'.
   *
   *  @param {Object} proposeOptions - (optional) for sending message proposing credential.
   *  @param {String} proposeOptions.from - (optional) option from DID option to customize sender DID..
   *  @param {Time} proposeOptions.timeout - (optional) to wait for offer credential message from relying party.
   *
   * @returns {Promise<Object>} - promise of object containing offer credential message from relying party or error if operation fails.
   */
  async proposeCredential(
    auth,
    invitation = {},
    {
      myLabel,
      routerConnections,
      reuseConnection,
      reuseAnyConnection = false,
      connectionTimeout,
    },
    { from, timeout } = {}
  ) {
    return await this.agent.vcwallet.proposeCredential({
      userID: this.user,
      auth,
      invitation,
      from,
      timeout,
      connectOptions: {
        myLabel,
        routerConnections,
        reuseConnection,
        reuseAnyConnection,
        timeout: connectionTimeout,
      },
    });
  }

  /**
   *  sends request credential message from wallet to issuer as part of ongoing credential issuance interaction.
   *
   *  @see {@link https://w3c-ccg.github.io/universal-wallet-interop-spec/#requestcredential|WACI Request Credential }
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {String} threadID - threadID of credential interaction.
   *  @param {Object} presentation - to be sent as part of request credential message..
   *
   *  @param {Object} options - (optional) for sending request credential message.
   *  @param {Bool} options.waitForDone - (optional) If true then wallet will wait for credential response message or problem report .
   *  @param {Time} options.WaitForDoneTimeout - (optional) timeout to wait for credential response or problem report.
   *
   * @returns {Promise<Object>} - promise of object containing request credential status & redirect info or error if operation fails.
   */
  async requestCredential(
    auth,
    threadID,
    presentation,
    { waitForDone, WaitForDoneTimeout } = {}
  ) {
    return await this.agent.vcwallet.requestCredential({
      userID: this.user,
      auth,
      threadID,
      presentation,
      waitForDone,
      WaitForDoneTimeout,
    });
  }

  /**
   *  resolves given credential manifest by credential response or credential.
   * Supports: https://identity.foundation/credential-manifest/
   *
   *  @see {@link https://identity.foundation/credential-manifest|Credential Manifest }
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} manifest - credential manifest to be used for resolving credential.
   *
   *  @param {Object} options - credential or response to resolve.
   *
   *  @param {String} options.response - (optional) credential response to be resolved.
   *  if provided, then this option takes precedence over credential resolve option.
   *
   *  @param {Object} options.credential - (optional) raw credential to be resolved (accepting 'ldp_vc' format only).
   *  This option has to be provided with descriptor ID.
   *  @param {String} options.credentialID - (optional) ID of the credential to be resolved which is persisted in wallet content store.
   *  This option has to be provided with descriptor ID.
   *  @param {String} options.descriptorID - (optional) output descriptor ID of the descriptor from manifest to be used for resolving given
   *  credential or credentialID. This option is required only when a raw credential or credential ID is to be resolved.
   *
   * @returns {Promise<Object>} - promise of object containing request credential status & redirect info or error if operation fails.
   */
  async resolveCredential(
    auth,
    manifest,
    { response, credential, credentialID = "", descriptorID = "" } = {}
  ) {
    return await this.agent.vcwallet.resolveCredentialManifest({
      userID: this.user,
      auth,
      manifest,
      response,
      credential,
      credentialID,
      descriptorID,
    });
  }

  /**
   *  signs a JWT using a key in wallet.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {String} request.headers - JWT token headers.
   *  @param {String} request.claims - JWT token claims.
   *  @param {String} request.kid - wallet's key id.
   *
   * @returns {Promise<Object>} - promise of object containing signed JWT string.
   */
  async signJWT({ auth, headers, claims, kid } = {}) {
    return await this.agent.vcwallet.signJWT({
      userID: this.user,
      auth,
      headers,
      claims,
      kid,
    });
  }

  /**
   *  verifies a JWT using wallet.
   *
   *  @param {Object} request
   *  @param {String} request.auth -  authorization token for performing this wallet operation.
   *  @param {String} request.jwt - JWT token to be verified.
   *
   * @returns {Promise<Object>} - promise of object containing a boolean representing verification result or an error if operation fails.
   */
     async verifyJWT({ auth, jwt } = {}) {
      return await this.agent.vcwallet.verifyJWT({
        userID: this.user,
        auth,
        jwt,
      });
    }
}

/**
 *  creates new wallet profile for given user.
 *
 *  @param {Object} agent - aries agent
 *  @param {String} userID - unique identifier of user for which the profile is being created.
 *  @param {String} profileOptions -  options for creating profile.
 *  @param {String} profileOptions.localKMSPassphrase - (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations.
 *  @param {String} profileOptions.keyStoreURL - (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations.
 *  @param {Object} profileOptions.edvConfiguration - (optional) EDV configuration if profile wants to use EDV as a wallet content store.
 *  By Default, aries context storage provider will be used.
 *
 *  @param {String} profileOptions.edvConfiguration.serverURL - EDV server URL for storing wallet contents.
 *  @param {String} profileOptions.edvConfiguration.vaultID - EDV vault ID for storing the wallet contents.
 *  @param {String} profileOptions.edvConfiguration.encryptionKID - Encryption key ID of already existing key in wallet profile kms.
 *  If profile is using localkms then wallet will create this key set for wallet user.
 *  @param {String} profileOptions.edvConfiguration.macKID -  MAC operation key ID of already existing key in wallet profile kms.
 *  If profile is using localkms then wallet will create this key set for wallet user.
 *
 *
 * @returns {Promise<Object>} - empty promise or error if operation fails.
 */
export async function createWalletProfile(
  agent,
  userID,
  { localKMSPassphrase, keyStoreURL, edvConfiguration } = {}
) {
  return await agent.vcwallet.createProfile({
    userID,
    localKMSPassphrase,
    keyStoreURL,
    edvConfiguration,
  });
}

/**
 *  updates existing wallet profile for given user.
 *  Caution:
 *  - you might lose your existing keys if you change kms options.
 *  - you might lose your existing wallet contents if you change storage/EDV options
 *  (ex: switching context storage provider or changing EDV settings).
 *
 *  @param {Object} agent - aries agent
 *  @param {String} userID - unique identifier of user for which the profile is being created.
 *  @param {String} profileOptions -  options for creating profile.
 *  @param {String} profileOptions.localKMSPassphrase - (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations.
 *  @param {String} profileOptions.keyStoreURL - (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations.
 *  @param {String} profileOptions.edvConfiguration - (optional) EDV configuration if profile wants to use EDV as a wallet content store.
 *  By Default, aries context storage provider will be used.
 *
 * @returns {Promise<Object>} - empty promise or error if operation fails.
 */
export async function updateWalletProfile(
  agent,
  userID,
  { localKMSPassphrase, keyStoreURL, edvConfiguration } = {}
) {
  return await agent.vcwallet.updateProfile({
    userID,
    localKMSPassphrase,
    keyStoreURL,
    edvConfiguration,
  });
}

/**
 *  check is profile exists for given wallet user.
 *
 *  @param {Object} agent - aries agent
 *  @param {String} userID - unique identifier of user for which the profile is being created.
 *  @param {String} profilestorage provider will be used.
 *
 * @returns {Promise<Object>} - empty promise or error if profile not found.
 */
export async function profileExists(agent, userID) {
  return await agent.vcwallet.profileExists({ userID });
}
