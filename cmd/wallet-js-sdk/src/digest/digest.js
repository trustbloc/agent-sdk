/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { fromUint8Array } from "js-base64";

/**
 * Generates and returns a hash for the data object provided using the specified encryption algorithm.
 *
 *  @param {String} name - name of the encryption algorithm to be used for generating the digest.
 *  @param {Object} data - data object to generate the digest for.
 *
 * @returns {String} - generated hash for the data provided
 */
export default async function getDigest(name, data) {
  switch (name) {
    case "SHA-256": {
      const encoder = new TextEncoder();
      // hash the data
      const hashBuffer = await window.crypto.subtle.digest(
        name,
        encoder.encode(data) // encode as (utf-8) Uint8Array
      );
      // convert buffer to byte array
      const hashArray = new Uint8Array(hashBuffer);
      // Encode byte array with base64
      const encodedHash = fromUint8Array(hashArray, true);
      return `sha-256=:${encodedHash}:`;
    }
    case "SHA-512": {
      const encoder = new TextEncoder();
      // hash the data
      const hashBuffer = await window.crypto.subtle.digest(
        name,
        encoder.encode(data) // encode as (utf-8) Uint8Array
      );
      // convert buffer to byte array
      const hashArray = new Uint8Array(hashBuffer);
      // Encode byte array with base64
      const encodedHash = fromUint8Array(hashArray, true);
      return `sha-512=:${encodedHash}:`;
    }
    default:
      throw new Error("unsupported digest encryption algorithm provided");
  }
}
