package ctu

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"skyra-v1/src/skyra/skyra-dna-v14/controller"
	"skyra-v1/src/skyra/skyra-dna-v14/driver"
	"skyra-v1/src/skyra/skyra-dna-v14/internal/support"
	"skyra-v1/src/skyra/skyra-dna-v14/probe"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
	"skyra-v1/src/skyra/skyra-dna-v14/scheduler"
)

func observation(s string) protocol.Observation {
	o := protocol.Observation{}
	o.Content.Bytes = []byte(s)
	return o
}

func rootMove() protocol.Move {
	return protocol.Move{Root: protocol.PossibilityID{Value: "workflow"}, Target: protocol.Target{Address: "root"}, Op: protocol.Invoke{InvocationID: protocol.InvocationID{Value: "root-id"}, Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("hello")}}}
}

func fixture(t *testing.T, echo driver.Driver, authority controller.Authority, done func(protocol.RootResult) error) CTU {
	t.Helper()
	ct, err := NewCTU("root", driver.DriverRegistration{Driver: &probe.RootDriver{EchoTarget: protocol.Target{Address: "echo"}}, Input: probe.ProbeInput{}}, authority, done)
	if err != nil {
		t.Fatal(err)
	}
	if err := ct.RegisterDriver("echo", driver.DriverRegistration{Driver: echo, Input: probe.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	return ct
}

func allow() controller.Authority {
	return &probe.ProbeAuthority{CheckMove: func(context.Context, protocol.Move) error { return nil }}
}

func checkRoot(t *testing.T, ct CTU, result protocol.RootResult) {
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
	c, ok := ct.Snapshot(protocol.InvocationID{Value: "root-id"})
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
		child, _ := ct.Snapshot(id)
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
	ct := fixture(t, &probe.EchoDriver{}, allow(), func(r protocol.RootResult) error { results = append(results, r); return nil })
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
	ct := fixture(t, &probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}, allow(), func(r protocol.RootResult) error { results = append(results, r); return nil })
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
	ct := fixture(t, &probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}, allow(), func(protocol.RootResult) error { mu.Lock(); count++; mu.Unlock(); return nil })
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

type testDriver struct {
	ingress func(context.Context, protocol.Move, func(string, error) error) error
	egress  func(string) ([]protocol.Move, error)
}

func (d *testDriver) Ingress(c context.Context, m protocol.Move, f func(string, error) error) error {
	return d.ingress(c, m, f)
}

func (d *testDriver) Egress(s string) ([]protocol.Move, error) {
	if d.egress != nil {
		return d.egress(s)
	}
	return (&probe.RootDriver{}).Egress(s)
}

func TestFailures(t *testing.T) {
	sentinel := errors.New("host callback")
	for _, mode := range []string{"deny", "malformed", "transport", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			count := 0
			var pending []func() error
			echo := driver.Driver(&testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
				if mode == "transport" {
					return out("", errors.New("transport"))
				}
				return out("invalid", nil)
			}})
			a := allow()
			if mode == "deny" {
				a = &probe.ProbeAuthority{CheckMove: func(_ context.Context, m protocol.Move) error {
					if m.Target.Address == "echo" {
						return errors.New("denied")
					}
					return nil
				}}
			}
			if mode == "cancel" {
				echo = &probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}
			}
			ct := fixture(t, echo, a, func(r protocol.RootResult) error {
				count++
				if r.Err == nil {
					t.Error("missing workflow error")
				}
				return sentinel
			})
			err := ct.Submit(ctx, rootMove())
			if mode == "cancel" {
				if err != nil {
					t.Fatal(err)
				}
				cancel()
				err = pending[0]()
			}
			if !errors.Is(err, sentinel) {
				t.Fatalf("callback error lost: %v", err)
			}
			if count != 1 {
				t.Fatal(count)
			}
			c, _ := ct.Snapshot(protocol.InvocationID{Value: "root-id"})
			if c.State != protocol.Faulted {
				t.Fatal("root not faulted")
			}
		})
	}
}

