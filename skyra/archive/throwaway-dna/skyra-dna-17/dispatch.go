package dna17

import (
	"errors"
	"fmt"
)

// Dispatch fires every authorized execution concurrently: one worker per fd. The
// interpreter's job ended at the commit; Dispatch launches and returns, never
// waiting on the world. A single commit may authorize several executions (the
// current move plus any it unblocked) and they fire together.
func (cu *CU) Dispatch(items []dispatchItem) {
	for _, item := range items {
		cu.launch(item)
	}
}

func (cu *CU) launch(item dispatchItem) {
	surface, ok := cu.state.surfaceOf(item.surface)
	if !ok {
		cu.commitFailure(item.id, fmt.Errorf("dispatch: no surface %q", item.surface))
		return
	}
	cu.workers.Add(1)
	go func() {
		defer cu.workers.Done()
		// The fd (item.id) is the worker's only authority; bind it to the output.
		output := cu.makeOutput(item.id, surface)
		if err := surface.Ingress(cu.ctx, item.input, output); err != nil {
			// A native failure or a synchronous ingress error becomes a fault Move,
			// not a crash. (For long-lived drivers Ingress returns nil and emissions
			// arrive later through output.)
			cu.commitFailure(item.id, err)
		}
	}()
}

// makeOutput builds the persistent execution-plane callback for one fd. Every
// emission a Surface produces over the fd's life flows through here: it is
// normalized (Egress), committed through stateVia(id) — checked live against the fd
// — and only then are host commands sent (commit-before-send). A dead fd rejects
// the commit and the emission is dropped.
func (cu *CU) makeOutput(id Identifier, surface Surface) func(Emission, error) error {
	return func(emission Emission, execErr error) error {
		if execErr != nil {
			cu.commitFailure(id, execErr)
			return nil
		}
		if emission.Failure != nil && (emission.Raw != "" || len(emission.Commands) > 0) {
			cu.commitFailure(id, errors.New("failed emission cannot carry output or commands"))
			return nil
		}

		var moves []Move
		if emission.Raw != "" {
			decoded, err := surface.Egress(emission.Raw)
			if err != nil {
				cu.commitFailure(id, err)
				return nil
			}
			if len(decoded) == 0 {
				cu.commitFailure(id, errors.New("nonempty driver output decoded to no moves"))
				return nil
			}
			moves = decoded
		}

		_, err := cu.state.stateVia(id).Commit(cu.ctx, EmissionCommit{
			State:   emission.State,
			Moves:   moves,
			Failure: emission.Failure,
		})
		if err != nil {
			// The settlement fence: the fd is dead (settled/faulted), so this late or
			// duplicate emission commits nothing and enqueues no Move. This is exactly
			// what prevents a post-fault agent message from producing a Skyra Return.
			return nil
		}

		if len(emission.Commands) > 0 {
			executor, ok := surface.(CommandExecutor)
			if !ok {
				cu.commitFailure(id, errors.New("emission carried host commands without a command executor"))
				return nil
			}
			for _, command := range emission.Commands {
				if command.ID == "" || len(command.Payload) == 0 {
					cu.commitFailure(id, errors.New("host command requires id and payload"))
					return nil
				}
				if err := executor.Execute(cu.ctx, command); err != nil {
					cu.commitFailure(id, err)
					return nil
				}
			}
		}
		return nil
	}
}

// commitFailure records an execution-plane failure for an fd (if still live). It is
// a no-op once the fd is dead, so it cannot resurrect settled work.
func (cu *CU) commitFailure(id Identifier, cause error) {
	_, _ = cu.state.stateVia(id).Commit(cu.ctx, EmissionCommit{Failure: cause})
}
