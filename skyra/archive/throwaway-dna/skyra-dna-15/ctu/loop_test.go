package ctu

import (
"skyra-v1/src/skyra/skyra-dna-15/probe"
"sync"
"errors"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/kernel"
"testing"
"context"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/scheduler"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"encoding/json"
)

// snapshotter is what checkRoot actually needs — a CTU satisfies it, and so does a
// bare *scheduler.Scheduler, letting scheduler-level tests reuse the same root check.
type snapshotter interface {
	Snapshot(protocol.InvocationID) (scheduler.Continuation, bool)
}

func checkRoot(t *testing.T, snap snapshotter, result protocol.RootResult) {
	t.Helper()
	if result.Err != nil || result.Observation == nil {
		t.Fatalf("bad result: %+v", result)
	}
	var values []protocol.Observation
	if err := json.Unmarshal(result.Observation.Content.Bytes, &values); err != nil {
		t.Fatal(err)
	}
	if len(values) != 2 || string(values[0].Content.Bytes) != "hello" || string(values[1].Content.Bytes) != "hello" {
		t.Fatalf("bad values: %+v", values)
	}
	c, ok := snap.Snapshot(protocol.InvocationID{Value: "root-id"})
	if !ok || !c.Resumed || c.Result == nil {
		t.Fatalf("bad root: %+v", c)
	}
	r := c.Move.Op.(protocol.Invoke).Register
	if len(r.Order) != 2 || r.Order[0] == r.Order[1] {
		t.Fatal("child identity/order lost")
	}
	for _, id := range r.Order {
		e := r.Children[id]
		if e.Observation == nil || e.Observation.InvocationID != id {
			t.Fatal("missing child observation")
		}
		child, _ := snap.Snapshot(id)
		if child.Result == nil || child.Move.Op.(protocol.Invoke).ReturnFd != result.Invocation {
			t.Fatal("child not linked")
		}
		if !support.EmptyRegister(e.Invocation.Move.Op.(protocol.Invoke).Register) {
			t.Fatal("embedded live frame")
		}
	}
}

func TestImmediateLoop(t *testing.T) {
	var results []protocol.RootResult
	ct := fixture(t, &probe.EchoDriver{}, func(r protocol.RootResult) error { results = append(results, r); return nil })
	if err := ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("results %d", len(results))
	}
	checkRoot(t, ct, results[0])
	if err := ct.Submit(context.Background(), rootMove()); err == nil {
		t.Fatal("accepted reused ID")
	}
}

func TestReverseAndSnapshotIsolation(t *testing.T) {
	var pending []func() error
	var results []protocol.RootResult
	ct := fixture(t, &probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}, func(r protocol.RootResult) error { results = append(results, r); return nil })
	ctx := context.WithValue(context.Background(), struct{}{}, "saved")
	if err := ct.Submit(ctx, rootMove()); err != nil {
		t.Fatal(err)
	}
	c, _ := ct.Snapshot(protocol.InvocationID{Value: "root-id"})
	inv := c.Move.Op.(protocol.Invoke)
	ids := append([]protocol.InvocationID(nil), inv.Register.Order...)
	inv.Params[0].Content.Bytes[0] = 'X'
	inv.Register.Order[0] = protocol.InvocationID{}
	delete(inv.Register.Children, ids[0])
	if len(pending) != 2 {
		t.Fatal("not atomic fanout")
	}
	if err := pending[1](); err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Fatal("premature result")
	}
	c, _ = ct.Snapshot(protocol.InvocationID{Value: "root-id"})
	if c.Context != ctx || c.Move.Op.(protocol.Invoke).Register.Children[ids[0]].Observation != nil {
		t.Fatal("wrong child or context")
	}
	if err := pending[0](); err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatal("wrong result count")
	}
	checkRoot(t, ct, results[0])
}

func TestConcurrentChildren(t *testing.T) {
	var pending []func() error
	var mu sync.Mutex
	count := 0
	ct := fixture(t, &probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}, func(protocol.RootResult) error { mu.Lock(); count++; mu.Unlock(); return nil })
	if err := ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	for _, f := range pending {
		wg.Add(1)
		go func(f func() error) {
			defer wg.Done()
			if err := f(); err != nil {
				t.Error(err)
			}
		}(f)
	}
	wg.Wait()
	if count != 1 {
		t.Fatal(count)
	}
}

