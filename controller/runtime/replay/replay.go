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
)

// Store is where replay reads a lineage from: the tree on disk (Replay), or
// bytes handed over whole, as the website does in a visitor's browser. A
// store is not trusted: everything read from one must hash to its name.
type Store interface {
	Get(kind, hash string) ([]byte, error)
}

// Lineage is Replay over any store: it walks from head back to genesis, then
// forward, and returns the number of records in the chain.
func Lineage(s Store, head string) (int, error) {
	_, n, err := walk(s, head)
	return n, err
}

// get reads kind/h from the store and checks that it hashes to its name.
func get(s Store, kind, h string) ([]byte, error) {
	b, err := s.Get(kind, h)
	if err != nil {
		return nil, err
	}
	if record.Sum(b) != h {
		return nil, fmt.Errorf("%s/%s does not hash to its name", kind, record.Short(h))
	}
	return b, nil
}

func load(s Store, stamp string) (record.Record, error) {
	b, err := get(s, "records", stamp)
	if err != nil {
		return record.Record{}, err
	}
	var r record.Record
	return r, json.Unmarshal(b, &r)
}

func walk(root Store, head string) (horizon.Horizon, int, error) {
	var none horizon.Horizon
	var chain []string
	for s := head; s != ""; {
		r, err := load(root, s)
		if err != nil {
			return none, 0, err
		}
		chain = append([]string{s}, chain...)
		s = r.Predecessor
	}
	g, err := load(root, chain[0])
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
		rec, err := load(root, stamp)
		if err != nil {
			return none, 0, err
		}
		in, err := get(root, "inputs", rec.Input)
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
func routed(root Store, waiter horizon.Horizon, discharge string, run machine.RouteRun) ([]byte, error) {
	if run.Child != discharge {
		return nil, errors.New("its route ran on another discharge")
	}
	if run.Input != "" {
		b, err := get(root, "inputs", run.Input)
		if err != nil {
			return nil, err
		}
		var in machine.RunnerInput
		if err := json.Unmarshal(b, &in); err != nil {
			return nil, err
		}
		d, err := load(root, discharge)
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
			out, err = get(root, "inputs", run.Output)
		} else {
			failed, err = get(root, "inputs", run.Failed)
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
