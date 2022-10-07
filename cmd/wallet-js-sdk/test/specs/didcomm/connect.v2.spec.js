/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";
import {
    loadFrameworks,
    testConfig,
    DIDEXCHANGE_STATE_TOPIC,
    POST_STATE,
    DIDEXCHANGE_STATE_REQUESTED,
    wait
} from "../common";
import {
    connectToMediator,
    createWalletProfile,
    DIDComm, DIDManager,
    getMediatorConnections,
    UniversalWallet,
    waitForEvent
} from "../../../src";

const ALICE_LABEL_V2 = "alice-agent-v2"
const BOB_LABEL_V2 = "bob-agent-v2"

let alice, bob
let aliceDID

before(async function () {
    alice = await loadFrameworks({
      name: ALICE_LABEL_V2,
      mediaTypeProfiles: ["didcomm/v2"],
      keyType: "ecdsap256ieee1363",
      keyAgreementType: "p256kw",
      enableDIDComm: true,
      contextProviderURL: ["http://localhost:10096/agent-startup-contexts.json"]
    });
    bob = await loadFrameworks({
      name: BOB_LABEL_V2,
      mediaTypeProfiles: ["didcomm/v2"],
      keyType: "ecdsap256ieee1363",
      keyAgreementType: "p256kw",
      enableDIDComm: true,
      contextProviderURL: ["http://localhost:10096/agent-startup-contexts.json"]
    });

    aliceDID = await createWalletAndPublicDID(alice, ALICE_LABEL_V2, aliceDID)
    // await createWalletAndPublicDID(bob, BOB_LABEL_V2, bobDID)
});

async function createWalletAndPublicDID(agent, agentLabel, agentDID) {
    try {
        await createWalletProfile(agent, agentLabel, {localKMSPassphrase: testConfig.walletUserPassphrase})

        let wallet = new UniversalWallet({agent: agent, user: agentLabel})
        let authResponse = await wallet.open({localKMSPassphrase: testConfig.walletUserPassphrase})
        expect(authResponse.token).to.not.empty

        let auth = authResponse.token
        let didManager = new DIDManager({agent: agent, user: agentLabel})
        let document = await didManager.createOrbDID(auth)

        expect(document).to.not.empty

        agentDID = document.didDocument.id

        expect(agentDID).to.not.empty

    } catch (e) {
        console.error('failed to create wallet and peer DID document from ${agentLabel} agent ',e)
        expect.fail(e);
    }

    return agentDID
}

after(function () {
    alice ? alice.destroy() : ''
    bob ? bob.destroy() : ''
});

describe('running DIDComm V2 connection tests', async function () {
    it('alice connected to V2 mediator', async function () {
        try {
            await connectToMediator(alice, testConfig.mediatorEndPoint, {isDIDCommV2: true})
            let conns = await getMediatorConnections(alice)
            expect(conns).to.not.empty
        } catch (e) {
            console.error('failed to connect alice to mediator ',e)
            expect.fail(e);
        }
    });

    let routerConnections
    it('bob connected to V2 mediator', async function () {
        try {

            await connectToMediator(bob, testConfig.mediatorEndPoint,{isDIDCommV2: true})
            let conns = await getMediatorConnections(bob, {single: true})
            expect(conns).to.not.empty

            routerConnections = [conns]
        } catch (e) {
            console.error('failed to connect bob to mediator ',e)
            expect.fail(e);
        }
    });

    let invitation
    it('alice creates oobv2 invitation for bob', async function () {
        try {
            let res = await alice.outofbandv2.createInvitation({
                label: ALICE_LABEL_V2,
                body: {accept: ["didcomm/v2"]},
                from: aliceDID,
            })

             invitation = res.invitation
            expect(invitation).to.not.empty
        } catch (e) {
            console.error('failed to create invitation from alice agent ',e)
            expect.fail(e);
        }
    });

    it('bob connect to alice by accepting invitation', async function () {
        try {
            await wait(5000)

            await bob.outofbandv2.acceptInvitation({
                my_label: BOB_LABEL_V2,
                invitation: invitation,
            })
        } catch (e) {
            console.error('bob fails to accept alice\'s invitation and connect',e)
            expect.fail(e);
        }
    });
});
