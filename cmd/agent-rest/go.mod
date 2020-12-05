// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

replace (
	github.com/kilic/bls12-381 => github.com/trustbloc/bls12-381 v0.0.0-20201104214312-31de2a204df8
	github.com/trustbloc/agent-sdk => ../../
)

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.5-0.20201205211909-fdef7da31ced
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20201113155502-c4ba5d2c7c0a
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20201113155502-c4ba5d2c7c0a
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20201205211909-fdef7da31ced
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
	github.com/trustbloc/trustbloc-did-method v0.1.5-0.20201203214019-c56f43ad3f6e
	golang.org/x/net v0.0.0-20201027133719-8eef5233e2a1 // indirect
)

go 1.15
