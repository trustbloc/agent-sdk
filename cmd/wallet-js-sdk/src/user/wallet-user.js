/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {
  contentTypes,
  createWalletProfile,
  profileExists,
  UniversalWallet,
  updateWalletProfile,
  definedProps,
  DIDManager,
} from "..";

const JSONLD_CTX_USER_PREFERENCE = [
  "https://w3id.org/wallet/v1",
  "https://trustbloc.github.io/context/wallet/user-preferences-v1.jsonld",
];
const METADATA_PREFIX = "user-preference-";
const USER_PREFERENCE_METADATA_TYPE = "UserPreferences";

/**
 *  wallet-user module provides wallet user specific features like maintaining profiles, preferences, locking and unlocking wallets.
 *
 *  @module wallet-user
 *
 */

/**
 *  WalletUser provides wallet user related features like,
 *
 *  - Creating and updating wallet user profiles.
 *  - Saving and updating user wallet preferences.
 *  - Unlocking and locking wallet.
 *
 *  @alias module:wallet-user
 *
 */
export class WalletUser {
  /**
   * @param {String} agent - aries agent.
   * @param {String} user -  unique wallet user identifier, the one used to create wallet profile.
   *
   */
  constructor({ agent, user } = {}) {
    this.agent = agent;
    this.user = user;
    this.wallet = new UniversalWallet({ agent: this.agent, user });
    this.didManager = new DIDManager({ agent: this.agent, user: user });
  }

  /**
   * Create wallet profile for the user and returns error if profile is already created.
   *
   *  @param {String} profileOptions -  options for creating profile.
   *  @param {String} profileOptions.localKMSPassphrase - (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations.
   *  @param {String} profileOptions.keyStoreURL - (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations.
   *  @param {String} profileOptions.edvConfiguration - (optional) EDV configuration if profile wants to use EDV as a wallet content store.
   *  By Default, aries context storage provider will be used.
   *
   *  @param {String} profileOptions.edvConfiguration.serverURL - EDV server URL for storing wallet contents.
   *  @param {String} profileOptions.edvConfiguration.vaultID - EDV vault ID for storing the wallet contents.
   *  @param {String} profileOptions.edvConfiguration.encryptionKID - Encryption key ID of already existing key in wallet profile kms.
   *  If profile is using localkms then wallet will create this key set for wallet user.
   *  @param {String} profileOptions.edvConfiguration.macKID -  MAC operation key ID of already existing key in wallet profile kms.
   *  If profile is using localkms then wallet will create this key set for wallet user.
   *
   * @returns {Promise} - empty promise or an error if operation fails..
   */
  async createWalletProfile({
    localKMSPassphrase,
    keyStoreURL,
    edvConfiguration,
  } = {}) {
    await createWalletProfile(this.agent, this.user, {
      localKMSPassphrase,
      keyStoreURL,
      edvConfiguration,
    });
  }

  /**
   * Updates wallet profile for the user and returns error if profile doesn't exists.
   * Caution:
   *  - you might lose your existing keys if you change kms options.
   *  - you might lose your existing wallet contents if you change storage/EDV options
   *  (ex: switching context storage provider or changing EDV settings).
   *
   *  @param {String} profileOptions -  options for creating profile.
   *  @param {String} profileOptions.localKMSPassphrase - (optional) passphrase for local kms for key operations. If provided then localkms will be used for this wallet profile's key operations.
   *  @param {String} profileOptions.keyStoreURL - (optional) key store URL for web/remote kms. If provided then webkms will be used for this wallet profile's key operations.
   *  @param {String} profileOptions.edvConfiguration - (optional) EDV configuration if profile wants to use EDV as a wallet content store.
   *  By Default, aries context storage provider will be used.
   *
   *  @param {String} profileOptions.edvConfiguration.serverURL - EDV server URL for storing wallet contents.
   *  @param {String} profileOptions.edvConfiguration.vaultID - EDV vault ID for storing the wallet contents.
   *  @param {String} profileOptions.edvConfiguration.encryptionKID - Encryption key ID of already existing key in wallet profile kms.
   *  If profile is using localkms then wallet will create this key set for wallet user.
   *  @param {String} profileOptions.edvConfiguration.macKID -  MAC operation key ID of already existing key in wallet profile kms.
   *  If profile is using localkms then wallet will create this key set for wallet user.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async updateWalletProfile({
    localKMSPassphrase,
    keyStoreURL,
    edvConfiguration,
  } = {}) {
    await updateWalletProfile(this.agent, this.user, {
      localKMSPassphrase,
      keyStoreURL,
      edvConfiguration,
    });
  }

  /**
   *  check is profile exists for given wallet user.
   *
   * @returns {Promise<Boolean>} - true if profile is found.
   */
  async profileExists() {
    let found = true;
    await profileExists(this.agent, this.user).catch((e) => (found = false));

    return found;
  }

