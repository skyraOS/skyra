// Package replay re-derives a lineage from genesis using machine.Step, and
// checks every record against what Step would produce. Nothing is re-executed.
package replay

import (
	"errors"
	"fmt"

	"skyra/controller/horizon"
	"skyra/controller/machine"
	"skyra/controller/move"
	"skyra/controller/record"
	"skyra/controller/tree"
)

// Replay walks from head back to genesis, then forward. For each record it
// loads the recorded input, re-derives the Horizon the input was emitted under
// (machine.Narrow, starting from genesis's root reach), runs the Intake step,
// and requires the committed record to hash to the recorded stamp. It returns
// the number of records in the chain.
func Replay(root, head string) (int, error) {
	var chain []string
	for s := head; s != ""; {
		r, err := tree.Load(root, s)
		if err != nil {
			return 0, err
		}
		chain = append([]string{s}, chain...)
		s = r.Predecessor
	}
	g, err := tree.Load(root, chain[0])
	if err != nil {
		return 0, err
	}
	if g.Predecessor != "" {
		return 0, errors.New("chain does not start at genesis")
	}

	h := horizon.Horizon{Stamp: chain[0], Reach: g.Reach}
	for i := 1; i < len(chain); i++ {
		stamp := chain[i]
		rec, err := tree.Load(root, stamp)
		if err != nil {
			return 0, err
		}
		if rec.Supervisor {
			// Not produced by intake. It is only allowed to be a plain fault.
			if rec.State != move.StateFault || rec.Move != nil {
				return 0, fmt.Errorf("supervisor record %s is not a plain fault", record.Short(stamp))
			}
			continue
		}
		in, err := tree.Get(root, "inputs", rec.Input)
		if err != nil {
			return 0, err
		}
		_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: h, In: in})
		if record.Stamp(machine.Committed(effs)) != stamp {
			return 0, fmt.Errorf("record %s is not what step produces from its input", record.Short(stamp))
		}
		if h, err = machine.Narrow(h, rec, stamp); err != nil {
			return 0, err
		}
	}
	return len(chain), nil
}
