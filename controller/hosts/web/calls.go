package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// call is one tool call in an agent turn, read from the lineage: an invoke of
// a tool and what discharged it, or a fault whose observed input addressed a
// tool. Clients draw it (a ledger line, a light in the sky); the webhost only
// reads records and decides nothing.
type call struct {
	To     string `json:"to"`
	Arg    string `json:"arg"`
	State  string `json:"state"`            // running | returned | fault
	Result string `json:"result,omitempty"` // returned: what came back, in a few words
	Fault  string `json:"fault,omitempty"`  // fault: its kind
	Cause  string `json:"cause,omitempty"`  // fault: its cause
	Stamp  string `json:"stamp"`            // the invoke's stamp, or the fault's
	MS     int64  `json:"ms,omitempty"`     // returned: from the invoke's commit to the return's
}

// lineage reads one conversation's record tree as it grows. Records are
// write-once, so each is loaded once.
type lineage struct {
	root string

	mu       sync.Mutex
	records  map[string]lrec
	turns    [][]call
	released map[string]horizon.Horizon // every Horizon released in it, by id
	horizons []horizonView
}

type lrec struct {
	r  record.Record
	at time.Time
}

func newLineage(root string) *lineage {
	return &lineage{root: root, records: map[string]lrec{}, released: map[string]horizon.Horizon{}}
}

// refresh loads records committed and Horizons released since the last call
// and re-derives the turns and Horizons. It reports whether either changed.
func (l *lineage) refresh() bool {
	files, err := os.ReadDir(filepath.Join(l.root, "records"))
	if err != nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	added := l.loadHorizons()
	for _, f := range files {
		name := f.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if _, ok := l.records[name]; ok {
			continue
		}
		r, err := tree.Load(l.root, name)
		if err != nil {
			continue
		}
		at := time.Time{}
		if info, err := f.Info(); err == nil {
			at = info.ModTime()
		}
		l.records[name] = lrec{r, at}
		added = true
	}
	if !added {
		return false
	}
	turns := l.derive()
	changed := !sameTurns(turns, l.turns)
	l.turns = turns
	var genesis *record.Record
	for _, x := range l.records {
		if x.r.Predecessor == "" {
			genesis = &x.r
		}
	}
	var released []horizon.Horizon
	for _, h := range l.released {
		released = append(released, h)
	}
	views := horizons(genesis, released)
	before, _ := json.Marshal(l.horizons)
	after, _ := json.Marshal(views)
	l.horizons = views
	return changed || string(before) != string(after)
}

// loadHorizons loads Horizons released since the last call. Caller holds mu.
// A conversation recorded before Horizons were stored has none: it is drawn
// from its genesis alone.
func (l *lineage) loadHorizons() bool {
	files, _ := os.ReadDir(filepath.Join(l.root, "horizons"))
	added := false
	for _, f := range files {
		id := f.Name()
		if _, ok := l.released[id]; ok || strings.HasPrefix(id, ".") {
			continue
		}
		h, err := tree.LoadHorizon(l.root, id)
		if err != nil {
			continue
		}
		l.released[id] = h
		added = true
	}
	return added
}

