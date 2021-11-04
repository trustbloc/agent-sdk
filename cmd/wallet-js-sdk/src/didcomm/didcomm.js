/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import { UniversalWallet, waitForEvent, POST_STATE } from "..";
import axios from "axios";

const STATE_COMPLETE_MSG_TOPIC = "didexchange-state-complete";
const STATE_COMPLETE_MSG_TYPE =
  "https://trustbloc.dev/didexchange/1.0/state-complete";
const PRESENT_PROOF_STATE_TOPIC = "present-proof_states";
const PRESENT_PROOF_ACTION_TOPIC = "present-proof_actions";
const PRESENTATION_SENT_STATE_ID = "presentation-sent";

const DEFAULT_LABEL = "agent-default-label";
const ROUTER_CREATE_INVITATION_PATH = `/didcomm/invitation`;

/**
 * didcomm module provides wallet based DIDComm features.
 *
 * @module didcomm
 */
/**
 * DIDComm module provides wallet based DIDComm features. Currently supporting DID-Exchange, Present Proof & WACI features.
 *
 * @alias module:didcomm
 */
export class DIDComm {
  /**
   *
   *
   * @param {string} agent - aries agent.
   * @param {string} user -  unique wallet user identifier, the one used to create wallet profile.
   *
   */
  constructor({ agent, user } = {}) {
    this.agent = agent;
    this.wallet = new UniversalWallet({ agent: this.agent, user });
  }

  /**
   *  accepts an out of band invitation, performs did-exchange and returns connection record of connection established.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} invitation - out of band invitation.
   *
   *  @param {Object} options - (optional) for accepting incoming out-of-band invitation and connecting to inviter.
   *  @param {String} options.myLabel - (optional) for providing label to be shared with the other agent during the subsequent did-exchange.
   *  @param {Array<string>} options.routerConnections - (optional) to provide router connection to be used.
   *  @param {Bool} options.userAnyRouterConnection=false - (optional) if true and options.routerConnections not provided then wallet will find
   *  an existing router connection and will use it for accepting invitation.
   *  @param {String} options.reuseConnection - (optional) to provide DID to be used when reusing a connection.
   *  @param {Bool} options.reuseAnyConnection=false - (optional) to use any recognized DID in the services array for a reusable connection.
   *  @param {Time} options.timeout - (optional) to wait for connection status to be 'completed'.
   *  @param {Bool} options.waitForCompletion - (optional) if true then wait for custom 'didexchange-state-complete' message to conclude connection as completed.
   *
   * @returns {Promise<Object>} - promise of object containing connection ID or error if operation fails.
   */
  async connect(
    auth,
    invitation = {},
    {
      myLabel,
      routerConnections = [],
      userAnyRouterConnection = false,
      reuseConnection,
      reuseAnyConnection = false,
      timeout,
      waitForCompletion,
    } = {}
  ) {
    let { connectionID } = await this.wallet.connect(auth, invitation, {
      myLabel,
      routerConnections:
        routerConnections.length == 0 && userAnyRouterConnection
          ? [
              await getMediatorConnections(this.agent, {
                single: true,
              }),
            ]
          : routerConnections,
      reuseConnection,
      reuseAnyConnection,
      timeout,
    });

    const record = await this.agent.didexchange.queryConnectionByID({
      id: connectionID,
    });

    if (waitForCompletion) {
      this.agent.messaging.registerService({
        name: STATE_COMPLETE_MSG_TOPIC,
        type: STATE_COMPLETE_MSG_TYPE,
      });

      try {
        await new Promise((resolve, reject) => {
          setTimeout(
            () =>
              reject(new Error("time out waiting for state complete message")),
            15000
          );
          const stop = this.agent.startNotifier(
            (msg) => {
              if (
                record.result.MyDID == msg.payload.mydid &&
                record.result.TheirDID == msg.payload.theirdid
              ) {
                stop();
                console.debug("received state complete msg received.");
                resolve(msg.payload.message);
              }
            },
            [STATE_COMPLETE_MSG_TOPIC]
          );
        });
      } catch (e) {
        console.warn("error while waiting for state complete msg !!", e);
      }
    }

    return record;
  }

