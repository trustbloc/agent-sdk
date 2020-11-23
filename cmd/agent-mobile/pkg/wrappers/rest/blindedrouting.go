/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest // nolint: dupl

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
)

// BlindedRouting contains necessary fields to support its operations.
type BlindedRouting struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// SendDIDDocRequest sends DID doc request over a connection.
func (br *BlindedRouting) SendDIDDocRequest(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return br.createRespEnvelope(request, blindedrouting.SendDIDDocRequest)
}

// SendRegisterRouteRequest sends register route request as a response to reply from send DID doc request.
func (br *BlindedRouting) SendRegisterRouteRequest(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return br.createRespEnvelope(request, blindedrouting.SendRegisterRouteRequest)
}

func (br *BlindedRouting) createRespEnvelope(request *models.RequestEnvelope,
	endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        br.URL,
		token:      br.Token,
		httpClient: br.httpClient,
		endpoint:   br.endpoints[endpoint],
		request:    request,
	})
}
