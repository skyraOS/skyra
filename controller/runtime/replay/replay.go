// Package replay re-derives a lineage from genesis using the machine, and
// checks every record against what the machine derives from its stored input.
// Nothing is re-executed. Every kind of record is derived, none is skipped:
//
//	intake       interpret(the Horizon it was emitted under, the bytes received)
//	supervisor   Observed(the instance's stamp, what it observed)
//	resolve      Resolved(the discharge's stamp, what it observed)
//
// Replay follows the Horizons the way step does: an invoke opens a frame, a
// return or fault discharges the current frame's slot, and a resolve resumes
// the continuation that was waiting on it.
package replay

import (
	"encoding/json"
	"errors"
	"fmt"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// Replay walks from head back to genesis, then forward, and returns the
// number of records in the chain.
func Replay(root, head string) (int, error) {
	_, n, err := walk(root, head)
	return n, err
}

// HorizonAt derives, from the tree alone, the Horizon the record at stamp
// released: the one whatever it runs, or whatever is emitted under it, runs
// under. Every record on the way is re-derived, as Replay does. Authority is
// never taken from anyone who holds the work (next/24, next/20 finding 1).
func HorizonAt(root, stamp string) (horizon.Horizon, error) {
	h, _, err := walk(root, stamp)
	return h, err
}

func walk(root, head string) (horizon.Horizon, int, error) {
	var none horizon.Horizon
	var chain []string
	for s := head; s != ""; {
		r, err := tree.Load(root, s)
		if err != nil {
			return none, 0, err
		}
		chain = append([]string{s}, chain...)
		s = r.Predecessor
	}
	g, err := tree.Load(root, chain[0])
	if err != nil {
		return none, 0, err
	}
	if g.Predecessor != "" {
		return none, 0, errors.New("chain does not start at genesis")
	}

	h := horizon.Horizon{Stamp: chain[0], Reach: g.Reach, Frame: chain[0]}
	if g.Authority != nil {
		h.Authority = *g.Authority
	}
	var sent []byte // a route's send: exactly what the next record must have received
	for _, stamp := range chain[1:] {
		rec, err := tree.Load(root, stamp)
		if err != nil {
			return none, 0, err
		}
		in, err := tree.Get(root, "inputs", rec.Input)
		if err != nil {
			return none, 0, err
		}
		if rec.Predecessor != h.Stamp {
			return none, 0, fmt.Errorf("record %s does not follow the Horizon it claims", record.Short(stamp))
		}
		if sent != nil && (rec.Supervisor || rec.State == record.Resolve || string(in) != string(sent)) {
			return none, 0, fmt.Errorf("record %s is not the send its route decided", record.Short(stamp))
		}
		sent = nil
		var derived record.Record
		switch {
		case rec.Supervisor:
			derived = machine.Observed(rec.Predecessor, in)
		case rec.State == record.Resolve:
			if h.Resume == nil {
				return none, 0, fmt.Errorf("resolve %s has no waiting continuation", record.Short(stamp))
			}
			var o machine.Observation
			if err := json.Unmarshal(in, &o); err != nil {
				return none, 0, err
			}
			if o.Frame != h.Resume.Frame || o.Round != h.Resume.Stamp {
				return none, 0, fmt.Errorf("resolve %s resumes a frame that was not waiting", record.Short(stamp))
			}
			if derived, err = machine.Resolved(rec.Predecessor, in); err != nil {
				return none, 0, err
			}
			if o.Route != nil {
				if sent, err = routed(root, *h.Resume, rec.Predecessor, *o.Route); err != nil {
					return none, 0, fmt.Errorf("resolve %s: %w", record.Short(stamp), err)
				}
			}
		default:
			_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: h, In: in})
			derived = machine.Committed(effs)
		}
		if record.Stamp(derived) != stamp {
			return none, 0, fmt.Errorf("record %s is not what step derives from its input", record.Short(stamp))
		}

		switch {
		case rec.Supervisor:
			// It discharged the frame of the Horizon it was committed under.
			h = horizon.Horizon{Stamp: stamp, Reach: h.Reach, Authority: h.Authority, Frame: h.Frame, Resume: h.Resume}
		case rec.State == record.Resolve:
			resumed := *h.Resume
			resumed.Stamp = stamp
			h = resumed
		default:
			if h, err = machine.Narrow(h, rec, stamp); err != nil {
				return none, 0, err
			}
		}
	}
	return h, len(chain), nil
}

// routed re-derives how a round was routed (next/30): the runner ran on the
// discharge this resolve follows, over that discharge's return, and its
// decision is what step derives from the runner's stored emission. It returns
// a send's envelope, which the next record must have received.
func routed(root string, waiter horizon.Horizon, discharge string, run machine.RouteRun) ([]byte, error) {
	if run.Child != discharge {
		return nil, errors.New("its route ran on another discharge")
	}
	if run.Input != "" {
		b, err := tree.Get(root, "inputs", run.Input)
		if err != nil {
			return nil, err
		}
		var in machine.RunnerInput
		if err := json.Unmarshal(b, &in); err != nil {
			return nil, err
		}
		d, err := tree.Load(root, discharge)
		if err != nil {
			return nil, err
		}
		if record.Sum([]byte(in.Source)) != run.Source || string(in.Return) != string(d.Payload) {
			return nil, errors.New("its route ran another source, or over another return")
		}
	}
	if run.Output != "" || run.Failed != "" {
		var out, failed []byte
		var err error
		if run.Output != "" {
			out, err = tree.Get(root, "inputs", run.Output)
		} else {
			failed, err = tree.Get(root, "inputs", run.Failed)
		}
		if err != nil {
			return nil, err
		}
		want, _ := json.Marshal(run)
		got, _ := json.Marshal(machine.Decided(waiter, run, out, string(failed)))
		if string(got) != string(want) {
			return nil, errors.New("its decision is not what step derives from the route runner's emission")
		}
	}
	if run.Decision == machine.Send {
		return run.Send, nil
	}
	return nil, nil
}
