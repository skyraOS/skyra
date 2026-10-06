package main

import (
	"bufio"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
const maxFrame = 16 << 20

// socket speaks the browser protocol's small WebSocket subset. All writes are
// masked, as required of a client. The webhost sends unfragmented text frames.
type socket struct {
	conn net.Conn
	r    *bufio.Reader
	mu   sync.Mutex
}

func connect(addr string) (*socket, error) {
	c, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		return nil, err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		c.Close()
		return nil, err
	}
	key := base64.StdEncoding.EncodeToString(nonce[:])
	if _, err = fmt.Fprintf(c, "GET /ws HTTP/1.1\r\nHost: %s\r\nUpgrade: websocket\r\nConnection: Upgrade\r\nSec-WebSocket-Version: 13\r\nSec-WebSocket-Key: %s\r\nAuthorization: Bearer %s\r\n\r\n", addr, key, os.Getenv("SKYRA_WS_TOKEN")); err != nil {
		c.Close()
		return nil, err
	}
	r := bufio.NewReader(c)
	status, err := r.ReadString('\n')
	if err != nil || !strings.Contains(status, " 101 ") {
		c.Close()
		return nil, fmt.Errorf("webhost did not upgrade WebSocket: %s", strings.TrimSpace(status))
	}
	var accept string
	for {
		line, e := r.ReadString('\n')
		if e != nil {
			c.Close()
			return nil, e
		}
		if line == "\r\n" {
			break
		}
		if k, v, ok := strings.Cut(line, ":"); ok && strings.EqualFold(k, "Sec-WebSocket-Accept") {
			accept = strings.TrimSpace(v)
		}
	}
	sum := sha1.Sum([]byte(key + wsGUID))
	if accept != base64.StdEncoding.EncodeToString(sum[:]) {
		c.Close()
		return nil, errors.New("webhost sent an invalid WebSocket accept key")
	}
	return &socket{conn: c, r: r}, nil
}

func (s *socket) close() { _ = s.conn.Close() }

func (s *socket) send(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.writeFrame(1, b)
}

func (s *socket) writeFrame(op byte, payload []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(payload) > maxFrame {
		return errors.New("WebSocket frame is too large")
	}
	var mask [4]byte
	if _, err := rand.Read(mask[:]); err != nil {
		return err
	}
	header := []byte{0x80 | op}
	switch n := len(payload); {
	case n < 126:
		header = append(header, 0x80|byte(n))
	case n < 65536:
		header = append(header, 0x80|126, byte(n>>8), byte(n))
	default:
		header = append(header, 0x80|127)
		var ext [8]byte
		binary.BigEndian.PutUint64(ext[:], uint64(n))
		header = append(header, ext[:]...)
	}
	header = append(header, mask[:]...)
	for i := range payload {
		payload[i] ^= mask[i%4]
	}
	frame := append(header, payload...)
	for len(frame) > 0 {
		n, err := s.conn.Write(frame)
		if err != nil {
			return err
		}
		if n == 0 {
			return io.ErrShortWrite
		}
		frame = frame[n:]
	}
	return nil
}

func (s *socket) read() ([]byte, error) {
	for {
		var header [2]byte
		if _, err := io.ReadFull(s.r, header[:]); err != nil {
			return nil, err
		}
		if header[0]&0x80 == 0 {
			return nil, errors.New("fragmented WebSocket frame")
		}
		op := header[0] & 0x0f
		n := uint64(header[1] & 0x7f)
		switch n {
		case 126:
			var ext [2]byte
			if _, err := io.ReadFull(s.r, ext[:]); err != nil {
				return nil, err
			}
			n = uint64(binary.BigEndian.Uint16(ext[:]))
		case 127:
			var ext [8]byte
			if _, err := io.ReadFull(s.r, ext[:]); err != nil {
				return nil, err
			}
			n = binary.BigEndian.Uint64(ext[:])
		}
		if n > maxFrame {
			return nil, errors.New("WebSocket frame is too large")
		}
		var mask [4]byte
		if header[1]&0x80 != 0 {
			if _, err := io.ReadFull(s.r, mask[:]); err != nil {
				return nil, err
			}
		}
		body := make([]byte, int(n))
		if _, err := io.ReadFull(s.r, body); err != nil {
			return nil, err
		}
		if header[1]&0x80 != 0 {
			for i := range body {
				body[i] ^= mask[i%4]
			}
		}
		switch op {
		case 1:
			return body, nil
		case 8:
			return nil, io.EOF
		case 9:
			if err := s.writeFrame(10, body); err != nil {
				return nil, err
			}
		case 10:
		default:
			return nil, fmt.Errorf("unsupported WebSocket opcode %d", op)
		}
	}
}
