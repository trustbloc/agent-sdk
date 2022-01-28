# Agent SDK - Build

## Prerequisites
- Go 1.17
- Docker
- Docker-Compose
- Make
- bash
- npm v7
- Configure Docker to use GitHub Packages - [Authenticate](https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages#authenticating-to-github-packages) 
using [GitHub token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line#creating-a-token) 
- Node.js (note: installation via [nvm](https://github.com/nvm-sh/nvm) is *recommended* to avoid errors due to local
  path permissions when running certain `npm` commands (eg. `npm link`). Otherwise, assign the proper permissions to the
  user account running `npm`)

## Targets
```
# run all the project build targets
make all

# run linter checks
make checks

# run unit tests
make unit-test

# run unit tests for wasm
# requires chrome to be installed
make unit-test-wasm

# run bdd tests
make bdd-test
```
