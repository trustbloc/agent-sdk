/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest // nolint: dupl

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/presentationclient"
)

// PresentationClient contains necessary fields to support its operations.
type PresentationClient struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// SavePresentation in the SDS.
func (pc *PresentationClient) SavePresentation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return pc.createRespEnvelope(request, presentationclient.SavePresentationCommandMethod)
}

func (pc *PresentationClient) createRespEnvelope(r *models.RequestEnvelope, endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        pc.URL,
		token:      pc.Token,
		httpClient: pc.httpClient,
		endpoint:   pc.endpoints[endpoint],
		request:    r,
	})
}
