/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {contentTypes, createWalletProfile, UniversalWallet, updateWalletProfile} from "..";

const JSONLD_CTX_USER_PREFERENCE = ['https://w3id.org/wallet/v1', 'https://trustbloc.github.io/context/wallet/user-preferences-v1.jsonld']
const METADATA_PREFIX = 'user-preference-'
const USER_PREFERENCE_METADATA_TYPE = 'UserPreferences'

/**
 *  WalletUser provides wallet user related features like,
 *
 *  - Creating and updating wallet user profiles.
 *  - Saving and updating user wallet preferences.
 *  - Unlocking and locking wallet.
 *
 */
export class WalletUser {
    /**
     *
     * @class WalletUser
     *
     * @param {string} agent - aries agent.
     * @param {string} user -  unique wallet user identifier, the one used to create wallet profile.
     *
     */
    constructor({agent, user} = {}) {
        this.agent = agent
        this.user = user
        this.wallet = new UniversalWallet({agent: this.agent, user})
    }

    /**
     * Create wallet profile for the user and returns error if profile is already created.
     *
     *  @param {string} profileOptions -  options for creating profile.
     *  @param {string} profileOptions.localKMSPassphrase - (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations.
     *  @param {string} profileOptions.keyStoreURL - (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations.
     *  @param {string} profileOptions.edvConfiguration - (optional) EDV configuration if profile wants to use EDV as a wallet content store.
     *  By Default, aries context storage provider will be used.
     *
     * @returns {Promise} - empty promise or an error if operation fails..
     */
    async createWalletProfile({localKMSPassphrase, keyStoreURL, edvConfiguration} = {}) {
        await createWalletProfile(this.agent, this.user, {localKMSPassphrase, keyStoreURL, edvConfiguration})
    }


    /**
     * Updates wallet profile for the user and returns error if profile doesn't exists.
     * Caution:
     *  - you might lose your existing keys if you change kms options.
     *  - you might lose your existing wallet contents if you change storage/EDV options
     *  (ex: switching context storage provider or changing EDV settings).
     *
     *  @param {string} profileOptions -  options for creating profile.
     *  @param {string} profileOptions.localKMSPassphrase - (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations.
     *  @param {string} profileOptions.keyStoreURL - (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations.
     *  @param {string} profileOptions.edvConfiguration - (optional) EDV configuration if profile wants to use EDV as a wallet content store.
     *  By Default, aries context storage provider will be used.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async updateWalletProfile({localKMSPassphrase, keyStoreURL, edvConfiguration} = {}) {
        await updateWalletProfile(this.agent, this.user, {localKMSPassphrase, keyStoreURL, edvConfiguration})
    }

    /**
     * Unlocks wallet and returns a authorization token to be used for performing wallet operations.
     *
     *  @param {Object} options
     *  @param {string} options.localKMSPassphrase - (optional) passphrase for local kms for key operations.
     *  @param {Object} options.webKMSAuth - (optional) WebKMSAuth for authorizing access to web/remote kms.
     *  @param {string} options.webKMSAuth.authToken - (optional) Http header 'authorization' bearer token to be used.
     *  @param {string} options.webKMSAuth.capability - (optional) Capability if ZCAP sign header feature to be used for authorizing access.
     *  @param {Object} options.edvUnlocks - (optional) for authorizing access to wallet's EDV content store.
     *  @param {string} options.edvUnlocks.authToken - (optional) Http header 'authorization' bearer token to be used.
     *  @param {string} options.edvUnlocks.capability - (optional) Capability if ZCAP sign header feature to be used for authorizing access.
     *
     * @returns {Promise<Object>} - 'object.token' - auth token subsequent use of wallet features.
     */
    //TODO unlock timeout
    async unlock({localKMSPassphrase, webKMSAuth, edvUnlocks} = {}) {
        return await this.wallet.open({localKMSPassphrase, webKMSAuth, edvUnlocks})
    }

    /**
     * locks wallet by invalidating previously issued wallet auth.
     * Wallet has to be unlocked again to perform any future wallet operations.
     *
     * @returns {Promise<Object>} - 'object.closed' -  bool flag false if token is not found or already expired for this wallet user.
     */
    async lock() {
        return await this.wallet.close({userID: this.user})
    }


    /**
     * Saves TrustBloc wallet user preferences.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *  @param {Object} preferences
     *  @param {string} preferences.name - (optional) wallet user display name.
     *  @param {Object} preferences.description - (optional) wallet user display description.
     *  @param {string} preferences.image - (optional)  wallet user display image in URL format.
     *  @param {string} preferences.controller - (optional) default controller to be used for digital proof for this wallet user.
     *  @param {Object} preferences.verificationMethod - (optional) default verificationMethod to be used for digital proof for this wallet user.
     *  @param {string} preferences.proofType - (optional) default proofType to be used for digital proof for this wallet user.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async savePreferences(auth, {name, description, image, controller, verificationMethod, proofType} = {}) {
        await this.saveMetadata(auth, {
            "@context": JSONLD_CTX_USER_PREFERENCE,
            id: `${METADATA_PREFIX}${this.user}`,
            type: USER_PREFERENCE_METADATA_TYPE,
            name, description, image, controller, verificationMethod, proofType
        })
    }

    /**
     * Updates TrustBloc wallet user preferences.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *  @param {Object} preferences
     *  @param {string} preferences.name - (optional) wallet user display name.
     *  @param {Object} preferences.description - (optional) wallet user display description.
     *  @param {string} preferences.image - (optional)  wallet user display image in URL format.
     *  @param {string} preferences.controller - (optional) default controller to be used for digital proof for this wallet user.
     *  @param {Object} preferences.verificationMethod - (optional) default verificationMethod to be used for digital proof for this wallet user.
     *  @param {string} preferences.proofType - (optional) default proofType to be used for digital proof for this wallet user.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async updatePreferences(auth, {name, description, image, controller, verificationMethod, proofType} = {}) {
        let {content} = await this.getPreferences(auth)
        if (!content) {
            throw 'user preference not found'
        }

        let remove = this.wallet.remove({
            auth,
            contentID: `${METADATA_PREFIX}${this.user}`,
            contentType: contentTypes.METADATA
        })
        let updates = {name, description, image, controller, verificationMethod, proofType}

        const newContent = {};
        Object.keys(content).forEach((key) => newContent[[key]] = updates[key] ? updates[key] : content[key]);

        await remove
        await this.saveMetadata(auth, newContent)
    }

    /**
     * Gets TrustBloc walletuser preference.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async getPreferences(auth) {
        return await this.wallet.get({
            auth,
            contentType: contentTypes.METADATA,
            contentID: `${METADATA_PREFIX}${this.user}`
        })
    }

    /**
     * Saves custom metadata data model into wallet.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *  @param {Object} content - metadata to be saved in wallet content store.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async saveMetadata(auth, content) {
        await this.wallet.add({auth, contentType: contentTypes.METADATA, content})
    }

    /**
     * Gets metadata by ID from wallet.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *  @param {string} contentID - ID of the metadata to be read from wallet content store.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async getMetadata(auth, contentID) {
        return await this.wallet.get({auth, contentType: contentTypes.METADATA, contentID})
    }

    /**
     * Gets All metadata data models from wallet.
     *
     *  @param {string} auth - authorization token for wallet operations.
     *
     *  @returns {Promise<Object>} - empty promise or error if operation fails.
     */
    async getAllMetadata(auth) {
        return await this.wallet.getAll({auth, contentType: contentTypes.METADATA})
    }
}
