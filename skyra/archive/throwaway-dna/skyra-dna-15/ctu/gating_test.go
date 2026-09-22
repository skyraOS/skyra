package ctu

import (
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"testing"
"context"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/probe"
"errors"
"skyra-v1/src/skyra/skyra-dna-15/controller"
"reflect"
)

func TestImmediateBufferAndFailure(t *testing.T) {
	count := 0
	var ct CTU
	root := &testDriver{ingress: func(_ context.Context, m protocol.Move, out func(string, error) error) error {
		if err := out("children", nil); err != nil {
			return err
		}
		c, _ := ct.Snapshot(m.Op.(protocol.Invoke).InvocationID)
		if len(c.Handoff.PendingOutput) != 2 || !support.EmptyRegister(c.Move.Op.(protocol.Invoke).Register) {
			t.Error("callback did not buffer before acknowledgement")
		}
		return errors.New("placement failed after output")
	}, egress: func(string) ([]protocol.Move, error) {
		return []protocol.Move{{Target: protocol.Target{Address: "echo"}, Op: protocol.Invoke{Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("a")}}}, {Target: protocol.Target{Address: "echo"}, Op: protocol.Invoke{Action: protocol.Action{Value: "echo"}, Params: []protocol.Observation{observation("b")}}}}, nil
	}}
	ct = fixtureRoot(t, root, &probe.EchoDriver{}, func(r protocol.RootResult) error {
		count++
		if r.Err == nil {
			t.Error("missing failure")
		}
		return nil
	})
	if err := ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	c, _ := ct.Snapshot(rootMove().Op.(protocol.Invoke).InvocationID)
	if count != 1 || len(c.Move.Op.(protocol.Invoke).Register.Children) != 0 {
		t.Fatal("failed placement admitted children")
	}
}

func TestControllerCallbackErrorsAndBinding(t *testing.T) {
	sentinel := errors.New("writer error")
	for _, swallow := range []bool{false, true} {
		endpoint := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
			err := out("ignored", nil)
			if swallow {
				return nil
			}
			return err
		}, egress: func(string) ([]protocol.Move, error) {
			return []protocol.Move{{Target: protocol.Target{Invocation: protocol.InvocationID{Value: "forged"}}, Op: protocol.Return{Observation: observation("x")}}}, nil
		}}
		r := driver.NewDriverRegistry()
		if err := r.Register("root", driver.DriverRegistration{Driver: endpoint, Input: probe.ProbeInput{}}); err != nil {
			t.Fatal(err)
		}
		cc, _ := controller.NewContinuationController(r.Resolver())
		var events []protocol.OpcodeEvent
		err := cc.Dispatch(context.Background(), rootMove(), 7, func(e protocol.OpcodeEvent) error { events = append(events, e); return sentinel })
		if !errors.Is(err, sentinel) || len(events) != 1 {
			t.Fatal("callback error replayed or lost", err, events)
		}
		batch, ok := events[0].(protocol.MoveBatch)
		if !ok || batch.Sequence != 7 || *batch.Source != rootMove().Op.(protocol.Invoke).InvocationID || batch.Moves[0].Target.Invocation != *batch.Source {
			t.Fatal("unbound output")
		}
	}
}

func TestMalformedOutputAndValidatorAtomicity(t *testing.T) {
	for _, mode := range []string{"parse", "transport", "batch"} {
		t.Run(mode, func(t *testing.T) {
			count := 0
			root := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error {
				if mode == "transport" {
					return out("", errors.New("transport"))
				}
				return out("invalid", nil)
			}}
			if mode == "batch" {
				root.egress = func(string) ([]protocol.Move, error) {
					a := rootMove()
					b := support.CloneMove(a)
					b.Target.Address = "missing"
					return []protocol.Move{a, b}, nil
				}
			}
			ct := fixtureRoot(t, root, &probe.EchoDriver{}, func(r protocol.RootResult) error {
				count++
				if r.Err == nil {
					t.Error("no fault")
				}
				return nil
			})
			if err := ct.Submit(context.Background(), rootMove()); err != nil {
				t.Fatal(err)
			}
			c, _ := ct.Snapshot(rootMove().Op.(protocol.Invoke).InvocationID)
			if count != 1 || !support.EmptyRegister(c.Move.Op.(protocol.Invoke).Register) {
				t.Fatal("partial installation")
			}
		})
	}
}

func TestRejectedAdmissionDoesNotFaultExisting(t *testing.T) {
	var pending []func() error
	ct := fixture(t, &probe.EchoDriver{Release: func(f func() error) error { pending = append(pending, f); return nil }}, func(protocol.RootResult) error { return nil })
	if err := ct.Submit(context.Background(), rootMove()); err != nil {
		t.Fatal(err)
	}
	id := rootMove().Op.(protocol.Invoke).InvocationID
	before, _ := ct.Snapshot(id)
	bad := rootMove()
	bad.Target.Address = "missing"
	if err := ct.Submit(context.Background(), bad); err == nil {
		t.Fatal("invalid submission accepted")
	}
	after, _ := ct.Snapshot(id)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("rejection faulted resident work")
	}
}

func TestDelayedDriverFaultUsesBoundRuntimeEvent(t *testing.T) {
	var output func(string, error) error
	r := driver.NewDriverRegistry()
	endpoint := &testDriver{ingress: func(_ context.Context, _ protocol.Move, out func(string, error) error) error { output = out; return nil }}
	if err := r.Register("root", driver.DriverRegistration{Driver: endpoint, Input: probe.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	cc, _ := controller.NewContinuationController(r.Resolver())
	var events []protocol.OpcodeEvent
	if err := cc.Dispatch(context.Background(), rootMove(), 3, func(e protocol.OpcodeEvent) error { events = append(events, e); return nil }); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("delayed transport")
	if err := output("", sentinel); err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatal("missing delayed feedback")
	}
	e, ok := events[1].(protocol.ExecutionFailure)
	if !ok || e.Sequence != 3 || e.Source != rootMove().Op.(protocol.Invoke).InvocationID || !errors.Is(e.Err, sentinel) {
		t.Fatal("unbound delayed fault")
	}
}

func TestDeniedChildBatchHasNoEffects(t *testing.T) {
	calls := 0
	root := &probe.RootDriver{EchoTarget: protocol.Target{Address: "echo"}}
	echo := &testDriver{ingress: func(context.Context, protocol.Move, func(string, error) error) error { calls++; return nil }}
	ct := fixtureRoot(t, root, echo, func(r protocol.RootResult) error {
		if r.Err == nil {
			t.Error("denied batch notified success")
		}
		return nil
	})
	m := rootMove()
	inv := m.Op.(protocol.Invoke)
	inv.Horizon.Capabilities = inv.Horizon.Capabilities[:1]
	m.Op = inv
	if err := ct.Submit(context.Background(), m); err == nil {
		t.Fatal("child authorization bypassed", err)
	}
	c, _ := ct.Snapshot(inv.InvocationID)
	if calls != 0 || len(c.Move.Op.(protocol.Invoke).Register.Children) != 0 || len(c.Handoff.PendingOutput) != 0 || !c.Handoff.OutputConsumed || c.State != protocol.Faulted {
		t.Fatal("denied batch had effects")
	}
}

