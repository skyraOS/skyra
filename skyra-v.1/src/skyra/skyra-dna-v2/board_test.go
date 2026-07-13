package dnav2

import (
	"sync"
	"testing"
)

// The join: motor fires only when done ⊇ expects, and fires EXACTLY once even
// though both fan-out branches race the check on completion.
func TestBoardJoinFiresOnceWhenComplete(t *testing.T) {
	b := NewBoard()
	r := b.Row("X", "")
	r.Declare("episode", "self", "other") // midlevel declares the expected set

	// not satisfied yet
	if r.TryFireMotor() {
		t.Fatal("motor fired before the join was satisfied")
	}
	r.Mark("episode", "self")
	if r.TryFireMotor() {
		t.Fatal("motor fired with only self done")
	}
	r.Mark("episode", "other")

	// both agents complete concurrently and both race TryFireMotor — exactly one wins
	var fires int32
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if r.TryFireMotor() {
				mu.Lock()
				fires++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if fires != 1 {
		t.Fatalf("motor fired %d times, want exactly 1 (atomic join guard failed)", fires)
	}
}

// The one-reframe rail: the first reframe is allowed, later ones suppressed.
func TestBoardOneReframeRail(t *testing.T) {
	b := NewBoard()
	r := b.Row("X", "")
	if !r.TryReframe() {
		t.Fatal("first reframe should be allowed")
	}
	if r.TryReframe() {
		t.Fatal("second reframe must be suppressed (the rail)")
	}
}

// Rows track the causal chain and pending/done lifecycle.
func TestBoardCausalChainAndLifecycle(t *testing.T) {
	b := NewBoard()
	x := b.Row("X", "")  // fresh external event
	y := b.Row("Y", "X") // reafference — links back to X
	if y.Prev != "X" {
		t.Fatalf("causal link = %q, want X", y.Prev)
	}
	if len(b.Pending()) != 2 {
		t.Fatalf("pending = %d, want 2", len(b.Pending()))
	}
	x.Append(Span{Step: "other:grade", Condition: "held", At: 1})
	x.Append(Span{Step: "motor", Condition: "emitted", At: 2})
	trace := x.Done()
	if len(trace) != 2 {
		t.Fatalf("flushed trace = %d spans, want 2", len(trace))
	}
	if len(b.Pending()) != 1 {
		t.Fatalf("pending after X done = %d, want 1 (Y)", len(b.Pending()))
	}
}

// Concurrent marks + history appends are race-clean (run with -race).
func TestBoardConcurrentAccess(t *testing.T) {
	b := NewBoard()
	r := b.Row("X", "")
	r.Declare("episode", "self", "other")
	var wg sync.WaitGroup
	wg.Add(3)
	go func() { defer wg.Done(); r.Mark("episode", "self") }()
	go func() { defer wg.Done(); r.Mark("episode", "other") }()
	go func() { defer wg.Done(); r.Append(Span{Step: "x", At: 1}) }()
	wg.Wait()
	if !r.TryFireMotor() {
		t.Fatal("join should be satisfied after both marks")
	}
}
