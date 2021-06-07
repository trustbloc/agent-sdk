// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

go 1.16

require (
	github.com/bluele/gcache v0.0.2 // indirect
	github.com/btcsuite/btcd v0.21.0-beta // indirect
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210607112508-8ef35c358338
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20210430213153-6c349de21198
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20210430213153-6c349de21198
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.1.1
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20210507165908-d8529097d7a0
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210520055214-ae429bb89bf7
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210520055214-ae429bb89bf7
	github.com/piprate/json-gold v0.4.0
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	github.com/teserakt-io/golang-ed25519 v0.0.0-20210104091850-3888c087a4c8 // indirect
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
)

replace github.com/trustbloc/agent-sdk => ../..
