package checkpoint

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func stamp(n int) string { return fmt.Sprintf("%064x", n) }
func fixture(t *testing.T) (Store, Checkpoint) {
	t.Helper()
	s := Store{t.TempDir()}
	c, e := s.Initialize("c-1", stamp(1), Contract{Target: "user", Schema: json.RawMessage(`{"type":"object"}`)})
	if e != nil {
		t.Fatal(e)
	}
	return s, c
}
func evidence(frontier []string, records map[string]Record) Evidence {
	return Evidence{Version: 1, Root: stamp(1), Frontier: frontier, Records: records}
}
func TestConcurrentBranchesRequireReconstructionAndNeverRegress(t *testing.T) {
	s, c := fixture(t)
	records := map[string]Record{stamp(1): {}, stamp(2): {Predecessor: stamp(1)}, stamp(3): {Predecessor: stamp(1)}}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, n := range []int{2, 3} {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			_, e := s.Advance("c-1", c, evidence([]string{stamp(n)}, records))
			results <- e
		}(n)
	}
	wg.Wait()
	close(results)
	wins, conflicts := 0, 0
	for e := range results {
		if e == nil {
			wins++
		} else if errors.Is(e, ErrConflict) {
			conflicts++
		} else {
			t.Fatal(e)
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("wins %d conflicts %d", wins, conflicts)
	}
	current, e := s.Read("c-1")
	if e != nil {
		t.Fatal(e)
	}
	both, e := s.Advance("c-1", current, evidence([]string{stamp(2), stamp(3)}, records))
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(both.Leaves, []string{stamp(2), stamp(3)}) {
		t.Fatalf("lost sibling: %+v", both)
	}
	records[stamp(4)] = Record{Predecessor: stamp(2)}
	partial, e := s.Advance("c-1", both, evidence([]string{stamp(4), stamp(3)}, records))
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(partial.Leaves, []string{stamp(3), stamp(4)}) {
		t.Fatalf("slow branch lost: %+v", partial)
	}
	records[stamp(5)] = Record{Predecessor: stamp(3), Deps: []string{stamp(4), stamp(3)}}
	joined, e := s.Advance("c-1", partial, evidence([]string{stamp(5)}, records))
	if e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(joined.Leaves, []string{stamp(5)}) {
		t.Fatalf("join did not cover leaves: %+v", joined)
	}
	if _, e := s.Advance("c-1", both, evidence([]string{stamp(2)}, records)); !errors.Is(e, ErrConflict) {
		t.Fatalf("stale writer restored superseded leaf: %v", e)
	}
	// Reconciliation of accepted progress published before the checkpoint is
	// the same compare/update with fresh syscall evidence; it runs no effects.
	records[stamp(6)] = Record{Predecessor: stamp(5)}
	recovered, e := s.Advance("c-1", joined, evidence([]string{stamp(6)}, records))
	if e != nil || !reflect.DeepEqual(recovered.Leaves, []string{stamp(6)}) {
		t.Fatalf("crash-gap recovery: %+v %v", recovered, e)
	}
}
func TestCheckpointRecoveryMetadataOnlyAndIdempotentInitialization(t *testing.T) {
	s, c := fixture(t)
	records := map[string]Record{stamp(1): {}, stamp(2): {Predecessor: stamp(1)}}
	advanced, e := s.Advance("c-1", c, evidence([]string{stamp(2), stamp(2)}, records))
	if e != nil {
		t.Fatal(e)
	}
	loaded, e := s.Initialize("c-1", c.RootInvocation, c.Contract)
	if e != nil || loaded.Key() != advanced.Key() {
		t.Fatal("initialization regressed an existing checkpoint")
	}
	if _, e := s.Initialize("c-1", stamp(9), c.Contract); !errors.Is(e, ErrConflict) {
		t.Fatal("anchor changed")
	}
	restarted := Store{Dir: s.Dir}
	again, e := restarted.Read("c-1")
	if e != nil || again.Key() != advanced.Key() {
		t.Fatal("checkpoint was not durable")
	}
	b, _ := json.Marshal(again)
	var fields map[string]any
	json.Unmarshal(b, &fields)
	if len(fields) != 3 {
		t.Fatalf("checkpoint stores a view: %s", b)
	}
}
func TestCheckpointRejectsIncompleteEvidence(t *testing.T) {
	s, c := fixture(t)
	for _, ev := range []Evidence{{}, evidence(nil, map[string]Record{stamp(1): {}}), evidence([]string{stamp(2)}, map[string]Record{stamp(1): {}})} {
		if _, e := s.Advance("c-1", c, ev); e == nil {
			t.Fatal("invalid evidence accepted")
		}
	}
	if _, e := s.Read("../escape"); e == nil {
		t.Fatal("path escaped checkpoint store")
	}
}
