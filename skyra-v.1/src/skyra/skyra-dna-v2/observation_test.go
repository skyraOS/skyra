package dnav2

import (
	"strings"
	"sync/atomic"
	"testing"

	"skyra-v1/src/skyra"
)

func TestUtteranceDecompositionPersistsObservedSubgraph(t *testing.T) {
	d := testDNA()
	observed := d.Nodes().EnsureEntity("michael")
	d.SetInfer(func(system, present string) (string, error) {
		if !strings.Contains(system, "decompose one observed utterance") {
			t.Fatalf("unexpected system prompt: %s", system)
		}
		return "michael[1] reports[1→2] {the api outage}[2]", nil
	})

	observation := decomposeUtterance(d, "michael", "skyra", "the api is down", 1)
	if observation == nil {
		t.Fatal("valid decomposition did not persist")
	}
	content, ok := observation.Cell.Content().(CausalNodeContent)
	if !ok || content.Provenance != ProvenanceObserved {
		t.Fatalf("observation content = %+v", observation.Cell.Content())
	}
	if content.Sentence != "michael reports the api outage" {
		t.Fatalf("observation sentence = %q", content.Sentence)
	}
	if len(observation.Cell.ChildrenIn(PortFormal)) != 1 || len(observation.Cell.ChildrenIn(PortMaterial)) != 2 {
		t.Fatal("observation birth subgraph was not published")
	}
	entityContent, _ := observed.Content().(Entity)
	if observation.Parents[0] != observed || entityContent.Provenance != ProvenanceObserved {
		t.Fatal("observed speaker must remain the observation-born semantic entity")
	}
	term := d.Nodes().Term("the api outage")
	termContent, _ := term.Content().(TermContent)
	if termContent.Provenance != ProvenanceAsserted {
		t.Fatalf("term provenance = %q", termContent.Provenance)
	}
}

func TestUtteranceDecompositionRetriesTransactionally(t *testing.T) {
	d := testDNA()
	var calls atomic.Int32
	d.SetInfer(func(system, present string) (string, error) {
		calls.Add(1)
		if strings.Contains(present, "could not be parsed") {
			return "michael[1] reports[1→2] outage[2]", nil
		}
		return "michael[1] reports[1→9] outage[2]", nil
	})

	observation := decomposeUtterance(d, "michael", "skyra", "outage", 1)
	if observation == nil || calls.Load() != 2 {
		t.Fatalf("observation=%v calls=%d, want corrected commit after two calls", observation, calls.Load())
	}
}

func TestCauseAssessmentRetriesMalformedGrade(t *testing.T) {
	var calls atomic.Int32
	grade, ok := inferCauseGrade(func(system, present string) (string, error) {
		calls.Add(1)
		if strings.Contains(present, "previous response") {
			return "<grade>0.6</grade>", nil
		}
		return "probably supported", nil
	}, "active and observation")
	if !ok || grade != 0.6 || calls.Load() != 2 {
		t.Fatalf("grade=%v ok=%v calls=%d", grade, ok, calls.Load())
	}
}

func TestCauseAssessmentNoGradeIsValidAndDistinctFromZero(t *testing.T) {
	var calls atomic.Int32
	if grade, ok := inferCauseGrade(func(system, present string) (string, error) {
		calls.Add(1)
		return "<no-grade/>", nil
	}, "unresolved"); ok || grade != 0 || calls.Load() != 1 {
		t.Fatalf("no-grade => grade=%v ok=%v calls=%d", grade, ok, calls.Load())
	}
	grade, hasGrade, valid := parseCauseGrade("<grade>0</grade>")
	if !valid || !hasGrade || grade != 0 {
		t.Fatalf("grade zero => grade=%v hasGrade=%v valid=%v", grade, hasGrade, valid)
	}
}

func TestObservationGradesOnlyEfficientAndFinal(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	var assessmentCalls atomic.Int32
	d.SetInfer(func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "decompose one observed utterance"):
			if strings.Contains(present, "changed") {
				return "michael[1] corrects[1→2] {the deadline}[2]", nil
			}
			return "michael[1] states[1→2] {the deadline}[2]", nil
		case strings.Contains(system, "grade one ACTIVE CAUSE"):
			assessmentCalls.Add(1)
			return "<grade>-0.8</grade>", nil
		case strings.Contains(system, "evolving being"):
			return "<requery>michael[1] discusses[1→2] {the deadline}[2]</requery>", nil
		default:
			return gradeMock("<match>none</match>")(system, present)
		}
	})
	d.SetEmit(func(skyra.Impulse) {})
	o, board := newOrch(d)

	fireTurn(t, o, "X1", "", "the deadline is Friday")
	first, ok := d.ActiveSet()
	if !ok {
		t.Fatal("first turn did not establish an active cause")
	}
	fireTurn(t, o, "X2", "X1", "the deadline changed to Monday")
	if assessmentCalls.Load() != 2 {
		t.Fatalf("Efficient/Final grading calls=%d, want 2", assessmentCalls.Load())
	}

	observationID := board.Row("X2", "").ObservationID()
	observation := d.CausalNodes().Get(observationID)
	if observation == nil {
		t.Fatal("impulse row did not retain its observation CausalNode")
	}
	assessments := d.CausalNodes().AssessmentsOf(observationID)
	if len(assessments) != 2 {
		t.Fatalf("assessments = %+v, want Efficient and Final", assessments)
	}
	for _, assessment := range assessments {
		if assessment.Role != CauseEfficient && assessment.Role != CauseFinal {
			t.Fatalf("Material/Formal entered signed grading: %+v", assessment)
		}
	}

	second, _ := d.ActiveSet()
	if second.PresentID == first.PresentID {
		t.Fatal("requery should replace the active present representation")
	}
	if inherited := d.CausalNodes().Get(second.PresentID).ChildrenIn(PortInherited); len(inherited) != 0 {
		t.Fatalf("standalone requery invented composition ancestry: %v", inherited)
	}
}
