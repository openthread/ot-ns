# Stage 0: build OpenThread
FROM openthread/environment

WORKDIR /openthread
RUN ./bootstrap
RUN make -f examples/Makefile-simulation OTNS=1

# Stage 1: build OTNS and dependencies
FROM golang:buster

RUN apt-get update
RUN apt-get install -y python3 python3-pip unzip

COPY . /otns
WORKDIR /otns
RUN ./script/install-deps
RUN ./script/install

# Stage 2: build the final image
FROM debian:buster

COPY --from=0 /openthread/output/x86_64-unknown-linux-gnu/bin/ot-cli-ftd /usr/bin/
COPY --from=1 /go/bin/grpcwebproxy /usr/bin/
COPY --from=1 /go/bin/otns /usr/bin/

ENTRYPOINT [ "otns" ]