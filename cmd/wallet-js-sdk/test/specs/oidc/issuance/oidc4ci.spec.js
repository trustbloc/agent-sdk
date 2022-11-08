/*
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import chai, {expect} from 'chai';
import moxios from 'moxios';
import {OpenID4CI} from '@';
import {decode} from "js-base64";

function parseTransaction(txn) {
  return JSON.parse(decode(txn));
}

const defaultIssuerURI = 'https://issuer.example.com';
const walletCallbackURI = 'https://wallet.example.com/cb';
const userDID = 'did:example:foo';
const clientID = '12345';
const issuerRequestURI = 'request-uri-value';
const mockCredJWT = 'foo.bar.baz';


describe('OpenID4CI Auth Code Flow', async function(){
  before(function () {
    moxios.install();
  });

  after(function () {
    moxios.uninstall();
  });

  const clientConfig = {
    walletCallbackURI,
    userDID,
    clientID,
  };

  const client = new OpenID4CI(clientConfig, function(a,b,c,d){return 'foo.jwt.signature'});

  let redirect, clientState;

  it('authorization request', async function() {
    moxios.stubRequest(/\/\.well-known\/openid-configuration/, {
      status: 200,
      response: {
        issuer: defaultIssuerURI,
        authorization_endpoint: defaultIssuerURI + '/auth',
        token_endpoint: defaultIssuerURI + '/token',
        pushed_authorization_request_endpoint: defaultIssuerURI + '/par',
        require_pushed_authorization_requests: true,
        credential_endpoint: defaultIssuerURI + '/credentials',
      }
    });

    moxios.stubRequest(/par/, {
      status: 200,
      response: {request_uri: issuerRequestURI, expires_in: 60},
    });

    const issuanceRequest = {
      issuer: defaultIssuerURI,
      credential_type: 'credential_type',
      user_pin_required: true,
      op_state: 'op_state_value',
    };

    const resp = await client.authorize(issuanceRequest, '');

    redirect = resp.redirect;
    clientState = resp.clientState;

    const redirectURL = new URL(redirect);
    expect(redirectURL.origin).to.equal(defaultIssuerURI);
    expect(redirectURL.pathname).to.equal('/auth');
    expect(redirectURL.searchParams.get('request_uri')).to.equal(issuerRequestURI);
    expect(redirectURL.searchParams.get('client_id')).to.equal(clientID);
  });

  it('issuer returns to wallet callback', async function() {
    const parsedState = parseTransaction(clientState);

    const returnCallback = new URL(walletCallbackURI);
    returnCallback.searchParams.append('code', 'abcdefg');
    returnCallback.searchParams.append('state', parsedState.oauthState);

    moxios.stubRequest( /token/, {
      status: 200,
      response: {
        access_token: 'ACCESS-TOKEN-VALUE',
        token_type: 'Bearer',
        c_nonce: 'ABCDEF123456',
      },
    });

    moxios.stubRequest( /credentials/, {
      status: 200,
      response: {
        format: 'jwt',
        credential: mockCredJWT,
      },
    });

    const {format, credential} = await client.callback(returnCallback.href, clientState);

    expect(format).to.equal('jwt');
    expect(credential).to.equal(mockCredJWT);
  });
});

describe('OpenID4CI Pre-Auth Flow', async function(){
  before(function () {
    moxios.install();
  });

  after(function () {
    moxios.uninstall();
  });

  const clientConfig = {
    walletCallbackURI,
    userDID,
    clientID,
  };

  const client = new OpenID4CI(clientConfig, function(a,b,c,d){return 'foo.jwt.signature'});

  it('pre-auth flow success', async function() {
    moxios.stubRequest(/\/\.well-known\/openid-configuration/, {
      status: 200,
      response: {
        issuer: defaultIssuerURI,
        authorization_endpoint: defaultIssuerURI + '/auth',
        token_endpoint: defaultIssuerURI + '/token',
        pushed_authorization_request_endpoint: defaultIssuerURI + '/par',
        require_pushed_authorization_requests: true,
        credential_endpoint: defaultIssuerURI + '/credentials',
      }
    });

    moxios.stubRequest( /token/, {
      status: 200,
      response: {
        access_token: 'ACCESS-TOKEN-VALUE',
        token_type: 'Bearer',
        c_nonce: 'ABCDEF123456',
      },
    });

    const mockCredJWT = 'foo.bar.baz';

    moxios.stubRequest( /credentials/, {
      status: 200,
      response: {
        format: 'jwt',
        credential: mockCredJWT,
      },
    });

    const issuanceRequest = {
      issuer: defaultIssuerURI,
      credential_type: 'credential_type',
      user_pin_required: true,
      op_state: 'op_state_value',
      'pre-authorized_code': 'foo',
    };

    const {format, credential} = await client.authorize(issuanceRequest, 'user-pin');

    expect(format).to.equal('jwt');
    expect(credential).to.equal(mockCredJWT);
  });
});
