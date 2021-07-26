/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package didclient

import (
	"github.com/hyperledger/aries-framework-go/pkg/doc/did"

	"github.com/trustbloc/agent-sdk/pkg/controller/command/didclient"
)

// createOrbDIDRequest model
//
// Request to create a new trust bloc DID.
//
// swagger:parameters createOrbDID
type createOrbDIDRequest struct { // nolint: unused,deadcode
	// Params for creating a Orb DID.
	//
	// in: body
	// required: true
	Request didclient.CreateOrbDIDRequest
}

// createPeerDIDRequest model
//
// Request to create a new peer DID.
//
// swagger:parameters createPeerDID
type createPeerDIDRequest struct { // nolint: unused,deadcode
	// Params for creating a TrustBlocDID.
	//
	// in: body
	// required: true
	Request didclient.CreatePeerDIDRequest
}

// createDIDResp model
//
// This is used as the response model for create TrustBloc/ DID operations.
//
// swagger:response createDIDResp
type createDIDResp struct { // nolint: unused,deadcode
	// in: body
	Response *did.DocResolution
}
