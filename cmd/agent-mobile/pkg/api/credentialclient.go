/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// CredentialClient  defines methods for the CredentialClient.
type CredentialClient interface {
	// SaveCredential received in the request.
	SaveCredential(request *models.RequestEnvelope) *models.ResponseEnvelope
}
