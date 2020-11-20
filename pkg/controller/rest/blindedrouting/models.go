/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package blindedrouting

import (
	"github.com/trustbloc/agent-sdk/pkg/controller/command/blindedrouting"
)

// didDocRequest model
//
// Request for sending did doc request.
//
// swagger:parameters didDocRequest
type didDocRequest struct { // nolint: unused,deadcode
	// Params for sending did doc request.
	//
	// in: body
	// required: true
	Request blindedrouting.DIDDocRequest
}

// didDocResponse model
//
// Response from a connection for a did doc request.
//
// swagger:response didDocResponse
type didDocResponse struct {
	// in: body
	Response blindedrouting.DIDDocResponse
}

// registerRouteRequest model
//
// This is used for sending register route request, often sent as response to `didDocResponse` from
// send did doc request operation.
//
// swagger:parameters registerRoute
type registerRouteRequest struct { // nolint: unused,deadcode
	// Params for sending did doc request.
	//
	// in: body
	// required: true
	Request blindedrouting.RegisterRouteRequest
}

// registerRouteResponse model
//
// Response for a register route request sent.
//
// swagger:response registerRouteResponse
type registerRouteResponse struct {
	// in: body
	Response blindedrouting.RegisterRouteResponse
}
