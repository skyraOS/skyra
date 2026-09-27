// Package replay re-derives a lineage from genesis using machine.Step, and
// checks every record, supervisor faults included, against what Step would
// derive from its stored input. Nothing is re-executed.
package replay

import (
	"errors"
	"fmt"

	"skyra/controller/horizon"
	"skyra/controller/machine"
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
		in, err := tree.Get(root, "inputs", rec.Input)
		if err != nil {
			return 0, err
		}
		var derived record.Record
		if rec.Supervisor {
			// A supervisor's fault: derived from what it observed.
			derived = machine.Observed(rec.Predecessor, in)
		} else {
			_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: h, In: in})
			derived = machine.Committed(effs)
		}
		if record.Stamp(derived) != stamp {
			return 0, fmt.Errorf("record %s is not what step derives from its input", record.Short(stamp))
		}
		if rec.Supervisor {
			break // it settled the frame it was committed for; nothing follows it
		}
		if h, err = machine.Narrow(h, rec, stamp); err != nil {
			return 0, err
		}
	}
	return len(chain), nil
}
