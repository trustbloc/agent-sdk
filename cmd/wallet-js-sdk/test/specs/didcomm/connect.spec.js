/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";
import {loadFrameworks, testConfig} from "../common";
import {connectToMediator, DIDExchange, getMediatorConnections, waitForEvent} from "../../../src";

const ALICE_LABEL = "alice-agent"
const BOB_LABEL = "bob-agent"
const DIDEX_STATE_TOPIC = "didexchange_states"
const POST_STATE = "post_state"
const REQUESTED_STATE = "requested"

let alice, bob

before(async function () {
    alice = await loadFrameworks({name: ALICE_LABEL, logLevel: 'debug'})
    bob = await loadFrameworks({name: BOB_LABEL, logLevel: 'debug'})
});

after(function () {
    alice ? alice.destroy() : ''
    bob ? bob.destroy() : ''
});

describe('running connect test 123', async function () {
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

    it('bob connected to mediator', async function () {
        try {
            await connectToMediator(bob, testConfig.mediatorEndPoint)
            let conns = await getMediatorConnections(bob)
            expect(conns).to.not.empty
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



    it('alice connects to bob', async function () {
        try {
            // listen for exchange request and accept it
            acceptExchangeRequest(alice)

            let didexchange = new DIDExchange(bob)
            let res = await didexchange.connect(invitation, {label: BOB_LABEL})

        } catch (e) {
            console.error('alice fails to connect to bob ',e)
            expect.fail(e);
        }
    });
});


async function acceptExchangeRequest(agent) {
    return waitForEvent(agent, {
        stateID: 'requested',
        type: POST_STATE,
        topic: DIDEX_STATE_TOPIC,
        callback: async (payload) => {
            await agent.didexchange.acceptExchangeRequest({
                id: payload.Properties.connectionID,
                router_connections: await getMediatorConnections(agent, {single:true}),
            })
        }
    })
}
