package dnav3

import "testing"

func TestPhase3AuditSourceIdentityIncludesPayloadAndRequirements(t *testing.T) {
	base := Act{ID: "a1", Primitive: "write", Target: "file:/x", Payload: "one"}
	differentPayload := base
	differentPayload.Payload = "two"
	differentRequirement := base
	differentRequirement.Requires = []Triple{{Subject: "file:/guard", Predicate: "exists", Object: "true"}}

	key := func(act Act) string {
		return edgeSourceKey(map[string]Act{"a1": act}, []string{"a1"})
	}
	if key(base) == key(differentPayload) {
		t.Fatal("different payloads aliased one action source")
	}
	if key(base) == key(differentRequirement) {
		t.Fatal("different requirements aliased one action source")
	}

	left := differentRequirement
	left.Requires = append(left.Requires,
		Triple{Subject: "build:/", Predicate: "status", Object: "ok"})
	right := left
	right.Requires[0], right.Requires[1] = right.Requires[1], right.Requires[0]
	if key(left) != key(right) {
		t.Fatal("requirement insertion order fragmented equivalent action identity")
	}
}

func TestPhase3AuditDensityUsesPredictionTimeSpanNotResolvingEffects(t *testing.T) {
	b := NewBeing(nil, nil)
	completionA := Triple{Subject: "slot:test", Predicate: "completion", Object: "A"}
	completionB := Triple{Subject: "slot:test", Predicate: "completion", Object: "B"}
	leakA := Triple{Subject: "outcome:/", Predicate: "error", Object: "after-A"}
	leakB := Triple{Subject: "outcome:/", Predicate: "error", Object: "after-B"}
	active := []SpanID{}
	var aEdge, bEdge string

	seed := func(id string, context, completion, effect Triple) string {
		source := b.startExternalSpan()
		b.appendObservation(toy8Shared, ProvenanceRaw, 0, nil, source)
		b.appendObservation(context, ProvenanceRaw, 0, nil, source)
		active = append(active, source)
		e := &Expression{
			EpisodeID: b.EpisodeID(), SpanID: source,
			Acts:    []Act{{ID: "a1", Primitive: "complete", Target: "slot:test"}},
			Expects: []Expect{{Of: []string{"a1"}, Triple: completion}},
		}
		b.Expr.Place(id, e)
		edge := b.Expr.Edge(id, e.Expects[0])
		edge.Observe(TruthTrue, b.tick)

		resolving := b.startExternalSpan()
		b.appendObservation(completion, ProvenanceRaw, 0, nil, resolving)
		b.appendObservation(effect, ProvenanceRaw, 0, nil, resolving)
		b.Expr.RecordSettlement(id, e.Expects[0], TruthTrue, resolving)
		return edgeKey(actMap(e), e.Expects[0].Of, e.Expects[0].Key())
	}
	for i := 0; i < 3; i++ {
		aEdge = seed("audit-a"+string(rune('1'+i)), toy8School, completionA, leakA)
		bEdge = seed("audit-b"+string(rune('1'+i)), toy8Skyra, completionB, leakB)
	}

	group := TargetGroup{Source: "audit", State: SourceAmbiguous, Committed: -1, PocketMatch: -1,
		Branches: []Branch{
			{Completions: []Completion{{Triple: completionA, EdgeKey: aEdge, Trials: 3}}},
			{Completions: []Completion{{Triple: completionB, EdgeKey: bEdge, Trials: 3}}},
		}}
	result := differentiateSource(group, b.Expr.SettlementSourceInSpan, b.Spans.Observations,
		active, "", DefaultActivationConfig().Differentiation)
	if !result.Differentiated {
		t.Fatalf("prediction-time contexts did not form pockets: %+v", result)
	}
	gotA := result.Branches[0].Pocket.Predicate.Triple
	gotB := result.Branches[1].Pocket.Predicate.Triple
	if gotA.Key() != toy8School.Key() || gotB.Key() != toy8Skyra.Key() {
		t.Fatalf("resolving effects leaked into pocket names: got A=%+v B=%+v", gotA, gotB)
	}
	if b.Expr.SettlementSourceInSpan(aEdge, TruthTrue, active[0]) != true {
		t.Fatal("source-span settlement provenance was not retained")
	}
}
