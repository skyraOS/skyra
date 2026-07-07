package dnav2

import (
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
	dna "skyra-v1/src/skyra/skyra-dna"
	"skyra-v1/src/skyra/world"
)

// sensory:intake resolves the sender into the entity index (on the nucleus, via
// the handed working set), records the utterance, and reports "external".
func TestSensoryIntakePortRecordsEntity(t *testing.T) {
	d := testDNA()
	WireGenes(d)

	// transcribe a sensory:intake worker with the entity index handed (as membrane
	// does), drive it through the kernel with an external engage.
	root := d.Transcribe(string(StepSensoryIntake), string(StepSensoryIntake), "",
		world.NewSlice("skyra:sensory", "skyra"), "X", nil)
	root.Entities = d.Entities()

	kernel := relation.NewRelation()
	kernel.Realize(root, root, skyra.Impulse{
		Primitive: "engage",
		Source:    &skyra.ImpulseSource{SpaceID: "michael"},
		Payload:   "hi skyra",
	})

	if root.Condition != "external" {
		t.Fatalf("condition = %q, want external", root.Condition)
	}
	ents := d.Entities().Snapshot()
	if len(ents) != 1 || ents[0].ID() != "michael" {
		t.Fatalf("entity index = %v, want [michael]", ents)
	}
	// the utterance landed as a KindResponse child
	var said string
	for _, c := range ents[0].Children() {
		if mc, ok := c.(*dna.MemoryCell); ok {
			if rec, ok := mc.Content().(dna.InteractionRecord); ok {
				said = rec.Utterance
			}
		}
	}
	if said != "hi skyra" {
		t.Fatalf("recorded utterance = %q, want %q", said, "hi skyra")
	}
}

// The three-way fork: external / reafference / internal.
func TestSensoryIntakeClassifies(t *testing.T) {
	cases := []struct {
		name string
		imp  skyra.Impulse
		want string
	}{
		{"external", skyra.Impulse{Primitive: "engage", Source: &skyra.ImpulseSource{SpaceID: "michael"}}, "external"},
		{"reafference", skyra.Impulse{Primitive: "move", Return: true, Source: &skyra.ImpulseSource{SpaceID: "world"}}, "reafference"},
		{"internal", skyra.Impulse{Primitive: "reframe", Source: &skyra.ImpulseSource{SpaceID: "skyra"}}, "internal"},
		{"self-engage", skyra.Impulse{Primitive: "engage", Source: &skyra.ImpulseSource{SpaceID: "skyra"}}, "internal"},
	}
	for _, tc := range cases {
		if got := classify(tc.imp, "skyra"); got != tc.want {
			t.Errorf("%s: classify = %q, want %q", tc.name, got, tc.want)
		}
	}
}
