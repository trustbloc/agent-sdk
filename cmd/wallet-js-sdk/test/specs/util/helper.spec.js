/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { expect } from "chai";
import { v4 as uuidv4 } from "uuid";
import jp from "jsonpath";
import {
  ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
  ATTACH_FORMAT_CREDENTIAL_MANIFEST,
  getJSONTestData,
  loadFrameworks,
  testConfig,
} from "../common";
import {
  CredentialManager,
  extractOOBGoalCode,
  findAttachmentByFormatV2,
  normalizePresentationSubmission,
  updatePresentationSubmission,
  createKeyPair,
  WalletUser,
} from "../../../src";
import { IssuerAdapter } from "../mocks/adapters";

const WALLET_USER = "max-agent";
const WALLET_QUERY_USER = "max-query-agent";
const WALLET_DIDCOMM_USER = "max-didcomm-agent";
const VC_ISSUER = "vc-issuer-agent-01";

let walletUserAgent,
  issuer,
  sampleUDC,
  samplePRC,
  sampleUDCBBS,
  sampleFrameDoc,
  manifest;

before(async function () {
  this.timeout(0);
  walletUserAgent = await loadFrameworks({
    name: WALLET_USER,
    enableDIDComm: true,
  });

  issuer = new IssuerAdapter(VC_ISSUER);
  await issuer.init();

  // load sample VCs from testdata
  let udcVC = getJSONTestData("udc-vc.json");
  let prcVC = getJSONTestData("prc-vc.json");
  sampleUDCBBS = getJSONTestData("udc-bbs-vc.json");
  sampleFrameDoc = getJSONTestData("udc-frame.json");
  manifest = getJSONTestData("manifest-vc.json");

  // issue sample credentials
  let [vc1, vc2] = await issuer.issue(udcVC, prcVC);
  expect(vc1.id).to.not.empty;
  expect(vc1.credentialSubject).to.not.empty;
  expect(vc2.id).to.not.empty;
  expect(vc2.credentialSubject).to.not.empty;

  sampleUDC = vc1;
  samplePRC = vc2;
});

after(function () {
  walletUserAgent ? walletUserAgent.destroy() : "";
  issuer ? issuer.destroy() : "";
});

