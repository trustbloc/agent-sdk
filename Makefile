# Copyright SecureKey Technologies Inc.
#
# SPDX-License-Identifier: Apache-2.0


GOBIN_PATH             = $(abspath .)/build/bin
HTTP_SERVER_PATH       = cmd/http-server

# GO version
ALPINE_VER ?= 3.12
GO_TAGS    ?=
GO_VER     ?= 1.15.2

# Namespace for the images
DOCKER_OUTPUT_NS         ?= docker.pkg.github.com
REPO_IMAGE_NAME          ?= trustbloc/agent-sdk

.PHONY: all
all: clean checks unit-test unit-test-wasm

.PHONY: checks
checks: license lint

.PHONY: license
license:
	@scripts/check_license.sh

.PHONY: lint
lint:
	@scripts/check_lint.sh

.PHONY: unit-test
unit-test:
	@scripts/check_unit.sh

.PHONY: unit-test-wasm
unit-test-wasm: export GOBIN=$(GOBIN_PATH)
unit-test-wasm: depend
	@scripts/check_unit_wasm.sh

.PHONY: depend
depend:
	@mkdir -p ./build/bin
	GO111MODULE=off GOBIN=$(GOBIN_PATH) go get github.com/agnivade/wasmbrowsertest

.PHONY: http-server
http-server:
	@echo "Building http-server"
	@cd ${HTTP_SERVER_PATH} && go build -o ../../build/bin/http-server main.go

.PHONY: agent-sdk-wasm-docker
agent-sdk-wasm-docker: clean http-server
	@echo "Building agent wasm docker image"
	@docker build -f ./images/agent-wasm/Dockerfile --no-cache -t $(DOCKER_OUTPUT_NS)/$(REPO_IMAGE_NAME)/agent-sdk-wasm:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) .

.PHONY: clean
clean: clean-build

.PHONY: clean-build
clean-build:
	@rm -Rf ./build
	@rm -Rf ./cmd/agent-js-worker/node_modules
	@rm -Rf ./cmd/agent-js-worker/dist