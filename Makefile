# Copyright SecureKey Technologies Inc.
#
# SPDX-License-Identifier: Apache-2.0


GOBIN_PATH             = $(abspath .)/build/bin
ARIES_AGENT_REST_PATH=cmd/agent-rest

# Namespace for the agent images
DOCKER_OUTPUT_NS   ?= docker.pkg.github.com
REPO_IMAGE_NAME   ?= trustbloc/agent-sdk
AGENT_NAME ?= agent-rest

ALPINE_VER ?= 3.12
GO_VER ?= 1.15

.PHONY: all
all: clean checks unit-test unit-test-wasm agent-rest agent-rest-docker

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

.PHONY: agent-rest
agent-rest:
	@echo "Building aries-agent-rest"
	@mkdir -p ./build/bin
	@cd ${ARIES_AGENT_REST_PATH} && go build -o ../../build/bin/agent-rest main.go

.PHONY: agent-rest-docker
agent-rest-docker:
	@echo "Building aries agent rest docker image"
	@docker build -f ./images/agent-rest/Dockerfile --no-cache -t $(DOCKER_OUTPUT_NS)/$(REPO_IMAGE_NAME)/${AGENT_NAME}:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .

.PHONY: depend
depend:
	@mkdir -p ./build/bin
	GO111MODULE=off GOBIN=$(GOBIN_PATH) go get github.com/agnivade/wasmbrowsertest

.PHONY: clean
clean: clean-build

.PHONY: clean-build
clean-build:
	@rm -Rf ./build
	@rm -Rf ./cmd/agent-js-worker/node_modules
	@rm -Rf ./cmd/agent-js-worker/dist