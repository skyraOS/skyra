package scheduler

import (
	"context"
	"errors"
	"reflect"
	"skyra-v1/src/skyra/skyra-dna-15/controller"
	"skyra-v1/src/skyra/skyra-dna-15/internal/support"
	"skyra-v1/src/skyra/skyra-dna-15/protocol"
	"time"
)

func NewScheduler(cc controller.Controller, complete func(protocol.RootResult) error) (*Scheduler, error) {
	if support.Missing(cc) || complete == nil {
		return nil, errors.New("controller and completion callback required")
	}
	return &Scheduler{controller: cc, complete: complete, continuations: make(map[protocol.InvocationID]Continuation)}, nil
}

func (s *Scheduler) interpret(live protocol.Register, compiled *CompiledRegister) Assessment {
	return Waiting
}

func (s *Scheduler) Snapshot(id protocol.InvocationID) (Continuation, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.continuations[id]
	return CloneContinuation(c), ok
}

func (s *Scheduler) rootLocked(id protocol.InvocationID) (protocol.InvocationID, Continuation, bool) {
	for {
		c, ok := s.continuations[id]
		if !ok {
			return id, c, false
		}
		parent := c.Move.Op.(protocol.Invoke).ReturnFd
		if parent.Value == "" {
			return id, c, true
		}
		id = parent
	}
}

func (s *Scheduler) faultLocked(id protocol.InvocationID, err error) (*protocol.RootResult, error) {
	c, ok := s.continuations[id]
	if !ok {
		return nil, errors.New("unknown failure source")
	}
	rootID, root, _ := s.rootLocked(id)
	if root.Result != nil || root.State == protocol.Faulted {
		return nil, nil
	}
	c.State = protocol.Faulted
	s.continuations[id] = c
	root = s.continuations[rootID]
	root.State = protocol.Faulted
	s.continuations[rootID] = root
	return &protocol.RootResult{Invocation: rootID, Err: err}, nil
}

func (s *Scheduler) selectMoves(m protocol.Move) []Continuation {
	inv, ok := m.Op.(protocol.Invoke)
	if !ok {
		return nil
	}
	c, ok := s.continuations[inv.InvocationID]
	if !ok || c.Result != nil || c.State == protocol.Faulted {
		return nil
	}
	live := c.Move.Op.(protocol.Invoke)
	if c.State != 0 {
		if c.Resumed || !support.CompleteRegister(live.Register) {
			return nil
		}
		_ = s.interpret(live.Register, c.CompiledRegister)
		c.Resumed = true
		c.Handoff.Sequence++
	}
	c.Handoff = Handoff{Sequence: c.Handoff.Sequence, Placement: PlacementPending}
	c.State = protocol.Dispatched
	s.continuations[inv.InvocationID] = c
	return []Continuation{CloneContinuation(c)}
}

// continuationOpcode requires mu. Full validation and authorization precede
// atomic installation. Endpoint validation was completed outside this lock.
func (s *Scheduler) continuationOpcode(ctx context.Context, e protocol.MoveBatch) ([]protocol.InvocationID, *protocol.RootResult, error) {
	if len(e.Moves) == 0 {
		return nil, nil, errors.New("empty admission batch")
	}
	source, producer, ctx, err := s.resolveSource(ctx, e.Source)
	if err != nil {
		return nil, nil, err
	}
	if err := s.authorizeBatch(e.Moves, e.Source, producer); err != nil {
		return nil, nil, err
	}
	if ret, ok := e.Moves[0].Op.(protocol.Return); ok {
		return s.applyReturn(e, source, producer, ret)
	}
	return s.admitInvokes(ctx, e, source, producer)
}

// resolveSource requires mu. Placement may release output long after controller
// validation, so bound output uses the resident execution lifetime.
func (s *Scheduler) resolveSource(ctx context.Context, sourceID *protocol.InvocationID) (Continuation, protocol.Invoke, context.Context, error) {
	var source Continuation
	var producer protocol.Invoke
	if sourceID != nil {
		var ok bool
		source, ok = s.continuations[*sourceID]
		if !ok {
			return source, producer, ctx, errors.New("unknown output source")
		}
		producer = source.Move.Op.(protocol.Invoke)
		ctx = source.Context
	}
	if ctx == nil {
		return source, producer, ctx, errors.New("context required")
	}
	if err := ctx.Err(); err != nil {
		return source, producer, ctx, err
	}
	return source, producer, ctx, nil
}

