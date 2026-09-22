package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
)

// browser is the private browser transport: a minimal RFC 6455 WebSocket server
// over a loopback TCP listener. It is plumbing, exactly as Codex's provider
// connection is private to that executable — a network port number is not a
// Skyra port or executable. One authorized session at a time for this proof.
type browser struct {
	listener net.Listener
	inbound  chan []byte
	connects chan struct{}

	mu   sync.Mutex
	conn net.Conn
}

const wsGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

func newBrowser() *browser {
	return &browser{inbound: make(chan []byte, 16), connects: make(chan struct{}, 1)}
}

// start binds addr (":0" for an ephemeral port) and returns the actual address.
func (b *browser) start(addr string) (string, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return "", err
	}
	b.listener = l
	go b.accept()
	return l.Addr().String(), nil
}

func (b *browser) accept() {
	for {
		conn, err := b.listener.Accept()
		if err != nil {
			return
		}
		go b.handle(conn)
	}
}

func (b *browser) handle(conn net.Conn) {
	r := bufio.NewReader(conn)
	if err := handshake(r, conn); err != nil {
		_ = conn.Close()
		return
	}
	b.mu.Lock()
	if b.conn != nil {
		_ = b.conn.Close()
	}
	b.conn = conn
	b.mu.Unlock()
	select {
	case b.connects <- struct{}{}:
	default:
	}
	for {
		opcode, payload, err := readFrame(r)
		if err != nil {
			b.drop(conn)
			return
		}
		switch opcode {
		case 0x1, 0x2: // text / binary — both carry a JSON message here
			select {
			case b.inbound <- payload:
			default:
			}
		case 0x8: // close
			b.drop(conn)
			return
		case 0x9: // ping -> pong
			b.write(conn, 0xA, payload)
		}
	}
}

func (b *browser) drop(conn net.Conn) {
	b.mu.Lock()
	if b.conn == conn {
		b.conn = nil
	}
	b.mu.Unlock()
	_ = conn.Close()
}

// send marshals v and writes it as a text frame to the current session, if any.
func (b *browser) send(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}
	b.mu.Lock()
	conn := b.conn
	b.mu.Unlock()
	if conn != nil {
		b.write(conn, 0x1, data)
	}
}

func (b *browser) write(conn net.Conn, opcode byte, payload []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	_ = writeFrame(conn, opcode, payload)
}

func (b *browser) close() {
	if b.listener != nil {
		_ = b.listener.Close()
	}
	b.mu.Lock()
	if b.conn != nil {
		_ = b.conn.Close()
		b.conn = nil
	}
	b.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Minimal RFC 6455 framing (text/close/ping only, unfragmented).
// ---------------------------------------------------------------------------

func handshake(r *bufio.Reader, conn net.Conn) error {
	var key string
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
			if name == "sec-websocket-key" {
				key = strings.TrimSpace(line[i+1:])
			}
		}
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
