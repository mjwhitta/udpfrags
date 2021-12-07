package udpfrags_test

import (
	"net"
	"testing"

	"gitlab.com/mjwhitta/udpfrags"
)

func TestPacketLoss(t *testing.T) {
	var addr *net.UDPAddr
	var e error
	var expected string = "missing 1 fragments"
	var pkt *udpfrags.UDPPkt

	if addr, e = net.ResolveUDPAddr("udp", ":1194"); e != nil {
		t.Errorf("got: %s; want: nil", e.Error())
	}

	pkt = udpfrags.NewUDPPkt(addr, 1)
	if _, e = pkt.Hash(); e == nil {
		t.Errorf("got: nil; want: %s", expected)
	} else if e.Error() != expected {
		t.Errorf("got: %s; want: %s", e.Error(), expected)
	}
}
