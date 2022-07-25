/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import getDigest from "../digest/digest";

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
    const digest = await getDigest("SHA-256", req);
    const url = this.gnapAuthServerURL + AUTH_REQUEST_PATH;
    const signatureName = "sig1";

    // Generate signature params
    const sigParams = this.signer.generateSignatureParams();
    // Get signature input
    const signatureInput = this.signer.getSignatureInput(
      signatureName,
      sigParams
    );
    // Get http signature
    const signature = await this.signer.sign(
      digest,
      url,
      signatureName,
      sigParams
    );

    const gnapResp = await axios.post(url, req, {
      headers: {
        "Content-Type": CONTENT_TYPE,
        "content-digest": digest,
        "Signature-Input": signatureInput,
        Signature: signature,
      },
    });

    return gnapResp;
  }

  // Continue gnap auth request containing interact_ref.
  async continue(req, continue_token) {
    const digest = await getDigest("SHA-256", req);
    const url = this.gnapAuthServerURL + AUTH_CONTINUE_PATH;
    const signatureName = "sig1";

    // Generate signature params
    const sigParams = this.signer.generateSignatureParams();
    // Get signature input
    const signatureInput = this.signer.getSignatureInput(
      signatureName,
      sigParams
    );
    // Get http signature
    const signature = await this.signer.sign(
      digest,
      url,
      signatureName,
      sigParams
    );

    const gnapResp = await axios.post(url, req, {
      headers: {
        "Content-Type": CONTENT_TYPE,
        Authorization: "GNAP " + continue_token,
        "Content-Digest": digest,
        "Signature-Input": signatureInput,
        Signature: signature,
      },
    });

    return gnapResp;
  }
}
