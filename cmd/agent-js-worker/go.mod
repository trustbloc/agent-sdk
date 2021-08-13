// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.16

require (
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/google/tink/go v1.6.1-0.20210519071714-58be99b3c4d0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210811135743-532e65035d3b
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.0.0-20210813123233-e22ddceee0b1
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210807121559-b41545a4f1e8
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210709155504-c8d5f0b43bb5
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210807121559-b41545a4f1e8
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210807121559-b41545a4f1e8
	github.com/mitchellh/mapstructure v1.3.3
	github.com/piprate/json-gold v0.4.0
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/edge-core v0.1.7-0.20210812092729-6c61997fa9dd
	github.com/trustbloc/kms v0.1.7-0.20210813141531-93ef7a32fb42
)

replace github.com/trustbloc/agent-sdk => ../..