  /**
   * Unlocks wallet and returns a authorization token to be used for performing wallet operations.
   *
   *  @param {Object} options
   *  @param {String} options.localKMSPassphrase - (optional) passphrase for local kms for key operations.
   *  @param {Object} options.webKMSAuth - (optional) WebKMSAuth for authorizing access to web/remote kms.
   *  @param {String} options.webKMSAuth.authToken - (optional) Http header 'authorization' bearer token to be used.
   *  @param {String} options.webKMSAuth.capability - (optional) Capability if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.webKMSAuth.authzKeyStoreURL - (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.webKMSAuth.secretShare - (optional) secret share if ZCAP sign header feature to be used for authorizing access.
   *  @param {Object} options.edvUnlocks - (optional) for authorizing access to wallet's EDV content store.
   *  @param {String} options.edvUnlocks.authToken - (optional) Http header 'authorization' bearer token to be used.
   *  @param {String} options.edvUnlocks.capability - (optional) Capability if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.edvUnlocks.authzKeyStoreURL - (optional) authz key store URL if ZCAP sign header feature to be used for authorizing access.
   *  @param {String} options.edvUnlocks.secretShare - (optional) secret share if ZCAP sign header feature to be used for authorizing access.
   *  @param {Time} options.expiry - (optional) time duration in milliseconds for which this profile will be unlocked.
   *
   * @returns {Promise<Object>} - 'object.token' - auth token subsequent use of wallet features.
   */
  //TODO unlock timeout
  async unlock({ localKMSPassphrase, webKMSAuth, edvUnlocks, expiry } = {}) {
    return await this.wallet.open({
      localKMSPassphrase,
      webKMSAuth,
      edvUnlocks,
      expiry,
    });
  }

  /**
   * locks wallet by invalidating previously issued wallet auth.
   * Wallet has to be unlocked again to perform any future wallet operations.
   *
   * @returns {Promise<Bool>} -  bool flag false if token is not found or already expired for this wallet user.
   */
  async lock() {
    return await this.wallet.close({ userID: this.user });
  }

