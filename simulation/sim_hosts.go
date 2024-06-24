// Copyright (c) 2024, The OTNS Authors.
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
// 3. Neither the name of the copyright holder nor the
//    names of its contributors may be used to endorse or promote products
//    derived from this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package simulation

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"

	"golang.org/x/net/ipv6"

	"github.com/openthread/ot-ns/event"
	"github.com/openthread/ot-ns/logger"
	"github.com/openthread/ot-ns/types"
)

const (
	udpHeaderLen = 8
	protocolUdp  = 17
)

// ConnId is a unique identifier/tuple for a TCP or UDP connection between a node and a simulated host.
type ConnId struct {
	NodeIp6Addr netip.Addr
	ExtIp6Addr  netip.Addr
	NodePort    uint16
	ExtPort     uint16
}

// SimConn is a two-way connection between a node's port and a sim-host's port.
type SimConn struct {
	Node               *Node // assumes a single BR also handles the return traffic.
	Conn               net.Conn
	Nat66State         *event.MsgToHostEventData
	PortMapped         uint16 // real localhost ::1 port on simulator machine, on which sim-host's port is mapped.
	UdpBytesUpstream   uint64 // total bytes UDP payload from node to sim-host (across all BRs)
	UdpBytesDownstream uint64 // total bytes UDP payload from sim-host to node (across all BRs)
}

// SimHostEndpoint represents a single endpoint (port) of a sim-host, potentially interacting with N >= 0 nodes.
type SimHostEndpoint struct {
	HostName   string
	Ip6Addr    netip.Addr
	Port       uint16 // destination UDP/TCP port as specified by the simulated node.
	PortMapped uint16 // actual sim-host port on [::1] to which specified port is mapped.
}

// SimHosts manages all connections between nodes and simulated hosts.
type SimHosts struct {
	sim   *Simulation
	Hosts map[SimHostEndpoint]struct{}
	Conns map[ConnId]*SimConn
}

func (sc *SimConn) close() {
	if sc.Conn != nil {
		_ = sc.Conn.Close()
	}
}

func NewSimHosts() *SimHosts {
	sh := &SimHosts{
		sim:   nil,
		Hosts: make(map[SimHostEndpoint]struct{}),
		Conns: make(map[ConnId]*SimConn),
	}
	return sh
}

func (sh *SimHosts) Init(sim *Simulation) {
	sh.sim = sim
}

func (sh *SimHosts) AddHost(host SimHostEndpoint) error {
	sh.Hosts[host] = struct{}{}
	// TODO check for conflicts with existing item
	return nil
}

func (sh *SimHosts) RemoveHost(host SimHostEndpoint) {
	delete(sh.Hosts, host)
	// delete and close all related connection state
	for id, conn := range sh.Conns {
		if conn.Nat66State.DstIp6Address == host.Ip6Addr &&
			conn.Nat66State.DstPort == host.Port {
			conn.close()
			delete(sh.Conns, id)
		}
	}
}

func (sh *SimHosts) GetTxBytes(host *SimHostEndpoint) uint64 {
	var n uint64 = 0
	for connId, simConn := range sh.Conns {
		if host.Ip6Addr == connId.ExtIp6Addr && host.Port == connId.ExtPort {
			n += simConn.UdpBytesDownstream
		}
	}
	return n
}

func (sh *SimHosts) GetRxBytes(host *SimHostEndpoint) uint64 {
	var n uint64 = 0
	for connId, simConn := range sh.Conns {
		if host.Ip6Addr == connId.ExtIp6Addr && host.Port == connId.ExtPort {
			n += simConn.UdpBytesUpstream
		}
	}
	return n
}

// handleUdpFromNode handles a UDP datagram coming from a node and checks to which sim-host to deliver it.
func (sh *SimHosts) handleUdpFromNode(node *Node, udpMetadata *event.MsgToHostEventData, udpData []byte) {
	var host SimHostEndpoint
	var err error
	var ok bool
	var simConn *SimConn

	found := false

	// find the first matching simulated host, if any.
	for host = range sh.Hosts {
		if host.Port == udpMetadata.DstPort && host.Ip6Addr == udpMetadata.DstIp6Address {
			found = true
		}
	}
	if !found {
		logger.Debugf("SimHosts: IPv6/UDP from node %d did not reach a sim-host destination: %+v", node.Id, udpMetadata)
		return
	}
	logger.Debugf("SimHosts: IPv6/UDP from node %d, to sim-host [::1]:%d (%d bytes)", node.Id, host.PortMapped, len(udpData))

	// FIXME
	/*
		// TODO if sending node doesn't know its own AIL IPv6 interface address, it uses 'unspecified'.
		if udpMetadata.SrcIp6Address == netip.IPv6Unspecified() {
			udpMetadata.SrcIp6Address, err = netip.ParseAddr(fmt.Sprintf("fc00::%d", node.Id)) // FIXME make configurable
			if err != nil {
				logger.Panicf("Unexpected error in IPv6 address generation for BR")
			}
		}
	*/

	// fetch existing conn object for the specific node/sim-host IP and port combo, if any.
	connId := ConnId{
		NodeIp6Addr: udpMetadata.SrcIp6Address,
		ExtIp6Addr:  udpMetadata.DstIp6Address,
		NodePort:    udpMetadata.SrcPort,
		ExtPort:     udpMetadata.DstPort,
	}
	if simConn, ok = sh.Conns[connId]; !ok {
		// create new connection
		simConn = &SimConn{
			Conn:               nil,
			Node:               node,
			Nat66State:         udpMetadata,
			PortMapped:         host.PortMapped,
			UdpBytesUpstream:   0,
			UdpBytesDownstream: 0,
		}
		simConn.Conn, err = net.Dial("udp", fmt.Sprintf("[::1]:%d", host.PortMapped))
		if err != nil {
			logger.Warnf("SimHosts could not connect to local UDP port %d: %v", host.PortMapped, err)
			simConn.close()
			return
		}

		// create reader thread - to process the sim-host's response traffic.
		go sh.udpReaderGoRoutine(simConn)

		// store created connection under its unique tuple ID
		sh.Conns[connId] = simConn
	}

	var n int
	n, err = simConn.Conn.Write(udpData)
	if err != nil {
		logger.Warnf("SimHosts could not write UDP data to [::1]:%d : %v", host.PortMapped, err)
		return
	}
	simConn.UdpBytesUpstream += uint64(n)
}

