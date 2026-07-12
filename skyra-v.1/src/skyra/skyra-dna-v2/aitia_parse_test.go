package dnav2

import (
	"math"
	"strings"
	"testing"
)

// The canonical example round-trips: three nodes, two edges, clean sentence.
func TestParseAnnotatedCanonical(t *testing.T) {
	p, err := ParseAnnotated("michael[1] asked[1→2] skyra[2] to check[2→3] the logs[3]")
	if err != nil {
		t.Fatal(err)
	}
	if p.Sentence != "michael asked skyra to check the logs" {
		t.Fatalf("clean sentence = %q", p.Sentence)
	}
	if len(p.Nodes) != 3 || len(p.Edges) != 2 {
		t.Fatalf("V=%d E=%d, want 3/2", len(p.Nodes), len(p.Edges))
	}
	if p.Nodes[0].Label != "michael" || p.Nodes[1].Label != "skyra" || p.Nodes[2].Label != "logs" {
		t.Fatalf("labels = %v", p.Nodes)
	}
	if p.Edges[0] != (ParsedEdge{Label: "asked", From: 1, To: 2}) {
		t.Fatalf("edge 0 = %+v", p.Edges[0])
	}
	if p.Edges[1] != (ParsedEdge{Label: "check", From: 2, To: 3}) {
		t.Fatalf("edge 1 = %+v", p.Edges[1])
	}
}

// Braces bind multi-word referents; ASCII arrow accepted; labels normalize.
func TestParseAnnotatedPhrasesAndArrow(t *testing.T) {
	p, err := ParseAnnotated("{The Skyra Project}[1] needs[1->2] help[2].")
	if err != nil {
		t.Fatal(err)
	}
	if p.Nodes[0].Label != "the skyra project" {
		t.Fatalf("phrase label = %q", p.Nodes[0].Label)
	}
	if p.Nodes[1].Label != "help" { // trailing period stripped by normalize... on the word before [2]? "help[2]." — annotation ends before '.', label is "help"
		t.Fatalf("label = %q", p.Nodes[1].Label)
	}
	if p.Edges[0].Label != "needs" || p.Edges[0].From != 1 || p.Edges[0].To != 2 {
		t.Fatalf("edge = %+v", p.Edges[0])
	}
	if p.Sentence != "The Skyra Project needs help." {
		t.Fatalf("sentence = %q", p.Sentence)
	}
}

// Coreference: a repeated ordinal re-references the node; the first occurrence
// names it; a stable id from any occurrence binds it.
func TestParseAnnotatedCoreference(t *testing.T) {
	p, err := ParseAnnotated("michael[1], he[1=entity:michael] wants[1→2] coffee[2]")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Nodes) != 2 {
		t.Fatalf("V=%d, want 2 (coreference collapses)", len(p.Nodes))
	}
	if p.Nodes[0].Label != "michael" || p.Nodes[0].StableID != "entity:michael" {
		t.Fatalf("node 1 = %+v", p.Nodes[0])
	}
}

// Semantic content outside annotations is not glue: sentence and subgraph must
// commit as one complete claim.
func TestParseAnnotatedRejectsUnrepresentedSemantics(t *testing.T) {
	_, err := ParseAnnotated("michael[1] asked[1→2] skyra[2] to investigate the outage")
	if err == nil {
		t.Fatal("unannotated semantic content must fail")
	}
	msg := err.Error()
	if !strings.Contains(msg, `"investigate"`) || !strings.Contains(msg, `"outage"`) {
		t.Fatalf("missing omitted-token errors: %s", msg)
	}

	p, err := ParseAnnotated("michael[1] asked[1→2] skyra[2]. skyra[2] investigates[2→3] the outage[3]")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Nodes) != 3 || len(p.Edges) != 2 {
		t.Fatalf("V=%d E=%d, want 3/2", len(p.Nodes), len(p.Edges))
	}
}

// Every problem is reported at once — the correction prompt needs them all.
func TestParseAnnotatedCollectsAllProblems(t *testing.T) {
	_, err := ParseAnnotated("michael[1=entity:a] loves[1→9] him[1=entity:b]")
	if err == nil {
		t.Fatal("expected parse errors")
	}
	pe, ok := err.(*ParseErrors)
	if !ok {
		t.Fatalf("error type %T", err)
	}
	msg := pe.Error()
	if !strings.Contains(msg, "two stable ids") || !strings.Contains(msg, "undeclared node 9") {
		t.Fatalf("problems missing from: %s", msg)
	}
}

// An emission with no nodes is an empty claim.
func TestParseAnnotatedEmptyClaim(t *testing.T) {
	if _, err := ParseAnnotated("just some plain words"); err == nil {
		t.Fatal("plain text should not parse as a compound")
	}
}

