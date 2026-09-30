package main

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"syscall"
)

const maxListenPort = 65535

// listenHTTP binds addr. If that port is in use, it tries the next port through maxListenPort.
// Port 0 is left for the OS to assign. shifted is true when the bound port differs from the request.
func listenHTTP(addr string) (net.Listener, string, bool, error) {
	return listenFrom(addr, maxListenPort)
}

func listenFrom(addr string, maxPort int) (net.Listener, string, bool, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, "", false, fmt.Errorf("listen address %q: %w", addr, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 0 || port > maxListenPort {
		return nil, "", false, fmt.Errorf("listen port %q: invalid", portStr)
	}
	if port == 0 {
		ln, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, "", false, fmt.Errorf("listen %s: %w", addr, err)
		}
		return ln, ln.Addr().String(), false, nil
	}
	if port > maxPort {
		return nil, "", false, fmt.Errorf("no free port from %d through %d", port, maxPort)
	}

	var lastErr error
	for p := port; p <= maxPort; p++ {
		try := net.JoinHostPort(host, strconv.Itoa(p))
		ln, err := net.Listen("tcp", try)
		if err == nil {
			return ln, try, p != port, nil
		}
		lastErr = err
		if !isAddrInUse(err) {
			return nil, "", false, fmt.Errorf("listen %s: %w", try, err)
		}
	}
	return nil, "", false, fmt.Errorf("no free port from %d through %d: %w", port, maxPort, lastErr)
}

// isAddrInUse reports a bind failure caused by a taken port.
// On Windows the errno is WSAEADDRINUSE (10048), which is not syscall.EADDRINUSE.
func isAddrInUse(err error) bool {
	return errors.Is(err, syscall.EADDRINUSE) || errors.Is(err, syscall.Errno(10048))
}

// httpURL turns a listen address into a browser URL. Unspecified hosts become localhost.
func httpURL(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "http://" + addr
	}
	switch host {
	case "", "0.0.0.0", "::", "[::]":
		host = "localhost"
	}
	return "http://" + net.JoinHostPort(host, port)
}
