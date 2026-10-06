package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
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
	req := "GET /ws HTTP/1.1\r\nHost: 127.0.0.1\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n"
	for _, h := range headers {
		req += h + "\r\n"
	}
	go io.WriteString(client, req+"\r\n")
	resp, _ := io.ReadAll(client)
	return <-done, string(resp)
}

func TestHandshakeWithoutTokenRefusesClients(t *testing.T) {
	err, resp := shake(t, "")
	if err == nil || !strings.HasPrefix(resp, "HTTP/1.1 401") {
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

func TestHandshakeRejectsForeignOriginAndHost(t *testing.T) {
	for _, headers := range [][]string{{"Origin: https://untrusted.example", "Authorization: Bearer secret"}, {"Host: untrusted.example", "Authorization: Bearer secret"}} {
		err, resp := shake(t, "secret", headers...)
		if err == nil || strings.HasPrefix(resp, "HTTP/1.1 101") {
			t.Fatalf("accepted %v: %s", headers, resp)
		}
	}
}

func TestFrameRejectsMalformedLengthsAndFlags(t *testing.T) {
	for _, frame := range [][]byte{
		{0x81, 0xff, 255, 255, 255, 255, 255, 255, 255, 255},
		{0x81, 0x7f, 255, 255, 255, 255, 255, 255, 255, 255},
		{0x01, 0x80, 0, 0, 0, 0}, {0xc1, 0x80, 0, 0, 0, 0},
		{0x89, 0xfe, 0, 126}, {0x81, 0xfe, 0, 1}, {0x83, 0x80},
	} {
		if _, _, err := readFrame(bufio.NewReader(bytes.NewReader(frame))); err == nil {
			t.Fatalf("accepted invalid frame %x", frame)
		}
	}
	payload := []byte("hello")
	frame := append([]byte{0x81, 0x85, 1, 2, 3, 4}, payload...)
	for i := range payload {
		frame[i+6] ^= frame[2+i%4]
	}
	op, got, err := readFrame(bufio.NewReader(bytes.NewReader(frame)))
	if err != nil || op != 1 || string(got) != "hello" {
		t.Fatalf("valid frame: %d %q %v", op, got, err)
	}
}

func TestBrowserSubprotocolRequiresAuthorizedOrigin(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	done := make(chan error, 1)
	go func() { done <- handshake(bufio.NewReader(server), server, "secret", "http://localhost:3000") }()
	go func() {
		fmt.Fprint(client, "GET /ws HTTP/1.1\r\nHost: localhost:8080\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\nOrigin: http://localhost:3000\r\nSec-WebSocket-Protocol: skyra, skyra-token.secret\r\n\r\n")
	}()
	response, err := http.ReadResponse(bufio.NewReader(client), nil)
	if err != nil || response.StatusCode != 101 || response.Header.Get("Sec-WebSocket-Protocol") != "skyra" {
		t.Fatalf("browser handshake: %v %v", response, err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}
