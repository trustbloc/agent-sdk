/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import { encode, decode } from "js-base64";

/**
 * OpenID4CI module provides APIs for wallets to receive verifiable credentials through OIDC for Credential Issuance.
 *
 * @module OpenID4CI
 *
 */

/**
 * Client providing support for OIDC4CI methods.
 *
 * @alias module:OpenID4CI
 */
export class OpenID4CI {
  /**
   * Creates a new OpenID4CI client.
   *
   * @param{Object} clientConfig -
   * @param{string} clientConfig.walletCallbackURI - wallet-hosted URI for callback redirect from issuer, after user authorizes issuer.
   * @param{string} clientConfig.userDID - the DID of the wallet user. The requested credential will be bound to this DID.
   * @param{string} clientConfig.clientID - the wallet app instance's OIDC client ID.
   * @param{function} jwtSigner - a handler for signing a JWT for proving possession of userDID.
   *    When this creates a JWT, it should add kid, jwk, or x5c field corresponding to a key bound to userDID.
   *
   */
  constructor(
    clientConfig,
    jwtSigner = async function (iss, aud, iat, c_nonce) {
      throw new Error("OpenID4CI client needs to be provided with a JWT Signer function.")
    },
  ) {
    this.clientConfig = clientConfig;
    this.jwtSigner = jwtSigner;
  }

  /**
   * authorize is used by a wallet to authorize an issuer's OIDC verifiable-credential Issuance Request.
   *
   * @param{Object} req - the Issuance Request from an OIDC Issuer that the wallet intends to authorize.
   * @param{string} userPin - Optional. A 2FA PIN provided by the Issuer to the wallet through a separate channel
   * from the request, for pre-authorized issuance flows.
   *
   * @returns{Object} response - An object containing either `credential` and `format` fields, or `redirect` and `clientState` fields.
   *
   * response.credential - The credential issued, returned if the Issuer has pre-authorized the wallet.
   *
   * response.format - The format of the credential issued.
   *
   * response.redirect - The URI for redirecting to the Issuer for authorization. The Wallet should redirect without changing this URI.
   *
   * response.clientState - A string encoding the client state that this API will need to be passed in to callback().
   * The wallet should save this client-side.
   */
  async authorize(
    req,
    userPin = '',
  ) {
    const {
      issuer,
      credential_type: credentialType,
      user_pin_required: userPINRequired,
      op_state: opState,
    } = req;
    const preAuthCode = req['pre-authorized_code'];

    if (preAuthCode && preAuthCode.length) {
      if (userPINRequired && (!userPin || !userPin.length)) {
        throw new Error("Issuance Request indicates a user PIN is required, but no user PIN was provided.");
      }

      return await preAuthorized(issuer, credentialType, preAuthCode, userPin, this.clientConfig, this.jwtSigner);
    }

    return await pushAuthRequest(issuer, credentialType, opState, this.clientConfig);
  }


  /**
   * callback is the OIDC issuance callback, used when the issuer has returned to the wallet, and the wallet user
   * has consented to have the credential created.
   *
   * @param {string} callbackURI - The complete wallet callback URI that the Issuer redirected to, after the Issuer
   * authenticated the user. This URI should include all query parameters, fragments, etc, that the Issuer added to the callback URI.
   * @param {string} clientState - The opaque client state string that was returned from the `authorize` call that started this issuance flow.
   *
   * @returns{Object} response - An object containing `credential` and `format` fields.
   *
   * response.credential - The credential issued.
   *
   * response.format - The format of the credential issued.
   *
   */
  async callback(
    callbackURI,
    clientState,
  ) {
    const callbackQuery = (new URL(callbackURI)).searchParams;

    const authorizationCode = callbackQuery.get('code');
    const oauthState = callbackQuery.get('state');

    if (!authorizationCode || !authorizationCode.length) {
      throw new Error('Callback URI is missing authorization `code` as a query parameter.');
    }

    if (!oauthState || !oauthState.length) {
      throw new Error('Issuer did not include state parameter in callback URI when returning user to wallet.');
    }

    if (!clientState || !clientState.length) {
      throw new Error('Cannot complete authorization flow, `callback` must be provided with clientState string returned by ' +
        'previous `authorize` call.');
    }

    const transactionData = parseTransaction(clientState);

    if (transactionData.oauthState !== oauthState) {
      throw new Error('Issuer provided a state parameter that does not match the state parameter of the client-side transaction. ' +
        'This may be an error in the Issuer, or a mix-up of state data in a client performing multiple flows at the same time.');
    }

    const tokenRequest = new URLSearchParams();
    tokenRequest.append('grant_type', 'authorization_code');
    tokenRequest.append('code', authorizationCode);
    tokenRequest.append('redirect_uri', this.clientConfig.walletCallbackURI);
    tokenRequest.append('client_id', this.clientConfig.clientID);

    const tokenResponse = await axios.post(
      transactionData.issuerMetadata.token_endpoint,
      tokenRequest,
    ).then((resp) => resp.data).catch((e) => {
      throw new Error('Error authorizing wallet with Issuer: failed to exchange auth code for token:', e);
    });

    return await getCredential(tokenResponse, transactionData, this.clientConfig, this.jwtSigner);
  }
}

