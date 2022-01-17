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

const WALLET_WACI_V2_USER = 'smith-waci-agent-v2'
const ISSUER_V2_LABEL = 'vc-issuer-agent'
const RP_V2_LABEL = 'vc-rp-agent'


let walletUserAgentV2, issuerV2, rpV2, sampleUDC, samplePRC, routerConnections, auth, controller

before(async function () {
    this.timeout(0)

    // wallet agent
    walletUserAgentV2 = await loadFrameworks({name: WALLET_WACI_V2_USER, mediaTypeProfiles:["didcomm/v2"], keyType:'ecdsap256ieee1363', keyAgreementType:'p256kw'})

    // issuer agent
    issuerV2 = new IssuerAdapter(ISSUER_V2_LABEL)
    await issuerV2.init({mediaTypeProfiles:["didcomm/v2"], keyType:'ecdsap256ieee1363', keyAgreementType:'p256kw'})

    // rp agent
    rpV2 = new VerifierAdapter(RP_V2_LABEL)
    await rpV2.init({mediaTypeProfiles:["didcomm/v2"], keyType:'ecdsap256ieee1363', keyAgreementType:'p256kw'})

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

    // register wallet to router
    await connectToMediator(walletUserAgentV2, testConfig.mediatorEndPoint, {}, {isDIDCommV2: true})
    let conns = await getMediatorConnections(walletUserAgentV2, {single: true})
    expect(conns).to.not.empty
    routerConnections = [conns]

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

    // pre-load wallet with university degree and permanent resident card credentials.
    let credentialManager = new CredentialManager({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
    await credentialManager.save(auth, {credentials: [sampleUDC, samplePRC]})
    let {contents} = await credentialManager.getAll(auth)
    expect(contents).to.not.empty
    expect(Object.keys(contents)).to.have.lengthOf(2)
});

after(function () {
    walletUserAgentV2 ? walletUserAgentV2.destroy() : ''
    issuerV2 ? issuerV2.destroy() : ''
});

describe('Wallet DIDComm WACI credential share flow', async function () {
    let credentialInteraction
    it('user accepts out-of-band invitation from relying party and initiates WACI credential interaction', async function () {
        let invitation = await rpV2.createInvitation({goal_code: 'streamlined-vp'})
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
        credentialInteraction = await didcomm.initiateCredentialShare(auth, invitation, {routerConnections})

        let {threadID, presentations, normalized} = credentialInteraction
        expect(threadID).to.not.empty
        expect(presentations).to.not.empty
        expect(normalized).to.not.empty
    })

    it('user gives consent and concludes credential interaction by presenting proof to relying party', async function () {
        let {threadID, presentations} = credentialInteraction

        const redirectURL = "http://example.com/success"
        let acceptPresentation = rpV2.acceptPresentProof({redirectURL})

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        const response = await didcomm.completeCredentialShare(auth, threadID, presentations, {controller}, {waitForDone: true})
        expect(response.status).to.be.equal("OK")
        expect(response.url).to.be.equal(redirectURL)


        let presentation = await acceptPresentation
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential[0].id).to.be.equal(samplePRC.id)
    })

    it('user accepts out-of-band invitation from relying party and initiates WACI credential interaction', async function () {
        let invitation = await rpV2.createInvitation({goal_code: 'streamlined-vp'})
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
        credentialInteraction = await didcomm.initiateCredentialShare(auth, invitation, {routerConnections})

        let {threadID, presentations} = credentialInteraction
        expect(threadID).to.not.empty
        expect(presentations).to.not.empty
    })

    it('user gives consent and concludes credential interaction by presenting proof to relying party, but relying party rejects it', async function () {
        let {threadID, presentations} = credentialInteraction

        const redirectURL = "http://example.com/error"
        let declinePresentProof = rpV2.declinePresentProof({redirectURL})


        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})

        const response = await didcomm.completeCredentialShare(auth, threadID, presentations, {controller}, {waitForDone: true, autoAccept: true})
        expect(response.status).to.be.equal("FAIL")
        expect(response.url).to.be.equal(redirectURL)

        let presentation = await declinePresentProof
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential[0].id).to.be.equal(samplePRC.id)
    })
})


