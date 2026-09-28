package main

import (
	"bufio"
	"io"
	"net"
	"strings"
	"testing"
)

// shake runs the server side of a handshake against a request with the given
// extra header lines, and returns the handshake's error and the response.
func shake(t *testing.T, token string, headers ...string) (error, string) {
	t.Helper()
	server, client := net.Pipe()
	defer client.Close()
	done := make(chan error, 1)
	go func() {
		done <- handshake(bufio.NewReader(server), server, token)
		server.Close()
	}()
	req := "GET /ws HTTP/1.1\r\nHost: 127.0.0.1\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n"
	for _, h := range headers {
		req += h + "\r\n"
	}
	go io.WriteString(client, req+"\r\n")
	resp, _ := io.ReadAll(client)
	return <-done, string(resp)
}

func TestHandshakeWithoutTokenAcceptsAnyClient(t *testing.T) {
	err, resp := shake(t, "")
	if err != nil || !strings.HasPrefix(resp, "HTTP/1.1 101") {
		t.Fatalf("err %v, response %q", err, resp)
	}
}

func TestHandshakeRequiresTheToken(t *testing.T) {
	for _, headers := range [][]string{nil, {"Authorization: Bearer wrong"}, {"Authorization: secret"}} {
		err, resp := shake(t, "secret", headers...)
		if err == nil || !strings.HasPrefix(resp, "HTTP/1.1 401") {
			t.Errorf("headers %q: err %v, response %q", headers, err, resp)
		}
	}
	err, resp := shake(t, "secret", "authorization: Bearer secret")
	if err != nil || !strings.HasPrefix(resp, "HTTP/1.1 101") {
		t.Fatalf("right token: err %v, response %q", err, resp)
	}
}
