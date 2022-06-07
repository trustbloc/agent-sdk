/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import {
  connectToMediator,
  getMediatorConnections,
  waitForEvent,
  findAttachmentByFormat,
} from "../../../src";

import {
  DIDEXCHANGE_STATE_REQUESTED,
  DIDEXCHANGE_STATE_TOPIC,
  loadFrameworks,
  POST_STATE,
  PRESENT_PROOF_ACTION_TOPIC,
  ISSUE_CREDENTIAL_ACTION_TOPIC,
  MSG_TYPE_OFFER_CREDENTIAL_V2,
  MSG_TYPE_OFFER_CREDENTIAL_V3,
  MSG_TYPE_PROPOSE_CREDENTIAL_V2,
  MSG_TYPE_PROPOSE_CREDENTIAL_V3,
  ATTACH_FORMAT_CREDENTIAL_MANIFEST,
  ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
  retryWithDelay,
  testConfig,
} from "../common";

import { expect } from "chai";
import { v4 as uuidv4 } from "uuid";

/**
 * Adapter mocks common issuer or rp adapter features
 *
 * @param agent instance
 * @class
 */
export class Adapter {
  constructor(label) {
    this.label = label;
  }

  async init({
    mediaTypeProfiles = ["didcomm/aip2;env=rfc587", "didcomm/aip2;env=rfc19"],
    keyType = "ed25519",
    keyAgreementType = "p256kw",
  } = {}) {
    this.agent = await loadFrameworks({
      name: this.label,
      mediaTypeProfiles: mediaTypeProfiles,
      keyType: keyType,
      keyAgreementType: keyAgreementType,
    });

    let mediatorURL = testConfig.mediatorEndPoint;
    let isDIDCommV2 = false;
    for (let mtp of mediaTypeProfiles) {
      if (mtp === "didcomm/v2") {
        isDIDCommV2 = true;
      }
    }

    let { invitation_did } = await connectToMediator(this.agent, mediatorURL, {
      isDIDCommV2: isDIDCommV2,
    });

    let conns = await getMediatorConnections(this.agent, { single: true });
    expect(conns).to.not.empty;

    return { connection_id: conns, router_did: invitation_did };
  }

  async createInvitation({ goal_code, from } = {}) {
    console.debug(
      "~ about to call mediatorclient.createInvitation() - mediatorclient: " +
        JSON.stringify(this.agent.mediatorclient)
    );
    console.debug(
      "  label: " + this.label + ", goal_code:" + goal_code + ", from: " + from
    );
    let response = await this.agent.mediatorclient.createInvitation({
      label: this.label,
      router_connection_id: await getMediatorConnections(this.agent, {
        single: true,
      }),
      goal_code,
      from,
    });

    console.debug(
      "createInvitation() called - invitation created: " +
        JSON.stringify(response)
    );

    if (response["invitation-v2"] !== null) {
      return response["invitation-v2"];
    }

    return response.invitation;
  }

  async acceptExchangeRequest(timeout) {
    return await waitForEvent(this.agent, {
      stateID: DIDEXCHANGE_STATE_REQUESTED,
      type: POST_STATE,
      topic: DIDEXCHANGE_STATE_TOPIC,
      timeout,
      callback: async (payload) => {
        await this.agent.didexchange.acceptExchangeRequest({
          id: payload.Properties.connectionID,
          router_connections: await getMediatorConnections(this.agent, {
            single: true,
          }),
        });
      },
    });
  }

  async destroy() {
    return await this.agent.destroy();
  }
}

/**
 * VerifierAdapter mocks verifier(relying party) adapter features.
 *
 * @param agent instance
 * @class
 */
export class VerifierAdapter extends Adapter {
  constructor(label) {
    super(label);
  }

  async init({
    mediaTypeProfiles = ["didcomm/aip2;env=rfc587", "didcomm/aip2;env=rfc19"],
    keyType = "ed25519",
    keyAgreementType = "p256kw",
  } = {}) {
    return await super.init({
      mediaTypeProfiles: mediaTypeProfiles,
      keyType: keyType,
      keyAgreementType: keyAgreementType,
    });
  }

