/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
// TODO : Implement http message signature module in JS (https://github.com/trustbloc/agent-sdk/issues/348)
// import { encode } from "js-base64";

const GNAP_BASE_PATH = "/gnap";
const AUTH_REQUEST_PATH = GNAP_BASE_PATH + "/auth";
const AUTH_CONTINUE_PATH = GNAP_BASE_PATH + "/continue";
const CONTENT_TYPE = "application/json";

/**
 *  client module is the Authentication Server client that requests GNAP tokens from the Authorization Server.
 *
 *  @module client
 *
 */

/**
 * Client requesting Gnap tokens from the Authorization Server.
 *
 * @alias module:client
 */
export class Client {
  // NewClient creates a new GNAP authorization client. It requires a signer for HTTP Signature header, an HTTP client
  // and a base URL of the authorization server.
  constructor({ signer, gnapAuthServerURL } = {}) {
    this.signer = signer;
    this.gnapAuthServerURL = gnapAuthServerURL;
  }

  // RequestAccess creates a GNAP grant access req then submit it to the server to receive a response with an
  // interact_ref value.
  async requestAccess(req) {
    // TODO : Implement http message signature module in JS (https://github.com/trustbloc/agent-sdk/issues/348)
    // const sig = this.signer.Sign(req);

    const url = this.gnapAuthServerURL + AUTH_REQUEST_PATH;

    const gnapResp = await axios.post(url, req, {
      headers: {
        "Content-Type": CONTENT_TYPE,
        // "Signature-Input": "TODO", // TODO update signature input
        // TODO : Implement http message signature module in JS (https://github.com/trustbloc/agent-sdk/issues/348)
        // Signature: encode(sig, true),
      },
    });

    return gnapResp;
  }

  // Continue gnap auth request containing interact_ref.
  async continue(req) {
    // TODO : Implement http message signature module in JS (https://github.com/trustbloc/agent-sdk/issues/348)
    // const sig = this.signer.Sign(req);

    const gnapResp = await axios.post(
      this.gnapAuthServerURL + AUTH_CONTINUE_PATH,
      req,
      {
        headers: {
          "Content-Type": CONTENT_TYPE,
          // "Signature-Input": "TODO", // TODO update signature input
          // TODO : Implement http message signature module in JS (https://github.com/trustbloc/agent-sdk/issues/348)
          // Signature: encode(sig, true),
        },
      }
    );

    return gnapResp;
  }
}
