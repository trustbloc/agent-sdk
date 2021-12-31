/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";

import {getJSONTestData, loadFrameworks, retryWithDelay, testConfig, wait} from "../common";
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

const WALLET_WACI_V2_USER = 'smith-waci-agent-v2'
const ISSUER_V2_LABEL = 'vc-issuer-agent-v2'
const RP_V2_LABEL = 'vc-rp-agent-v2'
const KEY_TYPE = 'ecdsap256ieee1363'
const KEY_AGREEMENT_TYPE = 'p256kw'
const MEDIA_TYPE_PROFILE = "didcomm/v2"
const KEY_TYPE_KMS = "ECDSAP256IEEEP1363"
const KEY_AGREEMENT_TYPE_KMS = "NISTP256ECDHKW"
const DIDCOMM_V2_SERVICE_TYPE = "DIDCommMessaging"


let invID, walletUserAgentV2, issuerV2, rpV2, sampleUDC, samplePRC, routerRPConnections, auth, controller

before(async function () {
    this.timeout(0)

    // wallet agent
    walletUserAgentV2 = await loadFrameworks({name: WALLET_WACI_V2_USER, mediaTypeProfiles:[MEDIA_TYPE_PROFILE], keyType: KEY_TYPE, keyAgreementType: KEY_AGREEMENT_TYPE})

    // issuer agent
    issuerV2 = new IssuerAdapter(ISSUER_V2_LABEL)
    await issuerV2.init({mediaTypeProfiles:[MEDIA_TYPE_PROFILE], keyType: KEY_TYPE, keyAgreementType: KEY_AGREEMENT_TYPE})

    // rp agent
    rpV2 = new VerifierAdapter(RP_V2_LABEL)
    let rpConns = await rpV2.init({mediaTypeProfiles:[MEDIA_TYPE_PROFILE], keyType: KEY_TYPE, keyAgreementType: KEY_AGREEMENT_TYPE})
    console.debug("before test - rpConnections: "+JSON.stringify(rpConns))
    routerRPConnections = [rpConns]

    // load sample VCs from testdata
    let udcVC = getJSONTestData('udc-vc.json')
    let prcVC = getJSONTestData('prc-vc.json')

    // issue sample credentials
    let [vc1, vc2] = await issuerV2.issue(udcVC, prcVC)
    expect(vc1.id).to.not.empty
    expect(vc1.credentialSubject).to.not.empty
    expect(vc2.id).to.not.empty
    expect(vc2.credentialSubject).to.not.empty

    sampleUDC = vc1
    samplePRC = vc2

    let routerInvID

    // register wallet to router
    null, routerInvID = await connectToMediator(walletUserAgentV2, testConfig.mediatorV2EndPoint, {}, {isDIDCommV2: true})
    let conns = await getMediatorConnections(walletUserAgentV2, {single: true})
    expect(conns).to.not.empty

    let routerDoc = await walletUserAgentV2.didclient.resolveOrbDID({did: routerInvID})

    console.debug("before test - routerDIDDoc: " + JSON.stringify(routerDoc, undefined, 2))

    // create wallet profile
    let walletUser = new WalletUser({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
    await walletUser.createWalletProfile({localKMSPassphrase: testConfig.walletUserPassphrase})

    // unlock wallet
    let authResponse = await walletUser.unlock({localKMSPassphrase: testConfig.walletUserPassphrase})
    expect(authResponse.token).to.not.empty
    auth = authResponse.token

    // create orb DID as controller for signing from wallet.
    let didManager = new DIDManager({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
    let docres = await didManager.createOrbDID(auth, {purposes: ["assertionMethod", "authentication"]})
    expect(docres).to.not.empty
    controller = docres.didDocument.id

    // create a new orbDID for rp adapter
    const [keySet, recoveryKeySet, updateKeySet, keyAgreementSet] = await Promise.all([
        rpV2.agent.kms.createKeySet({keyType: KEY_TYPE_KMS}),
        rpV2.agent.kms.createKeySet({keyType: KEY_TYPE_KMS}),
        rpV2.agent.kms.createKeySet({keyType: KEY_TYPE_KMS}),
        rpV2.agent.kms.createKeySet({keyType: KEY_AGREEMENT_TYPE_KMS})
    ])

    // add router doc's keyAgreement VM to the public keys request to register router keys in the rp DID doc's
    // service block as a routingKey.
    let routerKeyAgreementIDs
    for (let vm in routerDoc.didDocument.verificationMethod) {
        let routerVM = routerDoc.didDocument.verificationMethod[vm]
        console.debug("before test - routerVM: " + JSON.stringify(routerVM, undefined, 2))
        if (routerVM.id.includes("#key-2")) {
            routerKeyAgreementIDs = routerVM.id
        } else {
            // only add public key of keyAgreement VM for router
            continue
        }
    }

    const createRPDIDRequest = {
        "serviceID": "rpServiceID",
        "serviceEndpoint": testConfig.mediatorV2WSEndPoint,
        "didcommServiceType": DIDCOMM_V2_SERVICE_TYPE,
        "publicKeys": [{
            "id": keySet.keyID,
            "type": 'JsonWebKey2020',
            "value": keySet.publicKey,
            "encoding": "Jwk",
            "keyType": KEY_TYPE_KMS,
            "purposes": ["authentication"]
        }, {
            "id": recoveryKeySet.keyID,
            "type": 'JsonWebKey2020',
            "value": recoveryKeySet.publicKey,
            "encoding": "Jwk",
            "keyType": KEY_TYPE_KMS,
            "recovery": true
        }, {
            "id": updateKeySet.keyID,
            "type": 'JsonWebKey2020',
            "value": updateKeySet.publicKey,
            "encoding": "Jwk",
            "keyType": KEY_TYPE_KMS,
            "update": true
        }, {
            "id": keyAgreementSet.keyID,
            "type": 'JsonWebKey2020',
            "value": keyAgreementSet.publicKey,
            "encoding": "Jwk",
            "keyType": KEY_AGREEMENT_TYPE_KMS,
            "purposes": ["keyAgreement"]
        }],
        "routerKAIDS": [routerKeyAgreementIDs],
        "routerConnections": routerRPConnections,
    };

    console.debug("before test - about to call rpV2.agent.didclient.createOrbDID with createRPDIDRequest: "+JSON.stringify(createRPDIDRequest, undefined, 2)+"...")
    let rpDocRes = await rpV2.agent.didclient.createOrbDID(createRPDIDRequest)
    console.debug("before test - rpDocRes ORB DID created: "+JSON.stringify(rpDocRes, undefined, 2))
    expect(rpDocRes).to.not.empty
    invID = rpDocRes.didDocument.id

    // pre-load wallet with university degree and permanent resident card credentials.
    let credentialManager = new CredentialManager({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
    await credentialManager.save(auth, {credentials: [sampleUDC, samplePRC]})
    let {contents} = await credentialManager.getAll(auth)
    expect(contents).to.not.empty
    expect(Object.keys(contents)).to.have.lengthOf(2)
});

after(function () {
    walletUserAgentV2 ? walletUserAgentV2.destroy() : ''
    rpV2.agent ? rpV2.agent.destroy() : ''
    rpV2 ? rpV2.destroy() : ''
    issuerV2 ? issuerV2.destroy() : ''
});

describe('Wallet DIDComm V2 WACI credential share flow', async function () {
    let credentialInteraction
    it('user accepts out-of-band invitation from relying party and initiates WACI credential interaction', async function () {

        await wait(3000) // wait to make sure orb DID of invID was created in ORB sidetree

        let invitation = await rpV2.createInvitation({goal_code: 'streamlined-vp', from:invID})
        console.debug("rpV2.createInvitation() called, invitation:"+JSON.stringify(invitation))
        rpV2.acceptExchangeRequest()
        console.debug("debug acceptExchangeRequest() called")
        rpV2.acceptPresentationProposal({
            "id": "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
            "input_descriptors": [{
                "id": "20b073bb-cede-4912-9e9d-334e5702077b",
                "schema": [{"uri": "https://w3id.org/citizenship#PermanentResidentCard"}],
                "constraints": {"fields": [{"path": ["$.credentialSubject.familyName"]}]}
            }]
        })
        console.debug("debug acceptPresentationProposal() called with rrouterRPConnections:"+ routerRPConnections)
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        credentialInteraction = await didcomm.initiateCredentialShare(auth, invitation, {userAnyRouterConnection: true})
        console.debug("debug initiateCredentialShare() called, credentialInteraction:"+JSON.stringify(credentialInteraction, undefined, 2))

        let {threadID, presentations, normalized} = credentialInteraction
        expect(threadID).to.not.empty
        expect(presentations).to.not.empty
        expect(normalized).to.not.empty
    })

    it('user gives consent and concludes credential interaction by presenting proof to relying party', async function () {
        let {threadID, presentations} = credentialInteraction

        const redirectURL = "http://example.com/success"
        let acceptPresentation = rpV2.acceptPresentProof({redirectURL})
        console.debug("acceptPresentProof() called, acceptPresentation:"+JSON.stringify(acceptPresentation, undefined, 2))
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        const response = await didcomm.completeCredentialShare(auth, threadID, presentations, {controller}, {waitForDone: true})
        console.debug("completeCredentialShare() called, response:"+JSON.stringify(response, undefined, 2))

        expect(response.status).to.be.equal("OK")

        let presentation = await acceptPresentation
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential[0].id).to.be.equal(samplePRC.id)
    })

    it('user accepts out-of-band v2 invitation from relying party and initiates WACI credential interaction', async function () {
        let invitation = await rpV2.createInvitation({goal_code: 'streamlined-vp', from:invID})
        rpV2.acceptExchangeRequest()
        rpV2.acceptPresentationProposal({
            "id": "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
            "input_descriptors": [{
                "id": "20b073bb-cede-4912-9e9d-334e5702077b",
                "schema": [{"uri": "https://w3id.org/citizenship#PermanentResidentCard"}],
                "constraints": {"fields": [{"path": ["$.credentialSubject.familyName"]}]}
            }]
        })

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        console.debug("before test - about to call initiateCredentialShare(), invID: "+ invID +", invitation: " + JSON.stringify(invitation))
        credentialInteraction = await didcomm.initiateCredentialShare(auth, invitation, {userAnyRouterConnection: true})
        console.debug("debug initiateCredentialShare() called, credentialInteraction:"+JSON.stringify(credentialInteraction, undefined, 2))

        let {threadID, presentations} = credentialInteraction
        expect(threadID).to.not.empty
        expect(presentations).to.not.empty
    })

    it('user gives consent and concludes credential interaction by presenting proof to relying party, but relying party rejects it', async function () {
        let {threadID, presentations} = credentialInteraction

        const redirectURL = "http://example.com/error"
        let declinePresentProof = rpV2.declinePresentProof({redirectURL})
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})

        console.debug("completeCredentialShare() called, presentations:"+JSON.stringify(presentations, undefined, 2))
        const response = await didcomm.completeCredentialShare(auth, threadID, presentations, {controller}, {waitForDone: true, autoAccept: true})
        console.debug("completeCredentialShare() called, response:"+JSON.stringify(response, undefined, 2))

        expect(response.status).to.be.equal("FAIL")

        let presentation = await declinePresentProof
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential[0].id).to.be.equal(samplePRC.id)
    })
})
