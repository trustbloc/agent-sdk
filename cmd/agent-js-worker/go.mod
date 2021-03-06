// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

replace github.com/trustbloc/agent-sdk => ../..

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210305233053-d3d22c802c12
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210306141947-8094fee88506
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210306141947-8094fee88506
	github.com/mitchellh/mapstructure v1.3.3
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/edge-core v0.1.6-0.20210306154041-63c6b31a177c
	github.com/trustbloc/kms v0.1.6-0.20210306112542-4ebac0ebd739
)
