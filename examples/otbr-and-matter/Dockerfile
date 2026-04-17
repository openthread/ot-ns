# Dockerfile for OTNS + OT-BR + Matter all-clusters-app development environment
#
# Works on both arm64 and x86_64/amd64 hosts 

FROM ubuntu:24.04

ARG TARGETARCH
ARG GO_VERSION=1.26.2
ARG OTNS_BRANCH=pr-br
ARG OTNS_REPO=https://github.com/EskoDijk/ot-ns
ARG MATTER_BRANCH=openthread
ARG MATTER_REPO=https://github.com/gmarcosb/connectedhomeip

ENV DEBIAN_FRONTEND=noninteractive
ENV HOME=/home/otns
ENV GOPATH=${HOME}/go
ENV PATH=${GOPATH}/bin:/usr/local/go/bin:${PATH}

SHELL ["/bin/bash", "-c"]

# ---------------------------------------------------------------
# Stage 1: System packages
# ---------------------------------------------------------------
RUN apt-get clean && apt-get update && apt-get install -y \
        sudo \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user (after sudo is installed so /etc/sudoers exists cleanly)
RUN useradd -m -s /bin/bash -d ${HOME} otns && \
    echo "otns ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

RUN apt-get update && apt-get install -y \
        git \
        build-essential \
        g++ \
        cmake \
        ninja-build \
        patch \
        wget \
        curl \
        unzip \
        xdg-utils \
        python3 \
        python3-pip \
        python3-venv \
        pkg-config \
        libssl-dev \
        libdbus-1-dev \
        dbus \
        libavahi-client-dev \
        libgirepository1.0-dev \
        libglib2.0-dev \
        libreadline-dev \
        libncurses-dev \
        libprotobuf-dev \
        protobuf-compiler \
        libjsoncpp-dev \
        libgtest-dev \
        libgmock-dev \
        libnetfilter-queue1 \
        libnetfilter-queue-dev \
        iptables \
        ipset \
        bind9 \
        rsyslog \
        nodejs \
        npm \
        lsb-release \
        vim \
        nano \
        iproute2 \
        iputils-ping \
        net-tools \
        ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# ---------------------------------------------------------------
# Stage 2: Install Go (architecture-aware)
# ---------------------------------------------------------------
RUN wget -q https://go.dev/dl/go${GO_VERSION}.linux-${TARGETARCH}.tar.gz -O /tmp/go.tar.gz && \
    tar -C /usr/local -xzf /tmp/go.tar.gz && \
    rm /tmp/go.tar.gz

# ---------------------------------------------------------------
# Stage 2.5: Install grpcwebproxy for arm64 if we're running on
#            arm64 architecrure
# ---------------------------------------------------------------
# The OTNS install-deps script downloads an x86_64 grpcwebproxy which
# won't work on arm64. Pre-install the arm64 build so install-deps
# finds it already present and skips its own download.
RUN mkdir -p ${GOPATH}/bin && \
    if [ "${TARGETARCH}" = "arm64" ]; then \
        wget -q https://github.com/improbable-eng/grpc-web/releases/download/v0.14.0/grpcwebproxy-v0.14.0-arm64.zip \
             -O /tmp/grpcwebproxy.zip && \
        cd /tmp && unzip grpcwebproxy.zip && \
        mv dist/grpcwebproxy* ${GOPATH}/bin/grpcwebproxy && \
        chmod +x ${GOPATH}/bin/grpcwebproxy && \
        rm -rf /tmp/grpcwebproxy.zip /tmp/dist; \
    fi
RUN chown -R otns:otns ${GOPATH}

# ---------------------------------------------------------------
# Stage 3: Clone and build OTNS (Esko's branch with BR support)
# ---------------------------------------------------------------
USER otns
WORKDIR ${HOME}

RUN git clone --branch ${OTNS_BRANCH} ${OTNS_REPO} ot-ns

