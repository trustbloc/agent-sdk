// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0
module github.com/trustbloc/agent-sdk

go 1.15

replace github.com/kilic/bls12-381 => github.com/trustbloc/bls12-381 v0.0.0-20201104214312-31de2a204df8

require (
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.7.4
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201127190801-509fc3277e1e
	github.com/hyperledger/aries-framework-go/component/storage/jsindexeddb v0.0.0-20201127190801-509fc3277e1e
	github.com/igor-pavlenko/httpsignatures-go v0.0.21
	github.com/stretchr/testify v1.6.1
	github.com/trustbloc/edge-core v0.1.5-0.20201121214029-0646e96dbdcf
	github.com/trustbloc/trustbloc-did-method v0.1.5-0.20201113081448-0e789546b4d7
)