// Persist: fresh summary + fresh contextual edges; term nodes REUSED across
// compounds by normalized label (nodes reused, edges never).
func TestPersistReusesNodesFreshEdges(t *testing.T) {
	nr := NewNodeRegistry(nil)
	ci := NewCausalNodes(nil, nr)
	p1, _ := ParseAnnotated("michael[1] asked[1→2] skyra[2]")
	p2, _ := ParseAnnotated("skyra[1] answered[1→2] michael[2]")

	c1 := ci.Persist(p1, 1)
	c2 := ci.Persist(p2, 2)

	if c1.Cell.ID() == c2.Cell.ID() {
		t.Fatal("summaries must be fresh")
	}
	// same term nodes, both compounds
	if nr.Term("michael") == nil || nr.Term("skyra") == nil {
		t.Fatal("term registry missing shared nodes")
	}
	if c1.Parents[0] != c2.Parents[1] { // michael
		t.Fatal("michael must be ONE reused node")
	}
	// edges are contextual: no shared edge IDs
	seen := map[string]bool{}
	for _, e := range append(c1.Edges, c2.Edges...) {
		if seen[e.ID] {
			t.Fatalf("edge id %s reused across contexts", e.ID)
		}
		seen[e.ID] = true
	}
	// member edges point AT the summary (the subgraph points to it)
	memberEdges := 0
	for _, e := range c1.Edges {
		if e.Kind == EdgeMember {
			memberEdges++
			if e.ToID != c1.Cell.ID() {
				t.Fatalf("member edge points at %s, want the summary", e.ToID)
			}
		}
	}
	if memberEdges != 2 {
		t.Fatalf("member edges = %d, want 2", memberEdges)
	}
}

// Resolution is unified with the observed graph: a mentioned referent whose
// label matches an OBSERVED entity resolves to that entity node — one michael.
// A mentioned-but-never-observed referent stays a term (entities are born from
// observation, never from assertion).
func TestResolveUnifiesWithObservedEntities(t *testing.T) {
	nr := NewNodeRegistry(nil)
	observed := nr.EnsureEntity("michael") // born from observation (sensory/ingest)
	ci := NewCausalNodes(nil, nr)

	p, _ := ParseAnnotated("michael[1] mentioned[1→2] sarah[2]")
	c := ci.Persist(p, 1)

	if c.Parents[0] != observed {
		t.Fatal("mentioned michael must resolve to the OBSERVED entity node")
	}
	if nr.Term("michael") != nil {
		t.Fatal("no shadow term when an observed entity matches")
	}
	if c.Parents[1].Kind() != KindTerm {
		t.Fatal("never-observed sarah stays a term (assertion, not observation)")
	}

	// the explicit escape hatch reaches the entity too, prefix accepted
	p2, _ := ParseAnnotated("he[1=entity:michael] left[1→2] the maze[2]")
	c2 := ci.Persist(p2, 2)
	if c2.Parents[0] != observed {
		t.Fatal("[k=entity:id] must resolve to the observed entity node")
	}
}

// Composition: a stable compound id as referent resolves to the existing
// summary node — previous compounds are legal material.
func TestPersistComposesPreviousCompound(t *testing.T) {
	ci := NewCausalNodes(nil, NewNodeRegistry(nil))
	p1, _ := ParseAnnotated("{the outage}[1] hit[1→2] the api[2]")
	c1 := ci.Persist(p1, 1)

	p2, _ := ParseAnnotated("{the outage}[1=" + c1.Cell.ID() + "] caused[1→2] delays[2]")
	c2 := ci.Persist(p2, 2)

	if c2.Parents[0] != c1.Cell {
		t.Fatal("the compound summary node should be the reused referent")
	}
}

// Signed store: EMA math, lazy Familiarity decay, no-grade ≠ zero, clamping.
func TestSignedStoreMechanics(t *testing.T) {
	s := NewSignedStore()

	if _, ok := s.Strength("e1", 0); ok {
		t.Fatal("ungraded edge must have NO strength (no grade ≠ zero)")
	}

	s.Grade("e1", 1, 0)
	got, ok := s.Strength("e1", 0)
	want := (0.02 + 0.10) / 2
	if !ok || math.Abs(got-want) > 1e-12 {
		t.Fatalf("strength = %v, want %v", got, want)
	}

	// lazy decay: Familiarity fades over silence, Identity holds
	later, _ := s.Strength("e1", 50)
	wantLater := (0.02 + 0.10*math.Pow(signedDecay, 50)) / 2
	if math.Abs(later-wantLater) > 1e-12 {
		t.Fatalf("decayed strength = %v, want %v", later, wantLater)
	}
	if later >= got {
		t.Fatal("silence must fade the recent charge")
	}

	// signed: negative grades drive strength negative; clamp holds
	s.Grade("e2", -5, 0) // clamps to -1
	neg, _ := s.Strength("e2", 0)
	if neg >= 0 {
		t.Fatalf("negative grade should give negative strength, got %v", neg)
	}
	if Polarity(neg) != -1 || Magnitude(neg) != -neg {
		t.Fatal("polarity/magnitude split broken")
	}
	if s.GradeCount("e2") != 1 || s.GradeCount("nope") != 0 {
		t.Fatal("grade counts")
	}
}