func TestBoundReturnAndDuplicates(t *testing.T) {
	var count int
	evil := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
		return out("ignored", nil)
	}, egress: func(string) ([]protocol.Move, error) {
		return []protocol.Move{{Target: protocol.Target{Invocation: protocol.InvocationID{Value: "victim"}}, Op: protocol.Return{Observation: observation("hello")}}}, nil
	}}
	ct := fixture(t, evil, allow(), func(protocol.RootResult) error { count++; return nil })
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
	if err := s.Write(context.Background(), protocol.MoveBatch{Source: &id, Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	m.Op = protocol.Return{Observation: observation("conflict")}
	if err := s.Write(context.Background(), protocol.MoveBatch{Source: &id, Moves: []protocol.Move{m}}); err == nil {
		t.Fatal("conflict accepted")
	}
	if count != 1 {
		t.Fatal("duplicate completion")
	}
}

func TestMissingDependencies(t *testing.T) {
	if _, err := scheduler.NewScheduler(nil, func(protocol.RootResult) error { return nil }); err == nil {
		t.Fatal("nil controller")
	}
	if _, err := controller.NewContinuationController(driver.NewDriverRegistry(), nil); err == nil {
		t.Fatal("nil authority")
	}
	if err := driver.NewDriverRegistry().Register("echo", driver.DriverRegistration{Driver: &probe.EchoDriver{}}); err == nil {
		t.Fatal("missing validator")
	}
	var d *probe.EchoDriver
	if err := driver.NewDriverRegistry().Register("echo", driver.DriverRegistration{Driver: d, Input: probe.ProbeInput{}}); err == nil {
		t.Fatal("typed nil driver")
	}
	if _, err := NewCTU("root", driver.DriverRegistration{Driver: &probe.RootDriver{}, Input: probe.ProbeInput{}}, allow(), nil); err == nil {
		t.Fatal("nil callback")
	}
}

func TestNormalizeChildrenAndResumeAuthorization(t *testing.T) {
	grant := protocol.Capability{Target: protocol.Target{Address: "echo"}, Action: protocol.Action{Value: "echo"}, Sig: []byte("grant")}
	input := rootMove()
	inv := input.Op.(protocol.Invoke)
	inv.Horizon = protocol.Horizon{Capabilities: []protocol.Capability{grant}}
	input.Op = inv
	var ct CTU
	dispatches := 0
	rootCalls := 0
	results := 0
	a := &probe.ProbeAuthority{CheckMove: func(ctx context.Context, m protocol.Move) error {
		dispatches++
		if ctx.Value("key") != "lifetime" {
			t.Error("context lost")
		}
		if len(m.Op.(protocol.Invoke).Horizon.Capabilities) != 1 {
			t.Error("grant lost")
		}
		return nil
	}}
	root := &testDriver{ingress: func(ctx context.Context, m protocol.Move, out func(string, error) error) error {
		rootCalls++
		if rootCalls == 1 {
			return out("children", nil)
		}
		return (&probe.RootDriver{}).Ingress(ctx, m, out)
	}, egress: func(raw string) ([]protocol.Move, error) {
		if raw != "children" {
			return (&probe.RootDriver{}).Egress(raw)
		}
		return []protocol.Move{{Root: protocol.PossibilityID{Value: "forged"}, Target: protocol.Target{Address: "echo"}, Op: protocol.Invoke{InvocationID: protocol.InvocationID{Value: "forged"}, ReturnFd: protocol.InvocationID{Value: "forged"}, Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("hello")}, Register: protocol.Register{Order: []protocol.InvocationID{{Value: "forged"}}}, Horizon: protocol.Horizon{Capabilities: []protocol.Capability{{Sig: []byte("forged")}}}, Focus: []protocol.Capability{grant}}}}, nil
	}}
	var err error
	ct, err = NewCTU("root", driver.DriverRegistration{Driver: root, Input: probe.ProbeInput{}}, a, func(r protocol.RootResult) error {
		results++
		if r.Err != nil {
			t.Error(r.Err)
		}
		if _, ok := ct.Snapshot(r.Invocation); !ok {
			t.Error("callback cannot inspect root")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ct.RegisterDriver("echo", driver.DriverRegistration{Driver: &probe.EchoDriver{}, Input: probe.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	if err = ct.Submit(context.WithValue(context.Background(), "key", "lifetime"), input); err != nil {
		t.Fatal(err)
	}
	if dispatches != 3 || results != 1 || rootCalls != 2 {
		t.Fatalf("dispatches=%d results=%d root=%d", dispatches, results, rootCalls)
	}
	c, _ := ct.Snapshot(inv.InvocationID)
	id := c.Move.Op.(protocol.Invoke).Register.Order[0]
	child, _ := ct.Snapshot(id)
	ci := child.Move.Op.(protocol.Invoke)
	if id.Value == "forged" || ci.ReturnFd != inv.InvocationID || child.Move.Root != input.Root || !support.EmptyRegister(ci.Register) {
		t.Fatal("untrusted child metadata survived")
	}
	ci.Horizon.Capabilities[0].Sig[0] = 'X'
	again, _ := ct.Snapshot(id)
	if string(again.Move.Op.(protocol.Invoke).Horizon.Capabilities[0].Sig) != "grant" {
		t.Fatal("grant alias")
	}
}

func TestInvalidBatchIsAtomic(t *testing.T) {
	root := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
		return out("children", nil)
	}, egress: func(string) ([]protocol.Move, error) {
		return []protocol.Move{{Target: protocol.Target{Address: "echo"}, Op: protocol.Invoke{Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("hello")}}}, {Target: protocol.Target{Address: "missing"}, Op: protocol.Invoke{Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("hello")}}}}, nil
	}}
	count := 0
	ct, err := NewCTU("root", driver.DriverRegistration{Driver: root, Input: probe.ProbeInput{}}, allow(), func(r protocol.RootResult) error {
		count++
		if r.Err == nil {
			t.Error("missing failure")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ct.RegisterDriver("echo", driver.DriverRegistration{Driver: &probe.EchoDriver{}, Input: probe.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	if err = ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	c, _ := ct.Snapshot(protocol.InvocationID{Value: "root-id"})
	if count != 1 || len(c.Move.Op.(protocol.Invoke).Register.Children) != 0 {
		t.Fatal("partial batch admitted")
	}
}

func TestCancelledAdmissionAndDeniedResumption(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	count := 0
	a := &probe.ProbeAuthority{CheckMove: func(_ context.Context, m protocol.Move) error {
		if m.Target.Address == "root" && len(m.Op.(protocol.Invoke).Register.Children) > 0 {
			return errors.New("resume denied")
		}
		return nil
	}}
	ct := fixture(t, &probe.EchoDriver{}, a, func(r protocol.RootResult) error {
		count++
		if r.Err == nil {
			t.Error("missing denial")
		}
		return nil
	})
	if err := ct.Submit(ctx, rootMove()); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if _, ok := ct.Snapshot(protocol.InvocationID{Value: "root-id"}); ok {
		t.Fatal("cancelled root admitted")
	}
	if err := ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal(count)
	}
}

func TestRunSuppressesWorkAfterRootFault(t *testing.T) {
	calls := 0
	sentinel := errors.New("completion delivery failed")
	echo := &testDriver{ingress: func(context.Context, protocol.Move, func(string, error) error) error {
		calls++
		return errors.New("handoff failed")
	}}
	ct := fixture(t, echo, allow(), func(protocol.RootResult) error { return sentinel })
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

func TestCallbackWriteErrorIsNotReportedAsHandoffFailure(t *testing.T) {
	sentinel := errors.New("write failed")
	for _, swallow := range []bool{false, true} {
		endpoint := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
			raw, err := json.Marshal([]probe.ProbeOutput{{Kind: "return", Value: func() *protocol.Observation { o := observation("hello"); return &o }()}})
			if err != nil {
				return err
			}
			err = out(string(raw), nil)
			if swallow {
				return nil
			}
			return err
		}}
		registry := driver.NewDriverRegistry()
		if err := registry.Register("root", driver.DriverRegistration{Driver: endpoint, Input: probe.ProbeInput{}}); err != nil {
			t.Fatal(err)
		}
		cc, err := controller.NewContinuationController(registry, allow())
		if err != nil {
			t.Fatal(err)
		}
		var events []protocol.OpcodeEvent
		err = cc.Dispatch(context.Background(), rootMove(), func(e protocol.OpcodeEvent) error { events = append(events, e); return sentinel })
		if !errors.Is(err, sentinel) {
			t.Fatalf("callback error lost: %v", err)
		}
		if len(events) != 1 {
			t.Fatalf("callback failure replayed: %+v", events)
		}
		if _, ok := events[0].(protocol.MoveBatch); !ok {
			t.Fatalf("callback error reclassified: %T", events[0])
		}
	}
}
