/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package api

import "github.com/trustbloc/agent-sdk/cmd/agent-mobile/pkg/wrappers/models"

// VCWalletController is a Verifiable Credential Wallet based on Universal Wallet 2020
// https://w3c-ccg.github.io/universal-wallet-interop-spec/#interface.
//
// Refer to [OpenAPI spec](docs/rest/openapi_spec.md#generate-openapi-spec) for
// input params and output return json values.
type VCWalletController interface {

	// Creates new wallet profile and returns error if wallet profile is already created.
	CreateProfile(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Updates an existing wallet profile and returns error if profile doesn't exists.
	UpdateProfile(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Checks if profile exists for given wallet user.
	ProfileExists(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Unlocks given wallet's key manager instance & content store and
	// returns a authorization token to be used for performing wallet operations.
	Open(request *models.RequestEnvelope) *models.ResponseEnvelope

	// Expires token issued to this VC wallet, removes wallet's key manager instance and closes wallet content store.
	// returns response containing bool flag false if token is not found or already expired for this wallet user.
	Close(request *models.RequestEnvelope) *models.ResponseEnvelope

	// adds given data model to wallet content store.
	//
	//   Supported data models:
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Collection
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Credential
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#DIDResolutionResponse
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#meta-data
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#connection
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Key
	Add(request *models.RequestEnvelope) *models.ResponseEnvelope

	// removes given content from wallet content store.
	//
	//   Supported data models:
	//	     - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Collection
	//	     - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Credential
	//	     - https://w3c-ccg.github.io/universal-wallet-interop-spec/#DIDResolutionResponse
	//	     - https://w3c-ccg.github.io/universal-wallet-interop-spec/#meta-data
	//	     - https://w3c-ccg.github.io/universal-wallet-interop-spec/#connection
	Remove(request *models.RequestEnvelope) *models.ResponseEnvelope

	// gets content from wallet content store.
	//
	//   Supported data models:
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Collection
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Credential
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#DIDResolutionResponse
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#meta-data
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#connection
	Get(request *models.RequestEnvelope) *models.ResponseEnvelope

	// gets all contents from wallet content store for given content type.
	//
	//   Supported data models:
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Collection
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#Credential
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#DIDResolutionResponse
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#meta-data
	//	    - https://w3c-ccg.github.io/universal-wallet-interop-spec/#connection
	GetAll(request *models.RequestEnvelope) *models.ResponseEnvelope

	// runs query against wallet credential contents and returns presentation containing credential results.
	//
	// This function may return multiple presentations as a result based on combination of query types used.
	//
	// https://w3c-ccg.github.io/universal-wallet-interop-spec/#query
	//
	//	Supported Query Types:
	//	    - https://www.w3.org/TR/json-ld11-framing
	//	    - https://identity.foundation/presentation-exchange
	//	    - https://w3c-ccg.github.io/vp-request-spec/#query-by-example
	//	    - https://w3c-ccg.github.io/vp-request-spec/#did-authentication-request
	Query(request *models.RequestEnvelope) *models.ResponseEnvelope

	// adds proof to a Verifiable Credential.
	//
	// https://w3c-ccg.github.io/universal-wallet-interop-spec/#issue
	Issue(request *models.RequestEnvelope) *models.ResponseEnvelope

	// produces a Verifiable Presentation.
	//
	//  https://w3c-ccg.github.io/universal-wallet-interop-spec/#prove
	Prove(request *models.RequestEnvelope) *models.ResponseEnvelope

	// verifies a Verifiable Credential or a Verifiable Presentation.
	//
	// https://w3c-ccg.github.io/universal-wallet-interop-spec/#verify
	Verify(request *models.RequestEnvelope) *models.ResponseEnvelope

	// derives a Verifiable Credential.
	//
	//  https://w3c-ccg.github.io/universal-wallet-interop-spec/#derive
	Derive(request *models.RequestEnvelope) *models.ResponseEnvelope

	// creates a key pair from wallet.
	CreateKeyPair(request *models.RequestEnvelope) *models.ResponseEnvelope
}
