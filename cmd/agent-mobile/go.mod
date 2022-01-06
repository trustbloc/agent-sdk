// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.17

require (
	github.com/google/uuid v1.3.0
	github.com/hyperledger/aries-framework-go v0.1.8-0.20220106195936-a9d6794663ed
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.1.4-0.20211229160742-30796de12ba6
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20220106195936-a9d6794663ed
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20220106195936-a9d6794663ed
	github.com/hyperledger/aries-framework-go/test/component v0.0.0-20220106195936-a9d6794663ed
	github.com/piprate/json-gold v0.4.1-0.20210813112359-33b90c4ca86c
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	nhooyr.io/websocket v1.8.3
)

require (
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20211214153431-5c8a10d6e6ad // indirect
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324 // indirect
)

replace github.com/trustbloc/agent-sdk => ../../
