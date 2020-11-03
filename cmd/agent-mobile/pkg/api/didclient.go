/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// DIDClient  defines methods for the DIDClient.
type DIDClient interface {
	// CreateTrustBlocDID creates a new trust bloc DID.
	CreateTrustBlocDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// CreatePeerDID creates a new peer DID.
	CreatePeerDID(request *models.RequestEnvelope) *models.ResponseEnvelope
}