describe.skip("Presentation Submission Normalization Tests", async function () {
  let auth;
  it("user setups wallet", async function () {
    let walletUser = new WalletUser({
      agent: walletUserAgent,
      user: WALLET_QUERY_USER,
    });
    // create wallet profile
    await walletUser.createWalletProfile({
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });

    // unlock wallet
    let authResponse = await walletUser.unlock({
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
    expect(authResponse.token).to.not.empty;
    auth = authResponse.token;

    // save sample credentials
    const manifest = getJSONTestData("allvcs-cred-manifest.json");
    const descriptorMap = [
      {
        id: "udc_output",
        format: "ldp_vc",
        path: "$[0]",
      },
      {
        id: "prc_output",
        format: "ldp_vc",
        path: "$[1]",
      },
      {
        id: "udc_output",
        format: "ldp_vc",
        path: "$[2]",
      },
    ];

    let credentialManager = new CredentialManager({
      agent: walletUserAgent,
      user: WALLET_QUERY_USER,
    });
    await credentialManager.save(
      auth,
      { credentials: [sampleUDC, samplePRC, sampleUDCBBS] },
      { manifest, descriptorMap }
    );

    let { contents } = await credentialManager.getAll(auth);
    expect(contents).to.not.empty;
    expect(Object.keys(contents)).to.have.lengthOf(3);
  });

  let presentation;
  it("user performs presentation exchange query and normalizes multi credential results", async function () {
    let credentialManager = new CredentialManager({
      agent: walletUserAgent,
      user: WALLET_QUERY_USER,
    });

    const query = [
      {
        type: "PresentationExchange",
        credentialQuery: [
          {
            id: "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
            input_descriptors: [
              {
                id: "20b073bb-cede-4912-9e9d-334e5702077b",
                name: "Need All VCs",
                purpose: "Need All your W3 credentials.",
                schema: [
                  {
                    uri: "https://www.w3.org/2018/credentials#VerifiableCredential",
                  },
                ],
              },
            ],
          },
        ],
      },
    ];

    let { results } = await credentialManager.query(auth, query);
    expect(results).to.have.lengthOf(1);
    presentation = results[0];

    const normalized = normalizePresentationSubmission(query, results[0]);
    expect(normalized).to.have.lengthOf(1);
    expect(normalized[0].id).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[0].id
    );
    expect(normalized[0].name).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[0].name
    );
    expect(normalized[0].purpose).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[0].purpose
    );
    expect(normalized[0].format).to.not.empty;
    expect(normalized[0].credentials).to.have.lengthOf(3);
  });

  it("user selects only one credential and updates presentation", async function () {
    let updated = updatePresentationSubmission(presentation, {
      "20b073bb-cede-4912-9e9d-334e5702077b": samplePRC.id,
    });

    expect(updated).to.not.empty;
    expect(updated.presentation_submission.definition_id).to.be.equal(
      presentation.presentation_submission.definition_id
    );
    expect(updated.presentation_submission.descriptor_map).to.have.lengthOf(1);
    expect(updated.presentation_submission.descriptor_map[0].id).to.be.equal(
      "20b073bb-cede-4912-9e9d-334e5702077b"
    );
    expect(
      updated.presentation_submission.descriptor_map[0].format
    ).to.be.equal("ldp_vp");
    expect(updated.presentation_submission.descriptor_map[0].path).to.be.equal(
      "$.verifiableCredential[0]"
    );
  });

  const id1 = uuidv4(),
    id2 = uuidv4(),
    id3 = uuidv4();
  it("user performs presentation exchange query and normalizes multi credential results - multiple input descriptors", async function () {
    let credentialManager = new CredentialManager({
      agent: walletUserAgent,
      user: WALLET_QUERY_USER,
    });

    const query = [
      {
        type: "PresentationExchange",
        credentialQuery: [
          {
            id: "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
            input_descriptors: [
              {
                id: id1,
                name: "Need All VCs - 1",
                purpose: "Need All your W3 credentials - 1.",
                schema: [
                  {
                    uri: "https://www.w3.org/2018/credentials#VerifiableCredential",
                  },
                ],
              },
              {
                id: id2,
                name: "Family name VCs",
                purpose: "Need All your W3 credentials with family name.",
                schema: [
                  {
                    uri: "https://www.w3.org/2018/credentials#VerifiableCredential",
                  },
                ],
                constraints: {
                  fields: [{ path: ["$.credentialSubject.familyName"] }],
                },
              },
              {
                id: id3,
                name: "Need All VCs - 2",
                purpose: "Need All your W3 credentials -2.",
                schema: [
                  {
                    uri: "https://www.w3.org/2018/credentials#VerifiableCredential",
                  },
                ],
              },
            ],
          },
        ],
      },
    ];

    let { results } = await credentialManager.query(auth, query);
    expect(results).to.have.lengthOf(1);
    presentation = results[0];

    const normalized = normalizePresentationSubmission(query, results[0]);
    expect(normalized).to.have.lengthOf(3);

    const normalized1 = jp.query(normalized, `$[?(@.id=="${id1}")]`);
    expect(normalized1).to.have.lengthOf(1);
    expect(normalized1[0].id).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[0].id
    );
    expect(normalized1[0].name).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[0].name
    );
    expect(normalized1[0].purpose).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[0].purpose
    );
    expect(normalized1[0].credentials).to.have.lengthOf(3);

    const normalized2 = jp.query(normalized, `$[?(@.id=="${id2}")]`);
    expect(normalized2).to.have.lengthOf(1);
    expect(normalized2[0].id).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[1].id
    );
    expect(normalized2[0].name).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[1].name
    );
    expect(normalized2[0].purpose).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[1].purpose
    );
    expect(normalized2[0].credentials).to.have.lengthOf(1);

    const normalized3 = jp.query(normalized, `$[?(@.id=="${id3}")]`);
    expect(normalized3).to.have.lengthOf(1);
    expect(normalized3[0].id).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[2].id
    );
    expect(normalized3[0].name).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[2].name
    );
    expect(normalized3[0].purpose).to.be.equal(
      query[0].credentialQuery[0].input_descriptors[2].purpose
    );
    expect(normalized3[0].credentials).to.have.lengthOf(3);
  });

  it("user selects one credential for each descriptor ID and updates presentation", async function () {
    let updates = {};
    updates[id1] = sampleUDC.id;
    updates[id2] = samplePRC.id;
    updates[id3] = sampleUDCBBS.id;

    let updated = updatePresentationSubmission(presentation, updates);

    expect(updated).to.not.empty;
    expect(updated.presentation_submission.definition_id).to.be.equal(
      presentation.presentation_submission.definition_id
    );
    expect(updated.presentation_submission.descriptor_map).to.have.lengthOf(3);

    updated.presentation_submission.descriptor_map.forEach((descrMap) => {
      const vc = jp.query(updated, descrMap.path);
      expect(vc).to.have.lengthOf(1);

      switch (descrMap.id) {
        case id1:
          expect(vc[0].id).to.be.equal(sampleUDC.id);
          break;
        case id2:
          expect(vc[0].id).to.be.equal(samplePRC.id);
          break;
        case id3:
          expect(vc[0].id).to.be.equal(sampleUDCBBS.id);
          break;
        default:
          expect.fail("invalid descriptor ID");
      }
    });
  });

  it("user performs query by example and normalizes multi credential results", async function () {
    let credentialManager = new CredentialManager({
      agent: walletUserAgent,
      user: WALLET_QUERY_USER,
    });

    const query = [
      {
        type: "QueryByExample",
        credentialQuery: [
          {
            reason: "Please present your valid degree certificate.",
            example: {
              "@context": [
                "https://www.w3.org/2018/credentials/v1",
                "https://www.w3.org/2018/credentials/examples/v1",
              ],
              type: ["UniversityDegreeCredential"],
              trustedIssuer: [
                { issuer: "urn:some:required:issuer" },
                {
                  required: true,
                  issuer: "did:example:76e12ec712ebc6f1c221ebfeb1f",
                },
              ],
              credentialSubject: {
                id: "did:example:ebfeb1f712ebc6f1c276e12ec21",
              },
            },
          },
        ],
      },
    ];

    let { results } = await credentialManager.query(auth, query);
    expect(results).to.have.lengthOf(1);

    const normalized = normalizePresentationSubmission(query, results[0]);
    expect(normalized).to.have.lengthOf(2);

    normalized.forEach((result) => {
      expect(result.id).to.not.empty;
      expect(result.name).to.be.undefined;
      expect(result.purpose).to.be.undefined;
      expect(result.credentials).to.have.lengthOf(1);
    });

    let updated = updatePresentationSubmission(results[0]);
    expect(updated).to.not.empty;
    expect(updated.verifiableCredential).to.have.lengthOf(2);
  });
});

