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

// Persist: fresh summary + fresh contextual semantic edges; term nodes REUSED
// across CausalNodes by normalized label (nodes reused, edges never).
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
	// Material membership and reverse summary retrieval are typed projections,
	// not fake semantic member edges.
	if got := len(c1.Cell.ChildrenIn(PortMaterial)); got != 2 {
		t.Fatalf("material children = %d, want 2", got)
	}
	if got := c1.Parents[0].ChildrenIn(PortSummaries)[c1.Cell.ID()]; got != c1.Cell {
		t.Fatal("material node must reverse-index the containing summary")
	}
	for _, e := range c1.Edges {
		if e.Kind != EdgeStructural {
			t.Fatalf("unexpected edge kind in non-composed birth graph: %s", e.Kind)
		}
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
	inherited := c2.Cell.ChildrenIn(PortInherited)
	if len(inherited) != 1 {
		t.Fatalf("inherited edges = %d, want 1", len(inherited))
	}
	for _, child := range inherited {
		edge, ok := child.(*EdgeCell)
		if !ok {
			t.Fatalf("inherited child type = %T, want *EdgeCell", child)
		}
		if edge.Record.FromID != c1.Cell.ID() || edge.Record.ToID != c2.Cell.ID() {
			t.Fatalf("inherited topology = %s -> %s", edge.Record.FromID, edge.Record.ToID)
		}
		if edge.ChildrenIn(PortFrom)[c1.Cell.ID()] != c1.Cell || edge.ChildrenIn(PortTo)[c2.Cell.ID()] != c2.Cell {
			t.Fatal("inherited edge endpoint fields and ports disagree")
		}
	}
}

func TestPersistPublishesTraversableDirectedSemanticEdges(t *testing.T) {
	ci := NewCausalNodes(nil, NewNodeRegistry(nil))
	p, _ := ParseAnnotated("michael[1] asked[1→2] skyra[2]")
	c := ci.Persist(p, 1)

	formal := c.Cell.ChildrenIn(PortFormal)
	if len(formal) != 1 {
		t.Fatalf("formal children = %d, want 1", len(formal))
	}
	for id, child := range formal {
		edge, ok := child.(*EdgeCell)
		if !ok {
			t.Fatalf("formal child type = %T, want *EdgeCell", child)
		}
		if edge.Kind() != KindSemanticEdge || edge.Record.Kind != EdgeStructural {
			t.Fatalf("edge kinds = %q/%q", edge.Kind(), edge.Record.Kind)
		}
		if edge.Record.ID != id || edge.Record.Label != "asked" {
			t.Fatalf("edge identity = %+v, map key %q", edge.Record, id)
		}
		if edge.ChildrenIn(PortFrom)[c.Parents[0].ID()] != c.Parents[0] {
			t.Fatal("from port does not expose michael")
		}
		if edge.ChildrenIn(PortTo)[c.Parents[1].ID()] != c.Parents[1] {
			t.Fatal("to port does not expose skyra")
		}
		if _, ok := edge.Strength(0); ok {
			t.Fatal("semantic edges must never carry learned strength")
		}
	}
}

func TestQueryBySubgraphPreservesDirection(t *testing.T) {
	nr := NewNodeRegistry(nil)
	ci := NewCausalNodes(nil, nr)
	forward, _ := ParseAnnotated("michael[1] asked[1→2] skyra[2]")
	reverse, _ := ParseAnnotated("skyra[1] asked[1→2] michael[2]")
	f := ci.Persist(forward, 1)
	r := ci.Persist(reverse, 2)
	michael, skyra := nr.Term("michael"), nr.Term("skyra")

	candidates := ci.Candidates(SubgraphQuery{
		NodeIDs: []string{michael.ID(), skyra.ID()},
		Edges:   []SemanticTriple{{FromID: michael.ID(), Label: "asked", ToID: skyra.ID()}},
	})
	if len(candidates) != 2 {
		t.Fatalf("candidates = %d, want both node-overlap candidates", len(candidates))
	}
	exact := map[string]bool{}
	for _, candidate := range candidates {
		exact[candidate.CausalNodeID] = candidate.Exact
		if candidate.Occurrences < 2 {
			t.Fatalf("candidate %s occurrence score = %d", candidate.CausalNodeID, candidate.Occurrences)
		}
	}
	if !exact[f.Cell.ID()] || exact[r.Cell.ID()] {
		t.Fatalf("directional verification forward=%v reverse=%v", exact[f.Cell.ID()], exact[r.Cell.ID()])
	}
}

