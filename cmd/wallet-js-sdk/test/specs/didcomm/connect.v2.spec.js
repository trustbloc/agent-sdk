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

before(async function () {
    alice = await loadFrameworks({name: ALICE_LABEL_V2})
    bob = await loadFrameworks({name: BOB_LABEL_V2})
});

after(function () {
    alice ? alice.destroy() : ''
    bob ? bob.destroy() : ''
});

describe('running DIDComm V2 connection tests', async function () {
    it('alice connected to V2 mediator', async function () {
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
    let aliceOrbDID
    it('alice creates oobv2 invitation for bob', async function () {
        try {
            await createWalletProfile(alice, ALICE_LABEL_V2, {localKMSPassphrase: testConfig.walletUserPassphrase})

            let wallet = new UniversalWallet({agent: alice, user: ALICE_LABEL_V2})
            let authResponse = await wallet.open({localKMSPassphrase: testConfig.walletUserPassphrase})
            expect(authResponse.token).to.not.empty

            let auth = authResponse.token
            let didManager = new DIDManager({agent: alice, user: ALICE_LABEL_V2})
            let document = await didManager.createOrbDID(auth)

            expect(document).to.not.empty

            aliceOrbDID = document.didDocument.id

            expect(aliceOrbDID).to.not.empty
        } catch (e) {
            console.error('failed to create wallet and peer DID document from alice agent ',e)
            expect.fail(e);
        }

        try {
            let res = await alice.outofbandv2.createInvitation({
                label: ALICE_LABEL_V2,
                body: {accept: ["didcomm/aip2;env=rfc19"]},
                from: aliceOrbDID,
            })

             invitation = res.invitation
            expect(invitation).to.not.empty
        } catch (e) {
            console.error('failed to create invitation from alice agent ',e)
            expect.fail(e);
        }
    });

    it('bob creates his V2 wallet profile', async function () {
        await createWalletProfile(bob, BOB_LABEL_V2, {localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth
    it('bob opens his wallet', async function () {
        let wallet = new UniversalWallet({agent: bob, user: BOB_LABEL_V2})
        let authResponse = await wallet.open({localKMSPassphrase: testConfig.walletUserPassphrase})
        expect(authResponse.token).to.not.empty
        auth = authResponse.token
    })

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
