// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.16

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.0.0-20210430213153-6c349de21198
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210429205242-c5e97865879c
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210429205242-c5e97865879c
	github.com/mitchellh/mapstructure v1.3.3
	github.com/piprate/json-gold v0.4.0
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/edge-core v0.1.7-0.20210429222332-96b987820e63
	github.com/trustbloc/kms v0.1.7-0.20210430171137-77ee09acc581
)

replace github.com/trustbloc/agent-sdk => ../..
