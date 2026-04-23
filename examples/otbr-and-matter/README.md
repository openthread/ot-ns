# OTNS + Matter + OT-BR Docker Environment

This Docker image provides a complete development environment for simulating
Thread networks with OTNS, including Matter device support and an OpenThread
Border Router.

## What's included

- **OTNS** — OpenThread Network Simulator with web UI
- **OT-BR-POSIX** — OpenThread Border Router (with border routing, SRP server, NAT64)
- **Matter all-clusters-app** — A virtual Matter accessory that runs over Thread (via OTNS)
- **ot-rfsim node binaries** — FTD, MTD, BR, RCP, and Matter node types for OTNS

## Prerequisites

- Docker (Docker Desktop on macOS, or Docker Engine on Linux)

## Build (optional)

```bash
./run-docker.sh build
```

Or manually:

```bash
docker build -t otns-matter .
```

The build takes a long time (compiling OpenThread, Matter, OT-BR from source).

## Load an existing image

If you received the image as a TAR archive, you can load it by running:

```
docker load < <path_to_docker_image>
```

On some Docker versions, the image gets tagged with the `docker.io` prefix. if that is the case, remove that prefix by running:

```
docker tag docker.io/framichel/otns-matter framichel/otns-matter
```


## Run

If you haven't built the image in the previous step, it will pull the image from Docker Hub.

```bash
./run-docker.sh run
```

## Using OTNS inside the container

The entrypoint auto-detects the backbone interface and prints the recommended
command. Typically:

```bash
otns -realtime -listen 0.0.0.0:9000 -otbr-backbone-if eth0
```

Or just type `otns` — the entrypoint sets up an alias with the right flags.

Once OTNS is running, open `http://localhost:8997/visualize?addr=localhost:8998` in your browser to access
the OTNS web UI.

### Spawning nodes in OTNS

In the OTNS web UI or CLI, you can add:

- **FTD nodes** — regular Thread Full Thread Devices
- **OTBR nodes** — Thread Border Routers (with border routing and SRP server)
- **Matter nodes** — Matter all-clusters-app devices running over Thread

The OTBR node will automatically:
- Advertise an OMR prefix in Thread network data
- Run an SRP server and advertise it in Thread network data

(Requires `-otbr-backbone-if` to be set to a valid interface.)

## Commissioning a Matter device from outside the container

If you run the container on a Linux machine or VM with `--network host`, you
can commission a simulated Matter device from outside the container using
chip-tool.

### 1. Install avahi (for observation and debugging)

```bash
sudo apt-get update
sudo apt-get install -y avahi-utils
```

Useful commands:

```bash
avahi-browse -rt _meshcop._udp      # Browse Thread Border Routers
avahi-browse -rt _matterc._udp      # Browse Matter commissionable devices
```

### 2. Install chip-tool

The chip-tool is used to commission and control Matter accessories. You can either install it from snap or build it from source.

#### Installing it from snap

```bash
sudo snap install chip-tool
```

#### Building from sources

```bash
# Install build dependencies
sudo apt-get install -y libglib2.0-dev-bin libglib2.0-dev libgirepository1.0-dev libevent-dev

git clone https://github.com/project-chip/connectedhomeip.git
cd connectedhomeip
scripts/checkout_submodules.py --shallow --platform linux
source scripts/activate.sh
./scripts/build/build_examples.py --target linux-x64-chip-tool build
```

The binary will be at `out/linux-x64-chip-tool/chip-tool`.
For arm64 hosts, replace `x64` with `arm64`.

### 3. Commission and control a Matter lightbulb

#### Setup in OTNS

1. Start OTNS inside the container:
   ```
   otns -realtime -listen 0.0.0.0:9000 -otbr-backbone-if eth0
   ```
2. Add a Matter node first: click on "matter" or type `add matter` in the OTNS CLI
3. Add an OTBR node: click on "otbr" or type `add otbr` in the OTNS CLI
4. Add as many other nodes as you want: click on "router" or type `add router`
   in the OTNS CLI
5. Monitor the announced services from the Linux VM (outside the container) using `avahi-browse -a`.
   Wait until a `_matterc._udp` service shows up, meaning that the Matter node has registered its SRP service on the OTBR node. Once that is donem you can start pairing and controlling the accessory.

**Important:** Always add the Matter node before the OTBR node. Add nodes one
at a time and wait for each to fully join before adding the next. All Matter
nodes currently use the same default pairing code, so adding multiple
simultaneously will cause conflicts. Use a different node ID for each Matter
node you commission.

#### Pair the lightbulb

From the Linux host (outside the container), pair using the QR code payload:

```bash
chip-tool pairing onnetwork 1234 20202021
```

The `1234` is the node ID you assign to this device — use a different ID for each
Matter node you commission (e.g., 1234, 1235, 1, 2...).

#### Control the lightbulb

```bash
# Turn on
./chip-tool onoff on 1234 1

# Turn off
./chip-tool onoff off 1234 1

# Read current on/off state
./chip-tool onoff read on-off 1234 1
```

The `1234` is the node ID (assigned during pairing), it must be different for every matter node. The `1` is the endpoint ID. It is always `1` for every node.

#### Viewing Matter node logs

The Matter node writes its output to `/var/log/syslog` inside the container.
To follow logs:

```bash
# Inside the container
sudo cat /var/log/syslog | grep ot-matter
```

## Going further with OTNS automation

OTNS exposes a Python API that lets you script the full lifecycle of a simulated
Thread network — spawning nodes, sending node commands, and driving time — all
from a single Python script.

An example script `otns-automation.py` is provided in this repository. To run it
inside the container:

1. Copy `otns-automation.py` into your running container:
   ```bash
   ./run-docker.sh copy otns-automation.py
   ```
2. Inside the container, activate the OTNS Python virtual environment:
   ```bash
   source ~/ot-ns/.venv-otns/bin/activate
   ```
3. Run the script with root privileges (required to create virtual network interfaces):
   ```bash
   cd ~/ot-ns && sudo $(which python3) otns-automation.py
   ```
4. Open the OTNS web UI in your browser:
   `http://localhost:8997/visualize?addr=localhost:8998`

### What the script does

The script automates the following sequence:

1. **Detects the backbone interface** — finds the first non-loopback interface
   that is up (falls back to `eth0`).
2. **Starts OTNS** in realtime mode with the web UI enabled and the correct
   backbone interface set.
3. **Spawns a Matter lightbulb** node at the center of the canvas.
4. **Spawns 10 router nodes** arranged in a circle around the Matter node. Each
   router enables SRP client autostart, sets a hostname (`router-<id>`), and
   registers a custom SRP service (`_otns-handson._tcp` on port 12345).
5. **Spawns an OTBR node** on the opposite side of the circle. The OTBR
   advertises an OMR prefix and runs an SRP server, allowing the Matter node to
   publish its service and become commissionable from outside the container.

Once the mesh has converged you can commission and control the Matter lightbulb
exactly as described in the section above.
