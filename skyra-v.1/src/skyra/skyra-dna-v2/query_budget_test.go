package dnav2

import (
	"math"
	"sync"
	"testing"
)

func budgetTestStore(t *testing.T) *QueryBudgetStore {
	t.Helper()
	store, err := NewQueryBudgetStore(QueryBudgetConfig{Min: 20, Max: 200, Initial: 100, Alpha: 10, Beta: 0.25})
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestQueryBudgetContinuousAIMD(t *testing.T) {
	store := budgetTestStore(t)
	if got, existed := store.Current(CauseEfficient); got != 100 || existed {
		t.Fatalf("cold budget=%v existed=%v", got, existed)
	}
	if event, err := store.ApplyGrade(CauseEfficient, "e1", 1); err != nil || event.After != 110 {
		t.Fatalf("+1 event=%+v err=%v", event, err)
	}
	if event, err := store.ApplyGrade(CauseEfficient, "e2", 0.5); err != nil || event.After != 115 {
		t.Fatalf("+0.5 event=%+v err=%v", event, err)
	}
	// beta^(-g) at g=-0.5 = sqrt(0.25) = 0.5.
	if event, err := store.ApplyGrade(CauseEfficient, "e3", -0.5); err != nil || event.After != 57.5 {
		t.Fatalf("-0.5 event=%+v err=%v", event, err)
	}
	if event, err := store.ApplyGrade(CauseEfficient, "e4", -1); err != nil || event.After != 20 {
		t.Fatalf("-1/clamp event=%+v err=%v", event, err)
	}
}

func TestQueryBudgetGradeZeroAndNoGradeAreDistinct(t *testing.T) {
	store := budgetTestStore(t)
	zero, err := store.ApplyGrade(CauseFinal, "e1", 0)
	if err != nil {
		t.Fatal(err)
	}
	absent, err := store.NoGrade(CauseFinal, "e2")
	if err != nil {
		t.Fatal(err)
	}
	if zero.After != 100 || absent.After != 100 || !zero.HasGrade || absent.HasGrade {
		t.Fatalf("zero=%+v no-grade=%+v", zero, absent)
	}
	history := store.History(CauseFinal)
	if len(history) != 2 || !history[0].HasGrade || history[1].HasGrade {
		t.Fatalf("history lost zero/no-grade distinction: %+v", history)
	}
	if history[0].EdgeID != "e1" || history[1].EdgeID != "e2" || history[0].Lineage != CauseFinal {
		t.Fatalf("history lost sample provenance: %+v", history)
	}
}

func TestQueryBudgetBoundsAndLineageIndependence(t *testing.T) {
	store := budgetTestStore(t)
	for range 20 {
		if _, err := store.ApplyGrade(CauseEfficient, "e1", 1); err != nil {
			t.Fatal(err)
		}
	}
	if got, _ := store.Current(CauseEfficient); got != 200 {
		t.Fatalf("upper bound=%v", got)
	}
	if got, existed := store.Current(CauseFinal); got != 100 || existed {
		t.Fatalf("independent lineage budget=%v existed=%v", got, existed)
	}
	if _, err := store.ApplyGrade(CauseFinal, "e2", -1); err != nil {
		t.Fatal(err)
	}
	if got, _ := store.Current(CauseFinal); got != 25 {
		t.Fatalf("final lineage budget=%v", got)
	}
	if got, _ := store.Current(CauseEfficient); got != 200 {
		t.Fatalf("final update changed efficient budget=%v", got)
	}
}

func TestQueryBudgetRejectsInvalidInputWithoutState(t *testing.T) {
	for _, config := range []QueryBudgetConfig{
		{},
		{Min: 20, Max: 10, Initial: 15, Alpha: 1, Beta: 0.5},
		{Min: 10, Max: 20, Initial: 15, Alpha: 0, Beta: 0.5},
		{Min: 10, Max: 20, Initial: 15, Alpha: 1, Beta: 1},
	} {
		if _, err := NewQueryBudgetStore(config); err == nil {
			t.Fatalf("invalid config accepted: %+v", config)
		}
	}
	store := budgetTestStore(t)
	for _, grade := range []float64{-1.1, 1.1, math.NaN(), math.Inf(1)} {
		if _, err := store.ApplyGrade("bad", "e", grade); err == nil {
			t.Fatalf("invalid grade accepted: %v", grade)
		}
	}
	if _, err := store.ApplyGrade("", "e", 0.5); err == nil {
		t.Fatal("missing lineage accepted")
	}
	if len(store.All()) != 0 {
		t.Fatalf("invalid grades created state: %v", store.All())
	}
}

func TestQueryBudgetConcurrentLineages(t *testing.T) {
	store := budgetTestStore(t)
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lineage := CauseEfficient
			if i%2 == 1 {
				lineage = CauseFinal
			}
			_, _ = store.ApplyGrade(lineage, "e", 0.1)
		}(i)
	}
	wg.Wait()
	if len(store.History(CauseEfficient)) != 50 || len(store.History(CauseFinal)) != 50 {
		t.Fatalf("history counts efficient=%d final=%d", len(store.History(CauseEfficient)), len(store.History(CauseFinal)))
	}
}

func TestDNAOwnsQueryBudgetStore(t *testing.T) {
	d := testDNA()
	if got, existed := d.QueryBudgets().Current(CauseEfficient); got != 128 || existed {
		t.Fatalf("DNA default budget=%v existed=%v", got, existed)
	}
}
