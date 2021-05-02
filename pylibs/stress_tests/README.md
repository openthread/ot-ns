# Stress Tests

OpenThread Network Simulator (OTNS) is used to run stress tests to improve the robustness of OpenThread.

## Network Forming

### Variable Sized Network Forming

#### Test Purpose

Make sure different number of nodes can form a single partition within reasonable time limits.

#### Topology

- NxN Routers, for n in 1 to 7.

#### Procedure

- Boot up NxN Routers.
- Wait them to form a single partition.

#### Fault Injections

None.

#### Pass Criteria

- Network forming time is less than corresponding time limits.

### Large Size Network Forming

#### Test Purpose

Make sure large number of nodes can form a single partition within reasonable time limits.

#### Topology

- 8x8 Routers.

#### Procedure

- Boot up 8x8 Routers.
- Wait them to form a single partition.

#### Fault Injections

None.

#### Pass Criteria

- Network forming time is less than a time limit.

## Connectivity

### ML-EID Connectivity

#### Test Purpose

Make sure Nodes have connectivity to the Border Router's ML-EID under harsh environment.

#### Topology

- 20 Routers (one BR)
- 20 FEDs
- 10 MEDs
- 10 SEDs

#### Procedure

- Boot up Nodes.
- Ping BR's ML-EID from other Nodes and measure the connectivity.

#### Fault Injections

- Nodes are constantly moving.
- Nodes constantly fail for a short period.
- Packet loss radio set to 20%.

#### Pass Criteria

- Max Delay (of all nodes) <= 3600s

### ALOC Connectivity

#### Test Purpose

Make sure Nodes have connectivity to the Border Router's ALOC under harsh environment.

#### Topology

- 20 Routers
- 20 FEDs
- 10 MEDs
- 10 SEDs

#### Procedure

- Boot up Nodes.
- Ping BR's Service ALOC from other Nodes and measure the connectivity.

#### Fault Injections

- Nodes are constantly moving.
- Nodes constantly fail for a short period.
- Packet loss radio set to 20%.

#### Pass Criteria

- Max Delay (of all nodes) <= 3600s

## Latency

### Ping Latency

#### Test Purpose

Measure the ping delay for different hops and different data sizes.

#### Topology

- 6 Routers forming Circle 1
- 6 Routers forming Circle 2

#### Procedure

- Boot up Nodes.
- Wait for them to form a single partition.
- Ping Nodes within 1 hop.
- Ping Nodes within 2 hops.
- Ping Nodes without 3 hops.

#### Fault Injections

- None

#### Pass Criteria

- Max ping latency < 100ms

## Commissioning

### Commissioning

#### Test Purpose

Make sure Joiners can join a Thread network by Commissioning.

#### Topology

- 5x5 Nodes, with Commissioner at the center.

#### Procedure

- Boot up Nodes.
- Start the Commissioner.
- Start Joiners to join (2 Joiners joins concurrently).
- Wait for all Joiners to join successfully.

#### Fault Injections

- Packet loss ratio set to 20%

#### Pass Criteria

- All Joiners joined successfully
- Join process success rate >= 90%
- Average Join duration <= 20s

## Multicast

### CoAP Multicast

#### Test Purpose

Make sure CoAP multicast messgaes can reach all nodes with reasonable coverage and delay.

#### Topology

- 8 Routers (one BR)
- 8 FEDs
- 8 MEDs
- 8 SEDs

#### Procedure

- Boot up all Nodes.
- BR sends CoAP multicast to all Nodes.
- Measure the coverage and delay.

#### Fault Injections

- Packet loss ratio set to 50%

#### Pass Criteria

- Average coverage >= 70%
- Average delay < 200ms for non-SEDs
- Average delay < 2000ms for SEDs

## OTNS

### OTNS Performance

#### Test Purpose

Make sure OTNS can run large-scale simulations efficiently. Simulate a number of nodes at max speed without injected
traffic or failure for 1 hour (simulation time), and measure the execution time.

#### Topology

- 4x8 Routers

#### Procedure

- Boot up 4x8 Routers.
- Wait for them to form a single partition.
- Simulate for 1 hour.
- Measure the program execution time.

#### Fault Injections

None.

#### Pass Criteria

- Execution time <= 30s
