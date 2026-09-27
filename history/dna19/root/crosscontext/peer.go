// Package crosscontext implements the bounded, serial activation bridge. It only
// sees Moves and three connections, never either context's maps or Bookkeeper.
package crosscontext

import (
	"encoding/json"
	"fmt"
	"io"
	p "skyra/history/dna19/protocol"
)

func Serve(caller, activation, results io.ReadWriteCloser, executable p.ExecutableID) error {
	defer caller.Close()
	defer activation.Close()
	defer results.Close()
	input, output := json.NewDecoder(caller), json.NewEncoder(caller)
	activate, result := json.NewEncoder(activation), json.NewDecoder(results)
	var sequence uint64
	for {
		var call p.Move
		if err := input.Decode(&call); err != nil {
			return err
		}
		if err := call.Validate(); err != nil {
			return err
		}
		if call.Invoke == nil || call.Target != executable || call.Invoke.Frame == 0 {
			return fmt.Errorf("invalid activation")
		}
		if err := p.CheckArguments(call.Invoke.Args, "echo"); err != nil {
			return err
		}
		sequence++
		// Caller identities are retained locally, never accepted as B's frame identities.
		local := p.Move{Identity: sequence, Target: p.Echo, Invoke: &p.Invoke{Args: call.Clone().Invoke.Args}}
		if err := activate.Encode(local); err != nil {
			return err
		}
		var answer p.Move
		if err := result.Decode(&answer); err != nil {
			return err
		}
		if err := answer.Validate(); err != nil {
			return err
		}
		if answer.Identity != sequence || (answer.Return == nil && answer.Fault == nil) {
			return fmt.Errorf("activation correlation mismatch")
		}
		answer.Identity, answer.Context, answer.Source = call.Identity, call.Context, call.Invoke.Frame
		if answer.Return != nil {
			if answer.Return.Observation.Key != "text" {
				return fmt.Errorf("activation output contract rejected")
			}
			answer.Return.Answering = call.Invoke.Frame
		} else {
			answer.Fault.Frame = call.Invoke.Frame
		}
		if err := output.Encode(answer); err != nil {
			return err
		}
	}
}
