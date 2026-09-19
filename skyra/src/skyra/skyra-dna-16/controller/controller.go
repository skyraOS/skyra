package controller

import (
	"context"
	"errors"
	"reflect"
	"skyra-v1/src/skyra/skyra-dna-16/driver"
	"skyra-v1/src/skyra/skyra-dna-16/internal/support"
	"skyra-v1/src/skyra/skyra-dna-16/protocol"
	"strings"
	"sync"
)

func NewContinuationController(registry driver.Resolver) (*ContinuationController, error) {
	if support.Missing(registry) {
		return nil, errors.New("registry required")
	}
	return &ContinuationController{registry: registry}, nil
}

func (controller *ContinuationController) Validate(ctx context.Context, move protocol.Move) error {
	if ctx == nil {
		return errors.New("context required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	switch operation := move.Op.(type) {
	case protocol.Invoke:
		if operation.InvocationID.Value == "" || strings.TrimSpace(operation.Action.Value) == "" {
			return errors.New("invocation identity and action required")
		}
		registration, err := controller.registry.Resolve(move.Target.Address)
		if err != nil {
			return err
		}
		return registration.Input.Validate(support.CloneMove(move))
	case protocol.Return:
		if move.Target.Invocation.Value == "" {
			return errors.New("return target required")
		}
		return nil
	default:
		return errors.New("unknown opcode")
	}
}

func (controller *ContinuationController) Dispatch(ctx context.Context, input protocol.DriverInput, sequence uint64, write func(protocol.OpcodeEvent) error) error {
	invocation, ok := input.Move.Op.(protocol.Invoke)
	if !ok || write == nil || ctx == nil {
		return errors.New("dispatch requires Invoke, context, and writer")
	}
	registration, err := controller.registry.Resolve(input.Move.Target.Address)
	if err == nil {
		err = ctx.Err()
	}
	if err != nil {
		return write(protocol.DispatchResult{Invocation: invocation.InvocationID, Sequence: sequence, Err: err})
	}
	// Placement means the endpoint is resolved and has accepted execution. Commit
	// it before calling Ingress so every emission can commit before host commands.
	if err := write(protocol.DispatchResult{Invocation: invocation.InvocationID, Sequence: sequence}); err != nil {
		return err
	}
	return controller.execute(ctx, protocol.DriverInput{Move: support.CloneMove(input.Move), State: support.CloneDriverState(input.State)}, sequence, registration.Driver, write)
}

func (controller *ContinuationController) execute(ctx context.Context, input protocol.DriverInput, sequence uint64, endpoint driver.Driver, write func(protocol.OpcodeEvent) error) error {
	source := input.Move.Op.(protocol.Invoke)
	var callbackMu sync.Mutex
	var callbackErr error

	fail := func(err error) error {
		callbackErr = errors.Join(err, write(protocol.ExecutionFailure{Source: source.InvocationID, Sequence: sequence, Err: err}))
		return callbackErr
	}

	output := func(emission protocol.DriverEmission, executionErr error) error {
		callbackMu.Lock()
		defer callbackMu.Unlock()
		if callbackErr != nil {
			return callbackErr
		}
		if executionErr != nil {
			return fail(executionErr)
		}
		if emission.Failure != nil && (emission.Raw != "" || len(emission.Commands) != 0) {
			return fail(errors.New("failed driver emission cannot carry output or commands"))
		}
		if err := ctx.Err(); err != nil {
			return fail(err)
		}
		if err := driver.ValidateState(emission.State); err != nil {
			return fail(err)
		}
		for _, command := range emission.Commands {
			if command.ID == "" || len(command.Payload) == 0 {
				return fail(errors.New("host command requires id and payload"))
			}
		}

		var moves []protocol.Move
		if emission.Raw != "" {
			var err error
			moves, err = endpoint.Egress(emission.Raw)
			if err != nil {
				return fail(err)
			}
			if len(moves) == 0 {
				return fail(errors.New("nonempty driver output decoded to no moves"))
			}
			for index, candidate := range moves {
				switch operation := candidate.Op.(type) {
				case protocol.Invoke:
					id, err := protocol.NewInvocationID()
					if err != nil {
						return fail(err)
					}
					for _, focus := range operation.Focus {
						found := false
						for _, grant := range source.Horizon.Capabilities {
							if reflect.DeepEqual(focus, grant) {
								found = true
								break
							}
						}
						if !found {
							return fail(errors.New("focus exceeds inherited grants"))
						}
					}
					candidate.Root = input.Move.Root
					candidate.Target.Invocation = protocol.InvocationID{}
					operation.InvocationID = id
					operation.ReturnFd = source.InvocationID
					operation.Register = protocol.Register{}
					operation.Horizon = protocol.Horizon{Capabilities: support.CloneCapabilities(source.Horizon.Capabilities)}
					candidate.Op = operation
				case protocol.Return:
					if len(moves) != 1 {
						return fail(errors.New("return must be a single output"))
					}
					candidate = protocol.Move{Root: input.Move.Root, Target: input.Move.Target, Op: operation}
					candidate.Target.Invocation = source.InvocationID
				default:
					return fail(errors.New("unknown output opcode"))
				}
				if err := controller.Validate(ctx, candidate); err != nil {
					return fail(err)
				}
				moves[index] = support.CloneMove(candidate)
			}
		}

		transition := protocol.DriverTransition{
			Source:   source.InvocationID,
			Sequence: sequence,
			State:    support.CloneDriverState(emission.State),
			Moves:    support.CloneMoves(moves),
			Commands: support.CloneHostCommands(emission.Commands),
			Failure:  emission.Failure,
		}
		if err := write(transition); err != nil {
			return fail(err)
		}
		if len(emission.Commands) == 0 {
			return nil
		}
		executor, ok := endpoint.(driver.CommandExecutor)
		if !ok {
			return fail(errors.New("driver emitted host commands without a command executor"))
		}
		for _, command := range emission.Commands {
			if err := executor.Execute(ctx, command); err != nil {
				return fail(err)
			}
		}
		return nil
	}

	ingressErr := endpoint.Ingress(ctx, input, output)
	callbackMu.Lock()
	reportedErr := callbackErr
	callbackMu.Unlock()
	if reportedErr != nil {
		return reportedErr
	}
	if ingressErr != nil {
		return write(protocol.ExecutionFailure{Source: source.InvocationID, Sequence: sequence, Err: ingressErr})
	}
	return nil
}
