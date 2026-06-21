package tests

import (
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

// A heterogeneous being — MembraneCell → GlobalCell → MemoryCell(setpoint) →
// MemoryCell(entity) — traverses as one wave. Proves the cells implement Space,
// compose via Children, and the kernel walks the whole being top to bottom.
// (Activation is stubbed; this is structural traversal, not cognition yet.)
//
// Black-box: this is an integration test in its own package, so it exercises
// only the exported API across dna + relation + world.
func TestBeingTraverses(t *testing.T) {
	sl := world.NewSlice("being", "skyra") // threshold 0 → everything survives

	entity := dna.NewMemoryCell("michael", dna.KindEntity, "person:michael", sl)
	setpoint := dna.NewMemoryCell("greeted", dna.KindSetpoint, "the user has been greeted", sl)
	setpoint.AddChild("greeted→michael", entity) // closure snapshot edge

	global := dna.NewGlobalCell("global", sl)
	global.AddGrouping("global→greeted", setpoint)

	membrane := dna.NewMembraneCell("skyra", "be helpful to the user", sl)
	membrane.AddChild(global)

	kernel := relation.NewRelation()
	// Return:true keeps it to one wave (no second return-perception pass).
	kernel.Realize(membrane, membrane, skyra.Impulse{Primitive: "perceive", Return: true})

	// The finished tree should descend membrane → global → setpoint → entity.
	tree := sl.Tree()
	if len(tree.Current) != 1 || tree.Current[0].CellID != "skyra" {
		t.Fatalf("tree root = %+v, want membrane 'skyra'", tree.Current)
	}
	got := []string{tree.Current[0].CellID}
	node := tree.Current[0]
	for len(node.Children) == 1 {
		node = node.Children[0]
		got = append(got, node.CellID)
	}
	want := []string{"skyra", "global", "global→greeted", "greeted→michael"}
	if len(got) != len(want) {
		t.Fatalf("traversal chain = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("traversal chain = %v, want %v", got, want)
		}
	}
}