/**
 * pushAuthRequest performs a pushed authorization request to the issuer, and returns a redirect URI and client state.
 *
 * @param {string} issuerURI - uri of issuer server.
 * @param {string} credentialType - type of credential to request, as found in the Issuance Request.
 * @param {string} opState - (optional) op_state parameter for issuer-initiated transactions.
 * @param {object} clientConfig - client configuration.
 * @param {string} clientConfig.walletCallbackURI - wallet-hosted URI for callback redirect from issuer, after user authorizes issuer.
 * @param {string} clientConfig.clientID - the wallet app instance's OIDC client ID.
 *
 *
 * @returns {Object} - If successful, returns the redirect URI for redirecting to the Issuer for user consent,
 * and a client state string to be saved and provided to the follow-up wallet callback.
 */
async function pushAuthRequest(
  issuerURI ,
  credentialType ,
  opState = '',
  clientConfig ,
) {
  const issuerMetadata = await getIssuerMetadata(issuerURI);
  const oauthState = generateNonce();

  const authRequest = new URLSearchParams();
  authRequest.append('response_type', 'code');
  authRequest.append('client_id', clientConfig.clientID);
  authRequest.append('redirect_uri', clientConfig.walletCallbackURI);
  authRequest.append('state', oauthState);
  authRequest.append('op_state', opState);
  authRequest.append('authorization_details', JSON.stringify([
    {
      type: 'openid_credential',
      credential_type: credentialType,
    }
  ]));

  const {request_uri: requestURI} = await axios.post(
    issuerMetadata.pushed_authorization_request_endpoint,
    authRequest,
  ).then((resp) => resp.data).catch((e) => {
    throw new Error('Error authorizing wallet with Issuer: failed to send Pushed Auth Request:', e);
  });

  const redirectToIssuer = new URL(issuerMetadata.authorization_endpoint);
  redirectToIssuer.searchParams.append('request_uri', requestURI);
  redirectToIssuer.searchParams.append('client_id', clientConfig.clientID);

  const transactionData = {
    credentialType: credentialType,
    clientID: clientConfig.clientID,
    issuerMetadata: issuerMetadata,
    oauthState: oauthState,
  };

  return {redirect: redirectToIssuer.href, clientState: marshalTransaction(transactionData)};
}

async function preAuthorized(
  issuerURI,
  credentialType,
  preAuthorizedCode,
  userPIN,
  clientConfig,
  jwtSigner,
) {
  const issuerMetadata = await getIssuerMetadata(issuerURI);
  const transactionData = {
    issuerMetadata,
    credentialType,
    clientID: clientConfig.clientID,
  };

  const tokenRequest = new URLSearchParams();
  tokenRequest.append('grant_type', 'urn:ietf:params:oauth:grant-type:pre-authorized_code');
  tokenRequest.append('pre-authorized_code', preAuthorizedCode);
  tokenRequest.append('user_pin', userPIN);

  const tokenResponse = await axios.post(
    transactionData.issuerMetadata.token_endpoint,
    tokenRequest,
  ).then((resp)=> resp.data).catch((e) => {
    throw new Error('Error authorizing wallet with Issuer: failed to exchange pre-authorized code for token:', e);
  });
  // TODO: poll if response is deferred (if tokenResponse.authorization_pending is true),
  //  waiting tokenResponse.interval seconds before next request (or 5 seconds if interval missing or less than 5).

  return await getCredential(tokenResponse, transactionData, clientConfig, jwtSigner);
}

async function getCredential(
  tokenResponse,
  transactionData,
  clientConfig = {
    userDID: '',
  },
  jwtSigner,
) {
  const jwt = await jwtSigner(transactionData.clientID, transactionData.issuerMetadata.issuer, new Date().getTime() / 1000, tokenResponse.c_nonce);

  const credentialRequest = {
    type: transactionData.credentialType,
    did: clientConfig.userDID,
    proof: {
      proof_type: 'jwt',
      jwt,
    },
  };

  const credentialResponse = await axios.post(transactionData.issuerMetadata.credential_endpoint, credentialRequest, {
    headers: {
      Authorization: "Bearer " + tokenResponse.access_token,
    },
  }).then((resp) => resp.data).catch((e) => {
    throw new Error('Error fetching credential from Issuer:', e);
  });
  // TODO deferred flow implementation deferred

  return {format: credentialResponse.format, credential: credentialResponse.credential};
}

async function getIssuerMetadata(issuer_uri) {
  return await axios.get(issuer_uri + '/.well-known/openid-configuration')
    .then((resp) => {
      return resp.data;
    })
    .catch((e) => {
      throw new Error('Failed to fetch issuer server metadata:', e);
    });
}

function parseTransaction(txn) {
  return JSON.parse(decode(txn));
}

function marshalTransaction(txn) {
  return encode(JSON.stringify(txn), true);
}

function generateNonce() {
  return crypto.randomUUID();
}