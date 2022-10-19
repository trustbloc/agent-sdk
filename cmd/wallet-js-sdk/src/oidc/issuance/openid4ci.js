/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/**
 * PrepareIssuanceRequest fetches the issuer metadata and prepares an issuance transaction,
 * returning to the caller for an opportunity to consent.
 *
 * @param {string} issuer_uri - uri of issuer server: TODO decide if this is specifically the metadata endpoint.
 * @param {Object} options
 * @param {string} options.op_state - (optional) op_state parameter for issuer-initiated transactions.
 *
 * @returns {Promise<Object>} - If successful, returns the issuer metadata and a continuation handler function.
 *   Issuer metadata may be used to display a UI showing the fields the issuer will include in each potential credential,
 *   for user selection/consent. Call the continuation function if wallet user consents to request a credential.
 */
export async function PrepareIssuanceRequest(
  issuer_uri = '',
  {
    op_state = '',
  }
) {
  // stub
  const fetchIssuerMetadata = async function (uri) {
    // data returned by issuer server
    return {
      issuer: uri,
      authorization_endpoint: uri + '/auth',
      token_endpoint: uri + '/token',
      pushed_authorization_request_endpoint: uri + '/par',
      require_pushed_authorization_requests: true,
    };
  };

  // TODO: Will issuer metadata be at a .well-known location, or should we expect the issuer_uri parameter to be the
  //  complete uri for the metadata endpoint
  const issuerMetadata = await fetchIssuerMetadata(issuer_uri);

  const state = '';

  const requestIssuanceHandler = async function (
    wallet_callback_uri = '',
    {
      credential_type = '', // TODO: should we support requesting multiple credentials at once?
      credential_scope = '',
    },
    save_transaction = async function(txn_id='', data={}){},
    ) {
    return await requestIssuance(
      wallet_callback_uri,
      state,
      {
        credential_type,
        credential_scope,
        op_state,
      },
      issuerMetadata,
      save_transaction,
    );
  }

  return {issuer: issuerMetadata, accept: requestIssuanceHandler};
}

async function requestIssuance(
    wallet_callback_uri = '',
    oauth_state = '',
  {
    credential_type = '', // TODO: should we support requesting multiple credentials at once?
    credential_scope = '',
    op_state = '',
  },
  issuer_metadata = {},
  save_transaction = async function(txn_id='', data={}){},
) {
  // stub
  const pushAuthRequest = async function (url, req) {
    // data returned by issuer server
    return {request_uri: '', expires_in: 60};
  }

  // TODO create new client for each transaction, or use a caller-provided client ID?
  const client_id = '';

  const authRequest = {
    scope: credential_scope,
    response_type: 'code',
    client_id,
    redirect_uri: wallet_callback_uri,
    state: oauth_state,
    op_state: op_state,
    authorization_details: [
      {
        type: 'openid_credential',
        credential_type: credential_type,
      }
    ]
  };

  const pushedAuthResponse = await pushAuthRequest(issuer_metadata.pushed_authorization_request_endpoint, authRequest);

  // TODO query-encode
  const redirect_to_issuer = issuer_metadata.authorization_endpoint + '?request_uri=' + pushedAuthResponse.request_uri +
    '&client_id=' + client_id;

  const transaction_data = {
    credential_type,
    client_id,
    issuer_metadata,
  };

  await save_transaction(oauth_state, transaction_data);

  return {redirect: redirect_to_issuer, state: oauth_state};
}

/**
 * CompleteIssuance is the OIDC issuance callback, used when the issuer has returned to the wallet, and the wallet user
 * has consented to have the credential created.
 *
 * @param {string} oidc_state - the state parameter of this OIDC transaction.
 * @param {string} user_did - the DID of the wallet that is receiving this credential.
 * @param {Object} options
 * @param {function} jwt_signer - a handler for signing a JWT for proving possession of user_did.
 *    When this creates a JWT, it should add kid, jwk, or x5c field corresponding to a key bound to user_did.
 * @param {function} load_transaction - a handler for loading an OIDC issuance transaction under the given transaction ID.
 *    Used for loading transaction state saved by requestIssuance.
 * @param {function} delete_transaction - a handler for deleting an OIDC issuance transaction under the given transaction ID.
 */
export async function CompleteIssuance(
  oidc_state = '',
  user_did = '',
  {
    authorization_code = '',
    pre_authorized_code = '',
    user_pin = '',
  },
  jwt_signer = async function(iss, aud, iat, c_nonce){return {}},
  // TODO: should these handlers be parameterless instead, as caller-provided closures on the txn ID?
  //  then the caller doesn't need to trust that this API will only touch the relevant transaction.
  load_transaction = async function(txn_id= ''){return {}},
  delete_transaction = async function(txn_id = ''){},
) {

  const transaction_data = await load_transaction(oidc_state);

  // stub
  const exchangeAuthCode = async function(auth_code, token_endpoint) {
    // data returned by issuer server
    return {
      access_token: '',
      token_type: '',
      c_nonce: '',
    };
  };

  // stub
  const exchangePreAuthCode = async function(pre_auth_code, user_pin, token_endpoint) {
    // data returned by issuer server
    return {
      access_token: '',
      token_type: '',
      c_nonce: '',
      authorization_pending: true,
      interval: 1,
    };
  };

  let tokenResponse;

  if (authorization_code !== '') {
    tokenResponse = await exchangeAuthCode(authorization_code, transaction_data.issuer_metadata.token_endpoint);
  } else if (pre_authorized_code !== '') {
    tokenResponse = await exchangePreAuthCode(pre_authorized_code, user_pin, transaction_data.issuer_metadata.token_endpoint);
    // TODO: poll for token if authorization is pending
  }

  // stub
  const sendCredentialRequest = async function(req) {
    // data returned by issuer server
    return {
      format: '',
      credential: {},
    };
  }

  const jwt = await jwt_signer(transaction_data.client_id, transaction_data.issuer_metadata.issuer, /*time now*/0, tokenResponse.c_nonce)

  // TODO note: credential request seems to handle only one credential type at a time, while the original auth request can request multiple credentials.
  //  should we:
  //   A) be able to request multiple credentials in one call, or
  //   B) make multiple requests, one for each requested credential, or
  //   C) only request one credential at a time in a single issuance flow
  const credentialRequest = {
    type: transaction_data.credential_type,
    format: '',
    did: user_did,
    proof: {
      proof_type: 'jwt',
      jwt,
    },
  };

  const credentialResponse = await sendCredentialRequest(credentialRequest)
  // TODO poll for credential if response is deferred

  await delete_transaction(oidc_state);

  return {format: credentialResponse.format, credential: credentialResponse.credential};
}
