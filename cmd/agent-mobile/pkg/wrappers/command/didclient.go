/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
)

// DIDClient contains necessary fields to support its operations.
type DIDClient struct {
	handlers map[string]command.Exec
}

// CreateTrustBlocDID creates a new trust bloc DID.
func (de *DIDClient) CreateTrustBlocDID(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := didclient.CreateBlocDIDRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(de.handlers[didclient.CreateTrustBlocDIDCommandMethod], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}

// CreatePeerDID creates a new peer DID.
func (de *DIDClient) CreatePeerDID(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := didclient.CreatePeerDIDRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(de.handlers[didclient.CreatePeerDIDCommandMethod], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}
