#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -e

echo "Running $0"

GOLANGCI_LINT_VER=v1.31.0
DOCKER_CMD=${DOCKER_CMD:-docker}

if [ ! $(command -v ${DOCKER_CMD}) ]; then
    exit 0
fi

echo "Linting top-level module..."
${DOCKER_CMD} run --rm -v $(pwd):/opt/workspace -w /opt/workspace golangci/golangci-lint:$GOLANGCI_LINT_VER golangci-lint run
echo "Linting agent-rest module..."
${DOCKER_CMD} run --rm -v $(pwd):/opt/workspace -w /opt/workspace/cmd/agent-rest golangci/golangci-lint:$GOLANGCI_LINT_VER golangci-lint run -c ../../.golangci.yml
echo "Linting agent-js-worker module..."
${DOCKER_CMD} run --rm -e GOOS=js -e GOARCH=wasm -v $(pwd):/opt/workspace -w /opt/workspace/cmd/agent-js-worker golangci/golangci-lint:$GOLANGCI_LINT_VER golangci-lint run -c ../../.golangci.yml
