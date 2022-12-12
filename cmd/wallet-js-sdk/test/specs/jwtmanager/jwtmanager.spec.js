/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { expect } from "chai";
import { loadFrameworks, testConfig } from "../common";
import { createWalletProfile, DIDManager, JWTManager, UniversalWallet } from "@";

const RICK_USER = 'rick-agent-jwt'

let rick

before(async function () {
    rick = await loadFrameworks({
        name: RICK_USER,
        enableDIDComm: false,
        contextProviderURL: ["http://localhost:10096/agent-startup-contexts.json"]
    });
});

after(function () {
    rick ? rick.destroy() : ''
});


describe('JWT Manager tests', async function () {
    it('rick creates his wallet profile', async function () {
        await createWalletProfile(rick, RICK_USER, {localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    let auth, kid, jwtManager, jwt;
    it('rick opens wallet', async function () {
        const wallet = new UniversalWallet({agent: rick, user: RICK_USER})
        const authResponse = await wallet.open({localKMSPassphrase: testConfig.walletUserPassphrase})
        expect(authResponse.token).to.not.empty
        auth = authResponse.token
    })

    it('rick creates wallet DID', async function () {
        const didManager = new DIDManager({agent: rick, user: RICK_USER});
        const docRes = await didManager.createOrbDID(auth, {
            purposes: ["assertionMethod", "authentication"],
            keyType: 'ED25519',
            signType: 'Ed25519VerificationKey2018',
        })
        kid = docRes.didDocument.id;
    })

    it('rick initializes jwt manager', async function () {
        jwtManager = new JWTManager({agent: rick, user: RICK_USER})
        expect(jwtManager).to.be.an.instanceof(JWTManager);
    })

    it('rick signs jwt successfully', async function () {
        const result = await jwtManager.signJWT(auth, { headers: {}, claims: { foo: 'bar' }, kid });
        expect(result).to.be.an.instanceof(Object);
        expect(result).to.have.property('jwt');
        expect(result.jwt).to.have.a.lengthOf.above(0);
        jwt = result.jwt;
    })

    it('rick fails to verify invalid jwt', async function () {
        const result = await jwtManager.verifyJWT(auth, { jwt: 'foo.bar.baz' });
        expect(result).to.be.an.instanceof(Object);
        expect(result).to.have.property('verified');
        expect(result.verified).to.equal(false);
        expect(result).to.have.property('error');
        expect(result.error).to.have.a.lengthOf.above(0);
    })

    it('rick verifies jwt successfully', async function () {
        const result = await jwtManager.verifyJWT(auth, { jwt });
        expect(result).to.be.an.instanceof(Object);
        expect(result).to.have.property('verified');
        expect(result.verified).to.equal(true);
    })
})
