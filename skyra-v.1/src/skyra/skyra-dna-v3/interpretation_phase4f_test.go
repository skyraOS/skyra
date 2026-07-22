package dnav3

import (
	"fmt"
	"math"
	"reflect"
	"testing"
	"unicode/utf8"
)

func TestToy9FParserAcceptsCompactEvidenceSyntax(t *testing.T) {
	res, err := Parse(`<framing>use the observed clue</framing>
<act id="a1" target="file:/x" primitive="write">x</act>
<expect id="p1" of="a1">0.9 1 file:/x exists true</expect>
<evidence>[episode:1/observation:7][2:9]:-0.55->[p1.exact]</evidence>`)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Rejects) != 0 || len(res.Expression.Expects) != 1 ||
		res.Expression.Expects[0].PredictionID != "p1" {
		t.Fatalf("parsed prediction=%+v rejects=%+v", res.Expression.Expects, res.Rejects)
	}
	if len(res.Expression.PredictionAttributions) != 1 {
		t.Fatalf("attributions=%+v", res.Expression.PredictionAttributions)
	}
	a := res.Expression.PredictionAttributions[0]
	if a.ObservationID != "episode:1/observation:7" || a.StartChar != 2 || a.EndChar != 9 ||
		a.PredictionID != "p1" || a.DimensionID != "exact" || a.ClaimedInfluence != -.55 {
		t.Fatalf("parsed attribution=%+v", a)
	}
}

func TestToy9FDuplicateLocalPredictionIDsCannotFuseFields(t *testing.T) {
	b := NewBeing(nil, nil)
	e := &Expression{
		EpisodeID: b.EpisodeID(), SpanID: b.startExternalSpan(),
		Expects: []Expect{
			{PredictionID: "p1", Triple: Triple{Subject: "x", Predicate: "state", Object: "a"}},
			{PredictionID: "p1", Triple: Triple{Subject: "x", Predicate: "state", Object: "b"}},
		},
	}
	b.prepareExpressionPredictionFields("e1", e)
	b.Expr.Place("e1", e)
	if e.Expects[0].PredictionID == e.Expects[1].PredictionID {
		t.Fatalf("duplicate local IDs fused fields at %q", e.Expects[0].PredictionID)
	}
	if _, ok := b.Expr.PredictionField(e.Expects[0].PredictionID); !ok {
		t.Fatal("first prediction field missing")
	}
	if _, ok := b.Expr.PredictionField(e.Expects[1].PredictionID); !ok {
		t.Fatal("second prediction field missing")
	}
}

func TestToy9FRunNamespacesAndStoresModelAuthoredAttribution(t *testing.T) {
	var response string
	target := Triple{Subject: "file:/x", Predicate: "exists", Object: "true"}
	b := NewBeing(outcomeRealm{effect: target}, func(string, string) (string, error) { return response, nil })
	b.ObserveImpulse([]Triple{{Subject: "message:/request", Predicate: "text", Object: "prior clue"}}, nil)
	evidence := b.currentBornFrom[0]
	response = fmt.Sprintf(`<framing>use the clue</framing>
<act id="a1" target="file:/x" primitive="write">x</act>
<expect id="p1" of="a1">0.9 1 file:/x exists true</expect>
<evidence>[%s][0:10]:+0.75->[p1.exact]</evidence>`, evidence)

	if _, err := b.Run([]Triple{target}, 1); err != nil {
		t.Fatal(err)
	}
	field, ok := b.Expr.PredictionField("e1/p1")
	if !ok || len(field.Attributions) != 1 {
		t.Fatalf("stored field=%+v exists=%v", field, ok)
	}
	if field.Attributions[0].PredictionID != "e1/p1" || field.Attributions[0].ObservationID != evidence {
		t.Fatalf("stored attribution=%+v", field.Attributions[0])
	}
}

