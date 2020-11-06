/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package command // nolint: dupl

import (
	"encoding/json"

	"github.com/hyperledger/aries-framework-go/pkg/controller/command"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/credentialclient"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/sdscomm"
)

// CredentialClient contains necessary fields to support its operations.
type CredentialClient struct {
	handlers map[string]command.Exec
}

// SaveCredential received in the request.
func (cc *CredentialClient) SaveCredential(request *models.RequestEnvelope) *models.ResponseEnvelope {
	args := sdscomm.SaveCredentialToSDSRequest{}

	if err := json.Unmarshal(request.Payload, &args); err != nil {
		return &models.ResponseEnvelope{Error: &models.CommandError{Message: err.Error()}}
	}

	response, cmdErr := exec(cc.handlers[credentialclient.SaveCredentialCommandMethod], args)
	if cmdErr != nil {
		return &models.ResponseEnvelope{Error: cmdErr}
	}

	return &models.ResponseEnvelope{Payload: response}
}
