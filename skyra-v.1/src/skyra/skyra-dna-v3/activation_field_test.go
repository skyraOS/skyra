package dnav3

import (
	"fmt"
	"strings"
	"testing"
)

func TestActivationFieldDecayCurve(t *testing.T) {
	f := NewActivationField(DefaultActivationConfig())
	f.BumpSpan("s1", 1)
	want := []float64{0.5, 0.25, 0.125, 0.0625}
	for i, w := range want {
		f.BeginImpulse()
		if got := f.Spans["s1"]; got != w {
			t.Fatalf("decay %d = %.4f, want %.4f", i+1, got, w)
		}
	}
	f.BeginImpulse()
	if _, alive := f.Spans["s1"]; alive {
		t.Fatalf("activation below epsilon was not pruned: %v", f.Spans)
	}
}

func TestMemoryNodesCarryTokenCostAndDensityControlsSelection(t *testing.T) {
	b := NewBeing(nil, nil)
	b.CountTokens = func(s string) int { return len(strings.Fields(s)) }
	o := b.Observe(Triple{Subject: "project:skyra", Predicate: "uses", Object: "activation"}, ProvenanceRaw)
	if o.TokenCount != 3 {
		t.Fatalf("observation TokenCount=%d, want 3", o.TokenCount)
	}
	if got := b.Spans.SpanTokens(o.SpanID); got != o.TokenCount {
		t.Fatalf("span tokens=%d, want %d", got, o.TokenCount)
	}
	e := &Expression{
		SpanID: o.SpanID,
		Acts:   []Act{{ID: "a1", Primitive: "remember", Target: "project:skyra"}},
		Expects: []Expect{{
			Of: []string{"a1"}, Within: 1,
			Triple: Triple{Subject: "project:skyra", Predicate: "state", Object: "remembered"},
		}},
	}
	b.Expr.Place("token-expression", e)
	if e.TokenCount <= 0 {
		t.Fatal("expression was placed without a token cost")
	}

	nodes := []MemoryNode{
		{ID: "large", TokenCount: 10, Activation: 1, Text: "large"},
		{ID: "dense", TokenCount: 2, Activation: 0.8, Text: "dense"},
		{ID: "weak", TokenCount: 2, Activation: 0.1, Text: "weak"},
	}
	selected := SelectMemory(nodes, 10)
	if len(selected) != 2 || selected[0].ID != "dense" || selected[1].ID != "weak" {
		t.Fatalf("density selection=%v, want dense+weak within budget", selected)
	}
}

func TestColdMemoryDisappearsAndEntityReactivatesIt(t *testing.T) {
	b := NewBeing(nil, nil)
	old := Triple{
		Subject: "project:skyra", Predicate: "decision", Object: "activation-replaces-context",
	}
	b.ObserveImpulse([]Triple{old}, nil)
	oldSpan := b.CurrentSpanID()

	for i := 0; i < 5; i++ {
		b.ObserveImpulse([]Triple{{
			Subject:   fmt.Sprintf("unrelated:%d", i),
			Predicate: "mentions",
			Object:    fmt.Sprintf("topic:%d", i),
		}}, nil)
	}
	if b.Activation.SpanActive(oldSpan) {
		t.Fatalf("old region remained active after decay: %.4f", b.Activation.Spans[oldSpan])
	}
	for _, node := range b.ActiveMemory(nil) {
		if strings.Contains(node.Text, old.Object) {
			t.Fatal("cold memory remained in the active projection")
		}
	}

	b.ObserveImpulse([]Triple{{
		Subject: "project:skyra", Predicate: "mentioned-again", Object: "today",
	}}, nil)
	if !b.Activation.SpanActive(oldSpan) {
		t.Fatal("entity reverse lookup did not reactivate the old span")
	}
	found := false
	for _, node := range b.ActiveMemory(nil) {
		found = found || strings.Contains(node.Text, old.Object)
	}
	if !found {
		t.Fatal("reactivated memory was not rendered into the bounded present")
	}
}

