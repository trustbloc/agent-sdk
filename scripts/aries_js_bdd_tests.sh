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

function version { echo "$@" | awk -F. '{ printf("%d%03d%03d%03d\n", $1,$2,$3,$4); }'; }

git clone -b main https://github.com/hyperledger/aries-framework-go $framework_dir
cd $framework_dir || exit 1

git checkout ${ARIES_FRAMEWORK_COMMIT}

if [[ "$OSTYPE" == "darwin"* ]]; then
  mOSVer=$(sw_vers -productVersion)
  # MaxOS Big Sur executes sed just like GNU sed while previous versions require ''
  if [ $(version $mOSVer) -lt $(version "11") ]; then
    sed -i '' -e "1,/E2E_AGENT_REST_IMAGE.*/s/E2E_AGENT_REST_IMAGE.*/E2E_AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/aries-js-worker/fixtures/.env
  else
    sed -i -e "1,/E2E_AGENT_REST_IMAGE.*/s/E2E_AGENT_REST_IMAGE.*/E2E_AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/aries-js-worker/fixtures/.env
  fi
  sed -i '' -e "1,/E2E_AGENT_REST_IMAGE.*/s/E2E_AGENT_REST_IMAGE.*/E2E_AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/aries-js-worker/fixtures/.env
else
  sed -i -e "1,/E2E_AGENT_REST_IMAGE.*/s/E2E_AGENT_REST_IMAGE.*/E2E_AGENT_REST_IMAGE=ghcr.io\/trustbloc\/agent-sdk-server/" test/aries-js-worker/fixtures/.env
fi

make sidetree-cli

cd test/aries-js-worker || exit 1
npm install
npm link @trustbloc/agent-sdk
cd $root || exit 1

cp $working_dir/node_modules/@trustbloc/agent-sdk/dist/assets/agent-js-worker.wasm.gz $working_dir/node_modules/@trustbloc/agent-sdk/dist/assets/aries-js-worker.wasm.gz
mv $working_dir/node_modules/@trustbloc/agent-sdk/dist/rest/agent.js $working_dir/node_modules/@trustbloc/agent-sdk/dist/rest/aries.js
mv $working_dir/node_modules/@trustbloc/agent-sdk/dist/web/agent.js $working_dir/node_modules/@trustbloc/agent-sdk/dist/web/aries.js
mv $working_dir/node_modules/@trustbloc/agent-sdk $working_dir/node_modules/@trustbloc/aries-framework-go
mv $working_dir/node_modules/@trustbloc $working_dir/node_modules/@hyperledger

echo "gunzip public/aries-framework-go/assets/agent-js-worker.wasm.gz" >> $working_dir/scripts/setup_assets.sh

if [[ "$OSTYPE" == "darwin"* ]]; then
  if [ $(version $mOSVer) -lt $(version "11") ]; then
    sed -i '' -e "s/Aries.Framework/Agent.Framework/" $working_dir/test/common.js
    sed -i '' -e "s/db-namespace/indexedDB-namespace/" $working_dir/test/common.js
    sed -i '' -e "s/\"log-level\": \"debug\",/\"log-level\": \"debug\",\"storageType\": \"indexedDB\",/" $working_dir/test/common.js
    sed -i '' -e "s/15000/20000/" $working_dir/karma.conf.js
  else
    sed -i -e "s/Aries.Framework/Agent.Framework/" $working_dir/test/common.js
    sed -i -e "s/db-namespace/indexedDB-namespace/" $working_dir/test/common.js
    sed -i -e "s/\"log-level\": \"debug\",/\"log-level\": \"debug\",\"storageType\": \"indexedDB\",/" $working_dir/test/common.js
    sed -i -e "s/15000/20000/" $working_dir/karma.conf.js
  fi
else
  sed -i -e "s/Aries.Framework/Agent.Framework/" $working_dir/test/common.js
  sed -i -e "s/db-namespace/indexedDB-namespace/" $working_dir/test/common.js
  sed -i -e "s/\"log-level\": \"debug\",/\"log-level\": \"debug\",\"storageType\": \"indexedDB\",/" $working_dir/test/common.js
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