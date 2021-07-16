/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

export const PRE_STATE = "pre_state"
export const POST_STATE = "post_state"

const DEFAULT_TIMEOUT = 120000
const DEFAULT_TIMEOUT_ERR = "time out while waiting for event"
const DEFAULT_TOPIC = 'all'

export function waitForEvent(agent, {
    timeout = DEFAULT_TIMEOUT,
    timeoutError = DEFAULT_TIMEOUT_ERR,
    topic = DEFAULT_TOPIC,
    type, stateID, connectionID, callback = () => {}} = {}) {


    return new Promise((resolve, reject) => {
        setTimeout(() => reject(new Error(timeoutError)), timeout)
        const stop = agent.startNotifier(event => {
            try {
                let payload = event.payload;

                if (connectionID && payload.Properties &&
                    payload.Properties.connectionID !== connectionID) {
                    return
                }

                if (stateID && payload.StateID !== stateID) {
                    return
                }

                if (type && payload.Type !== type) {
                    return
                }

                stop()

                callback(payload)

                resolve(payload)
            } catch (e) {
                stop()
                reject(e)
            }
        }, [topic])
    })
}

// filter and return defined properties only
export const definedProps = obj => Object.fromEntries(
    Object.entries(obj).filter(([k, v]) => v !== undefined)
);