WORKDIR ${HOME}/ot-ns
RUN git submodule update --init --recursive

# Patch 1: Matter nodes may respond with "down" or "up" to ifconfig, like OTBR
RUN sed -i 's/if node.cfg.IsRcp && node.cfg.IsBorderRouter {/if node.cfg.IsRcp \&\& node.cfg.IsBorderRouter || node.cfg.Type == MATTER {/' simulation/node.go

# Patch 2: Drain Matter app stdout (CHIP logs) to prevent pipe blocking and write to node log
RUN sed -i '/go node.lineReaderStdOut(node.pipeOut)/a\\t\t} else if cfg.Type == MATTER {\n\t\t\tgo node.lineReaderStdErr(node.pipeOut) \/\/ drain Matter stdout (CHIP logs) and write to node log' simulation/node.go

# Install OTNS dependencies and OTNS itself
RUN ./script/install-deps && ./script/install

# Make otns and grpcwebproxy available system-wide (needed for sudo)
RUN sudo ln -sf ${GOPATH}/bin/otns /usr/local/bin/otns && \
    sudo ln -sf ${GOPATH}/bin/grpcwebproxy /usr/local/bin/grpcwebproxy

# ---------------------------------------------------------------
# Stage 4: Clone and build Matter connectedhomeip (OTNS-enabled)
# ---------------------------------------------------------------
WORKDIR ${HOME}

RUN git clone ${MATTER_REPO} connectedhomeip && \
    cd connectedhomeip && \
    git checkout ${MATTER_BRANCH}

WORKDIR ${HOME}/connectedhomeip

# Add ot-ns as a submodule pointing to Esko's branch, then check it out
RUN mkdir -p ./third_party/ot-ns && \
    git submodule add -b ${OTNS_BRANCH} ${OTNS_REPO} third_party/ot-ns/repo && \
    git submodule update --init third_party/ot-ns/repo

# Checkout submodules (includes openthread, etc.)
# Enable long paths to avoid failures with deeply nested submodules
RUN git config --global core.longpaths true && \
    git submodule update --init && \
    python3 scripts/checkout_submodules.py --shallow --platform linux

# Apply ot-rfsim build fixes needed for the Matter accessory
RUN cd third_party/ot-ns/repo && git apply - <<'PATCH'
diff --git a/ot-rfsim/src/ble.c b/ot-rfsim/src/ble.c
index 2a7f87f..7fdf05b 100644
--- a/ot-rfsim/src/ble.c
+++ b/ot-rfsim/src/ble.c
@@ -25,7 +25,7 @@
  *  ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
  *  POSSIBILITY OF SUCH DAMAGE.
  */
-
+#include "platform-rfsim.h"
 #if OPENTHREAD_CONFIG_BLE_TCAT_ENABLE
 
 #include "platform-rfsim.h"
diff --git a/ot-rfsim/src/openthread-core-rfsim-config.h b/ot-rfsim/src/openthread-core-rfsim-config.h
index acb7cf4..293029c 100644
--- a/ot-rfsim/src/openthread-core-rfsim-config.h
+++ b/ot-rfsim/src/openthread-core-rfsim-config.h
@@ -348,4 +348,19 @@
 #define OPENTHREAD_SIMULATION_VIRTUAL_TIME 1
 #endif
 
+#ifndef OPENTHREAD_CONFIG_ECDSA_ENABLE
+#define OPENTHREAD_CONFIG_ECDSA_ENABLE 1
+#endif
+
+#ifndef OPENTHREAD_CONFIG_BLE_TCAT_ENABLE
+#define OPENTHREAD_CONFIG_BLE_TCAT_ENABLE 0
+#endif
+
+#ifndef OPENTHREAD_CONFIG_BORDER_ROUTING_ENABLE
+#define OPENTHREAD_CONFIG_BORDER_ROUTING_ENABLE 0
+#endif
+
+#ifndef OPENTHREAD_SIMULATION_VIRTUAL_TIME_UART
+#define OPENTHREAD_SIMULATION_VIRTUAL_TIME_UART 1
+#endif
 #endif // OPENTHREAD_CORE_RFSIM_CONFIG_H_
