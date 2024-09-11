#!/usr/bin/env python3
# Copyright (c) 2020-2022, The OTNS Authors.
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
#
import logging
import os
import queue
import random
import socket
import struct
import threading
import time
import unittest
from typing import Dict, List, Tuple

import grpc

from OTNSTestCase import OTNSTestCase
from otns.cli import errors, OTNS
from otns.proto import visualize_grpc_pb2
from otns.proto import visualize_grpc_pb2_grpc


class UDPSignaler(object):
    """Signaler for UDP messages.
    """

    def __init__(self, id: int):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.source_addr = ("localhost", 9000 + id)
        self.dest_addr = ("localhost", 9000)
        self.sock.bind(self.source_addr)

    def emit_status(self, status: str) -> None:
        data = status.encode("ascii")
        event_packet = struct.pack("<QBH", 0, 5, len(data)) + data
        self.sock.sendto(event_packet, self.dest_addr)

    def close(self) -> None:
        self.sock.close()


class GRPCThread(threading.Thread):
    """Thread to listen for expected gRPC stream response.
    """

    def __init__(self,
                 stream,
                 contain: List[str],
                 not_contain: List[str],
                 exception_queue: queue.Queue,
                 time_limit=10):
        threading.Thread.__init__(self)
        self.stream = stream
        self.contain = contain
        self.not_contain = not_contain
        self.exception_queue = exception_queue
        self.time_limit = time_limit

    def run(self):
        start_time = time.time()
        try:
            for event in self.stream:
                if time.time() - start_time > self.time_limit:
                    self.exception_queue.put(errors.UnexpectedError("Expectation not fulfilled within time limit"))
                    return
                if (all(s in str(event) for s in self.contain) and not any(s in str(event) for s in self.not_contain)):
                    return
        except Exception as error:
            self.exception_queue.put(error)


class GRPCClient(object):
    """Listener for gRPC Visualize stream.
    """

    def __init__(self):
        channel = grpc.insecure_channel("localhost:8999")
        grpc.channel_ready_future(channel).result(timeout=10)
        self.stub = visualize_grpc_pb2_grpc.VisualizeGrpcServiceStub(channel)

    def send_command(self, command: str):
        self.stub.Command(visualize_grpc_pb2.CommandRequest(command=command))

    def expect_response(self, contain: List[str], not_contain: List[str]) -> GRPCThread:
        exception_queue = queue.Queue()
        stream = self.stub.Visualize(visualize_grpc_pb2.VisualizeRequest())
        expect_thread = GRPCThread(stream, contain, not_contain, exception_queue)
        expect_thread.start()
        return exception_queue, expect_thread


