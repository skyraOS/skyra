package dnav2

import (
	"strings"
	"sync"
	"testing"
)

func persistTestCause(t *testing.T, d *DNA, sentence string, ts int) *CausalNode {
	t.Helper()
	parsed, err := ParseAnnotated(sentence)
	if err != nil {
		t.Fatal(err)
	}
	return d.CausalNodes().Persist(parsed, ts)
}

func TestCausalExpansionWorkersStartInParallel(t *testing.T) {
	d := testDNA()
	present := persistTestCause(t, d, "situation[1] affects[1→2] system[2]", 1)
	started := make(chan string, 2)
	release := make(chan struct{})
	d.SetInfer(func(system, present string) (string, error) {
		role := ""
		switch {
		case strings.Contains(system, "emit the efficient"):
			role = CauseEfficient
		case strings.Contains(system, "emit the final"):
			role = CauseFinal
		}
		started <- role
		<-release
		if role == CauseEfficient {
			return "prior[1] produces[1→2] situation[2]", nil
		}
		return "situation[1] enables[1→2] outcome[2]", nil
	})
	done := make(chan error, 1)
	go func() {
		_, _, err := expandCausalBeliefs(d, nil, present.Cell.ID())
		done <- err
	}()
	roles := map[string]bool{<-started: true, <-started: true}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if !roles[CauseEfficient] || !roles[CauseFinal] {
		t.Fatalf("workers did not both start: %v", roles)
	}
}

// Fresh edges never inherit signed state (item 6, option B); the LINEAGE
// budget deliberately survives the set boundary — the voice keeps its earned
// specificity even though its edges die with their contexts.
func TestMintFullSetFreshEdgesDoNotInheritSignedStateButLineageBudgetPersists(t *testing.T) {
	d := testDNA()
	p1 := persistTestCause(t, d, "first[1] describes[1→2] situation[2]", 1)
	p2 := persistTestCause(t, d, "second[1] describes[1→2] situation[2]", 2)
	efficient := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	final := persistTestCause(t, d, "situation[1] enables[1→2] outcome[2]", 1)
	first, err := d.CausalNodes().MintFullSet(p1.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 1)
	if err != nil {
		t.Fatal(err)
	}
	d.Signed().Grade(first.EfficientEdgeID, 1, 1)
	_, _ = d.QueryBudgets().ApplyGrade(CauseEfficient, first.EfficientEdgeID, 1)

	second, err := d.CausalNodes().MintFullSet(p2.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if first.EfficientEdgeID == second.EfficientEdgeID || first.FinalEdgeID == second.FinalEdgeID {
		t.Fatal("new AitiaSet reused contextual edge identity")
	}
	if _, ok := d.Signed().Strength(second.EfficientEdgeID, 2); ok {
		t.Fatal("fresh edge inherited signed state")
	}
	if budget, existed := d.QueryBudgets().Current(CauseEfficient); !existed || budget != 138 {
		t.Fatalf("lineage budget=%v existed=%v, want earned 138 carried across sets", budget, existed)
	}
}

func TestActiveCausalGradesUpdateExactEdgesAndNoGradeDoesNotSampleEMA(t *testing.T) {
	d := testDNA()
	present := persistTestCause(t, d, "current[1] describes[1→2] situation[2]", 1)
	efficient := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	final := persistTestCause(t, d, "situation[1] enables[1→2] outcome[2]", 1)
	set, err := d.CausalNodes().MintFullSet(present.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 1)
	if err != nil {
		t.Fatal(err)
	}
	d.ActivateSet(set.ID)
	observedParsed, _ := ParseAnnotated("michael[1] reports[1→2] update[2]")
	observation := d.CausalNodes().PersistObserved(observedParsed, 2)
	d.SetInfer(func(system, present string) (string, error) {
		if strings.Contains(present, "ACTIVE CAUSE (efficient)") {
			return "<grade>0.5</grade>", nil
		}
		return "<no-grade/>", nil
	})

	assessments := d.AssessActiveCauses(observation.Cell.ID(), nil)
	if len(assessments) != 2 {
		t.Fatalf("assessments=%+v", assessments)
	}
	if strength, ok := d.Signed().Strength(set.EfficientEdgeID, d.Clock()); !ok || strength <= 0 {
		t.Fatalf("efficient strength=%v ok=%v", strength, ok)
	}
	if _, ok := d.Signed().Strength(set.FinalEdgeID, d.Clock()); ok {
		t.Fatal("no-grade sampled Final EMA")
	}
	if budget, _ := d.QueryBudgets().Current(CauseEfficient); budget != 133 {
		t.Fatalf("efficient lineage budget=%v", budget)
	}
	if budget, _ := d.QueryBudgets().Current(CauseFinal); budget != 128 {
		t.Fatalf("final no-grade lineage budget=%v", budget)
	}
	if history := d.QueryBudgets().History(CauseFinal); len(history) != 1 || history[0].EdgeID != set.FinalEdgeID {
		t.Fatalf("no-grade lost edge provenance: %+v", history)
	}
	for _, edgeID := range []string{set.EfficientEdgeID, set.FinalEdgeID} {
		edge := d.CausalNodes().Edge(edgeID)
		if edge.ChildrenIn(PortEvidence)[observation.Cell.ID()] != observation.Cell {
			t.Fatalf("edge %s missing evidence provenance", edgeID)
		}
	}
}

func TestCausalWorkerFailurePreservesPriorActiveSet(t *testing.T) {
	d, root, before, _ := seedStructuralRuntime(t)
	// Upgrade the seed so there is a complete prior set to preserve.
	efficient := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	final := persistTestCause(t, d, "situation[1] enables[1→2] outcome[2]", 1)
	full, _ := d.CausalNodes().MintFullSet(before.PresentID, efficient.Cell.ID(), final.Cell.ID(), 1)
	d.ActivateSet(full.ID)
	d.SetInfer(func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "evolving being"):
			return "<requery>replacement[1] describes[1→2] situation[2]</requery>", nil
		case strings.Contains(system, "emit the efficient"):
			return "not annotated", nil
		case strings.Contains(system, "emit the final"):
			return "situation[1] enables[1→2] outcome[2]", nil
		}
		return "<grades></grades>", nil
	})

	midlevelFrame(root, nil, nil)
	after, _ := d.ActiveSet()
	if root.Condition != "causal-format-error" || after.ID != full.ID {
		t.Fatalf("condition=%q before=%+v after=%+v", root.Condition, full, after)
	}
}