  /**
   *  Initiates WACI credential share interaction from wallet.
   *
   *  accepts an out of band invitation, sends propose presentation message to inviter, waits for request presentation message reply from inviter.
   *  reads presentation definition(s) from request presentation, performs query in wallet and returns response presentation(s) to be shared.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {Object} invitation - out of band invitation.
   *
   *  @param {Object} connectOptions - (optional) for accepting incoming out-of-band invitation and connecting to inviter.
   *  @param {String} connectOptions.myLabel - (optional) for providing label to be shared with the other agent during the subsequent did-exchange.
   *  @param {Array<string>} connectOptions.routerConnections - (optional) to provide router connection to be used.
   *  @param {Bool} options.userAnyRouterConnection=false - (optional) if true and options.routerConnections not provided then wallet will find
   *  an existing router connection and will use it for accepting invitation.
   *  @param {String} connectOptions.reuseConnection - (optional) to provide DID to be used when reusing a connection.
   *  @param {Bool} connectOptions.reuseAnyConnection=false - (optional) to use any recognized DID in the services array for a reusable connection.
   *  @param {timeout} connectOptions.connectionTimeout - (optional) to wait for connection status to be 'completed'.
   *
   *  @param {Object} proposeOptions - (optional) for sending message proposing presentation.
   *  @param {String} proposeOptions.from - (optional) option from DID option to customize sender DID..
   *  @param {Time} proposeOptions.timeout - (optional) to wait for request presentation message from relying party.
   *
   * @returns {Object} response - promise of object containing presentation request message from relying party or error if operation fails.
   * @returns {String} response.threadID - thread ID of credential interaction to be used for correlation.
   * @returns {Array<Object>} response.presentations - array of presentation responses from wallet query.
   */
  async initiateCredentialShare(
    auth,
    invitation = {},
    {
      myLabel,
      routerConnections = [],
      userAnyRouterConnection = false,
      reuseConnection,
      reuseAnyConnection = false,
      connectionTimeout,
    },
    { from, timeout } = {}
  ) {
    let { presentationRequest } = await this.wallet.proposePresentation(
      auth,
      invitation,
      {
        myLabel,
        routerConnections:
          routerConnections.length == 0 && userAnyRouterConnection
            ? [
                await getMediatorConnections(this.agent, {
                  single: true,
                }),
              ]
            : routerConnections,
        reuseConnection,
        reuseAnyConnection,
        connectionTimeout,
      },
      {
        from,
        timeout,
      }
    );

    console.debug(
      "presentation request from verifier",
      JSON.stringify(presentationRequest, null, 2)
    );

    //supports multiple, but only presentation_definition inside json data.
    let _query = (attachment) => {
      return {
        type: "PresentationExchange",
        credentialQuery: [attachment.data.json["presentation_definition"]],
      };
    };

    let query = presentationRequest["request_presentations~attach"].map(_query);

    let { results } = await this.wallet.query(auth, query);
    let { thid } = presentationRequest["~thread"];

    return { threadID: thid, presentations: results };
  }

