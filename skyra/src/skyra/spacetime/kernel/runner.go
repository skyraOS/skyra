package kernel

// A Runner holds several live executions and advances them one step at a time.
//
// # CURRENTLY UNREACHABLE — 2026-08-06
//
// specs/2026-08-03/execution.md adopted ONE STACK: one subgraph active at a time, connected,
// entered at a root. Under that rule nothing here runs — Capacity is never reached, Evict
// never fires, and round-robin has nothing to rotate between.
//
// It is kept rather than deleted because the concurrency deferral names "many processes, one
// filesystem" as the eventual target and this is that machinery. But the tests below describe
// behavior nothing currently exercises, and should not be read as describing the live runtime.
//
// What made one stack safe was actions-as-residual: a wave is graph reads and graph writes and
// nothing else, so it never blocks and there is no long-running execution for an arriving
// impulse to compete with. If a wave ever regains the ability to block — a model call inside
// the wave, a synchronous world call — that premise dies and this file becomes load-bearing
// again.
//
// The property this file WAS carrying, which must be preserved wherever impulses now queue:
//
//	ADMISSION MUST NOT CONSULT ATTENTION.
//
// Admit never refused; a FIFO queue must not either. A queue that orders or drops by relevance,
// trust, or activation is the sealing failure moved into the intake.
//
//	MULTIPLE STACKS IS NOT PARALLEL EXECUTION.
//
// One runner, one step at a time, strictly serial. What it buys is structural rather than
// throughput: an arriving impulse gets its own stack instead of competing for a single
// execution's attention, so goal-directed work and stimulus-driven work stop arbitrating
// over one pointer. Nothing here runs concurrently, and nothing shares mutable state, so
// the concurrency deferral in specs/2026-08-03/graph.md still holds.
//
// # Finite attention lives HERE, and that is the point
//
// Attention has to be bounded somewhere. Bounding it as an epistemic filter — deciding
// whether an arriving observation counts — is what seals the loop (2026-07-27 Toy 5, three
// times over). Bounding it as a RESOURCE LIMIT does not:
//
//	FILTER   at capacity, reject the arrival    → contradiction never enters
//	LIMIT    at capacity, evict something else  → arrival always enters
//
// So Admit ALWAYS SUCCEEDS. Pressure evicts an existing execution and never refuses a new
// one. You lose work rather than evidence, and lost work can be redone.
type Runner struct {
	// Capacity is how many executions may be live at once. Zero or less means
	// unbounded, which is only appropriate for tests.
	Capacity int

	// Evict chooses the victim when Admit is at capacity. Nil means EvictOldest.
	// It is a policy hook because eviction quietly decides what gets learned — an
	// evicted stack's accumulation dies with it, unwritten.
	Evict func(live []*Stack) int

	live []*Stack
	next int // round-robin cursor
}

// Admit adds an execution and returns whatever was evicted to make room, or nil.
//
// IT NEVER REFUSES. The returned stack is the caller's to inspect: its transcript is valid
// and says honestly how far it got before being dropped.
func (r *Runner) Admit(s *Stack) *Stack {
	if s == nil {
		return nil
	}
	var evicted *Stack
	if r.Capacity > 0 && len(r.live) >= r.Capacity {
		evicted = r.drop(r.victim())
	}
	r.live = append(r.live, s)
	return evicted
}

func (r *Runner) victim() int {
	if r.Evict != nil {
		if i := r.Evict(r.live); i >= 0 && i < len(r.live) {
			return i
		}
	}
	return EvictOldest(r.live)
}

// EvictOldest drops the execution admitted longest ago. The default, and defensible
// rather than principled: it is predictable, and it does not let a stack's own activity
// protect it from eviction — which a least-recently-stepped policy would, giving a busy
// execution the power to keep itself alive.
func EvictOldest(live []*Stack) int { return 0 }

func (r *Runner) drop(i int) *Stack {
	if i < 0 || i >= len(r.live) {
		return nil
	}
	s := r.live[i]
	r.live = append(r.live[:i], r.live[i+1:]...)
	if r.next > i {
		r.next--
	}
	return s
}

// Step advances ONE execution by one action, round-robin, and reports whether any work
// was done. Finished executions are reaped as they are encountered.
//
// Round-robin is deliberately the dumbest possible policy. Scheduling by trust, activation,
// or depth would put the controller back in charge of which executions make progress, which
// is the sealing failure relocated one level up.
func (r *Runner) Step() bool {
	for attempts := len(r.live); attempts > 0; attempts-- {
		if len(r.live) == 0 {
			return false
		}
		if r.next >= len(r.live) {
			r.next = 0
		}
		s := r.live[r.next]
		if s.Done() {
			r.drop(r.next)
			continue
		}
		s.Step()
		r.next++
		return true
	}
	return false
}

// RunAll advances every live execution to completion, interleaved.
func (r *Runner) RunAll() {
	for r.Step() {
	}
}

// Live is how many executions are currently held.
func (r *Runner) Live() int { return len(r.live) }

// Stacks exposes the live executions in admission order. For a scheduler or an inspector;
// the runner itself reads nothing off them but Done.
func (r *Runner) Stacks() []*Stack { return r.live }
