/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import { decode } from "js-base64";
import { CredentialManager, DIDManager } from "@";
import * as jose from "jose";

// TODO: replace mock function with an actual JWT signer once implemented
async function signJWT({ header, payload }) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve("mock.signed.jwt");
    }, 300);
  });
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

    this.client_id = payload.client_id;
    this.nonce = payload.nonce;
    this.redirect_uri = payload.redirect_uri;

    return response.results;
  }

  /**
   * submitOIDCPresentation performs an OIDC presentation submission
   * @param {string} kid - consumer's verification method's kid.
   * @param {Object} presentationQuery - presentation query object retrieved from user's wallet.
   * @param {string} issuer - the issuer's key id.
   * @param {number} expiry - time in seconds representing the expiry of the presentation.
   *
   *
   * @returns {Promise<Object>} - empty promise or error if operation fails.
   */
  async submitOIDCPresentation(kid, presentationQuery, issuer, expiry) {
    if (!kid) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: kid cannot be empty"
      );
    } else if (!presentationQuery) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: presentationQuery cannot be empty"
      );
    } else if (!presentationQuery.presentation_submission) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: presentation_submission is missing"
      );
    } else if (!presentationQuery.type) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: type is missing"
      );
    } else if (!presentationQuery.verifiableCredential) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: verifiableCredential is missing"
      );
    } else if (!issuer) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: issuer is missing"
      );
    } else if (!expiry) {
      throw new TypeError(
        "Error submitting OpenID4VP presentation: expiry is missing"
      );
    }

    const header = new Object();
    Object.defineProperty(header, "alg", {
      value: alg,
    });
    Object.defineProperty(header, "kid", {
      value: kid,
    });
    Object.defineProperty(header, "typ", {
      value: "JWT",
    });

    const idToken = await generateIdToken(
      kid,
      presentationQuery.presentation_submission,
      header,
      expiry
    );

    const vp = new Object();
    Object.defineProperty(vp, "@context", {
      value: presentationQuery["@context"],
    });
    Object.defineProperty(vp, "type", {
      value: presentationQuery.type,
    });
    Object.defineProperty(vp, "verifiableCredential", {
      // TODO: encode and sign with JWT
      value: presentationQuery.verifiableCredential,
    });

    const vpToken = await generateVpToken(kid, header, vp, expiry);

    const authRequest = new URLSearchParams();
    authRequest.append("id_token", idToken);
    authRequest.append("vp_token", vpToken);

    return await axios.post(this.redirect_uri, authRequest).catch((e) => {
      throw new Error("Error submitting OIDC presentation:", e);
    });
  }
}

/**
 * generateIdToken generates an ID Token for the presentation submission request
 * @param {string} kid - consumer's verification method's kid.
 * @param {Object} presentationSubmission - presentation submission.
 * @param {Object} header - header for the ID Token.
 * @param {number} expiry - time in seconds representing the expiry of the token.
 *
 *
 * @returns {Promise<string>} - a promise resolving with a string containing ID Token or an error.
 */
async function generateIdToken(kid, presentationSubmission, header, expiry) {
  if (!kid) {
    throw new TypeError("Error generating ID Token: kid cannot be empty");
  } else if (!presentationSubmission) {
    throw new TypeError(
      "Error generating ID Token: presentationSubmission cannot be empty"
    );
  } else if (!header) {
    throw new TypeError("Error generating ID Token: header cannot be empty");
  } else if (!expiry) {
    throw new TypeError("Error generating ID Token: expiry cannot be empty");
  }

  const vpToken = new Object();
  Object.defineProperty(vpToken, "presentation_submission", {
    value: presentationSubmission,
  });

  const payload = new Object();
  Object.defineProperty(payload, "sub", {
    value: kid.split("#")[0],
  });
  Object.defineProperty(payload, "nonce", {
    value: this.nonce,
  });
  Object.defineProperty(payload, "_vp_token", {
    value: vpToken,
  });
  Object.defineProperty(payload, "aud", {
    value: this.client_id,
  });
  Object.defineProperty(payload, "iss", {
    value: "https://self-issued.me/v2/openid-vc",
  });
  Object.defineProperty(payload, "exp", {
    value: expiry,
  });

  return await signJWT({ header, payload });
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
async function generateVpToken(kid, vp, header, expiry) {
  if (!kid) {
    throw new TypeError("Error generating VP Token: kid cannot be empty");
  } else if (!vp) {
    throw new TypeError("Error generating VP Token: vp cannot be empty");
  } else if (!header) {
    throw new TypeError("Error generating VP Token: header cannot be empty");
  } else if (!expiry) {
    throw new TypeError("Error generating VP Token: expiry cannot be empty");
  }

  const payload = new Object();
  Object.defineProperty(payload, "nonce", {
    value: this.nonce,
  });
  Object.defineProperty(payload, "vp", {
    value: vp,
  });
  Object.defineProperty(payload, "aud", {
    value: this.client_id,
  });
  Object.defineProperty(payload, "iss", {
    value: kid.split("#")[0],
  });
  Object.defineProperty(payload, "exp", {
    value: expiry,
  });

  return await signJWT({ header, payload });
}
