// Copyright (c) 2020-2024, The OTNS Authors.
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
	"encoding/binary"
	"math/rand"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/ipv6"
)

func removeAllFiles(globPath string) error {
	files, err := filepath.Glob(globPath)
	if err != nil {
		return err
	}
	for _, f := range files {
		if err := os.Remove(f); err != nil {
			return err
		}
	}
	return nil
}

func getCommitFromOtVersion(ver string) string {
	if strings.HasPrefix(ver, "OPENTHREAD/") && len(ver) >= 13 {
		commit := ver[11:]
		idx := strings.Index(commit, ";")
		if idx > 0 {
			commit = commit[0:idx]
			return commit
		}
	}
	return ""
}

func mergeNodeCounters(counters ...NodeCounters) NodeCounters {
	res := make(NodeCounters)
	for _, c := range counters {
		for k, v := range c {
			res[k] = v
		}
	}
	return res
}

func randomString(length int) string {
	chars := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	b := make([]byte, length)
	for i := 0; i < length; i++ {
		b[i] = chars[rand.Intn(len(chars))]
	}
	return string(b)
}

// SerializeIp6Header serializes an IPv6 header.
func serializeIp6Header(ipv6 *ipv6.Header, payloadLen int) []byte {
	data := make([]byte, 40)
	data[0] = uint8((ipv6.Version)<<4) | uint8(ipv6.TrafficClass>>4)
	data[1] = uint8(ipv6.TrafficClass<<4) | uint8(ipv6.FlowLabel>>16)
	binary.BigEndian.PutUint16(data[2:], uint16(ipv6.FlowLabel))
	binary.BigEndian.PutUint16(data[4:], uint16(payloadLen))
	data[6] = byte(ipv6.NextHeader)
	data[7] = byte(ipv6.HopLimit)
	copy(data[8:], ipv6.Src)
	copy(data[24:], ipv6.Dst)

	return data
}

// serializeUdpHeader serializes a UDP datagram header.
func serializeUdpHeader(udpHdr *UdpHeader) []byte {
	data := make([]byte, 8)
	binary.BigEndian.PutUint16(data, udpHdr.SrcPort)
	binary.BigEndian.PutUint16(data[2:], udpHdr.DstPort)
	binary.BigEndian.PutUint16(data[4:], udpHdr.Length)
	binary.BigEndian.PutUint16(data[6:], udpHdr.Checksum)
	return data
}

// calcUdpChecksum calculates the UDP checksum per RFC 2460 and writes it into udpHeader.Checksum.
func calcUdpChecksum(srcIp6Addr netip.Addr, dstIp6Addr netip.Addr, udpHeader *UdpHeader, udpPayload []byte) {
	sum := uint32(0)
	udpLen := uint16(len(udpPayload) + 8)

	pseudoHdr := make([]byte, 40)

	// IPv6 pseudo-header RFC 2460
	copy(pseudoHdr[0:16], srcIp6Addr.AsSlice())
	copy(pseudoHdr[16:32], dstIp6Addr.AsSlice())
	binary.BigEndian.PutUint32(pseudoHdr[32:36], uint32(udpLen)) // not including UDP header len
	pseudoHdr[39] = 17                                           // UDP next-header

	// append UDP header (with 0x0000 checksum) and UDP payload
	udpHeader.Checksum = 0x0000
	data := append(pseudoHdr, serializeUdpHeader(udpHeader)...)
	data = append(data, udpPayload...)

	for ; len(data) >= 2; data = data[2:] {
		sum += uint32(data[0])<<8 | uint32(data[1])
	}
	if len(data) > 0 {
		sum += uint32(data[0]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	csum := ^uint16(sum)
	if csum == 0 {
		csum = 0xffff
	}
	udpHeader.Checksum = csum
}

// createIp6UdpDatagram creates an IPv6+UDP datagram, including UDP payload.
func createIp6UdpDatagram(srcPort uint16, dstPort uint16, srcIp6Addr netip.Addr, dstIp6Addr netip.Addr, hopLimit int, udpPayload []byte) []byte {
	udpLen := len(udpPayload) + 8
	udpHeader := &UdpHeader{
		SrcPort:  srcPort,
		DstPort:  dstPort,
		Length:   uint16(udpLen),
		Checksum: 0,
	}
	calcUdpChecksum(srcIp6Addr, dstIp6Addr, udpHeader, udpPayload)
	udpHeaderSer := serializeUdpHeader(udpHeader)

	ip6Header := &ipv6.Header{
		Version:      6,
		TrafficClass: 0,
		FlowLabel:    0,
		PayloadLen:   udpLen,
		NextHeader:   17, // UDP next-header id
		HopLimit:     hopLimit,
		Src:          srcIp6Addr.AsSlice(),
		Dst:          dstIp6Addr.AsSlice(),
	}
	ip6Datagram := serializeIp6Header(ip6Header, udpLen)
	ip6Datagram = append(ip6Datagram, udpHeaderSer...)
	ip6Datagram = append(ip6Datagram, udpPayload...)

	return ip6Datagram
}
