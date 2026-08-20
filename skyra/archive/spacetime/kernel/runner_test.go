package kernel

import (
	"strings"
	"testing"
)

// Two executions advanced by one runner interleave rather than running to completion in
// turn. This is the whole structural claim: an arriving impulse gets its own stack and
// makes progress without waiting for the one already running.
func TestRunnerInterleaves(t *testing.T) {
	var log []string
	a := chain(&log, "a1", "a2", "a3")
	b := chain(&log, "b1", "b2", "b3")

	r := &Runner{}
	r.Admit(NewStack("", a, a, nil))
	r.Admit(NewStack("", b, b, nil))
	r.RunAll()

	joined := strings.Join(log, " ")
	if !strings.Contains(joined, "perturb:a1 perturb:b1") {
		t.Fatalf("executions did not interleave:\n%v", log)
	}
	// and both still ran correctly, post-order intact within each
	for _, want := range []string{"observe:a3<-nil", "observe:a1<-a2", "observe:b3<-nil", "observe:b1<-b2"} {
		if !strings.Contains(joined, want) {
			t.Errorf("missing %q in %v", want, log)
		}
	}
}

// A step is one action. Interleaving is only possible because no step runs a subtree.
func TestStepIsOneAction(t *testing.T) {
	var log []string
	s := NewStack("", nil, chain(&log, "x", "y"), nil)

	s.Step()
	if got := strings.Join(log, " "); got != "perturb:x" {
		t.Fatalf("first step did %q, want a single perturb", got)
	}
	s.Step()
	if got := strings.Join(log, " "); got != "perturb:x perturb:y" {
		t.Fatalf("second step did %q", got)
	}
}

// Admission ALWAYS succeeds. At capacity the runner evicts rather than refusing — the
// difference between a resource limit and an epistemic filter, and the reason this
// placement does not reproduce Toy 5.
func TestAdmitNeverRefuses(t *testing.T) {
	var log []string
	r := &Runner{Capacity: 2}

	first := NewStack("", nil, chain(&log, "a"), nil)
	r.Admit(first)
	r.Admit(NewStack("", nil, chain(&log, "b"), nil))

	evicted := r.Admit(NewStack("", nil, chain(&log, "c"), nil))
	if evicted == nil {
		t.Fatal("at capacity the runner refused instead of evicting")
	}
	if evicted != first {
		t.Error("evicted the wrong stack; default policy is oldest-first")
	}
	if r.Live() != 2 {
		t.Fatalf("live = %d, want 2", r.Live())
	}
}

// An evicted execution's transcript stays valid and says how far it got. A dropped stack
// is not a lost record.
func TestEvictedStackKeepsItsTranscript(t *testing.T) {
	var log []string
	r := &Runner{Capacity: 1}

	long := NewStack("", nil, chain(&log, "p", "q", "rr"), nil)
	r.Admit(long)
	r.Step() // descend into p
	r.Step() // descend into q

	evicted := r.Admit(NewStack("", nil, chain(&log, "other"), nil))
	if evicted != long {
		t.Fatal("expected the in-flight stack to be evicted")
	}
	tr := evicted.Transcript()
	if tr.Executed() != 3 {
		t.Fatalf("executed = %d, want 3 hops recorded before eviction", tr.Executed())
	}
	if evicted.Done() {
		t.Error("evicted stack reports Done; it was cut short, not finished")
	}
	if evicted.Depth() == 0 {
		t.Error("evicted stack lost its frames; it should say where it stopped")
	}
}

// The runner reads nothing off a stack but Done. Scheduling by trust, activation, or
// depth would put the controller back in charge of which executions progress — the
// sealing failure one level up. Round-robin is deliberate, not lazy.
func TestRoundRobinGivesEveryStackProgress(t *testing.T) {
	var log []string
	r := &Runner{}
	deep := NewStack("", nil, chain(&log, "d1", "d2", "d3", "d4", "d5"), nil)
	shallow := NewStack("", nil, chain(&log, "s1"), nil)
	r.Admit(deep)
	r.Admit(shallow)

	for i := 0; i < 4; i++ {
		r.Step()
	}
	// the shallow stack must have made progress despite being admitted second and
	// competing with a much deeper one
	if !strings.Contains(strings.Join(log, " "), "perturb:s1") {
		t.Fatalf("shallow stack starved: %v", log)
	}
}

// Finished executions are reaped, and a runner with nothing live does no work.
func TestRunnerDrains(t *testing.T) {
	var log []string
	r := &Runner{}
	r.Admit(NewStack("", nil, chain(&log, "a", "b"), nil))
	r.RunAll()

	if r.Live() != 0 {
		t.Fatalf("live = %d, want 0 after draining", r.Live())
	}
	if r.Step() {
		t.Error("empty runner reported work")
	}
}
