/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package protocol

import "github.com/hyperledger/aries-framework-go/pkg/didcomm/protocol/outofbandv2"

// MockOobServiceV2 is a mock of OobService V2 interface.
type MockOobServiceV2 struct {
	AcceptInvitationHandle func(*outofbandv2.Invitation) (string, error)
	SaveInvitationErr      error
}

// AcceptInvitation mock implementation.
func (m *MockOobServiceV2) AcceptInvitation(arg0 *outofbandv2.Invitation) (string, error) {
	if m.AcceptInvitationHandle != nil {
		return m.AcceptInvitationHandle(arg0)
	}

	return "", nil
}

// SaveInvitation mock implementation.
func (m *MockOobServiceV2) SaveInvitation(arg0 *outofbandv2.Invitation) error {
	if m.SaveInvitationErr != nil {
		return m.SaveInvitationErr
	}

	return nil
}
