package dnav3

import (
	"sort"
	"strings"
)

// Ingest — the observe-wave seam (spec delta 6, the-observe-wave.md §Ground).
//
// The seam does not replace Run(target); it changes WHO SUPPLIES THE TARGET:
//
//	impulse observed → wave resolves → target derived
//	                   (class completion; the framing call on novelty)
//	                 → Run(target)        ← unchanged
//
// ObserveImpulse is the top half. It writes the impulse (LAW 1), runs the wave (zero
// inference), commits the readings, and derives the target from the primary class's
// MEASURED edges — the settled completion: what this kind of situation reliably ends
// with, counted, not authored.
//
// When precedent cannot supply a target (novel stitch, unmeasured class, cold start),
// Target comes back nil with the Resolution's flags attached — and AUTHORING STAYS
// WITH THE CALLER'S FRAMING CALL. That is the design, not a gap: only the model can
// judge that a first-time situation warrants wanting something, and its commit founds
// the precedent the next impulse reads for free. cmd/dnav3's harness-supplied targets
// are exactly this caller-authored path until a worker owns it.
type IngestResult struct {
	Resolution Resolution
	Target     []Triple // nil when precedent cannot supply one — author via framing
	Authored   bool     // false: read from precedent. true is the CALLER's to set.
}

// ObserveImpulse runs the observe wave over one decomposed impulse. Everything here
// is arithmetic; segmentation and indexes are rebuilt from the log each call — at
// walking-skeleton scale brute force is correct, and derived indexes are cheap to be
// wrong about by construction.
func (b *Being) ObserveImpulse(impulse []Triple, context []Triple) IngestResult {
	// 1. The impulse crosses the boundary — raw, the world's (LAW 1).
	var written []Observation
	var derivedFrom []string
	for _, t := range impulse {
		o := b.Observe(t, ProvenanceRaw)
		written = append(written, o)
		derivedFrom = append(derivedFrom, obsID(o))
	}

	// 2. Derived substrate: spans, classes, lens, resolution history, class stats.
	log := b.Obs.Log()
	spans := Spans(log, 0)
	classes := Classes(log, spans, 0)
	lens := Lens(classes)
	rx := BuildResolutionIndex(log)
	byClass := b.Expr.ClassEdges(log, spans, classes)

	// 3. The wave.
	roots := Roots(written, context, rx)
	res := Resolve(log, spans, classes, lens, roots, classStatsFn(byClass))

	// 4. Commit the readings — interpreted, counted confidence, evidence attached.
	//    These found the resolution index the NEXT impulse proposes from.
	for _, o := range res.Commit(derivedFrom) {
		b.ObserveInterpreted(o.Triple, o.Confidence, o.DerivedFrom)
	}

	// 5. The target, read off precedent: the primary class's measured completions.
	out := IngestResult{Resolution: res}
	if res.Novel || res.Primary < 0 {
		return out // the framing call's job now — see the type comment.
	}
	seen := map[string]bool{}
	for _, ed := range byClass[res.Primary] {
		if ed.N == 0 || ed.P < 0.5 {
			continue // unmeasured edges have no opinion; unreliable ones are not wants
		}
		if !seen[ed.Target] {
			seen[ed.Target] = true
			out.Target = append(out.Target, keyTriple(ed.Target))
		}
	}
	sort.Slice(out.Target, func(i, j int) bool { return out.Target[i].Key() < out.Target[j].Key() })
	return out
}

// ObserveInterpreted writes a reading — a claim wearing interpreted provenance, with
// its counted confidence and the raw evidence it was derived from. Same write path as
// every observation (LAW 1); fidelity stays capped by provenance.
func (b *Being) ObserveInterpreted(t Triple, confidence float64, derivedFrom []string) Observation {
	b.seq++
	o := Observation{
		Triple:      t,
		Provenance:  ProvenanceInterpreted,
		Confidence:  confidence,
		DerivedFrom: derivedFrom,
		At:          b.tick,
	}
	o.Subject, o.Predicate, o.Object = t.Subject, t.Predicate, t.Object
	b.Obs.Write(o)
	b.grade(o)
	return o
}

// classStatsFn adapts class-grouped edges into the wave's Born-rule hook: the counted
// reliability of a reading within a class is the best measured edge whose claim
// mentions the term.
func classStatsFn(byClass map[int][]*Edge) Stats {
	return func(class int, term string) (float64, int) {
		bestP, bestN := 0.0, 0
		for _, ed := range byClass[class] {
			if ed.N == 0 {
				continue
			}
			t := keyTriple(ed.Target)
			if t.Subject != term && t.Predicate != term && t.Object != term {
				continue
			}
			if bestN == 0 || ed.P > bestP {
				bestP, bestN = ed.P, ed.N
			}
		}
		return bestP, bestN
	}
}

// keyTriple inverts Triple.Key — the derived target is world state by address, so the
// key carries everything Run needs (mode and confidence are a stance, not the claim).
func keyTriple(key string) Triple {
	parts := strings.SplitN(key, "\x00", 3)
	t := Triple{Mode: ModeExact}
	if len(parts) > 0 {
		t.Subject = parts[0]
	}
	if len(parts) > 1 {
		t.Predicate = parts[1]
	}
	if len(parts) > 2 {
		t.Object = parts[2]
	}
	return t
}
