package dnav3

import (
	"math"
	"strings"
	"testing"
)

var (
	toy8Shared = Triple{Mode: ModeExact, Subject: "topic:shared", Predicate: "present", Object: "true"}
	toy8School = Triple{Mode: ModeExact, Subject: "context:/", Predicate: "kind", Object: "school"}
	toy8Skyra  = Triple{Mode: ModeExact, Subject: "context:/", Predicate: "kind", Object: "skyra"}
	toy8Red    = Triple{Mode: ModeExact, Subject: "context:/", Predicate: "color", Object: "red"}
	toy8Blue   = Triple{Mode: ModeExact, Subject: "context:/", Predicate: "color", Object: "blue"}
)

func toy8Observation(span SpanID, triple Triple, provenance Provenance) Observation {
	return Observation{SpanID: span, Triple: triple, Provenance: provenance}
}

func toy8SettlementMembership(spans map[string][]SpanID) (settlementSpanContains, []SpanID) {
	activeSet := map[SpanID]bool{}
	for _, ids := range spans {
		for _, id := range ids {
			activeSet[id] = true
		}
	}
	active := make([]SpanID, 0, len(activeSet))
	for id := range activeSet {
		active = append(active, id)
	}
	contains := func(edgeKey string, outcome Truth, span SpanID) bool {
		if outcome != TruthTrue {
			return false
		}
		for _, candidate := range spans[edgeKey] {
			if candidate == span {
				return true
			}
		}
		return false
	}
	return contains, active
}

func toy8ResolvableFixture(n int) (TargetGroup, settlementSpanContains, spanObservationLookup, []SpanID, SpanID) {
	a := phase2Completion("S", "A", .5, n)
	b := phase2Completion("S", "B", .5, n)
	settled := map[string][]SpanID{}
	observed := map[SpanID][]Observation{}
	for i := 0; i < n; i++ {
		aSpan := SpanID("a" + string(rune('1'+i)))
		bSpan := SpanID("b" + string(rune('1'+i)))
		settled[a.EdgeKey] = append(settled[a.EdgeKey], aSpan)
		settled[b.EdgeKey] = append(settled[b.EdgeKey], bSpan)
		observed[aSpan] = []Observation{
			toy8Observation(aSpan, toy8Shared, ProvenanceRaw),
			toy8Observation(aSpan, toy8School, ProvenanceRaw),
			toy8Observation(aSpan, a.Triple, ProvenanceRaw), // must be excluded from geometry
		}
		observed[bSpan] = []Observation{
			toy8Observation(bSpan, toy8Shared, ProvenanceRaw),
			toy8Observation(bSpan, toy8Skyra, ProvenanceRaw),
			toy8Observation(bSpan, b.Triple, ProvenanceRaw), // must be excluded from geometry
		}
	}
	current := SpanID("current")
	observed[current] = []Observation{
		toy8Observation(current, toy8Shared, ProvenanceRaw),
		toy8Observation(current, toy8School, ProvenanceRaw),
	}
	group := TargetGroup{
		Source: "S", State: SourceAmbiguous, Committed: -1,
		Branches: []Branch{phase2Branch(a), phase2Branch(b)},
	}
	contains, active := toy8SettlementMembership(settled)
	obs := func(span SpanID) []Observation { return append([]Observation(nil), observed[span]...) }
	return group, contains, obs, active, current
}

