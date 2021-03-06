// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

replace github.com/trustbloc/agent-sdk => ../..

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210306003819-79bbc2abaa5f
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210305211919-90001416b848
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210305211919-90001416b848
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210303194824-a55a12f8d063
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20210306003819-79bbc2abaa5f
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210306003819-79bbc2abaa5f
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210306003819-79bbc2abaa5f
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/edge-core v0.1.6-0.20210305000733-14a89fe44ae8 // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
)

go 1.15
