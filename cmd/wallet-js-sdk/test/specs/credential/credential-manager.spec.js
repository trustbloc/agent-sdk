/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";

import {getJSONTestData, loadFrameworks, retryWithDelay, testConfig, wait} from "../common";
import {CredentialManager, DIDManager, WalletUser} from "../../../src";

var uuid = require('uuid/v4')

const WALLET_USER = 'smith-agent'
const WALLET_QUERY_USER = 'smith-query-agent'
const WALLET_DIDCOMM_USER = 'smith-didcomm-agent'
const VC_ISSUER = 'vc-issuer-agent'


let walletUserAgent, issuer, sampleUDC, samplePRC, sampleUDCBBS, sampleFrameDoc, manifest

before(async function () {
    this.timeout(0)
    walletUserAgent = await loadFrameworks({name: WALLET_USER})
    issuer = await loadFrameworks({name: VC_ISSUER})

    // load sample VCs from testdata
    let udcVC = getJSONTestData('udc-vc.json')
    let prcVC = getJSONTestData('prc-vc.json')
    sampleUDCBBS = getJSONTestData('udc-bbs-vc.json')
    sampleFrameDoc = getJSONTestData('udc-frame.json')
    manifest = getJSONTestData('manifest-vc.json')

    // issue sample credentials
    let [vc1, vc2] = await issueCredential(issuer, udcVC, prcVC)
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


    it('user saves a credential into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.save(auth, {credential: sampleUDC})
    })

    it('user saves a BBS credential into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.save(auth, {credential:sampleUDCBBS})
    })

    it('user saves a credential into wallet by verifying', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.save(auth, {credential:samplePRC}, {verify: true})
    })

    it('user gets all credentials from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(3)
    })

    it('user gets a credential from wallet by id', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {content} = await credentialManager.get(auth, sampleUDC.id)
        expect(content).to.not.empty
    })

    it('user removes a credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.remove(auth, samplePRC.id)
        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(2)
    })

    let did
    it('user creates Orb DID in wallet', async function () {
        let didManager = new DIDManager({agent: walletUserAgent, user: WALLET_USER})

        let docres = await didManager.createOrbDID(auth, {purposes: ["assertionMethod", "authentication"]})
        expect(docres).to.not.empty
        did = docres.DIDDocument.id
    })

    it('user issues a credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {credential} = await credentialManager.issue(auth, samplePRC, {controller: did})
        expect(credential).to.not.empty
        expect(credential.proof).to.not.empty
        expect(credential.proof).to.have.lengthOf(2)
    })

    it('user verifies a credential stored in wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {verified, error} = await credentialManager.verify(auth, {storedCredentialID: sampleUDC.id})
        expect(verified).to.be.true
        expect(error).to.be.undefined
    })

    it('user verifies a raw credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {verified, error} = await credentialManager.verify(auth, {rawCredential: sampleUDC})
        expect(verified).to.be.true
        expect(error).to.be.undefined
    })

    it('user verifies a tampered raw credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let tampered = JSON.parse(JSON.stringify(sampleUDC))
        tampered.credentialSubject.name = 'Mr.Fake'
        let {verified, error} = await credentialManager.verify(auth, {rawCredential: tampered})
        expect(verified).to.not.true
        expect(error).to.not.empty
    })

    it('user presents credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        let {presentation} = await credentialManager.present(auth, {
            rawCredentials: [samplePRC],
            storedCredentials: [sampleUDC.id]
        }, {controller: did})
        expect(presentation).to.not.empty
        expect(presentation.verifiableCredential).to.not.empty
        expect(presentation.verifiableCredential).to.have.lengthOf(2)
    })

    it('user derives a stored credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {credential} = await credentialManager.derive(auth, {storedCredentialID: sampleUDCBBS.id}, {
            frame: sampleFrameDoc, nonce: uuid()
        })
        expect(credential).to.not.empty
        expect(credential.credentialSubject).to.not.empty
        expect(credential.proof).to.not.empty
    })

    it('user derives a raw credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {credential} = await credentialManager.derive(auth, {rawCredential: sampleUDCBBS}, {
            frame: sampleFrameDoc, nonce: uuid()
        })
        expect(credential).to.not.empty
        expect(credential.credentialSubject).to.not.empty
        expect(credential.proof).to.not.empty
    })
})

