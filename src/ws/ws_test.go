package ws

import (
	"net"
	"testing"
)

func TestHandleRawTCPEchoesPayload(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer serverConn.Close()
	defer clientConn.Close()

	errCh := make(chan error, 1)
	go func() {
		errCh <- HandleRawTCP(serverConn)
	}()

	if _, err := clientConn.Write([]byte("hello\n")); err != nil {
		t.Fatalf("write failed: %v", err)
	}

	buf := make([]byte, 32)
	n, err := clientConn.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}

	if got := string(buf[:n]); got != "hello\n" {
		t.Fatalf("expected echoed payload, got %q", got)
	}

	if err := clientConn.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("handle raw tcp returned error: %v", err)
	}
}
