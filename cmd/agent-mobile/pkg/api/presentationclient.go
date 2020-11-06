/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// PresentationClient  defines methods for the PresentationClient.
type PresentationClient interface {
	// SavePresentation in the SDS.
	SavePresentation(request *models.RequestEnvelope) *models.ResponseEnvelope
}