func TestMintCausalEdgeUsesTypedPerspectivePortAndFreshActivation(t *testing.T) {
	signed := NewSignedStore()
	ci := NewCausalNodes(nil, NewNodeRegistry(nil), signed)
	priorParsed, _ := ParseAnnotated("rain[1] fell[1→2] yesterday[2]")
	presentParsed, _ := ParseAnnotated("streets[1] are[1→2] wet[2]")
	prior := ci.Persist(priorParsed, 1)
	present := ci.Persist(presentParsed, 2)

	edge, err := ci.MintCausalEdge(EdgeEfficient, "set:1", prior.Cell.ID(), present.Cell.ID())
	if err != nil {
		t.Fatal(err)
	}
	if present.Cell.ChildrenIn(PortEfficient)[edge.ID()] != edge {
		t.Fatal("present node must expose efficient relationship")
	}
	if edge.ChildrenIn(PortFrom)[prior.Cell.ID()] != prior.Cell || edge.ChildrenIn(PortTo)[present.Cell.ID()] != present.Cell {
		t.Fatal("causal edge endpoint projection disagrees with topology")
	}
	if _, ok := edge.Strength(0); ok {
		t.Fatal("fresh causal edge is untested, not zero-strength")
	}
	signed.Grade(edge.ID(), -1, 3)
	if strength, ok := edge.Strength(3); !ok || strength >= 0 {
		t.Fatalf("edge strength = %v, ok=%v; want learned tension", strength, ok)
	}

	edge2, err := ci.MintCausalEdge(EdgeEfficient, "set:2", prior.Cell.ID(), present.Cell.ID())
	if err != nil {
		t.Fatal(err)
	}
	if edge2.ID() == edge.ID() {
		t.Fatal("same nodes in a new context must mint a fresh edge")
	}
	if _, ok := edge2.Strength(3); ok {
		t.Fatal("fresh contextual edge must not inherit learned strength")
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

// The 2026-07-13 live-run failure class: verb-group particles. Brace-edges
// carry multi-word relations; particles are forgiven only ADJACENT to an
// annotated edge; a bare copula still fails (no silent relation drop).
func TestParseVerbGroupsAndParticles(t *testing.T) {
	// Brace-edge: multi-word relation label, correct direction, clean sentence.
	p, err := ParseAnnotated("michael[1] {asked about}[1→2] {skyra's well-being}[2]")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Edges) != 1 || p.Edges[0].Label != "asked about" || p.Edges[0].From != 1 || p.Edges[0].To != 2 {
		t.Fatalf("brace-edge = %+v", p.Edges)
	}
	if p.Sentence != "michael asked about skyra's well-being" {
		t.Fatalf("sentence = %q", p.Sentence)
	}

	// The exact shapes the live runs rejected, now legal via edge adjacency.
	for _, sentence := range []string{
		"michael[1] asked[1→2] about {skyra's well-being}[2]", // particle after edge
		"michael[1] is testing[1→2] {a harness}[2]",           // auxiliary before edge
		"skyra[1] will identify[1→2] {michael's location}[2]", // modal before edge
		"{the observation}[1] records[1→2] that michael[2]",   // complementizer after edge
	} {
		if _, err := ParseAnnotated(sentence); err != nil {
			t.Fatalf("%q should parse via edge-adjacent particle: %v", sentence, err)
		}
	}

	// No adjacent edge → particles stay errors: a bare copula can never
	// silently drop the relationship.
	if _, err := ParseAnnotated("streets[1] are wet[2]"); err == nil {
		t.Fatal("bare copula between two nodes must fail, not silently drop the relation")
	}
	// Particles forgiven only in the edge-adjacent gap, not sentence-wide.
	if _, err := ParseAnnotated("michael[1] greeted[1→2] skyra[2] happily with vigor"); err == nil {
		t.Fatal("non-particle unannotated token must still fail")
	}
}
