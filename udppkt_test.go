package udpfrags_test

import (
	"net"
	"strings"
	"testing"

	"gitlab.com/mjwhitta/udpfrags"
)

func TestPacketLoss(t *testing.T) {
	var addr *net.UDPAddr
	var e error
	var expected string = strings.Join(
		[]string{
			"udpfrags: failed to get reassembled data",
			"frgmnt: missing 1 fragments",
		},
		": ",
	)
	var pkt *udpfrags.UDPPkt

	if addr, e = net.ResolveUDPAddr("udp", ":1194"); e != nil {
		t.Errorf("\ngot: %s\nwant: nil", e.Error())
	}

	pkt = udpfrags.NewUDPPkt(addr, 1)
	if _, e = pkt.Hash(); e == nil {
		t.Errorf("\ngot: nil\nwant: %s", expected)
	} else if e.Error() != expected {
		t.Errorf("\ngot: %s\nwant: %s", e.Error(), expected)
	}
}
