#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Copying aries feature file..."
pwd=$(pwd)
rm -rf ./build/aries-framework-go
git clone -b main https://github.com/hyperledger/aries-framework-go ./build/aries-framework-go
cd ./build/aries-framework-go || exit

git checkout ${ARIES_FRAMEWORK_COMMIT}

if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i '' -e "1,/AGENT_REST_IMAGE.*/s/AGENT_REST_IMAGE.*/AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/bdd/fixtures/agent-rest/.env
  sed -i '' -e "s/aries-agent-rest /agent-rest /" test/bdd/fixtures/agent-rest/docker-compose.yml
else
  sed -i -e "1,/AGENT_REST_IMAGE.*/s/AGENT_REST_IMAGE.*/AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/bdd/fixtures/agent-rest/.env
  sed -i -e "s/aries-agent-rest /agent-rest /" test/bdd/fixtures/agent-rest/docker-compose.yml
fi

make clean generate-test-keys sample-webhook-docker sidetree-cli

cd test/bdd
go test -count=1 -v -cover . -p 1 -timeout=20m -race -run controller

cd $pwd || exit