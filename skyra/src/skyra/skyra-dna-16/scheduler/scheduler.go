package scheduler

import (
	"context"
	"errors"
	"reflect"
	"skyra-v1/src/skyra/skyra-dna-16/controller"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"time"
)

func CloneContinuation(continuation Continuation) Continuation {
	continuation.Move = support.CloneMove(continuation.Move)
	continuation.DriverState = support.CloneDriverState(continuation.DriverState)
	if continuation.Result != nil {
		result := support.CloneObservation(*continuation.Result)
		continuation.Result = &result
	}
	return continuation
}

func NewScheduler(controller controller.Controller, complete func(protocol.RootResult) error) (*Scheduler, error) {
	if support.Missing(controller) || complete == nil {
		return nil, errors.New("controller and completion callback required")
	}
	return &Scheduler{controller: controller, complete: complete, continuations: make(map[protocol.InvocationID]Continuation)}, nil
}

func (scheduler *Scheduler) interpret(protocol.Register, *CompiledRegister) Assessment {
	return Waiting
}

func (scheduler *Scheduler) Snapshot(id protocol.InvocationID) (Continuation, bool) {
	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	continuation, exists := scheduler.continuations[id]
	return CloneContinuation(continuation), exists
}

func (scheduler *Scheduler) Transition(ctx context.Context, event protocol.OpcodeEvent) error {
	scheduler.mu.Lock()
	affected, notification, err := scheduler.write(ctx, event)
	var selected []Continuation
	if err == nil && notification == nil {
		for _, id := range affected {
			continuation, exists := scheduler.continuations[id]
			if !exists {
				continue
			}
			_ = scheduler.interpret(continuation.Move.Op.(protocol.Invoke).Register, continuation.CompiledRegister)
			selected = append(selected, scheduler.selectMoves(continuation.Move)...)
		}
	}
	scheduler.mu.Unlock()

	if notification != nil {
		err = errors.Join(err, scheduler.complete(*notification))
	}
	if err != nil {
		return err
	}
	var dispatchErrors []error
	for _, continuation := range selected {
		invocation := continuation.Move.Op.(protocol.Invoke)
		scheduler.mu.Lock()
		_, root, _ := scheduler.rootLocked(invocation.InvocationID)
		stopped := root.State == protocol.Faulted || root.Result != nil
		scheduler.mu.Unlock()
		if stopped {
			continue
		}
		input := protocol.DriverInput{Move: support.CloneMove(continuation.Move), State: support.CloneDriverState(continuation.DriverState)}
		saved := continuation.Context
		if dispatchErr := scheduler.controller.Dispatch(saved, input, continuation.Handoff.Sequence, func(next protocol.OpcodeEvent) error {
			return scheduler.Transition(saved, next)
		}); dispatchErr != nil {
			dispatchErrors = append(dispatchErrors, dispatchErr)
		}
	}
	return errors.Join(dispatchErrors...)
}

// write runs under the scheduler mutex and is the only event-application door.
func (scheduler *Scheduler) write(ctx context.Context, event protocol.OpcodeEvent) ([]protocol.InvocationID, *protocol.RootResult, error) {
	switch typed := event.(type) {
	case protocol.AdmissionFailure:
		if typed.Err == nil {
			return nil, nil, errors.New("admission failure requires error")
		}
		return nil, nil, typed.Err
	case protocol.MoveBatch:
		return scheduler.continuationOpcode(ctx, protocol.MoveBatch{Source: typed.Source, Sequence: typed.Sequence, Moves: support.CloneMoves(typed.Moves)})
	case protocol.DispatchResult, protocol.ExecutionFailure:
		return scheduler.runtime(typed)
	case protocol.DriverTransition:
		return scheduler.applyDriverTransition(typed)
	case protocol.Envelope:
		if typed.Runtime == nil {
			return nil, nil, errors.New("envelope requires runtime event")
		}
		affected, notification, err := scheduler.runtime(typed.Runtime)
		if err != nil || notification != nil || typed.Continuation == nil {
			return affected, notification, err
		}
		more, notification, err := scheduler.continuationOpcode(ctx, *typed.Continuation)
		return append(affected, more...), notification, err
	default:
		return nil, nil, errors.New("unknown scheduler event")
	}
}

