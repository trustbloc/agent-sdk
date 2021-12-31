/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { contentTypes, getMediatorConnections, UniversalWallet } from "..";
import {expect} from "chai";

const DEFAULT_KEY_TYPE = "ED25519";
const DEFAULT_SIGNATURE_TYPE = "Ed25519VerificationKey2018";
const DEFAULT_KEYAGREEMENT_KEY_TYPE = "X25519ECDHKW";
const DEFAULT_KEYAGREEMENT_TYPE = "X25519KeyAgreementKey2019";
const JSONLD_CTX_KEY = ["https://w3id.org/wallet/v1"];

/**
 *  did-manager module provides DID related features for wallet like creating, importing & saving DIDs into wallets.
 *
 *  @module did-manager
 */

/**
 * DID Manger provides DID related features for wallet like,
 *
 *  - Creating Orb DIDs.
 *  - Creating Peer DIDs.
 *  - Saving Custom DIDs along with keys.
 *  - Getting all Saved DIDs.
 *
 *  @alias module:did-manager
 */
export class DIDManager {
  /**
   * @param {string} agent - aries agent.
   * @param {string} user -  unique wallet user identifier, the one used to create wallet profile.
   *
   */
  constructor({ agent, user } = {}) {
    this.agent = agent;
    this.wallet = new UniversalWallet({ agent: this.agent, user });
  }

  /**
   * Creates Orb DID and saves it in wallet content store.
   *
   * @see {@link https://trustbloc.github.io/did-method-orb|The did:orb Method}
   *
   *  @param {string} auth - authorization token for wallet operations.
   *  @param {Object} options - options for creating Orb DID.
   *  @param {Object} options.keyType=ED25519 - (optional, default ED25519) type of the key to be used for creating keys for the DID, Refer agent documentation for supported key types.
   *  @param {Object} options.keyAgreementKeyType=X25519ECDHKW - (optional, default X25519ECDHKW) type of the key to be used for creating keyAgreements for the DID, Refer agent documentation for supported key types.
   *  @param {String} options.signatureType=Ed25519VerificationKey2018 - (optional, default Ed25519VerificationKey2018) signature type to be used for DID verification methods.
   *  @param {String} options.keyAgreementType=X25519KeyAgreementKey2019 - (optional, default X25519KeyAgreementKey2019) keyAgreement VM type to be used for DID key agreement (payload encryption). For JWK type, use `JsonWebKey2020`.
   *  @param {Array<String>} options.purposes=authentication - (optional, default "authentication") purpose of the key.
   *  @param {String} options.collection - (optional, default no collection) collection to which this DID should belong in wallet content store.
   *
   * @returns {Promise<Object>} - Promise of DID Resolution response  or an error if operation fails..
   */
  async createOrbDID(
    auth,
    {
      keyType = DEFAULT_KEY_TYPE,
      keyAgreementKeyType = DEFAULT_KEYAGREEMENT_KEY_TYPE,
      signatureType = DEFAULT_SIGNATURE_TYPE,
      keyAgreementType = DEFAULT_KEYAGREEMENT_TYPE,
      purposes = ["authentication"],
      collection,
    } = {}
  ) {
    const [keySet, recoveryKeySet, updateKeySet, keyAgreementKeySet] = await Promise.all([
      this.wallet.createKeyPair(auth, { keyType }),
      this.wallet.createKeyPair(auth, { keyType: DEFAULT_KEY_TYPE }),
      this.wallet.createKeyPair(auth, { keyType: DEFAULT_KEY_TYPE }),
      this.wallet.createKeyPair(auth, { keyType: keyAgreementKeyType }),
    ]);

    const createDIDRequest = {
      publicKeys: [
        {
          id: keySet.keyID,
          type: signatureType,
          value: keySet.publicKey,
          encoding: "Jwk",
          keyType: keyType,
          purposes: purposes,
        },
        {
          id: recoveryKeySet.keyID,
          type: DEFAULT_SIGNATURE_TYPE,
          value: recoveryKeySet.publicKey,
          encoding: "Jwk",
          keyType: DEFAULT_KEY_TYPE,
          recovery: true,
        },
        {
          id: updateKeySet.keyID,
          type: DEFAULT_SIGNATURE_TYPE,
          value: updateKeySet.publicKey,
          encoding: "Jwk",
          keyType: DEFAULT_KEY_TYPE,
          update: true,
        },
        {
          id: keyAgreementKeySet.keyID,
          type: keyAgreementType,
          value: keyAgreementKeySet.publicKey,
          encoding: "JWK",
          keyType: keyAgreementKeyType,
          purposes: ["keyAgreement"],
        },
      ],
    };

    let content
    try {
      content = await this.agent.didclient.createOrbDID(createDIDRequest);
    } catch (e) {
      console.error('failed to create orbDID',e)
      expect.fail(e);
    }


    console.debug(
      "created and saved Orb DID successfully",
      content.didDocument.id
    );

    return content;
  }

