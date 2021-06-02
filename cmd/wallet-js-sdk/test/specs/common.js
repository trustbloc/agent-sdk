/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as Agent from "@trustbloc/agent-sdk-web"

export const testConfig = window.__ini__ ? window.__ini__['test/fixtures/config.ini'] : {}
console.debug('test configuration:', JSON.stringify(testConfig, null, 2))
const {agentStartupOpts} = testConfig

// loadFrameworks loads agent instance
export async function loadFrameworks({name = 'user-agent', logLevel = ''} = {}) {
    let agentOpts = JSON.parse(JSON.stringify(agentStartupOpts))
    agentOpts["indexedDB-namespace"] = `${name}db`
    agentOpts["agent-default-label"] = `${name}-wallet-web`

    if (logLevel) {
        agentOpts["log-level"] = logLevel
    }

    return new Agent.Framework(agentOpts)
}

