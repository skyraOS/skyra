package controller

import (
"skyra-v1/src/skyra/skyra-dna-15/protocol"
"context"
"strings"
"sync"
"reflect"
"skyra-v1/src/skyra/skyra-dna-15/driver"
"skyra-v1/src/skyra/skyra-dna-15/internal/support"
"errors"
)

func NewContinuationController(registry driver.Resolver) (*ContinuationController, error) {
	if support.Missing(registry) {
		return nil, errors.New("registry required")
	}
	return &ContinuationController{registry: registry}, nil
}

func (cc *ContinuationController) Validate(ctx context.Context, m protocol.Move) error {
	if ctx == nil {
		return errors.New("context required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	switch op := m.Op.(type) {
	case protocol.Invoke:
		if op.InvocationID.Value == "" || strings.TrimSpace(op.Action.Value) == "" {
			return errors.New("invocation identity and action required")
		}
		d, err := cc.registry.Resolve(m.Target.Address)
		if err != nil {
			return err
		}
		return d.Input.Validate(support.CloneMove(m))
	case protocol.Return:
		if m.Target.Invocation.Value == "" {
			return errors.New("return target required")
		}
		return nil
	default:
		return errors.New("unknown opcode")
	}
}

func (cc *ContinuationController) Dispatch(ctx context.Context, m protocol.Move, sequence uint64, write func(protocol.OpcodeEvent) error) error {
	op, ok := m.Op.(protocol.Invoke)
	if !ok || write == nil || ctx == nil {
		return errors.New("dispatch requires Invoke, context and writer")
	}
	d, err := cc.registry.Resolve(m.Target.Address)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return write(protocol.DispatchResult{Invocation: op.InvocationID, Sequence: sequence, Err: err})
	}
	// execute reports ingress and callback failures itself, avoiding error replay.
	return cc.execute(ctx, m, sequence, d.Driver, write)
}

func (cc *ContinuationController) execute(ctx context.Context, m protocol.Move, sequence uint64, driver driver.Driver, write func(protocol.OpcodeEvent) error) error {
	m = support.CloneMove(m)
	source := m.Op.(protocol.Invoke)
	// Per-handoff callback bookkeeping only; authoritative progress stays in Scheduler.
	var mu sync.Mutex
	emitted := false
	var callbackErr error
	output := func(raw string, executionErr error) error {
		mu.Lock()
		defer mu.Unlock()
		if emitted && executionErr == nil {
			return errors.New("handoff already emitted output")
		}
		if executionErr == nil {
			emitted = true
		}
		fail := func(err error) error {
			callbackErr = write(protocol.ExecutionFailure{Source: source.InvocationID, Sequence: sequence, Err: err})
			return callbackErr
		}
		if executionErr != nil {
			return fail(executionErr)
		}
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		moves, err := driver.Egress(raw)
		if err != nil {
			return fail(err)
		}
		if len(moves) == 0 {
			return fail(errors.New("empty output batch"))
		}
		for i, candidate := range moves {
			switch op := candidate.Op.(type) {
			case protocol.Invoke:
				id, e := protocol.NewInvocationID()
				if e != nil {
					return fail(e)
				}
				for _, f := range op.Focus {
					found := false
					for _, g := range source.Horizon.Capabilities {
						if reflect.DeepEqual(f, g) {
							found = true
							break
						}
					}
					if !found {
						return fail(errors.New("focus exceeds inherited grants"))
					}
				}
				candidate.Root = m.Root
				candidate.Target.Invocation = protocol.InvocationID{}
				op.InvocationID = id
				op.ReturnFd = source.InvocationID
				op.Register = protocol.Register{}
				op.Horizon = protocol.Horizon{Capabilities: support.CloneCapabilities(source.Horizon.Capabilities)}
				candidate.Op = op
			case protocol.Return:
				if len(moves) != 1 {
					return fail(errors.New("return must be a single output"))
				}
				candidate = protocol.Move{Root: m.Root, Target: m.Target, Op: op}
				candidate.Target.Invocation = source.InvocationID
			default:
				return fail(errors.New("unknown output opcode"))
			}
			if err := cc.Validate(ctx, candidate); err != nil {
				return fail(err)
			}
			moves[i] = support.CloneMove(candidate)
		}
		id := source.InvocationID
		callbackErr = write(protocol.MoveBatch{Source: &id, Sequence: sequence, Moves: moves})
		return callbackErr
	}
	ingressErr := driver.Ingress(ctx, support.CloneMove(m), output)
	mu.Lock()
	reportedErr := callbackErr
	mu.Unlock()
	if reportedErr != nil {
		return reportedErr
	}
	return write(protocol.DispatchResult{Invocation: source.InvocationID, Sequence: sequence, Err: ingressErr})
}