func (l *lineage) horizonsNow() []horizonView {
	if l == nil {
		return []horizonView{}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.horizons == nil {
		return []horizonView{}
	}
	return l.horizons
}

func (l *lineage) snapshot() [][]call {
	if l == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.turns
}

// live returns the records order walks, in its order.
func (l *lineage) live() []record.Record {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []record.Record
	for _, s := range l.order() {
		out = append(out, l.records[s].r)
	}
	return out
}

// order walks the chain from genesis, each record after its predecessor;
// siblings (a fan-out) in the order they were committed. A record whose
// predecessor also has a supervisor fault was never released: that instance
// kept custody and faulted its own frame, so the successor it had spawned
// (killed, or dead before its stamp arrived) is left out.
func (l *lineage) order() []string {
	children := map[string][]string{}
	for s, x := range l.records {
		children[x.r.Predecessor] = append(children[x.r.Predecessor], s)
	}
	for p, kids := range children {
		var faulted []string
		for _, k := range kids {
			if l.records[k].r.Supervisor {
				faulted = append(faulted, k)
			}
		}
		if faulted != nil {
			children[p] = faulted
		}
	}
	for _, kids := range children {
		sort.Slice(kids, func(i, j int) bool {
			a, b := l.records[kids[i]], l.records[kids[j]]
			if !a.at.Equal(b.at) {
				return a.at.Before(b.at)
			}
			return kids[i] < kids[j]
		})
	}
	var out []string
	stack := append([]string(nil), children[""]...)
	for len(stack) > 0 {
		s := stack[0]
		stack = stack[1:]
		out = append(out, s)
		stack = append(append([]string(nil), children[s]...), stack...)
	}
	return out
}

// derive groups tool calls by agent turn: a turn starts where the user side
// addresses codex, which it does once per human message.
func (l *lineage) derive() [][]call {
	turns := [][]call{}
	open := map[string]int{} // invoke stamp → index in the current turn
	for _, s := range l.order() {
		x := l.records[s]
		r := x.r
		switch r.State {
		case record.Invoke:
			if r.To == "codex" {
				turns = append(turns, []call{})
				open = map[string]int{}
				continue
			}
			if len(turns) == 0 || !isTool(r.To) {
				continue
			}
			t := len(turns) - 1
			open[s] = len(turns[t])
			turns[t] = append(turns[t], call{To: r.To, Arg: argument(r.Payload), State: "running", Stamp: s})
		case record.Return:
			if i, ok := open[r.Predecessor]; ok {
				c := &turns[len(turns)-1][i]
				c.State, c.Result = "returned", summarize(r.Payload)
				if inv := l.records[r.Predecessor]; !inv.at.IsZero() && !x.at.IsZero() {
					c.MS = x.at.Sub(inv.at).Milliseconds()
				}
				delete(open, r.Predecessor)
			}
		case record.Fault:
			if r.Fault == nil {
				continue
			}
			if i, ok := open[r.Predecessor]; ok {
				// The tool itself faulted: its call ends there.
				c := &turns[len(turns)-1][i]
				c.State, c.Fault, c.Cause = "fault", r.Fault.Kind, r.Fault.Cause
				delete(open, r.Predecessor)
				continue
			}
			// A refused emission: the input it observed names what was asked for.
			var in struct {
				To      string          `json:"to"`
				Payload json.RawMessage `json:"payload"`
			}
			b, err := tree.Get(l.root, "inputs", r.Input)
			if err != nil || json.Unmarshal(b, &in) != nil {
				continue
			}
			if in.To == "codex" {
				turns = append(turns, []call{})
				open = map[string]int{}
				continue
			}
			// A return refused by its own output contract addressed the
			// emitter's slot, not a tool.
			if len(turns) == 0 || !isTool(in.To) || strings.HasPrefix(r.Fault.Cause, "return: ") {
				continue
			}
			t := len(turns) - 1
			turns[t] = append(turns[t], call{To: in.To, Arg: argument(in.Payload), State: "fault", Fault: r.Fault.Kind, Cause: r.Fault.Cause, Stamp: s})
		}
	}
	return turns
}

// isTool reports whether an address is a tool the agent called, rather than
// the conversation's own plumbing: the user side, the browser, and the Codex
// adapter's steps.
func isTool(to string) bool {
	switch to {
	case "", "user", "browser", "codex":
		return false
	}
	return !strings.HasPrefix(to, "codex.")
}

// argument is a call's payload in a few words: its path or text when it has
// one, otherwise its JSON.
func argument(p json.RawMessage) string {
	var m map[string]json.RawMessage
	if json.Unmarshal(p, &m) == nil {
		for _, k := range []string{"path", "text", "command"} {
			var s string
			if json.Unmarshal(m[k], &s) == nil && s != "" {
				return clip(s)
			}
		}
		if len(m) == 0 {
			return ""
		}
	}
	var s string
	if json.Unmarshal(p, &s) == nil {
		return clip(s)
	}
	return clip(string(p))
}

// summarize is what a return carried: its error when it names one, else its size.
func summarize(p json.RawMessage) string {
	var m map[string]json.RawMessage
	if json.Unmarshal(p, &m) == nil {
		var e string
		if json.Unmarshal(m["error"], &e) == nil && e != "" {
			return "error: " + e
		}
	}
	n := len(p)
	if n < 1000 {
		return fmt.Sprintf("returned %d B", n)
	}
	return fmt.Sprintf("returned %.1f kB", float64(n)/1000)
}

func clip(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if r := []rune(s); len(r) > 60 {
		return string(r[:59]) + "…"
	}
	return s
}

func sameTurns(a, b [][]call) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}

// withCalls places each turn's calls after the human message that started
// it. When a turn ended in a refused call, the user side's line reporting the
// same fault is left out: the call's own line already says it.
func withCalls(entries []entry, turns [][]call) []entry {
	out := make([]entry, 0, len(entries))
	turn := -1
	var last *call
	for _, e := range entries {
		if last != nil && e.Role == "agent" && e.Text == fmt.Sprintf("⚠ the agent faulted (%s): %s", last.Fault, last.Cause) {
			last = nil
			continue
		}
		last = nil
		out = append(out, e)
		if e.Role != "user" {
			continue
		}
		turn++
		if turn >= len(turns) {
			continue
		}
		for i := range turns[turn] {
			c := turns[turn][i]
			out = append(out, entry{Role: "call", Text: strings.TrimSpace(c.To + " " + c.Arg), Call: &c})
			if c.State == "fault" {
				last = &c
			} else {
				last = nil
			}
		}
	}
	return stamp(out)
}
