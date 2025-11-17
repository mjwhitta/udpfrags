//nolint:godoclint // These are tests
package udpfrags_test

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/mjwhitta/udpfrags"
	assert "github.com/stretchr/testify/require"
)

func TestSendRecv(t *testing.T) {
	var actual string
	var addr *net.UDPAddr
	var address string = ":1194"
	var allErrs chan error = make(chan error, 16)
	var clientErrs chan error
	var clientPkts chan *udpfrags.UDPPkt
	var conn *net.UDPConn
	var data [4096]byte
	var e error
	var expected string
	var hash [sha256.Size]byte
	var n int
	var srv *net.UDPConn
	var svrErrs chan error
	var svrPkts chan *udpfrags.UDPPkt
	var wg sync.WaitGroup

	addr, e = net.ResolveUDPAddr("udp", address)
	assert.NoError(t, e)
	assert.NotNil(t, addr)

	e = udpfrags.SetBufferSize(10)
	assert.Error(t, e)

	e = udpfrags.SetBufferSize(256)
	assert.NoError(t, e)

	// Initialize UDP server
	srv, e = net.ListenUDP("udp", addr)
	assert.NoError(t, e)
	assert.NotNil(t, srv)
	defer func() {
		// Ensure close on test fail
		if srv != nil {
			_ = srv.Close()
		}
	}()

	// Start listening
	svrPkts, svrErrs, e = udpfrags.Recv(srv)
	assert.NoError(t, e)

	wg.Add(1)

	// Server goroutine
	go func() {
		var e error

		// Loop thru received messages
		for pkt := range svrPkts {
			// No need to fork as we are stopping after 1

			// Send echo
			_, e = udpfrags.Send(srv, pkt.Addr, pkt.Data)
			allErrs <- e

			// Close server to kill background receiving thread
			e = srv.Close()
			srv = nil

			allErrs <- e
		}

		// Loop thru errors
		for e = range svrErrs {
			allErrs <- e
		}

		wg.Done()
	}()

	// Generate random data
	n, e = rand.Read(data[:])
	assert.NoError(t, e)
	assert.Equal(t, len(data[:]), n)

	// Calculate hash
	hash = sha256.Sum256(data[:n])
	expected = hex.EncodeToString(hash[:])

	// Send data
	conn, e = udpfrags.Send(nil, addr, data[:n])
	assert.NoError(t, e)
	assert.NotNil(t, conn)
	defer func() {
		// Ensure close on test fail
		if conn != nil {
			_ = conn.Close()
		}
	}()

	// Set timeout
	e = conn.SetReadDeadline(time.Now().Add(time.Second))
	assert.NoError(t, e)

	// Receive echo
	clientPkts, clientErrs, e = udpfrags.Recv(conn)
	assert.NoError(t, e)

	wg.Add(1)

	// Client goroutine
	go func() {
		var e error

		// Get received message
		for pkt := range clientPkts {
			// Calculate hash
			actual, e = pkt.Hash()
			allErrs <- e

			// Close connection to kill background receiving thread
			e = conn.Close()
			conn = nil

			allErrs <- e
		}

		// Loop thru errors
		for e = range clientErrs {
			allErrs <- e
		}

		wg.Done()
	}()

	wg.Wait()
	close(allErrs)

	assert.Equal(t, expected, actual)

	for e = range allErrs {
		assert.NoError(t, e)
	}
}