func (scheduler *Scheduler) runtime(event protocol.OpcodeEvent) ([]protocol.InvocationID, *protocol.RootResult, error) {
	var id protocol.InvocationID
	var sequence uint64
	var failure error
	var placement bool
	switch typed := event.(type) {
	case protocol.DispatchResult:
		id, sequence, failure, placement = typed.Invocation, typed.Sequence, typed.Err, true
	case protocol.ExecutionFailure:
		id, sequence, failure = typed.Source, typed.Sequence, typed.Err
		if failure == nil {
			return nil, nil, errors.New("execution failure requires error")
		}
	default:
		return nil, nil, errors.New("unknown runtime event")
	}
	continuation, exists := scheduler.continuations[id]
	if !exists {
		return nil, nil, errors.New("unknown runtime source")
	}
	if sequence != continuation.Handoff.Sequence {
		return nil, nil, nil
	}
	_, root, _ := scheduler.rootLocked(id)
	if root.Result != nil || root.State == protocol.Faulted {
		return nil, nil, nil
	}
	if placement && continuation.Handoff.Placement != PlacementPending {
		return nil, nil, nil
	}
	if failure != nil {
		if placement {
			continuation.Handoff.Placement = PlacementFailed
		}
		continuation.Handoff.Rejection = failure
		scheduler.continuations[id] = continuation
		notification, err := scheduler.faultLocked(id, failure)
		return nil, notification, err
	}
	continuation.Handoff.Placement = PlacementSucceeded
	continuation.State = protocol.InFlight
	scheduler.continuations[id] = continuation
	return nil, nil, nil
}

// applyDriverTransition atomically replaces opaque driver state and applies all
// continuation moves. A failed transition restores the exact resident snapshot.
func (scheduler *Scheduler) applyDriverTransition(transition protocol.DriverTransition) ([]protocol.InvocationID, *protocol.RootResult, error) {
	continuation, exists := scheduler.continuations[transition.Source]
	if !exists {
		return nil, nil, errors.New("unknown driver transition source")
	}
	if transition.Sequence != continuation.Handoff.Sequence {
		return nil, nil, nil
	}
	if continuation.Handoff.Placement != PlacementSucceeded || continuation.State == protocol.Faulted || continuation.Result != nil {
		return nil, nil, errors.New("driver transition source is not active")
	}
	if err := driver.ValidateState(transition.State); err != nil {
		return nil, nil, err
	}
	before := scheduler.cloneContinuationsLocked()
	continuation.DriverState = support.CloneDriverState(transition.State)
	scheduler.continuations[transition.Source] = continuation
	if transition.Failure != nil {
		if len(transition.Moves) != 0 || len(transition.Commands) != 0 {
			scheduler.continuations = before
			return nil, nil, errors.New("failed driver transition cannot carry moves or commands")
		}
		notification, err := scheduler.faultLocked(transition.Source, transition.Failure)
		if err != nil {
			scheduler.continuations = before
			return nil, nil, err
		}
		return nil, notification, nil
	}
	if len(transition.Moves) == 0 {
		return nil, nil, nil
	}
	source := transition.Source
	affected, notification, err := scheduler.continuationOpcode(continuation.Context, protocol.MoveBatch{
		Source: &source, Sequence: transition.Sequence, Moves: support.CloneMoves(transition.Moves),
	})
	if err != nil {
		scheduler.continuations = before
		return nil, nil, err
	}
	return affected, notification, nil
}

func (scheduler *Scheduler) cloneContinuationsLocked() map[protocol.InvocationID]Continuation {
	cloned := make(map[protocol.InvocationID]Continuation, len(scheduler.continuations))
	for id, continuation := range scheduler.continuations {
		cloned[id] = CloneContinuation(continuation)
	}
	return cloned
}

func (scheduler *Scheduler) rootLocked(id protocol.InvocationID) (protocol.InvocationID, Continuation, bool) {
	for {
		continuation, exists := scheduler.continuations[id]
		if !exists {
			return id, continuation, false
		}
		parent := continuation.Move.Op.(protocol.Invoke).ReturnFd
		if parent.Value == "" {
			return id, continuation, true
		}
		id = parent
	}
}

func (scheduler *Scheduler) faultLocked(id protocol.InvocationID, cause error) (*protocol.RootResult, error) {
	continuation, exists := scheduler.continuations[id]
	if !exists {
		return nil, errors.New("unknown failure source")
	}
	rootID, root, _ := scheduler.rootLocked(id)
	if root.Result != nil || root.State == protocol.Faulted {
		return nil, nil
	}
	continuation.State = protocol.Faulted
	scheduler.continuations[id] = continuation
	root = scheduler.continuations[rootID]
	root.State = protocol.Faulted
	scheduler.continuations[rootID] = root
	return &protocol.RootResult{Invocation: rootID, Err: cause}, nil
}

