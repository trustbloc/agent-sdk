// Copyright SecureKey Technologies Inc. All Rights Reserved.
//
// SPDX-License-Identifier: Apache-2.0

module github.com/trustbloc/agent-sdk/cmd/agent-rest

replace github.com/trustbloc/agent-sdk => ../../

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/golang/snappy v0.0.2 // indirect
	github.com/gorilla/mux v1.8.0
	github.com/hyperledger/aries-framework-go v0.1.6-0.20210224230531-58e1368e5661
	github.com/hyperledger/aries-framework-go-ext/component/storage/couchdb v0.0.0-20201113155502-c4ba5d2c7c0a
	github.com/hyperledger/aries-framework-go-ext/component/storage/mysql v0.0.0-20201113155502-c4ba5d2c7c0a
	github.com/hyperledger/aries-framework-go-ext/component/vdr/trustbloc v0.0.0-20210125133828-10c25f5d6d37
	github.com/hyperledger/aries-framework-go/component/storage/leveldb v0.0.0-20201208011347-35f4bc183acf
	github.com/hyperledger/aries-framework-go/spi v0.0.0-20210208194322-89066ddae325
	github.com/rs/cors v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/trustbloc/agent-sdk v0.0.0-00010101000000-000000000000
)

go 1.15
