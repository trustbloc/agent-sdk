/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";

import {loadFrameworks, retryWithDelay, testConfig, getJSONTestData} from "../common";
import {CredentialManager, DIDManager, WalletUser} from "../../../src";

var uuid = require('uuid/v4')

const WALLET_USER = 'smith-agent'
const VC_ISSUER = 'vc-issuer-agent'


let walletUserAgent, issuer, sampleVC1, sampleVC2, sampleVCBBS, sampleFrameDoc

before(async function () {
    walletUserAgent = await loadFrameworks({name: WALLET_USER})
    issuer = await loadFrameworks({name: VC_ISSUER})

    // load sampel VCs from testdata
    sampleVC1 = getJSONTestData('udc-vc.json')
    sampleVC2 = getJSONTestData('prc-vc.json')
    sampleVCBBS = getJSONTestData('udc-bbs-vc.json')
    sampleFrameDoc = getJSONTestData('udc-frame.json')
});

after(function () {
    walletUserAgent ? walletUserAgent.destroy() : ''
    issuer ? issuer.destroy() : ''
});

describe('Credential Manager Tests', async function () {
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

    let credentials
    it('issuer issues credentials', async function () {
        this.timeout(0)
        credentials = await issueCredential(issuer, sampleVC1, sampleVC2)

        expect(credentials).to.not.empty
        expect(credentials).to.have.lengthOf(2)

        for (let credential of credentials) {
            expect(credential.id).to.not.empty
            expect(credential.credentialSubject).to.not.empty
        }
    })


    it('user saves a credential into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.save(auth, credentials[0])
    })

    it('user saves a BBS credential into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.save(auth, sampleVCBBS)
    })

    it('user saves a credential into wallet by verifying', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.save(auth, credentials[1], {verify: true})
    })

    it('user gets all credentials from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(3)
        console.log('saved credentials', Object.keys(contents))
    })

    it('user gets a credential from wallet by id', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {content} = await credentialManager.get(auth, credentials[0].id)
        expect(content).to.not.empty
    })

    it('user removes a credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.remove(auth, credentials[1].id)
        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(2)
    })

    let did
    it('user creates TrustBloc DID in wallet', async function () {
        let didManager = new DIDManager({agent: walletUserAgent, user: WALLET_USER})

        did = await didManager.createTrustBlocDID(auth, {purposes: ["assertionMethod", "authentication"]})
        expect(did).to.not.empty
    })

    it('user issues a credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {credential} = await credentialManager.issue(auth, credentials[1], {controller: did})
        expect(credential).to.not.empty
        expect(credential.proof).to.not.empty
        expect(credential.proof).to.have.lengthOf(2)
    })

    it('user verifies a credential stored in wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {verified, error} = await credentialManager.verify(auth, {storedCredentialID: credentials[0].id})
        expect(verified).to.be.true
        expect(error).to.be.undefined
    })

    it('user verifies a raw credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {verified, error} = await credentialManager.verify(auth, {rawCredential: credentials[0]})
        expect(verified).to.be.true
        expect(error).to.be.undefined
    })

    it('user verifies a tampered raw credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let tampered = JSON.parse(JSON.stringify(credentials[0]))
        tampered.credentialSubject.name = 'Mr.Fake'
        let {verified, error} = await credentialManager.verify(auth, {rawCredential: tampered})
        expect(verified).to.not.true
        expect(error).to.not.empty
    })

    it('user presents credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        let {presentation} = await credentialManager.present(auth, {
            rawCredentials: [credentials[1]],
            storedCredentials: [credentials[0].id]
        }, {controller: did})
        expect(presentation).to.not.empty
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential).to.have.lengthOf(2)
    })

    it('user derives a stored credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {credential}  = await credentialManager.derive(auth, {storedCredentialID: sampleVCBBS.id}, {
            frame: sampleFrameDoc, nonce: uuid()
        })
        expect(credential).to.not.empty
        expect(credential.credentialSubject).to.not.empty
        expect(credential.proof).to.not.empty
    })

    it('user derives a raw credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {credential} = await credentialManager.derive(auth, {rawCredential: sampleVCBBS}, {
            frame: sampleFrameDoc, nonce: uuid()
        })
        expect(credential).to.not.empty
        expect(credential.credentialSubject).to.not.empty
        expect(credential.proof).to.not.empty
    })
})

let issueCredential = async (issuer, ...credential) => {
    const keyType = 'ED25519'

    const [keySet, recoveryKeySet, updateKeySet] = await Promise.all([
        issuer.kms.createKeySet({keyType}),
        issuer.kms.createKeySet({keyType}),
        issuer.kms.createKeySet({keyType})
    ])

    const createDIDRequest = {
        "publicKeys": [{
            "id": keySet.keyID,
            "type": 'Ed25519VerificationKey2018',
            "value": keySet.publicKey,
            "encoding": "Jwk",
            keyType,
            "purposes": ["authentication"]
        }, {
            "id": recoveryKeySet.keyID,
            "type": 'Ed25519VerificationKey2018',
            "value": recoveryKeySet.publicKey,
            "encoding": "Jwk",
            keyType,
            "recovery": true
        }, {
            "id": updateKeySet.keyID,
            "type": 'Ed25519VerificationKey2018',
            "value": updateKeySet.publicKey,
            "encoding": "Jwk",
            keyType,
            "update": true
        }
        ]
    };

    let {DIDDocument} = await issuer.didclient.createTrustBlocDID(createDIDRequest)

    let resolveDID = async () => await issuer.vdr.resolveDID({id: DIDDocument.id})
    await retryWithDelay(resolveDID, 10, 5000)


    let signVCs = await Promise.all(credential.map(async (credential) => {
            let {verifiableCredential} = await issuer.verifiable.signCredential({
                credential,
                "did": DIDDocument.id,
                "signatureType": "Ed25519Signature2018"
            })

            return verifiableCredential
        }
    ))

    return signVCs
}
