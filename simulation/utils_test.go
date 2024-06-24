package simulation

import (
	"net/netip"
	"testing"

	"encoding/binary"
	"encoding/hex"

	"github.com/stretchr/testify/assert"
)

func TestCreateIp6UdpDatagram(t *testing.T) {
	src := netip.MustParseAddr("fc00::1234")
	dst := netip.MustParseAddr("fe80::abcd")
	udpPayload := []byte{1, 2, 3, 4, 5}

	ip6 := createIp6UdpDatagram(5683, 5683, src, dst, 64, udpPayload)
	assert.Equal(t, 53, len(ip6))
}

func TestUdpChecksum(t *testing.T) {
	// see example on https://stackoverflow.com/questions/30858973/udp-checksum-calculation-for-ipv6-packet
	src := netip.MustParseAddr("2100::1:abcd:0:0:1")
	dst := netip.MustParseAddr("fd00::160")
	udpPayload := []byte{0x12, 0x34, 0x56, 0x78}

	ip6 := createIp6UdpDatagram(9874, 9874, src, dst, 64, udpPayload)
	assert.Equal(t, 52, len(ip6))

	ip6HexStr := hex.EncodeToString(ip6)
	assert.Equal(t, "60000000000c11402100000000000001abcd000000000001fd00000000000000000000000000016026922692000c7ed512345678", ip6HexStr)
	udpChecksum := binary.BigEndian.Uint16(ip6[46:48])
	assert.Equal(t, uint16(0x7ed5), udpChecksum)
}
