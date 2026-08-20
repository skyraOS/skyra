package dnav3

import "testing"

// toyWorld bundles the paper-toy fixture with its derived segmentation.
func toyWorld(t *testing.T) ([]Observation, []Span, []Class, map[string][]int) {
	t.Helper()
	log := paperToy()
	spans := Spans(log, 0)
	classes := Classes(log, spans, 0)
	if len(classes) != 3 {
		t.Fatalf("fixture drift: want 3 classes, got %d", len(classes))
	}
	return log, spans, classes, Lens(classes)
}

func TestBuildResolutionIndexCountsPriors(t *testing.T) {
	log := []Observation{
		{Triple: Triple{Subject: "thing:my-water", Predicate: PredicateResolvedTo, Object: "thing:bottle"}},
		{Triple: Triple{Subject: "thing:my-water", Predicate: PredicateResolvedTo, Object: "thing:bottle"}},
		{Triple: Triple{Subject: "thing:my-water", Predicate: PredicateResolvedTo, Object: "thing:tap"}},
		{Triple: Triple{Subject: "x", Predicate: "unrelated", Object: "y"}},
	}
	rx := BuildResolutionIndex(log)
	cs := rx["thing:my-water"]
	if len(cs) != 2 {
		t.Fatalf("want 2 candidates, got %v", cs)
	}
	if cs[0].ID != "thing:bottle" || cs[0].Prior < 0.66 || cs[0].Prior > 0.67 {
		t.Fatalf("counted prior wrong: %+v", cs[0])
	}
	if cs[1].ID != "thing:tap" {
		t.Fatalf("candidates must sort by prior desc: %v", cs)
	}
}

// The toy impulse: "grab my drink from the fridge." Two ambiguous roots plus a verb
// root. Diversity must pick the water-routine class (3 roots beat 2), and the joint
// precedent (S1 holds bottle + kitchen + taken-from together) must attach.
func TestWaveResolvesToyImpulse(t *testing.T) {
	log, spans, classes, lens := toyWorld(t)
	rx := ResolutionIndex{
		"thing:my-drink":   {{ID: "thing:bottle", Prior: 0.5}, {ID: "thing:soda", Prior: 0.5}},
		"thing:the-fridge": {{ID: "fridge:kitchen", Prior: 0.5}, {ID: "fridge:garage", Prior: 0.5}},
	}
	impulse := []Observation{
		{Triple: Triple{Subject: "thing:my-drink", Predicate: "taken-from", Object: "thing:the-fridge"}},
	}
	roots := Roots(impulse, nil, rx)
	if len(roots) != 3 {
		t.Fatalf("want 3 roots (drink, fridge, verb), got %d: %+v", len(roots), roots)
	}

	res := Resolve(log, spans, classes, lens, roots, nil)
	if res.Primary != 0 {
		t.Fatalf("primary class = %d, want 0 (water routine: 3 distinct roots)", res.Primary)
	}
	if got := res.Ranked[0].Roots; got != 3 {
		t.Fatalf("winner heated by %d roots, want 3", got)
	}
	want := map[string]string{
		"thing:my-drink":   "thing:bottle",
		"thing:the-fridge": "fridge:kitchen",
	}
	for _, b := range res.Bindings {
		if w, ok := want[b.Root.Term]; ok && b.Candidate.ID != w {
			t.Fatalf("%s resolved to %s, want %s", b.Root.Term, b.Candidate.ID, w)
		}
	}
	if res.Novel || res.JointPrecedent != 1 {
		t.Fatalf("joint precedent = %d novel = %v, want 1/false (S1 holds all three)", res.JointPrecedent, res.Novel)
	}
}

// "Grab something from the fridge" — one contentful root, so diversity ties and the
// fridge-fetch class wins on heat. Within it, kitchen and garage tie on signature
// weight and prior: WITHOUT statistics the tie falls to the first candidate (wrong,
// and documented — this is exactly why the Born rule exists); WITH the class stats
// (kitchen 0/1, garage 2/2) garage wins.
func TestBornRuleSelectsGarage(t *testing.T) {
	log, spans, classes, lens := toyWorld(t)
	rx := ResolutionIndex{
		"thing:the-fridge": {{ID: "fridge:kitchen", Prior: 0.5}, {ID: "fridge:garage", Prior: 0.5}},
	}
	impulse := []Observation{
		{Triple: Triple{Subject: "thing:something", Predicate: "grab", Object: "thing:the-fridge"}},
	}
	roots := Roots(impulse, nil, rx)

	blind := Resolve(log, spans, classes, lens, roots, nil)
	if blind.Primary != 1 {
		t.Fatalf("primary = %d, want 1 (fridge-fetch on heat)", blind.Primary)
	}
	for _, b := range blind.Bindings {
		if b.Root.Term == "thing:the-fridge" && b.Candidate.ID != "fridge:kitchen" {
			t.Fatalf("without stats the tie documents itself as first-candidate: got %s", b.Candidate.ID)
		}
	}

	stats := func(class int, term string) (float64, int) {
		if class != 1 {
			return 0, 0
		}
		switch term {
		case "fridge:garage":
			return 1.0, 2
		case "fridge:kitchen":
			return 0.0, 1
		}
		return 0, 0
	}
	res := Resolve(log, spans, classes, lens, roots, stats)
	for _, b := range res.Bindings {
		if b.Root.Term == "thing:the-fridge" && b.Candidate.ID != "fridge:garage" {
			t.Fatalf("with class stats garage must win, got %s", b.Candidate.ID)
		}
	}
}

