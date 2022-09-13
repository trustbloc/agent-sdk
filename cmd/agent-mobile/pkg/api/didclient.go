/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// DIDClient  defines methods for the DIDClient.
type DIDClient interface {
	// CreateOrbDID creates a new orb DID.
	CreateOrbDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// CreatePeerDID creates a new peer DID.
	CreatePeerDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// ResolveOrbDID resolve orb DID.
	ResolveOrbDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// ResolveWebDIDFromOrbDID resolve orb DID.
	ResolveWebDIDFromOrbDID(request *models.RequestEnvelope) *models.ResponseEnvelope
}