func toy9FField(b *Being, evidence Observation) PredictionField {
	return PredictionField{
		ID:            "p17",
		EpisodeID:     b.EpisodeID(),
		SpanID:        evidence.SpanID,
		OwnerKind:     "consequence",
		OwnerID:       "c17",
		AuthoredAfter: evidence.Sequence,
		Dimensions: []PredictionDimension{
			{ID: "scale.tokens", Parser: ParserTokenCount, Range: NumberRange{Low: 80, High: 180}},
			{ID: "content.traversal", Parser: ParserStringMarkers,
				Markers: []string{"traversal", "graph walk", "depth-first"}, MinMatches: 1},
			{ID: "content.exact", Parser: ParserExactAddress,
				Exact: Triple{Subject: "message:/reply", Predicate: "topic", Object: "graph walk"}},
		},
		Attributions: []EvidenceAttribution{{
			ObservationID: evidence.ID, StartChar: 0, EndChar: utf8.RuneCountInString(evidence.Object),
			PredictionID: "p17", DimensionID: "content.traversal", ClaimedInfluence: .82,
		}},
	}
}

func TestToy9FAdmitsSignedPassageAttributionAndRetainsDimensionalComparison(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	evidence := b.appendObservation(Triple{
		Subject: "message:/request", Predicate: "text", Object: "α prior traversal discussion β",
	}, ProvenanceRaw, 0, nil, span)

	field := toy9FField(b, evidence)
	stored, err := b.AdmitPredictionField(field)
	if err != nil {
		t.Fatalf("admit prediction field: %v", err)
	}
	if len(stored.Attributions) != 1 || stored.Attributions[0].ClaimedInfluence != .82 {
		t.Fatalf("stored attribution=%+v, want signed birth evidence", stored.Attributions)
	}

	settlement := b.appendObservation(Triple{
		Subject: "message:/reply", Predicate: "topic", Object: "graph walk",
	}, ProvenanceRaw, 0, nil, b.startExternalSpan())
	settlement.TokenCount = 143
	comparison, err := b.CompareAndRecordPrediction(field.ID, []Observation{settlement})
	if err != nil {
		t.Fatalf("compare prediction field: %v", err)
	}
	if len(comparison.Dimensions) != 3 {
		t.Fatalf("comparison dimensions=%d, want 3", len(comparison.Dimensions))
	}
	byID := map[DimensionID]DimensionComparison{}
	for _, dimension := range comparison.Dimensions {
		byID[dimension.Dimension.ID] = dimension
	}
	if got := byID["scale.tokens"]; got.Truth != TruthTrue || got.Observed.Number != 143 || got.Relation != "in_range" {
		t.Fatalf("token comparison=%+v", got)
	}
	if got := byID["content.traversal"]; got.Truth != TruthTrue ||
		len(got.Observed.Markers) != 1 || got.Observed.Markers[0] != "graph walk" {
		t.Fatalf("marker comparison=%+v", got)
	}
	if got := byID["content.exact"]; got.Truth != TruthTrue || got.Observed.Exact == nil ||
		got.Observed.Exact.Object != "graph walk" {
		t.Fatalf("exact comparison=%+v", got)
	}
	if comparison.SettledSpan != settlement.SpanID {
		t.Fatalf("settled span=%q, want %q", comparison.SettledSpan, settlement.SpanID)
	}

	storedComparisons := b.Expr.Comparisons(field.ID)
	if len(storedComparisons) != 1 || storedComparisons[0].Dimensions[0].State != DimensionSettled {
		t.Fatalf("durable comparisons=%+v", storedComparisons)
	}
	comparison.Dimensions[0].Observed.Markers = append(comparison.Dimensions[0].Observed.Markers, "mutated")
	again := b.Expr.Comparisons(field.ID)
	if len(again) != 1 || len(again[0].Dimensions[0].Observed.Markers) !=
		len(storedComparisons[0].Dimensions[0].Observed.Markers) {
		t.Fatal("caller mutation changed durable comparison")
	}
}

