package dnav2

import (
	"strings"
	"testing"

	"skyra-v1/src/skyra"
)

// firePerceiveReturn drives a translated perceive-return through the pipeline.
func firePerceiveReturn(t *testing.T, o *Orchestrator, id, prev string, sur Surroundings) []Span {
	t.Helper()
	done := make(chan []Span, 1)
	o.SetFlush(func(_ string, trace []Span) { done <- trace })
	o.Fire(id, prev, skyra.Impulse{
		Primitive: "perceive",
		Return:    true,
		Payload:   sur,
		Source:    &skyra.ImpulseSource{SpaceID: "world"},
	})
	return <-done
}

// Cold-start deposit: no active episode → the per-being default episode (one
// empty-text frame) is created; one Kind:"perceive" record per entity (self via
// the origin coord + each swept agent) lands under the frame, points at its
// entity, the entity routes back to the episode, and the born edges are crossed
// (selectable immediately).
func TestIngestDepositsSurroundingsOnTheSpine(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(func(system, present string) (string, error) { return "", nil })
	o, _ := newOrch(d)

	firePerceiveReturn(t, o, "P1", "", Surroundings{
		Self:   skyra.Coord{X: 5, Y: 3},
		Agents: map[string]skyra.Coord{"michael": {X: 2, Y: 1}},
	})

	// the default episode is active, with its empty-text frame
	g := d.Episodes().Active()
	if g == nil || g.ID() != "skyra:episode:default" {
		t.Fatalf("active episode = %v, want the default singleton", g)
	}
	frame := latestFrame(g)
	if frame == nil {
		t.Fatal("default episode has no frame")
	}
	if frameText(g) != "" {
		t.Fatalf("default frame text = %q, want empty (unframed)", frameText(g))
	}

	// one perceive record per entity under the frame, coord in the payload,
	// pointing at its entity
	coords := map[string]skyra.Coord{}
	for _, c := range frame.Children() {
		rec, ok := c.(*MemoryCell)
		if !ok || rec.Kind() != Kind("perceive") {
			continue
		}
		pr, ok := rec.Content().(PerceiveRecord)
		if !ok {
			t.Fatalf("record %s content = %T, want PerceiveRecord", rec.ID(), rec.Content())
		}
		coords[pr.EntityID] = pr.Coord
		if _, ok := rec.Children()[rec.ID()+"→"+pr.EntityID]; !ok {
			t.Fatalf("record %s should point at its entity", rec.ID())
		}
	}
	if coords["skyra"] != (skyra.Coord{X: 5, Y: 3}) {
		t.Fatalf("self record coord = %v, want {5 3} (the origin)", coords["skyra"])
	}
	if coords["michael"] != (skyra.Coord{X: 2, Y: 1}) {
		t.Fatalf("michael record coord = %v, want {2 1}", coords["michael"])
	}

	// entity → episode route, for self and swept alike
	for _, id := range []string{"skyra", "michael"} {
		entity := d.Nodes().EnsureEntity(id)
		if _, ok := entity.Children()[g.ID()]; !ok {
			t.Fatalf("entity %s should route back to the episode", id)
		}
	}

	// born edges were crossed — selectable immediately
	for _, c := range frame.Children() {
		if rec, ok := c.(*MemoryCell); ok && rec.Kind() == Kind("perceive") {
			edge := frame.ID() + "→" + rec.ID()
			if s, ok := d.Activation().Strength(edge, 0); !ok || s <= 0 {
				t.Fatalf("born edge %s not selectable: s=%.3f ok=%v", edge, s, ok)
			}
		}
	}
}

