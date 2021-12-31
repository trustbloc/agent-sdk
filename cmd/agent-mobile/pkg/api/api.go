/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

// Package api describes agents api.
package api

// AriesController provides Aries agent protocols tailored to mobile platforms.
type AriesController interface {

	// GetIntroduceController returns an implementation of IntroduceController
	GetIntroduceController() (IntroduceController, error)

	// GetVerifiableController returns an implementation of VerifiableController
	GetVerifiableController() (VerifiableController, error)

	// GetIssueCredentialController returns an implementation of IssueCredentialController
	GetIssueCredentialController() (IssueCredentialController, error)

	// GetPresentProofController returns an implementation of PresentProofController
	GetPresentProofController() (PresentProofController, error)

	// GetDIDClient returns an implementation of DIDClient
	GetDIDClient() (DIDClient, error)

	// GetDIDExchangeController returns an implementation of DIDExchangeController
	GetDIDExchangeController() (DIDExchangeController, error)

	// GetVDRController returns an implementation of VDRController
	GetVDRController() (VDRController, error)

	// GetMediatorController returns an implementation of MediatorController
	GetMediatorController() (MediatorController, error)

	// GetMediatorClient returns an implementation of MediatorClient
	GetMediatorClientController() (MediatorClient, error)

	// GetMessagingController returns an implementation of MessagingController
	GetMessagingController() (MessagingController, error)

	// GetOutOfBandController returns an implementation of OutOfBandController
	GetOutOfBandController() (OutOfBandController, error)

	// GetOutOfBandV2Controller returns an implementation of OutOfBandController
	GetOutOfBandV2Controller() (OutOfBandV2Controller, error)

	// GetKMSController returns an implementation of KMSController
	GetKMSController() (KMSController, error)

	// GetBlindedRoutingController returns an implementation of BlindedRoutingController
	GetBlindedRoutingController() (BlindedRoutingController, error)

	// RegisterHandler registers handler for handling notifications
	RegisterHandler(h Handler, topics string) string

	// UnregisterHandler unregisters handler
	UnregisterHandler(id string)

	// GetVCWalletController returns an implementation of VCWalletController
	GetVCWalletController() (VCWalletController, error)

	// GetLDController returns an implementation of LDController
	GetLDController() (LDController, error)
}
