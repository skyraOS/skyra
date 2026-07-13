package dnav2

import (
	"testing"

	"skyra-v1/src/skyra"
	"skyra-v1/src/skyra/relation"
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
	root.Nodes = d.Nodes()

	kernel := relation.NewRelation()
	kernel.Realize(root, root, skyra.Impulse{
		Primitive: "engage",
		Source:    &skyra.ImpulseSource{SpaceID: "michael"},
		Payload:   "hi skyra",
	})

	if root.Condition != "external" {
		t.Fatalf("condition = %q, want external", root.Condition)
	}
	ents := d.Nodes().Snapshot()
	if len(ents) != 1 || ents[0].ID() != "michael" {
		t.Fatalf("entity index = %v, want [michael]", ents)
	}
	// the utterance landed as a KindResponse child
	var said string
	for _, c := range ents[0].Children() {
		if mc, ok := c.(*MemoryCell); ok {
			if rec, ok := mc.Content().(InteractionRecord); ok {
				said = rec.Utterance
			}
		}
	}
	if said != "hi skyra" {
		t.Fatalf("recorded utterance = %q, want %q", said, "hi skyra")
	}
}

// The intake fork: external / reafference (split by primitive) / internal. A
// perceive-return is a reafference whose record lands via the ingest, NOT an
// internal event (internal = interrupts, e.g. motor engaging the self).
func TestSensoryIntakeClassifies(t *testing.T) {
	cases := []struct {
		name string
		imp  skyra.Impulse
		want string
	}{
		{"external", skyra.Impulse{Primitive: "engage", Source: &skyra.ImpulseSource{SpaceID: "michael"}}, "external"},
		{"reafference", skyra.Impulse{Primitive: "move", Return: true, Source: &skyra.ImpulseSource{SpaceID: "world"}}, "reafference"},
		{"engage-return", skyra.Impulse{Primitive: "engage", Return: true, Source: &skyra.ImpulseSource{SpaceID: "world"}}, "reafference"},
		{"perceive-return", skyra.Impulse{Primitive: "perceive", Return: true, Source: &skyra.ImpulseSource{SpaceID: "world"}}, "perceive-return"},
		{"internal", skyra.Impulse{Primitive: "reframe", Source: &skyra.ImpulseSource{SpaceID: "skyra"}}, "internal"},
		{"self-engage", skyra.Impulse{Primitive: "engage", Source: &skyra.ImpulseSource{SpaceID: "skyra"}}, "internal"},
	}
	for _, tc := range cases {
		if got := classify(tc.imp, "skyra"); got != tc.want {
			t.Errorf("%s: classify = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// A perceive-return routes intake → ingest → midlevel:frame (decision 1b: the
// deposit continues into framing; with no exchange yet midlevel reports empty).
func TestPerceiveReturnRoutesThroughIngest(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	d.SetInfer(func(system, present string) (string, error) { return "", nil })
	o, _ := newOrch(d)

	done := make(chan []Span, 1)
	o.SetFlush(func(_ string, trace []Span) { done <- trace })
	o.Fire("P1", "", skyra.Impulse{
		Primitive: "perceive",
		Return:    true,
		Source:    &skyra.ImpulseSource{SpaceID: "world"},
	})
	trace := <-done

	if got := condOf(trace, StepSensoryIntake); got != "perceive-return" {
		t.Fatalf("intake condition = %q, want perceive-return", got)
	}
	if got := condOf(trace, StepSensoryIngest); got != "deposited" {
		t.Fatalf("ingest condition = %q, want deposited", got)
	}
	if got := condOf(trace, StepMidlevelFrame); got == "" {
		t.Fatal("perceive-return should continue into midlevel:frame")
	}
}
