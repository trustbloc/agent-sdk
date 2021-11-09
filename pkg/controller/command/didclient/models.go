/*
Copyright SecureKey Technologies Inc. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package didclient

// CreateOrbDIDRequest model
//
// This is used for creating orb DID.
//
type CreateOrbDIDRequest struct {
	PublicKeys []PublicKey `json:"publicKeys,omitempty"`
}

// ResolveOrbDIDRequest model
//
// This is used for resolving orb DID.
//
type ResolveOrbDIDRequest struct {
	DID string `json:"did,omitempty"`
}

// CreatePeerDIDRequest model
//
// This is used for creating peer DID.
//
type CreatePeerDIDRequest struct {
	RouterConnectionID string `json:"routerConnectionID,omitempty"`
}

// PublicKey public key.
type PublicKey struct {
	ID       string   `json:"id,omitempty"`
	Type     string   `json:"type,omitempty"`
	Encoding string   `json:"encoding,omitempty"`
	KeyType  string   `json:"keyType,omitempty"`
	Purposes []string `json:"purposes,omitempty"`
	Recovery bool     `json:"recovery,omitempty"`
	Update   bool     `json:"update,omitempty"`
	Value    string   `json:"value,omitempty"`
}
