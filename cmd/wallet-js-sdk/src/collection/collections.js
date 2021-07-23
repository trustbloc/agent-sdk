/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { contentTypes, definedProps, UniversalWallet } from "..";

const JSONLD_CTX_COLLECTION = [
  "https://w3id.org/wallet/v1",
  "https://trustbloc.github.io/context/wallet/collections-v1.jsonld",
];
const DEF_COLLECTION_TYPE = "Vault";

var uuid = require("uuid/v4");

/**
 *  collection module provides wallet collection data model features for grouping wallet contents.
 *  This is useful for implementing credential vaults.
 *
 *  @module collection
 *
 */

/**
 * Creating, updating, deleting, querying collections
 *
 * @alias module:collection
 */
export class CollectionManager {
  /**
   *
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
   * Creates new wallet collection model and adds it to wallet content store.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {Object} collection - collection data model.
   *  @param {String} collection.name - display name of the collection.
   *  @param {String} collection.description - display description of the collection.
   *  @param {String} collection.type=ContentVault - optional, custom collection type, default='ContentVault'.
   *
   *  TODO: support for more customized collection parameters like icon, color, storage params etc
   *  for supporting different types of credential vaults.
   *
   * @returns {Promise<string>} - promise containing collection ID or an error if operation fails..
   */
  async create(auth, { name, description, type = DEF_COLLECTION_TYPE } = {}) {
    let content = {
      "@context": JSONLD_CTX_COLLECTION,
      id: uuid(),
      type,
      name,
      description,
    };

    await this.wallet.add({
      auth,
      contentType: contentTypes.COLLECTION,
      content,
    });

    return content.id;
  }

  /**
   * Gets list of all collections from wallet content store.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *
   * @returns {Promise<Object>} - promise containing collection contents or an error if operation fails..
   */
  async getAll(auth) {
    return await this.wallet.getAll({
      auth,
      contentType: contentTypes.COLLECTION,
    });
  }

  /**
   * Gets a collection from wallet content store.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {String} collectionID - ID of the collection to retrieved from store.
   *
   *
   * @returns {Promise<Object>} - promise containing collection content or an error if operation fails..
   */
  async get(auth, collectionID) {
    return await this.wallet.get({
      auth,
      contentType: contentTypes.COLLECTION,
      contentID: collectionID,
    });
  }

  /**
   * Removes a collection from wallet content store and also deletes all contents which belongs to the collection.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {String} collectionID - ID of the collection to retrieved from store.
   *
   *
   * @returns {Promise} - empty promise or an error if operation fails..
   */
  async remove(auth, collectionID) {
    let { contents } = await this.wallet.getAll({
      auth,
      contentType: contentTypes.CREDENTIAL,
      collectionID,
    });

    let vcIDs = Object.keys(contents);
    console.debug(`deleting ${vcIDs.length} VCs from ${collectionID} vault`);

    await Promise.all([
      vcIDs.forEach(
        async (contentID) =>
          await this.wallet.remove({
            auth,
            contentType: contentTypes.CREDENTIAL,
            contentID,
          })
      ),
      this.wallet.remove({
        auth,
        contentType: contentTypes.COLLECTION,
        contentID: collectionID,
      }),
    ]);
  }

  /**
   * Removes a collection from wallet content store and also deletes all the contents which belongs to the collection.
   *
   *  @param {String} auth - authorization token for wallet operations.
   *  @param {String} collectionID - ID of the collection to retrieved from store.
   *  @param {Object} collection - collection data model.
   *  @param {String} collection.name - display name of the collection.
   *  @param {String} collection.description - display description of the collection.
   *
   *
   * @returns {Promise} - empty promise or an error if operation fails..
   */
  async update(auth, collectionID, { name, description } = {}) {
    let { content } = await this.get(auth, collectionID);
    let remove = this.wallet.remove({
      auth,
      contentType: contentTypes.COLLECTION,
      contentID: collectionID,
    });

    let updates = definedProps({ name, description });
    Object.keys(updates).forEach((key) => (content[key] = updates[key]));

    await remove;
    await this.wallet.add({
      auth,
      contentType: contentTypes.COLLECTION,
      content,
    });
  }
}