diff --git a/ot-rfsim/src/platform-rfsim.c b/ot-rfsim/src/platform-rfsim.c
index b98bac8..9059292 100644
--- a/ot-rfsim/src/platform-rfsim.c
+++ b/ot-rfsim/src/platform-rfsim.c
@@ -55,8 +55,9 @@
 
 #define VERIFY_EVENT_SIZE(X) OT_ASSERT((payloadLen >= sizeof(X)) && "received event payload too small");
 
+#if OPENTHREAD_CONFIG_BORDER_ROUTING_ENABLE && OPENTHREAD_CONFIG_UDP_FORWARD_ENABLE
 static const otIp6Address kUnspecifiedIp6Address = {.mFields = {.m32 = {0, 0, 0, 0}}};
-
+#endif
 extern int gSockFd;
 
 uint64_t     gLastMsgId = 0;
@@ -271,13 +272,13 @@ void handleUdpForwarding(otMessage    *aMessage,
     size_t                    msgLen = otMessageGetLength(aMessage);
 
     OT_ASSERT(msgLen <= sizeof(buf));
-
+
     evData.mSrcPort = aSockPort;
     evData.mDstPort = aPeerPort;
     memcpy(evData.mSrcIp6, &kUnspecifiedIp6Address, OT_IP6_ADDRESS_SIZE);
     memcpy(evData.mDstIp6, aPeerAddr, OT_IP6_ADDRESS_SIZE);
     otMessageRead(aMessage, 0, buf, msgLen);
-
+    (void) kUnspecifiedIp6Address;
     otSimSendMsgToHostEvent(OT_SIM_EVENT_UDP_TO_HOST, &evData, &buf[0], msgLen);
 }
 
diff --git a/ot-rfsim/src/platform-rfsim.cpp b/ot-rfsim/src/platform-rfsim.cpp
index a88c6d4..579f417 100644
--- a/ot-rfsim/src/platform-rfsim.cpp
+++ b/ot-rfsim/src/platform-rfsim.cpp
@@ -32,6 +32,7 @@
  *   This file includes the C++ portions of the OT-RFSIM platform.
  */
 
+#include "platform-rfsim.h"
 #if OPENTHREAD_CONFIG_BORDER_ROUTING_ENABLE
 
 #include "net/ip6.hpp"
diff --git a/ot-rfsim/src/uart.c b/ot-rfsim/src/uart.c
index 8ef034a..d5589ae 100644
--- a/ot-rfsim/src/uart.c
+++ b/ot-rfsim/src/uart.c
@@ -37,7 +37,6 @@
 #include "event-sim.h"
 #include "utils/code_utils.h"
 #include "utils/uart.h"
-
 #if OPENTHREAD_SIMULATION_VIRTUAL_TIME_UART == 0
 
 #include <assert.h>
PATCH

# Apply Matter build fixes using sed (inline git patches break in Docker heredocs):

# 1) Move otAppCliInit before Matter ThreadStack init so Matter's SRP callback
#    is not overwritten by the CLI SRP client module
RUN git apply - <<'PATCH'
diff --git a/examples/platform/linux/AppMain.cpp b/examples/platform/linux/AppMain.cpp
--- a/examples/platform/linux/AppMain.cpp
+++ b/examples/platform/linux/AppMain.cpp
@@ -753,7 +753,9 @@
         otSysInit(static_cast<int>(args.size()), args.data());
+#if CHIP_ENABLE_OTNS
+        // Init CLI before Matter ThreadStack so Matter's SRP callback
+        // is not overwritten by the CLI SRP client module.
+        otAppCliInit(otInstanceGetSingle());
+#endif
         SuccessOrExit(err = DeviceLayer::ThreadStackMgrImpl().InitThreadStack());
         SuccessOrExit(err = DeviceLayer::ThreadStackMgrImpl().StartThreadTask());
