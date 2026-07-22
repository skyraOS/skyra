package dnav3

import (
	"reflect"
	"testing"
)

func phase1Completion(source, object string, heat float64) Completion {
	triple := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "completion", Object: object}
	return Completion{Triple: triple, EdgeKey: source + "\x00" + triple.Key(), Heat: heat}
}

func phase1Lookup(spans map[string][]SpanID) settlementSpanLookup {
	return func(edgeKey string, outcome Truth) []SpanID {
		if outcome != TruthTrue {
			return nil
		}
		return append([]SpanID(nil), spans[edgeKey]...)
	}
}

func TestClassifySourceSingleAndConjunctiveBranchesAreConfident(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation
	a := phase1Completion("S", "A", 0.4)
	b := phase1Completion("S", "B", 0.3)

	single := classifySource("S", []Completion{a}, nil, cfg)
	if single.State != SourceConfident || single.Committed != 0 || single.Conjunction {
		t.Fatalf("single branch = %+v, want confident singleton", single)
	}

	conjunction := classifySource("S", []Completion{b, a}, phase1Lookup(map[string][]SpanID{
		a.EdgeKey: {"s1", "s2"},
		b.EdgeKey: {"s1", "s2"},
	}), cfg)
	if conjunction.State != SourceConfident || conjunction.Committed != 0 || !conjunction.Conjunction {
		t.Fatalf("one conjunctive branch = %+v, want confident+conjunction", conjunction)
	}
	if len(conjunction.Branches) != 1 || len(conjunction.Branches[0].Completions) != 2 {
		t.Fatalf("conjunctive partition = %+v", conjunction.Branches)
	}
}

func TestClassifySourceToy6TwoCompletionFixtures(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation
	tests := []struct {
		name        string
		aSpans      []SpanID
		bSpans      []SpanID
		aHeat       float64
		bHeat       float64
		state       SourceState
		branches    int
		conjunction bool
	}{
		{name: "AMB", aSpans: []SpanID{"p1", "p2"}, bSpans: []SpanID{"p3", "p4"}, aHeat: .4, bHeat: .4, state: SourceAmbiguous, branches: 2},
		{name: "AMB-NOISE", aSpans: []SpanID{"p1", "p2", "p3", "x"}, bSpans: []SpanID{"p4", "p5", "p6", "x"}, aHeat: .4, bHeat: .4, state: SourceAmbiguous, branches: 2},
		{name: "CONJ", aSpans: []SpanID{"q1", "q2", "q3"}, bSpans: []SpanID{"q1", "q2", "q3"}, aHeat: .4, bHeat: .4, state: SourceConfident, branches: 1, conjunction: true},
		{name: "CONJ-WEAK", aSpans: []SpanID{"q1", "q2", "q3", "q4", "q5", "q6", "q7", "q8"}, bSpans: []SpanID{"q1", "q2", "q3", "q4", "q5", "q6", "q7", "q8", "q9", "q10"}, aHeat: .4, bHeat: .4, state: SourceConfident, branches: 1, conjunction: true},
		{name: "AMB-EDGE", aSpans: []SpanID{"s1", "s2", "s3"}, bSpans: []SpanID{"s3", "s4"}, aHeat: .4, bHeat: .4, state: SourceConfident, branches: 1, conjunction: true},
		{name: "CONF", aSpans: []SpanID{"r1", "r2"}, bSpans: []SpanID{"r3"}, aHeat: .51, bHeat: .16, state: SourceConfident, branches: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := phase1Completion("S", "A", tt.aHeat)
			b := phase1Completion("S", "B", tt.bHeat)
			group := classifySource("S", []Completion{a, b}, phase1Lookup(map[string][]SpanID{
				a.EdgeKey: tt.aSpans,
				b.EdgeKey: tt.bSpans,
			}), cfg)
			if group.State != tt.state || len(group.Branches) != tt.branches || group.Conjunction != tt.conjunction {
				t.Fatalf("%s = %+v, want state=%s branches=%d conjunction=%v", tt.name, group, tt.state, tt.branches, tt.conjunction)
			}
		})
	}
}