// A second sweep APPENDS fresh records (the sequence — no overwrite, no eviction;
// "current" is the freshest record per entity), reusing the singleton default
// episode and the shared entity nodes.
func TestIngestAppendsFreshRecordsPerSweep(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(func(system, present string) (string, error) { return "", nil })
	o, _ := newOrch(d)

	firePerceiveReturn(t, o, "P1", "", Surroundings{
		Self: skyra.Coord{X: 5, Y: 3}, Agents: map[string]skyra.Coord{"michael": {X: 2, Y: 1}}})
	firePerceiveReturn(t, o, "P2", "P1", Surroundings{
		Self: skyra.Coord{X: 6, Y: 3}, Agents: map[string]skyra.Coord{"michael": {X: 2, Y: 2}}})

	if n := len(d.Episodes().All()); n != 1 {
		t.Fatalf("%d episodes, want 1 (the singleton default)", n)
	}
	g := d.Episodes().Active()
	frame := latestFrame(g)

	// two sweeps × two entities = four records, all retained
	perEntity := map[string][]PerceiveRecord{}
	for _, c := range frame.Children() {
		if rec, ok := c.(*MemoryCell); ok && rec.Kind() == Kind("perceive") {
			if pr, ok := rec.Content().(PerceiveRecord); ok {
				perEntity[pr.EntityID] = append(perEntity[pr.EntityID], pr)
			}
		}
	}
	if len(perEntity["skyra"]) != 2 || len(perEntity["michael"]) != 2 {
		t.Fatalf("records per entity = skyra:%d michael:%d, want 2 each (append, not overwrite)",
			len(perEntity["skyra"]), len(perEntity["michael"]))
	}

	// the freshest record carries the new position
	freshest := func(recs []PerceiveRecord) PerceiveRecord {
		best := recs[0]
		for _, r := range recs[1:] {
			if r.Timestamp > best.Timestamp {
				best = r
			}
		}
		return best
	}
	if got := freshest(perEntity["skyra"]).Coord; got != (skyra.Coord{X: 6, Y: 3}) {
		t.Fatalf("freshest self coord = %v, want {6 3}", got)
	}
	if got := freshest(perEntity["michael"]).Coord; got != (skyra.Coord{X: 2, Y: 2}) {
		t.Fatalf("freshest michael coord = %v, want {2 2}", got)
	}
}

// The normalization pass: the frame displays ONE location line per entity — the
// freshest — pinned at the top, stale coords normalized away (display-only; the
// stale records stay on the graph, decision 4).
func TestFrameNormalizesLocationsToFreshest(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(func(system, present string) (string, error) { return "", nil })
	o, _ := newOrch(d)

	firePerceiveReturn(t, o, "P1", "", Surroundings{
		Self: skyra.Coord{X: 5, Y: 3}, Agents: map[string]skyra.Coord{"michael": {X: 2, Y: 1}}})
	firePerceiveReturn(t, o, "P2", "P1", Surroundings{
		Self: skyra.Coord{X: 6, Y: 3}, Agents: map[string]skyra.Coord{"michael": {X: 2, Y: 2}}})

	present, _ := AssembleWithStore(d.Episodes().Active(), 10000, d.Activation(), d.Clock())
	lines := strings.Split(present, "\n")
	if len(lines) < 2 || lines[0] != "location: michael at 2,2" || lines[1] != "location: skyra at 6,3" {
		t.Fatalf("frame should open with the freshest location per entity:\n%s", present)
	}
	if strings.Contains(present, "2,1") || strings.Contains(present, "5,3") {
		t.Fatalf("stale coords should be normalized out of the display:\n%s", present)
	}
}

// The §5b lifecycle: a perceive-only turn (no exchange at all) still frames — the
// surroundings render into the framing present, the op sharpens the default
// episode (edit → real frame), and the sequence retains the empty seed frame.
func TestPerceiveOnlyTurnSharpensDefaultEpisode(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		if strings.Contains(system, "evolving being") { // frameOpSystem
			if !strings.Contains(present, "location: michael at 2,1") {
				t.Errorf("framing present missing the surroundings:\n%s", present)
			}
			return "<requery>exploring[1] the space[2] with[1→3] michael[3]</requery>", nil
		}
		return "", nil
	})
	o, _ := newOrch(d)

	trace := firePerceiveReturn(t, o, "P1", "", Surroundings{
		Self: skyra.Coord{X: 5, Y: 3}, Agents: map[string]skyra.Coord{"michael": {X: 2, Y: 1}}})

	if got := condOf(trace, StepMidlevelFrame); got != "framed" {
		t.Fatalf("midlevel condition = %q, want framed (spatial framing)", got)
	}
	g := d.Episodes().Active()
	if g == nil || g.ID() != "skyra:episode:default" {
		t.Fatalf("active episode = %v, want the sharpened default", g)
	}
	if got := frameText(g); got != "exploring the space with michael" {
		t.Fatalf("default episode frame = %q, want the sharpened framing", got)
	}
}
