/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package protocol provides useful mocks for testing different command controller features.
package protocol

import (
	"github.com/hyperledger/aries-framework-go/pkg/didcomm/common/service"
	didexchangesvc "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/didexchange"
	mockdidexchange "github.com/hyperledger/aries-framework-go/pkg/mock/didcomm/protocol/didexchange"
)

// MockDIDExchangeSvc is extension of aries mock did-exchange service.
type MockDIDExchangeSvc struct {
	*mockdidexchange.MockDIDExchangeSvc
	ConnID string
	State  string
}

// RegisterMsgEvent register message event.
func (m *MockDIDExchangeSvc) RegisterMsgEvent(ch chan<- service.StateMsg) error {
	if m.State == "" {
		m.State = didexchangesvc.StateIDCompleted
	}

	go func() {
		ch <- service.StateMsg{
			ProtocolName: didexchangesvc.DIDExchange,
			Type:         service.PostState,
			StateID:      m.State,
			Properties: &mockdidexchange.MockEventProperties{
				ConnID: m.ConnID,
			},
		}
	}()

	return nil
}
