package dnav2

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"skyra-v1/src/debug"
)

// InferLogger wraps an InferFunc so EVERY inference call — every gene, the edit/new op, and
// each settle pass — is timed and logged to the being's "inference" layer. It is the single
// chokepoint: all genes call DNA.Infer(), so wrapping that one seam instruments the whole
// run with no per-call-site changes. Built for live Codex runs, where calls are slow and we
// want to see where the wall-clock time goes.
//
// Per call it logs two lines (tied by #seq): a START line with the elapsed t+ since the
// first call and the gap since the previous call RETURNED (idle time on the pipe — negative/
// noisy under the concurrent predict fan-out, clean during the sequential midlevel steps),
// and an END line with the call's duration and outcome. Summary() rolls up the totals.
type InferLogger struct {
	name  string
	fn    InferFunc
	start time.Time

	mu       sync.Mutex
	seq      int
	lastEnd  time.Time
	total    time.Duration
	max      time.Duration
	errs     int
	firstSet bool
}

// NewInferLogger decorates fn. Use il.Infer as the DNA's InferFunc (SetInfer(il.Infer)) and
// print il.Summary() at the end of the run.
func NewInferLogger(name string, fn InferFunc) *InferLogger {
	return &InferLogger{name: name, fn: fn}
}

// Infer is the wrapped InferFunc — drop-in for CallCodex.
func (l *InferLogger) Infer(system, present string) (string, error) {
	l.mu.Lock()
	if !l.firstSet {
		l.start, l.firstSet = time.Now(), true
	}
	l.seq++
	seq := l.seq
	gap := time.Duration(0)
	if !l.lastEnd.IsZero() {
		gap = time.Since(l.lastEnd)
	}
	elapsed := time.Since(l.start)
	l.mu.Unlock()

	debug.Being(l.name, "inference", fmt.Sprintf("#%d → [%s] prompt=%dc t+%.1fs gap=%.1fs",
		seq, firstLine(system, 48), len(system)+len(present), elapsed.Seconds(), gap.Seconds()))

	callStart := time.Now()
	resp, err := l.fn(system, present)
	dur := time.Since(callStart)

	l.mu.Lock()
	l.lastEnd = time.Now()
	l.total += dur
	if dur > l.max {
		l.max = dur
	}
	if err != nil {
		l.errs++
	}
	l.mu.Unlock()

	if err != nil {
		debug.Being(l.name, "inference", fmt.Sprintf("#%d ✗ ERR dur=%.2fs %v", seq, dur.Seconds(), err))
	} else {
		debug.Being(l.name, "inference", fmt.Sprintf("#%d ✓ dur=%.2fs resp=%dc", seq, dur.Seconds(), len(resp)))
	}
	return resp, err
}

// Summary is a one-line rollup of the run's inference cost: call count, errors, summed call
// time (which exceeds wall time when calls overlapped), mean, max, and wall span.
func (l *InferLogger) Summary() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.seq == 0 {
		return "inference: no calls"
	}
	mean := l.total / time.Duration(l.seq)
	wall := time.Since(l.start)
	return fmt.Sprintf("inference: %d calls, %d errs — summed %.1fs, mean %.1fs, max %.1fs, over wall %.1fs",
		l.seq, l.errs, l.total.Seconds(), mean.Seconds(), l.max.Seconds(), wall.Seconds())
}

// firstLine returns s up to the first newline, trimmed and capped — a legible tag for the
// call (our system prompts open with a distinctive "You are the coherence check…" line).
func firstLine(s string, max int) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = s[:max]
	}
	return s
}
