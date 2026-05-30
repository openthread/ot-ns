#!/usr/bin/env python3
"""
OTNS demo: spawn a Matter lightbulb, 10 routers with custom SRP services, and an OTBR.

Usage (inside the Docker container):
    sudo $(which python3) otns-demo.py
"""

import glob
import math
import os
import time
import subprocess
from otns.cli import OTNS

NUM_ROUTERS = 10
CENTER_X = 400
CENTER_Y = 400
RADIUS = 250


def detect_backbone_interface():
    """Detect the first non-loopback interface that is up."""
    try:
        result = subprocess.run(
            ['ip', '-o', 'link', 'show', 'up'],
            capture_output=True, text=True, check=True
        )
        for line in result.stdout.strip().splitlines():
            parts = line.split(': ')
            if len(parts) >= 2:
                iface = parts[1].split('@')[0]
                if iface != 'lo':
                    return iface
    except (subprocess.CalledProcessError, FileNotFoundError):
        pass
    return 'eth0'


print("OTNS Web UI: http://localhost:8997/visualize?addr=localhost:8998")
print("")

backbone_if = detect_backbone_interface()
print(f"Using backbone interface: {backbone_if}")

# Clean up stale chip-tool/Matter state files from previous runs
for pattern in ['/tmp/chip_*', '/tmp/chip-*']:
    for f in glob.glob(pattern):
        os.remove(f)

# Start OTNS in realtime mode with web UI enabled
ns = OTNS(otns_args=[
    '-realtime',
    '-listen', '0.0.0.0:9000',
    '-web=true',
    '-otbr-backbone-if', backbone_if,
])

# Spawn the Matter lightbulb at the center
print("Adding Matter node at center...")
matter_id = ns.add("matter", x=CENTER_X, y=CENTER_Y)
time.sleep(2)

# Spawn router nodes in a circle around the Matter node
for i in range(NUM_ROUTERS):
    angle = 2 * math.pi * i / NUM_ROUTERS
    x = int(CENTER_X + RADIUS * math.cos(angle))
    y = int(CENTER_Y + RADIUS * math.sin(angle))

    print(f"Adding router {i+1}/{NUM_ROUTERS}...")
    router_id = ns.add("router", x=x, y=y)
    time.sleep(2)

    # Register a custom SRP service: "router-{node_id}" _otns-handson._tcp
    ns.node_cmd(router_id, 'srp client autostart enable')
    ns.node_cmd(router_id, f'srp client host name router-{router_id}')
    ns.node_cmd(router_id, 'srp client host address auto')
    ns.node_cmd(router_id, f'srp client service add router-{router_id} _otns-handson._tcp 12345')

# Spawn the OTBR node on the circle (opposite side from router 0)
otbr_angle = math.pi
otbr_x = int(CENTER_X + RADIUS * math.cos(otbr_angle))
otbr_y = int(CENTER_Y + RADIUS * math.sin(otbr_angle))
print("Adding OTBR node...")
otbr_id = ns.add("otbr", x=otbr_x, y=otbr_y)
time.sleep(2)

print(f"\nAll nodes added:")
print(f"  Matter node: {matter_id} (center)")
print(f"  Routers: {NUM_ROUTERS} nodes with _otns-handson._tcp services")
print(f"  OTBR node: {otbr_id}")

# Enable autogo so the simulation runs while we wait
ns.autogo = True

# Give time for the mesh to converge and for the Matter node to publish its SRP service
print("\nWaiting 20s for the mesh to converge...")
time.sleep(20)

print("\n=== Ready ===")
print("Commission the Matter lightbulb by running:")
print("  chip-tool pairing onnetwork 1234 20202021")
print("")
print("Then control it with:")
print("  chip-tool onoff on 1234 1")
print("  chip-tool onoff off 1234 1")
print("")
print("Press Ctrl+C to stop.")

# Keep the script running with interactive CLI
try:
    ns.interactive_cli()
except KeyboardInterrupt:
    print("\nShutting down...")
    ns.close()

