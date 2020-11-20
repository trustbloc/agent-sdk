/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package protocol

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	"github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol"
	mocksvc "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/service"
)

// MockProvider mock implementation of provider needed by sdk command controller.
type MockProvider struct {
	*protocol.MockProvider
	ServiceEndpointValue string
	CustomMessenger      service.Messenger
}

// NewMockProvider returns mock implementation of basic provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{&protocol.MockProvider{}, "", nil}
}

// ServiceEndpoint returns the service endpoint.
func (p *MockProvider) ServiceEndpoint() string {
	return p.ServiceEndpointValue
}

// Messenger return mock messenger.
func (p *MockProvider) Messenger() service.Messenger {
	if p.CustomMessenger != nil {
		return p.CustomMessenger
	}

	return &mocksvc.MockMessenger{}
}
