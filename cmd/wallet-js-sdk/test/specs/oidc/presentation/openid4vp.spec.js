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
  CollectionManager,
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
  "eyJhbGciOiJFUzI1NiIsImNydiI6IlAtMjU2Iiwia2lkIjoiZGlkOmtleTp6RG5hZVMzTnA3TUdWSzNodDU5UzZWNUFKU3B1YXI0cWQ5Z1IzVDVUUE5Sb2pBQzR1IiwidHlwIjoiSldUIn0.CgkJewoJCSAgImNsYWltcyI6IHsKCQkJInZwX3Rva2VuIjogewoJCQkgICJpZCI6ICIyMmM3NzE1NS1lZGYyLTRlYzUtOGQ0NC1iMzkzYjRlNGZhMzgiLAoJCQkgICJpbnB1dF9kZXNjcmlwdG9ycyI6IFsKCQkJCXsKCQkJCSAgImlkIjogIjIwYjA3M2JiLWNlZGUtNDkxMi05ZTlkLTMzNGU1NzAyMDc3YiIsCgkJCQkgICJzY2hlbWEiOiBbCgkJCQkJewoJCQkJCSAgInVyaSI6ICJodHRwczovL3d3dy53My5vcmcvMjAxOC9jcmVkZW50aWFscyNWZXJpZmlhYmxlQ3JlZGVudGlhbCIKCQkJCQl9CgkJCQkgIF0sCgkJCQkgICJjb25zdHJhaW50cyI6IHsKCQkJCQkiZmllbGRzIjogWwoJCQkJCSAgewoJCQkJCQkicGF0aCI6IFsKCQkJCQkJICAiJC5jcmVkZW50aWFsU3ViamVjdC5mYW1pbHlOYW1lIgoJCQkJCQldCgkJCQkJICB9CgkJCQkJXQoJCQkJICB9CgkJCQl9CgkJCSAgXQoJCQl9CgkJICB9CgkJfQoJ.o6ldJYTHMbmr7xhqaI-CznY8k9yNjQT7Rhzd4Ns_iBYg6Maj4qhiN7mKEFS9drWbj9v_bN65G44c0qyYcpYJEA";
const MOCK_DID_DOC = getJSONTestData("openid4vc-mock-did-doc.json");

let agent, issuer, openID4VP, samplePRC;

before(async function () {
  agent = await loadFrameworks({ name: USER_ID });

  issuer = new IssuerAdapter(VC_ISSUER);
  await issuer.init();

  // load sample VC from testdata
  const prcVC = getJSONTestData("prc-vc.json");
  console.log("prcVC", prcVC);

  // issue sample credential
  const [vc] = await issuer.issue(prcVC);
  console.log("vc", vc);
  expect(vc.id).to.not.empty;
  expect(vc.credentialSubject).to.not.empty;

  samplePRC = vc;

  openID4VP = new OpenID4VP({ agent: agent, user: USER_ID });

  sinon
    .stub(openID4VP.didManager, "resolveWebDIDFromOrbDID")
    .callsFake(() => MOCK_DID_DOC);
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
    const openID4VP = new OpenID4VP({ agent: agent, user: USER_ID });
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
