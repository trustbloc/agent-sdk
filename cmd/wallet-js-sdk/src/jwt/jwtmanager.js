/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { UniversalWallet } from "@";

/**
 *  jwt-manager module provides JWT related features for wallet like signing and verifying.
 *
 *  @module jwt-manager
 */

/**
 * JWT Manager provides JWT related features for wallet like,
 *
 *  - Signing JWT.
 *  - Verifying JWT.
 *
 *  @alias module:jwt-manager
 */
export class JWTManager {
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
   * Signs a JWT using a key in wallet.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {String} options.headers - JWT token headers.
   *  @param {String} options.claims - JWT token claims.
   *  @param {String} options.kid - wallet's key id.
   *
   * @returns {Promise} - promise of object containing signed JWT string.
   */
  async signJWT(auth, { headers, claims, kid } = {}) {
    return await this.wallet.signJWT({
      auth,
      headers,
      claims,
      kid,
    });
  }

  /**
   * Verifies a JWT using wallet.
   *
   *  @param {Object} options
   *  @param {string} options.auth - authorization token for wallet operations.
   *  @param {String} options.jwt - JWT token to be verified.
   *
   * @returns {Promise} - promise of object containing a boolean representing verification result or an error if operation fails.
   */
  async verifyJWT(auth, { jwt } = {}) {
    return await this.wallet.verifyJWT({
      auth,
      jwt,
    });
  }
}
