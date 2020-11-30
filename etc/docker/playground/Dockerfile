# Copyright (c) 2020, The OTNS Authors.
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

# Stage 0: build OpenThread
FROM openthread/environment

WORKDIR /openthread
RUN ./bootstrap
RUN make -f examples/Makefile-simulation OTNS=1 BACKBONE_ROUTER=1 DUA=1 MLR=1 THREAD_VERSION=1.2
RUN strip /openthread/output/simulation/bin/ot-cli-ftd

# Stage 1: build OTNS and dependencies
FROM golang:buster

RUN apt-get update
RUN apt-get install -y python3 python3-pip unzip

COPY . /otns
WORKDIR /otns
RUN ./script/install-deps
RUN ./script/install

RUN strip /go/bin/grpcwebproxy /go/bin/otns

# Stage 2: build the final image
FROM debian:buster

COPY --from=0 /openthread/output/simulation/bin/ot-cli-ftd /usr/bin/
COPY --from=1 /go/bin/grpcwebproxy /go/bin/otns /usr/bin/

EXPOSE 8997 8998 8999 9000

ENTRYPOINT [ "otns", "-ot-cli", "ot-cli-ftd" ]
