# UDPFrags

[![Yum](https://img.shields.io/badge/-Buy%20me%20a%20cookie-blue?style=for-the-badge&logo=cookiecutter)](https://www.buymeacoffee.com/mjwhitta)

[![Go Report Card](https://goreportcard.com/badge/github.com/mjwhitta/udpfrags)](https://goreportcard.com/report/github.com/mjwhitta/udpfrags)
![Workflow](https://github.com/mjwhitta/udpfrags/actions/workflows/ci.yaml/badge.svg?event=push)

## What is this?

A simple Go library to send and receive data via UDP in fragments.

## How to install

Open a terminal and run the following:

```
$ go get --ldflags="-s -w" --trimpath -u github.com/mjwhitta/udpfrags
```

## Usage

```
package main

import (
    "net"
    ...

    "github.com/mjwhitta/udpfrags"
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
        close(wait);
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

- [Source](https://github.com/mjwhitta/udpfrags)
