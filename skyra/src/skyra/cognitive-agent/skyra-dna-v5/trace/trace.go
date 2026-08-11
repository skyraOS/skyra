// Package trace writes a per-run record of everything that crossed a boundary.
//
// This is NOT part of the design — the graph is the durable record and the kernel
// holds what is happening. A trace is an operational facility, out of band,
// unavailable to anything inside the model. It exists because kernel.md is
// explicit that you cannot ask the graph what the system is doing, and something
// still has to be debuggable.
//
// Named here so nobody later mistakes it for structure and builds a dashboard out
// of engrams.
package trace

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Run struct {
	Dir   string
	mu    sync.Mutex
	main  *os.File
	steps int
}

// Open creates a run directory under root, named by start time.
func Open(root string) (*Run, error) {
	dir := filepath.Join(root, time.Now().Format("2006-01-02T15-04-05"))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(filepath.Join(dir, "run.log"))
	if err != nil {
		return nil, err
	}
	return &Run{Dir: dir, main: f}, nil
}

func (r *Run) Close() error {
	if r == nil {
		return nil
	}
	return r.main.Close()
}

// Logf writes one line to the run log and to stderr, so a long run is watchable.
func (r *Run) Logf(format string, a ...any) {
	if r == nil {
		return
	}
	line := fmt.Sprintf(format, a...)
	r.mu.Lock()
	defer r.mu.Unlock()
	stamp := time.Now().Format("15:04:05.000")
	fmt.Fprintf(r.main, "%s  %s\n", stamp, line)
	fmt.Fprintf(os.Stderr, "%s  %s\n", stamp, line)
}

// Step writes the full prompt and response of one entity crossing to its own
// file. These are large and the point of keeping them is to be able to read what
// an entity was ACTUALLY shown when its answer looks strange.
func (r *Run) Step(engram, ent, kind, prompt, response string) string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	r.steps++
	n := r.steps
	r.mu.Unlock()

	name := fmt.Sprintf("%03d-%s-%s.txt", n, sanitize(string(engram)), sanitize(string(ent)))
	path := filepath.Join(r.Dir, name)
	var b strings.Builder
	fmt.Fprintf(&b, "ENGRAM   %s\nENTITY   %s\nKIND     %s\n", engram, ent, kind)
	b.WriteString("\n==================== PROMPT ====================\n")
	b.WriteString(prompt)
	b.WriteString("\n\n=================== RESPONSE ===================\n")
	b.WriteString(response)
	b.WriteString("\n")
	_ = os.WriteFile(path, []byte(b.String()), 0o644)
	return name
}

// Write drops an arbitrary artifact into the run directory.
func (r *Run) Write(name, body string) {
	if r == nil {
		return
	}
	_ = os.WriteFile(filepath.Join(r.Dir, name), []byte(body), 0o644)
}

func sanitize(s string) string {
	s = strings.ReplaceAll(s, "/", "-")
	s = strings.ReplaceAll(s, " ", "-")
	if len(s) > 24 {
		s = s[:24]
	}
	return s
}
