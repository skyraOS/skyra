package boot_test

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"skyra/history/dna19/root/boot"
	"skyra/history/dna19/root/trace"
)

// End-to-end proof of the mediated loop with no provider:
//
//	browser (WS) -> user -> root -> echo -> Return -> user -> browser (WS)
//
// It builds the real user and echo binaries, boots them under root, drives a
// real WebSocket session, and asserts the agent answer flows back through the
// host and U's single terminal Return settles on close.
func buildUserFixture(t *testing.T) boot.UserConfig {
	t.Helper()
	dir := t.TempDir()
	c := boot.UserConfig{Binary: filepath.Join(dir, "user"), Echo: filepath.Join(dir, "echo")}
	for _, b := range []struct{ path, pkg string }{{c.Binary, "../../user"}, {c.Echo, "../echo"}} {
		cmd := exec.CommandContext(guard(t), "go", "build", "-race", "-o", b.path, b.pkg)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build %s: %v %s", b.pkg, err, out)
		}
	}
	return c
}

func TestUserBrowserAgentLoop(t *testing.T) {
	c := buildUserFixture(t)
	var diag bytes.Buffer
	c.Stderr = &diag
	ctx := guard(t)
	r, err := boot.User(ctx, c, &trace.Trace{})
	must(t, err)
	defer r.Close()

	ws := dialWS(t, r.Addr)
	defer ws.close()

	// Wait for U to open (a snapshot carrying the conversation).
	var incarnation, conversation string
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		m := ws.read(t)
		if m["type"] != "snapshot" {
			continue
		}
		if convs, _ := m["conversations"].([]any); len(convs) > 0 {
			incarnation = m["uiIncarnation"].(string)
			conversation = convs[0].(map[string]any)["conversation"].(string)
			break
		}
	}
	if incarnation == "" {
		t.Fatalf("U never opened; user stderr:\n%s", diag.String())
	}

	// A human message must round-trip through the real echo agent.
	ws.send(t, map[string]any{
		"type": "submit", "uiIncarnation": incarnation,
		"conversation": conversation, "requestId": "r1", "text": "hello world",
	})

	accepted, returned := false, false
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && !(accepted && returned) {
		m := ws.read(t)
		switch m["type"] {
		case "accepted":
			if m["requestId"] == "r1" {
				accepted = true
			}
		case "snapshot":
			convs, _ := m["conversations"].([]any)
			if len(convs) == 0 {
				continue
			}
			cv := convs[0].(map[string]any)
			entries, _ := cv["entries"].([]any)
			for _, e := range entries {
				em := e.(map[string]any)
				if em["role"] == "agent" && em["text"] == "hello world" {
					returned = true
				}
			}
			if returned && cv["phase"] != "ready" {
				t.Fatalf("agent returned but phase is %v", cv["phase"])
			}
		}
	}
	if !accepted {
		t.Fatal("no `accepted` ack for the submit")
	}
	if !returned {
		t.Fatalf("agent answer never routed back; user stderr:\n%s", diag.String())
	}

	// Closing the interaction settles U's single terminal Return to the host.
	ws.send(t, map[string]any{
		"type": "close-interaction", "uiIncarnation": incarnation,
		"conversation": conversation, "requestId": "close",
	})
	term, err := r.AwaitClose(ctx)
	must(t, err)
	if term.Identity != r.InteractID || term.Return == nil || string(term.Return.Observation.Bytes) != "interaction-closed" {
		t.Fatalf("bad terminal U return: %+v", term)
	}
}

// ---------------------------------------------------------------------------
// Minimal RFC 6455 client (text frames only), enough to drive the user binary.
// ---------------------------------------------------------------------------

type wsClient struct {
	conn net.Conn
	r    *bufio.Reader
}

func dialWS(t *testing.T, addr string) *wsClient {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	must(t, err)
	key := base64.StdEncoding.EncodeToString([]byte("0123456789abcdef"))
	_, err = io.WriteString(conn, "GET /ws HTTP/1.1\r\nHost: skyra\r\nUpgrade: websocket\r\n"+
		"Connection: Upgrade\r\nSec-WebSocket-Key: "+key+"\r\nSec-WebSocket-Version: 13\r\n\r\n")
	must(t, err)
	r := bufio.NewReader(conn)
	status, err := r.ReadString('\n')
	must(t, err)
	if !strings.Contains(status, "101") {
		t.Fatalf("no websocket upgrade: %q", status)
	}
	for {
		line, err := r.ReadString('\n')
		must(t, err)
		if strings.TrimRight(line, "\r\n") == "" {
			break
		}
	}
	return &wsClient{conn: conn, r: r}
}

func (c *wsClient) send(t *testing.T, v any) {
	t.Helper()
	data, err := json.Marshal(v)
	must(t, err)
	mask := []byte{0xa, 0xb, 0xc, 0xd}
	header := []byte{0x81}
	n := len(data)
	switch {
	case n < 126:
		header = append(header, byte(n)|0x80)
	case n < 65536:
		header = append(header, 126|0x80, byte(n>>8), byte(n))
	default:
		ext := make([]byte, 8)
		binary.BigEndian.PutUint64(ext, uint64(n))
		header = append(append(header, 127|0x80), ext...)
	}
	header = append(header, mask...)
	masked := make([]byte, n)
	for i := range data {
		masked[i] = data[i] ^ mask[i%4]
	}
	if _, err := c.conn.Write(append(header, masked...)); err != nil {
		t.Fatalf("ws write: %v", err)
	}
}

func (c *wsClient) read(t *testing.T) map[string]any {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	for {
		h := make([]byte, 2)
		if _, err := io.ReadFull(c.r, h); err != nil {
			t.Fatalf("ws read: %v", err)
		}
		opcode := h[0] & 0x0f
		n := uint64(h[1] & 0x7f)
		switch n {
		case 126:
			ext := make([]byte, 2)
			_, err := io.ReadFull(c.r, ext)
			must(t, err)
			n = uint64(binary.BigEndian.Uint16(ext))
		case 127:
			ext := make([]byte, 8)
			_, err := io.ReadFull(c.r, ext)
			must(t, err)
			n = binary.BigEndian.Uint64(ext)
		}
		payload := make([]byte, n)
		if _, err := io.ReadFull(c.r, payload); err != nil {
			t.Fatalf("ws read payload: %v", err)
		}
		if opcode != 0x1 { // ignore non-text control frames
			continue
		}
		var m map[string]any
		if err := json.Unmarshal(payload, &m); err != nil {
			t.Fatalf("ws json: %v (%s)", err, payload)
		}
		return m
	}
}

func (c *wsClient) close() { _ = c.conn.Close() }
