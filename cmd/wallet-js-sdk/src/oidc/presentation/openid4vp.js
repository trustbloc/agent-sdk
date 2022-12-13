/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import { decode } from "js-base64";
import { CredentialManager, JWTManager } from "@";

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
    this.jwtManager = new JWTManager({ agent, user });
  }

  /**
   *  Fetches OpenID4VP request object and retrieves presentation query for it.
   *
   *  @param {string} authToken - authorization token for wallet operations.
   *  @param {string} url - OpenID4VP presentation request url containing reference to the request object.
   *
   *  @returns {Promise<Array>} - presentation or error if operation fails.
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

    const jwtVerificationStatus = await this.jwtManager.verifyJWT(authToken, {
      jwt: encodedRequestToken,
    });

    if (!jwtVerificationStatus.verified) {
      throw new Error(
        "Error initiating OIDC presentation: failed to verify signature on the request token:",
        jwtVerificationStatus.error
      );
    }

    const payload = JSON.parse(decode(encodedRequestToken.split(".")[1]));
    const claims = payload.claims;

    // Get Presentation
    const response = await this.credentialManager
      .query(authToken, [
        {
          type: "PresentationExchange",
          credentialQuery: [claims.vp_token.presentation_definition],
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

    this.clientId = payload.client_id;
    this.nonce = payload.nonce;
    this.redirectUri = payload.redirect_uri;

    return response.results;
  }

  /**
   * submitOIDCPresentation performs an OIDC presentation submission
   * @param {string} kid - consumer's verification method's kid.
   * @param {Object} presentation - presentation array retrieved from user's wallet.
   * @param {number} expiry - time in seconds representing the expiry of the presentation.
   * @param {string} alg - encryption algorithm to be used for signing
   *
   *
   * @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async submitOIDCPresentation({ authToken, kid, presentation, expiry }) {
    if (!authToken) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: authToken is missing"
      );
    } else if (!kid) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: kid cannot be empty"
      );
    } else if (!presentation && !presentation.length) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: presentation cannot be empty"
      );
    } else if (!presentation[0].presentation_submission) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: presentation_submission is missing"
      );
    } else if (!presentation[0].type) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: type is missing"
      );
    } else if (!presentation[0].verifiableCredential) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: verifiableCredential is missing"
      );
    } else if (!expiry) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: expiry is missing"
      );
    }

    const headers = {};
    headers["kid"] = kid;
    headers["typ"] = "JWT";

    const { jwt: idToken } = await generateIdToken(this.jwtManager, authToken, {
      kid,
      presentation: presentation[0].presentation_submission,
      headers,
      expiry,
      nonce: this.nonce,
      clientId: this.clientId,
    });

    const vp = {};
    vp["@context"] = presentation["@context"];
    vp["type"] = presentation.type;
    vp["verifiableCredential"] = presentation[0].verifiableCredential;

    const { jwt: vpToken } = await generateVpToken(this.jwtManager, authToken, {
      kid,
      headers,
      vp,
      expiry,
      nonce: this.nonce,
      clientId: this.clientId,
    });

    const authRequest = new URLSearchParams();
    authRequest.append("id_token", idToken);
    authRequest.append("vp_token", vpToken);

    return await axios.post(this.redirectUri, authRequest).catch((e) => {
      throw new Error("Error submitting OIDC presentation:", e);
    });
  }
}

/**
 * generateIdToken generates an ID Token for the presentation submission request
 * @param {string} kid - consumer's verification method's kid.
 * @param {Object} presentation - presentation.
 * @param {Object} headers - headers for the ID Token.
 * @param {number} expiry - time in seconds representing the expiry of the token.
 *
 *
 * @returns {Promise<string>} - a promise resolving with a string containing ID Token or an error.
 */
async function generateIdToken(
  jwtManager,
  authToken,
  { kid, presentation, headers, expiry, nonce, clientId }
) {
  if (!kid) {
    throw new TypeError("Error generating ID Token: kid cannot be empty");
  } else if (!presentation) {
    throw new TypeError(
      "Error generating ID Token: presentation cannot be empty"
    );
  } else if (!headers) {
    throw new TypeError("Error generating ID Token: headers cannot be empty");
  } else if (!expiry) {
    throw new TypeError("Error generating ID Token: expiry cannot be empty");
  } else if (!nonce) {
    throw new TypeError("Error generating ID Token: nonce cannot be empty");
  } else if (!clientId) {
    throw new TypeError("Error generating ID Token: clientId cannot be empty");
  }

  const vpToken = {};
  vpToken["presentation_submission"] = presentation;

  const claims = {};
  claims["sub"] = kid.split("#")[0];
  claims["nonce"] = nonce;
  claims["_vp_token"] = vpToken;
  claims["aud"] = clientId;
  claims["iss"] = "https://self-issued.me/v2/openid-vc";
  claims["exp"] = expiry;

  return jwtManager.signJWT(authToken, { headers, claims, kid });
}

/**
 * generateVpToken generates a VP Token for the presentation submission request
 * @param {string} kid - consumer's verification method's kid.
 * @param {Object} vp - object containing details for the presentation.
 * @param {Object} headers - headers for the VP Token.
 * @param {number} expiry - time in seconds representing the expiry of the token.
 *
 *
 * @returns {Promise<string>} - a promise resolving with a string containing ID Token or an error.
 */
async function generateVpToken(
  jwtManager,
  authToken,
  { kid, vp, headers, expiry, nonce, clientId }
) {
  if (!kid) {
    throw new TypeError("Error generating VP Token: kid cannot be empty");
  } else if (!vp) {
    throw new TypeError("Error generating VP Token: vp cannot be empty");
  } else if (!headers) {
    throw new TypeError("Error generating VP Token: headers cannot be empty");
  } else if (!expiry) {
    throw new TypeError("Error generating VP Token: expiry cannot be empty");
  } else if (!nonce) {
    throw new TypeError("Error generating VD Token: nonce cannot be empty");
  } else if (!clientId) {
    throw new TypeError("Error generating VD Token: clientId cannot be empty");
  }

  const claims = {};
  claims["nonce"] = nonce;
  claims["vp"] = vp;
  claims["aud"] = clientId;
  claims["iss"] = kid.split("#")[0];
  claims["exp"] = expiry;

  return jwtManager.signJWT(authToken, { headers, claims, kid });
}
