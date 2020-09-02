package udpfrags

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
)

// Recv will create a background thread to receive UDP fragments. It
// will return a channel for incoming UDPPkts and a channel for
// errors. The background thread is terminated when the specified
// *net.UDPConn is closed by the caller.
func Recv(c *net.UDPConn) (chan *UDPPkt, chan error, error) {
	var msgs = make(chan *UDPPkt, 1024)
	var errs = make(chan error, 1024)

	if c == nil {
		return msgs, errs, fmt.Errorf("UDP connection is nil")
	}

	// Receive fragments in background thread
	go recvFrag(c, msgs, errs)

	// Return channels to the caller
	return msgs, errs, nil
}

func recvFrag(c *net.UDPConn, msgs chan *UDPPkt, errs chan error) {
	var addr *net.UDPAddr
	var e error
	var frag uint32
	var frags uint32
	var isClosed bool
	var n int
	var q = map[string]*UDPPkt{}
	var recv = make([]byte, bufSize)

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

			errs <- e

			if strings.HasSuffix(e.Error(), "i/o timeout") {
				break
			}

			continue
		} else if (addr == nil) || (n <= 8) {
			continue
		}

		// Parse header
		frag = binary.BigEndian.Uint32(recv[:4])
		frags = binary.BigEndian.Uint32(recv[4:8])

		// Create UDPPkt, if needed
		if _, ok := q[addr.String()]; !ok {
			q[addr.String()] = NewUDPPkt(addr, int(frags))
		}

		// Append buffer to appropriate UDPPkt
		q[addr.String()].AddFragment(int(frag), recv[8:n])

		// If last fragment, return data/error via channels
		if frag == frags {
			if e = q[addr.String()].Finalize(); e != nil {
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

	// Initialize connection, if needed
	if c == nil {
		if addr == nil {
			return nil, fmt.Errorf("UDPConn and UDPAddr are both nil")
		}

		// Create connection
		if c, e = net.DialUDP("udp", nil, addr); e != nil {
			return nil, e
		}

		// And set addr to nil
		addr = nil
	}

	// Send data in fragments
	if e = sendFrag(c, addr, data); e != nil {
		return nil, e
	}

	return c, nil
}

func sendFrag(c *net.UDPConn, addr *net.UDPAddr, data []byte) error {
	var e error
	var frag uint32
	var frags uint32
	var inc int = bufSize - 8
	var max int = len(data)
	var pkt []byte
	var start int
	var stop int

	if c == nil {
		return fmt.Errorf("UDP connection is nil")
	}

	// Determine number of fragments
	frags = uint32(max / inc)
	if max%inc != 0 {
		frags++
	}

	// Loop thru fragments
	for i := 0; i < max; i += inc {
		// Calculate fragment size
		start = i
		stop = i + inc
		if stop > max {
			stop = max
		}

		// Build packet
		frag++
		pkt = make([]byte, stop-start+8)
		binary.BigEndian.PutUint32(pkt[:4], uint32(frag))
		binary.BigEndian.PutUint32(pkt[4:8], uint32(frags))
		copy(pkt[8:], data[start:stop])

		// Send fragment
		if addr == nil {
			if _, e = c.Write(pkt); e != nil {
				return e
			}
		} else {
			if _, e = c.WriteTo(pkt, addr); e != nil {
				return e
			}
		}
	}

	return nil
}

// SetBufferSize will set the maximum size of each fragment.
func SetBufferSize(size int) error {
	if size < 256 {
		return fmt.Errorf("Buffer size should be >= 256")
	}

	bufSize = size
	return nil
}