func (scheduler *Scheduler) selectMoves(move protocol.Move) []Continuation {
	invocation, ok := move.Op.(protocol.Invoke)
	if !ok {
		return nil
	}
	continuation, exists := scheduler.continuations[invocation.InvocationID]
	if !exists || continuation.Result != nil || continuation.State == protocol.Faulted {
		return nil
	}
	live := continuation.Move.Op.(protocol.Invoke)
	if continuation.State != 0 {
		if continuation.Resumed || !support.CompleteRegister(live.Register) {
			return nil
		}
		continuation.Resumed = true
		continuation.Handoff.Sequence++
	}
	continuation.Handoff = Handoff{Sequence: continuation.Handoff.Sequence, Placement: PlacementPending}
	continuation.State = protocol.Dispatched
	scheduler.continuations[invocation.InvocationID] = continuation
	return []Continuation{CloneContinuation(continuation)}
}

func (scheduler *Scheduler) continuationOpcode(ctx context.Context, batch protocol.MoveBatch) ([]protocol.InvocationID, *protocol.RootResult, error) {
	if len(batch.Moves) == 0 {
		return nil, nil, errors.New("empty admission batch")
	}
	source, producer, executionContext, err := scheduler.resolveSource(ctx, batch.Source)
	if err != nil {
		return nil, nil, err
	}
	if err := scheduler.authorizeBatch(batch.Moves, batch.Source, producer); err != nil {
		return nil, nil, err
	}
	if returned, ok := batch.Moves[0].Op.(protocol.Return); ok {
		return scheduler.applyReturn(batch, source, producer, returned)
	}
	return scheduler.admitInvokes(executionContext, batch, source, producer)
}