  /**
   *  Completes WACI credential share flow.
   *
   *  Signs presentation(s) and sends them as part of present proof message to relying party.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {String} threadID - threadID of credential interaction.
   *  @param {Array<Object>} presentations - to be sent as part of present proof message..
   *
   *  @param {Object} proofOptions - proof options for signing presentation.
   *  @param {String} proofOptions.controller -  DID to be used for signing.
   *  @param {String} proofOptions.verificationMethod - (optional) VerificationMethod is the URI of the verificationMethod used for the proof.
   *  By default, Controller public key matching 'assertion' for issue or 'authentication' for prove functions.
   *  @param {String} proofOptions.created - (optional) Created date of the proof.
   *  By default, current system time will be used.
   *  @param {String} proofOptions.domain - (optional) operational domain of a digital proof.
   *  By default, domain will not be part of proof.
   *  @param {String} proofOptions.challenge - (optional) random or pseudo-random value option authentication.
   *  By default, challenge will not be part of proof.
   *  @param {String} proofOptions.proofType - (optional) signature type used for signing.
   *  By default, proof will be generated in Ed25519Signature2018 format.
   *  @param {String} proofOptions.proofRepresentation - (optional) type of proof data expected ( "proofValue" or "jws").
   *  By default, 'proofValue' will be used.
   *
   *  @param {Object} options - (optional) for sending message proposing presentation.
   *  @param {Bool} options.waitForDone - (optional) If true then wallet will wait for present proof protocol status to be done or abandoned .
   *  @param {Time} options.WaitForDoneTimeout - (optional) timeout to wait for present proof operation to be done.
   *  @param {Bool} options.autoAccept - (optional) can be used to auto accept any incoming problem reports while waiting for present proof protocol status to be done or abandoned.
   *
   * @returns {Promise<Object>} - promise of object containing present prof status & redirect info or error if operation fails.
   */
  async completeCredentialShare(
    auth,
    threadID,
    presentations,
    {
      controller,
      verificationMethod,
      created,
      domain,
      challenge,
      proofType,
      proofRepresentation,
    } = {},
    { waitForDone, WaitForDoneTimeout, autoAccept } = {}
  ) {
    let _prove = async (presentation) => {
      return await this.wallet.prove(
        auth,
        { presentation },
        {
          controller,
          verificationMethod,
          created,
          domain,
          challenge,
          proofType,
          proofRepresentation,
        }
      );
    };

    let vps = await Promise.all(presentations.map(_prove));

    if (autoAccept) {
      waitForEvent(this.agent, {
        topic: PRESENT_PROOF_ACTION_TOPIC,
        timeout: WaitForDoneTimeout,
        callback: async (payload) => {
          const { piid } = payload.Properties;

          await this.agent.presentproof.acceptProblemReport({
            piid,
          });
        },
      });
    }

    if (vps.length == 1) {
      let { presentation } = vps[0];
      return await this.wallet.presentProof(auth, threadID, presentation, {
        waitForDone,
        WaitForDoneTimeout,
      });
    } else {
      // typically only one presentation, if there are multiple then send them all as part of single attachment for now.
      return await this.wallet.presentProof(
        auth,
        threadID,
        { presentations: vps },
        { waitForDone, WaitForDoneTimeout }
      );
    }
  }
}

/**
 * Get router/mediator connections from agent.
 *
 * @param agent instance
 */
export async function getMediatorConnections(agent, { single } = {}) {
  let resp = await agent.mediator.getConnections();
  if (!resp.connections || resp.connections.length === 0) {
    return "";
  }

  if (single) {
    return resp.connections[
      Math.floor(Math.random() * resp.connections.length)
    ];
  }

  return resp.connections.join(",");
}

/**
 * Get DID Invitation from edge router.
 *
 * @param endpoint edge router endpoint
 */
export const createInvitationFromRouter = async (endpoint) => {
  console.log(
    "getting invitation from ",
    `${endpoint}${ROUTER_CREATE_INVITATION_PATH}`
  );
  let response = await axios.get(`${endpoint}${ROUTER_CREATE_INVITATION_PATH}`);
  return response.data.invitation;
};

/**
 * Connect given agent to edge mediator/router.
 *
 * @param agent trustbloc agent
 * @param endpoint edge router endpoint
 * @param wait for did exchange state complete message
 */
export async function connectToMediator(
  agent,
  mediatorEndpoint,
  { waitForStateComplete = true } = {}
) {
  let resp = await agent.mediatorclient.connect({
    myLabel: "agent-default-label",
    invitation: await createInvitationFromRouter(mediatorEndpoint),
    stateCompleteMessageType: waitForStateComplete
      ? STATE_COMPLETE_MSG_TYPE
      : "",
  });

  if (resp.connectionID) {
    console.log("router registered successfully!", resp.connectionID);
  } else {
    console.log("router was not registered!");
  }
}