describe('Wallet DIDComm WACI credential issuance flow - success scenarios', async function () {
    const fulfillmentJSON =  getJSONTestData("cred-fulfillment-DL.json")
    const sampleComment = "Offer to issue Drivers License for Mr.Smith"
    let udcVC = getJSONTestData('udc-vc.json')

    let credentialInteraction
    it('user accepts out-of-band invitation from issuer and initiates WACI credential interaction - presentation exchange flow', async function () {
        const manifestJSON =  getJSONTestData("cred-manifest-withdef.json")

        let invitation = await issuerV2.createInvitation({goal_code: 'streamlined-vc'})
        issuerV2.acceptExchangeRequest()
        issuerV2.acceptCredentialProposal({
            comment: sampleComment,
            manifest: manifestJSON,
            fulfillment:fulfillmentJSON,
        })

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        credentialInteraction = await didcomm.initiateCredentialIssuance(auth, invitation, {routerConnections})

        const { threadID, manifest, fulfillment, presentations, normalized, domain, challenge, comment, error } = credentialInteraction

        expect(threadID).to.not.empty
        expect(manifest).to.not.empty
        expect(manifest.id).to.be.equal(manifestJSON.id)
        expect(fulfillment).to.not.empty
        expect(presentations).to.not.empty
        expect(normalized).to.not.empty
        expect(threadID).to.not.empty
        expect(domain).to.be.equal(manifestJSON.options.domain)
        expect(challenge).to.be.equal(manifestJSON.options.challenge)
        expect(comment).to.be.equal(sampleComment)
        expect(error).to.be.undefined
    })

    it('user gives consent and concludes credential interaction by providing credential application in request credential message - presentation exchange flow', async function () {
        let {threadID, presentations} = credentialInteraction

        // setup issuer.
        udcVC.id = `http://example.edu/credentials/${uuid()}`
        let [credential] = await issuerV2.issue(udcVC)
        let acceptCredential = issuerV2.acceptRequestCredential({credential})

        // complete credential interaction.
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        const response = await didcomm.completeCredentialIssuance(auth, threadID, presentations[0], {controller}, {waitForDone: true, autoAccept: true})
        expect(response.status).to.be.equal("OK")

        // verify if issuer got expected message.
        let presentation = await acceptCredential
        expect(presentation.verifiableCredential).to.not.empty

        // verify if new credential is saved.
        let credentialManager = new CredentialManager({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(3)
    })

    it('user accepts out-of-band invitation from issuer and initiates WACI credential interaction - DID Auth flow', async function () {
        const manifestJSON =  getJSONTestData("cred-manifest-withoptions.json")

        let invitation = await issuerV2.createInvitation({goal_code: 'streamlined-vc'})
        issuerV2.acceptExchangeRequest()
        issuerV2.acceptCredentialProposal({
            comment: sampleComment,
            manifest: manifestJSON,
            fulfillment:fulfillmentJSON,
        })

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        credentialInteraction = await didcomm.initiateCredentialIssuance(auth, invitation, {routerConnections})

        const { threadID, manifest, fulfillment, presentations, normalized, domain, challenge, comment, error } = credentialInteraction

        expect(threadID).to.not.empty
        expect(manifest).to.not.empty
        expect(manifest.id).to.be.equal(manifestJSON.id)
        expect(fulfillment).to.not.empty
        expect(presentations).to.not.empty
        expect(normalized).to.be.undefined
        expect(domain).to.be.equal(manifestJSON.options.domain)
        expect(challenge).to.be.equal(manifestJSON.options.challenge)
        expect(comment).to.be.equal(sampleComment)
        expect(error).to.be.undefined
    })

    it('user gives consent and concludes credential interaction by providing credential application in request credential message (redirect flow) - DID Auth flow', async function () {
        let {threadID, presentations} = credentialInteraction

        // setup issuer.
        const redirect = "https://example.com/success"
        udcVC.id = `http://example.edu/credentials/${uuid()}`
        let [credential] = await issuerV2.issue(udcVC)
        let acceptCredential = issuerV2.acceptRequestCredential({credential, redirect})

        // complete credential interaction.
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        const response = await didcomm.completeCredentialIssuance(auth, threadID, presentations[0], {controller}, {waitForDone: true, autoAccept: true})
        expect(response.status).to.be.equal("OK")
        expect(response.url).to.be.equal(redirect)

        // verify if issuer got expected message.
        let presentation = await acceptCredential
        expect(presentation.verifiableCredential).to.be.null
        expect(presentation.proof).to.not.empty

        // verify if new credential is saved.
        let credentialManager = new CredentialManager({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(4)
    })

    it('user accepts out-of-band V2 invitation from issuer and initiates WACI credential interaction - Basic flow', async function () {
        const manifestJSON =  getJSONTestData("cred-manifest.json")

        let invitation = await issuerV2.createInvitation({goal_code: 'streamlined-vc'})
        issuerV2.acceptExchangeRequest()
        issuerV2.acceptCredentialProposal({
            comment: sampleComment,
            manifest: manifestJSON,
            fulfillment:fulfillmentJSON,
        })

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        credentialInteraction = await didcomm.initiateCredentialIssuance(auth, invitation, {routerConnections})

        const { threadID, manifest, fulfillment, presentations, normalized, domain, challenge, comment, error } = credentialInteraction

        expect(threadID).to.not.empty
        expect(manifest).to.not.empty
        expect(manifest.id).to.be.equal(manifestJSON.id)
        expect(fulfillment).to.not.empty
        expect(presentations).to.be.undefined
        expect(normalized).to.be.undefined
        expect(domain).to.be.undefined
        expect(challenge).to.be.undefined
        expect(comment).to.be.equal(sampleComment)
        expect(error).to.be.undefined
    })

    it('user gives consent and concludes credential interaction by providing credential application in request credential message - Basic flow', async function () {
        let {threadID, presentations} = credentialInteraction

        // setup issuer.
        udcVC.id = `http://example.edu/credentials/${uuid()}`
        let [credential] = await issuerV2.issue(udcVC)
        let acceptCredential = issuerV2.acceptRequestCredential({credential})

        // complete credential interaction.
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        const response = await didcomm.completeCredentialIssuance(auth, threadID, null, {controller}, {waitForDone: true, autoAccept: true})
        expect(response.status).to.be.equal("OK")

        // verify if issuer got expected message.
        let presentation = await acceptCredential
        expect(presentation).to.be.null

        // verify if new credential is saved.
        let credentialManager = new CredentialManager({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(5)
    })
})

describe('Wallet DIDComm WACI credential issuance flow - failure scenarios', async function () {
    const fulfillmentJSON =  getJSONTestData("cred-fulfillment-DL.json")
    const sampleComment = "Offer to issue Drivers License for Mr.Smith"
    let udcVC = getJSONTestData('udc-vc.json')

    it('user accepts out-of-band invitation from issuer, initiates WACI credential interaction and issuer declines proposal', async function () {
        const manifestJSON =  getJSONTestData("cred-manifest-withdef.json")
        const redirectURL = "https://example.com/error"

        let invitation = await issuerV2.createInvitation({goal_code: 'streamlined-vc'})
        issuerV2.acceptExchangeRequest()
        issuerV2.declineCredentialProposal({redirectURL})

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        credentialInteraction = await didcomm.initiateCredentialIssuance(auth, invitation, {routerConnections})
        expect(credentialInteraction).to.not.empty

        const {threadID, manifest, fulfillment, presentations, normalized, domain, challenge, comment, error} = credentialInteraction

        expect(threadID).to.be.undefined
        expect(manifest).to.be.undefined
        expect(fulfillment).to.be.undefined
        expect(presentations).to.be.undefined
        expect(normalized).to.be.undefined
        expect(domain).to.be.undefined
        expect(challenge).to.be.undefined
        expect(comment).to.be.undefined

        expect(error).to.not.empty
        const {status, url, code} = error
        expect(status).to.be.equal("FAIL")
        expect(url).to.be.equal(redirectURL)
        expect(code).to.be.equal("rejected")
    })

    let credentialInteraction
    it('user accepts out-of-band invitation from issuer and initiates WACI credential interaction', async function () {
        const manifestJSON =  getJSONTestData("cred-manifest-withdef.json")

        let invitation = await issuerV2.createInvitation({goal_code: 'streamlined-vc'})
        issuerV2.acceptExchangeRequest()
        issuerV2.acceptCredentialProposal({
            comment: sampleComment,
            manifest: manifestJSON,
            fulfillment:fulfillmentJSON,
        })

        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        credentialInteraction = await didcomm.initiateCredentialIssuance(auth, invitation, {routerConnections})

        const { threadID, manifest, fulfillment, presentations, normalized, domain, challenge, comment, error } = credentialInteraction

        expect(threadID).to.not.empty
        expect(manifest).to.not.empty
        expect(manifest.id).to.be.equal(manifestJSON.id)
        expect(fulfillment).to.not.empty
        expect(presentations).to.not.empty
        expect(normalized).to.not.empty
        expect(threadID).to.not.empty
        expect(domain).to.be.equal(manifestJSON.options.domain)
        expect(challenge).to.be.equal(manifestJSON.options.challenge)
        expect(comment).to.be.equal(sampleComment)
        expect(error).to.be.undefined
    })

    it('user gives consent by submitting credential application but issuer declines it', async function () {
        let {threadID, presentations} = credentialInteraction

        // setup issuer.
        const redirectURL = "https://example.com/error"
        udcVC.id = `http://example.edu/credentials/${uuid()}`
        let [credential] = await issuerV2.issue(udcVC)
        let acceptCredential = issuerV2.declineRequestCredential({redirectURL})

        // complete credential interaction.
        let didcomm = new DIDComm({agent: walletUserAgentV2, user: WALLET_WACI_V2_USER})
        const response = await didcomm.completeCredentialIssuance(auth, threadID, presentations[0], {controller}, {waitForDone: true, autoAccept: true})

        expect(response).to.not.empty
        const {status, url} = response
        expect(status).to.be.equal("FAIL")
        expect(url).to.be.equal(redirectURL)
    })
})
