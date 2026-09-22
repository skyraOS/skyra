package dna

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// These are component checks, not an echo-loop proof. No concrete runtime
// fixture types are introduced under the order's no-new-types constraint.
func TestConstructionRejectsMissingDependencies(t *testing.T) {
	if ct, err := NewCTU("root", DriverRegistration{}, nil); err == nil || ct != nil {
		t.Fatal("accepted missing dependencies")
	}
	r := NewDriverRegistry()
	if err := r.Register("root", DriverRegistration{}); err == nil {
		t.Fatal("accepted nil driver")
	}
	if _, err := r.Lookup("root"); err == nil {
		t.Fatal("failed registration became visible")
	}
	var ct *ctu
	if err := ct.Submit(context.Background(), Move{}); err == nil {
		t.Fatal("accepted uninitialized CTU")
	}
	if err := ct.RegisterDriver("root", DriverRegistration{}); err == nil {
		t.Fatal("accepted nil CTU")
	}
}

func TestValidationReportsFailureAndReturnsCallbackError(t *testing.T) {
	cc := NewContinuationController(NewDriverRegistry(), nil)
	want := errors.New("report unavailable")
	calls := 0
	err := cc.Validate(context.Background(), Move{}, func(event OpcodeEvent) error {
		calls++
		failure, ok := event.(ValidationFailure)
		if !ok || failure.Err == nil {
			t.Fatalf("expected validation failure: %#v", event)
		}
		return want
	})
	if !errors.Is(err, want) || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := cc.Validate(ctx, Move{}, func(event OpcodeEvent) error {
		if !errors.Is(event.(ValidationFailure).Err, context.Canceled) {
			t.Fatal("lost submitted cancellation")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDispatchLookupFailureIsAnOutcome(t *testing.T) {
	cc := NewContinuationController(NewDriverRegistry(), nil)
	move := Move{Invocation: InvocationID{Value: "child"}, Target: Target{Address: "missing"}}
	calls := 0
	if err := cc.Dispatch(context.Background(), move, func(event OpcodeEvent) error {
		calls++
		result, ok := event.(DispatchResult)
		if !ok || result.Invocation != move.Invocation || result.Authorized || result.Err == nil {
			t.Fatalf("bad outcome: %#v", event)
		}
		return nil
	}); err != nil || calls != 1 {
		t.Fatalf("err=%v calls=%d", err, calls)
	}
	want := errors.New("cannot write outcome")
	if err := cc.Dispatch(context.Background(), move, func(OpcodeEvent) error { return want }); !errors.Is(err, want) {
		t.Fatal(err)
	}
}

func TestBlockedSchedulerDoesNotAdmitOrMutateWork(t *testing.T) {
	s := NewScheduler(NewContinuationController(NewDriverRegistry(), nil))
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.Write(Move{Invocation: InvocationID{Value: "child"}, Op: Call{}}); err == nil {
				t.Error("unfinished loop reported success")
			}
		}()
	}
	wg.Wait()
	if len(s.continuations) != 0 {
		t.Fatal("blocked admission mutated resident state")
	}
	if err := s.run(Move{}); err == nil {
		t.Fatal("run discarded missing context")
	}
}

func TestMoveSnapshotDetachesMutableValues(t *testing.T) {
	obs := Observation{}
	obs.Content.Bytes = []byte("echo")
	cap := Capability{}
	cap.Sig.Bytes = []byte("sig")
	parent := InvocationID{Value: "parent"}
	move := Move{ParentFd: &parent, Params: []Observation{obs}, Horizon: Horizon{Capabilities: []Capability{cap}}, Focus: []Capability{cap}, Op: &Ret{InReplyTo: InvocationID{Value: "child"}, Value: obs}}
	snapshot := cloneMove(move)
	parent.Value = "changed"
	move.Params[0].Content.Bytes[0] = 'X'
	move.Horizon.Capabilities[0].Sig.Bytes[0] = 'X'
	move.Op.(*Ret).InReplyTo.Value = "changed"
	if snapshot.ParentFd.Value != "parent" || string(snapshot.Params[0].Content.Bytes) != "echo" || string(snapshot.Horizon.Capabilities[0].Sig.Bytes) != "sig" || string(snapshot.Focus[0].Sig.Bytes) != "sig" || string(snapshot.Op.(*Ret).Value.Content.Bytes) != "echo" || snapshot.Op.(*Ret).InReplyTo.Value != "child" {
		t.Fatal("snapshot aliases producer state")
	}
}

func TestReturnRequiresDistinctIdentity(t *testing.T) {
	move := Move{ID: MoveID{Value: "operation"}, Invocation: InvocationID{Value: "return"}, Target: Target{Address: "root"}, Action: Action{Value: "echo"}, Op: Ret{InReplyTo: InvocationID{Value: "return"}}}
	if err := validateStructure(move); err == nil {
		t.Fatal("accepted reused invocation identity")
	}
	move.Op = Ret{InReplyTo: InvocationID{Value: "child"}}
	if err := validateStructure(move); err != nil {
		t.Fatal(err)
	}
	var ret *Ret
	move.Op = ret
	if err := validateStructure(move); err == nil {
		t.Fatal("accepted nil opcode")
	}
}
