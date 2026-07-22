package dnav3

import "sort"

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
// Target is nil when measured precedent is absent OR when any source remains ambiguous or
// unresolved. TargetGroups preserves that distinction; Resolution.Novel is true only when no
// measured candidate group exists. Resolution remains as compatibility for the retired class-wave
// inspection types.
type IngestResult struct {
	Resolution     Resolution
	TargetGroups   []TargetGroup         // structured, source-owned candidates; order has no winner semantics
	Target         []Triple              // compatibility projection; nil unless every group is confident
	Interpretation InterpretationHandoff // four-route express handoff; no extra inference call
	Authored       bool                  // false: read from precedent. true is the CALLER's to set.
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
	groups := b.deriveActivatedTarget()
	handoff := buildInterpretationHandoff(groups)
	b.Interpretation = handoff
	out := IngestResult{
		TargetGroups:   groups,
		Target:         projectConfidentTarget(groups),
		Interpretation: handoff,
	}
	out.Resolution = Resolution{Primary: -1, Novel: len(groups) == 0}
	return out
}

func projectConfidentTarget(groups []TargetGroup) []Triple {
	if len(groups) == 0 {
		return nil
	}
	byKey := map[string]Triple{}
	for _, group := range groups {
		if group.State != SourceConfident || group.Committed < 0 || group.Committed >= len(group.Branches) {
			return nil
		}
		for _, completion := range group.Branches[group.Committed].Completions {
			byKey[completion.Triple.Key()] = completion.Triple
		}
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]Triple, len(keys))
	for i, key := range keys {
		out[i] = byKey[key]
	}
	return out
}

// ObserveInterpreted writes a decomposed reading with interpreted provenance. It is not
// automatically an admitted InterpretationClaim: durable learning additionally requires
// ProposeInterpretation's own context prediction and licensed-consequence reference.
// Same observation write path as every reading (LAW 1); fidelity stays provenance-capped.
func (b *Being) ObserveInterpreted(t Triple, confidence float64, derivedFrom []string) Observation {
	o := b.appendObservation(t, ProvenanceInterpreted, confidence, derivedFrom, b.ensureSpan())
	b.activateObservation(o)
	b.grade(o)
	return o
}
