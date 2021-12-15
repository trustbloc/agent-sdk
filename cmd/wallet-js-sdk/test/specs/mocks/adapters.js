/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {connectToMediator, getMediatorConnections, waitForEvent, findAttachmentByFormat} from "../../../src";

import {
    DIDEXCHANGE_STATE_REQUESTED,
    DIDEXCHANGE_STATE_TOPIC,
    loadFrameworks,
    POST_STATE,
    PRESENT_PROOF_ACTION_TOPIC,
    ISSUE_CREDENTIAL_ACTION_TOPIC,
    MSG_TYPE_OFFER_CREDENTIAL_V2,
    ATTACH_FORMAT_CREDENTIAL_MANIFEST,
    ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
    ATTACH_FORMAT_ISSUE_CREDENTIAL,
    retryWithDelay,
    testConfig
} from "../common";

import {expect} from "chai";

var uuid = require('uuid/v4')


/**
 * Adapter mocks common issuer or rp adapter features
 *
 * @param agent instance
 * @class
 */
export class Adapter {
    constructor(label) {
        this.label = label
    }

    async init() {
        this.agent = await loadFrameworks({name: this.label})

        await connectToMediator(this.agent, testConfig.mediatorEndPoint)

        let conns = await getMediatorConnections(this.agent)
        expect(conns).to.not.empty
    }

    async createInvitation({goal_code}={}) {
        let response = await this.agent.mediatorclient.createInvitation({
            label: this.label,
            router_connection_id: await getMediatorConnections(this.agent, {single: true}),
            goal_code,
        })

        return response.invitation
    }

    async acceptExchangeRequest(timeout) {
        return await waitForEvent(this.agent, {
            stateID: DIDEXCHANGE_STATE_REQUESTED,
            type: POST_STATE,
            topic: DIDEXCHANGE_STATE_TOPIC,
            timeout,
            callback: async (payload) => {
                await this.agent.didexchange.acceptExchangeRequest({
                    id: payload.Properties.connectionID,
                    router_connections: await getMediatorConnections(this.agent, {single: true}),
                })
            }
        })
    }

    async destroy() {
        return await this.agent.destroy()
    }
}

/**
 * VerifierAdapter mocks verifier(relying party) adapter features.
 *
 * @param agent instance
 * @class
 */
export class VerifierAdapter extends Adapter {
    constructor(label) {
        super(label)
    }

    async init() {
        return await super.init()
    }

    async acceptPresentationProposal(query = {}, timeout) {
        return await waitForEvent(this.agent, {
            topic: PRESENT_PROOF_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                let id = uuid()
                let {myDID, theirDID, piid} = payload.Properties
                await this.agent.presentproof.sendRequestPresentation({
                    my_did: myDID,
                    their_did: theirDID,
                    request_presentation: {
                        will_confirm: true,
                        formats: [
                            {
                                attach_id: id,
                                format: "dif/presentation-exchange/definitions@v1.0",
                            },
                        ],
                        "request_presentations~attach": [
                            {
                                "@id": id,
                                lastmod_time: "0001-01-01T00:00:00Z",
                                data: {
                                    json: {
                                        presentation_definition: query,
                                    },
                                },
                            },
                        ],
                    },
                });
            }
        })
    }

    async acceptPresentProof({timeout, redirectURL} = {}) {
        let presentation
        await waitForEvent(this.agent, {
            topic: PRESENT_PROOF_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                let {Message, Properties} = payload

                presentation = Message["presentations~attach"][0].data.json
                const { piid } = Properties

                return this.agent.presentproof.acceptPresentation({
                    piid, redirectURL
                });
            }
        })

        return presentation
    }

    async declinePresentProof({timeout, redirectURL} = {}) {
        let presentation
        await waitForEvent(this.agent, {
            topic: PRESENT_PROOF_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                let {Message, Properties} = payload

                presentation = Message["presentations~attach"][0].data.json
                const { piid } = Properties

                return this.agent.presentproof.declinePresentation({
                    piid, redirectURL, reason: "test"
                });
            }
        })

        return presentation
    }
}