// Paper test #3's contradiction: bottle grounds in the water routine, garage in the
// fridge-fetch class — a stitch across classes with zero witnessed co-occurrence.
// Legal, but the novelty flag must fly.
func TestNovelStitchFlagged(t *testing.T) {
	log, spans, classes, lens := toyWorld(t)
	rx := ResolutionIndex{
		"thing:my-water":   {{ID: "thing:bottle", Prior: 1}},
		"thing:the-fridge": {{ID: "fridge:garage", Prior: 1}},
	}
	impulse := []Observation{
		{Triple: Triple{Subject: "thing:my-water", Predicate: "put-into", Object: "thing:the-fridge"}},
	}
	res := Resolve(log, spans, classes, lens, Roots(impulse, nil, rx), nil)

	byTerm := map[string]Binding{}
	for _, b := range res.Bindings {
		byTerm[b.Root.Term] = b
	}
	if byTerm["thing:my-water"].Class == byTerm["thing:the-fridge"].Class {
		t.Fatal("fixture should force a stitch across two classes")
	}
	if !res.Novel || res.JointPrecedent != 0 {
		t.Fatalf("bottle+garage co-occur nowhere: want novel/0, got %v/%d", res.Novel, res.JointPrecedent)
	}
}

func TestCommitWritesInterpretedResolutions(t *testing.T) {
	log, spans, classes, lens := toyWorld(t)
	rx := ResolutionIndex{
		"thing:my-drink": {{ID: "thing:bottle", Prior: 0.75}, {ID: "thing:soda", Prior: 0.25}},
	}
	impulse := []Observation{
		{Triple: Triple{Subject: "thing:my-drink", Predicate: "taken-from", Object: "fridge:kitchen"}},
	}
	res := Resolve(log, spans, classes, lens, Roots(impulse, nil, rx), nil)
	commits := res.Commit([]string{"raw:1"})
	if len(commits) != 1 {
		t.Fatalf("want exactly one non-identity commit, got %d: %+v", len(commits), commits)
	}
	c := commits[0]
	if c.Subject != "thing:my-drink" || c.Predicate != PredicateResolvedTo || c.Object != "thing:bottle" {
		t.Fatalf("commit triple wrong: %+v", c.Triple)
	}
	if c.Provenance != ProvenanceInterpreted || c.Confidence != 0.75 {
		t.Fatalf("a resolution is a READING: interpreted provenance, counted confidence; got %+v", c)
	}
	if len(c.DerivedFrom) != 1 || c.DerivedFrom[0] != "raw:1" {
		t.Fatal("commit must point at its raw evidence — that pointer is the edge the feedback loop grades")
	}
	// And the loop closes: the commit feeds the next impulse's tier-1 priors.
	rx2 := BuildResolutionIndex(append(log, commits...))
	if got := rx2["thing:my-drink"]; len(got) != 1 || got[0].ID != "thing:bottle" {
		t.Fatalf("commit should found the resolution index entry, got %v", got)
	}
}

func TestClassEdgesJoinThroughEvidence(t *testing.T) {
	log, spans, classes, _ := toyWorld(t)
	g := NewExpressionGraph()
	e := &Expression{
		Acts: []Act{{ID: "a1", Target: "fridge:kitchen", Primitive: "fetch"}},
		Expects: []Expect{{
			Of:     []string{"a1"},
			Triple: Triple{Mode: ModeExact, Subject: "thing:bottle", Predicate: "taken-from", Object: "fridge:kitchen"},
			Within: 1,
		}},
	}
	g.Place("e1", e)
	// Resolve the bet by the evidence at log[1] — S1, class 0.
	e.Expects[0].Resolution = TruthTrue
	e.Expects[0].ResolvedBy = []string{obsID(log[1])}
	g.Edge("e1", e.Expects[0]).Observe(TruthTrue, 1)

	byClass := g.ClassEdges(log, spans, classes)
	if len(byClass[0]) != 1 {
		t.Fatalf("edge should group under class 0 via its evidence span, got %+v", byClass)
	}
	if byClass[0][0].N != 1 {
		t.Fatalf("the grouped edge is the LEARNED one (N=1), got N=%d", byClass[0][0].N)
	}
	if len(byClass[1]) != 0 && len(byClass[2]) != 0 {
		t.Fatal("evidence in one span must not leak into other classes")
	}
}
