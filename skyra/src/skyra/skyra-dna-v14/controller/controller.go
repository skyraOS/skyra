// Package controller validates and authorizes invocations and binds driver output to its source.
package controller

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"

	"skyra-v1/src/skyra/skyra-dna-v14/driver"
	"skyra-v1/src/skyra/skyra-dna-v14/internal/support"
	"skyra-v1/src/skyra/skyra-dna-v14/protocol"
)

// Validate returns a pure validation verdict. CTU.Submit and execute
// translate that verdict into Scheduler's event protocol. This lets execute check
// all outputs before ONE atomic batch write, without a second scheduler writer.
// Dispatch/execute retain write callbacks and no persistent invocation state.
type Controller interface {
	Validate(ctx context.Context, move protocol.Move) error
	Dispatch(ctx context.Context, move protocol.Move, write func(protocol.OpcodeEvent) error) error
}

type ContinuationController struct {
	registry  driver.Resolver
	authority Authority
}

type Authority interface {
	Check(ctx context.Context, move protocol.Move) error
}

func NewContinuationController(registry driver.Resolver, authority Authority) (*ContinuationController, error) {
	if registry == nil || support.Missing(authority) {
		return nil, errors.New("registry and authority required")
	}
	return &ContinuationController{registry: registry, authority: authority}, nil
}

// Validate checks the opcode shape. Invoke requires identity, a registered endpoint,
// and its input validator. Return requires a bound Target.Invocation and observation;
// it does not require an action/schema for driver ingress because it is not dispatched.
// Write verifies the captured Source equals Return.Target.Invocation and resolves
// that resident invocation. Neither raw data nor a public Submit may assert Source.
// Validate never checks Horizon; Invoke is authorized at actual dispatch.
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

// Dispatch binds one registration and authorizes its target/action. Lookup,
// authorization, and pre-handoff cancellation failures are reported here. After
// authorization, execute owns the ingress outcome; its error is propagated without
// a second report. Cancellation after acceptance does not prove non-execution.
func (cc *ContinuationController) Dispatch(ctx context.Context, m protocol.Move, write func(protocol.OpcodeEvent) error) error {
	op, ok := m.Op.(protocol.Invoke)
	if !ok || write == nil || ctx == nil {
		return errors.New("dispatch requires Invoke, context and writer")
	}
	d, err := cc.registry.Resolve(m.Target.Address)
	authorized := false
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		err = cc.authority.Check(ctx, support.CloneMove(m))
		authorized = err == nil
	}
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return write(protocol.DispatchResult{Invocation: op.InvocationID, Authorized: authorized, Err: err})
	}
	// execute reports ingress and callback failures itself, avoiding error replay.
	return cc.execute(ctx, m, d.Driver, write)
}

// execute captures Invoke.InvocationID, its endpoint, root, and grants in a closure,
// not an executor struct. It uses the SAME driver for Ingress and Egress.
//
// Candidate Invoke: assign a fresh InvocationID, set ReturnFd to the captured
// source, inherit root/grants, and initialize its scheduler-owned register on
// admission. Candidate Return: bind Target.Invocation to the captured source and
// carry its observation. Ignore parser-supplied identity, return route, register,
// and authority. A Return never gets a new InvocationID or goes through dispatch.
//
// output(raw, err) binds transport/parse/normalization failures to that same source.
// Otherwise validate all candidates and write one atomic MoveBatch. Errors from
// the write callback must be surfaced, not confused with endpoint failure.
// Resuming a parent captures its SAME Invoke.InvocationID and updated register.
// General retries/multiple output rounds need a later stale-callback policy.
//
// execute reports the ingress handoff outcome at most once. Success means authorized
// and accepted, not completed work. Output callbacks report their own moves/failures.
// A callback write error is surfaced as processing/reporting failure, never recast
// as failed ingress. Returned errors must not be reported again by Dispatch. Late
// acknowledgements cannot regress Result or Faulted state in Write. No retries.
func (cc *ContinuationController) execute(ctx context.Context, m protocol.Move, driver driver.Driver, write func(protocol.OpcodeEvent) error) error {
	m = support.CloneMove(m)
	source := m.Op.(protocol.Invoke)
	// Per-handoff callback bookkeeping only; authoritative progress stays in Scheduler.
	var mu sync.Mutex
	emitted := false
	var callbackErr error
	output := func(raw string, executionErr error) error {
		mu.Lock()
		defer mu.Unlock()
		if emitted {
			return errors.New("handoff already emitted output")
		}
		emitted = true
		fail := func(err error) error {
			callbackErr = write(protocol.ExecutionFailure{Source: source.InvocationID, Err: err})
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
		callbackErr = write(protocol.MoveBatch{Source: &id, Moves: moves})
		return callbackErr
	}
	ingressErr := driver.Ingress(ctx, support.CloneMove(m), output)
	mu.Lock()
	reportedErr := callbackErr
	mu.Unlock()
	if reportedErr != nil {
		return reportedErr
	}
	return write(protocol.DispatchResult{Invocation: source.InvocationID, Authorized: true, Err: ingressErr})
}
