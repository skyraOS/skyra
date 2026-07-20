package dnav3

import (
	"math"
	"sort"
)

type ActivationConfig struct {
	Lambda       float64
	Epsilon      float64
	SpanFloor    float64
	HopLimit     int
	MemoryTokens int
}

func DefaultActivationConfig() ActivationConfig {
	return ActivationConfig{
		Lambda:       0.5,
		Epsilon:      0.05,
		SpanFloor:    0.05,
		HopLimit:     2,
		MemoryTokens: 2048,
	}
}

// ActivationField is transient working memory. It is not a log and carries no
// truth statistic. BeginImpulse continuously decays the old field; bumps from the
// new impulse and causal traversal replace it with what matters now.
type ActivationField struct {
	Config ActivationConfig
	Tick   int

	Entities map[string]float64
	Triples  map[string]float64
	Spans    map[SpanID]float64
	Edges    map[string]float64
	Nodes    map[string]float64
}

func NewActivationField(config ActivationConfig) *ActivationField {
	if config.Lambda <= 0 || config.Lambda >= 1 {
		config.Lambda = 0.5
	}
	if config.Epsilon <= 0 {
		config.Epsilon = 0.05
	}
	if config.SpanFloor <= 0 {
		config.SpanFloor = config.Epsilon
	}
	if config.HopLimit <= 0 {
		config.HopLimit = 2
	}
	if config.MemoryTokens <= 0 {
		config.MemoryTokens = 2048
	}
	return &ActivationField{
		Config:   config,
		Entities: map[string]float64{},
		Triples:  map[string]float64{},
		Spans:    map[SpanID]float64{},
		Edges:    map[string]float64{},
		Nodes:    map[string]float64{},
	}
}

func (f *ActivationField) BeginImpulse() {
	if f == nil {
		return
	}
	f.Tick++
	decayStrings(f.Entities, f.Config.Lambda, f.Config.Epsilon)
	decayStrings(f.Triples, f.Config.Lambda, f.Config.Epsilon)
	decaySpans(f.Spans, f.Config.Lambda, f.Config.Epsilon)
	decayStrings(f.Edges, f.Config.Lambda, f.Config.Epsilon)
	decayStrings(f.Nodes, f.Config.Lambda, f.Config.Epsilon)
}

func (f *ActivationField) BumpEntity(id string, heat float64) {
	if f != nil && id != "" {
		f.Entities[id] = bump(f.Entities[id], heat)
	}
}

func (f *ActivationField) BumpTriple(key string, heat float64) {
	if f != nil && key != "" {
		f.Triples[key] = bump(f.Triples[key], heat)
	}
}

func (f *ActivationField) BumpSpan(id SpanID, heat float64) {
	if f != nil && id != "" {
		f.Spans[id] = bump(f.Spans[id], heat)
	}
}

func (f *ActivationField) BumpEdge(key string, heat float64) {
	if f != nil && key != "" {
		f.Edges[key] = bump(f.Edges[key], heat)
	}
}

func (f *ActivationField) BumpNode(id string, heat float64) {
	if f != nil && id != "" {
		f.Nodes[id] = bump(f.Nodes[id], heat)
	}
}

func (f *ActivationField) SpanActive(id SpanID) bool {
	return f != nil && f.Spans[id] >= f.Config.SpanFloor
}

func (f *ActivationField) RankedSpans() []SpanActivation {
	if f == nil {
		return nil
	}
	out := make([]SpanActivation, 0, len(f.Spans))
	for id, a := range f.Spans {
		if a >= f.Config.Epsilon {
			out = append(out, SpanActivation{SpanID: id, Activation: a})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Activation != out[j].Activation {
			return out[i].Activation > out[j].Activation
		}
		return out[i].SpanID < out[j].SpanID
	})
	return out
}

type SpanActivation struct {
	SpanID     SpanID
	Activation float64
}

