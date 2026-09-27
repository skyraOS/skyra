// Package boot is the operator's process: it writes genesis, then runs as an
// instance whose target (the operator) has just emitted the seed Move.
package boot

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"skyra/controller/horizon"
	"skyra/controller/instance"
	"skyra/controller/machine"
	"skyra/controller/record"
	"skyra/controller/registry"
	"skyra/controller/replay"
	"skyra/controller/trace"
	"skyra/controller/tree"
)

// Run boots: genesis from the operator's registry, then the operator emits
// the seed (addressed to a capability in the root reach), then boot waits for
// the lineage to settle, prints it, and replays it.
func Run(root, registryFile, to string, payload json.RawMessage) error {
	if err := tree.Init(root); err != nil {
		return err
	}
	reach, err := registry.Load(registryFile)
	if err != nil {
		return err
	}

	// Genesis: the root Horizon's reach comes from the operator.
	reachBytes, _ := json.Marshal(reach)
	g0, err := tree.Commit(root, record.Record{Input: record.Sum(reachBytes), State: record.Invoke, Reach: reach}) // frame 0: the operator's, in invoke
	if err != nil {
		return err
	}
	node0, err := tree.OpenFrame(root, g0)
	if err != nil {
		return err
	}
	defer node0.Close()
	h0 := horizon.Horizon{Stamp: g0, Reach: reach}
	trace.Logf("genesis  committed %s  root reach %v", record.Short(g0), h0.Names())

	// Boot is an instance whose target is the operator: it is already live
	// under the root Horizon, and its target has just emitted the seed. The
	// seed is operands like any other output; intake derives its state.
	seed, _ := json.Marshal(map[string]any{"to": to, "payload": payload})
	s0 := machine.Instance{Floor: root, State: record.Invoke, Phase: machine.Execute, Stamp: g0, Self: h0}
	s0, err = instance.Drive(root, nil, node0, s0, machine.Emitted{Out: seed})
	if err != nil {
		return err
	}
	s1 := s0.Child // "" if boot faulted frame 0 itself

	// The operator waits for the lineage to settle. An accepted seed settles
	// frame 1. A rejected seed settles frame 0.
	result, err := waitResult(root, s1, g0)
	if err != nil {
		return err
	}
	fmt.Println()
	fmt.Println("lineage (newest first):")
	for s := result; s != ""; {
		r, err := tree.Load(root, s)
		if err != nil {
			return err
		}
		fmt.Printf("  %s  %s\n", record.Short(s), record.Describe(r))
		s = r.Predecessor
	}

	n, err := replay.Replay(root, result)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}
	fmt.Printf("replay: %d records re-derived from genesis by step, byte-identical\n", n)
	return nil
}

func waitResult(root string, frames ...string) (string, error) {
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		for _, f := range frames {
			if f == "" {
				continue
			}
			if r, err := tree.Result(root, f); err == nil {
				return r, nil
			}
		}
		time.Sleep(5 * time.Millisecond)
	}
	return "", errors.New("lineage did not settle")
}
