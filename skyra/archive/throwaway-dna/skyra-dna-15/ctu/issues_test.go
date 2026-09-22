package ctu

import (
"testing"
"context"
"errors"
"skyra-v1/src/skyra/skyra-dna-15/probe"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/controller"
"skyra-v1/src/skyra/skyra-dna-15/scheduler"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
)

func TestBufferedChildrenRepeatedInEnvelopeCompleteLoop(t *testing.T) {
	rootCalls, childCalls, notifications := 0, 0, 0
	rootDriver := &probe.RootDriver{EchoTarget: protocol.Target{Address: "echo"}}
	root := &testDriver{ingress: func(ctx context.Context, m protocol.Move, out func(string, error) error) error {
		rootCalls++
		return rootDriver.Ingress(ctx, m, out)
	}}
	echo := &testDriver{ingress: func(ctx context.Context, m protocol.Move, out func(string, error) error) error {
		childCalls++
		return (&probe.EchoDriver{}).Ingress(ctx, m, out)
	}}
	// Build the registry and real controller directly so the scheduler is constructed
	// with a wrapping controller from the start — no post-hoc write of the scheduler's
	// now-private controller field across the package boundary.
	r := driver.NewDriverRegistry()
	if err := r.Register("root", driver.DriverRegistration{Driver: root, Input: probe.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	if err := r.Register("echo", driver.DriverRegistration{Driver: echo, Input: probe.ProbeInput{}}); err != nil {
		t.Fatal(err)
	}
	cc, err := controller.NewContinuationController(r.Resolver())
	if err != nil {
		t.Fatal(err)
	}

	var s *scheduler.Scheduler
	// Wrap the real controller: intercept the root's initial dispatch, buffer its child
	// batch, then replay it inside an Envelope alongside the placement result.
	wrapping := &testController{
		validate: cc.Validate,
		dispatch: func(ctx context.Context, m protocol.Move, seq uint64, write func(protocol.OpcodeEvent) error) error {
			if m.Target.Address != "root" || seq != 0 {
				return cc.Dispatch(ctx, m, seq, write)
			}
			var buffered *protocol.MoveBatch
			return cc.Dispatch(ctx, m, seq, func(e protocol.OpcodeEvent) error {
				switch event := e.(type) {
				case protocol.MoveBatch:
					batch := event
					batch.Moves = support.CloneMoves(event.Moves)
					buffered = &batch
					if err := write(event); err != nil {
						return err
					}
					source, _ := s.Snapshot(m.Op.(protocol.Invoke).InvocationID)
					if len(source.Handoff.PendingOutput) != 2 || len(source.Move.Op.(protocol.Invoke).Register.Children) != 0 {
						t.Fatal("batch not buffered")
					}
					return nil
				case protocol.DispatchResult:
					if buffered == nil {
						t.Fatal("placement preceded output")
					}
					return write(protocol.Envelope{Runtime: event, Continuation: buffered})
				default:
					return write(e)
				}
			})
		},
	}
	s, err = scheduler.NewScheduler(wrapping, func(res protocol.RootResult) error {
		notifications++
		checkRoot(t, s, res)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	// Replicate ctu.Submit's admission: validate the root, then drive the scheduler.
	m := rootMove()
	if err := cc.Validate(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	if err := s.Transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	if notifications != 1 || rootCalls != 2 || childCalls != 2 {
		t.Fatalf("notifications=%d root=%d children=%d", notifications, rootCalls, childCalls)
	}
}

func TestCancelledBufferedReturnFaultsBeforeEffects(t *testing.T) {
	for _, mode := range []string{"root", "child"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			var ct CTU
			var source protocol.InvocationID
			notifications, endpointCalls := 0, 0
			hostErr := errors.New("notification failed")
			endpoint := &testDriver{ingress: func(c context.Context, m protocol.Move, out func(string, error) error) error {
				endpointCalls++
				source = m.Op.(protocol.Invoke).InvocationID
				if err := (&probe.EchoDriver{}).Ingress(c, m, out); err != nil {
					return err
				}
				snap, _ := ct.Snapshot(source)
				if len(snap.Handoff.PendingOutput) != 1 || snap.Result != nil {
					t.Fatal("Return not buffered")
				}
				cancel()
				return nil
			}}
			done := func(r protocol.RootResult) error {
				notifications++
				if !errors.Is(r.Err, context.Canceled) || r.Observation != nil {
					t.Errorf("wrong cancellation notification: %+v", r)
				}
				return hostErr
			}
			if mode == "root" {
				ct = fixtureRoot(t, endpoint, &probe.EchoDriver{}, done)
			} else {
				ct = fixture(t, endpoint, done)
			}
			err := ct.Submit(ctx, rootMove())
			if !errors.Is(err, context.Canceled) || !errors.Is(err, hostErr) {
				t.Fatal("cancellation or callback error lost", err)
			}
			snap, _ := ct.Snapshot(source)
			if snap.Result != nil || snap.State != protocol.Faulted || !snap.Handoff.OutputConsumed || len(snap.Handoff.PendingOutput) != 0 {
				t.Fatal("cancelled Return had effects or was not consumed")
			}
			rootID := rootMove().Op.(protocol.Invoke).InvocationID
			root, _ := ct.Snapshot(rootID)
			if root.Result != nil || root.State != protocol.Faulted {
				t.Fatal("root not faulted")
			}
			if mode == "child" && root.Move.Op.(protocol.Invoke).Register.Children[source].Observation != nil {
				t.Fatal("cancelled child observation applied")
			}
			if err := ct.(*ctu).scheduler.Transition(context.Background(), protocol.DispatchResult{Invocation: source}); err != nil {
				t.Fatal(err)
			}
			if notifications != 1 || endpointCalls != 1 {
				t.Fatal("repeated notification or sibling dispatched after cancellation")
			}
		})
	}
}

