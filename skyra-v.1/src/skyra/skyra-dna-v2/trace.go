package dnav2

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// RunTrace — one impulse's content-level trace: every inference call with its
// FULL system/present/response, every contract rejection with the parse error
// that caused it, every step span, and worker notes. Written as JSONL and
// synced per event, so a killed run keeps everything recorded up to the kill —
// the property the first live run (2026-07-13) lacked: ten calls, zero
// recoverable outputs (size-only logging + codex --ephemeral + in-memory graph).
//
// The Board row owns the trace (one per board run); the membrane threads it to
// each worker at transcription. All methods are nil-safe: a nil *RunTrace is a
// no-op, so tests and trace-less runs pay nothing.
type RunTrace struct {
	mu    sync.Mutex
	f     *os.File
	seq   int
	start time.Time
}

// TraceEvent is one JSONL line. Kind is one of "span" (step completed),
// "infer" (one inference call, full content), "reject" (a returned response
// failed contract validation), "note" (worker-recorded fact).
type TraceEvent struct {
	Seq       int    `json:"seq"`
	TMs       int64  `json:"tMs"`
	Kind      string `json:"kind"`
	Stage     string `json:"stage,omitempty"`
	Step      string `json:"step,omitempty"`
	Condition string `json:"condition,omitempty"`
	System    string `json:"system,omitempty"`
	Present   string `json:"present,omitempty"`
	Response  string `json:"response,omitempty"`
	DurMs     int64  `json:"durMs,omitempty"`
	Err       string `json:"err,omitempty"`
	Note      string `json:"note,omitempty"`
}

// TraceStore mints one RunTrace per board run under dir.
type TraceStore struct {
	mu  sync.Mutex
	dir string
	seq int
}

func NewTraceStore(dir string) *TraceStore { return &TraceStore{dir: dir} }

// Open creates the trace file for one impulse. Errors degrade to a nil trace
// (tracing must never take down a turn).
func (s *TraceStore) Open(impulseID string) *RunTrace {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	s.seq++
	seq := s.seq
	s.mu.Unlock()
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return nil
	}
	name := fmt.Sprintf("%03d-%s.jsonl", seq, sanitize(impulseID))
	f, err := os.Create(filepath.Join(s.dir, name))
	if err != nil {
		return nil
	}
	return &RunTrace{f: f, start: time.Now()}
}

func sanitize(id string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		}
		return '_'
	}, id)
}

func (t *RunTrace) write(e TraceEvent) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.f == nil {
		return
	}
	t.seq++
	e.Seq = t.seq
	e.TMs = time.Since(t.start).Milliseconds()
	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	t.f.Write(append(line, '\n'))
	t.f.Sync() // per-event durability — the whole point
}

// Span records a completed workflow step and the condition it reported.
func (t *RunTrace) Span(step, condition string) {
	t.write(TraceEvent{Kind: "span", Step: step, Condition: condition})
}

// Reject records a response that returned fine but failed contract validation
// (parse error, offer violation, budget violation) — the retry's cause.
func (t *RunTrace) Reject(stage, response string, err error) {
	e := TraceEvent{Kind: "reject", Stage: stage, Response: response}
	if err != nil {
		e.Err = err.Error()
	}
	t.write(e)
}

// Note records a worker-level fact (persisted node IDs, chosen structural op,
// set activation) in plain text.
func (t *RunTrace) Note(stage, text string) {
	t.write(TraceEvent{Kind: "note", Stage: stage, Note: text})
}

// Wrap decorates an InferFunc so every call through it lands in the trace with
// full content and duration. Nil-safe on both sides: a nil trace returns fn
// unchanged (preserving genes' infer-is-nil checks), a nil fn stays nil.
func (t *RunTrace) Wrap(fn InferFunc, stage string) InferFunc {
	if t == nil || fn == nil {
		return fn
	}
	return func(system, present string) (string, error) {
		start := time.Now()
		resp, err := fn(system, present)
		e := TraceEvent{
			Kind: "infer", Stage: stage,
			System: system, Present: present, Response: resp,
			DurMs: time.Since(start).Milliseconds(),
		}
		if err != nil {
			e.Err = err.Error()
		}
		t.write(e)
		return resp, err
	}
}

// Close flushes and closes the file. Safe on nil; later writes become no-ops.
func (t *RunTrace) Close() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.f != nil {
		t.f.Close()
		t.f = nil
	}
}
