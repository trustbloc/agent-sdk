#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Copying aries feature file..."
pwd=$(pwd)
rm -rf ./build/aries-framework-go
git clone -b master https://github.com/hyperledger/aries-framework-go ./build/aries-framework-go
cd ./build/aries-framework-go || exit

git checkout ${ARIES_FRAMEWORK_COMMIT}

sed -i '' -e "1,/AGENT_REST_IMAGE.*/s/AGENT_REST_IMAGE.*/AGENT_REST_IMAGE=docker.pkg.github.com\/trustbloc\/agent-sdk\/agent-sdk-rest/" test/bdd/fixtures/agent-rest/.env
sed -i '' -e "s/aries-agent-rest /agent-rest /" test/bdd/fixtures/agent-rest/docker-compose.yml

make clean generate-test-keys sample-webhook-docker sidetree-cli bdd-test-go

go test -count=1 -v -cover . -p 1 -timeout=20m -race

rm -rf aries-framework-go
cd $pwd || exit