/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/mediatorclient"
)

// MediatorClient contains necessary fields to support its operations.
type MediatorClient struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// Connect  connects agent to given router endpoint.
func (mc *MediatorClient) Connect(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return mc.createRespEnvelope(request, mediatorclient.Connect)
}

// CreateInvitation creates out-of-band invitation from one of the mediator connections.
func (mc *MediatorClient) CreateInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return mc.createRespEnvelope(request, mediatorclient.CreateInvitation)
}

// SendCreateConnectionRequest sends create connection request to mediator.
func (mc *MediatorClient) SendCreateConnectionRequest(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return mc.createRespEnvelope(request, mediatorclient.SendCreateConnectionRequest)
}

func (mc *MediatorClient) createRespEnvelope(request *models.RequestEnvelope,
	endpoint string,
) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        mc.URL,
		token:      mc.Token,
		httpClient: mc.httpClient,
		endpoint:   mc.endpoints[endpoint],
		request:    request,
	})
}
