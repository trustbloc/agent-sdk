/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

var webpackConfig = require('./webpack.dev.js')

module.exports = function (config) {
    config.set({
        frameworks: ['mocha', 'chai', 'webpack'],
        files: [
            {pattern: "public/agent-js-worker/assets/*", included: false},
            {pattern: "test/**/*.ini", included: true},
            {pattern: "test/testdata/*", included: true},
            {pattern: "test/specs/**/*.spec.js", type: "module"},
        ],
        preprocessors: {
            'test/specs/**/*.spec.js': ['webpack', 'sourcemap'],
            'test/**/*.ini': ['ini2js'],
            'test/fixtures/testdata/*': ['file-fixtures']
        },
        webpack: webpackConfig,
        reporters: ['spec'],
        browsers: ['ChromeHeadless_cors'],
        browserNoActivityTimeout: 60000,
        browserDisconnectTimeout: 60000,
        customLaunchers: {
            ChromeHeadless_cors: {
                base: "ChromeHeadless",
                flags: ["--disable-web-security", "--allow-running-insecure-content", "--ignore-certificate-errors",
                    "--ignore-certificate-errors-spki-list", "--ignore-urlfetcher-cert-requests"]
            },
            Chrome_without_security: {
                base: 'Chrome',
                flags: ['--disable-web-security', '--disable-site-isolation-trials', '--auto-open-devtools-for-tabs']
            }
        },
        client: {
            captureConsole: false,
            mocha: {
                timeout: 30000
            }
        }
    })
}


