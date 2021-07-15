/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package rest

import (
	cmdvcwallet "github.com/hyperledger/aries-framework-go/pkg/controller/command/vcwallet"

	"github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"
)

// VCWallet contains necessary fields to support its operations.
type VCWallet struct {
	httpClient httpClient
	endpoints  map[string]*endpoint

	URL   string
	Token string
}


func (wallet *VCWallet) CreateProfile(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.CreateProfileMethod)
}

func (wallet *VCWallet) UpdateProfile(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.UpdateProfileMethod)
}

func (wallet *VCWallet) ProfileExists(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.ProfileExistsMethod)
}

func (wallet *VCWallet) Open(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.OpenMethod)
}

func (wallet *VCWallet) Close(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.CloseMethod)
}

func (wallet *VCWallet) Add(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.AddMethod)
}


func (wallet *VCWallet) Remove(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.RemoveMethod)
}

func (wallet *VCWallet) Get(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.GetMethod)
}

func (wallet *VCWallet) GetAll(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.GetAllMethod)
}

func (wallet *VCWallet) Query(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.QueryMethod)
}

func (wallet *VCWallet) Issue(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.IssueMethod)
}

func (wallet *VCWallet) Prove(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.ProveMethod)
}

func (wallet *VCWallet) Verify(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.VerifyMethod)
}

func (wallet *VCWallet) Derive(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.DeriveMethod)
}

func (wallet *VCWallet) CreateKeyPair(request *models.RequestEnvelope) *models.ResponseEnvelope {
	return wallet.createRespEnvelope(request, cmdvcwallet.CreateKeyPairMethod)
}


func (wallet *VCWallet) createRespEnvelope(request *models.RequestEnvelope, endpoint string) *models.ResponseEnvelope {
	return exec(&restOperation{
		url:        wallet.URL,
		token:      wallet.Token,
		httpClient: wallet.httpClient,
		endpoint:   wallet.endpoints[endpoint],
		request:    request,
	})
}
