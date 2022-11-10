/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import chai, { expect } from "chai";
import chaiAsPromised from "chai-as-promised";
import moxios from "moxios";
import sinon from "sinon";
import {
  getJSONTestData,
  loadFrameworks,
  testConfig,
  prepareTestManifest,
} from "../../common";
import {
  createWalletProfile,
  CredentialManager,
  DIDManager,
  OpenID4VP,
  UniversalWallet,
} from "@";
import { IssuerAdapter } from "../../mocks/adapters";

chai.use(chaiAsPromised);

const VC_ISSUER = "vc-issuer-agent";
const PRC_DESCRIPTOR_ID = "prc_output";
const VC_FORMAT = "ldp_vc";
const USER_ID = "agent-openid4vp-tests";
const MOCK_REQUEST_URL =
  "openid-vc://?request_uri=https://someverifierdomain.com/v1.0/verifiablecredentials/request/a0eed079-672f-4055-a4f5-e0f5d76ecdea";
const MOCK_INVALID_REQUEST_URL = "some-invalid-request-url-string";
const MOCK_JWT =
  "eyJhbGciOiJFUzI1NiIsImNydiI6IlAtMjU2Iiwia2lkIjoiZGlkOmtleTp6RG5hZWtkMndMa3ltcWhCWEs3SnJhTTN5c3lOOGNvZmtEcHJ2c3puTnNqb0hZd013IiwidHlwIjoiSldUIn0.CgkJewogICAgICAgICAgIm5vbmNlIjogIk8xbVpHbnVldCsrSWxnMmMxalI0akE9PSIsCiAgICAgICAgICAiY2xpZW50X2lkIjogImRpZDppb246RWlBdjBlSjVjQjBoR1dWSDVZYlktdXcxSzcxRXBPU1Q2enR1ZUVRelZDRWMwQTpleUprWld4MFlTSTZleUp3WVhSamFHVnpJanBiZXlKaFkzUnBiMjRpT2lKeVpYQnNZV05sSWl3aVpHOWpkVzFsYm5RaU9uc2ljSFZpYkdsalMyVjVjeUk2VzNzaWFXUWlPaUp6YVdkZlkyRmlOalZoWVRBaUxDSndkV0pzYVdOTFpYbEtkMnNpT25zaVkzSjJJam9pYzJWamNESTFObXN4SWl3aWEzUjVJam9pUlVNaUxDSjRJam9pT0cxNU1IRktVR3Q2T1ZOUlJUa3lSVGxtUkZnNFpqSjRiVFIyWDI5Wk1YZE5URXBXV2xRMVN6aFJkeUlzSW5raU9pSXhiMHhzVkc1ck56TTJSVE5IT1VOTlVUaDNXakpRU2xWQk0wcGhWblk1VnpGYVZHVkdTbUpSV1RGRkluMHNJbkIxY25CdmMyVnpJanBiSW1GMWRHaGxiblJwWTJGMGFXOXVJaXdpWVhOelpYSjBhVzl1VFdWMGFHOWtJbDBzSW5SNWNHVWlPaUpGWTJSellWTmxZM0F5TlRack1WWmxjbWxtYVdOaGRHbHZia3RsZVRJd01Ua2lmVjBzSW5ObGNuWnBZMlZ6SWpwYmV5SnBaQ0k2SW14cGJtdGxaR1J2YldGcGJuTWlMQ0p6WlhKMmFXTmxSVzVrY0c5cGJuUWlPbnNpYjNKcFoybHVjeUk2V3lKb2RIUndjem92TDNOM1pXVndjM1JoYTJWekxtUnBaQzV0YVdOeWIzTnZablF1WTI5dEx5SmRmU3dpZEhsd1pTSTZJa3hwYm10bFpFUnZiV0ZwYm5NaWZWMTlmVjBzSW5Wd1pHRjBaVU52YlcxcGRHMWxiblFpT2lKRmFVRndjbVZUTnkxRWN6aDVNREZuVXprMmNFNWlWbnBvUm1ZeFVscHZibFozVWtzd2JHOW1aSGRPWjJGQkluMHNJbk4xWm1acGVFUmhkR0VpT25zaVpHVnNkR0ZJWVhOb0lqb2lSV2xFTVdSRmRVVmxkRVJuTW5oaVZFczBVRFpWVFROdVdFTktWbkZNUkUxMU0yOUlWV05NYW10Wk1XRlRkeUlzSW5KbFkyOTJaWEo1UTI5dGJXbDBiV1Z1ZENJNklrVnBSRUZrU3pGV05rcGphMUJwWTBSQmNHRnhWMkl5WkU5NU1GUk5jbUpLVG1sbE5tbEtWems0Wms1NGJrRWlmWDAiLAogICAgICAgICAgInJlZGlyZWN0X3VyaSI6ICJodHRwczovL2JldGEuZGlkLm1zaWRlbnRpdHkuY29tL3YxLjAvZTFmNjZmMmUtYzA1MC00MzA4LTgxYjMtM2Q3ZWE3ZWYzYjFiL3ZlcmlmaWFibGVjcmVkZW50aWFscy9wcmVzZW50IiwKCQkgICJjbGFpbXMiOiB7CgkJCSJ2cF90b2tlbiI6IHsKCQkJICAiaWQiOiAiMjJjNzcxNTUtZWRmMi00ZWM1LThkNDQtYjM5M2I0ZTRmYTM4IiwKCQkJICAiaW5wdXRfZGVzY3JpcHRvcnMiOiBbCgkJCQl7CgkJCQkgICJpZCI6ICIyMGIwNzNiYi1jZWRlLTQ5MTItOWU5ZC0zMzRlNTcwMjA3N2IiLAoJCQkJICAic2NoZW1hIjogWwoJCQkJCXsKCQkJCQkgICJ1cmkiOiAiaHR0cHM6Ly93d3cudzMub3JnLzIwMTgvY3JlZGVudGlhbHMjVmVyaWZpYWJsZUNyZWRlbnRpYWwiCgkJCQkJfQoJCQkJICBdLAoJCQkJICAiY29uc3RyYWludHMiOiB7CgkJCQkJImZpZWxkcyI6IFsKCQkJCQkgIHsKCQkJCQkJInBhdGgiOiBbCgkJCQkJCSAgIiQuY3JlZGVudGlhbFN1YmplY3QuZmFtaWx5TmFtZSIKCQkJCQkJXQoJCQkJCSAgfQoJCQkJCV0KCQkJCSAgfQoJCQkJfQoJCQkgIF0KCQkJfQoJCSAgfQoJCX0KCQ.qVoI0Epd-4OWZ2I1rBZkN3l7uIu1QnNP1Ni9UnqMc_7EeYpI9QoJVwf6paxJ8lHsOUbta9cxLWvagos9A9H-2Q";
