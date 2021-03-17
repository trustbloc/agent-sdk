// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210317132839-dd5660169bae
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210306194409-6e4c5d622fbc
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210306194409-6e4c5d622fbc
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210306194409-6e4c5d622fbc
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20210310140909-2ae2d7df101e
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210310140909-2ae2d7df101e
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210310140909-2ae2d7df101e
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/agent-sdk => ../..
