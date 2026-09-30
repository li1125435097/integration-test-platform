package main

import (
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestListenFromSkipsBusyPort(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()

	_, portStr, err := net.SplitHostPort(busy.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}
	if port >= maxListenPort {
		t.Skip("no higher port to try")
	}

	ln, got, shifted, err := listenFrom(net.JoinHostPort("127.0.0.1", portStr), maxListenPort)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if !shifted {
		t.Fatalf("expected a shifted bind, got %s", got)
	}
	_, gotPortStr, err := net.SplitHostPort(got)
	if err != nil {
		t.Fatal(err)
	}
	gotPort, err := strconv.Atoi(gotPortStr)
	if err != nil {
		t.Fatal(err)
	}
	if gotPort <= port {
		t.Fatalf("expected port > %d, got %d", port, gotPort)
	}
}

func TestListenFromStopsAtMax(t *testing.T) {
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer busy.Close()

	_, portStr, err := net.SplitHostPort(busy.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}

	_, _, _, err = listenFrom(net.JoinHostPort("127.0.0.1", portStr), port)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no free port") {
		t.Fatal(err)
	}
}

func TestListenFromPortZero(t *testing.T) {
	ln, got, shifted, err := listenFrom("127.0.0.1:0", maxListenPort)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	if shifted {
		t.Fatalf("port 0 should not count as shifted, bound %s", got)
	}
	_, portStr, err := net.SplitHostPort(got)
	if err != nil {
		t.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatal(err)
	}
	if port == 0 {
		t.Fatalf("expected an assigned port, got %s", got)
	}
}

func TestListenFromRejectsBadAddress(t *testing.T) {
	_, _, _, err := listenFrom("127.0.0.1:abc", maxListenPort)
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), "no free port") {
		t.Fatal(err)
	}
}

func TestHTTPURL(t *testing.T) {
	if got := httpURL(":8081"); got != "http://localhost:8081" {
		t.Fatalf("got %s", got)
	}
	if got := httpURL("127.0.0.1:8081"); got != "http://127.0.0.1:8081" {
		t.Fatalf("got %s", got)
	}
}
