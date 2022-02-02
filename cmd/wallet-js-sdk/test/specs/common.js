/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import * as Agent from "@trustbloc/agent-sdk-web"

var uuid = require('uuid/v4')

export const DIDEXCHANGE_STATE_TOPIC = "didexchange_states"
export const POST_STATE = "post_state"
export const DIDEXCHANGE_STATE_REQUESTED = "requested"
export const DIDEXCHANGE_STATE_COMPLETED = "completed"
export const PRESENT_PROOF_ACTION_TOPIC = "present-proof_actions"
export const ISSUE_CREDENTIAL_STATE_TOPIC = "issue-credential_states"
export const ISSUE_CREDENTIAL_ACTION_TOPIC = "issue-credential_actions"
export const MSG_TYPE_OFFER_CREDENTIAL_V2 = "https://didcomm.org/issue-credential/2.0/offer-credential"
export const MSG_TYPE_OFFER_CREDENTIAL_V3 = "https://didcomm.org/issue-credential/3.0/offer-credential"
export const MSG_TYPE_PROPOSE_CREDENTIAL_V2 = "https://didcomm.org/issue-credential/2.0/propose-credential"
export const MSG_TYPE_PROPOSE_CREDENTIAL_V3 = "https://didcomm.org/issue-credential/3.0/propose-credential"
export const ATTACH_FORMAT_CREDENTIAL_MANIFEST = "dif/credential-manifest/manifest@v1.0"
export const ATTACH_FORMAT_CREDENTIAL_FULFILLMENT = "dif/credential-manifest/fulfillment@v1.0"
export const ATTACH_FORMAT_ISSUE_CREDENTIAL = "aries/ld-proof-vc@v1.0"

export const testConfig = window.__ini__ ? window.__ini__['test/fixtures/config.ini'] : {}
testConfig.walletUserPassphrase = uuid()
console.debug('test configuration:', JSON.stringify(testConfig, null, 2))
const {agentStartupOpts} = testConfig

// loads testdata from fixtures and returns string response.
export function getTestData(filename) {
    return window.__FIXTURES__[`test/fixtures/testdata/${filename}`]
}

// loads testdata from fixtures and returns JSON parsed response.
export function getJSONTestData(filename) {
    return JSON.parse(window.__FIXTURES__[`test/fixtures/testdata/${filename}`])
}

// loadFrameworks loads agent instance
export async function loadFrameworks({
      name = 'user-agent',
      logLevel = '',
      mediaTypeProfiles = ["didcomm/aip2;env=rfc19"],
      keyType = 'ed25519',
      keyAgreementType = 'p256kw'} = {}) {
    let agentOpts = JSON.parse(JSON.stringify(agentStartupOpts))
    agentOpts["indexedDB-namespace"] = `${name}db`
    agentOpts["agent-default-label"] = `${name}-wallet-web`
    agentOpts["media-type-profiles"] = mediaTypeProfiles
    agentOpts["key-type"] = keyType
    agentOpts["key-agreement-type"] = keyAgreementType

    if (logLevel) {
        agentOpts["log-level"] = logLevel
    }

    return new Agent.Framework(agentOpts)
}

export function wait(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

export const retryWithDelay = async (
    fn, retries = 3, interval = 50,
    finalErr = Error('retries exhausted.')
) => {
    try {
        await fn()
    } catch (err) {
        if (retries <= 0) {
            return Promise.reject(finalErr);
        }
        await wait(interval)
        return retryWithDelay(fn, (retries - 1), interval, finalErr);
    }
}

// read manifest and replace manifest ID so that same file can be reused for many tests.
export const prepareTestManifest = (file) => {
    const manifest = getJSONTestData(file)
    manifest.id = uuid()

    return manifest
}