func TestToy8DensityOpensValleyBeforePredicateNaming(t *testing.T) {
	cfg := DefaultActivationConfig().Differentiation

	sparse, settlements, observations, active, current := toy8ResolvableFixture(2)
	sparse = differentiateSource(sparse, settlements, observations, active, current, cfg)
	if sparse.Differentiated || sparse.Branches[0].Pocket != nil || sparse.State != SourceAmbiguous {
		t.Fatalf("two occurrences minted a premature pocket: %+v", sparse)
	}
	if math.Abs(sparse.RawValley-(2.0/3.0)) > 1e-12 || sparse.ValleyDepth >= cfg.SplitValleyDepth {
		t.Fatalf("sparse valley=(raw %.6f effective %.6f), want 2/3 and below %.3f",
			sparse.RawValley, sparse.ValleyDepth, cfg.SplitValleyDepth)
	}

	dense, settlements, observations, active, current := toy8ResolvableFixture(3)
	dense = differentiateSource(dense, settlements, observations, active, current, cfg)
	if !dense.Differentiated || dense.SelectedByPocket || dense.PocketMatch != 0 ||
		dense.State != SourceAmbiguous || dense.Committed != -1 {
		t.Fatalf("three occurrences did not retain provisional school pocket safely: %+v", dense)
	}
	if math.Abs(dense.RawValley-(2.0/3.0)) > 1e-12 || math.Abs(dense.ValleyDepth-.25) > 1e-12 {
		t.Fatalf("dense valley=(raw %.6f effective %.6f), want (2/3, .25)", dense.RawValley, dense.ValleyDepth)
	}
	if got := dense.Branches[0].Pocket.Predicate; got.Triple.Key() != toy8School.Key() ||
		math.Abs(got.Coverage-1) > 1e-12 || math.Abs(got.BaseRate-.5) > 1e-12 ||
		math.Abs(got.Diagnosticity-math.Log(2)) > 1e-12 {
		t.Fatalf("A pocket predicate=%+v, want school with (1,.5,log2)", got)
	}
	if got := dense.Branches[1].Pocket.Predicate.Triple; got.Key() != toy8Skyra.Key() {
		t.Fatalf("B pocket predicate=%+v, want skyra", got)
	}
}

func TestToy8IrreducibleEqualDistributionNeverSplits(t *testing.T) {
	a := phase2Completion("S", "A", .5, 100)
	b := phase2Completion("S", "B", .5, 100)
	settled := map[string][]SpanID{}
	observed := map[SpanID][]Observation{}
	colors := []Triple{toy8Red, toy8Red, toy8Blue, toy8Blue}
	for i, color := range colors {
		for _, item := range []struct {
			prefix string
			c      Completion
		}{{"a", a}, {"b", b}} {
			span := SpanID(item.prefix + string(rune('1'+i)))
			settled[item.c.EdgeKey] = append(settled[item.c.EdgeKey], span)
			observed[span] = []Observation{
				toy8Observation(span, toy8Shared, ProvenanceRaw),
				toy8Observation(span, color, ProvenanceRaw),
				toy8Observation(span, item.c.Triple, ProvenanceRaw),
			}
		}
	}
	group := TargetGroup{Source: "S", State: SourceAmbiguous, Committed: -1,
		Branches: []Branch{phase2Branch(a), phase2Branch(b)}}
	contains, active := toy8SettlementMembership(settled)
	result := differentiateSource(group, contains, func(span SpanID) []Observation {
		return observed[span]
	}, active, "", DefaultActivationConfig().Differentiation)
	if result.RawValley != 0 || result.ValleyDepth != 0 || result.Differentiated || result.State != SourceAmbiguous {
		t.Fatalf("equal-distribution coin flip was falsely split: %+v", result)
	}
}

func TestToy8OutcomeAndInterpretedFeaturesCannotCreatePocket(t *testing.T) {
	a := phase2Completion("S", "A", .5, 20)
	b := phase2Completion("S", "B", .5, 20)
	settled := map[string][]SpanID{a.EdgeKey: {"a1", "a2", "a3"}, b.EdgeKey: {"b1", "b2", "b3"}}
	observed := map[SpanID][]Observation{}
	for _, span := range settled[a.EdgeKey] {
		observed[span] = []Observation{
			toy8Observation(span, a.Triple, ProvenanceRaw),
			toy8Observation(span, toy8School, ProvenanceInterpreted),
		}
	}
	for _, span := range settled[b.EdgeKey] {
		observed[span] = []Observation{
			toy8Observation(span, b.Triple, ProvenanceRaw),
			toy8Observation(span, toy8Skyra, ProvenanceMark),
		}
	}
	group := TargetGroup{Source: "S", State: SourceAmbiguous, Committed: -1,
		Branches: []Branch{phase2Branch(a), phase2Branch(b)}}
	contains, active := toy8SettlementMembership(settled)
	result := differentiateSource(group, contains, func(span SpanID) []Observation {
		return observed[span]
	}, active, "", DefaultActivationConfig().Differentiation)
	if result.RawValley != 0 || result.Differentiated {
		t.Fatalf("outcome or non-raw label leaked into density: %+v", result)
	}
}

