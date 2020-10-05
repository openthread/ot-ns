# Stage 0: build OpenThread
FROM openthread/environment

WORKDIR /openthread
RUN ./bootstrap
RUN make -f examples/Makefile-simulation OTNS=1
RUN strip /openthread/output/x86_64-unknown-linux-gnu/bin/ot-cli-ftd

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

COPY --from=0 /openthread/output/x86_64-unknown-linux-gnu/bin/ot-cli-ftd /usr/bin/
COPY --from=1 /go/bin/grpcwebproxy /go/bin/otns /usr/bin/

EXPOSE 8997 8998 8999 9000

ENTRYPOINT [ "otns" ]
