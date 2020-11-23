/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
)

// MediatorClient contains necessary fields to support its operations.
type MediatorClient struct {
	handlers map[string]command.Exec
}

// Connect connects agent to given router endpoint.
func (mc *MediatorClient) Connect(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := mediatorclient.ConnectionRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(mc.handlers[mediatorclient.Connect], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}

// CreateInvitation creates out-of-band invitation from one of the mediator connections.
func (mc *MediatorClient) CreateInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := mediatorclient.CreateInvitationRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(mc.handlers[mediatorclient.CreateInvitation], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}

// SendCreateConnectionRequest sends create connection request to mediator.
func (mc *MediatorClient) SendCreateConnectionRequest(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := mediatorclient.CreateConnectionRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(mc.handlers[mediatorclient.SendCreateConnectionRequest], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}
