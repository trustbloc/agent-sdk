// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

replace github.com/trustbloc/agent-sdk => ../..

go 1.15

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210305233053-d3d22c802c12
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210305233053-d3d22c802c12
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210305233053-d3d22c802c12
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210306141947-8094fee88506
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
)
