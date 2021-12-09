// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

go 1.16

require (
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/cenkalti/backoff/v4 v4.1.1
	github.com/go-kivik/couchdb/v3 v3.2.8 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.8-0.20211209171241-d1472ffcaedc
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210909220549-ce3a2ee13e22
	github.com/hyperledger/aries-framework-go-ext/component/storage/mongodb v0.0.0-20211001130226-2e3422b613aa
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210909220549-ce3a2ee13e22
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.1.4-0.20211115235232-9c7453f469d0
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20211209171241-d1472ffcaedc
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20211209171241-d1472ffcaedc
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20211209171241-d1472ffcaedc
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
)

replace github.com/trustbloc/agent-sdk => ../..
