#!/bin/bash
#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#
set -e

echo "Running $0"

pwd=`pwd`
touch "$pwd"/coverage.out

amend_coverage_file () {
if [ -f profile.out ]; then
     cat profile.out >> "$pwd"/coverage.out
     rm profile.out
fi
}

# Running agent-sdk unit tests
PKGS=`go list github.com/trustbloc/agent-sdk/... 2> /dev/null`
go test $PKGS -count=1 -race -coverprofile=profile.out -covermode=atomic -timeout=10m
amend_coverage_file

# Running agent-mobile unit tests
cd cmd/agent-mobile
PKGS=`go list github.com/trustbloc/cmd/agent-mobile/... 2> /dev/null`
go test $PKGS -count=1 -race -coverprofile=profile.out -covermode=atomic -timeout=10m
amend_coverage_file
cd "$pwd"

# Running agent rest unit tests
cd cmd/agent-rest
PKGS=`go list github.com/trustbloc/agent-sdk/cmd/agent-rest/... 2> /dev/null`
go test $PKGS -count=1 -race -coverprofile=profile.out -covermode=atomic -timeout=10m
amend_coverage_file
cd "$pwd" || exit