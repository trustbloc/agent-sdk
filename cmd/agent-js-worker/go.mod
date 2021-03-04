// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210304035610-cb1ce0da0c3b
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210303194824-a55a12f8d063
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210302153503-0e00e248f14d
	github.com/mitchellh/mapstructure v1.4.1
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0
	github.com/trustbloc/edge-core v0.1.6-0.20210224175343-275d0e0370c4
	github.com/trustbloc/kms v0.1.6-0.20210302134939-3933de07039b
)

replace github.com/trustbloc/agent-sdk => ../../
