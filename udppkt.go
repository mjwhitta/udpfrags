package udpfrags

import (
	"fmt"
	"net"
)

// UDPPkt is a combination of data ([]byte) and the *net.UDPAddr that
// sent the data.
type UDPPkt struct {
	Addr      *net.UDPAddr
	Data      []byte
	fragments [][]byte
	Length    int
	numFrags  int
}

// NewUDPPkt will return a pointer to a new UDPPkt instance.
func NewUDPPkt(addr *net.UDPAddr, frags int) *UDPPkt {
	return &UDPPkt{
		Addr:      addr,
		fragments: make([][]byte, frags),
		numFrags:  frags,
	}
}

// AddFragment will add a fragment to the UDPPkt.
func (p *UDPPkt) AddFragment(frag int, data []byte) {
	// Have to copy b/c data will be overwritten by next received
	// fragment
	p.fragments[frag-1] = make([]byte, len(data))
	copy(p.fragments[frag-1], data)
}

// Finalize will put all the fragments back together.
func (p *UDPPkt) Finalize() error {
	for i := 0; i < p.numFrags; i++ {
		if len(p.fragments[i]) == 0 {
			return fmt.Errorf(
				"Packet loss detected from %s",
				p.Addr.String(),
			)
		}

		p.Data = append(p.Data, p.fragments[i]...)
	}

	p.Length = len(p.Data)
	return nil
}
