// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210226001736-7c70b7efda83
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210125133828-10c25f5d6d37
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210226001736-7c70b7efda83
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210226001736-7c70b7efda83
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210226001736-7c70b7efda83
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210226001736-7c70b7efda83
	github.com/mitchellh/mapstructure v1.4.1
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0
	github.com/trustbloc/edge-core v0.1.6-0.20210224175343-275d0e0370c4
	github.com/trustbloc/kms v0.1.6-0.20210226144927-6c67cc12839f
)

replace github.com/trustbloc/agent-sdk => ../../
