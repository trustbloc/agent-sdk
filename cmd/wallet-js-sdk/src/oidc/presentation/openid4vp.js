/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import { decode, encode } from "js-base64";
import { CredentialManager, DIDManager } from "@";

// TODO: replace mock function with an actual JWT signer once implemented
function signJWT({ header, payload }) {
  const encodedHeader = encode(JSON.stringify(header), true);
  const encodedPayload = encode(JSON.stringify(payload), true);
  const encodedSignature = encode("mock-signature", true);
  return `${encodedHeader}.${encodedPayload}.${encodedSignature}`;
}

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

    const encodedRequestTokenArray = encodedRequestToken.split(".");

    const header = JSON.parse(decode(encodedRequestTokenArray[0]));
    const payload = JSON.parse(decode(encodedRequestTokenArray[1]));

    const { didDocument } = await this.didManager.resolveWebDIDFromOrbDID(
      "",
      header.kid
    );

    this.verificationMethodId = didDocument.verificationMethod.id;

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
  async submitOIDCPresentation({ kid, presentation, expiry, alg = "ES256" }) {
    if (!kid) {
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

    const header = {};
    header["alg"] = alg;
    header["kid"] = kid;
    header["typ"] = "JWT";

    const idToken = await generateIdToken({
      kid,
      presentation: presentation[0].presentation_submission,
      header,
      expiry,
      nonce: this.nonce,
      clientId: this.clientId,
    });

    const vp = {};
    vp["@context"] = presentation["@context"];
    vp["type"] = presentation.type;
    // TODO: encode and sign with JWT
    vp["verifiableCredential"] = presentation.verifiableCredential;

    const vpToken = await generateVpToken({
      kid,
      header,
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
 * @param {Object} header - header for the ID Token.
 * @param {number} expiry - time in seconds representing the expiry of the token.
 *
 *
 * @returns {Promise<string>} - a promise resolving with a string containing ID Token or an error.
 */
async function generateIdToken({
  kid,
  presentation,
  header,
  expiry,
  nonce,
  clientId,
}) {
  if (!kid) {
    throw new TypeError("Error generating ID Token: kid cannot be empty");
  } else if (!presentation) {
    throw new TypeError(
      "Error generating ID Token: presentation cannot be empty"
    );
  } else if (!header) {
    throw new TypeError("Error generating ID Token: header cannot be empty");
  } else if (!expiry) {
    throw new TypeError("Error generating ID Token: expiry cannot be empty");
  } else if (!nonce) {
    throw new TypeError("Error generating ID Token: nonce cannot be empty");
  } else if (!clientId) {
    throw new TypeError("Error generating ID Token: clientId cannot be empty");
  }

  const vpToken = {};
  vpToken["presentation_submission"] = presentation;

  const payload = {};
  payload["sub"] = kid.split("#")[0];
  payload["nonce"] = nonce;
  payload["_vp_token"] = vpToken;
  payload["aud"] = clientId;
  payload["iss"] = "https://self-issued.me/v2/openid-vc";
  payload["exp"] = expiry;

  return signJWT({ header, payload });
}

/**
 * generateVpToken generates a VP Token for the presentation submission request
 * @param {string} kid - consumer's verification method's kid.
 * @param {Object} vp - object containing details for the presentation.
 * @param {Object} header - header for the VP Token.
 * @param {number} expiry - time in seconds representing the expiry of the token.
 *
 *
 * @returns {Promise<string>} - a promise resolving with a string containing ID Token or an error.
 */
async function generateVpToken({ kid, vp, header, expiry, nonce, clientId }) {
  if (!kid) {
    throw new TypeError("Error generating VP Token: kid cannot be empty");
  } else if (!vp) {
    throw new TypeError("Error generating VP Token: vp cannot be empty");
  } else if (!header) {
    throw new TypeError("Error generating VP Token: header cannot be empty");
  } else if (!expiry) {
    throw new TypeError("Error generating VP Token: expiry cannot be empty");
  } else if (!nonce) {
    throw new TypeError("Error generating ID Token: nonce cannot be empty");
  } else if (!clientId) {
    throw new TypeError("Error generating ID Token: clientId cannot be empty");
  }

  const payload = {};
  payload["nonce"] = nonce;
  payload["vp"] = vp;
  payload["aud"] = clientId;
  payload["iss"] = kid.split("#")[0];
  payload["exp"] = expiry;

  return signJWT({ header, payload });
}
