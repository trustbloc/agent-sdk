/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api // nolint: dupl

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// MediatorController defines methods for the Mediator controller.
type MediatorController interface {

	// Register registers the agent with the router.
	Register(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Unregister unregisters the agent with the router.
	Unregister(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Connections returns router`s connections.
	Connections(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Reconnect sends noop message to reestablish a connection when there is no other reason to message the mediator.
	Reconnect(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Status returns details about pending messages for given connection.
	Status(request *models.RequestEnvelope) *models.ResponseEnvelope

	// BatchPickup dispatches pending messages for given connection.
	BatchPickup(request *models.RequestEnvelope) *models.ResponseEnvelope
}
