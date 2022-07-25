/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { expect } from "chai";
import getDigest from "../../../src/digest/digest";

const kty = "EC";
const kid = "key1";
const crv = "P-256";
const alg = "ES256";
const x = "TDiE9hqocEOewpSgI5r_yhoOy1EuetUpWcKvlrC9Ix4";
const y = "pC76IX_6YRQ5CaxmcmcixTA-dWjHpuKoYICir7VYQ_A";

const expectThrowsAsync = async (method, errorMessage) => {
  let error = null;
  try {
    await method();
  } catch (err) {
    error = err;
  }
  expect(error).to.be.an("Error");
  if (errorMessage) {
    expect(error.message).to.equal(errorMessage);
  }
};

describe("HTTP Signature Module - Generating Digest", function () {
  const mockData = {
    access_token: [
      {
        access: [
          {
            type: "trustbloc.dev/types/gnap/userdata-access",
            actions: ["read", "write", "update", "delete"],
            locations: ["edv.provider.com/api"],
            "subject-keys": ["sub"],
          },
          {
            type: "trustbloc.dev/types/gnap/kms-access",
            actions: ["sign", "verify", "encrypt", "decrypt"],
            locations: ["kms.provider.com/api/user"],
            "subject-keys": ["sub"],
          },
        ],
        label: "access-1",
      },
      {
        access: [
          {
            type: "trustbloc.dev/types/gnap/legacy-api",
            actions: ["read", "append"],
            locations: ["log.legacy.com/user"],
          },
        ],
        label: "log-access",
        flags: ["bearer"],
      },
    ],
    client: {
      key: {
        proof: "httpsig",
        jwk: {
          kty,
          kid,
          crv,
          alg,
          x,
          y,
        },
      },
    },
    interact: {
      start: ["redirect"],
      finish: {
        method: "redirect",
        uri: "https://wallet.app/login/callback",
        nonce: "foo-bar-baz",
      },
    },
  };

  it("successfully generates a hash using a SHA-256 encryption algorithm name for the provided data", async () => {
    const mockValidAlgorithmName = "SHA-256";
    const hash = await getDigest(mockValidAlgorithmName, mockData);
    expect(hash).to.equal(
      `sha-256=:soyUshlcjtJZ8LQVqu4/ObCykgpFN2EUmfoESVaReiE=:`
    );
  });

  it("successfully generates a hash using a SHA-512 encryption algorithm name for the provided data", async () => {
    const mockValidAlgorithmName = "SHA-512";
    const hash = await getDigest(mockValidAlgorithmName, mockData);
    expect(hash).to.equal(
      `sha-512=:HcytP60FiinM744AP6cbur9YdDGsWlX7NiaL95WMXzyzERasnoVexhu5ty7L1IT3BL7gMnB/sOrSStK+6XuaOQ==:`
    );
  });

  it("throws an error for unsupported encryption algorithm provided", async () => {
    const mockInvalidAlgorithmName = "SHA-1";
    await expectThrowsAsync(
      () => getDigest(mockInvalidAlgorithmName, mockData),
      "unsupported digest encryption algorithm provided"
    );
  });
});
