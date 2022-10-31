/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import { decode } from "js-base64";
import { CredentialManager, DIDManager } from "@";
import * as jose from "jose";

/**
 *  OpenID4VP module is the oidc client that provides APIs for OIDC4VP flows.
 *
 *  @module OpenID4VP
 *
 */

/**
 * Client providing support for OIDC4VP methods.
 *
 * @alias module:OpenID4VP
 */
export class OpenID4VP {
  // Creates a new OpenID4VP client
  constructor({ agent, user }) {
    if (!agent) {
      throw new TypeError(
        "Error initializing OpenID4VP client: agent cannot be empty"
      );
    } else if (!user) {
      throw new TypeError(
        "Error initializing OpenID4VP client: user cannot be empty"
      );
    }
    this.user = user;
    this.credentialManager = new CredentialManager({ agent, user });
    this.didManager = new DIDManager({ agent, user });
  }

  /**
   *  Fetches OpenID4VP request object and retrieves presentation query for it.
   *
   *  @param {string} authToken - authorization token for wallet operations.
   *  @param {string} url - OpenID4VP presentation request url containing reference to the request object.
   *
   *  @returns {Promise<Array>} - presentation query or error if operation fails.
   */
  async initiateOIDCPresentation({ authToken, url }) {
    if (!authToken) {
      throw new TypeError(
        "Error initiating OIDC presentation: authToken is missing"
      );
    } else if (!url) {
      throw new TypeError("Error initiating OIDC presentation: url is missing");
    }
    const requestURI = new URLSearchParams(url).get("openid-vc://?request_uri");
    if (!requestURI) {
      throw new TypeError(
        "Error initiating OIDC presentation: invalid request url: request_uri is missing"
      );
    }

    // Get base64url-encoded request object
    const encodedRequestToken = await axios(requestURI)
      .then((resp) => resp.data)
      .catch((e) => {
        throw new Error(
          "Error initiating OIDC presentation: failed to get request token:",
          e
        );
      });

    const encodedRequestTokenArray = encodedRequestToken.split(".");

    const header = JSON.parse(decode(encodedRequestTokenArray[0]));
    const payload = JSON.parse(decode(encodedRequestTokenArray[1]));

    const didDoc = await this.didManager.resolveWebDIDFromOrbDID(
      "",
      header.kid
    );

    const { publicKeyJwk } = didDoc.verificationMethod.find(
      (keyPair) => keyPair.id.split("#")[0] === header.kid
    );
    const publicKey = await jose.importJWK(publicKeyJwk, header.alg);

    this.verificationMethodId = didDoc.verificationMethod.id;

    try {
      // TODO: https://github.com/trustbloc/agent-sdk/issues/449
      await jose.compactVerify(encodedRequestToken, publicKey);
    } catch (e) {
      console.error(e);
      throw new Error(
        "Error initiating OIDC presentation: signature verification failed",
        e
      );
    }

    const claims = payload.claims;

    // Get Presentation Query
    const response = await this.credentialManager
      .query(authToken, [
        {
          type: "PresentationExchange",
          credentialQuery: [claims.vp_token],
        },
      ])
      .catch((e) => {
        console.error(e);
        // Error code 12009 is for no result found message
        if (e.message.includes("code: 12009, message: no result found")) {
          throw new Error(
            "Error initiating credential share: requested credentials were not found"
          );
        }
      });

    return response.results;
  }
}
