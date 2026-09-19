package scheduler

import (
"context"
"errors"
"time"
"reflect"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"testing"
)

func TestPlacementOrders(t *testing.T) {
	for _, order := range []string{"output-first", "placement-first", "combined", "failure"} {
		t.Run(order, func(t *testing.T) {
			var callback func(protocol.OpcodeEvent) error
			notifications := 0
			s, err := NewScheduler(&testController{dispatch: func(_ context.Context, m protocol.Move, seq uint64, w func(protocol.OpcodeEvent) error) error {
				if seq != 0 || !m.Stamp.Allowed || m.Stamp.CheckedAt == 0 {
					t.Error("unstamped or wrong handoff")
				}
				callback = w
				return nil
			}}, func(r protocol.RootResult) error {
				notifications++
				if (order == "failure") != (r.Err != nil) {
					t.Error("wrong outcome")
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			m := rootMove()
			id := m.Op.(protocol.Invoke).InvocationID
			if err = s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
				t.Fatal(err)
			}
			ret := protocol.Move{Root: m.Root, Target: m.Target, Op: protocol.Return{Observation: observation("result")}}
			ret.Target.Invocation = id
			batch := protocol.MoveBatch{Source: &id, Moves: []protocol.Move{ret}}
			ack := protocol.DispatchResult{Invocation: id}
			switch order {
			case "output-first", "failure":
				if err = callback(batch); err != nil {
					t.Fatal(err)
				}
				c, _ := s.Snapshot(id)
				if c.Result != nil || len(c.Handoff.PendingOutput) != 1 || !support.EmptyRegister(c.Move.Op.(protocol.Invoke).Register) {
					t.Fatal("early output had effects")
				}
				c.Handoff.PendingOutput[0].Op = protocol.Return{Observation: observation("mutated")}
				if order == "failure" {
					ack.Err = errors.New("placement failed")
				}
				err = callback(ack)
			case "placement-first":
				if err = callback(ack); err != nil {
					t.Fatal(err)
				}
				err = callback(batch)
			case "combined":
				err = callback(protocol.Envelope{Runtime: ack, Continuation: &batch})
			}
			if err != nil {
				t.Fatal(err)
			}
			c, _ := s.Snapshot(id)
			if notifications != 1 || len(c.Handoff.PendingOutput) != 0 || !c.Handoff.OutputConsumed {
				t.Fatal("gate not consumed exactly once")
			}
			if order != "failure" && string(c.Result.Content.Bytes) != "result" {
				t.Fatal("snapshot alias")
			}
			if err = callback(ack); err != nil {
				t.Fatal(err)
			}
			if notifications != 1 {
				t.Fatal("repeated notification")
			}
		})
	}
}

func TestStalePlacementCannotConfirmResumption(t *testing.T) {
	var resumed func(protocol.OpcodeEvent) error
	cc := &testController{dispatch: func(_ context.Context, m protocol.Move, seq uint64, w func(protocol.OpcodeEvent) error) error {
		if seq == 1 {
			resumed = w
			return nil
		}
		return w(protocol.DispatchResult{Invocation: m.Op.(protocol.Invoke).InvocationID, Sequence: seq})
	}}
	s, _ := NewScheduler(cc, func(protocol.RootResult) error { return nil })
	ctx := context.Background()
	m := rootMove()
	id := m.Op.(protocol.Invoke).InvocationID
	if err := s.transition(ctx, protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	child := support.CloneMove(m)
	inv := child.Op.(protocol.Invoke)
	inv.InvocationID = protocol.InvocationID{Value: "child"}
	inv.ReturnFd = id
	child.Op = inv
	if err := s.transition(ctx, protocol.MoveBatch{Source: &id, Moves: []protocol.Move{child}}); err != nil {
		t.Fatal(err)
	}
	cid := inv.InvocationID
	ret := protocol.Move{Root: m.Root, Target: child.Target, Op: protocol.Return{Observation: observation("ok")}}
	ret.Target.Invocation = cid
	if err := s.transition(ctx, protocol.MoveBatch{Source: &cid, Moves: []protocol.Move{ret}}); err != nil {
		t.Fatal(err)
	}
	if resumed == nil {
		t.Fatal("no resumption")
	}
	if err := s.transition(ctx, protocol.DispatchResult{Invocation: id, Sequence: 0}); err != nil {
		t.Fatal(err)
	}
	c, _ := s.Snapshot(id)
	if c.Handoff.Sequence != 1 || c.Handoff.Placement != PlacementPending {
		t.Fatal("stale ack confirmed resumption")
	}
	if err := resumed(protocol.DispatchResult{Invocation: id, Sequence: 1}); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Snapshot(id)
	if c.Handoff.Placement != PlacementSucceeded {
		t.Fatal("matching ack lost")
	}
}

func TestWholeBatchRejection(t *testing.T) {
	s, _ := NewScheduler(&testController{dispatch: func(_ context.Context, m protocol.Move, seq uint64, w func(protocol.OpcodeEvent) error) error {
		return w(protocol.DispatchResult{Invocation: m.Op.(protocol.Invoke).InvocationID, Sequence: seq})
	}}, func(protocol.RootResult) error { return nil })
	m := rootMove()
	id := m.Op.(protocol.Invoke).InvocationID
	if err := s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	child := support.CloneMove(m)
	inv := child.Op.(protocol.Invoke)
	inv.InvocationID = protocol.InvocationID{Value: "child"}
	inv.ReturnFd = id
	child.Op = inv
	if err := s.transition(context.Background(), protocol.MoveBatch{Source: &id, Moves: []protocol.Move{child, child}}); err == nil {
		t.Fatal("duplicate batch accepted")
	}
	c, _ := s.Snapshot(id)
	if len(c.Move.Op.(protocol.Invoke).Register.Children) != 0 {
		t.Fatal("partial parent installation")
	}
	if _, ok := s.Snapshot(inv.InvocationID); ok {
		t.Fatal("partial child installation")
	}
}

func TestAuthorizationBeforeEffects(t *testing.T) {
	calls := 0
	s, _ := NewScheduler(&testController{dispatch: func(context.Context, protocol.Move, uint64, func(protocol.OpcodeEvent) error) error { calls++; return nil }}, func(r protocol.RootResult) error {
		if r.Err == nil {
			t.Error("unexpected success")
		}
		return nil
	})
	m := rootMove()
	inv := m.Op.(protocol.Invoke)
	inv.Horizon = protocol.Horizon{}
	m.Op = inv
	m.Stamp = protocol.AuthorizationStamp{Allowed: true}
	if err := s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err == nil {
		t.Fatal("forged stamp accepted")
	}
	if calls != 0 {
		t.Fatal("denied invoke dispatched")
	}
	if _, ok := s.Snapshot(inv.InvocationID); ok {
		t.Fatal("denied invoke installed")
	}
	// Deterministic expiry check uses an explicit timestamp, without a clock callback under mu.
	good := rootMove()
	h := good.Op.(protocol.Invoke).Horizon
	h.Capabilities[0].ExpiresAt = 10
	if err := s.authorize(&good, h, protocol.Action{Value: "echo"}, time.Unix(10, 0)); err == nil || good.Stamp.Allowed {
		t.Fatal("expired grant accepted")
	}
	// Authorize a real root, then expire trusted resident grant as a focused fixture.
	good = rootMove()
	id := good.Op.(protocol.Invoke).InvocationID
	if err := s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{good}}); err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	c := s.continuations[id]
	ci := c.Move.Op.(protocol.Invoke)
	ci.Horizon.Capabilities[0].ExpiresAt = 1
	c.Move.Op = ci
	s.continuations[id] = c
	s.mu.Unlock()
	if err := s.transition(context.Background(), protocol.DispatchResult{Invocation: id}); err != nil {
		t.Fatal("runtime performed authorization", err)
	}
	ret := protocol.Move{Root: good.Root, Target: good.Target, Op: protocol.Return{Observation: observation("denied")}, Stamp: protocol.AuthorizationStamp{Allowed: true}}
	ret.Target.Invocation = id
	if err := s.transition(context.Background(), protocol.MoveBatch{Source: &id, Moves: []protocol.Move{ret}}); err == nil {
		t.Fatal("denied Return accepted")
	}
	c, _ = s.Snapshot(id)
	if c.Result != nil {
		t.Fatal("denied return applied")
	}
}

func TestRejectedOutputConsumedAndNotRetried(t *testing.T) {
	var callback func(protocol.OpcodeEvent) error
	notifications := 0
	hostErr := errors.New("notification failed")
	s, _ := NewScheduler(&testController{dispatch: func(_ context.Context, _ protocol.Move, _ uint64, w func(protocol.OpcodeEvent) error) error { callback = w; return nil }}, func(r protocol.RootResult) error {
		notifications++
		if r.Err == nil {
			t.Error("missing rejection")
		}
		return hostErr
	})
	m := rootMove()
	id := m.Op.(protocol.Invoke).InvocationID
	if err := s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	if err := callback(protocol.DispatchResult{Invocation: id}); err != nil {
		t.Fatal(err)
	}
	child := support.CloneMove(m)
	inv := child.Op.(protocol.Invoke)
	inv.ReturnFd = id
	child.Op = inv
	batch := protocol.MoveBatch{Source: &id, Moves: []protocol.Move{child}}
	err := callback(batch)
	c, _ := s.Snapshot(id)
	if c.Handoff.Rejection == nil || !errors.Is(err, c.Handoff.Rejection) || !errors.Is(err, hostErr) {
		t.Fatal("processing or notification error lost", err)
	}
	if !c.Handoff.OutputConsumed || len(c.Handoff.PendingOutput) != 0 || !support.EmptyRegister(c.Move.Op.(protocol.Invoke).Register) || c.State != protocol.Faulted {
		t.Fatal("rejection not terminal")
	}
	if again := callback(batch); !errors.Is(again, c.Handoff.Rejection) || errors.Is(again, hostErr) {
		t.Fatal("rejection retried", again)
	}
	if err := callback(protocol.DispatchResult{Invocation: id}); err != nil {
		t.Fatal(err)
	}
	if notifications != 1 {
		t.Fatal("notification retried")
	}
}

func TestBufferedOutputIsAuthorizedOnlyOnRelease(t *testing.T) {
	var callback func(protocol.OpcodeEvent) error
	s, _ := NewScheduler(&testController{dispatch: func(_ context.Context, _ protocol.Move, _ uint64, w func(protocol.OpcodeEvent) error) error { callback = w; return nil }}, func(r protocol.RootResult) error {
		if r.Err == nil {
			t.Error("denied output notified success")
		}
		return nil
	})
	m := rootMove()
	id := m.Op.(protocol.Invoke).InvocationID
	if err := s.transition(context.Background(), protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
		t.Fatal(err)
	}
	ret := protocol.Move{Root: m.Root, Target: m.Target, Op: protocol.Return{Observation: observation("x")}}
	ret.Target.Invocation = id
	// Expire before output. Buffer acceptance must still succeed, since receipt is
	// not authorization or semantic acceptance.
	s.mu.Lock()
	c := s.continuations[id]
	inv := c.Move.Op.(protocol.Invoke)
	inv.Horizon.Capabilities[0].ExpiresAt = 1
	c.Move.Op = inv
	s.continuations[id] = c
	s.mu.Unlock()
	if err := callback(protocol.MoveBatch{Source: &id, Moves: []protocol.Move{ret}}); err != nil {
		t.Fatal("buffer performed opcode authorization", err)
	}
	if err := callback(protocol.DispatchResult{Invocation: id}); err == nil {
		t.Fatal("released Return bypassed authorization", err)
	}
	c, _ = s.Snapshot(id)
	if c.Result != nil || !c.Handoff.OutputConsumed || len(c.Handoff.PendingOutput) != 0 || c.State != protocol.Faulted {
		t.Fatal("denied Return effects or hidden consumption")
	}
}

func TestLateFaultPrecedence(t *testing.T) {
	for _, mode := range []string{"current", "returned-child", "older", "terminal"} {
		t.Run(mode, func(t *testing.T) {
			notifications := 0
			failure := errors.New("late execution failure")
			s, _ := NewScheduler(&testController{dispatch: func(_ context.Context, m protocol.Move, seq uint64, w func(protocol.OpcodeEvent) error) error {
				return w(protocol.DispatchResult{Invocation: m.Op.(protocol.Invoke).InvocationID, Sequence: seq})
			}}, func(r protocol.RootResult) error { notifications++; return nil })
			ctx := context.Background()
			m := rootMove()
			id := m.Op.(protocol.Invoke).InvocationID
			if err := s.transition(ctx, protocol.MoveBatch{Moves: []protocol.Move{m}}); err != nil {
				t.Fatal(err)
			}
			source := id
			if mode == "returned-child" || mode == "older" {
				child := support.CloneMove(m)
				inv := child.Op.(protocol.Invoke)
				inv.InvocationID = protocol.InvocationID{Value: "child"}
				inv.ReturnFd = id
				child.Op = inv
				if err := s.transition(ctx, protocol.MoveBatch{Source: &id, Moves: []protocol.Move{child}}); err != nil {
					t.Fatal(err)
				}
				cid := inv.InvocationID
				ret := protocol.Move{Root: m.Root, Target: child.Target, Op: protocol.Return{Observation: observation("ok")}}
				ret.Target.Invocation = cid
				if err := s.transition(ctx, protocol.MoveBatch{Source: &cid, Moves: []protocol.Move{ret}}); err != nil {
					t.Fatal(err)
				}
				c, _ := s.Snapshot(id)
				if c.Handoff.Sequence != 1 {
					t.Fatal("no resumption")
				}
				if mode == "returned-child" {
					source = cid
				}
			}
			if mode == "terminal" {
				ret := protocol.Move{Root: m.Root, Target: m.Target, Op: protocol.Return{Observation: observation("done")}}
				ret.Target.Invocation = id
				if err := s.transition(ctx, protocol.MoveBatch{Source: &id, Moves: []protocol.Move{ret}}); err != nil {
					t.Fatal(err)
				}
			}
			before, _ := s.Snapshot(id)
			event := protocol.ExecutionFailure{Source: source, Sequence: 0, Err: failure}
			if err := s.transition(ctx, event); err != nil {
				t.Fatal(err)
			}
			after, _ := s.Snapshot(id)
			if mode == "older" || mode == "terminal" {
				if !reflect.DeepEqual(before, after) {
					t.Fatal("stale or terminal fault changed root")
				}
				expected := 0
				if mode == "terminal" {
					expected = 1
				}
				if notifications != expected {
					t.Fatal("unexpected notification")
				}
			} else {
				c, _ := s.Snapshot(source)
				if after.State != protocol.Faulted || c.State != protocol.Faulted || c.Handoff.Placement != PlacementSucceeded || notifications != 1 {
					t.Fatal("current fault not applied or placement receipt corrupted")
				}
				if mode == "returned-child" && c.Result == nil {
					t.Fatal("accepted child result erased")
				}
			}
			if err := s.transition(ctx, event); err != nil {
				t.Fatal(err)
			}
			if mode != "older" && notifications != 1 {
				t.Fatal("fault notification repeated")
			}
		})
	}
}

