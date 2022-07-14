/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import jp from "jsonpath";
import { v4 as uuidv4 } from "uuid";

export const PRE_STATE = "pre_state";
export const POST_STATE = "post_state";

const DEFAULT_TIMEOUT = 120000;
const DEFAULT_TIMEOUT_ERR = "time out while waiting for event";
const DEFAULT_TOPIC = "all";

export function waitForEvent(
  agent,
  {
    timeout = DEFAULT_TIMEOUT,
    timeoutError = DEFAULT_TIMEOUT_ERR,
    topic = DEFAULT_TOPIC,
    type,
    stateID,
    connectionID,
    callback = () => {},
  } = {}
) {
  return new Promise((resolve, reject) => {
    setTimeout(() => reject(new Error(timeoutError)), timeout);
    const stop = agent.startNotifier(
      (event) => {
        try {
          let payload = event.payload;

          if (
            connectionID &&
            payload.Properties &&
            payload.Properties.connectionID !== connectionID
          ) {
            return;
          }

          if (stateID && payload.StateID !== stateID) {
            return;
          }

          if (type && payload.Type !== type) {
            return;
          }

          stop();

          callback(payload);

          resolve(payload);
        } catch (e) {
          stop();
          reject(e);
        }
      },
      [topic]
    );
  });
}

// filter and return defined properties only
export const definedProps = (obj) =>
  Object.fromEntries(Object.entries(obj).filter(([k, v]) => v !== undefined));

/**
 *  Scans through @see {@link https://identity.foundation/presentation-exchange/#presentation-submission|Presentation Submission} descriptor map and groups results by descriptor IDs [descriptor_id -> Array of verifiable Credentials].
 *  In many cases, a single input descriptor can find multiple credentials.
 *  So this function is useful in cases of grouping results by input descriptor ID and giving option to user to choose only one to avoid over sharing.
 *
 *  TODO: support for path_nested in descriptor map.
 *
 *  @param {Array<Object>} query - array of query, one of which could have produced the presentation.
 *  @param {Object} presentation - presentation response of query. If `presentation_submission` is missing, then normalization will treat this as non presentation exchange query
 *  and normalization logic will only flatten the credentials (grouping duplicate results logic won't be applied).
 *
 * @returns {Array<Object>} - Normalized result array with each objects containing input descriptor id, name, purpose, format and array of credentials.
 */
export const normalizePresentationSubmission = (query, presentation) => {
  if (!presentation.presentation_submission) {
    return presentation.verifiableCredential.map((credential) => {
      return {
        id: uuidv4(),
        credentials: [credential],
      };
    });
  }

  const queryMatches = jp.query(
    query,
    `$[?(@.type=="PresentationExchange")].credentialQuery[?(@.id=="${presentation.presentation_submission.definition_id}")]`
  );
  if (queryMatches.length == 0) {
    throw "couldn't find matching definition in query";
  }

  const queryMatch = queryMatches[0];

  let result = {};
  const _forEachDescriptor = (descr) => {
    const credentials = jp.query(presentation, descr.path);

    if (result[descr.id]) {
      result[descr.id].credentials.push(...credentials);
      return;
    }

    const inputDescrs = jp.query(
      queryMatch,
      `$..input_descriptors[?(@.id=="${descr.id}")]`
    );
    if (inputDescrs.length == 0) {
      throw "invalid result, unable to find input descriptor in query.";
    }

    result[descr.id] = {
      id: descr.id,
      name: inputDescrs[0].name,
      purpose: inputDescrs[0].purpose,
      format: descr.format,
      credentials,
    };
  };

  presentation.presentation_submission.descriptor_map.forEach(
    _forEachDescriptor
  );

  return Object.values(result);
};

/**
 *  Updates given presentation submission presentation by removing duplicate descriptor map entries.
 *
 *  Descriptor map might contain single input descriptor ID mapped to multiple credentials.
 *  So returning PresentationSubmission presentation will retain only mappings mentioned in updates Object{<inputDescriptorID>:<credentialID>} parameter.
 */
export const updatePresentationSubmission = (presentation, updates) => {
  if (!presentation.presentation_submission) {
    return presentation;
  }

  let verifiableCredential = [];
  let descriptorMap = [];
  const _forEach = (descriptor) => {
    const vcSelected = updates[descriptor.id];
    const vcMapped = jp.query(presentation, descriptor.path);

    if (vcMapped.length > 0 && vcMapped[0].id == vcSelected) {
      verifiableCredential.push(vcMapped[0]);
      descriptor.path = `$.verifiableCredential[${
        verifiableCredential.length - 1
      }]`;
      descriptorMap.push(descriptor);
    }
  };

  presentation.presentation_submission.descriptor_map.forEach(_forEach);

  presentation.verifiableCredential = verifiableCredential;
  presentation.presentation_submission.descriptor_map = descriptorMap;

  return presentation;
};

/**
 *  Finds attachment by given format.
 *  Supporting Attachment Format from DIDComm V1.
 *
 *  Note: Currently finding only one attachment per format.
 */
export const findAttachmentByFormat = (formats, attachments, format) => {
  const formatsFound = jp.query(formats, `$[?(@.format=="${format}")]`);

  let attachment;
  if (formatsFound.length > 0) {
    const { attach_id } = formatsFound[0];

    // read attachment
    const attachmentsFound = jp.query(
      attachments,
      `$[?(@.@id=="${attach_id}")]`
    );
    if (attachmentsFound.length == 1) {
      attachment = attachmentsFound[0].data.json;
    } else {
      throw `invalid attachment found in offer credential message for format : ${format}`;
    }
  }

  return attachment;
};

/**
 *  Finds attachment by given format.
 *  Supporting Attachment Format from DIDComm V2.
 *
 *  Note: Currently finding only one attachment per format.
 */
export const findAttachmentByFormatV2 = (attachments, format) => {
  const attachmentsFound = jp.query(attachments, `$[?(@.format=="${format}")]`);

  if (attachmentsFound.length == 1) {
    return attachmentsFound[0].data.json;
  }

  return;
};

/**
 *  Reads out-of-band invitation goal code.
 *  Supports DIDComm V1 & V2
 */
export const extractOOBGoalCode = (oob) => {
  if (oob["@type"] == "https://didcomm.org/out-of-band/1.0/invitation") {
    return oob.goal_code;
  } else if (
    oob.type == "https://didcomm.org/out-of-band/2.0/invitation" &&
    oob.body
  ) {
    // support for both goal_code & goal-code.
    if (oob.body["goal-code"]) {
      return oob.body["goal-code"];
    } else if (oob.body["goal_code"]) {
      return oob.body["goal_code"];
    }
  }

  return null;
};

/**
 * Wait for given duration in millisecond and return promise.
 */
export function waitFor(duration) {
  return new Promise((resolve, reject) => {
    setTimeout(() => {
      resolve();
    }, duration);
  });
}

export function createKeyPair(alg, curve, extractable, keyUsages) {
  return window.crypto.subtle.generateKey(
      {
        name: alg,
        namedCurve: curve,
      },
      extractable,
      keyUsages,
  );
}