func TestClassifySourceIgnoresOtherSourcesSettlementSpans(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation
	a := phase1Completion("S", "A", .4)
	b := phase1Completion("S", "B", .4)
	tA := phase1Completion("T", "A", .4)
	tB := phase1Completion("T", "B", .4)
	group := classifySource("S", []Completion{a, b}, phase1Lookup(map[string][]SpanID{
		a.EdgeKey:  {"s1", "s2"},
		b.EdgeKey:  {"s3", "s4"},
		tA.EdgeKey: {"x1", "x2"},
		tB.EdgeKey: {"x1", "x2"},
	}), cfg)
	if group.State != SourceAmbiguous || len(group.Branches) != 2 {
		t.Fatalf("source T contaminated source S partition: %+v", group)
	}
}

func TestClassifySourceMixedBranchAndThreeWayFork(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation
	a := phase1Completion("S", "A", 0.4)
	b := phase1Completion("S", "B", 0.4)
	c := phase1Completion("S", "C", 0.4)

	mixed := classifySource("S", []Completion{c, b, a}, phase1Lookup(map[string][]SpanID{
		a.EdgeKey: {"s1", "s2", "s3"},
		b.EdgeKey: {"s1", "s2", "s3"},
		c.EdgeKey: {"s4", "s5"},
	}), cfg)
	if mixed.State != SourceAmbiguous || !mixed.Conjunction || len(mixed.Branches) != 2 {
		t.Fatalf("mixed branch state = %+v, want conjunction-within and ambiguity-between", mixed)
	}

	fork := classifySource("S", []Completion{a, b, c}, phase1Lookup(map[string][]SpanID{
		a.EdgeKey: {"s1", "s2"},
		b.EdgeKey: {"s3", "s4"},
		c.EdgeKey: {"s5", "s6"},
	}), cfg)
	if fork.State != SourceAmbiguous || fork.Conjunction || len(fork.Branches) != 3 {
		t.Fatalf("three-way fork = %+v, want three ambiguous singleton branches", fork)
	}
}

func TestClassifySourceBridgeIsUnresolvedAndPermutationInvariant(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation
	a := phase1Completion("S", "A", 0.4)
	b := phase1Completion("S", "B", 0.4)
	c := phase1Completion("S", "C", 0.4)
	d := phase1Completion("S", "D", 0.5)
	lookup := phase1Lookup(map[string][]SpanID{
		a.EdgeKey: {"s1", "s2", "s3"},
		b.EdgeKey: {"s2", "s3", "s4"},
		c.EdgeKey: {"s4", "s5", "s6"},
		d.EdgeKey: {"s7", "s8"},
	})

	forward := classifySource("S", []Completion{a, b, c, d}, lookup, cfg)
	reverse := classifySource("S", []Completion{d, c, b, a}, lookup, cfg)
	if forward.State != SourceUnresolved || forward.Committed != -1 {
		t.Fatalf("bridge beside dominant branch = %+v, want unresolved veto", forward)
	}
	if len(forward.Unresolved) != 3 || len(forward.Branches) != 1 {
		t.Fatalf("bridge partition = %+v unresolved=%+v", forward.Branches, forward.Unresolved)
	}
	if !reflect.DeepEqual(forward, reverse) {
		t.Fatalf("partition depends on insertion order:\nforward=%+v\nreverse=%+v", forward, reverse)
	}
}

func TestClassifySourceMissingPositiveProvenanceIsUnresolved(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation
	a := phase1Completion("S", "A", 0.4)
	b := phase1Completion("S", "B", 0.4)
	group := classifySource("S", []Completion{a, b}, phase1Lookup(map[string][]SpanID{
		a.EdgeKey: {"s1"},
		// B has measured edge state but no indexed positive settlement provenance.
	}), cfg)
	if group.State != SourceUnresolved || group.Committed != -1 || len(group.Unresolved) != 1 {
		t.Fatalf("missing positive provenance = %+v, want unresolved", group)
	}
}

