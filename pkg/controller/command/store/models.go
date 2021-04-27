/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package store

import "github.com/hyperledger/aries-framework-go/spi/storage"

// PutRequest model
//
// This is used for putting data in the store.
//
type PutRequest struct {
	Key   string        `json:"key"`
	Value []byte        `json:"value"`
	Tags  []storage.Tag `json:"tags"`
}

// GetRequest model
//
// This is used for getting data (value or tags) from the store.
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

// QueryRequest model
//
// This is used for getting data (values only, without tags) from the store.
//
type QueryRequest struct {
	Expression string `json:"expression"`
	PageSize   int    `json:"pageSize"`
}

// QueryResponse model
//
// Represents a response of Query command.
//
type QueryResponse struct {
	Results [][]byte `json:"results"`
}

// DeleteRequest model
//
// This is used for deleting data from the store.
//
type DeleteRequest struct {
	Key string `json:"key"`
}
