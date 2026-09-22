package scheduler

import (
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"time"
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"context"
"errors"
"reflect"
)

// Transition is the exported progression entry for the CTU, the scheduler's only
// external driver. It forwards to the internal transition.
func (s *Scheduler) Transition(ctx context.Context, event protocol.OpcodeEvent) error {
	return s.transition(ctx, event)
}

// transition is the sole progression entry. Helpers require mu; all external
// work runs after unlock. Saved contexts travel with reserved snapshots.
func (s *Scheduler) transition(ctx context.Context, event protocol.OpcodeEvent) error {
	s.mu.Lock()
	affected, notification, err := s.write(ctx, event)
	var selected []Continuation
	if err == nil && notification == nil {
		for _, id := range affected {
			c := s.continuations[id]
			_ = s.interpret(c.Move.Op.(protocol.Invoke).Register, c.CompiledRegister)
			selected = append(selected, s.selectMoves(c.Move)...)
		}
	}
	s.mu.Unlock()
	if notification != nil {
		err = errors.Join(err, s.complete(*notification))
	}
	if err != nil {
		return err
	}
	var errs []error
	for _, c := range selected {
		s.mu.Lock()
		_, root, _ := s.rootLocked(c.Move.Op.(protocol.Invoke).InvocationID)
		stopped := root.State == protocol.Faulted || root.Result != nil
		s.mu.Unlock()
		if stopped {
			continue
		}
		saved := c.Context
		if err := s.controller.Dispatch(saved, c.Move, c.Handoff.Sequence, func(e protocol.OpcodeEvent) error { return s.transition(saved, e) }); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// write recognizes the two families and delegates application, under mu.
func (s *Scheduler) write(ctx context.Context, event protocol.OpcodeEvent) ([]protocol.InvocationID, *protocol.RootResult, error) {
	switch e := event.(type) {
	case protocol.AdmissionFailure:
		if e.Err == nil {
			return nil, nil, errors.New("admission failure requires error")
		}
		return nil, nil, e.Err
	case protocol.MoveBatch:
		return s.continuation(ctx, e)
	case protocol.DispatchResult, protocol.ExecutionFailure:
		return s.runtime(e)
	case protocol.Envelope:
		// Validate association before applying either portion.
		if e.Continuation == nil || e.Continuation.Source == nil {
			return nil, nil, errors.New("envelope requires bound output")
		}
		var id protocol.InvocationID
		var seq uint64
		switch r := e.Runtime.(type) {
		case protocol.DispatchResult:
			id, seq = r.Invocation, r.Sequence
		case protocol.ExecutionFailure:
			id, seq = r.Source, r.Sequence
		default:
			return nil, nil, errors.New("envelope requires runtime event")
		}
		if id != *e.Continuation.Source || seq != e.Continuation.Sequence {
			return nil, nil, errors.New("envelope association mismatch")
		}
		// Validate the complete envelope before any effects. Runtime release
		// owns a matching buffered batch; the continuation portion must not
		// apply it a second time or strand the affected children with an error.
		repeated := false
		if receipt, ok := e.Runtime.(protocol.DispatchResult); ok && receipt.Err == nil {
			source, exists := s.continuations[id]
			if exists && source.Handoff.Sequence == seq && source.Handoff.Placement == PlacementPending && source.Handoff.PendingOutput != nil {
				if !reflect.DeepEqual(source.Handoff.PendingOutput, e.Continuation.Moves) {
					return nil, nil, errors.New("conflicting envelope output")
				}
				repeated = true
			}
		}
		affected, n, err := s.runtime(e.Runtime)
		if err != nil || n != nil || repeated {
			return affected, n, err
		}
		more, n, err := s.continuation(ctx, *e.Continuation)
		return append(affected, more...), n, err
	default:
		return nil, nil, errors.New("unknown scheduler event")
	}
}

// runtime requires mu. Placement does not authorize; eligible output enters the
// continuation opcode path. Matching is by source AND handoff sequence.
func (s *Scheduler) runtime(event protocol.OpcodeEvent) ([]protocol.InvocationID, *protocol.RootResult, error) {
	var id protocol.InvocationID
	var seq uint64
	var failure error
	var placement bool
	switch e := event.(type) {
	case protocol.DispatchResult:
		id, seq, failure, placement = e.Invocation, e.Sequence, e.Err, true
	case protocol.ExecutionFailure:
		id, seq, failure = e.Source, e.Sequence, e.Err
		if failure == nil {
			return nil, nil, errors.New("execution failure requires error")
		}
	default:
		return nil, nil, errors.New("unknown runtime event")
	}
	c, ok := s.continuations[id]
	if !ok {
		return nil, nil, errors.New("unknown runtime source")
	}
	if seq != c.Handoff.Sequence {
		return nil, nil, nil
	}
	_, root, _ := s.rootLocked(id)
	if root.Result != nil || root.State == protocol.Faulted {
		return nil, nil, nil
	}
	if c.Handoff.Rejection != nil {
		return nil, nil, c.Handoff.Rejection
	}
	if placement && c.Handoff.Placement != PlacementPending {
		return nil, nil, nil
	}
	if failure != nil {
		if placement {
			c.Handoff.Placement = PlacementFailed
		}
		c.Handoff.PendingOutput = nil
		c.Handoff.OutputConsumed = true
		s.continuations[id] = c
		n, err := s.faultLocked(id, failure)
		return nil, n, err
	}
	c.Handoff.Placement = PlacementSucceeded
	c.State = protocol.InFlight
	s.continuations[id] = c
	if c.Handoff.PendingOutput == nil {
		return nil, nil, nil
	}
	return s.continuation(c.Context, protocol.MoveBatch{Source: &id, Sequence: seq, Moves: c.Handoff.PendingOutput})
}

// continuation requires mu. Early output is detached and accepted immediately;
// no semantic Register changes occur until matching placement succeeds.
func (s *Scheduler) continuation(ctx context.Context, e protocol.MoveBatch) ([]protocol.InvocationID, *protocol.RootResult, error) {
	e.Moves = support.CloneMoves(e.Moves)
	if e.Source == nil {
		return s.continuationOpcode(ctx, e)
	}
	id := *e.Source
	c, ok := s.continuations[id]
	if !ok {
		return nil, nil, errors.New("unknown output source")
	}
	if e.Sequence != c.Handoff.Sequence {
		return nil, nil, nil
	}
	if c.Handoff.Rejection != nil {
		return nil, nil, c.Handoff.Rejection
	}
	if len(e.Moves) == 0 {
		return nil, nil, errors.New("empty output batch")
	}
	if c.Handoff.OutputConsumed {
		if len(e.Moves) == 1 && c.Result != nil {
			if ret, ok := e.Moves[0].Op.(protocol.Return); ok {
				m := e.Moves[0]
				target := c.Move.Target
				target.Invocation = id
				if m.Target != target || m.Root != c.Move.Root {
					return nil, nil, errors.New("unbound return")
				}
				if reflect.DeepEqual(ret.Observation, *c.Result) {
					return nil, nil, nil
				}
				return nil, nil, errors.New("conflicting return")
			}
		}
		return nil, nil, errors.New("handoff output already consumed")
	}
	_, root, _ := s.rootLocked(id)
	if root.State == protocol.Faulted || root.Result != nil || c.Handoff.Placement == PlacementFailed {
		return nil, nil, nil
	}
	if c.Handoff.PendingOutput != nil && !reflect.DeepEqual(c.Handoff.PendingOutput, e.Moves) {
		return nil, nil, errors.New("conflicting pending output")
	}
	if c.Handoff.Placement == PlacementPending {
		c.Handoff.PendingOutput = e.Moves
		s.continuations[id] = c
		return nil, nil, nil
	}
	affected, n, err := s.continuationOpcode(c.Context, e)
	if err != nil {
		c = s.continuations[id]
		c.Handoff.PendingOutput = nil
		c.Handoff.OutputConsumed = true
		c.Handoff.Rejection = err
		s.continuations[id] = c
		notification, faultErr := s.faultLocked(id, err)
		return nil, notification, errors.Join(err, faultErr)
	}
	c = s.continuations[id]
	c.Handoff.PendingOutput = nil
	c.Handoff.OutputConsumed = true
	s.continuations[id] = c
	return affected, n, nil
}

// authorize is a bounded, trusted-fixture check under mu, never an arbitrary
// callback. ExpiresAt is Unix seconds, zero means no expiry. Input stamps cannot
// authorize themselves. Return uses the resident producer's action and Horizon.
func (s *Scheduler) authorize(m *protocol.Move, h protocol.Horizon, action protocol.Action, now time.Time) error {
	m.Stamp = protocol.AuthorizationStamp{CheckedAt: now.UnixNano()}
	if len(h.Capabilities) > 256 {
		return errors.New("too many grants")
	}
	if inv, ok := m.Op.(protocol.Invoke); ok {
		if len(inv.Focus) > 256 {
			return errors.New("too many focus grants")
		}
		for _, f := range inv.Focus {
			found := false
			for _, g := range h.Capabilities {
				if reflect.DeepEqual(f, g) {
					found = true
					break
				}
			}
			if !found {
				return errors.New("focus exceeds horizon")
			}
		}
	}
	for _, g := range h.Capabilities {
		if g.Target.Address == m.Target.Address && g.Target.Entity == m.Target.Entity && g.Target.Host == m.Target.Host && g.Action == action && (g.ExpiresAt == 0 || now.Unix() < g.ExpiresAt) {
			m.Stamp.Allowed = true
			return nil
		}
	}
	return errors.New("continuation authorization denied or expired")
}

