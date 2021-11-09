/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest // nolint: dupl

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
	"github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
)

// DIDClient contains necessary fields to support its operations.
type DIDClient struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// CreateOrbDID creates a new Orb DID.
func (dc *DIDClient) CreateOrbDID(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return dc.createRespEnvelope(request, didclient.CreateOrbDIDCommandMethod)
}

// ResolveOrbDID resolve Orb DID.
func (dc *DIDClient) ResolveOrbDID(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return dc.createRespEnvelope(request, didclient.ResolveOrbDIDCommandMethod)
}

// CreatePeerDID creates a new peer DID.
func (dc *DIDClient) CreatePeerDID(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return dc.createRespEnvelope(request, didclient.CreatePeerDIDCommandMethod)
}

func (dc *DIDClient) createRespEnvelope(request *models.RequestEnvelope, endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        dc.URL,
		token:      dc.Token,
		httpClient: dc.httpClient,
		endpoint:   dc.endpoints[endpoint],
		request:    request,
	})
}
