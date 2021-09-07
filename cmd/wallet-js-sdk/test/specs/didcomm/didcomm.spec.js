/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";

import {getJSONTestData, loadFrameworks, retryWithDelay, testConfig} from "../common";
import {
    connectToMediator,
    CredentialManager,
    DIDComm,
    DIDManager,
    getMediatorConnections,
    WalletUser
} from "../../../src";
import {IssuerAdapter, VerifierAdapter} from "../mocks/adapters";

var uuid = require('uuid/v4')

const WALLET_WACI_USER = 'smith-waci-agent'
const ISSUER_LABEL = 'vc-issuer-agent'
const RP_LABEL = 'vc-rp-agent'


let walletUserAgent, issuer, rp, sampleUDC, samplePRC

before(async function () {
    this.timeout(0)

    // wallet agent
    walletUserAgent = await loadFrameworks({name: WALLET_WACI_USER})

    // issuer agent
    issuer = new IssuerAdapter(ISSUER_LABEL)
    await issuer.init()

    // rp agent
    rp = new VerifierAdapter(RP_LABEL)
    await rp.init()

    // load sample VCs from testdata
    let udcVC = getJSONTestData('udc-vc.json')
    let prcVC = getJSONTestData('prc-vc.json')

    // issue sample credentials
    let [vc1, vc2] = await issuer.issue(udcVC, prcVC)
    expect(vc1.id).to.not.empty
    expect(vc1.credentialSubject).to.not.empty
    expect(vc2.id).to.not.empty
    expect(vc2.credentialSubject).to.not.empty

    sampleUDC = vc1
    samplePRC = vc2
});

after(function () {
    walletUserAgent ? walletUserAgent.destroy() : ''
    issuer ? issuer.destroy() : ''
});

describe('Wallet DIDComm WACI credential share flow', async function () {
    let routerConnections
    it('wallet agent connected to mediator for DIDComm', async function () {
        await connectToMediator(walletUserAgent, testConfig.mediatorEndPoint)

        let conns = await getMediatorConnections(walletUserAgent, {single: true})
        expect(conns).to.not.empty

        routerConnections = [conns]
    })

    it('user creates his wallet profile', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_WACI_USER})

        await walletUser.createWalletProfile({localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('user opens his wallet', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_WACI_USER})

        let authResponse = await walletUser.unlock({localKMSPassphrase: testConfig.walletUserPassphrase})

        expect(authResponse.token).to.not.empty

        auth = authResponse.token
    })

    let did
    it('user creates Orb DID in wallet', async function () {
        let didManager = new DIDManager({agent: walletUserAgent, user: WALLET_WACI_USER})

        let docres = await didManager.createOrbDID(auth, {purposes: ["assertionMethod", "authentication"]})
        expect(docres).to.not.empty
        did = docres.DIDDocument.id
    })

    it('user saves credentials into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_WACI_USER})

        await credentialManager.save(auth, {credentials: [sampleUDC, samplePRC]})
    })

    let credentialInteraction
    it('user accepts out-of-band invitation from relying party and initiates WACI credential interaction', async function () {
        let invitation = await rp.createInvitation()
        rp.acceptExchangeRequest()
        rp.acceptPresentationProposal({
            "id": "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
            "input_descriptors": [{
                "id": "20b073bb-cede-4912-9e9d-334e5702077b",
                "schema": [{"uri": "https://w3id.org/citizenship#PermanentResidentCard"}],
                "constraints": {"fields": [{"path": ["$.credentialSubject.familyName"]}]}
            }]
        })

        let didcomm = new DIDComm({agent: walletUserAgent, user: WALLET_WACI_USER})
        credentialInteraction = await didcomm.initiateCredentialShare(auth, invitation, {routerConnections})

        let {threadID, presentations} = credentialInteraction
        expect(threadID).to.not.empty
        expect(presentations).to.not.empty
    })

    it('user gives consent and concludes credential interaction by presenting proof to relying party', async function () {
        let {threadID, presentations} = credentialInteraction

        let didcomm = new DIDComm({agent: walletUserAgent, user: WALLET_WACI_USER})
        await didcomm.completeCredentialShare(auth, threadID, presentations, {controller: did}, true)

        let presentation = await rp.acceptPresentProof()
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential[0].id).to.be.equal(samplePRC.id)
    })

})