describe("Testing generic util functions", async function () {
  it("test extracting goal code from invitation", async function () {
    expect(
      extractOOBGoalCode({
        "@id": "6c6fbda3-60f4-4821-ab9f-dcbda4b94faa",
        "@type": "https://didcomm.org/out-of-band/1.0/invitation",
        accept: ["didcomm/aip2;env=rfc19"],
        goal_code: "streamlined-vc",
        handshake_protocols: ["https://didcomm.org/didexchange/1.0"],
        label: "vc-issuer-agent",
        services: [
          {
            id: "b10028a4-b71a-4113-a932-d3893e6cdfe9",
            recipientKeys: [
              "did:key:z6MkmiqyKcWh9Ee8iE3DPjoCHiZmdnP4armLD9Wjmm7PFZwP",
            ],
            routingKeys: [
              "did:key:z6MkmyYKpBNAtGK2BuTdDZNZajnJiSYtVDYiX21kanLr7uor",
            ],
            serviceEndpoint: "https://hub-router.trustbloc.local:10091",
            type: "did-communication",
          },
        ],
      })
    ).to.be.equal("streamlined-vc");

    expect(
      extractOOBGoalCode({
        "@id": "6c6fbda3-60f4-4821-ab9f-dcbda4b94faa",
        "@type": "https://didcomm.org/out-of-band/1.0/invitation",
        accept: ["didcomm/aip2;env=rfc19"],
        handshake_protocols: ["https://didcomm.org/didexchange/1.0"],
        label: "vc-issuer-agent",
        services: [
          {
            id: "b10028a4-b71a-4113-a932-d3893e6cdfe9",
            recipientKeys: [
              "did:key:z6MkmiqyKcWh9Ee8iE3DPjoCHiZmdnP4armLD9Wjmm7PFZwP",
            ],
            routingKeys: [
              "did:key:z6MkmyYKpBNAtGK2BuTdDZNZajnJiSYtVDYiX21kanLr7uor",
            ],
            serviceEndpoint: "https://hub-router.trustbloc.local:10091",
            type: "did-communication",
          },
        ],
      })
    ).to.be.undefined;

    expect(
      extractOOBGoalCode({
        type: "https://didcomm.org/out-of-band/2.0/invitation",
        id: "f137e0db-db7b-4776-9530-83c808a34a42",
        from: "did:example:issuer",
        body: {
          goal_code: "streamlined-vc",
          accept: ["didcomm/v2"],
        },
      })
    ).to.be.equal("streamlined-vc");

    expect(
      extractOOBGoalCode({
        type: "https://didcomm.org/out-of-band/2.0/invitation",
        id: "599f3638-b563-4937-9487-dfe55099d900",
        from: "did:example:verifier",
        body: {
          goal_code: "streamlined-vp",
          accept: ["didcomm/v2"],
        },
      })
    ).to.be.equal("streamlined-vp");

    expect(
      extractOOBGoalCode({
        type: "https://didcomm.org/out-of-band/2.0/invitation",
        id: "f137e0db-db7b-4776-9530-83c808a34a42",
        from: "did:example:issuer",
      })
    ).to.be.null;

    expect(
      extractOOBGoalCode({
        type: "https://didcomm.org/out-of-band/2.0/invitation",
        id: "f137e0db-db7b-4776-9530-83c808a34a42",
        from: "did:example:issuer",
        body: {
          accept: ["didcomm/v2"],
        },
      })
    ).to.be.null;

    expect(
      extractOOBGoalCode({
        type: "https://didcomm.org/out-of-band/2.0/invitation",
        id: "f137e0db-db7b-4776-9530-83c808a34a42",
        from: "did:example:issuer",
        body: {
          goal_code: "",
          accept: ["didcomm/v2"],
        },
      })
    ).to.be.null;

    expect(
      extractOOBGoalCode({
        type: "https://didcomm.org/out-of-band/2.0/invitation",
        id: "f137e0db-db7b-4776-9530-83c808a34a42",
        from: "did:example:issuer",
        body: {
          goal_code: "",
          accept: ["didcomm/v2"],
        },
      })
    ).to.be.null;

    const manifest = getJSONTestData("udc-cred-manifest.json");
    const fulfillment = getJSONTestData("cred-fulfillment-udc-vp.json");

    let attachments = [];

    attachments.push({
      id: uuidv4(),
      media_type: "application/json",
      format: ATTACH_FORMAT_CREDENTIAL_MANIFEST,
      data: {
        json: {
          credential_manifest: manifest,
        },
      },
    });

    attachments.push({
      id: uuidv4(),
      media_type: "application/json",
      format: ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
      data: {
        json: fulfillment,
      },
    });

    expect(
      findAttachmentByFormatV2(attachments, ATTACH_FORMAT_CREDENTIAL_MANIFEST)
        .credential_manifest
    ).to.be.equal(manifest);
    expect(
      findAttachmentByFormatV2(
        attachments,
        ATTACH_FORMAT_CREDENTIAL_FULFILLMENT
      )
    ).to.be.equal(fulfillment);
  });
  it("test creating crypto key pair", async function () {
    const keyPair = await createKeyPair("ECDSA", "P-256", true, [
      "sign",
      "verify",
    ]);
    expect(keyPair.publicKey).to.be.exist;
    expect(keyPair.privateKey).to.be.exist;
  });
});
