package engine

import (
	"reflect"
	"strings"
	"testing"

	c "skyra-v1/src/skyra/skyra-dna-v12/causal"
)

// Component evidence only: installation and execution methods remain unimplemented.
func TestLandPreservesOrderAndOwnsPublishedBytes(t *testing.T) {
	parent, first, second := MoveID{"parent"}, MoveID{"first"}, MoveID{"second"}
	frame := &Register{Children: []ChildObservation{{Fd: first}, {Fd: second}}}
	memory := &Memory{Fds: map[MoveID]*Fd{parent: {Move: Move{Fd: parent}, Register: frame}}}
	value := c.Observation{Content: c.Content{Bytes: []byte("second")}, Segment: c.Segment{By: c.Name{Value: "document"}}}
	if memory.Land(parent, second, value) != c.Success {
		t.Fatal("second child did not land")
	}
	if frame.Children[0].Value != nil || frame.Children[1].Fd != second || !reflect.DeepEqual(*frame.Children[1].Value, value) {
		t.Fatal("landing changed order or the wrong entry")
	}
	stored := frame.Children[1].Value
	if memory.Land(parent, second, value) != c.Success || frame.Children[1].Value != stored {
		t.Fatal("identical repeat must retain the published observation")
	}
	value.Content.Bytes[0] = 'X'
	if string(stored.Content.Bytes) != "second" || memory.Land(parent, second, value) != c.Failure {
		t.Fatal("producer mutation or conflicting repeat changed publication")
	}
	value.Content.Bytes = []byte("second")
	value.Segment.By.Value = "different"
	if memory.Land(parent, second, value) != c.Failure {
		t.Fatal("segment conflict accepted")
	}
	if memory.Land(parent, first, c.Observation{}) != c.Success || frame.Children[0].Value == nil || frame.Children[0].Fd != first {
		t.Fatal("zero observation must count as an arrival in its declared position")
	}
}

func TestLandRejectsMissingAndAmbiguousDestinations(t *testing.T) {
	parent, child := MoveID{"parent"}, MoveID{"child"}
	frame := &Register{Children: []ChildObservation{{Fd: child}, {Fd: child}}}
	memory := &Memory{Fds: map[MoveID]*Fd{parent: {Move: Move{Fd: parent}, Register: frame}}}
	for _, destination := range []MoveID{parent, {Value: "unknown"}} {
		if memory.Land(destination, child, c.Observation{}) != c.Failure {
			t.Fatal("invalid destination accepted")
		}
	}
	if memory.Land(parent, MoveID{"undeclared"}, c.Observation{}) != c.Failure {
		t.Fatal("undeclared child accepted")
	}
	if frame.Children[0].Value != nil || frame.Children[1].Value != nil || len(memory.Fds) != 1 {
		t.Fatal("failure mutated memory")
	}
	for _, empty := range []*Memory{nil, {}, {Fds: map[MoveID]*Fd{parent: nil}}, {Fds: map[MoveID]*Fd{parent: {Move: Move{Fd: parent}}}}} {
		if empty.Land(parent, child, c.Observation{}) != c.Failure {
			t.Fatal("absent frame accepted")
		}
	}
}

func TestSchedulerClaimAndCompletion(t *testing.T) {
	first, second, unknown := MoveID{"first"}, MoveID{"second"}, MoveID{"unknown"}
	// Seed Ready as a unit-test precondition, not a runtime admission workaround.
	scheduler := &Scheduler{states: map[MoveID]State{first: Ready, second: Ready}}
	batch := []MoveID{second, unknown, second, first}
	claimed := scheduler.Claim(batch)
	if !reflect.DeepEqual(claimed, []MoveID{second, first}) || scheduler.states[first] != Claimed || scheduler.states[second] != Claimed {
		t.Fatal("claim must select each ready fd once and record it before returning")
	}
	claimed[0] = unknown
	if batch[0] != second || len(scheduler.Claim(batch)) != 0 {
		t.Fatal("claim aliased input or reclaimed outstanding work")
	}
	scheduler.Apply(Landed{Fd: first})
	scheduler.Apply(Landed{Fd: first})
	if scheduler.states[first] != Completed || scheduler.states[second] != Claimed || len(scheduler.states) != 2 {
		t.Fatal("completion must be retained across duplicate completion")
	}
	if len((&Scheduler{}).Claim(batch)) != 0 {
		t.Fatal("zero scheduler admitted unknown work")
	}
}

func TestSchedulerRejectsUnclaimedEvents(t *testing.T) {
	fd := MoveID{"move"}
	for _, state := range []State{0, Ready} {
		for _, event := range []Event{Landed{Fd: fd}, Faulted{Move: fd}, nil} {
			scheduler := &Scheduler{states: map[MoveID]State{fd: state}}
			expectProbePanic(t, func() { scheduler.Apply(event) })
			if scheduler.states[fd] != state {
				t.Fatal("rejected event changed scheduling state")
			}
		}
	}
}

// These are executable boundary demonstrations, not successful loop evidence.
func TestLocalLoopBlockedAtAdmissionResolutionAndDelivery(t *testing.T) {
	fd := MoveID{"root"}
	runtime := &Runtime{
		Memory: &Memory{Fds: map[MoveID]*Fd{fd: {Move: Move{Fd: fd}, Register: &Register{}}}},
		Kernel: &Kernel{}, CTU: &CTU{},
	}
	runtime.Boot()
	expectProbePanic(t, func() { runtime.Kernel.Sched.Admit([]MoveID{fd}) })
	expectProbePanic(t, func() { runtime.Memory.Resolve(fd) })
	expectProbePanic(t, func() { runtime.CTU.Fire([]*Fd{runtime.Memory.Fds[fd]}) })
	expectProbePanic(t, func() {
		(Ret{}).Fire(&ExecCtx{Host: runtime.CTU, Move: &Move{Fd: fd}, Frame: runtime.Memory.Fds[fd].Register})
	})
	if len(runtime.Kernel.Sched.states) != 0 {
		t.Fatal("blocked execution must not claim progress")
	}
}

func expectProbePanic(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		t.Helper()
		message, ok := recover().(string)
		if !ok || (!strings.HasPrefix(message, "PROPOSAL:") && !strings.HasPrefix(message, "engine:")) {
			t.Fatalf("expected explicit engine failure, got %q", message)
		}
	}()
	run()
}
