#!/bin/bash
# Build or run the OTNS + Matter + OT-BR Docker image.
# Works on both macOS and Linux hosts.
#
# Usage:
#   ./run-docker.sh build          Build the Docker image
#   ./run-docker.sh run            Run an interactive container
#   ./run-docker.sh build-run      Build then run

set -e

IMAGE_NAME="framichel/otns-matter"
CONTAINER_NAME="otns-matter-dev"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

# Detect host architecture and map to Docker platform
detect_platform() {
    local arch
    arch="$(uname -m)"
    case "${arch}" in
        x86_64|amd64)   echo "linux/amd64" ;;
        arm64|aarch64)   echo "linux/arm64" ;;
        *)
            echo "Unsupported architecture: ${arch}" >&2
            exit 1
            ;;
    esac
}

do_build() {
    local platform
    platform="$(detect_platform)"
    echo "Building ${IMAGE_NAME} for platform ${platform} ..."
    docker build \
        --platform "${platform}" \
        -t "${IMAGE_NAME}" \
        "${SCRIPT_DIR}"
}

do_run() {
    local platform
    platform="$(detect_platform)"
    echo "Starting ${CONTAINER_NAME} (platform ${platform}) ..."

    local run_args=(
        -it --rm
        --platform "${platform}"
        --name "${CONTAINER_NAME}"
        --cap-add NET_ADMIN
        --device /dev/net/tun:/dev/net/tun
    )

    if [ "$(uname)" = "Linux" ]; then
        # On Linux, use host networking for direct access to host interfaces
        run_args+=(--network host)
    else
        # On macOS, use port forwarding and enable IPv6
        run_args+=(
            --sysctl net.ipv6.conf.all.disable_ipv6=0
            -p 8997:8997
            -p 8998:8998
            -p 8999:8999
            -p 9000:9000
            -p 8080:8080
        )
    fi

    docker run "${run_args[@]}" "${IMAGE_NAME}"
}

case "${1}" in
    build)
        do_build
        ;;
    run)
        do_run
        ;;
    build-run)
        do_build
        do_run
        ;;
    *)
        echo "Usage: $0 build|run|build-run" >&2
        exit 1
        ;;
esac
