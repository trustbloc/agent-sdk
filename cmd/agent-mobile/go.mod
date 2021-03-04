// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.15

require (
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210304035610-cb1ce0da0c3b
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210125133828-10c25f5d6d37
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210226001736-7c70b7efda83
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210302153503-0e00e248f14d
	github.com/hyperledger/aries-framework-go/test/component v0.0.0-20210226235232-298aa129d822
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	nhooyr.io/websocket v1.8.3
)

replace github.com/trustbloc/agent-sdk => ../../