  async acceptPresentationProposal(query = {}, timeout) {
    console.debug(
      "acceptPresentationProposal query:" + JSON.stringify(query, undefined, 2),
      "    timeout:" + timeout
    );
    return await waitForEvent(this.agent, {
      topic: PRESENT_PROOF_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        let id = uuidv4();
        let { myDID, theirDID, piid } = payload.Properties;
        // TODO create request_presentation based on DIDComm version. Right now, only DIDComm V1 is used.
        await this.agent.presentproof.sendRequestPresentation({
          my_did: myDID,
          their_did: theirDID,
          request_presentation: {
            will_confirm: true,
            formats: [
              {
                attach_id: id,
                format: "dif/presentation-exchange/definitions@v1.0",
              },
            ],
            "request_presentations~attach": [
              {
                "@id": id,
                lastmod_time: "0001-01-01T00:00:00Z",
                data: {
                  json: {
                    presentation_definition: query,
                  },
                },
              },
            ],
          },
        });
      },
    });
  }

  async acceptPresentProof({ timeout, redirectURL } = {}) {
    let presentation;
    await waitForEvent(this.agent, {
      topic: PRESENT_PROOF_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        let { Message, Properties } = payload;

        console.debug(
          "acceptPresentProof() Message:" +
            JSON.stringify(Message, undefined, 2),
          "    Properties:" + JSON.stringify(Properties, undefined, 2)
        );
        presentation = extractPresentation(Message);

        const { piid } = Properties;

        return this.agent.presentproof.acceptPresentation({
          piid,
          redirectURL,
        });
      },
    });

    return presentation;
  }

  async declinePresentProof({ timeout, redirectURL } = {}) {
    let presentation;
    await waitForEvent(this.agent, {
      topic: PRESENT_PROOF_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        let { Message, Properties } = payload;
        console.debug(
          "declinePresentProof() Message:" +
            JSON.stringify(Message, undefined, 2),
          "    Properties:" + JSON.stringify(Properties, undefined, 2)
        );
        presentation = extractPresentation(Message);

        const { piid } = Properties;

        return this.agent.presentproof.declinePresentation({
          piid,
          redirectURL,
          reason: "test",
        });
      },
    });

    return presentation;
  }
}

/**
 * Extracts Presentation JSON object from Message map based on DIDComm V1 or V2 formats.
 *
 * @param Message map instance
 */
function extractPresentation(Message) {
  let presentation;
  if (Message["presentations~attach"]) {
    // didcomm v1
    presentation = Message["presentations~attach"][0].data.json;
    console.log("didcomm v1 found");
  } else if (Message["attachments"]) {
    // didcomm v2
    presentation = Message["attachments"][0].data.json;
    console.log("didcomm v2 found");
  } else {
    console.error(
      "unrecognized presentation object: '" +
        JSON.stringify(Message, undefined, 2) +
        "'"
    );
  }

  return presentation;
}

/**
 * IssuerAdapter mocks issuer adapter features.
 *
 * @param agent instance
 * @class
 */
export class IssuerAdapter extends Adapter {
  constructor(label) {
    super(label);
  }

  async init({
    mediaTypeProfiles = ["didcomm/aip2;env=rfc587", "didcomm/aip2;env=rfc19"],
    keyType = "ed25519",
    keyAgreementType = "p256kw",
  } = {}) {
    return await super.init({
      mediaTypeProfiles: mediaTypeProfiles,
      keyType: keyType,
      keyAgreementType: keyAgreementType,
    });
  }

  async issue(...credential) {
    const keyType = "ED25519";

    const [keySet, recoveryKeySet, updateKeySet] = await Promise.all([
      this.agent.kms.createKeySet({ keyType }),
      this.agent.kms.createKeySet({ keyType }),
      this.agent.kms.createKeySet({ keyType }),
    ]);

    const createDIDRequest = {
      publicKeys: [
        {
          id: keySet.keyID,
          type: "Ed25519VerificationKey2018",
          value: keySet.publicKey,
          encoding: "Jwk",
          keyType,
          purposes: ["authentication"],
        },
        {
          id: recoveryKeySet.keyID,
          type: "Ed25519VerificationKey2018",
          value: recoveryKeySet.publicKey,
          encoding: "Jwk",
          keyType,
          recovery: true,
        },
        {
          id: updateKeySet.keyID,
          type: "Ed25519VerificationKey2018",
          value: updateKeySet.publicKey,
          encoding: "Jwk",
          keyType,
          update: true,
        },
      ],
    };

    let { didDocument } = await this.agent.didclient.createOrbDID(
      createDIDRequest
    );

    let resolveDID = async () =>
      await this.agent.vdr.resolveDID({ id: didDocument.id });
    await retryWithDelay(resolveDID, 10, 5000);

    let signVCs = await Promise.all(
      credential.map(async (credential) => {
        let { verifiableCredential } =
          await this.agent.verifiable.signCredential({
            credential,
            did: didDocument.id,
            signatureType: "Ed25519Signature2018",
          });

        return verifiableCredential;
      })
    );

    return signVCs;
  }

