/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package blindedrouting

import (
	"encoding/json"
)

// DIDDocRequest model
//
// This is used for sending did doc request to a connection.
//
type DIDDocRequest struct {
	// ConnectionID of the connection to which did doc request to be sent.
	ConnectionID string `json:"connectionID"`
}

// DIDDocResponse model
//
// This is the response from a connection for a  did doc request.
//
type DIDDocResponse struct {
	// Payload contains response from a connection for a did doc request.
	Payload json.RawMessage `json:"payload"`
}

// RegisterRouteRequest model
//
// This is used for sending register route request, often sent as response to `DIDDocResponse` from
// send did doc request command.
//
type RegisterRouteRequest struct {
	// MessageID of the conversation to which this request has to be sent.
	// '@id' of the previous response from this connection to maintain communication context.
	// previous response --> message ID from DIDDocResponse.
	MessageID string `json:"messageID"`

	// DIDDocument to be shared in raw format.
	DIDDocument json.RawMessage `json:"didDoc"`
}

// RegisterRouteResponse model
//
// This is the response for a register route request sent.
//
type RegisterRouteResponse struct {
	// Payload contains response from a connection for a register route request.
	Payload json.RawMessage `json:"payload"`
}
