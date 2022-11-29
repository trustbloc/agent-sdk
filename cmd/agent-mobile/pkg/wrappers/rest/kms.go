/*
 Copyright SecureKey Technologies Inc. All Rights Reserved.
 Copyright Avast Software. All Rights Reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package rest //nolint: dupl

import (
	"github.com/hyperledger/aries-framework-go/pkg/controller/command/kms"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

// KMS contains necessary fields to support its operations.
type KMS struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}

// CreateKeySet create a new public/private encryption and signature key pairs set.
func (k *KMS) CreateKeySet(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return k.createRespEnvelope(request, kms.CreateKeySetCommandMethod)
}

// ImportKey imports a key.
func (k *KMS) ImportKey(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return k.createRespEnvelope(request, kms.ImportKeyCommandMethod)
}

func (k *KMS) createRespEnvelope(request *models.RequestEnvelope, endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        k.URL,
		token:      k.Token,
		httpClient: k.httpClient,
		endpoint:   k.endpoints[endpoint],
		request:    request,
	})
}
