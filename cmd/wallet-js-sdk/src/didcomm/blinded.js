/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

/**
 * blinded-routing module provides features supporting blinded DIDComm routing features.
 *
 * @module blinded-routing
 */
/**
 * BlindedRouter provides DIDComm message based blinded routing features.
 *
 * @alias module:blinded-routing
 */
export class BlindedRouter {
  /**
   * @param {string} agent - aries agent.
   */
  constructor(agent) {
    this.agent = agent;
  }

  /**
   *
   * This function provides functionality of sharing peer DID with connecting party for blinded DIDComm.
   *
   * @param {Object} connection - connection record of the connection established.
   */
  async sharePeerDID(connection) {
    let { ConnectionID } = connection;

    // request peer DID from other party
    console.debug("Sending DID Doc request");
    let response = await this.agent.blindedrouting.sendDIDDocRequest({
      connectionID: ConnectionID,
    });
    console.log("payload from did doc response", response);

    if (!response || !response.payload || !response.payload.message) {
      throw "no response DID found in did doc response";
    }

    let { message } = response.payload;

    let peerDID = message.data.didDoc;
    if (!peerDID) {
      console.error(
        "failed to get peerDID from inviter, could not find peer DID in response message."
      );
      throw "failed to get peer DID from inviter";
    }
    console.debug("received peer DID");

    // request wallet peer DID from router by sending peer DID from other party
    console.debug("requesting peer DID from wallet");
    let walletDID = await requestDIDFromMediator(this.agent, peerDID);
    console.debug(`wallet's peer DID: ${JSON.stringify(walletDID, null, 2)}`);

    console.log("sharing wallet peer DID to inviter");
    // share wallet peer DID to other party
    await this.agent.blindedrouting.sendRegisterRouteRequest({
      messageID: message["@id"],
      didDoc: walletDID,
    });
  }
}

async function requestDIDFromMediator(agent, reqDoc) {
  let res = await agent.mediatorclient.sendCreateConnectionRequest({
    didDoc: reqDoc,
  });

  if (res.payload && res.payload.message) {
    let response = res.payload.message;
    // TODO currently getting routerDIDDoc as byte[], to be fixed
    if (response.data.didDoc && response.data.didDoc.length > 0) {
      return JSON.parse(
        String.fromCharCode.apply(String, response.data.didDoc)
      );
    }
  }

  console.error(
    "failed to request DID from router, failed to get connection response"
  );
  throw "failed to request DID from router, failed to get connection response";
}