func TestBoundReturnAndDuplicates(t *testing.T) {
	var count int
	evil := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
		return out("ignored", nil)
	}, egress: func(string) ([]protocol.Move, error) {
		return []protocol.Move{{Target: protocol.Target{Invocation: protocol.InvocationID{Value: "victim"}}, Op: protocol.Return{Observation: observation("hello")}}}, nil
	}}
	ct := fixture(t, evil, func(protocol.RootResult) error { count++; return nil })
	if err := ct.Submit(context.Background(), protocol.Move{Target: protocol.Target{Invocation: protocol.InvocationID{Value: "root-id"}}, Op: protocol.Return{Observation: observation("forged")}}); err == nil {
		t.Fatal("external return accepted")
	}
	if err := ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal(count)
	}
	if _, ok := ct.Snapshot(protocol.InvocationID{Value: "victim"}); ok {
		t.Fatal("raw identity trusted")
	}
	s := ct.(*ctu).scheduler
	c, _ := ct.Snapshot(protocol.InvocationID{Value: "root-id"})
	id := c.Move.Op.(protocol.Invoke).Register.Order[0]
	child, _ := ct.Snapshot(id)
	m := child.Move
	m.Target.Invocation = id
	m.Op = protocol.Return{Observation: *child.Result}
	if err := s.Transition(context.Background(), protocol.MoveBatch{Source: &id, Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	m.Op = protocol.Return{Observation: observation("conflict")}
	if err := s.Transition(context.Background(), protocol.MoveBatch{Source: &id, Moves: []protocol.Move{m}}); err == nil {
		t.Fatal("conflict accepted")
	}
	if count != 1 {
		t.Fatal("duplicate completion")
	}
}

func TestRunSuppressesWorkAfterRootFault(t *testing.T) {
	calls := 0
	sentinel := errors.New("completion delivery failed")
	echo := &testDriver{ingress: func(context.Context, protocol.Move, func(string, error) error) error {
		calls++
		return errors.New("handoff failed")
	}}
	ct := fixture(t, echo, func(protocol.RootResult) error { return sentinel })
	if err := ct.Submit(context.Background(), rootMove()); !errors.Is(err, sentinel) {
		t.Fatalf("callback error lost: %v", err)
	}
	if calls != 1 {
		t.Fatalf("dispatched %d siblings after first fault", calls)
	}
	c, _ := ct.Snapshot(protocol.InvocationID{Value: "root-id"})
	if len(c.Move.Op.(protocol.Invoke).Register.Children) != 2 {
		t.Fatal("batch was not admitted before dispatch")
	}
}

func fixture(t *testing.T, echo driver.Driver, done func(protocol.RootResult) error) CTU {
	t.Helper()
	return fixtureRoot(t, &probe.RootDriver{EchoTarget: protocol.Target{Address: "echo"}}, echo, done)
}

func fixtureRoot(t *testing.T, root driver.Driver, echo driver.Driver, done func(protocol.RootResult) error) CTU {
	t.Helper()
	endpoint := kernel.Endpoint{Address: "root", Registration: driver.DriverRegistration{Driver: root, Input: probe.ProbeInput{}}}
	kernel := &kernel.FixtureKernel{Root: endpoint, Local: []kernel.Endpoint{{Address: "echo", Registration: driver.DriverRegistration{Driver: echo, Input: probe.ProbeInput{}}}}}
	ct, err := NewCTU(BootConfig{KernelAddress: "kernel", Kernel: kernel, KernelInput: probe.ProbeInput{}, Root: endpoint}, done)
	if err != nil {
		t.Fatal(err)
	}
	return ct
}

func TestDistinctOrderedResultsAndSavedContexts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var pending []func() error
	rootCalls := 0
	childCalls := 0
	results := 0
	root := &testDriver{ingress: func(c context.Context, m protocol.Move, out func(string, error) error) error {
		rootCalls++
		if c != ctx {
			t.Error("root context replaced")
		}
		if rootCalls == 1 {
			return out("children", nil)
		}
		if len(m.Op.(protocol.Invoke).Register.Order) != 2 {
			t.Error("missing resume register")
		}
		return (&probe.RootDriver{}).Ingress(c, m, out)
	}, egress: func(raw string) ([]protocol.Move, error) {
		if raw != "children" {
			return (&probe.RootDriver{}).Egress(raw)
		}
		return []protocol.Move{{Target: protocol.Target{Address: "echo"}, Op: protocol.Invoke{Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("first")}}}, {Target: protocol.Target{Address: "echo"}, Op: protocol.Invoke{Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("second")}}}}, nil
	}}
	echo := &testDriver{ingress: func(c context.Context, m protocol.Move, out func(string, error) error) error {
		childCalls++
		if c != ctx {
			t.Error("child context replaced")
		}
		return (&probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}).Ingress(c, m, out)
	}}
	ct := fixtureRoot(t, root, echo, func(r protocol.RootResult) error {
		results++
		var values []protocol.Observation
		if r.Err != nil || r.Observation == nil {
			t.Fatal("failed result", r.Err)
		}
		if err := json.Unmarshal(r.Observation.Content.Bytes, &values); err != nil {
			t.Fatal(err)
		}
		if len(values) != 2 || string(values[0].Content.Bytes) != "first" || string(values[1].Content.Bytes) != "second" {
			t.Fatal("completion order replaced declaration order")
		}
		return nil
	})
	if err := ct.Submit(ctx, rootMove()); err != nil {
		t.Fatal(err)
	}
	if len(pending) != 2 {
		t.Fatal("fanout incomplete")
	}
	if err := pending[1](); err != nil {
		t.Fatal(err)
	}
	if results != 0 || rootCalls != 1 {
		t.Fatal("premature resumption")
	}
	if err := pending[0](); err != nil {
		t.Fatal(err)
	}
	if results != 1 || rootCalls != 2 || childCalls != 2 {
		t.Fatal("execution counts", results, rootCalls, childCalls)
	}
}

func TestValidatorAndHostCallbacksOutsideSchedulerLock(t *testing.T) {
	var ct CTU
	validations := 0
	results := 0
	endpoint := kernel.Endpoint{Address: "root", Registration: driver.DriverRegistration{Driver: &probe.EchoDriver{}, Input: &testValidator{validate: func(m protocol.Move) error {
		validations++
		if _, ok := ct.Snapshot(m.Op.(protocol.Invoke).InvocationID); ok {
			t.Error("admitted before validation")
		}
		m.Op.(protocol.Invoke).Params[0].Content.Bytes[0] = 'X'
		return nil
	}}}}
	kernel := &kernel.FixtureKernel{Root: endpoint}
	var err error
	ct, err = NewCTU(BootConfig{KernelAddress: "kernel", Kernel: kernel, KernelInput: probe.ProbeInput{}, Root: endpoint}, func(r protocol.RootResult) error {
		results++
		if c, ok := ct.Snapshot(r.Invocation); !ok || c.Result == nil {
			t.Error("notification before result")
		}
		if string(r.Observation.Content.Bytes) != "hello" {
			t.Error("validator mutated resident input")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	if validations != 1 || results != 1 {
		t.Fatal("missing callback")
	}
}

