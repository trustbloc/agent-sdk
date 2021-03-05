// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.15

require (
	github.com/google/uuid v1.2.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210305213256-1b30fec59b51
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210303194824-a55a12f8d063
	github.com/hyperledger/aries-framework-go/component/storageutil v0.0.0-20210305213256-1b30fec59b51
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210305213256-1b30fec59b51
	github.com/hyperledger/aries-framework-go/test/component v0.0.0-20210305213256-1b30fec59b51
	github.com/stretchr/testify v1.7.0
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	nhooyr.io/websocket v1.8.3
)

replace github.com/trustbloc/agent-sdk => ../../
