package udpfrags_test

import (
	"net"
	"testing"

	"gitlab.com/mjwhitta/udpfrags"
)

func TestPacketLoss(t *testing.T) {
	var addr *net.UDPAddr
	var e error
	var expected string
	var pkt *udpfrags.UDPPkt

	if addr, e = net.ResolveUDPAddr("udp", ":1194"); e != nil {
		t.Errorf("got: %s; want: nil", e.Error())
	}

	expected = "Packet loss detected from :1194"
	pkt = udpfrags.NewUDPPkt(addr, 3)
	pkt.AddFragment(1, []byte("Hello"))
	pkt.AddFragment(3, []byte("!"))

	if e = pkt.Finalize(); e == nil {
		t.Errorf("got: nil; want: %s", expected)
	} else {
		if e.Error() != expected {
			t.Errorf("got: %s; want: %s", e.Error(), expected)
		}
	}
}
