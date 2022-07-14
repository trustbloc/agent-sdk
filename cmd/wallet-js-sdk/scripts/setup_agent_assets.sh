#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

rm -rf public
mkdir -p public/agent-js-worker/assets

pwd=$(pwd)
echo "currently running in $pwd"
ls
echo $(grep "@trustbloc-cicd/agent-sdk-web" "package.json")

if [[ $(grep "@trustbloc-cicd/agent-sdk-web" "package.json") ]] ; then
  echo "finding asset in trustbloc-cicd/"
  cp -Rp node_modules/@trustbloc-cicd/agent-sdk-web/dist/assets/* public/agent-js-worker/assets
  cp -Rp node_modules/@trustbloc-cicd/agent-lite-sdk-web/dist/assets/agent-lite-js-worker.wasm.gz public/agent-js-worker/assets
else
  echo "finding asset in trustbloc/"
  cp -Rp node_modules/@trustbloc/agent-sdk-web/dist/assets/* public/agent-js-worker/assets
  cp -Rp node_modules/@trustbloc/agent-lite-sdk-web/dist/assets/agent-lite-js-worker.wasm.gz public/agent-js-worker/assets
fi

gunzip public/agent-js-worker/assets/agent-js-worker.wasm.gz
gunzip public/agent-js-worker/assets/agent-lite-js-worker.wasm.gz
