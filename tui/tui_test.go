package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"testing"
	"time"
)

// Exercise the wire boundary with a real TCP listener: the client must verify
// the upgrade, read a snapshot, and mask its outbound browser message.
func TestWebSocketConversationExchange(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	done := make(chan error, 1)
	go func() {
		c, err := l.Accept()
		if err != nil {
			done <- err
			return
		}
		defer c.Close()
		r := bufio.NewReader(c)
		var key string
		for {
			line, e := r.ReadString('\n')
			if e != nil {
				done <- e
				return
			}
			if line == "\r\n" {
				break
			}
			if strings.HasPrefix(strings.ToLower(line), "sec-websocket-key:") {
				key = strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			}
		}
		sum := sha1.Sum([]byte(key + wsGUID))
		accept := base64.StdEncoding.EncodeToString(sum[:])
		if _, e := fmt.Fprintf(c, "HTTP/1.1 101 Switching Protocols\r\nSec-WebSocket-Accept: %s\r\n\r\n", accept); e != nil {
			done <- e
			return
		}
		snapshot := []byte(`{"type":"snapshot","uiIncarnation":"host-1","revision":1,"conversations":[]}`)
		if _, e := c.Write(append([]byte{0x81, byte(len(snapshot))}, snapshot...)); e != nil {
			done <- e
			return
		}
		var head [2]byte
		if _, e := io.ReadFull(r, head[:]); e != nil {
			done <- e
			return
		}
		if head[0] != 0x81 || head[1]&0x80 == 0 {
			done <- fmt.Errorf("client did not mask a text frame: %x", head)
			return
		}
		n := int(head[1] & 0x7f)
		var mask [4]byte
		if _, e := io.ReadFull(r, mask[:]); e != nil {
			done <- e
			return
		}
		body := make([]byte, n)
		if _, e := io.ReadFull(r, body); e != nil {
			done <- e
			return
		}
		for i := range body {
			body[i] ^= mask[i%4]
		}
		var msg map[string]string
		if e := json.Unmarshal(body, &msg); e != nil {
			done <- e
			return
		}
		if msg["type"] != "open-conversation" || msg["requestId"] != "req-1" {
			done <- fmt.Errorf("unexpected request: %s", body)
			return
		}
		done <- nil
	}()
	s, err := connect(l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer s.close()
	b, err := s.read()
	if err != nil {
		t.Fatal(err)
	}
	var m message
	if err := json.Unmarshal(b, &m); err != nil || m.Type != "snapshot" || m.UIIncarnation != "host-1" {
		t.Fatalf("snapshot: %s, %v", b, err)
	}
	if err := s.send(map[string]string{"type": "open-conversation", "requestId": "req-1"}); err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

func TestHistoryDoesNotBecomeCurrentSession(t *testing.T) {
	a := &app{revision: -1}
	a.receive([]byte(`{"type":"snapshot","uiIncarnation":"host-1","revision":1,"conversations":[{"conversation":"c-1","phase":"pending","entries":[]}]}`))
	if a.selected != "" || a.current() != nil {
		t.Fatalf("old history became current: %+v", a)
	}
	a.selected = "c-1"
	a.input = []rune("hello")
	a.key("\r")
	if string(a.input) != "hello" {
		t.Fatal("pending session consumed an unsent message")
	}
	a.receive([]byte(`{"type":"snapshot","uiIncarnation":"host-1","revision":0,"conversations":[]}`))
	if a.current() == nil {
		t.Fatal("stale snapshot replaced a newer one")
	}
	a.key("\x0e")
	if a.selected != "" || a.current() != nil || string(a.input) != "hello" {
		t.Fatalf("new session did not clear the view and retain the draft: %+v", a)
	}
}

func TestFirstMessageOpensFreshSessionAndSends(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()
	_ = client.SetDeadline(time.Now().Add(3 * time.Second))
	_ = server.SetDeadline(time.Now().Add(3 * time.Second))
	requests := make(chan []map[string]string, 1)
	go func() {
		peer := &socket{conn: server, r: bufio.NewReader(server)}
		var received []map[string]string
		for range 2 {
			body, err := peer.read()
			if err != nil {
				requests <- nil
				return
			}
			var msg map[string]string
			if json.Unmarshal(body, &msg) != nil {
				requests <- nil
				return
			}
			received = append(received, msg)
		}
		requests <- received
	}()
	a := &app{sock: &socket{conn: client}, connected: true, incarnation: "host-1", revision: 1,
		conversations: []conversation{{ID: "old", Phase: "closed"}}, input: []rune("hello")}
	a.key("\r")
	if !a.opening || a.pendingText != "hello" || len(a.input) != 0 {
		t.Fatalf("draft was not queued: %+v", a)
	}
	a.receive([]byte(`{"type":"snapshot","uiIncarnation":"host-1","revision":2,"conversations":[{"conversation":"new","phase":"pending","entries":[]},{"conversation":"old","phase":"closed","entries":[]}]}`))
	if a.selected != "new" || !a.opening {
		t.Fatalf("new chat was not selected: %+v", a)
	}
	a.receive([]byte(`{"type":"snapshot","uiIncarnation":"host-1","revision":3,"conversations":[{"conversation":"new","phase":"ready","entries":[]},{"conversation":"old","phase":"closed","entries":[]}]}`))
	got := <-requests
	if len(got) != 2 || got[0]["type"] != "open-conversation" || got[1]["type"] != "submit" || got[1]["conversation"] != "new" || got[1]["text"] != "hello" {
		t.Fatalf("wrong request sequence: %#v", got)
	}
	if a.opening || a.pendingText != "" || a.notice != "" {
		t.Fatalf("draft or opening status not settled: %+v", a)
	}
}

func TestCtrlCClearsTheComposerAndTwiceQuits(t *testing.T) {
	a := app{input: []rune("half a thought"), conversations: []conversation{{ID: "c", Phase: "ready", Entries: []entry{{Role: "user", Text: "hi"}}}}, selected: "c"}
	if a.key("\x03") || len(a.input) != 0 {
		t.Fatal("ctrl+c did not clear the composer, or quit")
	}
	if len(a.current().Entries) != 1 {
		t.Fatal("ctrl+c touched the transcript")
	}
	if a.key("\x03") {
		t.Fatal("one ctrl+c on an empty composer quit")
	}
	a.key("x")
	if a.quitArmed {
		t.Fatal("typing did not disarm quit")
	}
	if a.key("\x03") || len(a.input) != 0 {
		t.Fatal("ctrl+c did not clear the draft")
	}
	if a.key("\x03") {
		t.Fatal("quit armed too early")
	}
	if !a.key("\x03") {
		t.Fatal("second ctrl+c on an empty composer did not quit")
	}
}

func TestTitleRecedesInAConversationAndReturnsWhenEmpty(t *testing.T) {
	a := app{width: 120, height: 35}
	for i := 0; i < 25; i++ {
		before := a.titleOpen
		a.stepSky()
		if a.titleOpen-before > titleFade+1e-9 {
			t.Fatal("the title opened in a jump")
		}
	}
	if a.titleOpen != 1 {
		t.Fatalf("an empty session left the title at %.2f", a.titleOpen)
	}
	a.conversations = []conversation{{ID: "c", Phase: "ready", Entries: []entry{{Role: "user", Text: "hi"}}}}
	a.selected = "c"
	for i := 0; i < 25; i++ {
		a.stepSky()
	}
	if a.titleOpen != 0 {
		t.Fatalf("a conversation left the title at %.2f", a.titleOpen)
	}
}

func TestHeaderFloatsOnTheSky(t *testing.T) {
	a := app{width: 120, height: 35}
	screen := a.compose()
	sky := newCanvas(120, 35, 0, nil, nil).pixels()
	for x := 20; x < 80; x++ { // between the title and the status
		if screen[0][x].bg != sky[0][x].bg {
			t.Fatalf("the header row is covered at column %d", x)
		}
	}
}

func TestLightBrightensTheTitle(t *testing.T) {
	dark, _ := titleColors(0, 0)
	lit, _ := titleColors(0, 1)
	if dark == lit {
		t.Fatal("light did not change the title")
	}
	_, receded := titleColors(0, 0)
	_, open := titleColors(1, 0)
	if receded == open {
		t.Fatal("opening did not change the title")
	}
}
