/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { expect } from "chai";
import moxios from "moxios";
import { GNAPClient } from "../../../src";

const GNAP_AUTH_SERVER_URL = "https://auth.trustbloc.local:8070";

describe("GNAP Auth Client - Requesting Access", function () {
  beforeEach(function () {
    moxios.install();
  });

  afterEach(function () {
    moxios.uninstall();
  });

  const mockRequest = {
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
          kty: "EC",
          kid: "key1",
          crv: "P-256",
          alg: "ES256",
          x: "TDiE9hqocEOewpSgI5r_yhoOy1EuetUpWcKvlrC9Ix4",
          y: "pC76IX_6YRQ5CaxmcmcixTA-dWjHpuKoYICir7VYQ_A",
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

  it("success requesting gnap access", async () => {
    const expectedResp = {
      continue: {
        uri: "https://auth.trustbloc.local:8070/gnap/continue",
        access_token: {
          value: "4ebb032b-9fba-4eb0-bf1f-0d2e5d10d9bf",
          label: "",
          manage: "",
          access: null,
          expires_in: 0,
          key: "",
          flags: null,
        },
        wait: 0,
      },
      interact: {
        redirect:
          "https://auth.trustbloc.local:8070/gnap/interact?txnID=dYnH3kb29HK-vq0f7Bd9",
        finish: "",
      },
      instance_id: "e27f2921-9f44-439a-b077-f159b56d4baa",
    };

    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: expectedResp });
    }, 5);

    const signer = {
      signingKey: "signature",
      generateSignatureParams: () => "mock-sig-params",
      getSignatureInput: () => "mock-sig-input",
      sign: async () => Promise.resolve("mock-signature"),
    };
    const client = new GNAPClient({
      signer,
      GNAP_AUTH_SERVER_URL,
    });

    const result = await client.requestAccess(mockRequest);

    expect(result.status).to.be.equal(200);
    expect(result.data).to.be.equal(expectedResp);
  });
});

describe("GNAP Auth Client - Continuing Request", function () {
  beforeEach(function () {
    moxios.install();
  });

  afterEach(function () {
    moxios.uninstall();
  });

  const mockRequest = {
    interact_ref: "022ba1df-1f44-4353-8abf-be541423ba48",
  };

  it("success continuing gnap auth request", async () => {
    const expectedResp = {
      access_token: [
        {
          value: "bbb590ff-9f39-4cb1-9308-4d7fe7d3c9f9",
          label: "example-token",
          manage: "",
          access: null,
          expires_in: 3600,
          key: "",
          flags: null,
        },
      ],
      instance_id: "42dafcba-ddf3-4734-97ff-db6e7d9c85bd",
    };

    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: expectedResp });
    }, 5);

    const MOCK_CONTINUE_TOKEN = "mock-continue-token-value";

    const signer = {
      authorization: MOCK_CONTINUE_TOKEN,
      signingKey: "signature",
      generateSignatureParams: () => "mock-sig-params",
      getSignatureInput: () => "mock-sig-input",
      sign: async () => Promise.resolve("mock-signature"),
    };
    const client = new GNAPClient({
      signer,
      GNAP_AUTH_SERVER_URL,
    });

    const result = await client.continue(mockRequest, MOCK_CONTINUE_TOKEN);

    expect(result.status).to.be.equal(200);
    expect(result.data).to.be.equal(expectedResp);
  });
});