const MOCK_DID_DOC = getJSONTestData("openid4vc-mock-did-doc.json");
const MOCK_PRESENTATION_QUERY = getJSONTestData(
  "openid4vc-mock-presentation-query.json"
);
const MOCK_KID = "did:key:mockjwt#sign";

let agent, issuer, openID4VP, samplePRC;

before(async function () {
  agent = await loadFrameworks({ name: USER_ID });

  issuer = new IssuerAdapter(VC_ISSUER);
  await issuer.init();

  // load sample VC from testdata
  const prcVC = getJSONTestData("prc-vc.json");

  // issue sample credential
  const [vc] = await issuer.issue(prcVC);
  expect(vc.id).to.not.empty;
  expect(vc.credentialSubject).to.not.empty;

  samplePRC = vc;
});

after(function () {
  agent ? agent.destroy() : "";
  issuer ? issuer.destroy() : "";
});

describe("OpenID4VP - Constructor", async function () {
  it("throws an error constructing OpenID4VP instance when agent is missing", function () {
    expect(() => new OpenID4VP({ user: USER_ID })).to.throw(
      TypeError,
      "Error initializing OpenID4VP client: agent cannot be empty"
    );
  });

  it("throws an error constructing OpenID4VP instance when user is missing", function () {
    expect(() => new OpenID4VP({ agent: agent })).to.throw(
      TypeError,
      "Error initializing OpenID4VP client: user cannot be empty"
    );
  });

  it("successfully creates OpenID4VP client", function () {
    openID4VP = new OpenID4VP({ agent: agent, user: USER_ID });
    expect(openID4VP).to.be.an.instanceof(OpenID4VP);
    expect(openID4VP).to.have.property("user");
    expect(openID4VP.user).to.equal(USER_ID);
    expect(openID4VP).to.have.property("credentialManager");
    expect(openID4VP.credentialManager).to.be.an.instanceof(CredentialManager);
    expect(openID4VP).to.have.property("didManager");
    expect(openID4VP.didManager).to.be.an.instanceof(DIDManager);
  });
});

