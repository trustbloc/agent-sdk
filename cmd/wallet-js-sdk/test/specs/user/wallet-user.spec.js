/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {expect} from "chai";
import {loadFrameworks, testConfig} from "../common";
import {WalletUser} from "../../../src";

var uuid = require('uuid/v4')

const JOHN_USER = 'john-agent'

let john

before(async function () {
    john = await loadFrameworks({name: JOHN_USER})
});

after(function () {
    john ? john.destroy() : ''
});


describe('Wallet user tests', async function () {
    let preferences = {
        name: 'Mr. John Smith',
        description: 'preferences for my wallet',
        image: 'https://via.placeholder.com/150',
        controller: 'did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5',
        verificationMethod: 'did:key:z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5#z6MknC1wwS6DEYwtGbZZo2QvjQjkh2qSBjb4GYmbye8dv4S5',
        proofType: 'Ed25519Signature2018'
    }

    it('john checks if his wallet profile exists', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        let found = await walletUser.profileExists()
        expect(found).to.be.false
    })

    it('john creates his wallet profile', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        await walletUser.createWalletProfile({localKMSPassphrase: testConfig.walletUserPassphrase})
    })

    it('john again checks if his wallet profile exists', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        let found = await walletUser.profileExists()
        expect(found).to.be.true
    })

    let auth
    it('john opens wallet', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        let authResponse = await walletUser.unlock({localKMSPassphrase: testConfig.walletUserPassphrase})

        expect(authResponse.token).to.not.empty

        auth = authResponse.token
    })

    it('john saves his wallet preferences', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        await walletUser.savePreferences(auth, preferences)
    })

    it('john reads his wallet preferences', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        let {content} = await walletUser.getPreferences(auth)
        expect(preferences.name).to.be.equal(content.name)
        expect(preferences.description).to.be.equal(content.description)
        expect(preferences.image).to.be.equal(content.image)
        expect(preferences.controller).to.be.equal(content.controller)
        expect(preferences.verificationMethod).to.be.equal(content.verificationMethod)
        expect(preferences.proofType).to.be.equal(content.proofType)
    })

    it('john updates wallet preferences', async function () {
        const newDescription = 'updated description...'

        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        await walletUser.updatePreferences(auth, {description: newDescription, verificationMethod: ""})

        let {content} = await walletUser.getPreferences(auth)
        expect(preferences.name).to.be.equal(content.name)
        expect(newDescription).to.be.equal(content.description)
        expect(preferences.image).to.be.equal(content.image)
        expect(preferences.controller).to.be.equal(content.controller)
        expect(content.verificationMethod).to.be.empty
        expect(preferences.proofType).to.be.equal(content.proofType)
    })

    it('john saves his custom metadata into wallet', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})

        let custom = {
            '@context': 'https://w3id.org/wallet/v1',
            id: uuid(), name: uuid()
        }

        await walletUser.saveMetadata(auth, custom)

        let {content} = await walletUser.getMetadata(auth, custom.id)

        expect(custom.id).to.be.equal(content.id)
        expect(custom.name).to.be.equal(content.name)
    })

    it('john gets all metadata from wallet', async function () {
        let walletUser = new WalletUser({agent: john, user: JOHN_USER})


        let {contents} = await walletUser.getAllMetadata(auth)
        expect(Object.keys(contents)).to.have.lengthOf(2)
    })

})
