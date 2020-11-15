package udpfrags

import (
	"net"

	"gitlab.com/mjwhitta/frgmnt"
)

// UDPPkt is a combination of data ([]byte) and the *net.UDPAddr that
// sent the data.
type UDPPkt struct {
	Addr    *net.UDPAddr
	builder *frgmnt.Builder
	Data    []byte
	Length  int
}

// NewUDPPkt will return a pointer to a new UDPPkt instance.
func NewUDPPkt(addr *net.UDPAddr, frags int) *UDPPkt {
	return &UDPPkt{Addr: addr, builder: frgmnt.NewByteBuilder(frags)}
}

func (p *UDPPkt) addFragment(frag int, data []byte) error {
	return p.builder.Add(frag, data)
}

func (p *UDPPkt) finalize() error {
	var e error

	if p.Data, e = p.builder.Get(); e != nil {
		return e
	}

	p.Length = len(p.Data)

	return nil
}

func (p *UDPPkt) finished() bool {
	return p.builder.Finished()
}

// Hash will return a SHA256 has of the packet's data.
func (p *UDPPkt) Hash() (string, error) {
	return p.builder.Hash()
}
