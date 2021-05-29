/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as Agent from "@trustbloc-cicd/agent-sdk-web"

// TODO endpoints should be read from configurations
const agentStartupOpts = {
    assetsPath: "/base/public/agent-js-worker/assets",
    'outbound-transport': ['ws', 'http'],
    'transport-return-route': 'all',
    // "http-resolver-url": ["trustbloc@http://localhost:9080/1.0/identifiers", "v1@http://localhost:9080/1.0/identifiers"],
    'agent-default-label': 'wallet-test-agent',
    'auto-accept': true,
    'log-level': 'debug',
    'indexedDB-namespace': 'agent',
    storageType: `indexedDB`,
    edvServerURL: '',
    didAnchorOrigin: 'origin'
}


// TODO move to config file
export const testConfig = {
    mediatorEndPoint : "https://localhost:10093"
}

export async function loadFrameworks({name = 'user-agent', logLevel= 'debug'}) {
    let agentOpts = agentStartupOpts

    if (name) {
        agentOpts = JSON.parse(JSON.stringify(agentStartupOpts))
        agentOpts["indexedDB-namespace"] = `${name}db`
        agentOpts["agent-default-label"] = `${name}-wallet-web`
        agentOpts["log-level"] = logLevel
        agentOpts["auto-accept"] = true
    }

    return new Agent.Framework(agentOpts)
}

