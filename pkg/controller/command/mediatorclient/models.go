/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mediatorclient

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
)

// ConnectionRequest model
//
// This is used for connecting to given router.
//
type ConnectionRequest struct {
	// Invitation is out-of-band invitation from mediator.
	Invitation *outofband.Invitation `json:"invitation"`

	// MyLabel is custom label to be used as receiver label of this invitation
	// Optional: if missing, agent default label will be used.
	MyLabel string `json:"mylabel,omitempty"`

	// StateCompleteMessageType is optional did exchange completion notification message type from inviter.
	// If provided, then agent will wait for notification of this message type from inviter before performing
	// mediator registration.
	// If not provided, then this agent will go ahead with mediator registration once did exchange state is
	// completed at invitee.
	StateCompleteMessageType string `json:"stateCompleteMessageType,omitempty"`
}

// ConnectionResponse contains response.
type ConnectionResponse struct {
	ConnectionID   string   `json:"connectionID"`
	RouterEndpoint string   `json:"routerEndpoint"`
	RoutingKeys    []string `json:"routingKeys"`
}

// CreateInvitationRequest model
//
// This is used for creating an invitation using mediator.
//
type CreateInvitationRequest struct {
	Label     string        `json:"label"`
	Goal      string        `json:"goal"`
	GoalCode  string        `json:"goal_code"`
	Service   []interface{} `json:"service"`
	Protocols []string      `json:"protocols"`
}

// CreateInvitationResponse model
//
// Response for creating invitation through mediator.
//
type CreateInvitationResponse struct {
	// Invitation is out-of-band invitation from mediator.
	Invitation *outofband.Invitation `json:"invitation"`
}

// CreateConnectionRequest model
//
// This is used for sending create connection request.
//
type CreateConnectionRequest struct {
	DIDDocument json.RawMessage `json:"didDoc"`
}

// CreateConnectionResponse model
//
// This is used for getting create connection response.
//
type CreateConnectionResponse struct {
	Payload json.RawMessage `json:"payload"`
}