class RealTests(OTNSTestCase):

    def setUp(self) -> None:
        self.ns = OTNS(otns_args=[
            "-ot-script", "none", "-real", "-ot-cli", "otns-silk-proxy", "-listen", ":9000", "-log", "debug"
        ])
        # wait for OTNS gRPC server to start
        time.sleep(0.3)
        self.grpc_client = GRPCClient()
        self.udp_signalers = {}

    def tearDown(self):
        for udp_signaler in self.udp_signalers.values():
            udp_signaler.close()

        super().tearDown()

    def _wait_for_expect(self, exception_queue, expect_thread):
        while expect_thread.is_alive():
            try:
                exception = exception_queue.get(block=False)
            except queue.Empty:
                pass
            else:
                self.fail(exception)

            expect_thread.join(0.1)

    def create_expectation(self, contain: List[str] = None, not_contain: List[str] = None):
        return self.grpc_client.expect_response(contain=contain if contain is not None else [],
                                                not_contain=not_contain if not_contain is not None else [])

    def expect(self, expectation: Tuple[queue.Queue, GRPCThread]):
        exception_queue, expect_thread = expectation
        self._wait_for_expect(exception_queue, expect_thread)

    def expect_response(self,
                        contain: List[str] = None,
                        not_contain: List[str] = None,
                        action=None,
                        go_step: float = 0.1):
        exception_queue, expect_thread = self.grpc_client.expect_response(contain=contain, not_contain=not_contain)
        action()
        self.go(go_step)
        self._wait_for_expect(exception_queue, expect_thread)

    def testAddNode(self):
        node_id = random.randint(1, 10)
        self.expect_response(contain=["add_node", f"node_id: {node_id}", "x: 100", "y: 100"],
                             action=lambda: self.grpc_client.send_command(f"add router x 100 y 100 id {node_id}"))

    def testUpdateExtaddr(self):
        # create node
        node_id = random.randint(1, 10)
        signaler = UDPSignaler(node_id)
        self.udp_signalers[node_id] = signaler
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id}")
        self.go(0.1)

        extaddr = random.getrandbits(64)
        self.expect_response(contain=["on_ext_addr_change", f"node_id: {node_id}", f"ext_addr: {extaddr:d}"],
                             action=lambda: signaler.emit_status(f"extaddr={extaddr:016x}"))

    def testUpdateRLOC16(self):
        # create node
        node_id = random.randint(1, 10)
        signaler = UDPSignaler(node_id)
        self.udp_signalers[node_id] = signaler
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id}")
        self.go(0.1)

        rloc16 = random.getrandbits(16)
        self.expect_response(contain=["set_node_rloc16", f"node_id: {node_id}", f"rloc16: {rloc16:05d}"],
                             action=lambda: signaler.emit_status(f"rloc16={rloc16:05d}"))

    def testUpdatePartitionID(self):
        # create node
        node_id = random.randint(1, 10)
        signaler = UDPSignaler(node_id)
        self.udp_signalers[node_id] = signaler
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id}")
        self.go(0.1)

        par_id = random.getrandbits(32)
        self.expect_response(contain=["set_node_partition_id", f"node_id: {node_id}", f"partition_id: {par_id:d}"],
                             action=lambda: signaler.emit_status(f"parid={par_id:08x}"))

    def testUpdateRole(self):
        # create node
        node_id = random.randint(1, 10)
        signaler = UDPSignaler(node_id)
        self.udp_signalers[node_id] = signaler
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id}")
        self.go(0.1)

        self.expect_response(contain=["set_node_role", f"node_id: {node_id}", "role: OT_DEVICE_ROLE_LEADER"],
                             action=lambda: signaler.emit_status("role=4"))

        self.expect_response(contain=["set_node_role", f"node_id: {node_id}", "role: OT_DEVICE_ROLE_ROUTER"],
                             action=lambda: signaler.emit_status("role=3"))

        self.expect_response(contain=["set_node_role", f"node_id: {node_id}", "role: OT_DEVICE_ROLE_CHILD"],
                             action=lambda: signaler.emit_status("role=2"))

        self.expect_response(contain=["set_node_role", f"node_id: {node_id}", "role: OT_DEVICE_ROLE_DETACHED"],
                             action=lambda: signaler.emit_status("role=1"))

        self.expect_response(contain=["set_node_role", f"node_id: {node_id}"],
                             not_contain=["role: OT_DEVICE_ROLE"],
                             action=lambda: signaler.emit_status("role=0"))

    def testUpdateMode(self):
        # create node
        node_id = random.randint(1, 10)
        signaler = UDPSignaler(node_id)
        self.udp_signalers[node_id] = signaler
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id}")
        self.go(0.1)

        self.expect_response(
            contain=["set_node_mode", f"node_id: {node_id}", "secure_data_requests: true", "full_network_data: true"],
            not_contain=["rx_on_when_idle: true", "full_thread_device: true"],
            action=lambda: signaler.emit_status("mode=n"))

        self.expect_response(contain=[
            "set_node_mode", f"node_id: {node_id}", "rx_on_when_idle: true", "secure_data_requests: true",
            "full_thread_device: true", "full_network_data: true"
        ],
                             action=lambda: signaler.emit_status("mode=rdn"))

    def testUpdateChildren(self):
        # create node
        node_id_1 = random.randint(1, 5)
        extaddr_1 = random.getrandbits(64)
        signaler_1 = UDPSignaler(node_id_1)
        self.udp_signalers[node_id_1] = signaler_1
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id_1}")
        self.go(0.1)
        signaler_1.emit_status(f"extaddr={extaddr_1:016x}")

        node_id_2 = random.randint(6, 10)
        extaddr_2 = random.getrandbits(64)
        signaler_2 = UDPSignaler(node_id_2)
        self.udp_signalers[node_id_2] = signaler_2
        self.grpc_client.send_command(f"add router x 200 y 200 id {node_id_2}")
        self.go(0.1)
        signaler_2.emit_status(f"extaddr={extaddr_2:016x}")

        expectation = self.create_expectation(["add_child_table", f"node_id: {node_id_1}", f"ext_addr: {extaddr_2:d}"])
        signaler_1.emit_status("role=3")
        signaler_2.emit_status("role=2")
        signaler_1.emit_status(f"child_added={extaddr_2:016x}")
        self.go(0.1)
        self.expect(expectation)

        expectation = self.create_expectation(
            ["remove_child_table", f"node_id: {node_id_1}", f"ext_addr: {extaddr_2:d}"])
        signaler_1.emit_status("role=3")
        signaler_2.emit_status("role=1")
        signaler_1.emit_status(f"child_removed={extaddr_2:016x}")
        self.go(0.1)
        self.expect(expectation)

    def testUpdateRouter(self):
        # create node
        node_id_1 = random.randint(1, 5)
        extaddr_1 = random.getrandbits(64)
        signaler_1 = UDPSignaler(node_id_1)
        self.udp_signalers[node_id_1] = signaler_1
        self.grpc_client.send_command(f"add router x 100 y 100 id {node_id_1}")
        self.go(0.1)
        signaler_1.emit_status(f"extaddr={extaddr_1:016x}")

        node_id_2 = random.randint(6, 10)
        extaddr_2 = random.getrandbits(64)
        signaler_2 = UDPSignaler(node_id_2)
        self.udp_signalers[node_id_2] = signaler_2
        self.grpc_client.send_command(f"add router x 200 y 200 id {node_id_2}")
        self.go(0.1)
        signaler_2.emit_status(f"extaddr={extaddr_2:016x}")

        signaler_1.emit_status("role=3")
        signaler_2.emit_status("role=3")

        self.expect_response(contain=["add_router_table", f"node_id: {node_id_1}", f"ext_addr: {extaddr_2:d}"],
                             action=lambda: signaler_1.emit_status(f"router_added={extaddr_2:016x}"))

        self.expect_response(contain=["remove_router_table", f"node_id: {node_id_1}", f"ext_addr: {extaddr_2:d}"],
                             action=lambda: signaler_1.emit_status(f"router_removed={extaddr_2:016x}"))


if __name__ == '__main__':
    unittest.main()