// The controller is closed-loop: a taxed lineage's next emission reads the
// reduced budget, and an over-budget emission is corrected into a coarser
// complete claim rather than committed or truncated.
func TestExpansionReadsLineageBudgetAndVerifiesEmissionSize(t *testing.T) {
	d := testDNA()
	present := persistTestCause(t, d, "situation[1] affects[1→2] system[2]", 1)
	// Tax the efficient voice hard: 128 * 0.5^1 = 64, twice → 32 (the floor).
	_, _ = d.QueryBudgets().ApplyGrade(CauseEfficient, "prior-edge", -1)
	_, _ = d.QueryBudgets().ApplyGrade(CauseEfficient, "prior-edge", -1)

	longEmission := "prior[1] produces[1→2] situation[2] " + strings.Repeat("with[2→3] {greatly elaborated causal detail}[3] ", 12)
	var sawBudget, corrected bool
	d.SetInfer(func(system, present string) (string, error) {
		switch {
		case strings.Contains(system, "emit the efficient"):
			if strings.Contains(system, "budget is 32") {
				sawBudget = true
			}
			if strings.Contains(present, "token budget") {
				corrected = true
				return "prior[1] produces[1→2] situation[2]", nil
			}
			return longEmission, nil
		case strings.Contains(system, "emit the final"):
			return "situation[1] enables[1→2] outcome[2]", nil
		}
		return "", nil
	})

	efficient, _, err := expandCausalBeliefs(d, nil, present.Cell.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !sawBudget {
		t.Fatal("efficient worker did not receive the lineage's reduced budget")
	}
	if !corrected {
		t.Fatal("over-budget emission was not returned for correction")
	}
	if efficient.Sentence != "prior produces situation" {
		t.Fatalf("committed emission = %q, want the corrected coarser claim", efficient.Sentence)
	}
}

func TestAllocateAmplitudeUsesMagnitudeAndZeroGetsNone(t *testing.T) {
	got := AllocateAmplitude(12, map[string]float64{"resonance": 1, "tension": -0.5, "defeated": 0})
	if got["resonance"] != 8 || got["tension"] != 4 {
		t.Fatalf("allocation=%v", got)
	}
	if _, ok := got["defeated"]; ok {
		t.Fatal("zero-strength edge received excitation")
	}
	if got := AllocateAmplitude(12, map[string]float64{"a": 0, "b": 0}); len(got) != 0 {
		t.Fatalf("zero-total allocation=%v", got)
	}
}

func TestMintFullSetConcurrentContexts(t *testing.T) {
	d := testDNA()
	present := persistTestCause(t, d, "current[1] describes[1→2] situation[2]", 1)
	efficient := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	final := persistTestCause(t, d, "situation[1] enables[1→2] outcome[2]", 1)
	var wg sync.WaitGroup
	ids := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			set, err := d.CausalNodes().MintFullSet(present.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 1)
			if err == nil {
				ids <- set.ID
			}
		}()
	}
	wg.Wait()
	close(ids)
	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("duplicate set ID %s", id)
		}
		seen[id] = true
	}
	if len(seen) != 20 {
		t.Fatalf("sets=%d, want 20", len(seen))
	}
}
