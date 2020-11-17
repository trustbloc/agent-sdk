/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mediatorclient

import (
	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
)

// connectionRequest model
//
// Request for connecting agent to given mediator with invitation provided.
//
// swagger:parameters connect
type connectionRequest struct { // nolint: unused,deadcode
	// Params for connecting to mediator.
	//
	// in: body
	// required: true
	Request mediatorclient.ConnectionRequest
}

// connectionResponse model
//
// This is used as the response model for connecting agent to a mediator.
//
// swagger:response connectionResponse
type connectionResponse struct {
	// in: body
	Response mediatorclient.ConnectionResponse
}

// createInvitationRequest model
//
// Request for creating out-of-band invitation through mediator client.
//
// swagger:parameters createMediatorInvitation
type createInvitationRequest struct { // nolint: unused,deadcode
	// Params for creating invitation.
	//
	// in: body
	// required: true
	Request mediatorclient.CreateInvitationRequest
}

// createInvitationResponse model
//
//  Response of creating out-of-band invitation through mediator client.
//
// swagger:response createInvitationResponse
type createInvitationResponse struct { // nolint: unused,deadcode
	// in: body
	Response mediatorclient.CreateInvitationResponse
}