func TestToy9FAttributionAdmissionRejectsInvalidOrFutureClaims(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	prior := b.appendObservation(Triple{Subject: "message:/", Predicate: "text", Object: "prior"},
		ProvenanceRaw, 0, nil, span)
	boundary := prior.Sequence
	future := b.appendObservation(Triple{Subject: "message:/", Predicate: "text", Object: "future"},
		ProvenanceRaw, 0, nil, span)

	tests := []struct {
		name        string
		attribution EvidenceAttribution
	}{
		{name: "future", attribution: EvidenceAttribution{ObservationID: future.ID, StartChar: 0, EndChar: 1,
			PredictionID: "bad-future", DimensionID: "content", ClaimedInfluence: .5}},
		{name: "range", attribution: EvidenceAttribution{ObservationID: prior.ID, StartChar: 0, EndChar: 99,
			PredictionID: "bad-range", DimensionID: "content", ClaimedInfluence: .5}},
		{name: "influence", attribution: EvidenceAttribution{ObservationID: prior.ID, StartChar: 0, EndChar: 1,
			PredictionID: "bad-influence", DimensionID: "content", ClaimedInfluence: 1.1}},
		{name: "dimension", attribution: EvidenceAttribution{ObservationID: prior.ID, StartChar: 0, EndChar: 1,
			PredictionID: "bad-dimension", DimensionID: "missing", ClaimedInfluence: .5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field := PredictionField{
				ID: tt.attribution.PredictionID, EpisodeID: b.EpisodeID(), SpanID: span, AuthoredAfter: boundary,
				Dimensions:   []PredictionDimension{{ID: "content", Parser: ParserStringMarkers, Markers: []string{"prior"}}},
				Attributions: []EvidenceAttribution{tt.attribution},
			}
			if _, err := b.AdmitPredictionField(field); err == nil {
				t.Fatalf("invalid %s attribution admitted", tt.name)
			}
		})
	}
}

func TestToy9FComparisonIsOrderInvariantAndDoesNotGradeByItself(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	evidence := b.appendObservation(Triple{Subject: "message:/", Predicate: "text", Object: "traversal"},
		ProvenanceRaw, 0, nil, span)
	field := toy9FField(b, evidence)
	if _, err := b.AdmitPredictionField(field); err != nil {
		t.Fatal(err)
	}
	before := b.Metrics
	observation := Observation{ID: "settlement", EpisodeID: b.EpisodeID(), SpanID: "settled", Sequence: 99,
		TokenCount: 100, Provenance: ProvenanceRaw,
		Triple: Triple{Subject: "message:/reply", Predicate: "topic", Object: "graph walk"}}
	forward := ComparePredictionField(field, []Observation{observation})
	for i, j := 0, len(field.Dimensions)-1; i < j; i, j = i+1, j-1 {
		field.Dimensions[i], field.Dimensions[j] = field.Dimensions[j], field.Dimensions[i]
	}
	reversed := ComparePredictionField(field, []Observation{observation})
	if len(forward.Dimensions) != len(reversed.Dimensions) {
		t.Fatal("dimension reordering changed comparison size")
	}
	for i := range forward.Dimensions {
		if forward.Dimensions[i].Dimension.ID != reversed.Dimensions[i].Dimension.ID ||
			forward.Dimensions[i].Truth != reversed.Dimensions[i].Truth ||
			math.Abs(forward.Dimensions[i].Residual-reversed.Dimensions[i].Residual) > 1e-12 {
			t.Fatalf("dimension reordering changed comparison:\n%+v\n%+v", forward, reversed)
		}
	}
	if !reflect.DeepEqual(b.Metrics, before) {
		t.Fatal("pure comparison directly changed learned metrics")
	}
}

