/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// VDRController defines methods for the VDR controller.
type VDRController interface {
	// ResolveDID resolve did.
	ResolveDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// SaveDID saves the did doc to the store.
	SaveDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetDID retrieves the did from the store.
	GetDID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetDIDRecords retrieves the did doc containing name and didID.
	GetDIDRecords(request *models.RequestEnvelope) *models.ResponseEnvelope
}
