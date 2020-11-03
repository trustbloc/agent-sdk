// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
module github.com/trustbloc/agent-sdk

go 1.15

replace (
	github.com/kilic/bls12-381 => github.com/trustbloc/bls12-381 v0.0.0-20201008080608-ba2e87ef05ef
	github.com/phoreproject/bls => github.com/trustbloc/bls v0.0.0-20201023141329-a1e218beb89e
)

require (
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.4
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201029183113-1e234a0af6c6
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/trustbloc/edge-core v0.1.5-0.20200916124536-c32454a16108
	github.com/trustbloc/edv v0.1.4
	github.com/trustbloc/trustbloc-did-method v0.1.5-0.20201030202533-dc1622875c56
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)