-#if CHIP_ENABLE_OTNS
-        otAppCliInit(otInstanceGetSingle());
-#endif
         ChipLogProgress(NotSpecified, "Thread initialized.");
PATCH

# 2) Enable OTNS status push, virtual-time UART define, and verbose CHIP logging
RUN sed -i "/chip_enable_otns=true/a\            self.extra_gn_options.append('openthread_config_otns_enable=true')" scripts/build/builders/host.py && \
    sed -i "/openthread_config_otns_enable=true/a\            self.extra_gn_options.append('chip_detail_logging=true')" scripts/build/builders/host.py && \
    sed -i "/chip_detail_logging=true/a\            self.extra_gn_options.append('chip_progress_logging=true')" scripts/build/builders/host.py && \
    sed -i "/chip_progress_logging=true/a\            self.extra_gn_options.append('chip_automation_logging=true')" scripts/build/builders/host.py

# 3) Add OPENTHREAD_SIMULATION_VIRTUAL_TIME_UART=1 define to OTNS platform's BUILD.gn
RUN sed -i '/include_dirs = \[ "\${chip_root}\/third_party\/ot-ns\/repo\/ot-rfsim\/src" \]/a\  defines = [ "OPENTHREAD_SIMULATION_VIRTUAL_TIME_UART=1" ]' third_party/openthread/platforms/otns/BUILD.gn

# Bootstrap the Matter build environment and build the all-clusters app
# The build target adapts to the host architecture
RUN source ./scripts/activate.sh && \
    if [ "${TARGETARCH}" = "arm64" ]; then \
        ARCH="arm64"; \
    else \
        ARCH="x64"; \
    fi && \
    ./scripts/build/build_examples.py \
        --target linux-${ARCH}-all-clusters-no-wifi-no-ble-openthread-endpoint-otns \
        build

# Build ot-rfsim node binaries (FTD, MTD, BR, RCP)
WORKDIR ${HOME}/ot-ns/ot-rfsim
RUN ./script/build_latest
RUN ./script/build_br
RUN ./script/build_rcp

# Create convenience symlinks expected by OTNS
WORKDIR ${HOME}/ot-ns
RUN ln -sf ./ot-rfsim/ot-versions/ot-cli-ftd ./ && \
    ln -sf ./ot-rfsim/ot-versions/ot-cli-ftd_br ./

# Symlink the Matter all-clusters binary as ot-matter-node so OTNS can find it
RUN MATTER_BIN=$(ls ${HOME}/connectedhomeip/out/linux-*-all-clusters-no-wifi-no-ble-openthread-endpoint-otns/chip-all-clusters-app 2>/dev/null | head -1) && \
    if [ -n "$MATTER_BIN" ]; then \
        ln -sf "$MATTER_BIN" ./ot-rfsim/ot-versions/ot-matter-node; \
    else \
        echo "WARNING: Matter binary not found, ot-matter-node symlink not created"; \
    fi

# ---------------------------------------------------------------
# Stage 5: OT-BR-POSIX (build during image creation)
# ---------------------------------------------------------------
# OT-BR-POSIX is configured with:
#   - Web GUI for browser-based management
#   - Border Routing for OMR prefix advertisement
#   - SRP server for Thread service registration in network data
#   - NAT64 for IPv4 connectivity from Thread devices
#   - mDNSResponder instead of avahi for DNS-SD
WORKDIR ${HOME}

RUN git clone https://github.com/openthread/ot-br-posix && \
    cd ot-br-posix && \
    git submodule update --init --recursive

# Bootstrap installs apt packages AND builds/installs mDNSResponder from source
USER root

# Bootstrap installs apt packages AND builds/installs mDNSResponder from source
RUN HOME=/root && cd /home/otns/ot-br-posix && \
    INFRA_IF_NAME=eth0 WEB_GUI=1 NAT64=1 OTBR_MDNS=mDNSResponder ./script/bootstrap

