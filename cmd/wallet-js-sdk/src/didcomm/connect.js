/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {POST_STATE, waitForEvent} from "../util/event.js";
import axios from 'axios';

const STATE_COMPLETED = 'completed'
const DID_EXCHANGE_STATES_TOPIC = 'didexchange_states'
const STATE_COMPLETE_MSG_TOPIC = 'didexchange-state-complete'
const STATE_COMPLETE_MSG_TYPE = 'https://trustbloc.dev/didexchange/1.0/state-complete'
const DEFAULT_LABEL = 'agent-default-label'
const ROUTER_CREATE_INVITATION_PATH = `/didcomm/invitation`

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

    //TODO save connection in wallet store
    async connect(invitation, {waitForCompletion = '', label = ''} = {}) {
        let conn = await this.agent.outofband.acceptInvitation({
            my_label: label ? label : DEFAULT_LABEL,
            invitation: invitation,
            router_connections: await getMediatorConnections(this.agent, {single: true}),
        })

        let connectionID = conn['connection_id']

        let checked = await waitForEvent(this.agent, {
            type: POST_STATE,
            stateID: STATE_COMPLETED,
            topic: DID_EXCHANGE_STATES_TOPIC,
            connectionID
        })

        console.log('go connection completed event for connection', connectionID)

        const record = await this.agent.didexchange.queryConnectionByID({id: connectionID})

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

/**
 * Get DID Invitation from edge router.
 *
 * @param endpoint edge router endpoint
 */
export const createInvitationFromRouter = async (endpoint) => {
    let response = await axios.get(`${endpoint}${ROUTER_CREATE_INVITATION_PATH}`)
    return response.data.invitation
}


/**
 * Connect given agent to edge mediator/router.
 *
 * @param agent trustbloc agent
 * @param endpoint edge router endpoint
 * @param wait for did exchange state complete message
 */
export async function connectToMediator(agent, mediatorEndpoint, {waitForStateComplete = true} = {}) {
    let resp = await agent.mediatorclient.connect({
        myLabel: 'agent-default-label',
        invitation: await createInvitationFromRouter(mediatorEndpoint),
        stateCompleteMessageType: waitForStateComplete ? STATE_COMPLETE_MSG_TYPE : ''
    })

    if (resp.connectionID) {
        console.log("router registered successfully!", resp.connectionID)
    } else {
        console.log("router was not registered!")
    }
}
