# Copyright SecureKey Technologies Inc.
#
# SPDX-License-Identifier: Apache-2.0


GOBIN_PATH             = $(abspath .)/build/bin
ARIES_AGENT_REST_PATH=cmd/agent-rest
ARIES_AGENT_MOBILE_PATH=cmd/agent-mobile
ARIES_FRAMEWORK_COMMIT=bb5bedb39f3610a362829bb0b79b8ad2aca71b72
PROJECT_ROOT = github.com/trustbloc/agent-sdk
OPENAPI_SPEC_PATH=build/rest/openapi/spec
OPENAPI_DOCKER_IMG=quay.io/goswagger/swagger
OPENAPI_DOCKER_IMG_VERSION=v0.23.0

# Namespace for the agent images
DOCKER_OUTPUT_NS   ?= ghcr.io
REPO_IMAGE_NAME   ?= trustbloc
DOCKER_AGENT_NAME ?= agent-sdk-server

ALPINE_VER ?= 3.15
GO_VER ?= 1.17

.PHONY: all
all: clean checks unit-test unit-test-wasm agent-rest agent-server-docker bdd-test

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

.PHONY: bdd-test
bdd-test: aries-js-bdd rest-api-bdd

.PHONY: unit-test-wasm
unit-test-wasm: export GOBIN=$(GOBIN_PATH)
unit-test-wasm: depend
	@scripts/check_unit_wasm.sh

.PHONY: agent-rest
agent-rest:
	@echo "Building aries-agent-rest"
	@mkdir -p ./build/bin
	@cd ${ARIES_AGENT_REST_PATH} && go build -o ../../build/bin/agent-rest main.go

.PHONY: agent-server-docker
agent-server-docker:
	@echo "Building aries agent rest docker image"
	@docker build -f ./images/agent-rest/Dockerfile --no-cache -t $(DOCKER_OUTPUT_NS)/$(REPO_IMAGE_NAME)/${DOCKER_AGENT_NAME}:latest \
	--build-arg GO_VER=$(GO_VER) \
	--build-arg ALPINE_VER=$(ALPINE_VER) \
	--build-arg GO_TAGS=$(GO_TAGS) \
	--build-arg GOPROXY=$(GOPROXY) .

.PHONY: rest-api-bdd
rest-api-bdd: clean agent-server-docker
	@ARIES_FRAMEWORK_COMMIT=$(ARIES_FRAMEWORK_COMMIT) scripts/aries_bdd_tests.sh

.PHONY: aries-js-bdd
aries-js-bdd: clean agent-server-docker
	@ARIES_FRAMEWORK_COMMIT=$(ARIES_FRAMEWORK_COMMIT) scripts/aries_js_bdd_tests.sh

.PHONY: wallet-sdk-tests
wallet-sdk-tests:
	@set -e
	@cd cmd/agent-js-worker && npm install
	@cd cmd/wallet-js-sdk && npm install && npm run test

.PHONY: unit-test-mobile
unit-test-mobile:
	@echo "Running unit tests for mobile"
	@cd ${ARIES_AGENT_MOBILE_PATH} && $(MAKE) unit-test

.PHONY: agent-mobile
agent-mobile:
	@echo "Building agent-mobile"
	@cd ${ARIES_AGENT_MOBILE_PATH} && $(MAKE) all

.PHONY: depend
depend:
	@mkdir -p ./build/bin
	GO111MODULE=off GOBIN=$(GOBIN_PATH) go get github.com/agnivade/wasmbrowsertest

.PHONY: generate-openapi-spec
generate-openapi-spec: clean
	@echo "Generating and validating controller API specifications using Open API"
	@mkdir -p build/rest/openapi/spec
	@SPEC_LOC=${OPENAPI_SPEC_PATH}  \
	DOCKER_IMAGE=$(OPENAPI_DOCKER_IMG) DOCKER_IMAGE_VERSION=$(OPENAPI_DOCKER_IMG_VERSION)  \
	scripts/generate-openapi-spec.sh

.PHONY: generate-openapi-demo-specs
generate-openapi-demo-specs: clean generate-openapi-spec agent-server-docker
	@echo "Generate demo agent rest controller API specifications using Open API"
	@SPEC_PATH=${OPENAPI_SPEC_PATH} OPENAPI_DEMO_PATH=deployments/demo/openapi \
    	DOCKER_IMAGE=$(OPENAPI_DOCKER_IMG) DOCKER_IMAGE_VERSION=$(OPENAPI_DOCKER_IMG_VERSION)  \
    	scripts/generate-openapi-demo-specs.sh

generate-test-keys: clean
	@mkdir -p -p deployments/keys/tls
	@docker run -i --rm \
		-v $(abspath .):/opt/go/src/$(PROJECT_ROOT) \
		--entrypoint "/opt/go/src/$(PROJECT_ROOT)/scripts/generate_test_keys.sh" \
		frapsoft/openssl

.PHONY: run-openapi-demo
run-openapi-demo: generate-test-keys generate-openapi-demo-specs agent-server-docker
	@echo "Starting demo agent rest containers ..."
	@DEMO_COMPOSE_PATH=deployments/demo/openapi SIDETREE_COMPOSE_PATH=deployments/sidetree-mock AGENT_REST_COMPOSE_PATH=deployments/agent-rest  \
        scripts/run-openapi-demo.sh

.PHONY: stop-openapi-demo
stop-openapi-demo:
	@echo "Stopping demo agent rest containers ..."
	@DEMO_COMPOSE_PATH=deployments/demo/openapi SIDETREE_COMPOSE_PATH=deployments/sidetree-mock AGENT_REST_COMPOSE_PATH=deployments/agent-rest  \
        DEMO_COMPOSE_OP=down scripts/run-openapi-demo.sh

.PHONY: clean
clean:
	@rm -Rf ./build
	@rm -Rf $(ARIES_AGENT_MOBILE_PATH)/build
	@rm -Rf ./cmd/agent-js-worker/node_modules
	@rm -Rf ./cmd/agent-js-worker/dist
	@rm -Rf ./deployments/keys/tls
	@rm -Rf ./deployments/demo/openapi/specs
