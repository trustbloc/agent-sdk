// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.16

require (
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210709155504-c8d5f0b43bb5
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.1.1
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210709155504-c8d5f0b43bb5
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210709155504-c8d5f0b43bb5
	github.com/hyperledger/aries-framework-go/test/component v0.0.0-20210709155504-c8d5f0b43bb5
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	nhooyr.io/websocket v1.8.3
)

replace github.com/trustbloc/agent-sdk => ../../
