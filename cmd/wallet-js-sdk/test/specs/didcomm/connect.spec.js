/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";
import {loadFrameworks, testConfig, DIDEXCHANGE_STATE_TOPIC, POST_STATE, DIDEXCHANGE_STATE_REQUESTED} from "../common";
import {
    connectToMediator,
    createWalletProfile,
    DIDComm,
    getMediatorConnections,
    UniversalWallet,
    waitForEvent
} from "../../../src";

const ALICE_LABEL = "alice-agent"
const BOB_LABEL = "bob-agent"

let alice, bob

before(async function () {
    alice = await loadFrameworks({name: ALICE_LABEL})
    bob = await loadFrameworks({name: BOB_LABEL})
});

after(function () {
    alice ? alice.destroy() : ''
    bob ? bob.destroy() : ''
});

describe('running DIDComm connection tests', async function () {
    it('alice connected to mediator', async function () {
        try {
            await connectToMediator(alice, testConfig.mediatorEndPoint)
            let conns = await getMediatorConnections(alice)
            expect(conns).to.not.empty
        } catch (e) {
            console.error('failed to connect alice to mediator ',e)
            expect.fail(e);
        }
    });

    let routerConnections
    it('bob connected to mediator', async function () {
        try {
            await connectToMediator(bob, testConfig.mediatorEndPoint)
            let conns = await getMediatorConnections(bob, {single: true})
            expect(conns).to.not.empty

            routerConnections = [conns]
        } catch (e) {
            console.error('failed to connect bob to mediator ',e)
            expect.fail(e);
        }
    });

    let invitation
    it('alice creates invitation for bob', async function () {
        try {
            let res = await alice.outofband.createInvitation({
                label: ALICE_LABEL,
                router_connection_id: await getMediatorConnections(alice, {single:true})
            })

             invitation = res.invitation

            expect(invitation).to.not.empty
        } catch (e) {
            console.error('failed to create invitation from alice agent ',e)
            expect.fail(e);
        }
    });

    it('bob creates his wallet profile', async function () {
        await createWalletProfile(bob, BOB_LABEL, {localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('bob opens his wallet', async function () {
        let wallet = new UniversalWallet({agent: bob, user: BOB_LABEL})
        let authResponse = await wallet.open({localKMSPassphrase: testConfig.walletUserPassphrase})
        expect(authResponse.token).to.not.empty
        auth = authResponse.token
    })

    it('bob connect to alice by accepting invitation', async function () {
        try {
            // listen for exchange request and accept it
            acceptExchangeRequest(alice)

            let didcomm = new DIDComm({agent: bob, user: BOB_LABEL})
            let res = await didcomm.connect(auth, invitation, {myLabel: BOB_LABEL, routerConnections})

        } catch (e) {
            console.error('alice fails to connect to bob ',e)
            expect.fail(e);
        }
    });
});


async function acceptExchangeRequest(agent) {
    return waitForEvent(agent, {
        stateID: DIDEXCHANGE_STATE_REQUESTED,
        type: POST_STATE,
        topic: DIDEXCHANGE_STATE_TOPIC,
        callback: async (payload) => {
            await agent.didexchange.acceptExchangeRequest({
                id: payload.Properties.connectionID,
                router_connections: await getMediatorConnections(agent, {single:true}),
            })
        }
    })
}
