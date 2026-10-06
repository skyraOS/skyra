// Command host is the universal actor host (next/15): the one long-lived
// process that sits in front of an asynchronous entity the runtime does not
// control. It owns the actor's connection and nothing else. It makes no
// decisions: it never parses what the actor says, never retries, never
// times anything out on the actor's behalf. Adapter steps, which are
// stateless instances, make every decision by reading its inbox.
//
//	host -sock PATH -state DIR -- ACTOR [ARGS...]
//
// Every host answers the same universal contracts, one JSON request per line
// on a unix socket:
//
//	{"op":"deliver","id":ID,"line":LINE}  write LINE to the actor, once per ID.
//	                                       The same ID and LINE again is a no-op
//	                                       that returns the first answer; the same
//	                                       ID with another LINE is a conflict
//	{"op":"read","from":N,"wait_ms":T}     inbox events from sequence N, waiting
//	                                       up to T ms for one to arrive
//	{"op":"status","id":ID}                whether ID was delivered, and where
//	{"op":"ping"}                          liveness, as an observation
//
// The inbox is every line the actor wrote, in order, plus the host's own
// observations of the actor (its exit), each written durably before it is
// served.
package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"skyra/controller/runtime/custody"
)

type event struct {
	Seq  int             `json:"seq"`
	Line json.RawMessage `json:"line,omitempty"` // what the actor wrote, verbatim (JSON when it is JSON)
	Text string          `json:"text,omitempty"` // what the actor wrote, when it is not JSON
	Host string          `json:"host,omitempty"` // the host's own observation
}

type delivery struct {
	ID   string `json:"id"`
	Hash string `json:"hash"`
	Seq  int    `json:"seq"` // the inbox length when it was first delivered: read from here for its effects
}

type host struct {
	mu         sync.Mutex
	arrived    *sync.Cond
	events     []event
	delivered  map[string]delivery
	inbox, log *os.File
	actor      io.Writer
	alive      bool
}

func main() {
	custody.SealInherited() // nothing the launcher leaked reaches the actor
	sock := flag.String("sock", "", "unix socket to serve the universal contracts on")
	state := flag.String("state", "", "directory for the durable inbox and delivery log")
	flag.Parse()
	if *sock == "" || *state == "" || flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: host -sock PATH -state DIR -- ACTOR [ARGS...]")
		os.Exit(2)
	}
	if err := run(*sock, *state, flag.Args()); err != nil {
		fmt.Fprintln(os.Stderr, "host:", err)
		os.Exit(1)
	}
}

func run(sock, state string, argv []string) error {
	if err := os.MkdirAll(state, 0o700); err != nil {
		return err
	}
	h := &host{delivered: map[string]delivery{}}
	h.arrived = sync.NewCond(&h.mu)
	var err error
	if h.inbox, err = os.OpenFile(filepath.Join(state, "inbox.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err != nil {
		return err
	}
	if h.log, err = os.OpenFile(filepath.Join(state, "deliveries.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600); err != nil {
		return err
	}

	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stderr = os.Stderr
	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	h.actor, h.alive = in, true
	go func() {
		r := bufio.NewReaderSize(out, 1<<20)
		for {
			line, err := r.ReadBytes('\n')
			if len(line) > 0 {
				h.append(event{Line: trim(line)})
			}
			if err != nil {
				break
			}
		}
		status := cmd.Wait()
		h.mu.Lock()
		h.alive = false
		h.mu.Unlock()
		h.append(event{Host: fmt.Sprintf("actor exited: %v", status)})
	}()

	os.Remove(sock)
	l, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	defer os.Remove(sock)
	fmt.Fprintf(os.Stderr, "host: serving %s for %v (pid %d)\n", sock, argv, cmd.Process.Pid)
	for {
		c, err := l.Accept()
		if err != nil {
			return err
		}
		go h.serve(c)
	}
}

func trim(line []byte) []byte {
	for len(line) > 0 && (line[len(line)-1] == '\n' || line[len(line)-1] == '\r') {
		line = line[:len(line)-1]
	}
	return line
}

// append records one inbox event durably, then wakes readers.
func (h *host) append(e event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	e.Seq = len(h.events)
	if e.Line != nil && !json.Valid(e.Line) {
		e.Text, e.Line = string(e.Line), nil
	}
	b, _ := json.Marshal(e)
	h.inbox.Write(append(b, '\n'))
	h.inbox.Sync()
	h.events = append(h.events, e)
	h.arrived.Broadcast()
}

type request struct {
	Op     string          `json:"op"`
	ID     string          `json:"id"`
	Line   json.RawMessage `json:"line"`
	From   int             `json:"from"`
	WaitMS int             `json:"wait_ms"`
}

func (h *host) serve(c net.Conn) {
	defer c.Close()
	r := bufio.NewReaderSize(c, 1<<20)
	enc := json.NewEncoder(c)
	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			return
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			enc.Encode(map[string]string{"error": "malformed request"})
			continue
		}
		enc.Encode(h.handle(req))
	}
}

func (h *host) handle(req request) any {
	switch req.Op {
	case "deliver":
		d, again, err := h.deliver(req.ID, req.Line)
		if err != nil {
			return map[string]string{"error": err.Error()}
		}
		return map[string]any{"seq": d.Seq, "again": again}
	case "read":
		return h.read(req.From, time.Duration(req.WaitMS)*time.Millisecond)
	case "status":
		h.mu.Lock()
		defer h.mu.Unlock()
		d, ok := h.delivered[req.ID]
		return map[string]any{"delivered": ok, "seq": d.Seq}
	case "ping":
		h.mu.Lock()
		defer h.mu.Unlock()
		return map[string]any{"alive": h.alive, "events": len(h.events), "now_ms": time.Now().UnixMilli()}
	}
	return map[string]string{"error": "unknown op"}
}

// deliver writes a line to the actor at most once per id. The delivery is
// logged durably before the actor sees it.
func (h *host) deliver(id string, line json.RawMessage) (delivery, bool, error) {
	if id == "" || len(line) == 0 {
		return delivery{}, false, errors.New("deliver needs an id and a line")
	}
	sum := sha256.Sum256(line)
	hash := hex.EncodeToString(sum[:])
	h.mu.Lock()
	defer h.mu.Unlock()
	if d, ok := h.delivered[id]; ok {
		if d.Hash != hash {
			return delivery{}, false, fmt.Errorf("conflict: %s was delivered with a different line", id)
		}
		return d, true, nil
	}
	if !h.alive {
		return delivery{}, false, errors.New("actor is not alive")
	}
	d := delivery{ID: id, Hash: hash, Seq: len(h.events)}
	b, _ := json.Marshal(d)
	h.log.Write(append(b, '\n'))
	h.log.Sync()
	if _, err := h.actor.Write(append(append([]byte{}, line...), '\n')); err != nil {
		return delivery{}, false, err
	}
	h.delivered[id] = d
	return d, false, nil
}

// read returns every event from seq on, waiting up to wait for the first.
func (h *host) read(from int, wait time.Duration) any {
	deadline := time.Now().Add(wait)
	h.mu.Lock()
	defer h.mu.Unlock()
	for len(h.events) <= from && time.Now().Before(deadline) {
		timer := time.AfterFunc(time.Until(deadline), func() {
			h.mu.Lock()
			h.arrived.Broadcast()
			h.mu.Unlock()
		})
		h.arrived.Wait()
		timer.Stop()
	}
	var out []event
	if from < len(h.events) {
		out = append(out, h.events[from:]...)
	}
	return map[string]any{"events": out, "next": len(h.events)}
}
