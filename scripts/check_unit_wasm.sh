#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running $0"

# Running wasm unit test
# TODO Support collecting code coverage

PATH="$GOBIN:$PATH" GOOS=js GOARCH=wasm go test "github.com/trustbloc/agent-sdk/pkg/storage/jsindexeddbcache" -count=1 -exec=wasmbrowsertest -cover -timeout=10m


# run unit test for agent-js-worker
cd cmd/agent-js-worker
PATH="$GOBIN:$PATH" GOOS=js GOARCH=wasm go test "github.com/trustbloc/agent-sdk/cmd/agent-js-worker" -count=1 -exec=wasmbrowsertest -cover -timeout=10m
