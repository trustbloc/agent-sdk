/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package protocol

import "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol"

// MockProvider mock implementation of provider needed by sdk command controller.
type MockProvider struct {
	*protocol.MockProvider
	ServiceEndpointValue string
}

// NewMockProvider returns mock implementation of basic provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{&protocol.MockProvider{}, ""}
}

// ServiceEndpoint returns the service endpoint.
func (p *MockProvider) ServiceEndpoint() string {
	return p.ServiceEndpointValue
}
