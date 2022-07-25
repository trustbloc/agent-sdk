/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { fromUint8Array } from "js-base64";

/**
 *  HTTPSigner provides functionality to generate http signatures for requests
 *
 *  @module HTTPSigner
 *
 */

/**
 *  - generate signature params.
 *  - get signature input.
 *  - sign requests.
 *
 * @alias module:HTTPSigner
 */
export class HTTPSigner {
  constructor({ authorization = "", signingKey }) {
    if (!signingKey) {
      throw new Error(
        "Error initializing HTTPSigner: signingKey cannot be empty"
      );
    } else if (!signingKey.publicKey) {
      throw new Error(
        "Error initializing HTTPSigner: publicKey cannot be empty"
      );
    } else if (!signingKey.privateKey) {
      throw new Error(
        "Error initializing HTTPSigner: privateKey cannot be empty"
      );
    } else if (!signingKey.kid) {
      throw new Error("Error initializing HTTPSigner: kid cannot be empty");
    }
    this.authorization = authorization;
    this.signingKey = signingKey;
  }

  /**
   * Generates and returns signature params string.
   *
   * @returns {String} - generated signature params
   */
  generateSignatureParams() {
    if (!this.signingKey.kid)
      throw new Error("Error generating signature params: kid is required");

    const created = Math.floor(Date.now() / 1000);

    const signatureParams = `("@method" "@target-uri" "content-digest"${
      this.authorization && ' "authorization"'
    });created=${created};keyid="${this.signingKey.kid}"`;

    return signatureParams;
  }

  /**
   * Generates and returns a signature input for the name and signature params provided.
   *
   *  @param {String} name - name of the signature
   *  @param {String} sigParams - signature parameters
   *
   * @returns {String} - generated signature input
   */
  getSignatureInput(name, sigParams) {
    if (!name)
      throw new Error("Error getting signature input: name is required");
    return `${name}=${sigParams}`;
  }

  /**
   * Returns the proof type of the http signature.
   *
   * @returns {String} - proof type of the http signature
   */
  proofType() {
    return "httpsig";
  }

  /**
   * Generates and returns http signature for the request provided.
   *
   *  @param {String} digest - digest value for the request that's being signed
   *  @param {String} url - url of the server
   *  @param {String} name - signature name
   *  @param {String} sigParams - signature parameters
   *
   * @returns {String} - generated signature for the data provided
   */
  async sign(digest, url, name, sigParams) {
    if (!digest)
      throw new Error("Error generating a signature: digest is missing");
    if (!url)
      throw new Error("Error generating a signature: server url is missing");
    if (!name) throw new Error("Error generating a signature: name is missing");
    if (!sigParams)
      throw new Error(
        "Error generating a signature: signature parameters are missing"
      );

    const method = "POST";

    const signatureBase = this.authorization
      ? `"@method": ${method}\n"@target-uri": ${url}\n"content-digest": ${digest}\n"authorization": GNAP ${this.authorization}\n"@signature-params": ${sigParams}`
      : `"@method": ${method}\n"@target-uri": ${url}\n"content-digest": ${digest}\n"@signature-params": ${sigParams}`;

    const encoder = new TextEncoder();
    const encodedSignatureBase = encoder.encode(signatureBase);

    const sigBuffer = await window.crypto.subtle.sign(
      {
        name: "ECDSA",
        hash: { name: "SHA-256" },
      },
      this.signingKey.privateKey,
      encodedSignatureBase
    );

    // convert buffer to byte array
    const sigArray = new Uint8Array(sigBuffer);

    // Encode byte array with base64
    const encodedSig = fromUint8Array(sigArray);

    return `${name}=:${encodedSig}:`;
  }
}
