// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-mobile

go 1.15

require (
	github.com/google/uuid v1.1.2
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201029183113-1e234a0af6c6
	github.com/stretchr/testify v1.6.1
	github.com/trustbloc/trustbloc-did-method v0.1.5-0.20201020134433-7a5917ab71d7
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	nhooyr.io/websocket v1.8.3
)

replace (
	github.com/kilic/bls12-381 => github.com/trustbloc/bls12-381 v0.0.0-20201008080608-ba2e87ef05ef
	github.com/phoreproject/bls => github.com/trustbloc/bls v0.0.0-20201023141329-a1e218beb89e
)
