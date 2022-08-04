/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

const path = require("path")

const { CleanWebpackPlugin } = require('clean-webpack-plugin')
const WebpackShellPlugin = require('webpack-shell-plugin')

const { PATHS } = require("./webpack.config.common.js")

const OUTPUT = path.join(PATHS.buildDir, "web/")

module.exports = {
    entry: path.join(PATHS.srcDir, "agent.js"),
    target: 'web',
    output: {
        path: OUTPUT,
        publicPath: "dist/web",
        libraryTarget: "umd",
        filename: 'agent.js',
        library: 'Agent'
    },
    plugins: [
        new CleanWebpackPlugin(),
        new WebpackShellPlugin({
            onBuildStart: [
                "mkdir -p " + OUTPUT,
                "cp -p src/worker-impl-web.js dist/assets",
            ]
        })
    ],
    resolve: {
        alias: {
            worker_loader$: path.join(PATHS.srcDir, "worker-loader-web.js")
        }
    }
}
