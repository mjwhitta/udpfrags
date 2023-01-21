package udpfrags_test

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"testing"
	"time"

	"github.com/mjwhitta/udpfrags"
	assert "github.com/stretchr/testify/require"
)

func echoClient(t *testing.T, address string) {
	var actual string
	var addr *net.UDPAddr
	var conn *net.UDPConn
	var data [4096]byte
	var e error
	var errs chan error
	var expected string
	var hash [sha256.Size]byte
	var n int
	var pkts chan *udpfrags.UDPPkt
	var wait = make(chan struct{}, 1)

	// Resolve address
	addr, e = net.ResolveUDPAddr("udp", address)
	assert.Nil(t, e)
	assert.NotNil(t, addr)

	// Generate random data
	n, e = rand.Read(data[:])
	assert.Nil(t, e)
	assert.Equal(t, len(data[:]), n)

	// Calculate hash
	hash = sha256.Sum256(data[:n])
	expected = hex.EncodeToString(hash[:])

	// Send data
	conn, e = udpfrags.Send(nil, addr, data[:n])
	assert.Nil(t, e)
	assert.NotNil(t, conn)
	defer conn.Close()

	// Set timeout
	e = conn.SetReadDeadline(time.Now().Add(time.Second))
	assert.Nil(t, e)

	// Receive echo
	pkts, errs, e = udpfrags.Recv(conn)
	assert.Nil(t, e)

	// Loop thru errors
	go func() {
		for e := range errs {
			assert.Nil(t, e)
		}

		wait <- struct{}{}
		close(wait)
	}()

	// Get received message
	for pkt := range pkts {
		// Calculate hash
		actual, e = pkt.Hash()
		assert.Nil(t, e)
		assert.Equal(t, expected, actual)

		// Close connection to kill background receiving thread
		e = conn.Close()
		assert.Nil(t, e)
	}

	<-wait
}

func echoServer(t *testing.T, address string) {
	var addr *net.UDPAddr
	var e error
	var errs chan error
	var pkts chan *udpfrags.UDPPkt
	var srv *net.UDPConn
	var wait = make(chan struct{}, 1)

	// Initialize UDP server
	addr, e = net.ResolveUDPAddr("udp", address)
	assert.Nil(t, e)
	assert.NotNil(t, addr)

	srv, e = net.ListenUDP("udp", addr)
	assert.Nil(t, e)
	assert.NotNil(t, srv)
	defer srv.Close() // Close connection to kill background thread

	// Start listening
	pkts, errs, e = udpfrags.Recv(srv)
	assert.Nil(t, e)

	// Loop thru errors
	go func() {
		for e := range errs {
			assert.Nil(t, e)
		}

		wait <- struct{}{}
		close(wait)
	}()

	// Loop thru received messages
	for pkt := range pkts {
		// No need to create thread as we are stopping after 1

		// Send echo
		_, e = udpfrags.Send(srv, pkt.Addr, pkt.Data)
		assert.Nil(t, e)

		// Close connection to kill background receiving thread
		e = srv.Close()
		assert.Nil(t, e)
	}

	<-wait
}

func TestSendRecv(t *testing.T) {
	var addr string = ":1194"
	var e error
	var wait = make(chan struct{}, 1)

	e = udpfrags.SetBufferSize(10)
	assert.NotNil(t, e)

	e = udpfrags.SetBufferSize(256)
	assert.Nil(t, e)

	go func() {
		echoServer(t, addr)
		wait <- struct{}{}
		close(wait)
	}()

	time.Sleep(time.Second)
	echoClient(t, addr)
	<-wait
}
