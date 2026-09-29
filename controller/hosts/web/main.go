// Command webhost is the actor host in front of the human (next/16): it
// serves the browser over a loopback WebSocket, in the protocol DNA-19's UI
// speaks, and holds custody of every conversation frame parked with it.
//
//	webhost -ws ADDR -park SOCK -skyra BIN -registry FILE -conversations DIR
//
// With SKYRA_WS_TOKEN set, only a client presenting that token may connect
// (next/23). It is read from the environment so it never shows in ps.
//
// It follows the browser surface's registered procedure and decides nothing:
//
//	a frame is parked      show its transcript; wait for the human
//	the human sends text   resume the frame with {"to":"reply","payload":TEXT}:
//	                       the runtime checks it like any output
//	the human closes       resume it with "/close": the user side ends the conversation
//	the human opens one    seed a new conversation, as the operator
//
// A conversation's history is its lineage. Past conversations are listed by
// deriving their transcript from their records; they are shown as ended,
// because their parked frames died with the host that held them.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"skyra/controller/core/record"
	"skyra/controller/runtime/park"
)

type entry struct {
	Role      string  `json:"role"`
	Text      string  `json:"text"`
	TS        int64   `json:"ts"`
	RequestID *string `json:"requestId"`
	Call      *call   `json:"call,omitempty"` // role "call": a tool call in the agent's turn
}

type conversation struct {
	id      string
	started time.Time
	phase   string // ready | pending | failed | closed
	entries []entry
	parked  *park.Frame
	lineage *lineage // its records, read as they are committed
}

type webhost struct {
	skyra, registry, dir string
	incarnation          string
	token                string // empty: any loopback client may connect
	log                  *os.File

	mu       sync.Mutex
	convos   map[string]*conversation
	revision int
	seen     map[string]bool // request ids already answered
	conn     net.Conn
	wmu      sync.Mutex
}

func main() {
	ws := flag.String("ws", "127.0.0.1:8080", "loopback address for the browser")
	parkSock := flag.String("park", "", "unix socket instances park browser frames on")
	skyra := flag.String("skyra", "", "the skyra binary, to resume parked frames and seed conversations")
	registry := flag.String("registry", "", "the operator's registry, for new conversations")
	dir := flag.String("conversations", "", "directory holding one .skyra per conversation")
	flag.Parse()
	if *parkSock == "" || *skyra == "" || *registry == "" || *dir == "" {
		flag.Usage()
		os.Exit(2)
	}
	h := &webhost{skyra: *skyra, registry: *registry, dir: *dir, convos: map[string]*conversation{}, seen: map[string]bool{},
		incarnation: fmt.Sprintf("webhost-%d-%d", os.Getpid(), time.Now().UnixMilli()), token: os.Getenv("SKYRA_WS_TOKEN")}
	h.log = os.Stderr
	if err := os.MkdirAll(*dir, 0o755); err != nil {
		fail(err)
	}
	h.loadPast()
	go h.watch()

	os.Remove(*parkSock)
	pl, err := net.Listen("unix", *parkSock)
	if err != nil {
		fail(err)
	}
	defer os.Remove(*parkSock)
	go h.acceptParks(pl)

	wl, err := net.Listen("tcp", *ws)
	if err != nil {
		fail(err)
	}
	fmt.Fprintf(os.Stderr, "webhost: browser on ws://%s/ws, parking on %s, %d past conversations\n", *ws, *parkSock, len(h.convos))
	for {
		c, err := wl.Accept()
		if err != nil {
			fail(err)
		}
		go h.serveBrowser(c)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "webhost:", err)
	os.Exit(1)
}

// loadPast derives each past conversation's transcript from its records: the
// latest browser invocation carries the whole transcript it showed.
func (h *webhost) loadPast() {
	dirs, _ := os.ReadDir(h.dir)
	for _, d := range dirs {
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") || h.deleted(d.Name()) {
			continue
		}
		root := filepath.Join(h.dir, d.Name(), ".skyra")
		if _, err := os.Stat(filepath.Join(root, "records")); err != nil {
			continue
		}
		c := &conversation{id: d.Name(), phase: "closed", lineage: newLineage(root)}
		c.lineage.refresh()
		if info, err := d.Info(); err == nil {
			c.started = info.ModTime()
		}
		for _, r := range c.lineage.live() {
			if r.State != record.Invoke || r.To != "browser" {
				continue
			}
			var p struct {
				Transcript []entry `json:"transcript"`
			}
			if json.Unmarshal(r.Payload, &p) == nil && len(p.Transcript) >= len(c.entries) {
				c.entries = stamp(p.Transcript)
			}
		}
		h.convos[c.id] = c
	}
}

