/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

const path = require('path');
const TerserPlugin = require("terser-webpack-plugin");
const isSnapshot = require('./package.json').dependencies.hasOwnProperty("@trustbloc-cicd/agent-sdk-web")
const isSnapshotDev = require('./package.json').devDependencies.hasOwnProperty("@trustbloc-cicd/agent-sdk-web")
const agent_sdk = (isSnapshot || isSnapshotDev) ? "@trustbloc-cicd/agent-sdk-web" : "@trustbloc/agent-sdk-web"

module.exports = {
    entry: {
        'wallet-sdk':'./src/index.js',
    },
    output: {
        path: path.resolve(__dirname, 'dist'),
        filename: '[name].min.js',
        libraryTarget: 'umd',
        library: {
            name: 'walletSDK',
            type: 'umd',
        },
        clean: true,
    },
    optimization: {
        minimize: true,
        minimizer: [
            new TerserPlugin({
                extractComments: false,
            }),
        ],
    },
    resolve: {
        alias: {
            "@trustbloc/agent-sdk-web": path.resolve(__dirname, 'node_modules/' + agent_sdk)
        }
    }
};

