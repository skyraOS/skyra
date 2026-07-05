package dnav2

import (
	"math"
	"sync"
)

// Phase 2 (traversal-implementation-plan.md → traversal-mechanics.md "Activation & the
// bump"): the edge-owned activation record. Four factors stored SEPARATELY, maintained
// by the bump. Strength (the read) is their product; Cross updates F (a traversal),
// Grade updates A/S/N (an outcome). This is isolated dnav2 code — the shared
// dna.memoryActivation stays untouched until the Phase-3 collapse reconciles them.

// Ship defaults (the three knobs from the spec, with defaults flagged there).
const (
	alphaFloor  = 0.1  // A learning-rate floor: running-mean while N small → EMA once floored
	lambdaDecay = 0.98 // F recency decay per tick
	alphaF      = 0.3  // F reinforcement rate on a crossing
	softFloor   = 0.05 // rect floor: even a fully-failed edge stays barely alive (recoverable)
)

// Activation is one edge's record (A→B distinct from C→B).
type Activation struct {
	A     float64 // alignment — signed running outcome value (birth 0 = neutral)
	S     float64 // salience — arousal at reinforcement; immutable once stamped (birth 1)
	F     float64 // frequency/recency — lazy timestamped EMA (birth 0; Cross-at-birth lifts it)
	Flast int     // tick of the last F update
	N     int     // gradings seen — sets the learning rate AND confidence
	sSet  bool    // whether S has been stamped (immutable-at-birth)
}

func newActivation() *Activation { return &Activation{S: 1} }

// Strength is the edge's read-time weight — rect(A)·S·F·conf(N). Non-mutating.
func (a *Activation) Strength(tick int) float64 {
	return rectSoft(a.A) * a.S * a.readF(tick) * a.conf()
}

// readF decays F to `tick` without mutating (the lazy timestamped read).
func (a *Activation) readF(tick int) float64 {
	if d := tick - a.Flast; d > 0 {
		return a.F * math.Pow(lambdaDecay, float64(d))
	}
	return a.F
}

// Cross bumps F on a traversal (lazy timestamped EMA). Fired whenever the edge is
// crossed — including at node birth, since a born edge was just active.
func (a *Activation) Cross(tick int) {
	decayed := a.readF(tick)
	a.F = decayed + alphaF*(1-decayed)
	a.Flast = tick
}

// Grade bumps A/S/N on a closure outcome g ∈ [−1,1] at the being's current stress.
func (a *Activation) Grade(g, stress float64) {
	a.N++
	alpha := 1.0 / float64(a.N)
	if alpha < alphaFloor {
		alpha = alphaFloor
	}
	a.A += alpha * (g - a.A)
	if !a.sSet { // salience immutable at birth (knob 1 default)
		a.S, a.sSet = stress, true
	}
}

// rectSoft maps signed alignment to a non-negative weight; soft with a floor, so even a
// fully-failed edge (A=−1) stays barely alive (softFloor) and can recover if later
// graded up — instead of one catastrophic grade zeroing it forever (knob 3 default).
func rectSoft(x float64) float64 { return math.Max(softFloor, (x+1)/2) }

// conf is a mild exploration bonus for low-N edges (knob 2 default): 2 at N=0 → 1 as
// N→∞, so unproven edges still get looked at and earn their alignment.
func (a *Activation) conf() float64 { return 1 + 1/(1+float64(a.N)) }

// ActivationStore holds edge-owned records keyed by edge identity (A→B), on the nucleus.
// The traversal reads Strength off it; the bump writes it. Mutex-guarded for now —
// lock-free/atomic reads are open decision #2, deferred.
type ActivationStore struct {
	mu sync.Mutex
	m  map[string]*Activation
}

func NewActivationStore() *ActivationStore { return &ActivationStore{m: map[string]*Activation{}} }

func (s *ActivationStore) ensure(edge string) *Activation {
	if a := s.m[edge]; a != nil {
		return a
	}
	a := newActivation()
	s.m[edge] = a
	return a
}

// Strength reads an edge's strength; ok=false if the edge has no record yet.
func (s *ActivationStore) Strength(edge string, tick int) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.m[edge]
	if a == nil {
		return 0, false
	}
	return a.Strength(tick), true
}

// Cross bumps an edge's frequency (get-or-create).
func (s *ActivationStore) Cross(edge string, tick int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(edge).Cross(tick)
}

// Grade bumps an edge's alignment/salience/count (get-or-create) — the credit write-back.
func (s *ActivationStore) Grade(edge string, g, stress float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensure(edge).Grade(g, stress)
}