func stamp(es []entry) []entry {
	for i := range es {
		es[i].TS = int64(i)
	}
	return es
}

// acceptParks takes custody of frames the runtime parks with the browser.
func (h *webhost) acceptParks(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			return
		}
		go func() {
			defer c.Close()
			uc := c.(*net.UnixConn)
			f, err := park.Receive(uc)
			if err != nil {
				park.Answer(uc, err.Error())
				return
			}
			var p struct {
				Conversation string  `json:"conversation"`
				Transcript   []entry `json:"transcript"`
			}
			if err := json.Unmarshal(f.Payload, &p); err != nil || p.Conversation == "" {
				f.Close()
				park.Answer(uc, "the payload names no conversation")
				return
			}
			h.mu.Lock()
			if !validConversationID(p.Conversation) || h.deleted(p.Conversation) {
				h.mu.Unlock()
				f.Close()
				park.Answer(uc, "the conversation was deleted or is invalid")
				return
			}
			conv := h.convos[p.Conversation]
			if conv == nil {
				conv = &conversation{id: p.Conversation, started: time.Now(), lineage: newLineage(filepath.Join(h.dir, p.Conversation, ".skyra"))}
				h.convos[p.Conversation] = conv
			}
			if conv.parked != nil {
				h.mu.Unlock()
				f.Close()
				park.Answer(uc, "the conversation is already waiting on the human")
				return
			}
			conv.parked, conv.phase, conv.entries = &f, "ready", stamp(p.Transcript)
			l := conv.lineage
			h.mu.Unlock()
			l.refresh()
			if err := park.Answer(uc, ""); err != nil {
				return
			}
			fmt.Fprintf(h.log, "webhost: %s is waiting on the human\n", p.Conversation)
			h.broadcast()
		}()
	}
}

func (h *webhost) serveBrowser(conn net.Conn) {
	r := bufio.NewReader(conn)
	if err := handshake(r, conn, h.token); err != nil {
		conn.Close()
		return
	}
	h.wmu.Lock()
	if h.conn != nil {
		h.conn.Close()
	}
	h.conn = conn
	h.wmu.Unlock()
	h.broadcast()
	for {
		op, payload, err := readFrame(r)
		if err != nil || op == 0x8 {
			conn.Close()
			return
		}
		if op == 0x9 {
			h.wmu.Lock()
			writeFrame(conn, 0xA, payload)
			h.wmu.Unlock()
			continue
		}
		if op == 0x1 || op == 0x2 {
			h.handle(payload)
		}
	}
}

type inbound struct {
	Type          string `json:"type"`
	UIIncarnation string `json:"uiIncarnation"`
	Conversation  string `json:"conversation"`
	RequestID     string `json:"requestId"`
	Text          string `json:"text"`
}

func (h *webhost) handle(raw []byte) {
	var m inbound
	if err := json.Unmarshal(raw, &m); err != nil || m.RequestID == "" {
		return
	}
	reject := func(code, msg string) {
		h.send(map[string]string{"type": "rejected", "requestId": m.RequestID, "code": code, "message": msg})
	}
	h.mu.Lock()
	if h.seen[m.RequestID] {
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		return
	}
	if m.Type != "open-conversation" && m.UIIncarnation != h.incarnation {
		h.mu.Unlock()
		reject("stale", "the page is out of date; it will refresh")
		h.broadcast()
		return
	}
	switch m.Type {
	case "delete-conversation":
		if err := h.deleteConversation(m.Conversation); err != nil {
			h.mu.Unlock()
			reject("delete-failed", err.Error())
			return
		}
		h.seen[m.RequestID] = true
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
	case "open-conversation":
		id := fmt.Sprintf("c-%s", time.Now().Format("20060102-150405.000"))
		h.convos[id] = &conversation{id: id, started: time.Now(), phase: "pending", lineage: newLineage(filepath.Join(h.dir, id, ".skyra"))}
		h.seen[m.RequestID] = true
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
		go h.seed(id)
	case "submit", "close-interaction":
		conv := h.convos[m.Conversation]
		if conv == nil || conv.parked == nil || conv.phase != "ready" {
			h.mu.Unlock()
			reject("busy", "this conversation is not waiting for you")
			return
		}
		text := m.Text
		if m.Type == "close-interaction" {
			text = "/close"
			conv.phase = "closed"
		} else {
			id := m.RequestID
			conv.entries = append(conv.entries, entry{Role: "user", Text: text, TS: int64(len(conv.entries)), RequestID: &id})
			conv.phase = "pending"
		}
		f := *conv.parked
		conv.parked = nil
		h.seen[m.RequestID] = true
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
		// The event goes back in as a reply: whatever the human typed is only text.
		event, _ := json.Marshal(map[string]string{"to": "reply", "payload": text})
		if err := park.Resume(h.skyra, f, event, h.log); err != nil {
			h.mu.Lock()
			conv.phase = "failed"
			conv.entries = append(conv.entries, entry{Role: "agent", Text: "⚠ could not resume: " + err.Error(), TS: int64(len(conv.entries))})
			h.mu.Unlock()
			h.broadcast()
		}
	default:
		h.mu.Unlock()
		reject("unknown", "unknown message type")
	}
}

