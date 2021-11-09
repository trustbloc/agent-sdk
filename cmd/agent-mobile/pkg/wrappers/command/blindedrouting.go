/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
)

// BlindedRouting contains necessary fields to support its operations.
type BlindedRouting struct {
	handlers map[string]command.Exec
}

// SendDIDDocRequest sends DID doc request over a connection.
func (br *BlindedRouting) SendDIDDocRequest(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := blindedrouting.DIDDocRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(br.handlers[blindedrouting.SendDIDDocRequest], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}

// SendRegisterRouteRequest sends register route request as a response to reply from send DID doc request.
func (br *BlindedRouting) SendRegisterRouteRequest(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := blindedrouting.RegisterRouteRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(br.handlers[blindedrouting.SendRegisterRouteRequest], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}