func TestToy8ZeroOrMultipleCurrentPredicatesPreserveAmbiguity(t *testing.T) {
	for _, tt := range []struct {
		name     string
		features []Triple
	}{
		{name: "zero", features: []Triple{toy8Shared}},
		{name: "multiple", features: []Triple{toy8Shared, toy8School, toy8Skyra}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			group, settlements, observations, active, current := toy8ResolvableFixture(3)
			base := observations
			observations = func(span SpanID) []Observation {
				if span != current {
					return base(span)
				}
				out := make([]Observation, 0, len(tt.features))
				for _, feature := range tt.features {
					out = append(out, toy8Observation(current, feature, ProvenanceRaw))
				}
				return out
			}
			result := differentiateSource(group, settlements, observations, active, current,
				DefaultActivationConfig().Differentiation)
			if !result.Differentiated || result.SelectedByPocket || result.PocketMatch != -1 ||
				result.State != SourceAmbiguous || result.Committed != -1 {
				t.Fatalf("%s current evidence appointed a winner: %+v", tt.name, result)
			}
		})
	}
}

func TestToy8UnpromotedPocketCannotOverrideOrVetoGenericHeat(t *testing.T) {
	group, settlements, observations, active, current := toy8ResolvableFixture(3)
	group.State, group.Committed = SourceConfident, 1 // generic heat favored B
	selected := differentiateSource(group, settlements, observations, active, current,
		DefaultActivationConfig().Differentiation)
	if selected.SelectedByPocket || selected.PocketMatch != 0 ||
		selected.Committed != 1 || selected.State != SourceConfident {
		t.Fatalf("unpromoted school pocket changed generic B commit: %+v", selected)
	}

	base := observations
	observations = func(span SpanID) []Observation {
		if span != current {
			return base(span)
		}
		return []Observation{
			toy8Observation(current, toy8School, ProvenanceRaw),
			toy8Observation(current, toy8Skyra, ProvenanceRaw),
		}
	}
	vetoed := differentiateSource(group, settlements, observations, active, current,
		DefaultActivationConfig().Differentiation)
	if vetoed.State != SourceConfident || vetoed.Committed != 1 || vetoed.SelectedByPocket || vetoed.PocketMatch != -1 {
		t.Fatalf("unpromoted multiple pockets changed generic heat commit: %+v", vetoed)
	}
}

func TestToy8ConjunctiveOccurrenceUsesSpanIntersection(t *testing.T) {
	a := phase2Completion("S", "A", .5, 3)
	a2 := phase2Completion("S", "A2", .5, 2)
	branch := Branch{Completions: []Completion{a, a2}}
	contains, active := toy8SettlementMembership(map[string][]SpanID{
		a.EdgeKey:  {"s1", "s2", "s3"},
		a2.EdgeKey: {"s2", "s3", "s4"},
	})
	got := branchOccurrenceSpans(branch, contains, active, "")
	if len(got) != 2 || got[0] != "s2" || got[1] != "s3" {
		t.Fatalf("conjunctive occurrence spans=%v, want intersection [s2 s3]", got)
	}
}