describe('Credential Query Tests', async function () {
    it('user creates his wallet profile', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_QUERY_USER})

        await walletUser.createWalletProfile({localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('user opens his wallet', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_QUERY_USER})

        let authResponse = await walletUser.unlock({localKMSPassphrase: testConfig.walletUserPassphrase})

        expect(authResponse.token).to.not.empty

        auth = authResponse.token
    })


    it('user saves multiple credentials into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_QUERY_USER})

        await credentialManager.save(auth, {credentials:[sampleUDC, samplePRC, sampleUDCBBS]})

        let {contents} = await credentialManager.getAll(auth)

        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(3)
    })

    it('user performs DIDAuth in wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_QUERY_USER})

        let {results} = await credentialManager.query(auth, [{
            "type": "DIDAuth",
        }])

        expect(results).to.have.lengthOf(1)
        expect(results[0].verifiableCredential).to.have.lengthOf(0)
    })

    it('user performs QueryByExample in wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_QUERY_USER})

        let {results} = await credentialManager.query(auth, [{
            "type": "QueryByExample",
            "credentialQuery": [{
                "reason": "Please present your valid degree certificate.",
                "example": {
                    "@context": ["https://www.w3.org/2018/credentials/v1", "https://www.w3.org/2018/credentials/examples/v1"],
                    "type": ["UniversityDegreeCredential"],
                    "trustedIssuer": [
                        {"issuer": "urn:some:required:issuer"},
                        {
                            "required": true,
                            "issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f"
                        }
                    ],
                    "credentialSubject": {"id": "did:example:ebfeb1f712ebc6f1c276e12ec21"}
                }
            }]
        }])

        expect(results).to.have.lengthOf(1)
        expect(results[0].verifiableCredential).to.have.lengthOf(2)
    })

    it('user performs QueryByFrame in wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_QUERY_USER})

        let {results} = await credentialManager.query(auth, [{
            "type": "QueryByFrame",
            "credentialQuery": [{
                "reason": "Please provide your Passport details.",
                "frame": {
                    "@context": ["https://www.w3.org/2018/credentials/v1", "https://w3id.org/citizenship/v1", "https://w3id.org/security/bbs/v1"],
                    "type": ["VerifiableCredential", "PermanentResidentCard"],
                    "@explicit": true,
                    "identifier": {},
                    "issuer": {},
                    "issuanceDate": {},
                    "credentialSubject": {"@explicit": true, "name": {}, "spouse": {}}
                },
                "trustedIssuer": [{"issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f", "required": true}],
                "required": true
            }]
        }])

        expect(results).to.have.lengthOf(1)
        expect(results[0].verifiableCredential).to.have.lengthOf(1)
    })

    it('user performs PresentationExchange in wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_QUERY_USER})

        let {results} = await credentialManager.query(auth, [{
            "type": "PresentationExchange",
            "credentialQuery": [{
                "id": "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
                "input_descriptors": [{
                    "id": "20b073bb-cede-4912-9e9d-334e5702077b",
                    "schema": [{"uri": "https://www.w3.org/2018/credentials/v1#VerifiableCredential"}],
                    "constraints": {"fields": [{"path": ["$.credentialSubject.familyName"]}]}
                }]
            }]
        }])

        expect(results).to.have.lengthOf(1)
        expect(results[0].verifiableCredential).to.have.lengthOf(1)
    })

    it('user performs mixed query in wallet - PresentationExchange, QueryByFrame, QueryByExample ', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_QUERY_USER})

        let {results} = await credentialManager.query(auth, [{
            "type": "PresentationExchange",
            "credentialQuery": [{
                "id": "22c77155-edf2-4ec5-8d44-b393b4e4fa38",
                "input_descriptors": [{
                    "id": "20b073bb-cede-4912-9e9d-334e5702077b",
                    "schema": [{"uri": "https://www.w3.org/2018/credentials/v1#VerifiableCredential"}],
                    "constraints": {"fields": [{"path": ["$.credentialSubject.familyName"]}]}
                }]
            }]
        }, {
            "type": "QueryByFrame",
            "credentialQuery": [{
                "reason": "Please provide your Passport details.",
                "frame": {
                    "@context": ["https://www.w3.org/2018/credentials/v1", "https://w3id.org/citizenship/v1", "https://w3id.org/security/bbs/v1"],
                    "type": ["VerifiableCredential", "PermanentResidentCard"],
                    "@explicit": true,
                    "identifier": {},
                    "issuer": {},
                    "issuanceDate": {},
                    "credentialSubject": {"@explicit": true, "name": {}, "spouse": {}}
                },
                "trustedIssuer": [{"issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f", "required": true}],
                "required": true
            }]
        }, {
            "type": "QueryByExample",
            "credentialQuery": [{
                "reason": "Please present your valid degree certificate.",
                "example": {
                    "@context": ["https://www.w3.org/2018/credentials/v1", "https://www.w3.org/2018/credentials/examples/v1"],
                    "type": ["UniversityDegreeCredential"],
                    "trustedIssuer": [
                        {"issuer": "urn:some:required:issuer"},
                        {
                            "required": true,
                            "issuer": "did:example:76e12ec712ebc6f1c221ebfeb1f"
                        }
                    ],
                    "credentialSubject": {"id": "did:example:ebfeb1f712ebc6f1c276e12ec21"}
                }
            }]
        }
        ])

        expect(results).to.have.lengthOf(2)
    })
})

describe('Credential Manager DIDComm Tests', async function () {
    it('user creates his wallet profile', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        await walletUser.createWalletProfile({localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('user opens his wallet', async function () {
        let walletUser = new WalletUser({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        let authResponse = await walletUser.unlock({localKMSPassphrase: testConfig.walletUserPassphrase})

        expect(authResponse.token).to.not.empty

        auth = authResponse.token
    })

    let connectionID = uuid()
    it('user saves manifest credential with connection info', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        await credentialManager.saveManifestCredential(auth, manifest, connectionID)
    })

    it('user gets connection info from manifest', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        let mconnectionID = await credentialManager.getManifestConnection(auth, manifest.id)
        expect(mconnectionID).to.be.equal(connectionID)
    })

    // TODO currently failing due to document loader issue.
    it.skip('user queries manifest credentials', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        // save one more
        let manifest2 = JSON.parse(JSON.stringify(manifest))
        manifest2.id = `http://example.gov/credentials/${uuid()}`
        await credentialManager.saveManifestCredential(auth, manifest2, uuid())

        let {contents} = await credentialManager.getAllManifests(auth)
        console.log(JSON.stringify(contents))

        expect(Object.keys(contents)).to.have.lengthOf(2)
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

    let {DIDDocument} = await issuer.didclient.createOrbDID(createDIDRequest)

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
