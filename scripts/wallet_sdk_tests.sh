#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running wallet SDK tests..."
root=$(pwd)

cd ${root}/cmd//agent-js-worker
npm install

cd ${root}/cmd/wallet-js-sdk
npm install
npm run test
