/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
Copyright Avast Software. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import (
	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

// VerifiableController defines methods for the verifiable controller.
type VerifiableController interface { //nolint: interfacebloat
	// ValidateCredential validates the verifiable credential.
	ValidateCredential(request *models.RequestEnvelope) *models.ResponseEnvelope

	// SaveCredential saves the verifiable credential to the store.
	SaveCredential(request *models.RequestEnvelope) *models.ResponseEnvelope

	// SavePresentation saves the presentation to the store.
	SavePresentation(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetCredential retrieves the verifiable credential from the store.
	GetCredential(request *models.RequestEnvelope) *models.ResponseEnvelope

	// SignCredential adds proof to given verifiable credential.
	SignCredential(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetPresentation retrieves the verifiable presentation from the store.
	GetPresentation(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetCredentialByName retrieves the verifiable credential by name from the store.
	GetCredentialByName(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetCredentials retrieves the verifiable credential records containing name and fields of interest.
	GetCredentials(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GetPresentations retrieves the verifiable presentation records containing name and fields of interest.
	GetPresentations(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GeneratePresentation generates a verifiable presentation from a verifiable credential.
	GeneratePresentation(request *models.RequestEnvelope) *models.ResponseEnvelope

	// GeneratePresentationByID generates verifiable presentation from a stored verifiable credential.
	GeneratePresentationByID(request *models.RequestEnvelope) *models.ResponseEnvelope

	// RemoveCredentialByName will remove a VC that matches the specified name from the verifiable store.
	RemoveCredentialByName(request *models.RequestEnvelope) *models.ResponseEnvelope

	// RemovePresentationByName will remove a VP that matches the specified name from the verifiable store.
	RemovePresentationByName(request *models.RequestEnvelope) *models.ResponseEnvelope
}
