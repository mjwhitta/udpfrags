package udpfrags_test

import (
	"net"
	"testing"

	"github.com/mjwhitta/udpfrags"
	assert "github.com/stretchr/testify/require"
)

func TestPacketLoss(t *testing.T) {
	var addr *net.UDPAddr
	var e error

	addr, e = net.ResolveUDPAddr("udp", ":1194")
	assert.Nil(t, e)
	assert.NotNil(t, addr)

	_, e = udpfrags.NewUDPPkt(addr, 1).Hash()
	assert.NotNil(t, e)
}
