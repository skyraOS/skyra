package main

import (
	"bufio"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// Minimal RFC 6455 framing (text/close/ping only, unfragmented), from
// DNA-19's user peer, the predecessor's browser host.
//
// When token is set, the client must present it as "Authorization: Bearer
// TOKEN" (next/23): whoever holds this socket speaks as the human.
func handshake(r *bufio.Reader, conn net.Conn, token string) error {
	var key, auth string
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if i := strings.IndexByte(line, ':'); i > 0 {
			name := strings.ToLower(strings.TrimSpace(line[:i]))
			switch name {
			case "sec-websocket-key":
				key = strings.TrimSpace(line[i+1:])
			case "authorization":
				auth = strings.TrimSpace(line[i+1:])
			}
		}
	}
	if token != "" && subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+token)) != 1 {
		io.WriteString(conn, "HTTP/1.1 401 Unauthorized\r\nContent-Length: 0\r\n\r\n")
		return fmt.Errorf("missing or wrong token")
	}
	if key == "" {
		return fmt.Errorf("missing Sec-WebSocket-Key")
	}
	sum := sha1.Sum([]byte(key + wsGUID))
	accept := base64.StdEncoding.EncodeToString(sum[:])
	_, err := io.WriteString(conn, "HTTP/1.1 101 Switching Protocols\r\n"+
		"Upgrade: websocket\r\n"+
		"Connection: Upgrade\r\n"+
		"Sec-WebSocket-Accept: "+accept+"\r\n\r\n")
	return err
}

func readFrame(r *bufio.Reader) (opcode byte, payload []byte, err error) {
	h := make([]byte, 2)
	if _, err = io.ReadFull(r, h); err != nil {
		return 0, nil, err
	}
	opcode = h[0] & 0x0f
	masked := h[1]&0x80 != 0
	n := uint64(h[1] & 0x7f)
	switch n {
	case 126:
		ext := make([]byte, 2)
		if _, err = io.ReadFull(r, ext); err != nil {
			return 0, nil, err
		}
		n = uint64(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err = io.ReadFull(r, ext); err != nil {
			return 0, nil, err
		}
		n = binary.BigEndian.Uint64(ext)
	}
	var mask []byte
	if masked {
		mask = make([]byte, 4)
		if _, err = io.ReadFull(r, mask); err != nil {
			return 0, nil, err
		}
	}
	payload = make([]byte, n)
	if _, err = io.ReadFull(r, payload); err != nil {
		return 0, nil, err
	}
	if masked {
		for i := range payload {
			payload[i] ^= mask[i%4]
		}
	}
	return opcode, payload, nil
}

// writeFrame writes an unmasked server frame (FIN set, no fragmentation).
func writeFrame(w io.Writer, opcode byte, payload []byte) error {
	header := []byte{0x80 | opcode}
	n := len(payload)
	switch {
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
	if _, err := w.Write(header); err != nil {
		return err
	}
	_, err := w.Write(payload)
	return err
}
