package dnav2

import (
	"strings"
	"sync/atomic"
	"testing"
)

func TestParseStructuralOpContract(t *testing.T) {
	cases := []struct {
		raw, op, body string
	}{
		{"<retain/>", StructuralRetain, ""},
		{"<compose>michael[1] extends[1→2] plan[2]</compose>", StructuralCompose, "michael[1] extends[1→2] plan[2]"},
		{"<requery>michael[1] changes[1→2] plan[2]</requery>", StructuralRequery, "michael[1] changes[1→2] plan[2]"},
	}
	for _, tc := range cases {
		op, body, ok := ParseStructuralOp(tc.raw)
		if !ok || op != tc.op || body != tc.body {
			t.Fatalf("%q => op=%q body=%q ok=%v", tc.raw, op, body, ok)
		}
	}
	for _, invalid := range []string{"<retain>text</retain>", "<new>x</new>", "<retain/><retain/>", "plain"} {
		if _, _, ok := ParseStructuralOp(invalid); ok {
			t.Fatalf("invalid structural operation accepted: %q", invalid)
		}
	}
}

func seedStructuralRuntime(t *testing.T) (*DNA, *EphemeralRoot, AitiaSet, *CausalNode) {
	t.Helper()
	d := testDNA()
	WireGenes(d)
	p, err := ParseAnnotated("michael[1] discusses[1→2] project[2]")
	if err != nil {
		t.Fatal(err)
	}
	present := d.CausalNodes().Persist(p, 1)
	set := d.CausalNodes().MintSet(present.Cell.ID(), 1)
	d.ActivateSet(set.ID)
	d.Episodes().Create("skyra:episode:1", p.Sentence, 1)
	entity := d.Nodes().EnsureEntity("michael")
	d.Nodes().AppendInteraction(entity, InteractionRecord{From: "michael", To: "skyra", Utterance: "an update", Timestamp: 2})

	op, _ := ParseAnnotated("michael[1] reports[1→2] update[2]")
	observation := d.CausalNodes().PersistObserved(op, 2)
	root := d.Transcribe(string(StepMidlevelFrame), string(StepMidlevelFrame), "", nil, "X", nil)
	root.Nodes = d.Nodes()
	root.ObservationID = observation.Cell.ID()
	return d, root, set, observation
}

func TestStructuralRetainPreservesActivePresent(t *testing.T) {
	d, root, before, _ := seedStructuralRuntime(t)
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		if strings.Contains(system, "evolving being") {
			return "<retain/>", nil
		}
		return "<grades></grades>", nil
	})

	midlevelFrame(root, nil, nil)
	after, _ := d.ActiveSet()
	if root.Condition != "framed" || after.PresentID != before.PresentID || after.EfficientEdgeID == "" || after.FinalEdgeID == "" {
		t.Fatalf("retain changed active state: before=%+v after=%+v condition=%q", before, after, root.Condition)
	}
}

func TestStructuralComposeMintsExplicitAncestry(t *testing.T) {
	d, root, before, observation := seedStructuralRuntime(t)
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		if strings.Contains(system, "evolving being") {
			return "<compose>{prior situation}[1=" + before.PresentID + "] incorporates[1→2] update[2]</compose>", nil
		}
		return "<grades></grades>", nil
	})

	midlevelFrame(root, nil, nil)
	after, _ := d.ActiveSet()
	if root.Condition != "framed" || after.PresentID == before.PresentID {
		t.Fatalf("compose did not replace present: before=%+v after=%+v", before, after)
	}
	var ancestry *EdgeCell
	for _, child := range d.CausalNodes().Get(after.PresentID).ChildrenIn(PortInherited) {
		edge, _ := child.(*EdgeCell)
		if edge != nil && edge.Record.FromID == before.PresentID {
			ancestry = edge
		}
	}
	if ancestry == nil {
		t.Fatal("compose did not preserve its explicit prior-cause ancestry")
	}
	if ancestry.ChildrenIn(PortEvidence)[observation.Cell.ID()] != observation.Cell {
		t.Fatal("composition ancestry has no observation provenance")
	}
}

func TestStructuralCandidatesAllocateAmplitudeAndBoundComposition(t *testing.T) {
	d, _, before, observation := seedStructuralRuntime(t)
	offered := structuralCandidateIDs(d.CausalNodes(), before.PresentID, observation.Cell.ID())
	if offered[before.PresentID] <= 0 || offered[observation.Cell.ID()] <= 0 {
		t.Fatalf("active/observation candidates were not offered: %v", offered)
	}
	total := 0.0
	for _, amplitude := range offered {
		total += amplitude
	}
	if total < 0.999999 || total > 1.000001 {
		t.Fatalf("candidate amplitude = %f, want normalized 1", total)
	}

	otherParsed, _ := ParseAnnotated("unrelated[1] describes[1→2] thing[2]")
	other := d.CausalNodes().Persist(otherParsed, 3)
	raw := "<compose>{unoffered cause}[1=" + other.Cell.ID() + "] changes[1→2] topic[2]</compose>"
	if _, _, err := validateStructuralOp(d.CausalNodes(), raw, true, offered); err == nil || !strings.Contains(err.Error(), "was not offered") {
		t.Fatalf("unoffered composition reference error = %v", err)
	}
}

func TestStructuralRequeryIsStandalone(t *testing.T) {
	d, root, before, _ := seedStructuralRuntime(t)
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		if strings.Contains(system, "evolving being") {
			return "<requery>michael[1] changes[1→2] topic[2]</requery>", nil
		}
		return "<grades></grades>", nil
	})

	midlevelFrame(root, nil, nil)
	after, _ := d.ActiveSet()
	if after.PresentID == before.PresentID {
		t.Fatal("requery did not replace the active present")
	}
	if inherited := d.CausalNodes().Get(after.PresentID).ChildrenIn(PortInherited); len(inherited) != 0 {
		t.Fatalf("standalone requery has ancestry: %v", inherited)
	}
}

func TestStructuralFailureRetainsPriorState(t *testing.T) {
	d, root, before, _ := seedStructuralRuntime(t)
	var calls atomic.Int32
	d.SetInfer(func(system, present string) (string, error) {
		if response, ok := causalTestResponse(system); ok {
			return response, nil
		}
		if strings.Contains(system, "evolving being") {
			calls.Add(1)
			return "<compose>michael[1] changes[1→2] topic[2]</compose>", nil // no causal reference
		}
		return "<grades></grades>", nil
	})

	midlevelFrame(root, nil, nil)
	after, _ := d.ActiveSet()
	if root.Condition != "query-format-error" || calls.Load() != 3 {
		t.Fatalf("condition=%q calls=%d", root.Condition, calls.Load())
	}
	if after.ID != before.ID || after.PresentID != before.PresentID {
		t.Fatalf("failed operation changed active state: before=%+v after=%+v", before, after)
	}
}