// handleUdpFromSimHost handles a UDP message coming from a sim-host and checks to which (BR) node to deliver it
// as an IPv6+UDP datagram. It performs a form of NAT66 with NPT to achieve this.
func (sh *SimHosts) handleUdpFromSimHost(simConn *SimConn, udpData []byte) {
	logger.Debugf("SimHosts: UDP datagram from sim-host [::1]:%d (%d bytes)", simConn.PortMapped, len(udpData))
	simConn.UdpBytesDownstream += uint64(len(udpData))
	hopLim := 63 // assume sim-host used 64 and BR decreases by 1. TODO: the local sim-host case (must be 64?)

	if simConn.Nat66State.SrcIp6Address == netip.IPv6Unspecified() {
		// send as UDP-event to node itself to handle.
		ev := &event.Event{
			Delay:  0,
			Type:   event.EventTypeUdpFromHost,
			Data:   udpData,
			NodeId: simConn.Node.Id,
			MsgToHostData: event.MsgToHostEventData{
				SrcPort:       simConn.Nat66State.DstPort, // simulates response back: ports reversed
				DstPort:       simConn.Nat66State.SrcPort,
				SrcIp6Address: simConn.Nat66State.DstIp6Address, // simulates response: addrs reversed
				DstIp6Address: simConn.Nat66State.SrcIp6Address,
			},
		}
		sh.sim.Dispatcher().PostEventAsync(ev)
		logger.Debugf("sh.sim.Dispatcher().PostEventAsync(ev) FIXME-UDP path %v, %+v", ev, ev.MsgToHostData)
		logger.Debugf("simConn.Nat66State UDP-path = %v", simConn.Nat66State)
		logger.Debugf("udpData = %s", hex.EncodeToString(udpData))
	} else {
		// send as IPv6-event to node, to let it forward to others on mesh.
		ip6Datagram := createIp6UdpDatagram(simConn.Nat66State.DstPort, simConn.Nat66State.SrcPort,
			simConn.Nat66State.DstIp6Address, simConn.Nat66State.SrcIp6Address, hopLim, udpData)

		// send IPv6+UDP datagram as event to node
		ev := &event.Event{
			Delay:  0,
			Type:   event.EventTypeIp6FromHost,
			Data:   ip6Datagram,
			NodeId: simConn.Node.Id,
			MsgToHostData: event.MsgToHostEventData{
				SrcPort:       simConn.Nat66State.DstPort, // simulates response back: ports reversed
				DstPort:       simConn.Nat66State.SrcPort,
				SrcIp6Address: simConn.Nat66State.DstIp6Address, // simulates response: addrs reversed
				DstIp6Address: simConn.Nat66State.SrcIp6Address,
			},
		}
		sh.sim.Dispatcher().PostEventAsync(ev)
		logger.Debugf("sh.sim.Dispatcher().PostEventAsync(ev) FIXME %v", ev)
		logger.Debugf("simConn.Nat66State = %v", simConn.Nat66State)
	}
}

func (sh *SimHosts) udpReaderGoRoutine(simConn *SimConn) {
	buf := make([]byte, types.OtMaxIp6DatagramLength)
	defer func() {
		if simConn.Conn != nil {
			_ = simConn.Conn.Close()
		}
	}()

	for {
		rlen, err := simConn.Conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			logger.Warnf("Error reading from sim-host [::1]:%d, closing : %v", simConn.PortMapped, err)
			break
		}
		if rlen >= types.OtMaxUdpPayloadLength {
			logger.Warnf("sim-host sent too large UDP data (%d bytes > %d), dropped", rlen, types.OtMaxUdpPayloadLength)
			continue
		}
		sh.handleUdpFromSimHost(simConn, buf[:rlen])
	}
}

func (sh *SimHosts) handleIp6FromNode(node *Node, ip6Metadata *event.MsgToHostEventData, ip6Data []byte) {
	var ip6Header *ipv6.Header
	var err error

	// check if header is IPv6?
	if ip6Header, err = ipv6.ParseHeader(ip6Data); err != nil {
		logger.Warnf("SimHosts could not parse as IPv6: %v", err)
		return
	}
	// if it's UDP - attempt to handle the datagram by a sim-host
	if ip6Header.Version == 6 && ip6Header.NextHeader == protocolUdp && len(ip6Data) > ipv6.HeaderLen+udpHeaderLen {
		udpData := ip6Data[ipv6.HeaderLen+udpHeaderLen:]
		sh.handleUdpFromNode(node, ip6Metadata, udpData)
	}
	// TODO TCP
}
