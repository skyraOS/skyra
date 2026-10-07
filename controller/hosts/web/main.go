// Command webhost is the actor host in front of the human (next/16): it
// serves the browser over a loopback WebSocket, in the protocol DNA-19's UI
// speaks, and holds custody of every conversation frame parked with it.
//
//	webhost -ws ADDR -park SOCK -skyra BIN -registry FILE -conversations DIR
//
// SKYRA_WS_TOKEN is required; only a client presenting that token may connect
// (next/23). It is read from the environment so it never shows in ps.
//
// It follows the browser surface's registered procedure and decides nothing:
//
//	a frame is parked      show its transcript; wait for the human
//	the human sends text   resume the frame with {"to":<its slot>,"payload":TEXT}:
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
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"skyra/controller/core/record"
	"skyra/controller/runtime/boot"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/tree"

	"skyra/controller/runtime/custody"
)

type entry struct {
	Role      string  `json:"role"`
	Text      string  `json:"text"`
	TS        int64   `json:"ts"`
	RequestID *string `json:"requestId"`
	Call      *call   `json:"call,omitempty"` // role "call": a tool call in the agent's turn
}

type conversation struct {
	id               string
	started          time.Time
	phase            string // ready | pending | failed | closed
	entries          []entry
	parked           *park.Frame
	pending          string
	procedure        string // the root it runs, fixed at creation; "" for a conversation from before procedures
	home             string // its directory: under its procedure's, or at the top for one from before procedures
	recoveryRequired bool
	lineage          *lineage // its records, read as they are committed
	pastCalls        [][]call // completed runs, indexed by conversation turn
}

type webhost struct {
	skyra, registry, dir string
	scratch              string // operator-selected shared write space; empty means per-conversation
	incarnation          string
	token                string
	origin               string
	log                  *os.File

	mu           sync.Mutex
	convos       map[string]*conversation
	revision     int
	requests     map[string]*requestRecord
	destinations []destination // the roots a human may select, from the registry's bundles
	conn         net.Conn
	wmu          sync.Mutex
}

func main() {
	custody.SealInherited() // nothing the launcher leaked reaches a resumed instance
	ws := flag.String("ws", "127.0.0.1:8080", "loopback address for the browser")
	parkSock := flag.String("park", "", "unix socket instances park browser frames on")
	skyra := flag.String("skyra", "", "the skyra binary, to resume parked frames and seed conversations")
	registry := flag.String("registry", "", "the operator's registry, for new conversations")
	dir := flag.String("conversations", "", "directory holding one .skyra per conversation")
	scratch := flag.String("scratch", "", "shared write space for new runs; default is conversation-local scratch")
	flag.Parse()
	if *parkSock == "" || *skyra == "" || *registry == "" || *dir == "" {
		flag.Usage()
		os.Exit(2)
	}
	h := &webhost{skyra: *skyra, registry: *registry, dir: *dir, convos: map[string]*conversation{},
		scratch:     *scratch,
		incarnation: fmt.Sprintf("webhost-%d-%d", os.Getpid(), time.Now().UnixMilli()), token: os.Getenv("SKYRA_WS_TOKEN"), origin: os.Getenv("SKYRA_WS_ORIGIN")}
	if h.token == "" {
		fail(fmt.Errorf("SKYRA_WS_TOKEN is required"))
	}
	h.log = os.Stderr
	if err := os.MkdirAll(*dir, 0o700); err != nil {
		fail(err)
	}
	if err := h.loadRequests(); err != nil {
		fail(err)
	}
	destinations, err := loadDestinations(h.registry)
	if err != nil {
		fail(err)
	}
	h.destinations = destinations
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
// latest browser invocation carries the whole transcript it showed. A
// conversation lives under its procedure's directory, kept with the surface
// it runs (proposal 38); one from before procedures stays at the top.
func (h *webhost) loadPast() {
	dirs, _ := os.ReadDir(h.dir)
	for _, d := range dirs {
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			continue
		}
		if strings.HasPrefix(d.Name(), "c-") {
			h.loadConversation(filepath.Join(h.dir, d.Name()), "")
			continue
		}
		inside, _ := os.ReadDir(filepath.Join(h.dir, d.Name()))
		for _, c := range inside {
			if c.IsDir() && strings.HasPrefix(c.Name(), "c-") {
				h.loadConversation(filepath.Join(h.dir, d.Name(), c.Name()), d.Name())
			}
		}
	}
}

