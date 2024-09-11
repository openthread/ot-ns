# Stress Tests

OpenThread Network Simulator (OTNS) is used to run stress tests to improve the robustness of OpenThread.

## Test Suite: Network Forming

### Variable Sized Network Forming

Test Purpose --- Make sure different number of nodes can form a single partition within a reasonable time limit.

- Topology
  - NxN Routers, for n in 1 to 7.
- Procedure
  1. Boot up NxN Routers.
  1. Wait for them to form a single partition.
- Fault Injections
  - None
- Pass Criteria
  - Network forming time is less than the time limit.

### Large Size Network Forming

Test Purpose: Ensure that a large number of nodes can form a single partition within a reasonable time limit.

- Topology
  - 8x8 Routers.
- Procedure
  1. Boot up 8x8 Routers.
  1. Wait for them to form a single partition.
- Fault Injections
  - None.
- Pass Criteria
  - Network forming time is less than the time limit.

## Test Suite: Connectivity

### ML-EID Connectivity

Test Purpose: Make sure Nodes have connectivity to the Border Router's ML-EID under a harsh environment.

- Topology
  - 20 Rout`ers (one BR)
  - 20 FEDs
  - 10 MEDs
  - 10 SEDs`
- Procedure
  1. Boot up Nodes.
  1. Ping BR's ML-EID from other Nodes and measure the connectivity.
- Fault Injections
  - Nodes are constantly moving.
  - Nodes constantly fail for a short period.
  - Packet loss radio set to 20%.
- Pass Criteria
  - Max Delay (of all nodes) <= 3600s

### ALOC Connectivity

Test Purpose: Make sure Nodes have connectivity to the Border Router's ALOC under a harsh environment.

- Topology
  - 20 Routers
  - 20 FEDs
  - 10 MEDs
  - 10 SEDs
- Procedure
  1. Boot up Nodes.
  1. Ping BR's Service ALOC from other Nodes and measure the connectivity.
- Fault Injections
  - Nodes are constantly moving.
  - Nodes constantly fail for a short period.
  - Packet loss radio set to 20%.
- Pass Criteria
  - Max Delay (of all nodes) <= 3600s

## Test Suite: Latency

### Ping Latency

Test Purpose: Measure the ping delay for different hops and different data sizes.

- Topology
  - 6 Routers forming Circle 1
  - 6 Routers forming Circle 2
- Procedure
  1. Boot up Nodes.
  1. Wait for them to form a single partition.
  1. Ping Nodes within 1 hop.
  1. Ping Nodes within 2 hops.
  1. Ping Nodes without 3 hops.
- Fault Injections
  - None
- Pass Criteria
  - Max ping latency < 100ms

## Test Suite: Commissioning

### Commissioning

Test Purpose: Make sure Joiners can join a Thread network by Commissioning.

- Topology
  - 5x5 Nodes, with Commissioner at the center.
- Procedure
  1. Boot up Nodes.
  1. Start the Commissioner.
  1. Start Joiners (2 Joiners joins concurrently).
  1. Wait for all Joiners to join successfully.
- Fault Injections
  - Packet loss ratio set to 20%
- Pass Criteria
  - All Joiners joined successfully
  - Join process success rate >= 90%
  - Average Join duration <= 20s

## Test Suite: Multicast

### CoAP Multicast

Test Purpose: Make sure CoAP multicast messgaes can reach all nodes with reasonable coverage and delay.

- Topology
  - 8 Routers (one BR)
  - 8 FEDs
  - 8 MEDs
  - 8 SEDs
- Procedure
  1. Boot up all Nodes.
  1. BR sends CoAP multicast to all Nodes.
  1. Measure the coverage and delay.
- Fault Injections
  - Packet loss ratio set to 50%
- Pass Criteria
  - Average coverage >= 70%
  - Average delay < 200ms for non-SEDs
  - Average delay < 2000ms for SEDs

## Test Suite: OTNS

### OTNS Performance

Test Purpose: Make sure OTNS can run large-scale simulations efficiently.

- Topology
  - 4x8 Routers
- Procedure
  1. Boot up 4x8 Routers.
  1. Wait for them to form a single partition.
  1. Simulate for 1 hour.
  1. Measure the program execution time.
- Fault Injections
  - None.
- Pass Criteria
  - Execution time <= 30s
