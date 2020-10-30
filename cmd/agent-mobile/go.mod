// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.15

require (
	github.com/google/go-cmp v0.5.1 // indirect
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201022202135-f8f69217453b
	github.com/multiformats/go-multihash v0.0.14 // indirect
	github.com/stretchr/testify v1.6.1
	golang.org/x/mobile v0.0.0-20200801112145-973feb4309de // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	nhooyr.io/websocket v1.8.3
)

replace (
	github.com/kilic/bls12-381 => github.com/trustbloc/bls12-381 v0.0.0-20201008080608-ba2e87ef05ef
	github.com/phoreproject/bls => github.com/trustbloc/bls v0.0.0-20201023141329-a1e218beb89e
)
