/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package mediatorclient

import (
	"github.com/hyperledger/aries-framework-go/pkg/client/outofband"
)

// ConnectionRequest model
//
// This is used for connecting to given router.
//
type ConnectionRequest struct {
	// Invitation is out-of-band invitation from mediator.
	Invitation *outofband.Invitation `json:"invitation"`

	// MyLabel is custom label to be used as receiver label of this invitation
	// Optional: if missing, agent default label will be used.
	MyLabel string `json:"mylabel"`
}

// ConnectionResponse contains response.
type ConnectionResponse struct {
	ConnectionID   string   `json:"connection_id"`
	RouterEndpoint string   `json:"routerEndpoint"`
	RoutingKeys    []string `json:"routingKeys"`
}