describe("OpenID4VP - Initiate Presentation", async function () {
  beforeEach(function () {
    moxios.install();
  });

  afterEach(function () {
    moxios.uninstall();
  });

  it("user creates his wallet profile", async function () {
    await createWalletProfile(agent, USER_ID, {
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
  });

  let authToken;
  it("user opens wallet", async function () {
    const wallet = new UniversalWallet({ agent: agent, user: USER_ID });
    const authResponse = await wallet.open({
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
    expect(authResponse).to.have.property("token");
    expect(authResponse.token).to.have.a.lengthOf.above(0);
    authToken = authResponse.token;
  });

  it("user saves a credential into wallet", async function () {
    const credentialManager = new CredentialManager({
      agent: agent,
      user: USER_ID,
    });

    await credentialManager.save(
      authToken,
      { credentials: [samplePRC] },
      {
        manifest: prepareTestManifest("prc-cred-manifest.json"),
        descriptorMap: [
          {
            id: PRC_DESCRIPTOR_ID,
            format: VC_FORMAT,
            path: "$[0]",
          },
        ],
      }
    );
  });

  it("user successfully initiates presentation", async function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: MOCK_JWT });
    }, 5);

    sinon
      .stub(openID4VP.didManager, "resolveWebDIDFromOrbDID")
      .callsFake(() => MOCK_DID_DOC);

    const presentationQuery = await openID4VP.initiateOIDCPresentation({
      authToken,
      url: MOCK_REQUEST_URL,
    });

    expect(presentationQuery).to.have.lengthOf(1);
    expect(presentationQuery[0].verifiableCredential).to.have.lengthOf(1);
  });

  it("throws an error when authToken parameter is missing", function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: MOCK_JWT });
    }, 5);

    expect(
      openID4VP.initiateOIDCPresentation({
        MOCK_REQUEST_URL,
      })
    ).to.eventually.throw(
      TypeError,
      "Error initiating OIDC presentation: authToken is missing"
    );
  });

  it("throws an error when url parameter is missing", function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: MOCK_JWT });
    }, 5);

    expect(
      openID4VP.initiateOIDCPresentation({
        authToken,
      })
    ).to.eventually.throw(
      TypeError,
      "Error initiating OIDC presentation: url is missing"
    );
  });

  it("throws an error when url parameter is missing", function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: MOCK_JWT });
    }, 5);

    expect(
      openID4VP.initiateOIDCPresentation({
        authToken,
        url: MOCK_INVALID_REQUEST_URL,
      })
    ).to.eventually.throw(
      TypeError,
      "Error initiating OIDC presentation: invalid request url: request_uri is missing"
    );
  });

  it("throws an error when request_uri is missing in the url", function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: MOCK_JWT });
    }, 5);

    expect(
      openID4VP.initiateOIDCPresentation({
        authToken,
        url: MOCK_INVALID_REQUEST_URL,
      })
    ).to.eventually.throw(
      TypeError,
      "Error initiating OIDC presentation: invalid request url: request_uri is missing"
    );
  });
});

describe("OpenID4VP - Submit Presentation", async function () {
  it("throws an error when kid parameter is missing", function () {
    expect(
      openID4VP.submitOIDCPresentation({
        presentationQuery: MOCK_PRESENTATION_QUERY,
        expiry: new Date().getTime() / 1000 + 60 * 10,
      })
    ).to.eventually.throw(
      TypeError,
      "Error submitting OpenID4VP presentation: kid cannot be empty"
    );
  });
  it("throws an error when presentationQuery parameter is missing", function () {
    expect(
      openID4VP.submitOIDCPresentation({
        kid: MOCK_KID,
        expiry: new Date().getTime() / 1000 + 60 * 10,
      })
    ).to.eventually.throw(
      TypeError,
      "Error submitting OpenID4VP presentation: kid cannot be empty"
    );
  });
  it("throws an error when expiry parameter is missing", function () {
    expect(
      openID4VP.submitOIDCPresentation({
        kid: MOCK_KID,
        presentationQuery: MOCK_PRESENTATION_QUERY,
      })
    ).to.eventually.throw(
      TypeError,
      "Error submitting OpenID4VP presentation: expiry cannot be empty"
    );
  });
  it("successfuly submits oidc presentation", async function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: "" });
    }, 5);

    expect(
      openID4VP.submitOIDCPresentation({
        kid: MOCK_KID,
        presentationQuery: MOCK_PRESENTATION_QUERY,
        expiry: new Date().getTime() / 1000 + 60 * 10,
      })
    ).to.eventually.not.throw();
  });
});
