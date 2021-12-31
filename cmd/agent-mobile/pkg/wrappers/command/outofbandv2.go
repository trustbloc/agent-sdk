/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command //nolint:dupl // outofbandv2 is separate

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"
	"github.com/hyperledger/aries-framework-go/pkg/controller/command/outofbandv2"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

// OutOfBandV2 contains necessary fields to support its operations.
type OutOfBandV2 struct {
	handlers map[string]command.Exec
}

// CreateInvitation creates and saves an out-of-band 2.0 invitation.
func (oob *OutOfBandV2) CreateInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := outofbandv2.CreateInvitationArgs{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(oob.handlers[outofbandv2.CreateInvitation], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}

// AcceptInvitation from another agent and return the ID of the new connection records.
func (oob *OutOfBandV2) AcceptInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := outofbandv2.AcceptInvitationArgs{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(oob.handlers[outofbandv2.AcceptInvitation], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}
