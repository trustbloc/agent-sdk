/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Supported content type from this UniversalWallet.
export const contentTypes = {
    COLLECTION: "collection",
    CREDENTIAL: "credential",
    DID_RESOLUTION_RESPONSE: "didResolutionResponse",
    METADATA: "metadata",
    CONNECTION: "connection",
    KEY: "key",
}

/**
 * UniversalWallet is universal wallet SDK built on top aries universal wallet controller (vcwallet).
 *
 * https://w3c-ccg.github.io/universal-wallet-interop-spec/
 *
 * Refer Agent SDK Open API spec for detailed vcwallet request response models.
 *
 */
export class UniversalWallet {
    /**
     *
     * @class UniversalWallet
     *
     * @param agent - aries agent.
     * @param user -  unique wallet user identifier, the one used to create wallet profile.
     *
     */
    constructor({agent, user} = {}) {
        this.agent = agent
        this.user = user
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
    async open({localKMSPassphrase, webKMSAuth, edvUnlocks, expiry} = {}) {
        console.log('@got expiry', expiry)
        return await this.agent.vcwallet.open({userID: this.user, localKMSPassphrase, webKMSAuth, edvUnlocks, expiry})
    }

    /**
     * Expires token issued to this VC wallet, removes wallet's key manager instance and closes wallet content store.
     *
     * @returns {Promise<Object>} - 'object.closed' -  bool flag false if token is not found or already expired for this wallet user.
     */
    async close() {
        return await this.agent.vcwallet.close({userID: this.user})
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
    async add({auth, contentType, content = {}, collectionID} = {}) {
        return await this.agent.vcwallet.add({userID: this.user, auth, contentType, collectionID, content})
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
    async remove({auth = '', contentType = '', contentID = ''} = {}) {
        return await this.agent.vcwallet.remove({userID: this.user, auth, contentType, contentID})
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
    async get({auth = '', contentType = '', contentID = ''} = {}) {
        return await this.agent.vcwallet.get({userID: this.user, auth, contentType, contentID})
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
    async getAll({auth, contentType, collectionID} = {}) {
        return await this.agent.vcwallet.getAll({userID: this.user, auth, contentType, collectionID})
    }

    /**
     *  runs credential queries against wallet credential contents.
     *
     *  @param {String} auth -  authorization token for performing this wallet operation.
     *  @param {Object} query - credential query, refer: https://w3c-ccg.github.io/vp-request-spec/#format
     *
     * @returns {Promise<Object>} - promise of presentation(s) containing credential results or an error if operation fails.
     */
    async query(auth = '', query = []) {
        return await this.agent.vcwallet.query({userID: this.user, auth, query})
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
    async issue(auth = '', credential = {}, proofOptions = {}) {
        return await this.agent.vcwallet.issue({
            userID: this.user,
            auth,
            credential,
            proofOptions
        })
    }

    /**
     *  produces a Verifiable Presentation from wallet.
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
    async prove(auth = '', {storedCredentials = [], rawCredentials = [], presentation = {}} = {},
                proofOptions = {}) {
        return await this.agent.vcwallet.prove({
            userID: this.user,
            auth,
            storedCredentials,
            rawCredentials,
            presentation,
            proofOptions
        })
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
    async verify(auth = '', {storedCredentialID = '', rawCredential = {}, presentation = {}} = {}) {
        return await this.agent.vcwallet.verify({
            userID: this.user,
            auth,
            storedCredentialID,
            rawCredential,
            presentation
        })
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
    async derive(auth = '', {storedCredentialID = '', rawCredential = {}}, deriveOption = {}) {
        return await this.agent.vcwallet.derive({
            userID: this.user,
            auth,
            storedCredentialID,
            rawCredential,
            deriveOption
        })
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
    async createKeyPair(auth, {keyType} = {}) {
        return await this.agent.vcwallet.createKeyPair({
            userID: this.user,
            auth,
            keyType
        })
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
export async function createWalletProfile(agent, userID, {localKMSPassphrase, keyStoreURL, edvConfiguration} = {}) {
    return await agent.vcwallet.createProfile({userID, localKMSPassphrase, keyStoreURL, edvConfiguration})
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
export async function updateWalletProfile(agent, userID, {localKMSPassphrase, keyStoreURL, edvConfiguration} = {}) {
    return await agent.vcwallet.updateProfile({userID, localKMSPassphrase, keyStoreURL, edvConfiguration})
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
    return await agent.vcwallet.profileExists({userID})
}
