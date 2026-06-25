package tests

import (
	"testing"
	"time"

	"skyra-v1/src/skyra"
	dna "skyra-v1/src/skyra/skyra-dna"
)

// End-to-end boundary hop (real kernel): an external wave hits the thin membrane,
// which trampolines — Observe nothing, no children, Express spawns a wave into
// sensory — and the perception lands in sensory's own index/slice, asynchronously.
//
// Black-box: integration across dna + relation + world via the exported API.
func TestBoundaryHopLandsSurroundings(t *testing.T) {
	m, kernel := newBeing("skyra")

	kernel.Realize(nil, m, skyra.Impulse{
		Primitive: "perceive",
		Source:    &skyra.ImpulseSource{SpaceID: "michael", Coord: skyra.Coord{X: 1, Y: 2}},
		Payload:   "hello skyra",
	})

	// The membrane's own traversal is a thin leaf: root = membrane, no children.
	tree := m.Slice().Tree()
	if len(tree.Current) != 1 || tree.Current[0].CellID != "skyra" {
		t.Fatalf("membrane tree root = %+v, want thin membrane 'skyra'", tree.Current)
	}
	if len(tree.Current[0].Children) != 0 {
		t.Fatalf("membrane tree has children = %+v, want none (thin)", tree.Current[0].Children)
	}

	// The spawned sensory wave is async — poll its index (lock-guarded) for michael.
	deadline := time.Now().Add(2 * time.Second)
	var entity skyra.Space
	for time.Now().Before(deadline) {
		if e, ok := sensoryEntity(m.Sensory(), "michael"); ok {
			entity = e
			break
		}
		time.Sleep(2 * time.Millisecond)
	}
	if entity == nil {
		t.Fatal("boundary hop did not land michael in sensory's surroundings")
	}
	if cell, ok := entity.(*dna.MemoryCell); !ok || cell.Kind() != dna.KindEntity {
		t.Fatalf("michael = %T kind, want *MemoryCell entity", entity)
	}
}
