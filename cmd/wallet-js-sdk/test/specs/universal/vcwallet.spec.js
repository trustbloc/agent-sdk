/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import chai, { expect } from "chai";
import chaiAsPromised from "chai-as-promised";
import { v4 as uuidv4 } from "uuid";

import { getJSONTestData, loadFrameworks, testConfig } from "../common";
import { VerifierAdapter } from "../mocks/adapters";
import {
  connectToMediator,
  contentTypes,
  createWalletProfile,
  DIDManager,
  getMediatorConnections,
  UniversalWallet,
} from "../../../src";

chai.use(chaiAsPromised);

const WALLET_USER = "wallet-lite-user";
const RELYING_PARTY = "relying-party-01";
const keyType = "ED25519";
const signatureType = "Ed25519VerificationKey2018";

let walletAgent, rp, sampleMetadata;


before(async function () {
  walletAgent = await loadFrameworks({
    name: WALLET_USER,
    contextProviderURL: ["http://localhost:10096/agent-startup-contexts.json"]
  });

  rp = new VerifierAdapter(RELYING_PARTY);
  await rp.init();

  sampleMetadata = getJSONTestData("wallet-metadata.json");
});

after(function () {
  walletAgent ? walletAgent.destroy() : "";
  rp ? rp.destroy() : "";
});

