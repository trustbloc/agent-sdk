/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

// BlindedRoutingController  defines methods for the MediatorClient.
type BlindedRoutingController interface {
	// SendDIDDocRequest sends DID doc request over a connection.
	SendDIDDocRequest(request *models.RequestEnvelope) *models.ResponseEnvelope

	// SendRegisterRouteRequest sends register route request as a response to reply from send DID doc request.
	SendRegisterRouteRequest(request *models.RequestEnvelope) *models.ResponseEnvelope
}
