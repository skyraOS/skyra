package dnav3

// Ingest — the continuous-replacement seam.
//
// The seam does not replace Run(target); it changes WHO SUPPLIES THE TARGET and
// which durable memory is visible:
//
//	impulse appended → activation decays/bumps → direct spans + causal edges
//	                 → bounded active memory → target derived
//	                 → Run(target)
//
// ObserveImpulse is the top half. It writes the impulse (LAW 1), updates the
// transient field with zero inference, grades eligible active-span predictions, and
// derives a target from activated measured causal edges.
//
// When measured precedent cannot supply a target, Target is nil and AUTHORING STAYS
// WITH FRAMING. Resolution remains in the result as compatibility for the retired
// class-wave inspection types; Primary is -1 and Novel records this target decision.
type IngestResult struct {
	Resolution Resolution
	Target     []Triple // nil when precedent cannot supply one — author via framing
	Authored   bool     // false: read from precedent. true is the CALLER's to set.
}

// ObserveImpulse runs continuous replacement over one decomposed impulse. Durable
// history is appended; transient activation decays, retrieves spans directly, and
// propagates through measured causal edges. No conversation buffer is accumulated
// and no compaction step exists.
func (b *Being) ObserveImpulse(impulse []Triple, context []Triple) IngestResult {
	// 1. The impulse crosses the boundary — raw, the world's (LAW 1).
	spanID := b.startExternalSpan()
	var written []Observation
	var derivedFrom []string
	for _, t := range impulse {
		o := b.appendObservation(t, ProvenanceRaw, 0, nil, spanID)
		written = append(written, o)
		derivedFrom = append(derivedFrom, obsID(o))
	}
	b.currentBornFrom = make([]ObservationID, 0, len(written))
	for _, o := range written {
		b.currentBornFrom = append(b.currentBornFrom, o.ID)
	}

	// The complete impulse is durable before activation and grading begin.
	b.activateImpulse(written, context)
	for _, o := range written {
		b.grade(o)
	}

	// A measured causal continuation can supply the target for free. With no reliable
	// continuation the situation is novel and framing authors what to pursue.
	out := IngestResult{Target: b.deriveActivatedTarget()}
	out.Resolution = Resolution{Primary: -1, Novel: len(out.Target) == 0}
	return out
}

// ObserveInterpreted writes a reading — a claim wearing interpreted provenance, with
// its counted confidence and the raw evidence it was derived from. Same write path as
// every observation (LAW 1); fidelity stays capped by provenance.
func (b *Being) ObserveInterpreted(t Triple, confidence float64, derivedFrom []string) Observation {
	o := b.appendObservation(t, ProvenanceInterpreted, confidence, derivedFrom, b.ensureSpan())
	b.activateObservation(o)
	b.grade(o)
	return o
}
