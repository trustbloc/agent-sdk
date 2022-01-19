// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

go 1.17

require (
	github.com/cenkalti/backoff/v4 v4.1.2
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.8-0.20220119221523-d82510cb10e8
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210909220549-ce3a2ee13e22
	github.com/hyperledger/aries-framework-go-ext/component/storage/mongodb v0.0.0-20211219215001-23cd75276fdc
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210909220549-ce3a2ee13e22
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.1.4-0.20220114172935-0e96d787f80f
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20220106195936-a9d6794663ed
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20220106195936-a9d6794663ed
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20220106195936-a9d6794663ed
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
)

require (
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/go-kivik/couchdb/v3 v3.2.8 // indirect
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20211214153431-5c8a10d6e6ad // indirect
)

replace github.com/trustbloc/agent-sdk => ../..