func TestSettlementSpanIndexSeparatesPolarity(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.ensureSpan()
	a := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "value", Object: "A"}
	bValue := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "value", Object: "B"}
	e := &Expression{
		EpisodeID: b.EpisodeID(),
		SpanID:    span,
		Acts:      []Act{{ID: "a1", Primitive: "choose", Target: "slot:test"}},
		Expects: []Expect{
			{Of: []string{"a1"}, Armed: true, Triple: a},
			{Of: []string{"a1"}, Armed: true, Triple: bValue},
		},
	}
	b.Expr.Place("polarity", e)
	b.Observe(bValue, ProvenanceRaw)

	am := actMap(e)
	aKey := edgeKey(am, e.Expects[0].Of, a.Key())
	bKey := edgeKey(am, e.Expects[1].Of, bValue.Key())
	if got := b.Expr.SettlementSpans(aKey, TruthTrue); len(got) != 0 {
		t.Fatalf("false A settlement entered positive spans: %v", got)
	}
	if got := b.Expr.SettlementSpans(aKey, TruthFalse); !reflect.DeepEqual(got, []SpanID{span}) {
		t.Fatalf("A false spans=%v, want [%s]", got, span)
	}
	if got := b.Expr.SettlementSpans(bKey, TruthTrue); !reflect.DeepEqual(got, []SpanID{span}) {
		t.Fatalf("B true spans=%v, want [%s]", got, span)
	}
}

func TestObserveImpulsePreservesBalancedSourceAndWithholdsLegacyTarget(t *testing.T) {
	b := NewBeing(nil, nil)
	a := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "completion", Object: "A"}
	bValue := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "completion", Object: "B"}
	root := func(object string) Triple {
		return Triple{Mode: ModeExact, Subject: "topic:shared", Predicate: "mentions", Object: object}
	}
	seed := func(id string, completion Triple) {
		b.ObserveImpulse([]Triple{root(id), completion}, nil)
		span := b.CurrentSpanID()
		e := &Expression{
			EpisodeID: b.EpisodeID(), SpanID: span,
			Acts:    []Act{{ID: "a1", Primitive: "complete", Target: "slot:test"}},
			Expects: []Expect{{Of: []string{"a1"}, Triple: completion}},
		}
		b.Expr.Place(id, e)
		edge := b.Expr.Edge(id, e.Expects[0])
		edge.Observe(TruthTrue, b.tick)
		edge.Observe(TruthTrue, b.tick)
		b.Expr.RecordSettlement(id, e.Expects[0], TruthTrue, span)
	}
	seed("history-a", a)
	seed("history-b", bValue)
	// This fixture tests preserved ambiguity, not the still-experimental placement of the
	// dominance cutover. Keep both measured branches inside the configured no-dominance band.
	b.Activation.Config.Differentiation.DominanceMargin = 0.2

	result := b.ObserveImpulse([]Triple{root("now")}, nil)
	if result.Resolution.Novel {
		t.Fatal("balanced measured precedent was mislabeled novel")
	}
	if result.Target != nil {
		t.Fatalf("ambiguous source leaked into legacy target: %+v groups=%+v", result.Target, result.TargetGroups)
	}
	if len(result.TargetGroups) != 1 {
		t.Fatalf("target groups=%+v, want one source group", result.TargetGroups)
	}
	group := result.TargetGroups[0]
	if group.State != SourceAmbiguous || len(group.Branches) != 2 || group.Committed != -1 {
		t.Fatalf("balanced group=%+v, want two preserved ambiguous branches", group)
	}
	if result.Interpretation.Route != InterpretationAmbiguous || result.Interpretation.EffectiveN != 2 {
		t.Fatalf("express handoff=%+v, want ambiguous counted support N=2", result.Interpretation)
	}
	if b.Interpretation.Route != InterpretationAmbiguous {
		t.Fatalf("current-span handoff was not installed on being: %+v", b.Interpretation)
	}
}

func TestLegacyTargetRequiresEveryGroupConfident(t *testing.T) {
	a := phase1Completion("S1", "A", 0.5)
	b := phase1Completion("S2", "B", 0.5)
	confident := []TargetGroup{
		{Source: "S1", State: SourceConfident, Committed: 0, Branches: []Branch{{Completions: []Completion{a}, Heat: a.Heat}}},
		{Source: "S2", State: SourceConfident, Committed: 0, Branches: []Branch{{Completions: []Completion{b}, Heat: b.Heat}}},
	}
	if got := projectConfidentTarget(confident); len(got) != 2 {
		t.Fatalf("two confident sources projected %v, want both completions", got)
	}
	confident[1].State = SourceAmbiguous
	confident[1].Committed = -1
	if got := projectConfidentTarget(confident); got != nil {
		t.Fatalf("partial interpretation leaked into legacy target: %v", got)
	}
}