func TestToy8PocketDiscoveryIsPermutationInvariant(t *testing.T) {
	group, settled, observations, active, current := toy8ResolvableFixture(3)
	baseline := differentiateSource(group, settled, observations, active, current,
		DefaultActivationConfig().Differentiation)

	for i, j := 0, len(active)-1; i < j; i, j = i+1, j-1 {
		active[i], active[j] = active[j], active[i]
	}
	group.Branches[0], group.Branches[1] = group.Branches[1], group.Branches[0]
	reversedObservations := func(span SpanID) []Observation {
		out := observations(span)
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
		return out
	}
	reversed := differentiateSource(group, settled, reversedObservations, active, current,
		DefaultActivationConfig().Differentiation)

	if math.Abs(reversed.RawValley-baseline.RawValley) > 1e-12 ||
		math.Abs(reversed.ValleyDepth-baseline.ValleyDepth) > 1e-12 || reversed.SelectedByPocket ||
		reversed.PocketMatch < 0 {
		t.Fatalf("permutation changed pocket geometry/selection:\nbaseline=%+v\nreversed=%+v", baseline, reversed)
	}
	selected := reversed.Branches[reversed.PocketMatch]
	if len(selected.Completions) != 1 || selected.Completions[0].Triple.Object != "A" ||
		selected.Pocket.Predicate.Triple.Key() != toy8School.Key() {
		t.Fatalf("permutation changed semantic winner: %+v", selected)
	}
}

func TestToy8CurrentAndInactiveHistoryCannotTrainTheirOwnSelection(t *testing.T) {
	group, baseSettled, observations, active, current := toy8ResolvableFixture(2)
	active = append(active, current)
	settled := func(edgeKey string, outcome Truth, span SpanID) bool {
		if span == current && outcome == TruthTrue {
			return true // adversarial: current span appears under every candidate
		}
		if span == "cold-history" {
			t.Fatal("density pass queried a span outside the active region")
		}
		return baseSettled(edgeKey, outcome, span)
	}
	result := differentiateSource(group, settled, observations, active, current,
		DefaultActivationConfig().Differentiation)
	if result.Differentiated || result.ValleyDepth >= DefaultActivationConfig().Differentiation.SplitValleyDepth {
		t.Fatalf("current span trained its own pocket selection: %+v", result)
	}
}

func TestToy8ObserveImpulseKeepsUnpromotedPocketProvisionalEndToEnd(t *testing.T) {
	b := NewBeing(nil, nil)
	b.Activation.Config.Differentiation.DominanceMargin = .9
	a := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "completion", Object: "A"}
	bValue := Triple{Mode: ModeExact, Subject: "slot:test", Predicate: "completion", Object: "B"}

	seed := func(id string, marker, completion Triple) {
		b.ObserveImpulse([]Triple{toy8Shared, marker, completion}, nil)
		span := b.CurrentSpanID()
		e := &Expression{
			EpisodeID: b.EpisodeID(), SpanID: span,
			Acts:    []Act{{ID: "a1", Primitive: "complete", Target: "slot:test"}},
			Expects: []Expect{{Of: []string{"a1"}, Triple: completion}},
		}
		b.Expr.Place(id, e)
		edge := b.Expr.Edge(id, e.Expects[0])
		edge.Observe(TruthTrue, b.tick)
		b.Expr.RecordSettlement(id, e.Expects[0], TruthTrue, span)
	}
	for i := 0; i < 3; i++ {
		seed("a"+string(rune('1'+i)), toy8School, a)
		seed("b"+string(rune('1'+i)), toy8Skyra, bValue)
	}

	result := b.ObserveImpulse([]Triple{toy8Shared, toy8School}, nil)
	if len(result.TargetGroups) != 1 {
		t.Fatalf("target groups=%+v, want one differentiated source", result.TargetGroups)
	}
	group := result.TargetGroups[0]
	if !group.Differentiated || group.SelectedByPocket || group.PocketMatch < 0 || group.State != SourceAmbiguous {
		t.Fatalf("live density pocket did not remain provisional: %+v", group)
	}
	if result.Target != nil || result.Interpretation.Route != InterpretationAmbiguous {
		t.Fatalf("unpromoted pocket reached target/handoff authority: target=%v handoff=%+v", result.Target, result.Interpretation)
	}
	present := b.present(result.Target, nil)
	for _, want := range []string{"matched provisionally and cannot commit", "pocket when RAW", "context:/ kind school"} {
		if !strings.Contains(present, want) {
			t.Fatalf("differentiation missing from express context (%q):\n%s", want, present)
		}
	}
}
