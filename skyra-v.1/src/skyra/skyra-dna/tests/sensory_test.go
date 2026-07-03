package tests

import (
	"testing"
	"time"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// Perception is SENSORY's, not the membrane's. These exercise sensory directly
// (synchronously) — the boundary hop that gets a wave here is in membrane_test.go
// / cells_test.go.

// An external wave names its sender (Source.SpaceID) and where it came from
// (Source.Coord); sensory records who/where/what into its entity-keyed index.
func TestSensoryWritesEntityIndex(t *testing.T) {
	m, _ := newBeing("skyra")
	sensory := m.Sensory()
	sensory.Slice().SetCount(7)

	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", PointID: "3,4", Coord: skyra.Coord{X: 3, Y: 4}},
		Primitive: "engage",
		Payload: "hello skyra",
	})

	got, ok := sensory.ByEntity["michael"]
	if !ok {
		t.Fatalf("michael not in entity index: %+v", sensory.ByEntity)
	}
	if got.Content() != (dna.Entity{ID: "michael"}) {
		t.Errorf("entity lookup content = %+v, want Entity{ID:michael}", got.Content())
	}

	entity, ok := got.(*dna.MemoryCell)
	if !ok {
		t.Fatalf("michael = %T, want *MemoryCell", got)
	}
	if entity.Kind() != dna.KindEntity {
		t.Fatalf("michael kind = %q, want entity", entity.Kind())
	}
	if len(entity.Children()) != 2 {
		t.Fatalf("entity record children = %d, want 2", len(entity.Children()))
	}
	seenState, seenInteraction := false, false
	for _, child := range entity.Children() {
		switch record := child.Content().(type) {
		case dna.EntityState:
			seenState = true
			if record.Location != (skyra.Coord{X: 3, Y: 4}) || record.Timestamp != 7 {
				t.Fatalf("state = %+v, want location 3,4 timestamp 7", record)
			}
		case dna.InteractionRecord:
			seenInteraction = true
			if record.From != "michael" || record.To != "skyra" || record.Utterance != "hello skyra" {
				t.Fatalf("record = %+v, want michael -> skyra utterance", record)
			}
			if record.Timestamp != 7 {
				t.Fatalf("record timestamp = %d, want 7", record.Timestamp)
			}
		default:
			t.Fatalf("entity record content = %T, want EntityState or InteractionRecord", child.Content())
		}
	}
	if !seenState || !seenInteraction {
		t.Fatalf("children missing state=%v interaction=%v", seenState, seenInteraction)
	}
}

// A fresh entity is warmed (its records pre-created) so it is scored within the
// SAME sensory wave that first perceives it.
func TestSensoryWarmsNewEntityInSameTraversal(t *testing.T) {
	m, kernel := newBeing("skyra")
	sensory := m.Sensory()

	// traversed = sensory directly → one synchronous wave on sensory's slice.
	kernel.Realize(nil, sensory, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 3, Y: 4}},
		Primitive: "engage",
		Payload: "hello skyra",
	})

	seenEntity, seenState, seenInteraction := false, false, false
	var walk func(*skyra.Node)
	walk = func(n *skyra.Node) {
		if n == nil {
			return
		}
		if n.Cell != nil && n.Cell.ID() == "michael" {
			seenEntity = true
		}
		switch n.Cell.Content().(type) {
		case dna.EntityState:
			seenState = true
		case dna.InteractionRecord:
			seenInteraction = true
		}
		for _, child := range n.Children {
			walk(child)
		}
	}
	tree := sensory.Slice().Tree()
	if len(tree.Current) != 1 {
		t.Fatalf("tree roots = %d, want 1", len(tree.Current))
	}
	walk(tree.Current[0])
	if !seenEntity || !seenState || !seenInteraction {
		t.Fatalf("same traversal saw entity=%v state=%v interaction=%v", seenEntity, seenState, seenInteraction)
	}
	entity := sensory.ByEntity["michael"].(*dna.MemoryCell)
	if len(entity.Children()) != 2 {
		t.Fatalf("same traversal entity children = %d, want 2", len(entity.Children()))
	}
}