  async acceptCredentialProposal(
    { comment, manifest, fulfillment, noChallenge = false } = {},
    timeout
  ) {
    return await waitForEvent(this.agent, {
      topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        if (
          payload.Message.type &&
          payload.Message.type == MSG_TYPE_PROPOSE_CREDENTIAL_V3
        ) {
          await acceptProposalV3(this.agent, payload, {
            manifest,
            fulfillment,
            comment,
            noChallenge,
          });
        } else if (
          payload.Message["@type"] &&
          payload.Message["@type"] == MSG_TYPE_PROPOSE_CREDENTIAL_V2
        ) {
          await acceptProposalV2(this.agent, payload, {
            manifest,
            fulfillment,
            comment,
            noChallenge,
          });
        } else {
          console.error("unexpected message type received");
          return;
        }
      },
    });
  }

  async acceptRequestCredential({ timeout, credential, redirect } = {}) {
    let attachment;
    await waitForEvent(this.agent, {
      topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        const { Message, Properties, formats } = payload;
        const { piid } = Properties;

        if (Message["@type"] || Message["@id"]) {
          attachment = findAttachmentByFormat(
            Message.formats,
            Message["requests~attach"],
            "application/ld+json"
          );
          return await acceptRequestCredentialV2(this.agent, piid, {
            credential,
            redirect,
          });
        } else {
          if (!Message.attachments && Message.attachments.length === 0) {
            throw "no didcomm v2 attachments";
          }

          attachment = Message.attachments[0].data.json;
          return await acceptRequestCredentialV3(this.agent, piid, {
            credential,
            redirect,
          });
        }
      },
    });

    return attachment;
  }

  async declineCredentialProposal({ redirectURL } = {}, timeout) {
    return await waitForEvent(this.agent, {
      topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        let { piid } = payload.Properties;

        await this.agent.issuecredential.declineProposal({
          piid,
          redirectURL,
        });
      },
    });
  }

  async declineRequestCredential({ redirectURL } = {}, timeout) {
    await waitForEvent(this.agent, {
      topic: ISSUE_CREDENTIAL_ACTION_TOPIC,
      timeout,
      callback: async (payload) => {
        const { piid } = payload.Properties;

        return this.agent.issuecredential.declineRequest({
          piid,
          redirectURL,
        });
      },
    });
  }
}

async function acceptProposalV2(
  agent,
  payload,
  { manifest, fulfillment, comment, noChallenge }
) {
  let { piid } = payload.Properties;
  let attachID1 = uuidv4();
  let attachID2 = uuidv4();

  let formats = [];
  let attachments = [];

  if (manifest) {
    let attachId = uuidv4();
    formats.push({
      attach_id: attachId,
      format: ATTACH_FORMAT_CREDENTIAL_MANIFEST,
    });
    attachments.push({
      "@id": attachId,
      "mime-type": "application/json",
      data: {
        json: {
          credential_manifest: manifest,
          options: {
            challenge: noChallenge ? undefined : uuidv4(),
            domain: noChallenge ? undefined : uuidv4(),
          },
        },
      },
    });
  }

  if (fulfillment) {
    let attachId = uuidv4();
    formats.push({
      attach_id: attachId,
      format: ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
    });
    attachments.push({
      "@id": attachId,
      "mime-type": "application/json",
      data: {
        json: fulfillment,
      },
    });
  }

  await agent.issuecredential.acceptProposal({
    piid,
    offer_credential: {
      "@type": MSG_TYPE_OFFER_CREDENTIAL_V2,
      comment,
      formats,
      "offers~attach": attachments,
    },
  });
}

async function acceptProposalV3(
  agent,
  payload,
  { manifest, fulfillment, noChallenge }
) {
  let { piid } = payload.Properties;

  let attachments = [];

  if (manifest) {
    attachments.push({
      id: uuidv4(),
      media_type: "application/json",
      format: ATTACH_FORMAT_CREDENTIAL_MANIFEST,
      data: {
        json: {
          credential_manifest: manifest,
          options: {
            challenge: noChallenge ? undefined : uuidv4(),
            domain: noChallenge ? undefined : uuidv4(),
          },
        },
      },
    });
  }

  if (fulfillment) {
    attachments.push({
      id: uuidv4(),
      media_type: "application/json",
      format: ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
      data: {
        json: fulfillment,
      },
    });
  }

  await agent.issuecredential.acceptProposal({
    piid,
    offer_credential: {
      type: MSG_TYPE_OFFER_CREDENTIAL_V3,
      attachments: attachments,
    },
  });
}

async function acceptRequestCredentialV2(
  agent,
  piid,
  { credential, redirect }
) {
  let attachID = uuidv4();
  let icFormats = [];
  let icAttachments = [];

  if (credential) {
    icFormats.push({
      attach_id: attachID,
      format: ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
    });

    icAttachments.push({
      "@id": attachID,
      "mime-type": "application/json",
      data: {
        json: credential,
      },
    });
  }

  await agent.issuecredential.acceptRequest({
    piid,
    issue_credential: {
      "@type": "https://didcomm.org/issue-credential/2.0/issue-credential",
      formats: icFormats,
      "credentials~attach": icAttachments,
      "~web-redirect": {
        status: "OK",
        url: redirect,
      },
    },
  });
}

async function acceptRequestCredentialV3(
  agent,
  piid,
  { credential, redirect }
) {
  let attachments = [];

  if (credential) {
    attachments.push({
      id: uuidv4(),
      media_type: "application/json",
      format: ATTACH_FORMAT_CREDENTIAL_FULFILLMENT,
      data: {
        json: credential,
      },
    });
  }

  await agent.issuecredential.acceptRequest({
    piid,
    issue_credential: {
      type: "https://didcomm.org/issue-credential/3.0/issue-credential",
      attachments: attachments,
      "web-redirect": {
        status: "OK",
        url: redirect,
      },
    },
  });
}
