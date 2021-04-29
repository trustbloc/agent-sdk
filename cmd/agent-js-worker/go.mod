// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.16

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210429013345-a595aa0b19c4
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.0.0-20210422102350-1c5d6f027647
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210422133815-2ef2d99cb692
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210310140909-2ae2d7df101e
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210427144858-06fb8b7d2d30
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210422133815-2ef2d99cb692
	github.com/mitchellh/mapstructure v1.3.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/edge-core v0.1.7-0.20210429084532-c385778b4d8b
	github.com/trustbloc/kms v0.1.7-0.20210428154127-bc15ec954519
)

replace github.com/trustbloc/agent-sdk => ../..
