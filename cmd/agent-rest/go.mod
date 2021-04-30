// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

go 1.16

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210430213153-6c349de21198
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210430213153-6c349de21198
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.0.0-20210430213153-6c349de21198
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210429205242-c5e97865879c
	github.com/piprate/json-gold v0.4.0
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
)

replace github.com/trustbloc/agent-sdk => ../..
