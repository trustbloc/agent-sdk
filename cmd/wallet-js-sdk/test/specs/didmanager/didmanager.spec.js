/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";
import {loadFrameworks, testConfig} from "../common";
import {connectToMediator, createWalletProfile, getMediatorConnections, UniversalWallet, DIDManager} from "../../../src";

const RICK_USER = 'rick-agent'
const keyType = 'ED25519'
const signatureType = 'Ed25519VerificationKey2018'

let rick

before(async function () {
    rick = await loadFrameworks({name: RICK_USER})
});

after(function () {
    rick ? rick.destroy() : ''
});


describe('DID Manager tests', async function () {
    it('rick creates his wallet profile', async function () {
        await createWalletProfile(rick, RICK_USER, {localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('rick opens wallet', async function () {
        let wallet = new UniversalWallet({agent: rick, user: RICK_USER})
        let authResponse = await wallet.open({localKMSPassphrase: testConfig.walletUserPassphrase})
        expect(authResponse.token).to.not.empty
        auth = authResponse.token
    })

    it('rick creates trustbloc DID in wallet', async function () {
        let didManager = new DIDManager({agent: rick, user: RICK_USER})
        await didManager.createTrustBlocDID({auth})
    })

    it('rick connects to mediator', async function () {
        await connectToMediator(rick, testConfig.mediatorEndPoint)
        let conns = await getMediatorConnections(rick)
        expect(conns).to.not.empty
    });

    it('rick creates peer DID in wallet', async function () {
        let didManager = new DIDManager({agent: rick, user: RICK_USER})
        await didManager.createPeerDID({auth})
    })

    it('rick imports a DID with key into wallet', async function () {
        let didManager = new DIDManager({agent: rick, user: RICK_USER})
        await didManager.importDID({
            auth,
            did: 'did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5',
            key: {
                keyType: 'Ed25519VerificationKey2018',
                privateKeyBase58: '2MP5gWCnf67jvW3E4Lz8PpVrDWAXMYY1sDxjnkEnKhkkbKD7yP2mkVeyVpu5nAtr3TeDgMNjBPirk2XcQacs3dvZ',
                keyID: 'did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5',
            }
        })
    })

    it('rick lists all DIDs from wallet', async function () {
        let didManager = new DIDManager({agent: rick, user: RICK_USER})
        let {contents} = await didManager.getAllDIDs({auth})
        expect(Object.keys(contents)).to.have.lengthOf(3)
    })
})
