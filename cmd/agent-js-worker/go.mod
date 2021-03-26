// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-js-worker

go 1.15

require (
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/cenkalti/backoff/v4 v4.1.0 // indirect
	github.com/golang/protobuf v1.5.1 // indirect
	github.com/google/tink/go v1.5.0
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.7-0.20210325221830-6ab3160b7588
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210325134531-84a30b2ecacb
	github.com/hyperledger/aries-framework-go/component/storage/edv v0.0.0-20210326171010-c7ce51b1d6cb
	github.com/hyperledger/aries-framework-go/component/storage/indexeddb v0.0.0-20210310140909-2ae2d7df101e
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210325221830-6ab3160b7588
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210326171010-c7ce51b1d6cb
	github.com/mitchellh/mapstructure v1.3.3
	github.com/stretchr/testify v1.7.0
	github.com/teserakt-io/golang-ed25519 v0.0.0-20210104091850-3888c087a4c8 // indirect
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/edge-core v0.1.7-0.20210324113338-e0047bbbfdff
	github.com/trustbloc/kms v0.1.7-0.20210323140543-8c8c56dac24b
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2 // indirect
	golang.org/x/sys v0.0.0-20210324051608-47abb6519492 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/trustbloc/agent-sdk => ../..
