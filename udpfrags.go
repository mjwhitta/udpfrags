package udpfrags

import (
	"encoding/binary"
	"net"
	"strings"

	"github.com/mjwhitta/errors"
	"github.com/mjwhitta/frgmnt"
)

// Recv will create a background thread to receive UDP fragments. It
// will return a channel for incoming UDPPkts and a channel for
// errors. The background thread is terminated when the specified
// *net.UDPConn is closed by the caller.
func Recv(c *net.UDPConn) (chan *UDPPkt, chan error, error) {
	var errs chan error = make(chan error, QueueSize)
	var msgs chan *UDPPkt = make(chan *UDPPkt, QueueSize)

	if c == nil {
		return msgs, errs, errors.New("UDP connection is nil")
	}

	// Receive fragments in background thread
	go recvFrag(c, msgs, errs)

	// Return channels to the caller
	return msgs, errs, nil
}

func recvFrag(c *net.UDPConn, msgs chan *UDPPkt, errs chan error) {
	var addr *net.UDPAddr
	var e error
	var frag uint64
	var frags uint64
	var isClosed bool
	var n int
	var q map[string]*UDPPkt = map[string]*UDPPkt{}
	var recv []byte = make([]byte, bufSize)

	for {
		// Read incoming fragment
		if n, addr, e = c.ReadFromUDP(recv); e != nil {
			isClosed = strings.HasSuffix(
				e.Error(),
				"closed network connection",
			)
			if isClosed {
				break
			}

			errs <- errors.Newf("failed to read: %w", e)

			if strings.HasSuffix(e.Error(), "i/o timeout") {
				break
			}

			continue
		} else if (addr == nil) || (n <= fragHdrSize) {
			continue
		}

		// Parse header
		frag = binary.BigEndian.Uint64(recv[:fragHdrSize/2])
		frags = binary.BigEndian.Uint64(
			recv[fragHdrSize/2 : fragHdrSize],
		)

		// Create UDPPkt, if needed
		if _, ok := q[addr.String()]; !ok {
			q[addr.String()] = NewUDPPkt(addr, frags)
		}

		// Append buffer to appropriate UDPPkt
		e = q[addr.String()].addFragment(frag, recv[fragHdrSize:n])
		if e != nil {
			errs <- e
			continue
		}

		// If last fragment, return data/error via channels
		if q[addr.String()].finished() {
			if e = q[addr.String()].finalize(); e != nil {
				errs <- e
			} else {
				msgs <- q[addr.String()]
			}

			delete(q, addr.String())
		}
	}

	close(errs)
	close(msgs)
}

// Send will create a *net.UDPConn if one isn't provided. It will then
// send the specified data in fragments that can be put back together
// with Recv.
func Send(
	c *net.UDPConn,
	addr *net.UDPAddr,
	data []byte,
) (*net.UDPConn, error) {
	var e error
	var s *frgmnt.Streamer = frgmnt.NewByteStreamer(
		data,
		//nolint:gosec // G115 - size is 16, no overflow here
		bufSize-uint64(fragHdrSize),
	)

	// Initialize connection, if needed
	if c == nil {
		if addr == nil {
			e = errors.New("UDPConn and UDPAddr are both nil")
			return nil, e
		}

		// Create connection
		if c, e = net.DialUDP("udp", nil, addr); e != nil {
			e = errors.Newf("failed to create connection: %w", e)
			return nil, e
		}

		// And set addr to nil
		addr = nil
	}

	// Send data in fragments
	e = s.Each(
		func(fragNum uint64, numFrags uint64, frag []byte) error {
			var pkt []byte = make([]byte, len(frag)+fragHdrSize)

			binary.BigEndian.PutUint64(pkt[:fragHdrSize/2], fragNum)
			binary.BigEndian.PutUint64(
				pkt[fragHdrSize/2:fragHdrSize],
				numFrags,
			)
			copy(pkt[fragHdrSize:], frag)

			if addr == nil {
				if _, e = c.Write(pkt); e != nil {
					e = errors.Newf("failed to send fragment: %w", e)
					return e
				}
			} else {
				if _, e = c.WriteTo(pkt, addr); e != nil {
					e = errors.Newf("failed to send fragment: %w", e)
					return e
				}
			}

			return nil
		},
	)
	if e != nil {
		return nil, errors.Newf("failed to write: %w", e)
	}

	return c, nil
}

// SetBufferSize will set the maximum size of each fragment.
func SetBufferSize(size uint64) error {
	//nolint:gosec // G115 - size is 16, no overflow here
	if size <= uint64(fragHdrSize) {
		return errors.Newf("buffer size should be > %d", fragHdrSize)
	}

	bufSize = size

	return nil
}
