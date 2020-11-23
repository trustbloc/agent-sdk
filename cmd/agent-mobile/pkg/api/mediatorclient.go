/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// MediatorClient  defines methods for the MediatorClient.
type MediatorClient interface {
	// Connect  connects agent to given router endpoint.
	Connect(request *models.RequestEnvelope) *models.ResponseEnvelope

	// CreateInvitation creates out-of-band invitation from one of the mediator connections.
	CreateInvitation(request *models.RequestEnvelope) *models.ResponseEnvelope

	// SendCreateConnectionRequest sends create connection request to mediator.
	SendCreateConnectionRequest(request *models.RequestEnvelope) *models.ResponseEnvelope
}
