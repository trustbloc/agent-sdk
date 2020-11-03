// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

replace (
	github.com/kilic/bls12-381 => github.com/trustbloc/bls12-381 v0.0.0-20201008080608-ba2e87ef05ef
	github.com/phoreproject/bls => github.com/trustbloc/bls v0.0.0-20201023141329-a1e218beb89e
	github.com/trustbloc/agent-sdk => ../../
)

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/gorilla/mux v1.7.4
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201029183113-1e234a0af6c6
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20201030114218-27cdc521d9fc
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20201030114218-27cdc521d9fc
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20201029183113-1e234a0af6c6
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/trustbloc-did-method v0.1.5-0.20201030202533-dc1622875c56
	golang.org/x/net v0.0.0-20201027133719-8eef5233e2a1 // indirect
)

go 1.15