// loadConversation reads one conversation at home. Its procedure is what its
// genesis recorded; the directory it sits in is only where it is kept.
func (h *webhost) loadConversation(home, under string) {
	id := filepath.Base(home)
	if h.deleted(id) {
		return
	}
	roots := conversationRuns(home)
	root := roots[len(roots)-1]
	info, err := os.Stat(filepath.Join(root, "records"))
	if err != nil {
		return
	}
	c := &conversation{id: id, phase: "closed", home: home, procedure: under, lineage: newLineage(root), started: info.ModTime()}
	for _, prior := range roots[:len(roots)-1] {
		l := newLineage(prior)
		l.refresh()
		c.pastCalls = mergeCalls(c.pastCalls, l.snapshot())
	}
	c.lineage.refresh()
	if info, err := os.Stat(home); err == nil {
		c.started = info.ModTime()
	}
	for _, r := range c.lineage.live() {
		if r.State == record.Invoke && r.To == "user" {
			var seeded struct {
				Procedure string `json:"procedure"`
			}
			if json.Unmarshal(r.Payload, &seeded) == nil && seeded.Procedure != "" {
				if under != "" && seeded.Procedure != under {
					fmt.Fprintf(h.log, "webhost: %s is kept under %s but runs %s\n", id, under, seeded.Procedure)
				}
				c.procedure = seeded.Procedure // recorded at creation: the lineage says what it runs
			}
		}
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

// homeFor is where a new conversation that runs procedure is kept: with the
// surface it runs.
func (h *webhost) homeFor(procedure, id string) string {
	if procedure == "" {
		return filepath.Join(h.dir, id)
	}
	return filepath.Join(h.dir, procedure, id)
}

// homeOf is a conversation's directory: where it was opened or loaded, or
// found on disk.
func (h *webhost) homeOf(c *conversation) string {
	if c.home != "" {
		return c.home
	}
	return h.locate(c.id)
}

// locate finds a conversation's directory on disk when it is not loaded: at
// the top, or under any procedure's directory.
func (h *webhost) locate(id string) string {
	if _, err := os.Stat(filepath.Join(h.dir, id)); err == nil {
		return filepath.Join(h.dir, id)
	}
	dirs, _ := os.ReadDir(h.dir)
	for _, d := range dirs {
		if d.IsDir() && !strings.HasPrefix(d.Name(), ".") && !strings.HasPrefix(d.Name(), "c-") {
			if _, err := os.Stat(filepath.Join(h.dir, d.Name(), id)); err == nil {
				return filepath.Join(h.dir, d.Name(), id)
			}
		}
	}
	return filepath.Join(h.dir, id)
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
				home := h.locate(p.Conversation)
				conv = &conversation{id: p.Conversation, started: time.Now(), home: home, lineage: newLineage(filepath.Join(home, ".skyra"))}
				h.convos[p.Conversation] = conv
			}
			if conv.recoveryRequired || conv.parked != nil {
				h.mu.Unlock()
				f.Close()
				park.Answer(uc, "the conversation is already waiting on the human")
				return
			}
			h.acceptTranscript(conv, p.Transcript)
			conv.parked, conv.phase = &f, "ready"
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
	conn.SetDeadline(time.Now().Add(10 * time.Second))
	r := bufio.NewReader(conn)
	if err := handshake(r, conn, h.token, h.origin); err != nil {
		conn.Close()
		return
	}
	conn.SetDeadline(time.Time{})
	defer conn.Close()
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
			conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
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
	Type          string   `json:"type"`
	UIIncarnation string   `json:"uiIncarnation"`
	Conversation  string   `json:"conversation"`
	RequestID     string   `json:"requestId"`
	Text          string   `json:"text"`
	Procedure     string   `json:"procedure,omitempty"` // open-conversation: the root selected; the first when empty
	Images        []pasted `json:"images,omitempty"`    // submit: images pasted with the text
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
	if m.Type != "open-conversation" && m.UIIncarnation != h.incarnation {
		h.mu.Unlock()
		reject("stale", "the page is out of date; it will refresh")
		h.broadcast()
		return
	}
	if old := h.requests[m.RequestID]; old != nil {
		if old.Fingerprint != fingerprint(m) {
			h.mu.Unlock()
			reject("conflict", "this request id belongs to different work")
			return
		}
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
		return
	}
	admit := func(conversation string) bool {
		if err := h.admit(m, conversation); err != nil {
			h.mu.Unlock()
			reject("storage", err.Error())
			return false
		}
		return true
	}

	switch m.Type {
	case "delete-conversation":
		if !admit(m.Conversation) {
			return
		}
		if err := h.deleteConversation(m.Conversation); err != nil {
			h.finishRequest(m.RequestID, "failed", "delete-failed", err.Error())
			h.mu.Unlock()
			reject("delete-failed", err.Error())
			return
		}
		h.finishRequest(m.RequestID, "returned", "", "")
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
	case "open-conversation":
		d, ok := h.destinationFor(m.Procedure)
		if !ok {
			h.mu.Unlock()
			reject("unknown-procedure", "no procedure "+strconv.Quote(m.Procedure)+" can be selected")
			return
		}
		nonce := make([]byte, 16)
		if _, err := rand.Read(nonce); err != nil {
			h.mu.Unlock()
			reject("entropy", err.Error())
			return
		}
		id := "c-" + hex.EncodeToString(nonce)
		if !admit(id) {
			return
		}
		home := h.homeFor(d.Title, id)
		h.convos[id] = &conversation{id: id, started: time.Now(), phase: "pending", pending: m.RequestID, procedure: d.Title, home: home, lineage: newLineage(filepath.Join(home, ".skyra"))}
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
		go h.seed(id, d.Title, home)
	case "submit", "close-interaction":
		conv := h.convos[m.Conversation]
		if m.Type == "submit" && conv != nil && conv.parked == nil && conv.pending == "" && (conv.phase == "closed" || conv.phase == "failed") {
			h.continueConversation(conv, m, admit, reject)
			return
		}
		if conv == nil || conv.parked == nil || conv.phase != "ready" {
			h.mu.Unlock()
			reject("busy", "this conversation is not waiting for you")
			return
		}
		// Pasted images are checked and kept before admission, so a bad
		// one is refused whole and nothing runs.
		var images []attachment
		if m.Type == "submit" && len(m.Images) > 0 {
			var err error
			if images, err = saveAttachments(h.scratchFor(h.homeOf(conv)), m.Images); err != nil {
				h.mu.Unlock()
				reject("invalid-image", err.Error())
				return
			}
		}
		if !admit(m.Conversation) {
			return
		}
		conv.pending = m.RequestID
		text := m.Text
		if m.Type == "close-interaction" {
			text = "/close"
			conv.phase = "closed"
		} else {
			id := m.RequestID
			conv.entries = append(conv.entries, entry{Role: "user", Text: withImages(text, images), TS: int64(len(conv.entries)), RequestID: &id})
			conv.phase = "pending"
		}
		f := *conv.parked
		conv.parked = nil
		h.mu.Unlock()
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
		h.broadcast()
		// The event goes back in as a return, addressed to the parked frame's
		// slot: whatever the human typed is only text, and pasted images are
		// only their paths and hashes beside it.
		var payload any = text
		if len(images) > 0 {
			payload = map[string]any{"text": text, "images": images}
		}
		event, _ := json.Marshal(map[string]any{"to": f.Stamp, "payload": payload}) // a parked invoke's frame is its own stamp
		if err := park.Resume(h.skyra, f, event, h.log); err != nil {
			h.mu.Lock()
			conv.phase = "failed"
			conv.recoveryRequired = true
			h.finishRequest(m.RequestID, "failed", "delivery-unknown", err.Error())
			conv.pending = ""
			conv.entries = append(conv.entries, entry{Role: "agent", Text: "⚠ could not resume: " + err.Error(), TS: int64(len(conv.entries))})
			h.mu.Unlock()
			h.broadcast()
		}
	default:
		h.mu.Unlock()
		reject("unknown", "unknown message type")
	}
}

// withImages is a message as the transcript shows it: its text, then a line
// naming each image pasted with it.
func withImages(text string, images []attachment) string {
	for _, im := range images {
		text += "\n🖼 " + im.Name
	}
	return text
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
	home := h.homeOf(c)
	// Only ever a conversation's own directory, inside the conversations'.
	if rel, err := filepath.Rel(h.dir, home); err != nil || filepath.Base(home) != id || strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
		os.Remove(marker)
		return fmt.Errorf("conversation %s is not where conversations are kept", id)
	}
	if err := os.RemoveAll(home); err != nil {
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
func (h *webhost) seed(id, procedure, home string) {
	h.seedRun(id, filepath.Join(home, ".skyra"), home, map[string]any{"conversation": id, "procedure": procedure})
}

// seedRun starts a run of a conversation at root: genesis, then the user side
// with payload. A conversation runs one procedure, fixed at creation
// (proposal 38): the root the human selected, recorded in its genesis.
func (h *webhost) scratchFor(home string) string {
	if h.scratch != "" {
		return h.scratch
	}
	return filepath.Join(home, "scratch")
}

func (h *webhost) seedRun(id, root, home string, seeded map[string]any) {
	payload, _ := json.Marshal(seeded)
	// The operator may choose a persistent shared space; otherwise scratch
	// belongs to this conversation and is deleted with it.
	scratch := h.scratchFor(home)
	cmd := exec.Command(h.skyra, "boot", "-root", root, "-registry", h.registry, "-scratch", scratch, "-to", "user", "-payload", string(payload), "-detach")
	cmd.Stdout, cmd.Stderr = h.log, h.log
	if err := cmd.Run(); err != nil {
		h.mu.Lock()
		if c := h.convos[id]; c != nil && c.parked == nil {
			c.phase = "failed"
			h.finishRequest(c.pending, "failed", "startup", err.Error())
			c.pending = ""
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
			if (c.phase == "pending" || c.pending != "") && c.lineage != nil {
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
		changed = h.reconcileRequests() || changed
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
		entries := withCalls(append([]entry(nil), c.entries...), c.calls())
		views = append(views, map[string]any{"conversation": c.id, "phase": c.phase, "entries": entries, "horizons": c.lineage.horizonsNow(), "pendingRequestId": pendingID(c.pending), "procedure": c.procedure})
	}
	snap := map[string]any{"type": "snapshot", "uiIncarnation": h.incarnation, "revision": h.revision, "conversations": views, "requests": h.requestSnapshot(), "destinations": h.destinations}
	h.mu.Unlock()
	b, omitted := fitSnapshot(snap, views, maxFrame)
	if omitted > 0 {
		fmt.Fprintf(h.log, "webhost: snapshot over %d bytes; omitted older history from %d views (it remains on disk)\n", maxFrame, omitted)
	}
	h.sendRaw(b)
}

// fitSnapshot keeps a snapshot within one frame, so a large history cannot
// put every client into a reconnect loop. It first lists the oldest finished
// conversations without entries, then drops the oldest half of the longest
// remaining history, until it fits. Views run newest first.
func fitSnapshot(snap map[string]any, views []map[string]any, limit int) ([]byte, int) {
	b, _ := json.Marshal(snap)
	omitted := 0
	for i := len(views) - 1; i >= 0 && len(b) > limit; i-- {
		if p := views[i]["phase"]; p == "ready" || p == "pending" {
			continue
		}
		if es := views[i]["entries"].([]entry); len(es) > 0 {
			views[i]["entries"] = []entry{}
			omitted++
			b, _ = json.Marshal(snap)
		}
	}
	for len(b) > limit {
		longest := -1
		for i, v := range views {
			if n := len(v["entries"].([]entry)); n > 1 && (longest < 0 || n > len(views[longest]["entries"].([]entry))) {
				longest = i
			}
		}
		if longest < 0 {
			break // nothing left to omit; send reports the oversized frame
		}
		es := views[longest]["entries"].([]entry)
		views[longest]["entries"] = es[len(es)/2:]
		omitted++
		b, _ = json.Marshal(snap)
	}
	return b, omitted
}

func (h *webhost) send(v any) {
	b, _ := json.Marshal(v)
	h.sendRaw(b)
}

func (h *webhost) sendRaw(b []byte) {
	h.wmu.Lock()
	defer h.wmu.Unlock()
	if h.conn != nil {
		h.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		if err := writeFrame(h.conn, 0x1, b); err != nil {
			h.conn.Close()
			h.conn = nil
		}
	}
}

func pendingID(id string) any {
	if id == "" {
		return nil
	}
	return id
}

func (h *webhost) reconcileRequests() bool {
	changed := false
	h.mu.Lock()
	for _, c := range h.convos {
		if c.pending == "" || c.parked != nil {
			continue
		}
		root := filepath.Join(h.homeOf(c), ".skyra")
		if c.lineage != nil {
			root = c.lineage.root // the run in progress
		}
		if failure, err := tree.ReadFailure(root); err == nil && failure != nil {
			c.phase, c.recoveryRequired = "failed", true
			h.finishRequest(c.pending, "failed", "delivery-unknown", failure.Message)
			c.pending = ""
			c.entries = append(c.entries, entry{Role: "agent", Text: "⚠ runtime interrupted; inspect history before retrying: " + failure.Message, TS: int64(len(c.entries))})
			changed = true
			continue
		}
		if head, err := boot.Head(root); err == nil {
			r, err := tree.Load(root, head)
			if err == nil && r.State == record.Resolve && len(r.Deps) == 1 {
				r, err = tree.Load(root, r.Deps[0])
			}
			if err == nil && (r.State == record.Return || r.State == record.Fault) {
				status := "returned"
				c.phase = "closed"
				if r.State == record.Fault {
					status = "failed"
					c.phase = "failed"
					c.entries = append(c.entries, entry{Role: "agent", Text: "⚠ the runtime faulted: " + r.Fault.Cause, TS: int64(len(c.entries))})
				}
				h.finishRequest(c.pending, status, "", "")
				c.pending = ""
				changed = true
			}
		}
	}
	h.mu.Unlock()
	return changed
}

// latestRun is a conversation's newest run: the first at home/.skyra, each
// continuation under home/runs/<n>/.skyra. Its transcript is the whole
// conversation, since each run is seeded with what came before it.
func latestRun(home string) (string, int) {
	roots := conversationRuns(home)
	root := roots[len(roots)-1]
	n := 1
	if len(roots) > 1 {
		n, _ = strconv.Atoi(filepath.Base(filepath.Dir(root)))
	}
	return root, n
}

// conversationRuns lists committed runs in numeric order, starting with the
// original run. A continuation still starting has no records and is ignored.
func conversationRuns(home string) []string {
	type run struct {
		n    int
		root string
	}
	committed := []run{}
	runs, _ := os.ReadDir(filepath.Join(home, "runs"))
	for _, r := range runs {
		k, err := strconv.Atoi(r.Name())
		if err != nil || k <= 1 || !r.IsDir() {
			continue
		}
		root := filepath.Join(home, "runs", r.Name(), ".skyra")
		if _, err := os.Stat(filepath.Join(root, "records")); err == nil {
			committed = append(committed, run{k, root})
		}
	}
	sort.Slice(committed, func(i, j int) bool { return committed[i].n < committed[j].n })
	roots := []string{filepath.Join(home, ".skyra")}
	for _, r := range committed {
		roots = append(roots, r.root)
	}
	return roots
}

// continueConversation carries a conversation on after the run it was in
// ended: the host restarted, it was closed, or it failed. A conversation is
// the user's and the agent's, not the kernel's, so it goes on in a new run of
// the same slot, seeded with the transcript so far as text and the message
// that continues it, which runs at once with the transcript as its context.
// The scratch space is the conversation's, so it carries over. Nothing of the
// old run is repeated. Caller holds mu; it is released.
func (h *webhost) continueConversation(c *conversation, m inbound, admit func(string) bool, reject func(string, string)) {
	d, ok := h.destinationFor(c.procedure)
	if !ok {
		d, ok = h.destinationFor("") // one from before procedures continues in the first root
	}
	if !ok {
		h.mu.Unlock()
		reject("unknown-procedure", "no procedure can carry this conversation on")
		return
	}
	home := h.homeOf(c)
	images, err := saveAttachments(h.scratchFor(home), m.Images)
	if err != nil {
		h.mu.Unlock()
		reject("invalid-image", err.Error())
		return
	}
	if !admit(c.id) {
		return
	}
	said := []map[string]string{}
	for _, e := range c.entries {
		if e.Role == "user" || e.Role == "agent" {
			said = append(said, map[string]string{"role": e.Role, "text": e.Text})
		}
	}
	_, n := latestRun(home)
	root := filepath.Join(home, "runs", strconv.Itoa(n+1), ".skyra")
	id := m.RequestID
	c.entries = append(c.entries, entry{Role: "user", Text: withImages(m.Text, images), TS: int64(len(c.entries)), RequestID: &id})
	c.phase, c.pending, c.recoveryRequired, c.procedure = "pending", m.RequestID, false, d.Title
	c.startLineage(root)
	h.mu.Unlock()
	h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
	h.broadcast()
	seeded := map[string]any{"conversation": c.id, "procedure": d.Title, "transcript": said, "text": m.Text}
	if len(images) > 0 {
		seeded["images"] = images
	}
	go h.seedRun(c.id, root, home, seeded)
}