func TestToy9FLegacyExpectationsProjectIntoSeparateDurableFields(t *testing.T) {
	b, consequence, interpretation := toy9Setup(t)
	if consequence.Expects[0].PredictionID == "" || interpretation.Context.PredictionID == "" ||
		consequence.Expects[0].PredictionID == interpretation.Context.PredictionID {
		t.Fatalf("claim prediction IDs are missing or fused: consequence=%q interpretation=%q",
			consequence.Expects[0].PredictionID, interpretation.Context.PredictionID)
	}
	b.ObserveImpulse([]Triple{toy9Traversal, toy9School}, nil)

	consequenceComparisons := b.Expr.Comparisons(consequence.Expects[0].PredictionID)
	interpretationComparisons := b.Expr.Comparisons(interpretation.Context.PredictionID)
	if len(consequenceComparisons) != 1 || len(interpretationComparisons) != 1 {
		t.Fatalf("comparison histories consequence=%d interpretation=%d, want 1 each",
			len(consequenceComparisons), len(interpretationComparisons))
	}
	if consequenceComparisons[0].Dimensions[0].Truth != TruthTrue ||
		interpretationComparisons[0].Dimensions[0].Truth != TruthFalse {
		t.Fatalf("two-axis comparison projections=(%v,%v), want (true,false)",
			consequenceComparisons[0].Dimensions[0].Truth,
			interpretationComparisons[0].Dimensions[0].Truth)
	}
	if consequenceComparisons[0].Dimensions[0].Observed.Evidence[0] ==
		interpretationComparisons[0].Dimensions[0].Observed.Evidence[0] {
		t.Fatal("two distinct addresses recorded one settling observation")
	}
}

func TestToy9FAbsentParagraphCarriersVoidTheirContentPredictions(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	prior := b.appendObservation(Triple{
		Subject: "message:/request", Predicate: "text", Object: "please answer in three paragraphs",
	}, ProvenanceRaw, 0, nil, span)
	field := PredictionField{
		ID: "p-structural", EpisodeID: b.EpisodeID(), SpanID: span, AuthoredAfter: prior.Sequence,
		Dimensions: []PredictionDimension{
			{ID: "scale.paragraphs", Parser: ParserParagraphCount, Range: NumberRange{Low: 3, High: 3}},
			{ID: "paragraph.1.topic", Parser: ParserStringMarkers, Markers: []string{"alpha"},
				CarrierDimensionID: "scale.paragraphs", CarrierSlot: 1},
			{ID: "paragraph.2.topic", Parser: ParserStringMarkers, Markers: []string{"beta"},
				CarrierDimensionID: "scale.paragraphs", CarrierSlot: 2},
			{ID: "paragraph.3.topic", Parser: ParserStringMarkers, Markers: []string{"gamma"},
				CarrierDimensionID: "scale.paragraphs", CarrierSlot: 3},
		},
	}
	if _, err := b.AdmitPredictionField(field); err != nil {
		t.Fatal(err)
	}
	settlement := b.appendObservation(Triple{
		Subject: "message:/reply", Predicate: "text",
		Object: "alpha beta gamma all appear, but only inside the one paragraph that arrived",
	}, ProvenanceRaw, 0, nil, b.startExternalSpan())
	comparison, err := b.CompareAndRecordPrediction(field.ID, []Observation{settlement})
	if err != nil {
		t.Fatal(err)
	}

	byID := map[DimensionID]DimensionComparison{}
	for _, dimension := range comparison.Dimensions {
		byID[dimension.Dimension.ID] = dimension
	}
	if got := byID["scale.paragraphs"]; got.State != DimensionSettled || got.Truth != TruthFalse ||
		got.Observed.Number != 1 || got.Relation != "below_range" {
		t.Fatalf("paragraph scale=%+v, want settled one-versus-three miss", got)
	}
	if got := byID["paragraph.1.topic"]; got.State != DimensionSettled || got.Truth != TruthTrue {
		t.Fatalf("materialized paragraph content=%+v, want settled true", got)
	}
	for _, id := range []DimensionID{"paragraph.2.topic", "paragraph.3.topic"} {
		got := byID[id]
		if got.State != DimensionVoid || got.Truth != TruthUnknown || got.Relation != "carrier_absent" ||
			len(got.Observed.Evidence) != 0 {
			t.Fatalf("absent carrier child %q=%+v, want evidence-free void", id, got)
		}
	}
	settled := comparison.SettledDimensions()
	if len(settled) != 2 {
		t.Fatalf("learning projection contains %d dimensions, want only scale and paragraph 1", len(settled))
	}
	for _, dimension := range settled {
		if dimension.State != DimensionSettled {
			t.Fatalf("learning projection admitted non-settled dimension %+v", dimension)
		}
	}
}
