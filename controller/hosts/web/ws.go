package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"unicode/utf8"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
const maxFrame = 16 << 20
const maxHandshake = 16 << 10

// Native clients use Bearer auth. Browsers carry the session token in a
// subprotocol and must match the launcher's explicit UI origin. No token is
// ever accepted from a URL query or an unrelated browser origin.
func handshake(r *bufio.Reader, conn net.Conn, token string, origins ...string) error {
	reject := func(status int, reason string) error {
		fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: 0\r\nConnection: close\r\n\r\n", status, http.StatusText(status))
		return fmt.Errorf("%s", reason)
	}
	var headers bytes.Buffer
	for {
		line, err := r.ReadSlice('\n')
		if err != nil || headers.Len()+len(line) > maxHandshake {
			return reject(431, "handshake headers exceed limit or are incomplete")
		}
		headers.Write(line)
		if bytes.Equal(line, []byte("\r\n")) {
			break
		}
	}
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(headers.Bytes())))
	if err != nil {
		return reject(400, "invalid HTTP upgrade")
	}
	defer req.Body.Close()
	host := req.Host
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	ip := net.ParseIP(host)
	if host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return reject(403, "Host must be loopback")
	}
	if req.Method != "GET" || req.URL.Path != "/ws" || req.URL.RawQuery != "" || req.Header.Get("Sec-WebSocket-Version") != "13" || !strings.EqualFold(req.Header.Get("Upgrade"), "websocket") || !headerToken(req.Header.Get("Connection"), "upgrade") {
		return reject(400, "invalid WebSocket upgrade")
	}
	origin := req.Header.Get("Origin")
	if origin != "" {
		allowed := false
		for _, o := range origins {
			if o != "" && origin == o {
				allowed = true
			}
		}
		if !allowed {
			return reject(403, "browser origin is not authorized")
		}
	}
	auth := req.Header.Get("Authorization")
	browserProtocol := false
	for _, protocol := range strings.Split(req.Header.Get("Sec-WebSocket-Protocol"), ",") {
		protocol = strings.TrimSpace(protocol)
		if origin != "" && strings.HasPrefix(protocol, "skyra-token.") {
			auth = "Bearer " + strings.TrimPrefix(protocol, "skyra-token.")
		}
		if protocol == "skyra" {
			browserProtocol = true
		}
	}
	if token == "" || subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+token)) != 1 {
		return reject(401, "missing or wrong token")
	}
	key := req.Header.Get("Sec-WebSocket-Key")
	decoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil || len(decoded) != 16 {
		return reject(400, "invalid Sec-WebSocket-Key")
	}
	sum := sha1.Sum([]byte(key + wsGUID))
	response := "HTTP/1.1 101 Switching Protocols\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Accept: " + base64.StdEncoding.EncodeToString(sum[:]) + "\r\n"
	if browserProtocol {
		response += "Sec-WebSocket-Protocol: skyra\r\n"
	}
	_, err = io.WriteString(conn, response+"\r\n")
	return err
}

func headerToken(header, token string) bool {
	for _, v := range strings.Split(header, ",") {
		if strings.EqualFold(strings.TrimSpace(v), token) {
			return true
		}
	}
	return false
}

// Client frames are masked, bounded and unfragmented. Invalid framing is
// rejected before allocation so malformed traffic cannot panic the host.
func readFrame(r *bufio.Reader) (opcode byte, payload []byte, err error) {
	var h [2]byte
	if _, err = io.ReadFull(r, h[:]); err != nil {
		return
	}
	opcode = h[0] & 0x0f
	if h[0]&0x80 == 0 || h[0]&0x70 != 0 || h[1]&0x80 == 0 {
		return 0, nil, fmt.Errorf("invalid client frame flags")
	}
	switch opcode {
	case 1, 2, 8, 9, 10:
	default:
		return 0, nil, fmt.Errorf("unsupported opcode")
	}
	n := uint64(h[1] & 0x7f)
	marker := n
	switch marker {
	case 126:
		var ext [2]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return
		}
		n = uint64(binary.BigEndian.Uint16(ext[:]))
		if n < 126 {
			return 0, nil, fmt.Errorf("noncanonical frame length")
		}
	case 127:
		var ext [8]byte
		if _, err = io.ReadFull(r, ext[:]); err != nil {
			return
		}
		n = binary.BigEndian.Uint64(ext[:])
		if n < 65536 {
			return 0, nil, fmt.Errorf("noncanonical frame length")
		}
	}
	if n > maxFrame || (opcode >= 8 && n > 125) {
		return 0, nil, fmt.Errorf("frame exceeds limit")
	}
	var mask [4]byte
	if _, err = io.ReadFull(r, mask[:]); err != nil {
		return
	}
	payload = make([]byte, int(n))
	if _, err = io.ReadFull(r, payload); err != nil {
		return
	}
	for i := range payload {
		payload[i] ^= mask[i%4]
	}
	if (opcode == 1 && !utf8.Valid(payload)) || (opcode == 8 && (len(payload) == 1 || len(payload) > 2 && !utf8.Valid(payload[2:]))) {
		return 0, nil, fmt.Errorf("invalid frame text")
	}
	return
}

func writeFrame(w io.Writer, opcode byte, payload []byte) error {
	if len(payload) > maxFrame {
		return fmt.Errorf("frame exceeds limit")
	}
	header := []byte{0x80 | opcode}
	switch n := len(payload); {
	case n < 126:
		header = append(header, byte(n))
	case n < 65536:
		header = append(header, 126, byte(n>>8), byte(n))
	default:
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(n))
		header = append(header, 127)
		header = append(header, ext...)
	}
	_, err := io.Copy(w, io.MultiReader(bytes.NewReader(header), bytes.NewReader(payload)))
	return err
}
