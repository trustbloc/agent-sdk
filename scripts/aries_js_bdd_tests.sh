#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

echo "Copying aries feature file..."
# defines root directory
root=$(pwd)
# defines framework directory
framework_dir=./build/aries-framework-go
# defines working directory
working_dir=$framework_dir/test/aries-js-worker

rm -rf $framework_dir

cd cmd/agent-js-worker
npm install
npm link
cd $root

git clone -b main https://github.com/hyperledger/aries-framework-go $framework_dir
cd $framework_dir || exit 1

git checkout ${ARIES_FRAMEWORK_COMMIT}

if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i '' -e "1,/E2E_AGENT_REST_IMAGE.*/s/E2E_AGENT_REST_IMAGE.*/E2E_AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/aries-js-worker/fixtures/.env
else
  sed -i -e "1,/E2E_AGENT_REST_IMAGE.*/s/E2E_AGENT_REST_IMAGE.*/E2E_AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/aries-js-worker/fixtures/.env
fi

make sidetree-cli

cd test/aries-js-worker || exit 1
npm install
npm link @trustbloc/agent-sdk-web
cd $root || exit 1

cp $working_dir/node_modules/@trustbloc/agent-sdk-web/dist/assets/agent-js-worker.wasm.gz $working_dir/node_modules/@trustbloc/agent-sdk-web/dist/assets/aries-js-worker.wasm.gz
mv $working_dir/node_modules/@trustbloc/agent-sdk-web/dist/rest/agent.js $working_dir/node_modules/@trustbloc/agent-sdk-web/dist/rest/aries.js
mv $working_dir/node_modules/@trustbloc/agent-sdk-web/dist/web/agent.js $working_dir/node_modules/@trustbloc/agent-sdk-web/dist/web/aries.js
mv $working_dir/node_modules/@trustbloc/agent-sdk-web $working_dir/node_modules/@trustbloc/aries-framework-go
mv $working_dir/node_modules/@trustbloc $working_dir/node_modules/@hyperledger

echo "gunzip public/aries-framework-go/assets/agent-js-worker.wasm.gz" >> $working_dir/scripts/setup_assets.sh

if [[ "$OSTYPE" == "darwin"* ]]; then
  sed -i '' -e "s/Aries.Framework/Agent.Framework/" $working_dir/test/common.js
  sed -i '' -e "s/db-namespace/indexed-db-namespace/" $working_dir/test/common.js
  sed -i '' -e "s/\"log-level\": environment.LOG_LEVEL,/\"log-level\": \"debug\",\"storage-type\": \"indexedDB\",\"enableDIDComm\": \"true\",/" $working_dir/test/common.js
  sed -i '' -e "s/15000/20000/" $working_dir/karma.conf.js
else
  sed -i -e "s/Aries.Framework/Agent.Framework/" $working_dir/test/common.js
  sed -i -e "s/db-namespace/indexed-db-namespace/" $working_dir/test/common.js
  sed -i -e "s/\"log-level\": environment.LOG_LEVEL,/\"log-level\": \"debug\",\"storage-type\": \"indexedDB\",\"enableDIDComm\": \"true\",/" $working_dir/test/common.js
  sed -i -e "s/15000/20000/" $working_dir/karma.conf.js
fi

cd $working_dir/fixtures || exit 1
docker-compose down --remove-orphans && docker-compose up -d

cd ..
npm run test || exit 1
cd $root || exit 1

cd $working_dir/fixtures || exit 1
docker-compose stop

cd $root || exit