// activateImpulse is the continuous-replacement step. Decay happens exactly once
// per external impulse; causal descendants and planning queries bump the same field
// without advancing it.
func (b *Being) activateImpulse(written []Observation, context []Triple) {
	if b.Activation == nil {
		return
	}
	b.Activation.BeginImpulse()
	for _, o := range written {
		b.bumpObservation(o, true)
	}
	for _, t := range context {
		b.heatTriple(t, false)
	}
	b.propagateActivation()
}

func (b *Being) activateObservation(o Observation) {
	if b.Activation == nil {
		return
	}
	b.bumpObservation(o, true)
	b.propagateActivation()
}

func (b *Being) bumpObservation(o Observation, retrieve bool) {
	b.Activation.BumpSpan(o.SpanID, 1)
	b.Activation.BumpNode(string(o.ID), 1)
	for _, term := range []string{o.Subject, o.Predicate, o.Object} {
		b.Activation.BumpEntity(term, 1)
		if retrieve {
			b.heatPosting(b.Spans.Entity(term), 1, func(candidate Observation) bool {
				return candidate.Subject == term || candidate.Predicate == term || candidate.Object == term
			})
		}
	}
	b.Activation.BumpTriple(o.Key(), 1)
	if retrieve {
		b.heatPosting(b.Spans.Address(o.Triple), 0.5, func(candidate Observation) bool {
			return sp(candidate.Triple) == sp(o.Triple)
		})
		b.heatPosting(b.Spans.Triple(o.Triple), 0.5, func(candidate Observation) bool {
			return candidate.Key() == o.Key()
		})
	}
}

func (b *Being) heatTriple(t Triple, propagate bool) {
	if b.Activation == nil {
		return
	}
	for _, term := range []string{t.Subject, t.Predicate, t.Object} {
		b.Activation.BumpEntity(term, 1)
		b.heatPosting(b.Spans.Entity(term), 1, func(candidate Observation) bool {
			return candidate.Subject == term || candidate.Predicate == term || candidate.Object == term
		})
	}
	b.Activation.BumpTriple(t.Key(), 1)
	b.heatPosting(b.Spans.Address(t), 0.5, func(candidate Observation) bool {
		return sp(candidate.Triple) == sp(t)
	})
	b.heatPosting(b.Spans.Triple(t), 0.5, func(candidate Observation) bool {
		return candidate.Key() == t.Key()
	})
	if propagate {
		b.propagateActivation()
	}
}

func (b *Being) heatPosting(spans []SpanID, amplitude float64, matches func(Observation) bool) bool {
	if b.Activation == nil || len(spans) == 0 {
		return false
	}
	share := amplitude / float64(len(spans))
	// A posting so broad that no member can clear the activation floor carries no
	// discriminating information. Do not scan an unbounded "said"/"state"/user list;
	// exact address/triple postings and rarer co-roots still retrieve the region.
	if share < b.Activation.Config.Epsilon {
		return false
	}
	changed := false
	for _, span := range spans {
		before := b.Activation.Spans[span]
		b.Activation.BumpSpan(span, share)
		changed = changed || b.Activation.Spans[span] > before
		var matched []Observation
		for _, o := range b.Spans.Observations(span) {
			if matches == nil || matches(o) {
				matched = append(matched, o)
			}
		}
		if len(matched) == 0 {
			continue
		}
		nodeShare := share / float64(len(matched))
		if nodeShare < b.Activation.Config.Epsilon {
			continue
		}
		for _, o := range matched {
			b.Activation.BumpNode(string(o.ID), nodeShare)
		}
	}
	return changed
}