# Stub systemctl so setup doesn't try to start/enable services during docker build
RUN ln -sf /bin/true /usr/local/bin/systemctl

# Run setup to install binaries, DBus config, ot-ctl, etc.
RUN HOME=/root && cd /home/otns/ot-br-posix && \
    INFRA_IF_NAME=eth0 WEB_GUI=1 NAT64=1 OTBR_MDNS=mDNSResponder DOCKER=1 FIREWALL=0 ./script/setup

# Restore real systemctl for runtime use
RUN rm -f /usr/local/bin/systemctl

# Rebuild OT-BR-POSIX using the OTNS-specific build script from ot-rfsim.
# This overwrites the otbr-agent binary with one that has the correct flags
# (OTNS support, no backbone router multicast routing which fails on macOS Docker).
RUN cd /home/otns/ot-ns/ot-rfsim && \
    yes | ./script/build_otbr \
        -DOTBR_BACKBONE_ROUTER=OFF \
        -DOT_BACKBONE_ROUTER_MULTICAST_ROUTING=OFF

USER otns

# ---------------------------------------------------------------
# Stage 6: Entrypoint with OT-BR services
# ---------------------------------------------------------------
WORKDIR ${HOME}

# Create a startup script that launches supporting services.
# OTNS manages the border router (ot-rfsim BR) itself - we only
# start the infrastructure services it depends on.
COPY --chown=otns:otns <<'ENTRYPOINT_SCRIPT' /home/otns/entrypoint.sh
#!/bin/bash

# Detect the main non-loopback interface for OTBR backbone
BACKBONE_IF=$(ip -o link show up | awk -F': ' '!/lo/{print $2; exit}')
BACKBONE_IF="${BACKBONE_IF:-eth0}"

echo "=== OTNS + Matter + OT-BR Development Environment ==="
echo ""
echo "Detected backbone interface: ${BACKBONE_IF}"
echo ""
echo "Once the OTBR node is started, start the OT-BR web interface:"
echo "  otbr-web -p 8080"
echo ""
echo "Matter all-clusters-app:"
ARCH_DIR=$(ls -d ~/connectedhomeip/out/linux-*-all-clusters-* 2>/dev/null | head -1)
if [ -n "$ARCH_DIR" ]; then
    echo "  ${ARCH_DIR}/chip-all-clusters-app"
else
    echo "  (not found - build may have failed)"
fi
echo ""
echo "OTNS node binaries:"
echo "  ~/ot-ns/ot-rfsim/ot-versions/"
echo ""
echo "Start OTNS with border routing (OMR prefix + SRP server):"
echo "  sudo otns -realtime -listen 0.0.0.0:9000 -otbr-backbone-if ${BACKBONE_IF}"
echo ""

# Alias so that bare 'otns' uses the right flags
alias otns="sudo otns -realtime -listen 0.0.0.0:9000 -otbr-backbone-if ${BACKBONE_IF}"

# Start D-Bus (needed by OT-BR services)
sudo mkdir -p /run/dbus
sudo dbus-daemon --system --nopidfile || true

# Start rsyslog so Matter syslog logging works
sudo rsyslogd 2>/dev/null || true

# Start mDNSResponder (needed by otbr-agent for DNS-SD)
sudo mdnsd > /tmp/mdnsd.log 2>&1 || echo "ERROR: mdnsd failed to start, see /tmp/mdnsd.log"

exec "$@"
ENTRYPOINT_SCRIPT
RUN chmod +x ${HOME}/entrypoint.sh

# Expose OTNS web ports and OT-BR web UI
EXPOSE 8997 8998 8999 9000 8080

ENTRYPOINT ["/home/otns/entrypoint.sh"]
WORKDIR ${HOME}/ot-ns/ot-rfsim/ot-versions
CMD ["/bin/bash"]
