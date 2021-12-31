/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest //nolint:dupl // outofbandv2 is separate from outofband

import (
	"github.com/hyperledger/aries-framework-go/pkg/controller/command/outofbandv2"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

// OutOfBandV2 contains necessary fields to support its operations.
type OutOfBandV2 struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// CreateInvitation creates and saves an out-of-band 2.0 invitation.
func (oob *OutOfBandV2) CreateInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return oob.createRespEnvelope(request, outofbandv2.CreateInvitation)
}

// AcceptInvitation from another agent and return the ID of the new connection records.
func (oob *OutOfBandV2) AcceptInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return oob.createRespEnvelope(request, outofbandv2.AcceptInvitation)
}

func (oob *OutOfBandV2) createRespEnvelope(request *models.RequestEnvelope, endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        oob.URL,
		token:      oob.Token,
		httpClient: oob.httpClient,
		endpoint:   oob.endpoints[endpoint],
		request:    request,
	})
}