// authorizeBatch requires mu. It stamps and validates the complete batch before
// either opcode branch mutates resident continuation state.
func (s *Scheduler) authorizeBatch(moves []protocol.Move, sourceID *protocol.InvocationID, producer protocol.Invoke) error {
	for i := range moves {
		m := &moves[i]
		var horizon protocol.Horizon
		var action protocol.Action
		switch op := m.Op.(type) {
		case protocol.Invoke:
			horizon, action = op.Horizon, op.Action
		case protocol.Return:
			if sourceID == nil {
				return errors.New("unbound return")
			}
			horizon, action = producer.Horizon, producer.Action
		default:
			return errors.New("unknown opcode")
		}
		if err := s.authorize(m, horizon, action, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

// applyReturn requires mu. It owns the source Result and the corresponding
// parent Register observation write.
func (s *Scheduler) applyReturn(e protocol.MoveBatch, source Continuation, producer protocol.Invoke, ret protocol.Return) ([]protocol.InvocationID, *protocol.RootResult, error) {
	m := e.Moves[0]
	if len(e.Moves) != 1 || e.Source == nil || m.Target.Invocation != *e.Source || m.Root != source.Move.Root || m.Target.Entity != source.Move.Target.Entity || m.Target.Address != source.Move.Target.Address || m.Target.Host != source.Move.Target.Host {
		return nil, nil, errors.New("unbound return")
	}
	if source.Result != nil {
		if reflect.DeepEqual(*source.Result, ret.Observation) {
			return nil, nil, nil
		}
		return nil, nil, errors.New("conflicting return")
	}
	_, root, _ := s.rootLocked(*e.Source)
	if root.State == protocol.Faulted || source.State == protocol.Faulted {
		return nil, nil, errors.New("workflow faulted")
	}
	if len(producer.Register.Children) > 0 && !source.Resumed {
		return nil, nil, errors.New("parent returned before resumption")
	}
	result := support.CloneObservation(ret.Observation)
	source.Result = &result
	s.continuations[*e.Source] = source
	if producer.ReturnFd.Value == "" {
		o := support.CloneObservation(result)
		return nil, &protocol.RootResult{Invocation: *e.Source, Observation: &o}, nil
	}
	parent := s.continuations[producer.ReturnFd]
	inv := parent.Move.Op.(protocol.Invoke)
	entry, exists := inv.Register.Children[*e.Source]
	if !exists {
		return nil, nil, errors.New("missing parent child entry")
	}
	entry.Observation = &protocol.ChildObservation{InvocationID: *e.Source, Observation: support.CloneObservation(result)}
	inv.Register.Children[*e.Source] = entry
	parent.Move.Op = inv
	s.continuations[producer.ReturnFd] = parent
	return []protocol.InvocationID{producer.ReturnFd}, nil, nil
}

// admitInvokes requires mu. It validates the complete batch before atomically
// installing every child continuation and parent Register entry.
func (s *Scheduler) admitInvokes(ctx context.Context, e protocol.MoveBatch, source Continuation, producer protocol.Invoke) ([]protocol.InvocationID, *protocol.RootResult, error) {
	if e.Source == nil {
		if ctx == nil {
			return nil, nil, errors.New("context required")
		}
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
		if len(e.Moves) != 1 {
			return nil, nil, errors.New("submit one root")
		}
	} else {
		_, root, _ := s.rootLocked(*e.Source)
		if source.Result != nil || root.State == protocol.Faulted || source.State == protocol.Faulted || source.Resumed || len(producer.Register.Children) > 0 {
			return nil, nil, errors.New("source cannot declare another batch")
		}
		ctx = source.Context
		if err := ctx.Err(); err != nil {
			return nil, nil, err
		}
	}
	seen := make(map[protocol.InvocationID]bool)
	for _, m := range e.Moves {
		inv, ok := m.Op.(protocol.Invoke)
		if !ok || inv.InvocationID.Value == "" || !support.EmptyRegister(inv.Register) {
			return nil, nil, errors.New("invalid invoke admission")
		}
		if _, exists := s.continuations[inv.InvocationID]; exists || seen[inv.InvocationID] {
			return nil, nil, errors.New("invocation identity reused")
		}
		seen[inv.InvocationID] = true
		if e.Source == nil {
			if inv.ReturnFd.Value != "" {
				return nil, nil, errors.New("root cannot name parent")
			}
		} else if inv.ReturnFd != *e.Source || m.Root != source.Move.Root || !reflect.DeepEqual(inv.Horizon, producer.Horizon) {
			return nil, nil, errors.New("invalid child linkage or horizon")
		}
	}
	if e.Source != nil {
		producer.Register.Children = make(map[protocol.InvocationID]protocol.RegisterEntry)
	}
	for _, m := range e.Moves {
		m = support.CloneMove(m)
		inv := m.Op.(protocol.Invoke)
		s.continuations[inv.InvocationID] = Continuation{Move: m, Context: ctx}
		if e.Source != nil {
			producer.Register.Order = append(producer.Register.Order, inv.InvocationID)
			producer.Register.Children[inv.InvocationID] = protocol.RegisterEntry{Invocation: protocol.ChildInvocation{InvocationID: inv.InvocationID, Move: support.CloneMove(m)}}
		}
	}
	if e.Source != nil {
		source.Move.Op = producer
		s.continuations[*e.Source] = source
	}
	affected := make([]protocol.InvocationID, 0, len(e.Moves))
	for _, m := range e.Moves {
		affected = append(affected, m.Op.(protocol.Invoke).InvocationID)
	}
	return affected, nil, nil
}