func TestDirectMatchIsNotDilutedByLargeSpan(t *testing.T) {
	b := NewBeing(nil, nil)
	var wanted Observation
	for i := 0; i < 100; i++ {
		o := b.Observe(Triple{
			Subject:   fmt.Sprintf("entity:%d", i),
			Predicate: "appeared-in-large-span",
			Object:    fmt.Sprintf("value:%d", i),
		}, ProvenanceRaw)
		if i == 73 {
			wanted = o
		}
	}
	for i := 0; i < 5; i++ {
		b.ObserveImpulse([]Triple{{
			Subject: fmt.Sprintf("elsewhere:%d", i), Predicate: "unrelated", Object: "noise",
		}}, nil)
	}
	if b.Activation.Nodes[string(wanted.ID)] != 0 {
		t.Fatal("fixture did not let the large-span node decay")
	}
	b.ObserveImpulse([]Triple{{
		Subject: wanted.Subject, Predicate: "returned", Object: "now",
	}}, nil)
	if got := b.Activation.Nodes[string(wanted.ID)]; got < 0.5 {
		t.Fatalf("directly matched node activation=%.3f, want >=0.5 independent of 100-node span", got)
	}
}

func TestActivePredictionCanResolveAcrossCausallyRetrievedSpan(t *testing.T) {
	b := NewBeing(nil, nil)
	claim := Triple{
		Mode: ModeExact, Confidence: 0.9,
		Subject: "device:lamp", Predicate: "state", Object: "on",
	}
	b.ObserveImpulse([]Triple{{
		Subject: "device:lamp", Predicate: "commanded", Object: "on",
	}}, nil)
	owner := b.CurrentSpanID()
	e := &Expression{
		EpisodeID: b.EpisodeID(), SpanID: owner,
		Acts:    []Act{{ID: "a1", Primitive: "switch", Target: "device:lamp"}},
		Expects: []Expect{{Of: []string{"a1"}, Within: 2, Triple: claim}},
	}
	b.Expr.Place("active-bet", e)
	e.Complete("a1", b.tick)

	b.ObserveImpulse([]Triple{claim}, nil)
	if e.Expects[0].Resolution != TruthTrue {
		t.Fatal("incoming roots activated the owner span but its armed prediction did not resolve")
	}
}

func TestHundredKHistoryUsesFixedProjectionWithoutCompaction(t *testing.T) {
	b := NewBeing(nil, nil)
	b.Activation.Config.MemoryTokens = 256
	const impulses = 4_000
	var first Triple
	totalTokens := 0
	for i := 0; i < impulses; i++ {
		triple := Triple{
			Subject:   fmt.Sprintf("thread:%d", i),
			Predicate: "said",
			Object:    fmt.Sprintf("this-is-a-deliberately-substantial-memory-node-number-%d-with-unique-context", i),
		}
		if i == 0 {
			first = triple
		}
		b.ObserveImpulse([]Triple{triple}, nil)
		totalTokens += b.Spans.SpanTokens(b.CurrentSpanID())
	}
	if totalTokens <= 100_000 {
		t.Fatalf("fixture stored only %d tokens, want >100k", totalTokens)
	}
	active := b.ActiveMemory(nil)
	spent := 0
	for _, node := range active {
		spent += node.TokenCount
	}
	if spent > b.Activation.Config.MemoryTokens {
		t.Fatalf("active projection spent %d tokens, budget %d", spent, b.Activation.Config.MemoryTokens)
	}
	for _, node := range active {
		if strings.Contains(node.Text, first.Object) {
			t.Fatal("the first cold memory remained present across 100k history")
		}
	}

	b.ObserveImpulse([]Triple{{
		Subject: first.Subject, Predicate: "reopened", Object: "now",
	}}, nil)
	found := false
	for _, node := range b.ActiveMemory(nil) {
		found = found || strings.Contains(node.Text, first.Object)
	}
	if !found {
		t.Fatal("the first memory did not return by activation after 100k history")
	}
}
