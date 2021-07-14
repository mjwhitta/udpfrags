# UDPFrags

<a href="https://www.buymeacoffee.com/mjwhitta">🍪 Buy me a cookie</a>

[![Go Report Card](https://goreportcard.com/badge/gitlab.com/mjwhitta/udpfrags)](https://goreportcard.com/report/gitlab.com/mjwhitta/udpfrags)

## What is this?

A simple Go library to send and receive data via UDP in fragments.

## How to install

Open a terminal and run the following:

```
$ go get --ldflags="-s -w" --trimpath -u gitlab.com/mjwhitta/udpfrags
```

## Usage

```
package main

import (
    "net"
    ...

    "gitlab.com/mjwhitta/udpfrags"
    ...
)

func clientExample() error {
    var addr *net.UDPAddr
    var conn *net.UDPConn
    var e error
    var errs chan error
    var pkts chan *udpfrags.UDPPkt
    var wait = make(chan struct{}, 1)

    // Resolve address
    if addr, e = net.ResolveUDPAddr("udp", ":1194"); e != nil {
        return e
    }

    // Send data
    if conn, e = udpfrags.Send(nil, addr, []byte("hello")); e != nil {
        return e
    }
    defer conn.Close()

    // Set timeout
    e = conn.SetReadDeadline(time.Now().Add(time.Second))
    if e != nil {
        return e
    }

    // Receive response
    if pkts, errs, e = udpfrags.Recv(conn); e != nil {
        return e
    }

    // Loop thru errors
    go func() {
        for e := range errs {
            // Handle errors
        }

        wait <- struct{}{}
        close(wait)kwj;
    }()

    // Get received message
    for pkt := range pkts {
        // Handle message

        // Close connection to kill background thread
        if e = conn.Close(); e != nil {
            // Handle error
        }
    }

    <-wait

    return nil
}

func serverExample() error {
    var addr *net.UDPAddr
    var e error
    var errs chan error
    var pkts chan *udpfrags.UDPPkt
    var srv *net.UDPConn
    var wait = make(chan struct{}, 1)

    // Initialize UDP server
    if addr, e = net.ResolveUDPAddr("udp", ":1194"); e != nil {
        return e
    } else if srv, e = net.ListenUDP("udp", addr); e != nil {
        return e
    }
    defer srv.Close()

    // Start listening
    if pkts, errs, e = udpfrags.Recv(srv); e != nil {
        return e
    }

    // Loop thru errors
    go func() {
        for e := range errs {
            // Handle errors
        }

        wait <- struct{}{}
        close(wait)
    }()

    // Loop thru received messages
    for pkt := range pkts {
        go func(p *udpfrags.UDPPkt) {
            var echo []byte = p.Data

            // Handle message in new thread

            // Send response (here an echo)
            if _, e = udpfrags.Send(srv, p.Addr, echo); e != nil {
                // Handle error
            }

            // Only when finished:
            // Close connection to kill background receiving thread
            if e = srv.Close(); e != nil {
                // Handle error
            }
        }(pkt)
    }

    <-wait

    return nil
}
```

## Links

- [Source](https://gitlab.com/mjwhitta/udpfrags)
