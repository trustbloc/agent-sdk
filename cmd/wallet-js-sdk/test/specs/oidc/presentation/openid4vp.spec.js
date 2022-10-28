/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { expect } from "chai";
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
  OpenID4VP,
  UniversalWallet,
} from "@";
import { IssuerAdapter } from "../../mocks/adapters";

const VC_ISSUER = "vc-issuer-agent";
const PRC_DESCRIPTOR_ID = "prc_output";
const VC_FORMAT = "ldp_vc";
const USER_ID = "agent";
const MOCK_REQUEST_URL =
  "openid-vc://?request_uri=https://someverifierdomain.com/v1.0/verifiablecredentials/request/a0eed079-672f-4055-a4f5-e0f5d76ecdea";
const MOCK_JWT =
  "eyJhbGciOiJFUzI1NiIsImNydiI6IlAtMjU2Iiwia2lkIjoiZGlkOmtleTp6RG5hZVMzTnA3TUdWSzNodDU5UzZWNUFKU3B1YXI0cWQ5Z1IzVDVUUE5Sb2pBQzR1IiwidHlwIjoiSldUIn0.CgkJewoJCSAgImNsYWltcyI6IHsKCQkJInZwX3Rva2VuIjogewoJCQkgICJpZCI6ICIyMmM3NzE1NS1lZGYyLTRlYzUtOGQ0NC1iMzkzYjRlNGZhMzgiLAoJCQkgICJpbnB1dF9kZXNjcmlwdG9ycyI6IFsKCQkJCXsKCQkJCSAgImlkIjogIjIwYjA3M2JiLWNlZGUtNDkxMi05ZTlkLTMzNGU1NzAyMDc3YiIsCgkJCQkgICJzY2hlbWEiOiBbCgkJCQkJewoJCQkJCSAgInVyaSI6ICJodHRwczovL3d3dy53My5vcmcvMjAxOC9jcmVkZW50aWFscyNWZXJpZmlhYmxlQ3JlZGVudGlhbCIKCQkJCQl9CgkJCQkgIF0sCgkJCQkgICJjb25zdHJhaW50cyI6IHsKCQkJCQkiZmllbGRzIjogWwoJCQkJCSAgewoJCQkJCQkicGF0aCI6IFsKCQkJCQkJICAiJC5jcmVkZW50aWFsU3ViamVjdC5mYW1pbHlOYW1lIgoJCQkJCQldCgkJCQkJICB9CgkJCQkJXQoJCQkJICB9CgkJCQl9CgkJCSAgXQoJCQl9CgkJICB9CgkJfQoJ.o6ldJYTHMbmr7xhqaI-CznY8k9yNjQT7Rhzd4Ns_iBYg6Maj4qhiN7mKEFS9drWbj9v_bN65G44c0qyYcpYJEA";
const MOCK_DID_DOC = getJSONTestData("openid4vc-mock-did-doc.json");

let agent, issuer, samplePRC;

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
});

describe("OpenID4VP tests", async function () {
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

  let collection;
  it("user saves a credential into wallet", async function () {
    const credentialManager = new CredentialManager({
      agent: agent,
      user: USER_ID,
    });

    const collectionManager = new CollectionManager({
      agent: agent,
      user: USER_ID,
    });
    collection = await collectionManager.create(authToken, {
      name: "sample-name",
      description: "sample description",
    });
    expect(collection).to.not.empty;

    await credentialManager.save(
      authToken,
      { credentials: [samplePRC] },
      {
        verify: true,
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

  let openID4VP;
  it("successfully creates OpenID4VP client", function () {
    openID4VP = new OpenID4VP({ agent: agent, user: USER_ID });
    expect(openID4VP).to.be.an.instanceof(Object);
  });

  it("successfully initiates presentation", async function () {
    moxios.wait(() => {
      const request = moxios.requests.mostRecent();
      request.respondWith({ status: 200, response: MOCK_JWT });
    }, 5);

    sinon
      .stub(openID4VP.didManager, "resolveWebDIDFromOrbDID")
      .callsFake(() => MOCK_DID_DOC);

    const presentationQuery = await openID4VP.initiateOIDCPresentation(
      MOCK_REQUEST_URL,
      authToken
    );

    expect(presentationQuery).to.have.lengthOf(1);
    expect(presentationQuery[0].verifiableCredential).to.have.lengthOf(1);
  });
});
