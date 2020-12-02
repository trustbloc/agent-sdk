/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

// GetRequest model
//
// This is used for getting data from the store
//
type GetRequest struct {
	Key string `json:"key"`
}

// GetResponse model
//
// Represents a response of Get command.
//
type GetResponse struct {
	Result []byte `json:"result"`
}

// PutRequest model
//
// This is used for putting data to the store
//
type PutRequest struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

// IteratorRequest model
//
// This is used for getting data from the store.
//
type IteratorRequest struct {
	StartKey string `json:"start_key"`
	EndKey   string `json:"end_key"`
}

// IteratorResponse model
//
// Represents a response of Iterator command.
//
type IteratorResponse struct {
	Results [][]byte `json:"results"`
}

// DeleteRequest model
//
// This is used for deleting data from the store
//
type DeleteRequest struct {
	Key string `json:"key"`
}