// activateQuery makes target/holes ordinary roots. Exact precedent is allowed to
// recruit its birth span with measured retrieval strength, including across episodes;
// episode identity still prevents old predictions from grading in the current world.
func (b *Being) activateQuery(query []Triple) {
	if b.Activation == nil {
		return
	}
	for _, q := range query {
		b.heatTriple(q, false)
		for _, e := range b.Expr.Precedent(q, b.tick) {
			x := matching(e, q)
			ed := b.Expr.Edge(e.ID, x)
			heat := ed.RetrievalWeight(b.tick)
			if heat == 0 {
				heat = 0.1 // unmeasured precedent may be inspected, never treated as reliable.
			}
			b.Activation.BumpSpan(e.SpanID, math.Min(1, heat))
			b.Activation.BumpNode("expression:"+e.ID, math.Min(1, heat))
		}
	}
	b.propagateActivation()
}

// propagateActivation walks span -> expression -> measured causal edge -> target
// triple -> witnessed spans. It is bounded and deterministic; P/N are only read.
func (b *Being) propagateActivation() {
	if b.Activation == nil || b.Expr == nil || b.Spans == nil {
		return
	}
	visited := map[string]bool{}
	for hop := 0; hop < b.Activation.Config.HopLimit; hop++ {
		changed := false
		for _, source := range b.Activation.RankedSpans() {
			for _, e := range b.Expr.ExpressionsInSpan(source.SpanID) {
				b.Activation.BumpNode("expression:"+e.ID, source.Activation)
				am := actMap(e)
				for i := range e.Expects {
					x := e.Expects[i]
					ek := edgeKey(am, x.Of, x.Key())
					visit := string(source.SpanID) + "\x00" + ek
					if visited[visit] {
						continue
					}
					visited[visit] = true
					ed := b.Expr.Edge(e.ID, x)
					if ed.N == 0 {
						continue
					}
					crossing := source.Activation * ed.P * ed.confidence()
					if crossing < b.Activation.Config.Epsilon {
						continue
					}
					b.Activation.BumpEdge(ek, crossing)
					b.Activation.BumpTriple(x.Key(), crossing)
					changed = b.heatPosting(
						b.Spans.Triple(x.Triple),
						crossing,
						func(candidate Observation) bool { return candidate.Key() == x.Key() },
					) || changed
				}
			}
		}
		if !changed {
			return
		}
	}
}

func (b *Being) deriveActivatedTarget() []Triple {
	if b.Activation == nil {
		return nil
	}
	type candidate struct {
		t    Triple
		heat float64
	}
	byKey := map[string]candidate{}
	for _, source := range b.Activation.RankedSpans() {
		for _, e := range b.Expr.ExpressionsInSpan(source.SpanID) {
			am := actMap(e)
			for _, x := range e.Expects {
				ek := edgeKey(am, x.Of, x.Key())
				ed := b.Expr.Edge(e.ID, x)
				if ed.N == 0 || ed.P < 0.5 {
					continue
				}
				heat := b.Activation.Edges[ek]
				if heat < b.Activation.Config.Epsilon {
					continue
				}
				if old, ok := byKey[x.Key()]; !ok || heat > old.heat {
					byKey[x.Key()] = candidate{t: x.Triple, heat: heat}
				}
			}
		}
	}
	out := make([]candidate, 0, len(byKey))
	for _, c := range byKey {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].heat != out[j].heat {
			return out[i].heat > out[j].heat
		}
		return out[i].t.Key() < out[j].t.Key()
	})
	if len(out) == 0 {
		return nil
	}
	targets := make([]Triple, len(out))
	for i, c := range out {
		targets[i] = c.t
	}
	return targets
}

func bump(current, heat float64) float64 {
	if heat <= 0 {
		return current
	}
	return math.Min(1, current+heat)
}

func decayStrings(m map[string]float64, lambda, epsilon float64) {
	for k, a := range m {
		a *= lambda
		if a < epsilon {
			delete(m, k)
		} else {
			m[k] = a
		}
	}
}

func decaySpans(m map[SpanID]float64, lambda, epsilon float64) {
	for k, a := range m {
		a *= lambda
		if a < epsilon {
			delete(m, k)
		} else {
			m[k] = a
		}
	}
}
