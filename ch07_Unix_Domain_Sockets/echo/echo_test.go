package echo

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
)

func TestEchoServerUnix(t *testing.T) {
	dir, err := os.MkdirTemp("", "echo_unix")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if rErr := os.RemoveAll(dir); rErr != nil {
			t.Error(rErr)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())
	socket := filepath.Join(dir, fmt.Sprintf("%d.sock", os.Getpid()))
	rAddr, err := streamingEchoServer(ctx, "unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer cancel()

	err = os.Chmod(socket, os.ModeSocket|0666)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.Dial("unix", rAddr.String())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = conn.Close() }()

	msg := []byte("ping")
	for i := 0; i < 3; i++ { // write 3 "ping" messages
		_, err = conn.Write(msg)
		if err != nil {
			t.Fatal(err)
		}
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf) // read once from the server
	if err != nil {
		t.Fatal(err)
	}

	expected := bytes.Repeat(msg, 3)
	if !bytes.Equal(expected, buf[:n]) {
		t.Fatalf("expected reply %q; actual reply %q", expected, buf[:n])
	}
}

func datagramEchoServer(ctx context.Context, network string, addr string) (net.Addr, error) {
	s, err := net.ListenPacket(network, addr)
	if err != nil {
		return nil, err
	}

	go func() {
		go func() {
			<-ctx.Done()
			_ = s.Close()
			if network == "unixgram" {
				_ = os.Remove(addr)
			}
		}()

		buf := make([]byte, 1024)
		for {
			n, clientAddr, err := s.ReadFrom(buf)
			if err != nil {
				return
			}

			_, err = s.WriteTo(buf[:n], clientAddr)
			if err != nil {
				return
			}
		}
	}()

	return s.LocalAddr(), nil
}
