// Package boot is the operator's process: it writes genesis, then runs as an
// instance whose target (the operator) has just emitted the seed Move.
package boot

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"skyra/controller/horizon"
	"skyra/controller/instance"
	"skyra/controller/machine"
	"skyra/controller/move"
	"skyra/controller/record"
	"skyra/controller/replay"
	"skyra/controller/trace"
	"skyra/controller/tree"
)

// Run boots to echo: genesis, the seed Move, then waits for the lineage to
// settle, prints it, and replays it.
func Run(root, echoPath string, args map[string]string) error {
	if err := tree.Init(root); err != nil {
		return err
	}
	echoPath, err := filepath.Abs(echoPath)
	if err != nil {
		return err
	}
	bin, err := os.ReadFile(echoPath)
	if err != nil {
		return err
	}

	// Genesis: the root Horizon's reach comes from the operator.
	reach := map[string]horizon.Surface{"echo": {Path: echoPath, SHA256: record.Sum(bin), Leaf: true}}
	reachBytes, _ := json.Marshal(reach)
	g0, err := tree.Commit(root, record.Record{Input: record.Sum(reachBytes), State: move.StateInvoke, Reach: reach}) // frame 0: the operator's, in invoke
	if err != nil {
		return err
	}
	node0, err := tree.OpenFrame(root, g0)
	if err != nil {
		return err
	}
	defer node0.Close()
	h0 := horizon.Horizon{Stamp: g0, Reach: reach}
	trace.Logf("genesis  committed %s  root reach %v (echo is a leaf)", record.Short(g0), h0.Names())

	// Boot is an instance whose target is the operator: it is already live
	// under the root Horizon, and its target has just emitted the seed Move.
	seed, _ := json.Marshal(move.Move{Invoke: &move.Invoke{Target: "echo", Args: args}})
	s0 := machine.Instance{Floor: root, State: move.StateInvoke, Phase: machine.Execute, Stamp: g0, Self: h0}
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
