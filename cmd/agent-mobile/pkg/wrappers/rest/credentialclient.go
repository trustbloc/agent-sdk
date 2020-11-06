/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest // nolint: dupl

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/credentialclient"
)

// CredentialClient contains necessary fields to support its operations.
type CredentialClient struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// SaveCredential received in the request.
func (cc *CredentialClient) SaveCredential(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return cc.createRespEnvelope(request, credentialclient.SaveCredentialCommandMethod)
}

func (cc *CredentialClient) createRespEnvelope(r *models.RequestEnvelope, endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        cc.URL,
		token:      cc.Token,
		httpClient: cc.httpClient,
		endpoint:   cc.endpoints[endpoint],
		request:    r,
	})
}
