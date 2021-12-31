/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {
  contentTypes,
  findAttachmentByFormat,
  normalizePresentationSubmission,
  UniversalWallet,
  waitForEvent,
} from "..";

import axios from "axios";
import jp from "jsonpath";

const STATE_COMPLETE_MSG_TOPIC = "didexchange-state-complete";
const STATE_COMPLETE_MSG_TYPE =
  "https://trustbloc.dev/didexchange/1.0/state-complete";
const PRESENT_PROOF_STATE_TOPIC = "present-proof_states";
const PRESENT_PROOF_ACTION_TOPIC = "present-proof_actions";
const ISSUE_CREDENTIAL_ACTION_TOPIC = "issue-credential_actions";
const PRESENTATION_SENT_STATE_ID = "presentation-sent";

const DEFAULT_LABEL = "agent-default-label";
const ROUTER_CREATE_INVITATION_PATH = `/didcomm/invitation`;
const ROUTER_CREATE_INVITATION_V2_PATH = `/didcomm/invitation-v2`;
const ATTACH_FORMAT_CREDENTIAL_MANIFEST =
  "dif/credential-manifest/manifest@v1.0";
const ATTACH_FORMAT_CREDENTIAL_FULFILLMENT =
  "dif/credential-manifest/fulfillment@v1.0";
const MSG_TYPE_ISSUE_CREDENTIAL_V2 =
    "https://didcomm.org/issue-credential/2.0/issue-credential";
const MSG_TYPE_ISSUE_CREDENTIAL_V3 =
    "https://didcomm.org/issue-credential/3.0/issue-credential";
const MSG_TYPE_ISSUE_CREDENTIAL_PROBLEM_REPORT_V2 =
  "https://didcomm.org/issue-credential/2.0/problem-report";
