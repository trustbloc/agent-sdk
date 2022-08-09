#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

GOOS=js GOARCH=wasm go build -o src/agent-js-worker.wasm main.go
GOOS=js GOARCH=wasm go build -o src/agent-vcwallet-js-worker.wasm vcwallet/main.go

gzip -f src/agent-js-worker.wasm
gzip -f src/agent-vcwallet-js-worker.wasm
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" src/

rm -rf dist/assets
mkdir -p dist/assets
cp -p src/agent-js-worker.wasm.gz dist/assets
cp -p src/agent-vcwallet-js-worker.wasm.gz dist/assets
cp -p src/wasm_exec.js dist/assets
