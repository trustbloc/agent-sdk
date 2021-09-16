// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.16

require (
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.8-0.20210916154931-0196c3a2d102
	github.com/hyperledger/aries-framework-go-ext/component/vdr/orb v0.1.3
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210916154931-0196c3a2d102
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210916154931-0196c3a2d102
	github.com/hyperledger/aries-framework-go/test/component v0.0.0-20210916154931-0196c3a2d102
	github.com/piprate/json-gold v0.4.1-0.20210813112359-33b90c4ca86c
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	nhooyr.io/websocket v1.8.3
)

replace github.com/trustbloc/agent-sdk => ../../