func validConversationID(id string) bool {
	return id != "" && !strings.HasPrefix(id, ".") && filepath.Base(id) == id && !strings.ContainsAny(id, `/\`)
}

func (h *webhost) deleted(id string) bool {
	_, err := os.Stat(filepath.Join(h.dir, ".deleted", id))
	return err == nil
}

// Caller holds mu. Keep a durable tombstone so late parks or writes cannot
// resurrect a deleted conversation. Provider-side retention is separate.
func (h *webhost) deleteConversation(id string) error {
	if !validConversationID(id) {
		return fmt.Errorf("invalid conversation")
	}
	c := h.convos[id]
	if c == nil {
		if h.deleted(id) {
			return nil
		}
		return fmt.Errorf("conversation not found")
	}
	if c.phase == "pending" {
		return fmt.Errorf("wait for the agent to finish before deleting this chat")
	}
	markers := filepath.Join(h.dir, ".deleted")
	if err := os.MkdirAll(markers, 0o700); err != nil {
		return err
	}
	marker := filepath.Join(markers, id)
	if err := os.WriteFile(marker, nil, 0o600); err != nil {
		return err
	}
	if err := os.RemoveAll(filepath.Join(h.dir, id)); err != nil {
		os.Remove(marker)
		return fmt.Errorf("could not remove chat history: %w", err)
	}
	if c.parked != nil {
		c.parked.Close()
	}
	delete(h.convos, id)
	return nil
}

// seed starts a new conversation as the operator: genesis, then the user side.
func (h *webhost) seed(id string) {
	root := filepath.Join(h.dir, id, ".skyra")
	payload, _ := json.Marshal(map[string]string{"conversation": id})
	cmd := exec.Command(h.skyra, "boot", "-root", root, "-registry", h.registry, "-to", "user", "-payload", string(payload), "-detach")
	cmd.Stdout, cmd.Stderr = h.log, h.log
	if err := cmd.Run(); err != nil {
		h.mu.Lock()
		if c := h.convos[id]; c != nil && c.parked == nil {
			c.phase = "failed"
			c.entries = append(c.entries, entry{Role: "agent", Text: "⚠ could not start the conversation: " + err.Error()})
		}
		h.mu.Unlock()
		h.broadcast()
	}
}

// watch reads the lineage of every conversation whose agent is working, so a
// tool call shows while it runs rather than when the turn ends.
func (h *webhost) watch() {
	for range time.Tick(200 * time.Millisecond) {
		h.mu.Lock()
		var working []*lineage
		for _, c := range h.convos {
			if c.phase == "pending" && c.lineage != nil {
				working = append(working, c.lineage)
			}
		}
		h.mu.Unlock()
		changed := false
		for _, l := range working {
			if l.refresh() {
				changed = true
			}
		}
		if changed {
			h.broadcast()
		}
	}
}

func (h *webhost) broadcast() {
	h.mu.Lock()
	h.revision++
	var list []*conversation
	for _, c := range h.convos {
		list = append(list, c)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].started.After(list[j].started) })
	views := []map[string]any{}
	for _, c := range list {
		entries := withCalls(append([]entry(nil), c.entries...), c.lineage.snapshot())
		views = append(views, map[string]any{"conversation": c.id, "phase": c.phase, "entries": entries, "pendingRequestId": nil})
	}
	snap := map[string]any{"type": "snapshot", "uiIncarnation": h.incarnation, "revision": h.revision, "conversations": views, "requests": []any{}}
	h.mu.Unlock()
	h.send(snap)
}

func (h *webhost) send(v any) {
	b, _ := json.Marshal(v)
	h.wmu.Lock()
	defer h.wmu.Unlock()
	if h.conn != nil {
		writeFrame(h.conn, 0x1, b)
	}
}
