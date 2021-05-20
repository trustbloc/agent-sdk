/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/


import {POST_STATE, waitForEvent} from "../util/event.js";

const STATE_COMPLETED = 'completed'
const DID_EXCHANGE_STATES_TOPIC = 'didexchange_states'
const STATE_COMPLETE_MSG_TOPIC = 'didexchange-state-complete'
const STATE_COMPLETE_MSG_TYPE = 'https://trustbloc.dev/didexchange/1.0/state-complete'
const DEFAULT_LABEL = 'agent-default-label'

/**
 * DIDExchange provides aries DID exchange features.
 *
 * @param agent instance
 * @class
 */
export class DIDExchange {
    constructor(agent) {
        this.agent = agent
    }

    async connect(invitation, {waitForCompletion = ''} = {}) {
        let conn = await this.agent.outofband.acceptInvitation({
            my_label: DEFAULT_LABEL,
            invitation: invitation,
            router_connections: await getMediatorConnections(this.agent, {single:true}),
        })

        let connID = conn['connection_id']

        await waitForEvent(this.agent, {
            type: POST_STATE,
            stateID: STATE_COMPLETED,
            connectionID: connID,
            topic: DID_EXCHANGE_STATES_TOPIC,
        })

        const record = await this.agent.didexchange.queryConnectionByID({id: connID})

        if (waitForCompletion) {
            this.agent.messaging.registerService({
                "name": STATE_COMPLETE_MSG_TOPIC,
                "type": STATE_COMPLETE_MSG_TYPE,
            })

            try {
                await new Promise((resolve, reject) => {
                    setTimeout(() => reject(new Error("time out waiting for state complete message")), 15000)
                    const stop = this.agent.startNotifier(msg => {
                        if (record.result.MyDID == msg.payload.mydid && record.result.TheirDID == msg.payload.theirdid) {
                            stop()
                            console.debug('received state complete msg received.')
                            resolve(msg.payload.message)
                        }
                    }, [STATE_COMPLETE_MSG_TOPIC])
                })

            } catch (e) {
                console.warn('error while waiting for state complete msg !!', e)
            }
        }

        return record
    }
}

/**
 * Get router/mediator connections from agent.
 *
 * @param agent instance
 */
export async function getMediatorConnections(agent, {single} = {}) {
    let resp = await agent.mediator.getConnections()
    if (!resp.connections || resp.connections.length === 0) {
        return ""
    }

    if (single) {
        return resp.connections[Math.floor(Math.random() * resp.connections.length)]
    }

    return resp.connections.join(",");
}



