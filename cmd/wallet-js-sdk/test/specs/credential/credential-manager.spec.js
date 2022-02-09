/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";

import {getJSONTestData, loadFrameworks, testConfig, prepareTestManifest, wait} from "../common";
import {contentTypes, CredentialManager, DIDManager, UniversalWallet, WalletUser} from "../../../src";
import {IssuerAdapter} from "../mocks/adapters";
import jp from "jsonpath";

var uuid = require('uuid/v4')

const WALLET_USER = 'smith-agent'
const WALLET_QUERY_USER = 'smith-query-agent'
const WALLET_DIDCOMM_USER = 'smith-didcomm-agent'
const VC_ISSUER = 'vc-issuer-agent'

const UDC_DESCRIPTOR_ID = "udc_output"
const PRC_DESCRIPTOR_ID = "prc_output"
const VC_FORMAT = "ldp_vc"

let walletUserAgent, issuer, sampleUDC, samplePRC, sampleUDCBBS, sampleFrameDoc, manifestVC

before(async function () {
    this.timeout(0)
    walletUserAgent = await loadFrameworks({name: WALLET_USER})

    issuer = new IssuerAdapter(VC_ISSUER)
    await issuer.init()

    // load sample VCs from testdata
    let udcVC = getJSONTestData('udc-vc.json')
    let prcVC = getJSONTestData('prc-vc.json')
    sampleUDCBBS = getJSONTestData('udc-bbs-vc.json')
    sampleFrameDoc = getJSONTestData('udc-frame.json')
    manifestVC = getJSONTestData('manifest-vc.json')


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

describe('Credential Manager data model tests', async function () {
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
        const sampleUDCManifest = prepareTestManifest('udc-cred-manifest.json')

        await credentialManager.save(auth, {credentials: [sampleUDC]},
            {
                manifest: prepareTestManifest('udc-cred-manifest.json'),
                descriptorMap: [
                    {
                        id: UDC_DESCRIPTOR_ID,
                        format: VC_FORMAT,
                        path: "$[0]"
                    }
                ]
            })
    })

    it('user saves a BBS credential into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        const sampleUDCManifest = getJSONTestData('udc-cred-manifest.json')

        await credentialManager.save(auth, {credentials: [sampleUDCBBS]},
            {
                manifest: prepareTestManifest('udc-cred-manifest.json'),
                descriptorMap: [
                    {
                        id: UDC_DESCRIPTOR_ID,
                        format: VC_FORMAT,
                        path: "$[0]"
                    }
                ]
            })
    })

    it('user saves a credential into wallet by verifying', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        const samplePRCManifest = getJSONTestData('prc-cred-manifest.json')

        await credentialManager.save(auth, {credentials: [samplePRC]}, {
            verify: true,
            manifest: prepareTestManifest('prc-cred-manifest.json'),
            descriptorMap: [
                {
                    id: PRC_DESCRIPTOR_ID,
                    format: VC_FORMAT,
                    path: "$[0]"
                }
            ]
        })
    })

    it('user saves fulfillment presentation into wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        const manifest = getJSONTestData('udc-cred-manifest.json')
        const presentation = getJSONTestData('cred-fulfillment-udc-vp.json')

        await credentialManager.save(auth, {presentation}, {manifest})
    })

    // confirm number of credential manifests saved in DB
    it('user verified all credential related data models exists', async function () {
        const vcwallet = new UniversalWallet({agent: walletUserAgent, user: WALLET_USER});

        let {contents} = await vcwallet.getAll({auth, contentType: contentTypes.METADATA})
        expect(Object.keys(contents)).to.have.lengthOf(4) // 4 credential metadata
    })

    it('user gets credential metadata by credential ID', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let metadata = await credentialManager.getCredentialMetadata(auth, sampleUDC.id)
        expect(metadata).to.not.empty

        expect(metadata.credentialType).to.have.lengthOf(2)
        expect(metadata.name).to.be.equal(sampleUDC.name)
        expect(metadata.description).to.be.equal(sampleUDC.description)
        expect(metadata.expirationDate).to.be.equal(sampleUDC.expirationDate)
        expect(metadata.type).to.be.equal("CredentialMetadata")
        expect(metadata.resolved).to.have.lengthOf(1)
        expect(metadata.resolved[0].properties).to.have.lengthOf(2)

        // resolve credential
        const manifest = getJSONTestData('udc-cred-manifest.json')
        let resolved  = await credentialManager.resolveManifest(auth, {
            credentialID: sampleUDC.id,
            manifest,
            descriptorID: "udc_output",
        })
        expect(resolved).to.have.lengthOf(1)
        expect(resolved[0].properties).to.have.lengthOf(2)

        // resolve credential, using fulfillment & manifest objects
        resolved  = await credentialManager.resolveManifest(auth, {
            manifest: getJSONTestData('udc-cred-manifest.json'),
            fulfillment: getJSONTestData('cred-fulfillment-udc-vp.json'),
        })
        expect(resolved).to.have.lengthOf(1)
        expect(resolved[0].properties).to.have.lengthOf(2)
    })

    it('user updates credential metadata', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})
        const sampleName = "SAMPLE NAME"
        const sampleDescription = "SAMPLE DESCRIPTION"

        await credentialManager.updateCredentialMetadata(auth, sampleUDC.id, {
            name: sampleName,
            description: sampleDescription
        })

        let metadata = await credentialManager.getCredentialMetadata(auth, sampleUDC.id)
        expect(metadata).to.not.empty
        expect(metadata.name).to.be.equal(sampleName)
        expect(metadata.description).to.be.equal(sampleDescription)
    })

    it('user gets all credential metadata', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        // get all credential metadata
        let metadataList = await credentialManager.getAllCredentialMetadata(auth)
        expect(metadataList).to.have.lengthOf(4)
        metadataList.forEach(metadata => {
            expect(metadata.credentialType).to.have.lengthOf(2)
            expect(metadata.type).to.be.equal("CredentialMetadata")
            expect(metadata.resolved).to.not.empty
        })

        // filter credential metadata by credential IDs
        metadataList = await credentialManager.getAllCredentialMetadata(auth, {credentialIDs:[sampleUDC.id, samplePRC.id]})
        expect(metadataList).to.have.lengthOf(2)
        metadataList.forEach(metadata => {
            expect(metadata.credentialType).to.have.lengthOf(2)
            expect(metadata.type).to.be.equal("CredentialMetadata")
            expect(metadata.resolved).to.not.empty
        })

        metadataList = await credentialManager.getAllCredentialMetadata(auth, {credentialIDs:["invalid"]})
        expect(metadataList).to.have.lengthOf(0)

        // get all credential metadata
        metadataList = await credentialManager.getAllCredentialMetadata(auth, {resolve: true})
        expect(metadataList).to.have.lengthOf(4)
        metadataList.forEach(metadata => {
            expect(metadata.credentialType).to.have.lengthOf(2)
            expect(metadata.type).to.be.equal("CredentialMetadata")
            expect(metadata.resolved).to.not.empty
        })
    })

    it('user gets all credentials from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(4)
    })

    it('user gets a credential from wallet by id', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        let {content} = await credentialManager.get(auth, sampleUDC.id)
        expect(content).to.not.empty
    })

    it('user removes a credential from wallet', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_USER})

        await credentialManager.remove(auth, samplePRC.id)

        // get all credentials and verify
        let {contents} = await credentialManager.getAll(auth)
        expect(contents).to.not.empty
        expect(Object.keys(contents)).to.have.lengthOf(3)

        // get all credential metadata. and verify.
        let metadataList = await credentialManager.getAllCredentialMetadata(auth)
        expect(metadataList).to.have.lengthOf(3)
    })

    let did
    it('user creates Orb DID in wallet and resolve it', async function () {
        let didManager = new DIDManager({agent: walletUserAgent, user: WALLET_USER})

        let docres = await didManager.createOrbDID(auth, {purposes: ["assertionMethod", "authentication"]})
        expect(docres.didDocument.id).to.not.empty

        await wait(5000) // TODO trying polling/retry instead of wait

        let resolveDID = await didManager.resolveOrbDID(auth, docres.didDocument.id)

        expect(resolveDID.didDocument.id).to.not.empty

        did = resolveDID.didDocument.id
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
            { manifest, descriptorMap})

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
                    "schema": [{"uri": "https://www.w3.org/2018/credentials#VerifiableCredential"}],
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
                    "schema": [{"uri": "https://www.w3.org/2018/credentials#VerifiableCredential"}],
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

        // getting metadata of query results
        const credentialIDs = jp.query(results, '$[*].verifiableCredential[*].id');
        const metadataList = await Promise.all(credentialIDs.map(async id => await credentialManager.getCredentialMetadata(auth, id)))

        expect(metadataList).to.not.empty
        for (const metadata of metadataList) {
            expect(metadata).to.not.empty
            expect(metadata.resolved).to.not.empty
        }

    })
})

describe('Credential Manager blinded routing Manifest Credential Tests', async function () {
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

        await credentialManager.saveManifestVC(auth, manifestVC, connectionID)
    })

    it('user gets connection info from manifest', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        let mconnectionID = await credentialManager.getManifestConnection(auth, manifestVC.id)
        expect(mconnectionID).to.be.equal(connectionID)
    })

    // TODO currently failing due to document loader issue.
    it.skip('user queries manifest credentials', async function () {
        let credentialManager = new CredentialManager({agent: walletUserAgent, user: WALLET_DIDCOMM_USER})

        // save one more
        let manifest2 = JSON.parse(JSON.stringify(manifestVC))
        manifest2.id = `http://example.gov/credentials/${uuid()}`
        await credentialManager.saveManifestVC(auth, manifest2, uuid())

        let {contents} = await credentialManager.getAllManifestVCs(auth)
        expect(Object.keys(contents)).to.have.lengthOf(2)
    })
})
