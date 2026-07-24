package dnav3

import (
	"math"
	"sort"
	"strings"
)

type ActivationConfig struct {
	Lambda          float64
	Epsilon         float64
	SpanFloor       float64
	HopLimit        int
	MemoryTokens    int
	Differentiation DifferentiationConfig
}

// DifferentiationConfig keeps related separation decisions together without pretending that
// overlap, heat dominance, density valleys, and promotion evidence share one scalar or unit.
type DifferentiationConfig struct {
	ConjunctionOverlap float64
	DominanceMargin    float64
	SplitValleyDepth   float64
	PromotionEvidence  float64
}

func DefaultActivationConfig() ActivationConfig {
	return ActivationConfig{
		Lambda:       0.5,
		Epsilon:      0.05,
		SpanFloor:    0.05,
		HopLimit:     2,
		MemoryTokens: 2048,
		Differentiation: DifferentiationConfig{
			ConjunctionOverlap: 0.2,
			DominanceMargin:    0.1,
			SplitValleyDepth:   0.2,
			PromotionEvidence:  1,
		},
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
	defaults := DefaultActivationConfig().Differentiation
	if config.Differentiation.ConjunctionOverlap <= 0 || config.Differentiation.ConjunctionOverlap > 1 {
		config.Differentiation.ConjunctionOverlap = defaults.ConjunctionOverlap
	}
	if config.Differentiation.DominanceMargin <= 0 {
		config.Differentiation.DominanceMargin = defaults.DominanceMargin
	}
	if config.Differentiation.SplitValleyDepth <= 0 {
		config.Differentiation.SplitValleyDepth = defaults.SplitValleyDepth
	}
	if config.Differentiation.PromotionEvidence <= 0 {
		config.Differentiation.PromotionEvidence = defaults.PromotionEvidence
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

type SourceState string

const (
	SourceConfident  SourceState = "confident"
	SourceAmbiguous  SourceState = "ambiguous"
	SourceUnresolved SourceState = "unresolved"
)

type Completion struct {
	Triple  Triple
	EdgeKey string
	Heat    float64
	Trials  int // settled trials on the causal edge; Phase 2 uses this as counted support
}

type Branch struct {
	Completions []Completion
	Heat        float64
	Pocket      *DensityPocket // fitted only after Phase 3's density valley earns a split
}

type TargetGroup struct {
	Source      string
	Branches    []Branch
	Unresolved  []Completion
	State       SourceState
	Conjunction bool
	Committed   int // branch index when State is confident; -1 otherwise

	// Phase 3 readings. RawValley is topology alone; ValleyDepth includes recurrence
	// maturity. Differentiated means every branch earned a diagnostic pocket name.
	RawValley        float64
	ValleyDepth      float64
	Differentiated   bool
	SelectedByPocket bool // true only after a separately validated promotion gate authorizes commitment
	PocketMatch      int  // provisional unique predicate match; descriptive only, -1 when none/ambiguous
}

type settlementSpanLookup func(edgeKey string, outcome Truth) []SpanID

func (b *Being) deriveActivatedTarget() []TargetGroup {
	if b.Activation == nil {
		return nil
	}
	bySource := map[string]map[string]Completion{}
	for _, source := range b.Activation.RankedSpans() {
		for _, e := range b.Expr.ExpressionsInSpan(source.SpanID) {
			am := actMap(e)
			for _, x := range e.Expects {
				ek := edgeKey(am, x.Of, x.Key())
				sk := edgeSourceKey(am, x.Of)
				ed := b.Expr.Edge(e.ID, x)
				if ed.N == 0 || ed.P < 0.5 {
					continue
				}
				heat := b.Activation.Edges[ek]
				if heat < b.Activation.Config.Epsilon {
					continue
				}
				if bySource[sk] == nil {
					bySource[sk] = map[string]Completion{}
				}
				if old, ok := bySource[sk][ek]; !ok || heat > old.Heat {
					bySource[sk][ek] = Completion{Triple: x.Triple, EdgeKey: ek, Heat: heat, Trials: ed.N}
				}
			}
		}
	}
	sources := make([]string, 0, len(bySource))
	for source := range bySource {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	groups := make([]TargetGroup, 0, len(sources))
	active := b.Activation.RankedSpans()
	activeSpans := make([]SpanID, len(active))
	for i, span := range active {
		activeSpans[i] = span.SpanID
	}
	for _, source := range sources {
		completions := make([]Completion, 0, len(bySource[source]))
		for _, completion := range bySource[source] {
			completions = append(completions, completion)
		}
		group := classifySource(source, completions, b.Expr.SettlementSpans,
			b.Activation.Config.Differentiation)
		group = differentiateSource(group, b.Expr.SettlementSourceInSpan, b.Spans.Observations,
			activeSpans, b.currentSpan, b.Activation.Config.Differentiation)
		groups = append(groups, group)
	}
	return groups
}

func classifySource(source string, completions []Completion, spans settlementSpanLookup, cfg DifferentiationConfig) TargetGroup {
	defaults := DefaultActivationConfig().Differentiation
	if cfg.ConjunctionOverlap <= 0 || cfg.ConjunctionOverlap > 1 {
		cfg.ConjunctionOverlap = defaults.ConjunctionOverlap
	}
	if cfg.DominanceMargin <= 0 {
		cfg.DominanceMargin = defaults.DominanceMargin
	}
	completions = append([]Completion(nil), completions...)
	sort.Slice(completions, func(i, j int) bool {
		if completions[i].Triple.Key() != completions[j].Triple.Key() {
			return completions[i].Triple.Key() < completions[j].Triple.Key()
		}
		return completions[i].EdgeKey < completions[j].EdgeKey
	})
	g := TargetGroup{Source: source, Committed: -1, PocketMatch: -1}
	if len(completions) == 0 {
		return g
	}
	if len(completions) == 1 {
		branch := Branch{Completions: append([]Completion(nil), completions...), Heat: completions[0].Heat}
		g.Branches = []Branch{branch}
		g.State, g.Committed = SourceConfident, 0
		return g
	}

	adjacent := make([][]bool, len(completions))
	positive := make([][]SpanID, len(completions))
	missing := make([]bool, len(completions))
	for i := range completions {
		positive[i] = lookupSpans(spans, completions[i].EdgeKey)
		missing[i] = len(positive[i]) == 0
	}
	for i := range adjacent {
		adjacent[i] = make([]bool, len(completions))
		adjacent[i][i] = true
		for j := 0; j < i; j++ {
			if !missing[i] && !missing[j] && jaccardSpans(positive[i], positive[j]) >= cfg.ConjunctionOverlap {
				adjacent[i][j], adjacent[j][i] = true, true
			}
		}
	}

	visited := make([]bool, len(completions))
	for start := range completions {
		if visited[start] {
			continue
		}
		if missing[start] {
			visited[start] = true
			g.Unresolved = append(g.Unresolved, completions[start])
			continue
		}
		queue := []int{start}
		visited[start] = true
		var component []int
		for len(queue) > 0 {
			i := queue[0]
			queue = queue[1:]
			component = append(component, i)
			for j := range completions {
				if adjacent[i][j] && !visited[j] {
					visited[j] = true
					queue = append(queue, j)
				}
			}
		}
		clique := true
		for _, i := range component {
			for _, j := range component {
				clique = clique && adjacent[i][j]
			}
		}
		if !clique {
			for _, i := range component {
				g.Unresolved = append(g.Unresolved, completions[i])
			}
			continue
		}
		branch := Branch{}
		for _, i := range component {
			branch.Completions = append(branch.Completions, completions[i])
			branch.Heat = math.Max(branch.Heat, completions[i].Heat)
		}
		g.Conjunction = g.Conjunction || len(branch.Completions) > 1
		g.Branches = append(g.Branches, branch)
	}

	sort.Slice(g.Branches, func(i, j int) bool { return branchKey(g.Branches[i]) < branchKey(g.Branches[j]) })
	if len(g.Unresolved) > 0 {
		g.State = SourceUnresolved
		return g
	}
	if len(g.Branches) == 1 {
		g.State, g.Committed = SourceConfident, 0
		return g
	}
	best, second := -1, -1
	for i := range g.Branches {
		if best < 0 || g.Branches[i].Heat > g.Branches[best].Heat {
			second, best = best, i
		} else if second < 0 || g.Branches[i].Heat > g.Branches[second].Heat {
			second = i
		}
	}
	if best >= 0 && second >= 0 && g.Branches[best].Heat-g.Branches[second].Heat >= cfg.DominanceMargin {
		g.State, g.Committed = SourceConfident, best
	} else {
		g.State = SourceAmbiguous
	}
	return g
}

func lookupSpans(lookup settlementSpanLookup, edgeKey string) []SpanID {
	if lookup == nil {
		return nil
	}
	return lookup(edgeKey, TruthTrue)
}

func jaccardSpans(a, b []SpanID) float64 {
	left := make(map[SpanID]struct{}, len(a))
	union := make(map[SpanID]struct{}, len(a)+len(b))
	for _, span := range a {
		left[span], union[span] = struct{}{}, struct{}{}
	}
	intersection := 0
	for _, span := range b {
		if _, ok := left[span]; ok {
			intersection++
		}
		union[span] = struct{}{}
	}
	if len(union) == 0 {
		return 0
	}
	return float64(intersection) / float64(len(union))
}

func branchKey(branch Branch) string {
	keys := make([]string, len(branch.Completions))
	for i, completion := range branch.Completions {
		keys[i] = completion.Triple.Key()
	}
	return strings.Join(keys, "\x00branch\x00")
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