  /**
   * Saves TrustBloc wallet user preferences.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {Object} preferences
   *  @param {String} preferences.name - (optional) wallet user display name.
   *  @param {Object} preferences.description - (optional) wallet user display description.
   *  @param {String} preferences.image - (optional)  wallet user display image in URL format.
   *  @param {String} preferences.controller - (optional) default controller to be used for digital proof for this wallet user.
   *  @param {Boolean} preferences.controllerPublished - (optional) represents whether controller is published or not.
   *  @param {Object} preferences.verificationMethod - (optional) default verificationMethod to be used for digital proof for this wallet user.
   *  @param {String} preferences.proofType - (optional) default proofType to be used for digital proof for this wallet user.
   *  @param {Boolean} preferences.skipWelcomeMsg - (optional) represents whether this wallet user has dismissed a welcome message in the UI.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async savePreferences(
    auth,
    {
      name = "",
      description = "",
      image = "",
      controller = "",
      controllerPublished = false,
      verificationMethod = "",
      proofType = "",
      skipWelcomeMsg = false,
    } = {}
  ) {
    await this.saveMetadata(auth, {
      "@context": JSONLD_CTX_USER_PREFERENCE,
      id: `${METADATA_PREFIX}${this.user}`,
      type: USER_PREFERENCE_METADATA_TYPE,
      name,
      description,
      image,
      controller,
      controllerPublished,
      verificationMethod,
      proofType,
      skipWelcomeMsg,
    });
  }

  /**
   * Updates TrustBloc wallet user preferences.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {Object} preferences
   *  @param {String} preferences.name - (optional) wallet user display name.
   *  @param {Object} preferences.description - (optional) wallet user display description.
   *  @param {String} preferences.image - (optional)  wallet user display image in URL format.
   *  @param {String} preferences.controller - (optional) default controller to be used for digital proof for this wallet user.
   *  @param {Boolean} preferences.controllerPublished - (optional) represents whether controller is published or not.
   *  @param {Object} preferences.verificationMethod - (optional) default verificationMethod to be used for digital proof for this wallet user.
   *  @param {String} preferences.proofType - (optional) default proofType to be used for digital proof for this wallet user.
   *  @param {Boolean} preferences.skipWelcomeMsg - (optional) represents whether this wallet user has dismissed a welcome message in the UI.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async updatePreferences(
    auth,
    {
      name,
      description,
      image,
      controller,
      controllerPublished,
      verificationMethod,
      proofType,
      skipWelcomeMsg,
    }
  ) {
    let { content } = await this.getPreferences(auth);
    if (!content) {
      throw "user preference not found";
    }

    let remove = this.wallet.remove({
      auth,
      contentID: `${METADATA_PREFIX}${this.user}`,
      contentType: contentTypes.METADATA,
    });

    let updates = definedProps({
      name,
      description,
      image,
      controller,
      controllerPublished,
      verificationMethod,
      proofType,
      skipWelcomeMsg,
    });

    Object.keys(updates).forEach((key) => (content[key] = updates[key]));

    await remove;

    await this.saveMetadata(auth, content);
  }

  /**
   * Gets TrustBloc wallet user preference.
   *
   * If controller not published, then this function checks if that controller is published.
   * If published then it change published to true in underlying wallet content store and updates user preference.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @returns {Promise<Object>} - promise containing preference metadata or error if operation fails.
   */
  async getPreferences(auth) {
    let result = await this.wallet.get({
      auth,
      contentType: contentTypes.METADATA,
      contentID: `${METADATA_PREFIX}${this.user}`,
    });

    if (
      result.content.controllerPublished === false && result.content.controller.includes(`did:web`)
    ) {
      console.log("check DID is published")

      let published = await this.didManager.checkControllerIsPublished(
        auth,
        result.content.controller
      );

      if (published === true) {
        await this.wallet.remove({
          auth,
          contentID: `${METADATA_PREFIX}${this.user}`,
          contentType: contentTypes.METADATA,
        });

        console.log(
          "DID is published"
        );
        result.content.controllerPublished = published;

        await this.saveMetadata(auth, result.content);
      }
    }

    return result;
  }

  /**
   * Saves custom metadata data model into wallet.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {Object} content - metadata to be saved in wallet content store.
   *
   *  @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async saveMetadata(auth, content) {
    await this.wallet.add({
      auth,
      contentType: contentTypes.METADATA,
      content,
    });
  }

  /**
   * Gets metadata by ID from wallet.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {String} contentID - ID of the metadata to be read from wallet content store.
   *
   *  @returns {Promise<Object>} result.content - promise containing metadata or error if operation fails.
   */
  async getMetadata(auth, contentID) {
    return await this.wallet.get({
      auth,
      contentType: contentTypes.METADATA,
      contentID,
    });
  }

  /**
   * Gets All metadata data models from wallet.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *
   *  @returns {Promise<Object>} result.contents - promise containing result or error if operation fails.
   */
  async getAllMetadata(auth) {
    return await this.wallet.getAll({
      auth,
      contentType: contentTypes.METADATA,
    });
  }
}
