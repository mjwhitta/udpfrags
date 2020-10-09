package udpfrags_test

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"net"
	"testing"
	"time"

	"gitlab.com/mjwhitta/udpfrags"
)

func echoClient(t *testing.T, address string) {
	var actual string
	var addr *net.UDPAddr
	var conn *net.UDPConn
	var data [4096]byte
	var e error
	var errs chan error
	var expected string
	var pkts chan *udpfrags.UDPPkt
	var wait = make(chan struct{}, 1)

	// Resolve address
	if addr, e = net.ResolveUDPAddr("udp", address); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}

	// Generate random data
	if _, e = rand.Read(data[:]); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}

	// Calculate hash
	expected = fmt.Sprintf("%x", sha256.Sum256(data[:]))

	// Send data
	if conn, e = udpfrags.Send(nil, addr, data[:]); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}
	defer conn.Close()

	// Set timeout
	e = conn.SetReadDeadline(time.Now().Add(time.Second))
	if e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}

	// Receive echo
	if pkts, errs, e = udpfrags.Recv(conn); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}

	// Loop thru errors
	go func() {
		for e := range errs {
			t.Errorf("got: %s; want: nil", e.Error())
		}

		wait <- struct{}{}
		close(wait)
	}()

	// Get received message
	for pkt := range pkts {
		// Calculate hash
		actual = fmt.Sprintf("%x", sha256.Sum256(pkt.Data))

		if actual != expected {
			t.Errorf("got: %s; want: %s", actual, expected)
		}

		// Close connection to kill background receiving thread
		if e = conn.Close(); e != nil {
			t.Errorf("got: %s; want: nil", e.Error())
		}
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
	if addr, e = net.ResolveUDPAddr("udp", address); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	} else if srv, e = net.ListenUDP("udp", addr); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}
	defer srv.Close() // Close connection to kill background thread

	// Start listening
	if pkts, errs, e = udpfrags.Recv(srv); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}

	// Loop thru errors
	go func() {
		for e := range errs {
			t.Errorf("got: %s; want: nil", e.Error())
		}

		wait <- struct{}{}
		close(wait)
	}()

	// Loop thru received messages
	for pkt := range pkts {
		// No need to create thread as we are stopping after 1

		// Send echo
		if _, e = udpfrags.Send(srv, pkt.Addr, pkt.Data); e != nil {
			t.Errorf("got: %s; want: nil", e.Error())
		}

		// Close connection to kill background receiving thread
		if e = srv.Close(); e != nil {
			t.Errorf("got: %s; want: nil", e.Error())
		}
	}

	<-wait
}

func TestSendRecv(t *testing.T) {
	var addr string = ":1194"
	var e error
	var wait = make(chan struct{}, 1)

	if e = udpfrags.SetBufferSize(10); e == nil {
		t.Errorf("got: nil; want: %s", "Buffer size should be >= 256")
	}

	if e = udpfrags.SetBufferSize(256); e != nil {
		t.Fatalf("got: %s; want: nil", e.Error())
	}

	go func() {
		echoServer(t, addr)
		wait <- struct{}{}
		close(wait)
	}()

	time.Sleep(time.Second)
	echoClient(t, addr)
	<-wait
}
