/*
 Copyright SecureKey Technologies Inc. All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package api //nolint:dupl

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// LDController defines methods for the JSON-LD controller.
type LDController interface {
	// AddContexts adds JSON-LD contexts to the underlying storage.
	AddContexts(request *models.RequestEnvelope) *models.ResponseEnvelope

	// AddRemoteProvider adds remote provider and JSON-LD contexts from that provider.
	AddRemoteProvider(request *models.RequestEnvelope) *models.ResponseEnvelope

	// RefreshRemoteProvider updates contexts from the remote provider.
	RefreshRemoteProvider(request *models.RequestEnvelope) *models.ResponseEnvelope

	// DeleteRemoteProvider deletes remote provider and contexts from that provider.
	DeleteRemoteProvider(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetAllRemoteProviders gets all remote providers.
	GetAllRemoteProviders(request *models.RequestEnvelope) *models.ResponseEnvelope

	// RefreshAllRemoteProviders updates contexts from all remote providers.
	RefreshAllRemoteProviders(request *models.RequestEnvelope) *models.ResponseEnvelope
}