func (scheduler *Scheduler) resolveSource(ctx context.Context, sourceID *protocol.InvocationID) (Continuation, protocol.Invoke, context.Context, error) {
	var source Continuation
	var producer protocol.Invoke
	if sourceID != nil {
		var exists bool
		source, exists = scheduler.continuations[*sourceID]
		if !exists {
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

func (scheduler *Scheduler) authorizeBatch(moves []protocol.Move, sourceID *protocol.InvocationID, producer protocol.Invoke) error {
	for index := range moves {
		move := &moves[index]
		var horizon protocol.Horizon
		var action protocol.Action
		switch operation := move.Op.(type) {
		case protocol.Invoke:
			horizon, action = operation.Horizon, operation.Action
		case protocol.Return:
			if sourceID == nil {
				return errors.New("unbound return")
			}
			horizon, action = producer.Horizon, producer.Action
		default:
			return errors.New("unknown opcode")
		}
		if err := scheduler.authorize(move, horizon, action, time.Now()); err != nil {
			return err
		}
	}
	return nil
}

func (scheduler *Scheduler) applyReturn(batch protocol.MoveBatch, source Continuation, producer protocol.Invoke, returned protocol.Return) ([]protocol.InvocationID, *protocol.RootResult, error) {
	move := batch.Moves[0]
	if len(batch.Moves) != 1 || batch.Source == nil || move.Target.Invocation != *batch.Source || move.Root != source.Move.Root || move.Target.Entity != source.Move.Target.Entity || move.Target.Address != source.Move.Target.Address || move.Target.Host != source.Move.Target.Host {
		return nil, nil, errors.New("unbound return")
	}
	if source.Result != nil {
		if reflect.DeepEqual(*source.Result, returned.Observation) {
			return nil, nil, nil
		}
		return nil, nil, errors.New("conflicting return")
	}
	_, root, _ := scheduler.rootLocked(*batch.Source)
	if root.State == protocol.Faulted || source.State == protocol.Faulted {
		return nil, nil, errors.New("workflow faulted")
	}
	if len(producer.Register.Children) > 0 && !source.Resumed {
		return nil, nil, errors.New("parent returned before resumption")
	}
	result := support.CloneObservation(returned.Observation)
	source.Result = &result
	scheduler.continuations[*batch.Source] = source
	if producer.ReturnFd.Value == "" {
		observation := support.CloneObservation(result)
		return nil, &protocol.RootResult{Invocation: *batch.Source, Observation: &observation}, nil
	}
	parent, exists := scheduler.continuations[producer.ReturnFd]
	if !exists {
		return nil, nil, errors.New("missing parent continuation")
	}
	parentInvoke := parent.Move.Op.(protocol.Invoke)
	entry, exists := parentInvoke.Register.Children[*batch.Source]
	if !exists {
		return nil, nil, errors.New("missing parent child entry")
	}
	entry.Observation = &protocol.ChildObservation{InvocationID: *batch.Source, Observation: support.CloneObservation(result)}
	parentInvoke.Register.Children[*batch.Source] = entry
	parent.Move.Op = parentInvoke
	scheduler.continuations[producer.ReturnFd] = parent
	return []protocol.InvocationID{producer.ReturnFd}, nil, nil
}

func (scheduler *Scheduler) admitInvokes(ctx context.Context, batch protocol.MoveBatch, source Continuation, producer protocol.Invoke) ([]protocol.InvocationID, *protocol.RootResult, error) {
	if batch.Source == nil {
		if len(batch.Moves) != 1 {
			return nil, nil, errors.New("submit one root")
		}
	} else {
		_, root, _ := scheduler.rootLocked(*batch.Source)
		if source.Result != nil || root.State == protocol.Faulted || source.State == protocol.Faulted || source.Resumed || len(producer.Register.Children) > 0 {
			return nil, nil, errors.New("source cannot declare another batch")
		}
		ctx = source.Context
	}
	if ctx == nil {
		return nil, nil, errors.New("context required")
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	seen := make(map[protocol.InvocationID]bool)
	for _, move := range batch.Moves {
		invocation, ok := move.Op.(protocol.Invoke)
		if !ok || invocation.InvocationID.Value == "" || !support.EmptyRegister(invocation.Register) {
			return nil, nil, errors.New("invalid invoke admission")
		}
		if _, exists := scheduler.continuations[invocation.InvocationID]; exists || seen[invocation.InvocationID] {
			return nil, nil, errors.New("invocation identity reused")
		}
		seen[invocation.InvocationID] = true
		if batch.Source == nil {
			if invocation.ReturnFd.Value != "" {
				return nil, nil, errors.New("root cannot name parent")
			}
		} else if invocation.ReturnFd != *batch.Source || move.Root != source.Move.Root || !reflect.DeepEqual(invocation.Horizon, producer.Horizon) {
			return nil, nil, errors.New("invalid child linkage or horizon")
		}
	}
	if batch.Source != nil {
		producer.Register.Children = make(map[protocol.InvocationID]protocol.RegisterEntry)
	}
	affected := make([]protocol.InvocationID, 0, len(batch.Moves))
	for _, move := range batch.Moves {
		move = support.CloneMove(move)
		invocation := move.Op.(protocol.Invoke)
		scheduler.continuations[invocation.InvocationID] = Continuation{Move: move, Context: ctx}
		affected = append(affected, invocation.InvocationID)
		if batch.Source != nil {
			producer.Register.Order = append(producer.Register.Order, invocation.InvocationID)
			producer.Register.Children[invocation.InvocationID] = protocol.RegisterEntry{Invocation: protocol.ChildInvocation{InvocationID: invocation.InvocationID, Move: support.CloneMove(move)}}
		}
	}
	if batch.Source != nil {
		source.Move.Op = producer
		scheduler.continuations[*batch.Source] = source
	}
	return affected, nil, nil
}

func (scheduler *Scheduler) authorize(move *protocol.Move, horizon protocol.Horizon, action protocol.Action, now time.Time) error {
	move.Stamp = protocol.AuthorizationStamp{CheckedAt: now.UnixNano()}
	if len(horizon.Capabilities) > 256 {
		return errors.New("too many grants")
	}
	if invocation, ok := move.Op.(protocol.Invoke); ok {
		if len(invocation.Focus) > 256 {
			return errors.New("too many focus grants")
		}
		for _, focus := range invocation.Focus {
			found := false
			for _, grant := range horizon.Capabilities {
				if reflect.DeepEqual(focus, grant) {
					found = true
					break
				}
			}
			if !found {
				return errors.New("focus exceeds horizon")
			}
		}
	}
	for _, grant := range horizon.Capabilities {
		if grant.Target.Address == move.Target.Address && grant.Target.Entity == move.Target.Entity && grant.Target.Host == move.Target.Host && grant.Action == action && (grant.ExpiresAt == 0 || now.Unix() < grant.ExpiresAt) {
			move.Stamp.Allowed = true
			return nil
		}
	}
	return errors.New("continuation authorization denied or expired")
}