  /**
   * Creates Peer DID and saves it in wallet content store.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {string} options.collection - (optional, default no collection) collection to which this DID should belong in wallet content store.
   *
   * @returns {Promise} - empty promise or an error if operation fails..
   */
  async createPeerDID(auth, { collection } = {}) {
    let content = await this.agent.didclient.createPeerDID({
      routerConnectionID: await getMediatorConnections(this.agent, {
        single: true,
      }),
    });

    await this.saveDID(auth, { content, collection });

    console.debug("created and saved peer DID successfully");
    return content;
  }

  /**
   * Saves given DID content to wallet content store.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {string} options.content - DID document content.
   *  @param {string} options.collection - (optional, default no collection) collection to which this DID should belong in wallet content store.
   *
   * @returns {Promise} - empty promise or an error if operation fails..
   */
  async saveDID(auth, { content, collection } = {}) {
    await this.wallet.add({
      auth,
      contentType: contentTypes.DID_RESOLUTION_RESPONSE,
      collectionID: collection,
      content,
    });
  }

  /**
   * Resolves and saves DID document into wallet content store along with keys.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {string} options.did - ID of the DID to be imported.
   *  @param {string} options.key - (optional, default no collection) collection to which this DID should belong in wallet content store.
   *  @param {string} options.collection - (optional, default no collection) collection to which this DID should belong in wallet content store.
   *
   * @returns {Promise} - empty promise or an error if operation fails..
   */
  async importDID(auth, { did, key, collection } = {}) {
    if (key) {
      let { privateKeyJwk, privateKeyBase58, keyType, keyID } = key;
      await this.wallet.add({
        auth,
        contentType: contentTypes.KEY,
        collectionID: collection,
        content: {
          "@context": JSONLD_CTX_KEY,
          id: keyID,
          type: keyType,
          privateKeyJwk,
          privateKeyBase58,
        },
      });
    }

    let content = await this.agent.vdr.resolveDID({ id: did });

    await this.saveDID(auth, { content, collection });
  }

  /**
   * gets all DID contents from wallet content store.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {string} options.collection - (optional, default no collection) to filter DID contents based on collection ID.
   *
   * @returns {Promise<Object>} - result.contents - collection of DID documents by IDs.
   */
  async getAllDIDs(auth, { collection } = {}) {
    return await this.wallet.getAll({
      auth,
      contentType: contentTypes.DID_RESOLUTION_RESPONSE,
      collectionID: collection,
    });
  }

  /**
   * get DID content from wallet content store.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {string} options.contentID - DID ID.
   *
   * @returns {Promise<Object>} - result.content - DID document resolution from wallet content store.
   */
  async getDID(auth, contentID) {
    return await this.wallet.get({
      auth,
      contentType: contentTypes.DID_RESOLUTION_RESPONSE,
      contentID,
    });
  }

  /**
   * resolve orb DID.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {string} options.contentID - DID ID.
   *
   * @returns {Promise<Object>} - result.content - DID document resolution from did resolver.
   */
  async resolveOrbDID(auth, contentID) {
    const createDIDRequest = {
      did: contentID,
    };

    let content = await this.agent.didclient.resolveOrbDID(createDIDRequest);

    console.debug("resolve Orb DID successfully", content.didDocument.id);

    return content;
  }
}