/**
 * IssuerAdapter mocks issuer adapter features.
 *
 * @param agent instance
 * @class
 */
export class IssuerAdapter extends Adapter {
    constructor(label) {
        super(label)
    }

    async init() {
        return await super.init()
    }

    async issue(...credential) {
        const keyType = 'ED25519'

        const [keySet, recoveryKeySet, updateKeySet] = await Promise.all([
            this.agent.kms.createKeySet({keyType}),
            this.agent.kms.createKeySet({keyType}),
            this.agent.kms.createKeySet({keyType})
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

        let {didDocument} = await this.agent.didclient.createOrbDID(createDIDRequest)

        let resolveDID = async () => await this.agent.vdr.resolveDID({id: didDocument.id})
        await retryWithDelay(resolveDID, 10, 5000)


        let signVCs = await Promise.all(credential.map(async (credential) => {
                let {verifiableCredential} = await this.agent.verifiable.signCredential({
                    credential,
                    "did": didDocument.id,
                    "signatureType": "Ed25519Signature2018"
                })

                return verifiableCredential
            }
        ))

        return signVCs
    }

    async acceptCredentialProposal({comment, manifest, fulfillment} = {}, timeout) {
        return await waitForEvent(this.agent, {
            topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                let {piid} = payload.Properties
                let attachID1 = uuid()
                let attachID2 = uuid()

                let formats = []
                let attachments = []

                if (manifest) {
                    let attachId = uuid()
                    formats.push({
                        "attach_id": attachId,
                        "format": ATTACH_FORMAT_CREDENTIAL_MANIFEST
                    })
                    attachments.push({
                        "@id": attachId,
                        "mime-type": "application/json",
                        data: {
                            json: manifest
                        }
                    },)
                }

                if (fulfillment) {
                    let attachId = uuid()
                    formats.push({
                        "attach_id": attachId,
                        "format": ATTACH_FORMAT_CREDENTIAL_FULFILLMENT
                    })
                    attachments.push({
                        "@id": attachId,
                        "mime-type": "application/json",
                        data: {
                            json: fulfillment
                        }
                    },)
                }

                await this.agent.issuecredential.acceptProposal({
                    piid,
                    offer_credential: {
                        "@type": MSG_TYPE_OFFER_CREDENTIAL_V2,
                        comment,
                        formats,
                        "offers~attach": attachments,
                    }
                });
            }
        })
    }

    async acceptRequestCredential({timeout, credential, redirect} = {}) {
        let attachment
        await waitForEvent(this.agent, {
            topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                const {Message, Properties, formats} = payload
                const { piid } = Properties

                attachment = findAttachmentByFormat(Message.formats, Message["requests~attach"],"application/ld+json")

                let attachID = uuid()
                let icFormats = []
                let icAttachments = []

                if (credential) {
                    icFormats.push({
                        "attach_id": attachID,
                        "format": ATTACH_FORMAT_ISSUE_CREDENTIAL
                    })

                    icAttachments.push({
                        "@id": attachID,
                        "mime-type": "application/ld+json",
                        data: {
                            json: credential
                        }
                    },)
                }

                return this.agent.issuecredential.acceptRequest({
                    piid,
                    issue_credential: {
                        "@type": "https://didcomm.org/issue-credential/2.0/issue-credential",
                        formats: icFormats,
                        "credentials~attach": icAttachments,
                        "~web-redirect": {
                            status: "OK",
                            url: redirect
                        },
                    }
                });
            }
        })

        return attachment
    }

    async declineCredentialProposal({redirectURL} = {}, timeout) {
        return await waitForEvent(this.agent, {
            topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                let {piid} = payload.Properties

                await this.agent.issuecredential.declineProposal({
                    piid,
                    redirectURL
                });
            }
        })
    }

    async declineRequestCredential({redirectURL} = {}, timeout) {
        await waitForEvent(this.agent, {
            topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
            timeout,
            callback: async (payload) => {
                const { piid } = payload.Properties

                return this.agent.issuecredential.declineRequest({
                    piid,
                    redirectURL
                });
            }
        })
    }

}
