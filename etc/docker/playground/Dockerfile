# Copyright (c) 2020-2025, The OTNS Authors.
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. Neither the name of the copyright holder nor the
#    names of its contributors may be used to endorse or promote products
#    derived from this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

# Builds the OTNS2 playground Docker, for easily trying out OTNS2.
# The .dockerignore file from the repo root is used during build.

# Stage 0: build OTNS, OT nodes, and any dependencies
FROM golang:1.23

ARG GITHUB_REPO="openthread/ot-ns"
ARG GIT_COMMIT="HEAD"

WORKDIR /otns

RUN set -x \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
           git \
           python3 \
           python3-pip \
           sudo \
           unzip \
    \
    && git clone --depth 1 -b main https://github.com/"${GITHUB_REPO}".git . \
    && git fetch origin "${GIT_COMMIT}" \
    && git checkout "${GIT_COMMIT}" \
    && git submodule update --init \
    \
    && ./script/bootstrap \
    && strip /go/bin/grpcwebproxy /go/bin/otns # /go/bin is the default location for "FROM golang:*" Dockers.

# Stage 1: build the final image
FROM ubuntu:24.04

RUN apt-get update && apt-get install -y xdg-utils
# Copy over the produced binaries (including all built node types) from stage 0
COPY --from=0 /go/bin/grpcwebproxy /go/bin/otns /otns/ot-rfsim/ot-versions/ot-* /otns/etc/docker/playground/run-playground.sh /usr/bin/
EXPOSE 8997 8998 8999 9000

ENTRYPOINT [ "run-playground.sh" ]