// Presence decays via the perceive wave: the perceived sender bumps, everyone
// absent decays. The fast EMA (presence) loses a larger fraction than the slow
// EMA (groove) — recency fades, familiarity holds.
func TestPresenceDecaysFasterThanGroove(t *testing.T) {
	m, kernel := newBeing("skyra")
	sensory := m.Sensory()

	perceive := func(sender string) {
		kernel.Realize(nil, sensory, skyra.Impulse{
			Source:  &skyra.ImpulseSource{SpaceID: sender, Coord: skyra.Coord{X: 1, Y: 1}},
			Payload: "hi",
		})
	}

	perceive("michael")
	presenceHot := sensory.Presence("michael")
	grooveHot := sensory.Groove("michael")
	if presenceHot <= 0 || grooveHot <= 0 {
		t.Fatalf("michael not warmed: presence=%v groove=%v", presenceHot, grooveHot)
	}

	// Attention moves to sarah; michael is now absent on every wave.
	perceive("sarah")
	perceive("sarah")

	presenceCold := sensory.Presence("michael")
	grooveCold := sensory.Groove("michael")

	if presenceCold >= presenceHot {
		t.Fatalf("michael presence did not decay: %v -> %v", presenceHot, presenceCold)
	}
	if grooveCold <= 0 {
		t.Fatalf("michael groove erased (%v); the durable EMA should survive absence", grooveCold)
	}
	if sensory.Presence("sarah") <= presenceCold {
		t.Fatalf("sarah (present) %v should outrank absent michael %v", sensory.Presence("sarah"), presenceCold)
	}
	if grooveCold/grooveHot <= presenceCold/presenceHot {
		t.Fatalf("groove should decay slower than presence: grooveRetained=%v presenceRetained=%v",
			grooveCold/grooveHot, presenceCold/presenceHot)
	}
}

// Return-perception waves are not sensory intake — they must not pollute the
// surroundings.
func TestSensoryObserveSkipsReturnWave(t *testing.T) {
	m, _ := newBeing("skyra")
	sensory := m.Sensory()

	sensory.Observe(nil, skyra.Impulse{
		Return:  true,
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Payload: "inner",
	})

	if len(sensory.ByEntity) != 0 {
		t.Fatalf("return wave wrote entity index: %+v", sensory.ByEntity)
	}
}

// A later observation of the same entity reuses the lookup and appends fresh
// state/interaction child nodes.
func TestSensoryObserveUpdatesEntityTimestamp(t *testing.T) {
	m, _ := newBeing("skyra")
	sensory := m.Sensory()

	sensory.Slice().SetCount(1)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Primitive: "engage",
		Payload: "first",
	})

	sensory.Slice().SetCount(2)
	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 2, Y: 3}},
		Primitive: "engage",
		Payload: "second",
	})

	entity, ok := sensory.ByEntity["michael"].(*dna.MemoryCell)
	if !ok {
		t.Fatalf("michael = %T, want *MemoryCell", sensory.ByEntity["michael"])
	}
	if len(entity.Children()) != 4 {
		t.Fatalf("entity record children = %d, want 4", len(entity.Children()))
	}
	seenSecondState, seenSecond := false, false
	for _, child := range entity.Children() {
		switch record := child.Content().(type) {
		case dna.EntityState:
			if record.Location == (skyra.Coord{X: 2, Y: 3}) && record.Timestamp == 2 {
				seenSecondState = true
			}
		case dna.InteractionRecord:
			if record.Utterance == "second" {
				seenSecond = true
				if record.Timestamp != 2 {
					t.Fatalf("second record = %+v, want timestamp 2", record)
				}
			}
		}
	}
	if !seenSecondState || !seenSecond {
		t.Fatalf("second observation not appended: state=%v interaction=%v", seenSecondState, seenSecond)
	}
}

// The next pipeline link: on Express, sensory fires a forward wave into cognition
// (Global), owned by Global (origin = traversed = Global) — sensory only launches.
func TestSensoryFiresMidLevelOnExpress(t *testing.T) {
	spy := newSpyKernel()
	m := newBeingWithKernel("skyra", spy)
	sensory := m.Sensory()

	imp := skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 1}},
		Primitive: "engage",
		Payload: "hello skyra",
	}
	sensory.Express(nil, imp)

	select {
	case call := <-spy.calls:
		if call.origin == nil || call.origin.ID() != "skyra:midlevel" {
			t.Fatalf("spawn origin = %v, want skyra:midlevel (cognition owns its wave)", call.origin)
		}
		if call.traversed == nil || call.traversed.ID() != "skyra:midlevel" {
			t.Fatalf("spawn traversed = %v, want skyra:midlevel", call.traversed)
		}
		got, ok := call.impulse.(skyra.Impulse)
		if !ok || got.Payload != "hello skyra" {
			t.Fatalf("spawn impulse = %+v, want the forwarded impulse", call.impulse)
		}
	case <-time.After(time.Second):
		t.Fatal("sensory did not spawn a cognition (MidLevel) wave on Express")
	}
}

// A self-originated wave isn't a surrounding entity.
func TestSensoryObserveSkipsSelf(t *testing.T) {
	m, _ := newBeing("skyra")
	sensory := m.Sensory()

	sensory.Observe(nil, skyra.Impulse{
		Source:  &skyra.ImpulseSource{SpaceID: "skyra", Coord: skyra.Coord{X: 0, Y: 0}},
		Payload: "self",
	})

	if len(sensory.ByEntity) != 0 {
		t.Fatalf("self wave wrote entity index: %+v", sensory.ByEntity)
	}
}
