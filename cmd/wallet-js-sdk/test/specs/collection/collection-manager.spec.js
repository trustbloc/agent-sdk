/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import chai, {expect} from "chai";
import chaiAsPromised from 'chai-as-promised';

import {getJSONTestData, loadFrameworks, prepareTestManifest, testConfig} from "../common";
import {CollectionManager, contentTypes, CredentialManager, UniversalWallet, WalletUser} from "../../../src";

var uuid = require('uuid/v4')

const WALLET_USER = 'smith-collection-agent'
const UDC_DESCRIPTOR_ID = "udc_output"
const PRC_DESCRIPTOR_ID = "prc_output"
const VC_FORMAT = "ldp_vc"

let walletUserAgent, sampleUDC, samplePRC, sampleUDCBBS

chai.use(chaiAsPromised);

before(async function () {
    walletUserAgent = await loadFrameworks({name: WALLET_USER})

    // load sample VCs from testdata.
    sampleUDC = getJSONTestData('udc-vc.json')
    samplePRC = getJSONTestData('prc-vc.json')
    sampleUDCBBS = getJSONTestData('udc-bbs-vc.json')
});

after(function () {
    walletUserAgent ? walletUserAgent.destroy() : ''
});

describe('Collection Manager Tests', async function () {
    it('user creates his wallet profile', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_USER})

        await walletUser.createWalletProfile({localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('user opens his wallet', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_USER})

        let authResponse = await walletUser.unlock({localKMSPassphrase: testConfig.walletUserPassphrase})

        expect(authResponse.token).to.not.empty

        auth = authResponse.token
    })

    const name = "my collection", description = "my description"

    let previous
    it('user creates a collection in wallet', async function () {
        let collectionManager = new CollectionManager({agent: walletUserAgent, user: WALLET_USER})

        previous = await collectionManager.create(auth, {name, description})
        expect(previous).to.not.empty
    })

    it('user gets a collection from wallet', async function () {
        let collectionManager = new CollectionManager({agent: walletUserAgent, user: WALLET_USER})

        let {content} = await collectionManager.get(auth, previous)

        expect(content).to.not.empty
        expect(content.id).to.be.equal(previous)
        expect(name).to.be.equal(content.name)
        expect(description).to.be.equal(content.description)
        expect("Vault").to.be.equal(content.type)
    })

    it('user adds credentials to the collection', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        const manifest = getJSONTestData('allvcs-cred-manifest.json')
        const descriptorMap = [
            {
                "id": "udc_output",
                "format": "ldp_vc",
                "path": "$[0]"
            },
            {
                "id": "prc_output",
                "format": "ldp_vc",
                "path": "$[1]"
            },
            {
                "id": "udc_output",
                "format": "ldp_vc",
                "path": "$[2]"
            }
        ]

        await credentialManager.save(auth, {credentials: [sampleUDC, samplePRC, sampleUDCBBS]},
            { manifest, descriptorMap, collection:previous})
    })

    it('user lists credentials under a collection', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let metadataList = await credentialManager.getAllCredentialMetadata(auth, {collectionID: previous})
        expect(metadataList).to.have.lengthOf(3)

        // get raw credential from store
        let {contents} = await credentialManager.getAll(auth, {collectionID: previous})
        expect(Object.keys(contents)).to.have.lengthOf(3)
    })

    it('user updates a collection in wallet', async function () {
        let collectionManager = new CollectionManager({agent: walletUserAgent, user: WALLET_USER})
        const newName = "New Name"

        await collectionManager.update(auth, previous, {name: newName})

        let {content} = await collectionManager.get(auth, previous)

        expect(content).to.not.empty
        expect(content.id).to.be.equal(previous)
        expect(newName).to.be.equal(content.name)
        expect(description).to.be.equal(content.description)
        expect("Vault").to.be.equal(content.type)
    })

    it('user lists credentials again under a collection after update', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {contents} = await credentialManager.getAll(auth, {collectionID: previous})
        expect(Object.keys(contents)).to.have.lengthOf(3)
    })

    it('user removes a credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.remove(auth, sampleUDC.id)

        let {contents} = await credentialManager.getAll(auth, {collectionID: previous})
        expect(Object.keys(contents)).to.have.lengthOf(2)
    })

    it('user removes a collection from wallet', async function () {
        let collectionManager = new CollectionManager({agent: walletUserAgent, user: WALLET_USER})

        await collectionManager.remove(auth, previous)

        expect(collectionManager.get(auth, previous)).to.eventually.be.rejected
    })

    it('user finds credentials belonging to collection removed from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        // credential metadata gone from collection
        let metadataList = await credentialManager.getAllCredentialMetadata(auth, {collectionID: previous})
        expect(metadataList).to.have.lengthOf(2)

        let {contents} = await credentialManager.getAll(auth)
        expect(Object.keys(contents)).to.be.empty
    })

    it('user gets all collections from wallet', async function () {
        let collectionManager = new CollectionManager({agent: walletUserAgent, user: WALLET_USER})

        const len = 10

        let batch = []
        for (let i = 0; i < len; i++) {
            batch.push(collectionManager.create(auth, {name: `${name}-${i}`}))
        }

        await Promise.all(batch)

        let {contents} = await collectionManager.getAll(auth)
        expect(Object.keys(contents)).to.have.lengthOf(len)
    })

})
