package dnav2

import "sync"

// Board — the coordination slice's job board: one Row per in-flight impulse, keyed
// by impulseID. The dynamic runtime state of the workflow engine (the STATIC
// workflow definition lives elsewhere, on the orchestrator). Holds the join
// (expects/done), the one-reframe rail, the atomic motor guard, the append-only
// trace, and the causal link. "In flight" = pending rows; motor reads them.
//
// Concurrency: the Board's lock guards only row creation/lookup (brief); each Row
// carries its own lock for the actual coordination — the row-level locking from the
// substrate model, never held across inference.
type Board struct {
	mu   sync.Mutex
	rows map[string]*Row
}

func NewBoard() *Board { return &Board{rows: map[string]*Row{}} }

// Row is one impulse's transaction record.
type Row struct {
	mu sync.Mutex

	ImpulseID string
	Prev      string // causal link — the impulse whose action provoked this one (reafference chain)

	expects    map[string]bool // set of "grouping:agent" the join waits on
	done       map[string]bool // completed "grouping:agent"
	reframed   bool            // the one-reframe rail
	motorFired bool            // the atomic join guard
	status     string          // "pending" | "done"
	history    []Span          // append-only trace (immutable spans; flush on done)
}

// Span is one step's trace entry (append-only, immutable once written).
type Span struct {
	Step      string
	Condition string
	At        int
}

// Row returns the row for impulseID, creating it (pending) on first sight. prev is
// the causal predecessor (empty for a fresh external event).
func (b *Board) Row(impulseID, prev string) *Row {
	b.mu.Lock()
	defer b.mu.Unlock()
	r := b.rows[impulseID]
	if r == nil {
		r = &Row{
			ImpulseID: impulseID,
			Prev:      prev,
			expects:   map[string]bool{},
			done:      map[string]bool{},
			status:    "pending",
		}
		b.rows[impulseID] = r
	}
	return r
}

// Pending returns the impulseIDs whose join isn't satisfied yet (what motor/other
// inspection reads — "what's in flight").
func (b *Board) Pending() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	var out []string
	for id, r := range b.rows {
		r.mu.Lock()
		if r.status == "pending" {
			out = append(out, id)
		}
		r.mu.Unlock()
	}
	return out
}

// key joins a grouping + agent into the set key.
func key(grouping, agent string) string { return grouping + ":" + agent }

// Declare registers what the join waits on (midlevel declares the expected set).
// Additive — a warm impulse can declare several groupings.
func (r *Row) Declare(grouping string, agents ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, a := range agents {
		r.expects[key(grouping, a)] = true
	}
}

// Mark records a completed (grouping, agent). Idempotent.
func (r *Row) Mark(grouping, agent string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.done[key(grouping, agent)] = true
}

// TryFireMotor is the atomic join: returns true to EXACTLY ONE caller when the join
// is satisfied and motor hasn't fired. Both fan-out branches call it on completion;
// the row lock makes the check-and-set atomic so motor fires once.
func (r *Row) TryFireMotor() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.motorFired || len(r.expects) == 0 {
		return false
	}
	for k := range r.expects {
		if !r.done[k] {
			return false // join not yet satisfied
		}
	}
	r.motorFired = true
	return true
}

// TryReframe is the one-reframe rail: returns true (allowed) the first time only;
// subsequent reframe attempts on the same impulse return false (suppressed).
func (r *Row) TryReframe() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.reframed {
		return false
	}
	r.reframed = true
	return true
}

// Append adds a span to the trace (append-only, row-locked write). Reads of already-
// written spans are lock-free (spans are immutable) — see History.
func (r *Row) Append(s Span) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.history = append(r.history, s)
}

// Done flips the row to done (the join completed, motor fired). Returns the trace
// for the flush hook (the external store; deferred).
func (r *Row) Done() []Span {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.status = "done"
	out := make([]Span, len(r.history))
	copy(out, r.history)
	return out
}

func (r *Row) Status() string { return r.status }
