package udpfrags

import (
	"net"

	"github.com/mjwhitta/errors"
	"github.com/mjwhitta/frgmnt"
)

// UDPPkt is a combination of data ([]byte) and the *net.UDPAddr that
// sent the data.
type UDPPkt struct {
	Addr    *net.UDPAddr
	builder *frgmnt.Builder
	Data    []byte
	Length  uint64
}

// NewUDPPkt will return a pointer to a new UDPPkt instance.
func NewUDPPkt(addr *net.UDPAddr, frags uint64) *UDPPkt {
	return &UDPPkt{Addr: addr, builder: frgmnt.NewByteBuilder(frags)}
}

func (p *UDPPkt) addFragment(frag uint64, data []byte) error {
	if e := p.builder.Add(frag, data); e != nil {
		return errors.Newf("failed to add fragment %d: %w", frag, e)
	}

	return nil
}

func (p *UDPPkt) finalize() error {
	var e error

	if p.Data, e = p.builder.Get(); e != nil {
		return errors.Newf("failed to get reassembled data: %w", e)
	}

	p.Length = uint64(len(p.Data))

	return nil
}

func (p *UDPPkt) finished() bool {
	return p.builder.Finished()
}

// Hash will return a SHA256 has of the packet's data.
func (p *UDPPkt) Hash() (s string, e error) {
	if s, e = p.builder.Hash(); e != nil {
		e = errors.Newf("failed to get reassembled data: %w", e)
	}

	return
}