const WEB_REDIRECT_STATUS_OK = "OK";

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
                console.debug("received state complete msg!.");
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
   *  @see {@link https://identity.foundation/waci-presentation-exchange/#presentation-2|WACI Presentation flow } for more details.
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
   * @returns {Array<Object>} response.normalized - normalized version of `response.presentations` where response credentials are grouped by input descriptors.
   * Can be used to detect multiple credential result for same query.
   *
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

    let {attachment, threadID} = getPresentationAttachmentAndThreadID(presentationRequest)
    const query = attachment.map(_query);
    const { results } = await this.wallet.query(auth, query);

    const normalized = results.map((result) =>
      normalizePresentationSubmission(query, result)
    );

    return { threadID: threadID, presentations: results, normalized };
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

  /**
   *  Initiates WACI credential issuance interaction from wallet.
   *
   *  accepts an out of band invitation, sends request credential message to inviter, waits for offer credential message response from inviter.
   *
   *  If present, reads presentation definition(s) from offer credential message, performs query in wallet and returns response presentation(s) to be shared.
   *
   *  @see {@link https://identity.foundation/waci-presentation-exchange/#issuance-2|WACI Issuance flow } for more details.
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
   *  @param {Object} proposeOptions - (optional) for sending message proposing credential.
   *  @param {String} proposeOptions.from - (optional) option from DID option to customize sender DID..
   *  @param {Time} proposeOptions.timeout - (optional) to wait for offer credential message from relying party.
   *
   * @returns {Object} response - promise of object containing offer credential message from issuer or error if operation fails.
   * @returns {String} response.threadID - thread ID of credential interaction, to be used for correlation in future.
   * @returns {Object} response.error - error containing status, code and redirect URL if requested by issuer.
   * @returns {Object} manifest - credential manifest sent by issuer.
   * @returns {Object} fulfillment - credential fulfillment sent by issuer.
   * @returns {Array<Object>} response.presentations - array of presentation responses from wallet query.
   * @returns {Array<Object>} response.normalized - normalized version of `response.presentations` where response credentials are grouped by input descriptors.
   * @returns {String} domain - domain parameter sent by issuer for proving ownership of DID or freshness of proof.
   * @returns {String} challenge - challenge parameter sent by issuer for proving ownership of DID or freshness of proof..
   * @returns {String} comment - custom comment sent by issuer along with credential fulfillment.
   * Can be used to detect multiple credential result for same query.
   */
  async initiateCredentialIssuance(
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
    // propose credential and expect offer credential response.
    let { offerCredential } = await this.wallet.proposeCredential(
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
      "offer credential from issuer",
      JSON.stringify(offerCredential, null, 2)
    );

    if (
      offerCredential["@type"] === MSG_TYPE_ISSUE_CREDENTIAL_PROBLEM_REPORT_V2
    ) {
      const { status, url } = offerCredential["~web-redirect"];
      const { code } = offerCredential["description"];

      return { error: { status, url, code } };
    }

    // find manifest
    // TODO : for now, assuming there will only be one manifest per offer credential msg
    const manifest = findAttachmentByFormat(
      offerCredential.formats,
      offerCredential["offers~attach"],
      ATTACH_FORMAT_CREDENTIAL_MANIFEST
    );

    const { presentation_definition, options } = manifest;
    const { domain, challenge } = options ? options : {};

    // perform presentation-exchange or DID Auth.
    let presentations, normalized;
    if (presentation_definition) {
      const query = [
        {
          type: "PresentationExchange",
          credentialQuery: [presentation_definition],
        },
      ];

      // supporting only one presentation definition for now.
      const { results } = await this.wallet.query(auth, query);

      // for grouping duplicate query results
      normalized = results.map((result) =>
        normalizePresentationSubmission(query, result)
      );
      presentations = results;
    } else if (domain || challenge) {
      const { results } = await this.wallet.query(auth, [
        {
          type: "DIDAuth",
        },
      ]);

      presentations = results;
    }

    // find fulfillment
    // TODO : for now, assuming there will only be one fulfillment per offer credential msg
    const fulfillment = findAttachmentByFormat(
      offerCredential.formats,
      offerCredential["offers~attach"],
      ATTACH_FORMAT_CREDENTIAL_FULFILLMENT
    );

    // TODO read descriptors from manifests & fulfillments for credential preview in UI. (Pending support in vcwallet).

    const { comment } = offerCredential;
    const { thid } = offerCredential["~thread"];

    return {
      threadID: thid,
      manifest,
      fulfillment,
      presentations,
      normalized,
      domain,
      challenge,
      comment,
    };
  }

  /**
   *  Completes WACI credential issuance flow.
   *
   *  Sends request credential message to issuer as part of ongoing WACI issuance flow and waits for credential fulfillment response from issuer.
   *  Optionally sends presentations as credential application attachments as part of request credential message.
   *
   *  Response credentials from credential fulfillment will be saved to collection of choice.
   *
   *  @see {@link https://identity.foundation/waci-presentation-exchange/#issuance-2|WACI Issuance flow } for more details.
   *
   *  @param {String} auth -  authorization token for performing this wallet operation.
   *  @param {String} threadID - threadID of credential interaction.
   *  @param {Object} presentation - to be sent as part of credential fulfillment. This presentations will be converted into credential fulfillment format
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
   *  @param {Object} options - (optional) for sending message requesting credential.
   *  @param {Bool} options.waitForDone - (optional) If true then wallet will wait for credential fulfillment message or problem report to arrive.
   *  @param {Time} options.WaitForDoneTimeout - (optional) timeout to wait for for credential fulfillment message or problem report to arrive. Will be considered only
   *  when `options.waitForDone` is true.
   *  @param {Bool} options.autoAccept - (optional) if enabled then incoming issue credential or problem report will be auto accepted. If not provided then
   *  wallet will rely on underlying agent to accept incoming actions.
   *  @param {String} options.collection - (optional) ID of the wallet collection to which the credential should belong.
   *
   * @returns {Promise<Object>} - promise of object containing request credential status & redirect info or error if operation fails.
   */
  async completeCredentialIssuance(
    auth,
    threadID,
    presentation,
    {
      controller,
      verificationMethod,
      created,
      domain,
      challenge,
      proofType,
      proofRepresentation,
    } = {},
    { waitForDone, WaitForDoneTimeout, autoAccept, collectionID } = {}
  ) {
    // TODO convert presentation to credential fulfillment type and then sign.
    let signedPresentation;
    if (presentation) {
      let signed = await this.wallet.prove(
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

      signedPresentation = signed.presentation;
    }

    let credentials;
    if (autoAccept) {
      waitForEvent(this.agent, {
        topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
        timeout: WaitForDoneTimeout,
        callback: async (payload) => {
          const { Message, Properties } = payload;
          const { piid } = Properties;
          const type = Message["@type"];

          if (type === MSG_TYPE_ISSUE_CREDENTIAL_V2 || type === MSG_TYPE_ISSUE_CREDENTIAL_V3) {
            await this.agent.issuecredential.acceptCredential({
              piid,
              skipStore: true,
            });

            credentials = jp.query(
              Message["credentials~attach"],
              `$[*].data.json.verifiableCredential[*]`
            );
          } else if (type === MSG_TYPE_ISSUE_CREDENTIAL_PROBLEM_REPORT_V2) {
            await this.agent.issuecredential.acceptProblemReport({
              piid,
            });
          }
        },
      });
    }

    const response = await this.wallet.requestCredential(
      auth,
      threadID,
      signedPresentation,
      { waitForDone, WaitForDoneTimeout }
    );

    if (response.status != WEB_REDIRECT_STATUS_OK) {
      return response;
    }

    // expecting only one credential for now,  TODO it has to be in credential fulfillment
    if (credentials.length == 0) {
      throw "no incoming credentials found";
    }

    const contentType = contentTypes.CREDENTIAL;
    await Promise.all(
      credentials.map(
        async (content) =>
          await this.wallet.add({
            auth,
            contentType,
            content,
            collectionID,
          })
      )
    );

    return response;
  }
}

/**
 * Get attachment and threadID from presentationRequest instance based on DIDComm V1 or V2 formats.
 *
 * @param presentationRequest instance
 */
function getPresentationAttachmentAndThreadID(presentationRequest) {
    let attachment, threadID
    if (presentationRequest["request_presentations~attach"]) { // didcomm v1
        attachment = presentationRequest["request_presentations~attach"]
        console.log("GetPresentationAttachmentAndThreadID - didcomm v1 attachment key set: "+attachment)
    } else if (presentationRequest["attachments"]) { // didcomm v2
        attachment = presentationRequest["attachments"]
        console.log("GetPresentationAttachmentAndThreadID - didcomm v2 attachment key set: "+attachment)
    } else {
        console.error("GetPresentationAttachmentAndThreadID - unrecognized presentationRequest object: '"+ JSON.stringify(presentationRequest, undefined, 2) + "'\n   attachments key not found")
        return
    }

    if (presentationRequest["~thread"]) { // didcomm v1
        const { thid } = presentationRequest["~thread"];
        threadID = thid
    } else if (presentationRequest["thid"]) { // didcomm v2
        threadID = presentationRequest["thid"]
    } else {
        console.error("GetPresentationAttachmentAndThreadID - unrecognized presentationRequest object: '"+ JSON.stringify(presentationRequest, undefined, 2) + "'\n   thread ID key not found")
        return
    }

    return {
        attachment: attachment,
        threadID: threadID
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
 * @param isDIDCommV2 flag using DIDComm V2
 */
export const createInvitationFromRouter = async (endpoint, { isDIDCommV2 = false } = {}) => {
  let routerInvitationURL = `${endpoint}`
  console.log("createInvitationFromRouter isDIDCommV2 ? " + isDIDCommV2);

  if (isDIDCommV2 == true) {
      routerInvitationURL += `${ROUTER_CREATE_INVITATION_V2_PATH}`
  } else {
      routerInvitationURL += `${ROUTER_CREATE_INVITATION_PATH}`
  }

  console.log(
    "getting invitation from ",
    `${routerInvitationURL}`
  );
  let response = await axios.get(`${routerInvitationURL}`);
  return response.data.invitation;
};

/**
 * Connect given agent to edge mediator/router.
 *
 * @param agent trustbloc agent
 * @param endpoint edge router endpoint
 * @param waitForStateComplete wait for did exchange state complete message
 * @param isDIDCommV2 flag using DIDComm V2
 */
export async function connectToMediator(
  agent,
  mediatorEndpoint,
  { waitForStateComplete = true } = {},
  { isDIDCommV2 = false } = {},
) {
  let inv = await createInvitationFromRouter(mediatorEndpoint, {isDIDCommV2: isDIDCommV2})
  let resp = await agent.mediatorclient.connect({
    myLabel: "agent-default-label",
    invitation: inv,
    stateCompleteMessageType: waitForStateComplete
      ? STATE_COMPLETE_MSG_TYPE
      : "",
  });

  if (resp.connectionID) {
    console.log("router registered successfully!", resp.connectionID);
  } else {
    console.log("router was not registered!");
  }

  console.debug("in connectToMediator() - invitation: "+ JSON.stringify(inv))

  let invID = inv["@id"]
  if (!invID) {
      invID = inv["from"]
  }

  return resp.connectionID, invID
}
