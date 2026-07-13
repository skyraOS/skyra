package dnav2

import (
	"math"
	"sync"
)

// The signed edge store — the minimal loop's learned state
// (minimal-aitia-set.md "Edge Learning And Defeat"). One record per contextual
// edge (keyed by EdgeRecord.ID). Signed grades in [-1,1] sample two EMAs;
// strength is their mean and stays SIGNED: polarity is meaning (resonance /
// tension), absolute magnitude is the transport claim. Distinct from the
// existing ActivationStore (A/S/F/N) by design — new model beside old, not a
// mutation of it.
//
// Invariants carried: no grade is NOT grade zero (absent record ≠ zero record);
// only Familiarity decays lazily without new grades; Identity holds; defeat is
// an edge trending to zero magnitude, never a node property.

const (
	signedAlphaIdentity    = 0.02 // slow — what this relationship persistently is
	signedAlphaFamiliarity = 0.10 // fast — the recent charge
	signedDecay            = 0.98 // per-tick lazy Familiarity decay (tuning parameter)
)

// SignedEdge is one edge's learned state.
type SignedEdge struct {
	Identity    float64
	Familiarity float64
	FLast       int // tick of the last grade (decay reference)
	GradeCount  int // grades seen — interprets zero (untested vs defeated vs contested)
}

// readF lazily decays Familiarity to `tick` without mutating.
func (e *SignedEdge) readF(tick int) float64 {
	if d := tick - e.FLast; d > 0 {
		return e.Familiarity * math.Pow(signedDecay, float64(d))
	}
	return e.Familiarity
}

// grade samples a signed grade into both EMAs (clamped to [-1,1]). Familiarity
// decays to now first, then samples — elapsed silence has already faded the
// recent charge.
func (e *SignedEdge) grade(g float64, tick int) {
	g = math.Max(-1, math.Min(1, g))
	e.Familiarity = e.readF(tick)
	e.Identity += signedAlphaIdentity * (g - e.Identity)
	e.Familiarity += signedAlphaFamiliarity * (g - e.Familiarity)
	e.FLast = tick
	e.GradeCount++
}

// strength is the signed read: (Identity + decayed Familiarity) / 2.
func (e *SignedEdge) strength(tick int) float64 {
	return (e.Identity + e.readF(tick)) / 2
}

// SignedStore holds the edge records, on the nucleus.
type SignedStore struct {
	mu sync.Mutex
	m  map[string]*SignedEdge
}

func NewSignedStore() *SignedStore { return &SignedStore{m: map[string]*SignedEdge{}} }

// Grade records one signed grade on an edge (get-or-create — the first grade
// births the record).
func (s *SignedStore) Grade(edgeID string, g float64, tick int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.m[edgeID]
	if e == nil {
		e = &SignedEdge{}
		s.m[edgeID] = e
	}
	e.grade(g, tick)
}

// Strength reads an edge's SIGNED strength; ok=false when the edge has never
// been graded (no grade ≠ grade zero).
func (s *SignedStore) Strength(edgeID string, tick int) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e := s.m[edgeID]
	if e == nil {
		return 0, false
	}
	return e.strength(tick), true
}

// Polarity and Magnitude split a signed strength for transport: polarity is
// meaning, |magnitude| is the traversal claim
// (proportional-amplitude-allocation.md).
func Polarity(strength float64) float64 {
	switch {
	case strength > 0:
		return 1
	case strength < 0:
		return -1
	}
	return 0
}

func Magnitude(strength float64) float64 { return math.Abs(strength) }

// AllocateAmplitude distributes only by nonzero absolute signed strength.
// Polarity remains edge meaning; magnitude is excitation. Zero and absent
// alignments receive nothing. If total magnitude is zero, amplitude remains
// undistributed rather than inventing an exploration floor.
func AllocateAmplitude(remaining float64, strengths map[string]float64) map[string]float64 {
	out := map[string]float64{}
	if remaining <= 0 {
		return out
	}
	total := 0.0
	for _, strength := range strengths {
		total += Magnitude(strength)
	}
	if total == 0 {
		return out
	}
	for edgeID, strength := range strengths {
		magnitude := Magnitude(strength)
		if magnitude > 0 {
			out[edgeID] = remaining * magnitude / total
		}
	}
	return out
}

// GradeCount reads an edge's grade count (0 when ungraded) — the zero-
// interpretation input (untested vs defeated vs contested).
func (s *SignedStore) GradeCount(edgeID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.m[edgeID]; e != nil {
		return e.GradeCount
	}
	return 0
}

// All returns a copy of every record (inspection / the Praxis screen).
func (s *SignedStore) All() map[string]SignedEdge {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]SignedEdge, len(s.m))
	for k, e := range s.m {
		out[k] = *e
	}
	return out
}
