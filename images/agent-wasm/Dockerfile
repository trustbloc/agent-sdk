#
# Copyright SecureKey Technologies Inc. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

ARG GO_VER
ARG ALPINE_VER

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	git \
	libtool \
	make;
ADD . /opt/workspace/agent-sdk
WORKDIR /opt/workspace/agent-sdk
ENV EXECUTABLES go git

FROM golang as golang_build
ARG GO_TAGS
RUN GO_TAGS=${GO_TAGS} make http-server

FROM alpine:${ALPINE_VER} as base
COPY --from=golang_build /opt/workspace/agent-sdk/build/bin/http-server /usr/local/bin
ENTRYPOINT ["http-server"]
