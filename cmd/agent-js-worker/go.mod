// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210305161732-29081cf55c77
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210303194824-a55a12f8d063
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210303162231-46716728d6eb
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210303162231-46716728d6eb
	github.com/mitchellh/mapstructure v1.4.1
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0
	github.com/trustbloc/edge-core v0.1.6-0.20210305000733-14a89fe44ae8
	github.com/trustbloc/kms v0.1.6-0.20210304191421-0ebf2bf45b54
)

replace github.com/trustbloc/agent-sdk => ../../
