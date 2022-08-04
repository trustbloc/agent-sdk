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
  walletAgent = await loadFrameworks({ name: WALLET_USER, onlyWallet: true });

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
