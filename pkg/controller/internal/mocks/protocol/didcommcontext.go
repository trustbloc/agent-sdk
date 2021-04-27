/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package protocol

// MockDIDCommContext mock aries-framework-go/pkg/didcomm/common/service.DIDCommContext.
type MockDIDCommContext struct {
	MyDIDValue    string
	TheirDIDValue string
	Properties    map[string]interface{}
}

// MyDID return my DID.
func (c *MockDIDCommContext) MyDID() string {
	return c.MyDIDValue
}

// TheirDID return their DID.
func (c *MockDIDCommContext) TheirDID() string {
	return c.TheirDIDValue
}

// All return all properties.
func (c *MockDIDCommContext) All() map[string]interface{} {
	return c.Properties
}
