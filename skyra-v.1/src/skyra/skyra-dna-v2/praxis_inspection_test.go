package dnav2

import (
	"fmt"
	"testing"
)

func TestPraxisInspectionShowsUngradedAndZeroStrengthHistory(t *testing.T) {
	d := testDNA()
	present := persistTestCause(t, d, "current[1] describes[1→2] situation[2]", 1)
	efficient := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	final := persistTestCause(t, d, "situation[1] enables[1→2] outcome[2]", 1)
	set, err := d.CausalNodes().MintFullSet(present.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 1)
	if err != nil {
		t.Fatal(err)
	}
	d.ActivateSet(set.ID)
	d.Signed().Grade(set.EfficientEdgeID, 0, 1)

	snapshot := d.InspectPraxis(set.ID)
	if snapshot.ActiveSetID != set.ID || len(snapshot.Sets) != 1 {
		t.Fatalf("snapshot=%+v", snapshot)
	}
	edges := map[string]PraxisEdgeView{}
	for _, edge := range snapshot.Edges {
		edges[edge.ID] = edge
	}
	if edge := edges[set.EfficientEdgeID]; !edge.HasGrade || edge.Strength != 0 || edge.GradeCount != 1 {
		t.Fatalf("explicit zero missing: %+v", edge)
	}
	if edge := edges[set.FinalEdgeID]; edge.HasGrade {
		t.Fatalf("ungraded edge reported graded: %+v", edge)
	}
	if len(snapshot.Sets[0].Amplitude) != 0 {
		t.Fatalf("zero/absent edge got amplitude: %v", snapshot.Sets[0].Amplitude)
	}

	hops := d.WalkHistory(present.Cell.ID(), 2, PortEfficient, PortFinal, PortFrom, PortTo)
	found := false
	for _, hop := range hops {
		if hop.ToID == set.EfficientEdgeID {
			found = true
		}
	}
	if !found {
		t.Fatalf("deterministic history hid zero-strength edge: %+v", hops)
	}
}

func TestPraxisInspectionFiltersContextsWithoutLosingSharedNodes(t *testing.T) {
	d := testDNA()
	present := persistTestCause(t, d, "current[1] describes[1→2] situation[2]", 1)
	efficient := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	final := persistTestCause(t, d, "situation[1] enables[1→2] outcome[2]", 1)
	first, _ := d.CausalNodes().MintFullSet(present.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 1)
	second, _ := d.CausalNodes().MintFullSet(present.Cell.ID(), efficient.Cell.ID(), final.Cell.ID(), 2)

	snapshot := d.InspectPraxis(second.ID)
	if len(snapshot.Sets) != 1 || snapshot.Sets[0].ID != second.ID {
		t.Fatalf("sets=%+v", snapshot.Sets)
	}
	edges := map[string]bool{}
	for _, edge := range snapshot.Edges {
		edges[edge.ID] = true
	}
	if edges[first.EfficientEdgeID] || !edges[second.EfficientEdgeID] {
		t.Fatalf("context edge filter=%v", edges)
	}
	if len(snapshot.Nodes) != 3 {
		t.Fatalf("shared endpoint nodes missing: %+v", snapshot.Nodes)
	}
	if all := d.InspectPraxis(""); len(all.Sets) != 2 {
		t.Fatalf("unfiltered sets=%d", len(all.Sets))
	}
}

func TestFinalClosureCheckerIsReadOnlyAndUsesLaterEfficientLineage(t *testing.T) {
	d := testDNA()
	present1 := persistTestCause(t, d, "current[1] describes[1→2] situation[2]", 1)
	prior1 := persistTestCause(t, d, "prior[1] produces[1→2] situation[2]", 1)
	target := persistTestCause(t, d, "team[1] restores[1→2] service[2]", 1)
	first, _ := d.CausalNodes().MintFullSet(present1.Cell.ID(), prior1.Cell.ID(), target.Cell.ID(), 1)

	// The later efficient compound explicitly cites the earlier Final target.
	parsed, err := ParseAnnotated(fmt.Sprintf("{restoration progress}[1=%s] explains[1→2] recovery[2]", target.Cell.ID()))
	if err != nil {
		t.Fatal(err)
	}
	laterEfficient := d.CausalNodes().Persist(parsed, 2)
	present2 := persistTestCause(t, d, "service[1] enters[1→2] recovery[2]", 2)
	final2 := persistTestCause(t, d, "recovery[1] enables[1→2] stability[2]", 2)
	later, _ := d.CausalNodes().MintFullSet(present2.Cell.ID(), laterEfficient.Cell.ID(), final2.Cell.ID(), 2)

	matches := d.FinalClosureCandidates()
	if len(matches) != 1 || matches[0].FinalSetID != first.ID || matches[0].LaterSetID != later.ID || matches[0].TargetID != target.Cell.ID() {
		t.Fatalf("matches=%+v", matches)
	}
	if _, ok := d.Signed().Strength(first.FinalEdgeID, 2); ok {
		t.Fatal("closure checker mutated signed state")
	}
}