describe("Lite wallet tests", async function () {
  it("wallet user creates wallet profile", async function () {
    await createWalletProfile(walletAgent, WALLET_USER, {
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
  });

  let auth;
  it("wallet user opens wallet", async function () {
    let wallet = new UniversalWallet({ agent: walletAgent, user: WALLET_USER });
    let authResponse = await wallet.open({
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
    expect(authResponse.token).to.not.empty;
    auth = authResponse.token;
  });

  it("wallet user adds contents to wallet", async function () {
    let wallet = new UniversalWallet({ agent: walletAgent, user: WALLET_USER });

    // save sample metadata.
    await wallet.add({
      auth,
      contentType: contentTypes.METADATA,
      content: sampleMetadata,
    });

    // resolve and save a DID.
    let content = await walletAgent.vdr.resolveDID({
      id: "did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5",
    });
    await wallet.add({
      auth,
      contentType: contentTypes.DID_RESOLUTION_RESPONSE,
      content,
    });
  });

  it("wallet user adds, removes, gets, gets all contents from wallet", async function () {
    let wallet = new UniversalWallet({ agent: walletAgent, user: WALLET_USER });

    let ids = [uuidv4(), uuidv4(), uuidv4(), uuidv4(), uuidv4()];

    // add few sample data
    let addMetadata = async (id) => {
      await wallet.add({
        auth,
        contentType: contentTypes.METADATA,
        content: {
          "@context": ["https://w3id.org/wallet/v1"],
          id: id,
          type: "Person",
          name: "John Smith",
        },
      });
    };

    for (let id of ids) {
      await addMetadata(id);
    }

    // get by id
    let getMetadata = async (id) => {
      let content = await wallet.get({
        auth,
        contentType: contentTypes.METADATA,
        contentID: id,
      });
      expect(content).to.not.empty;
    };
    for (let id of ids) {
      await getMetadata(id);
    }

    // remove one
    await wallet.remove({
      auth,
      contentType: contentTypes.METADATA,
      contentID: ids[0],
    });

    // get all
    let all = await wallet.getAll({ auth, contentType: contentTypes.METADATA });
    expect(Object.keys(all.contents)).to.have.lengthOf(ids.length);
  });

  it("wallet user creates a key pair inside wallet", async function () {
    let wallet = new UniversalWallet({ agent: walletAgent, user: WALLET_USER });

    let keyPair = await wallet.createKeyPair(auth, { keyType: "ED25519" });
    expect(keyPair.keyID).to.not.empty;
    expect(keyPair.publicKey).to.not.empty;
  });

  it("wallet user closes wallet", async function () {
    let wallet = new UniversalWallet({ agent: walletAgent, user: WALLET_USER });

    expect((await wallet.close()).closed).to.be.true;
    expect((await wallet.close()).closed).to.be.false;

    // any operation should fail
    expect(wallet.getAll({ auth, contentType: contentTypes.METADATA })).to
      .eventually.be.rejected;
  });
});

describe.only("Lite wallet tests", async function () {
  it("wallet user creates wallet profile", async function () {
    await createWalletProfile(walletAgent, WALLET_USER, {
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
  });

  let auth;
  let wallet;

  it("wallet user opens wallet", async function () {
    wallet = new UniversalWallet({ agent: walletAgent, user: WALLET_USER });
    let authResponse = await wallet.open({
      localKMSPassphrase: testConfig.walletUserPassphrase,
    });
    expect(authResponse.token).to.not.empty;
    auth = authResponse.token;
  });


  let publicDID
  it("wallet user creates an Orb DID", async function () {
    let didManager = new DIDManager({agent: walletAgent, user: WALLET_USER})
    let didResolution = await didManager.createOrbDID(auth, {
      keyType: "ED25519",
      purposes: ["assertionMethod", "authentication"]
    })
    expect(didResolution).to.not.empty

    publicDID = didResolution.didDocument
  });

  let cred
  it("wallet user issues self-issued credential", async function () {
    let templateCred = {
      "@context": [
        "https://www.w3.org/2018/credentials/v1",
        "https://www.w3.org/2018/credentials/examples/v1"
      ],
      "credentialSchema": [{
        "id": "https://www.w3.org/2018/credentials/v1",
        "type": "JsonSchemaValidator2018"
      }],
      "credentialSubject": {
        "degree": {
          "type": "BachelorDegree",
          "university": "MIT"
        },
        "id": "did:example:ebfeb1f712ebc6f1c276e12ec21",
        "name": "Jayden Doe",
        "spouse": "did:example:c276e12ec21ebfeb1f712ebc6f1"
      },
      "expirationDate": "2020-01-01T19:23:24Z",
      "id": "http://example.edu/credentials/1872",
      "issuanceDate": "2010-01-01T19:23:24Z",
      "issuer": {
        "id": publicDID.id,
        "name": "Example University"
      },
      "referenceNumber": 83294847,
      "type": [
        "VerifiableCredential",
        "UniversityDegreeCredential"
      ]
    };

    cred = await wallet.issue(auth, templateCred, {
      controller: publicDID.id,
      proofFormat: "ExternalJWTProofFormat",
    });
    expect(cred).to.not.empty;
    expect(cred.credential).to.not.empty;
  });

  it("wallet user verifies credential", async function () {
    let res = await wallet.verify(auth, {rawCredential: cred.credential});
    expect(res) .to.not.empty;
    expect(res.error).to.undefined;
    expect(res.verified).to.eq(true);
  });

  // TODO: query by pEx

  it("wallet user stores credential", async function () {
    let res = await wallet.add({auth, contentType: contentTypes.CREDENTIAL, content: cred.credential});
    expect(res).to.empty;
  });

  it( "wallet user queries by example", async function () {
    let queryByEx = {
      "reason": "Please present your identity document.",
      "example": {
        "@context": [
          "https://www.w3.org/2018/credentials/v1",
          "https://www.w3.org/2018/credentials/examples/v1"
        ],
        "type": ["UniversityDegreeCredential"],
        "trustedIssuer": [
          {
            "required": true,
            "issuer": publicDID.id
          }
        ],
        "credentialSubject": {
          "id": "did:example:ebfeb1f712ebc6f1c276e12ec21"
        },
        "credentialSchema": {
          "id": "https://www.w3.org/2018/credentials/v1",
          "type": "JsonSchemaValidator2018"
        }
      }
    }

    let res = await wallet.query(auth, [{ "type": "QueryByExample", "credentialQuery": [queryByEx]}]);
    expect(res).to.not.empty;
    expect(res.results).to.not.undefined;
    expect(res.results).to.not.empty;
  });

  it( "wallet user queries by presentation definition", async function () {
    let presDef = {
      "id": "ec2f83c5-eac4-4d04-b41e-6636d6670a2e",
      "input_descriptors": [{
        "id": "105b1d58-71f8-4d1e-be71-36c9c6f600c9",
        "constraints": {
          "fields": [
            {
              "path": [
                "$.issuer.id"
              ],
              "filter": {
                "type": "string",
                "const": publicDID.id
              }
            }
          ]
        },
        "format":{"jwk":{"alg": "EdDSA"}}
      }]
    }

    let res = await wallet.query(auth, [{ "type": "PresentationExchange", "credentialQuery": [presDef]}]);
    expect(res).to.not.empty;
    expect(res.results).to.not.undefined;
    expect(res.results).to.not.empty;
  });

  let presentation
  it("wallet user creates presentation", async function () {
    presentation = await wallet.prove(auth, {
        rawCredentials: [cred.credential]
      },
      {
        controller: publicDID.id,
        proofFormat: "ExternalJWTProofFormat"
      });
    expect(presentation) .to.not.empty;
    expect(presentation.presentation).to.not.undefined;
  });

  it("wallet user verifies presentation", async function () {
    let res = await wallet.verify(auth, {presentation: presentation.presentation});
    expect(res) .to.not.empty;
    expect(res.error).to.undefined;
    expect(res.verified).to.eq(true);
  });

  it("wallet user closes wallet", async function () {
    expect((await wallet.close()).closed).to.be.true;
    expect((await wallet.close()).closed).to.be.false;

    // any operation should fail
    expect(wallet.getAll({ auth, contentType: contentTypes.METADATA })).to
      .eventually.be.rejected;
  });
});
