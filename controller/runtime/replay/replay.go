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
// the continuation that was waiting on it, after every sibling it names is
// re-derived too.
package replay

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/tree"
)

// Replay walks from head back to genesis, then forward, and returns the
// number of records re-derived: the chain, and every sibling a resolve names.
func Replay(root, head string) (int, error) {
	w := walker{root: root, seen: map[string]horizon.Horizon{}}
	if _, err := w.at(head); err != nil {
		return 0, err
	}
	return len(w.seen), nil
}

// HorizonAt derives, from the tree alone, the Horizon the record at stamp
// released: the one whatever it runs, or whatever is emitted under it, runs
// under. Every record on the way is re-derived, as Replay does. Authority is
// never taken from anyone who holds the work (next/24, next/20 finding 1).
func HorizonAt(root, stamp string) (horizon.Horizon, error) {
	w := walker{root: root, seen: map[string]horizon.Horizon{}}
	return w.at(stamp)
}

// Walker re-derives Horizons for many records, each record once: for a
// reader that asks about a whole tree (skyra metric).
type Walker struct{ w walker }

// NewWalker is a Walker over a tree.
func NewWalker(root string) *Walker {
	return &Walker{walker{root: root, seen: map[string]horizon.Horizon{}}}
}

// At is the Horizon the record at stamp released, as HorizonAt.
func (w *Walker) At(stamp string) (horizon.Horizon, error) { return w.w.at(stamp) }

// walker re-derives records, each once: seen holds the Horizon each record
// released, so siblings that share a chain share its re-derivation.
type walker struct {
	root string
	seen map[string]horizon.Horizon
}

// at re-derives every record from genesis (or from the last record already
// re-derived) to head, and returns the Horizon head released.
func (w *walker) at(head string) (horizon.Horizon, error) {
	var none horizon.Horizon
	var chain []string
	s := head
	for s != "" {
		if _, done := w.seen[s]; done {
			break
		}
		r, err := tree.Load(w.root, s)
		if err != nil {
			return none, err
		}
		chain = append([]string{s}, chain...)
		s = r.Predecessor
	}
	var h horizon.Horizon
	if s != "" {
		h = w.seen[s]
	} else {
		g, err := tree.Load(w.root, chain[0])
		if err != nil {
			return none, err
		}
		if g.Predecessor != "" {
			return none, errors.New("chain does not start at genesis")
		}
		h = horizon.Horizon{Stamp: chain[0], Reach: g.Reach, Frame: chain[0]}
		if g.Authority != nil {
			h.Authority = *g.Authority
		}
		w.seen[chain[0]] = h
		chain = chain[1:]
	}
	for _, stamp := range chain {
		rec, err := tree.Load(w.root, stamp)
		if err != nil {
			return none, err
		}
		in, err := tree.Get(w.root, "inputs", rec.Input)
		if err != nil {
			return none, err
		}
		if rec.Predecessor != h.Stamp {
			return none, fmt.Errorf("record %s does not follow the Horizon it claims", record.Short(stamp))
		}
		var derived record.Record
		var o machine.Observation
		switch {
		case rec.Supervisor:
			derived = machine.Observed(rec.Predecessor, in)
		case rec.State == record.Resolve:
			if h.Resume == nil {
				return none, fmt.Errorf("resolve %s has no waiting continuation", record.Short(stamp))
			}
			if err := json.Unmarshal(in, &o); err != nil {
				return none, err
			}
			if o.Frame != h.Resume.Frame || o.Round != h.Resume.Stamp || !slices.Contains(o.Discharges, rec.Predecessor) {
				return none, fmt.Errorf("resolve %s resumes a frame that was not waiting", record.Short(stamp))
			}
			if derived, err = machine.Resolved(rec.Predecessor, in); err != nil {
				return none, err
			}
		default:
			_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: h, In: in})
			derived = machine.Committed(effs)
		}
		if record.Stamp(derived) != stamp {
			return none, fmt.Errorf("record %s is not what step derives from its input", record.Short(stamp))
		}

		switch {
		case rec.Supervisor:
			// It discharged the frame of the Horizon it was committed under.
			h = horizon.Horizon{Stamp: stamp, Reach: h.Reach, Authority: h.Authority, Frame: h.Frame, Resume: h.Resume}
		case rec.State == record.Resolve:
			// Every other slot of the round was discharged too: re-derive
			// each sibling's chain, and check it discharged a slot of this
			// round (next/32, H2: a round of N).
			for _, d := range o.Discharges {
				if d == rec.Predecessor {
					continue
				}
				sib, err := w.at(d)
				if err != nil {
					return none, fmt.Errorf("resolve %s, sibling %s: %w", record.Short(stamp), record.Short(d), err)
				}
				if sib.Resume == nil || sib.Resume.Frame != o.Frame || sib.Resume.Stamp != o.Round {
					return none, fmt.Errorf("resolve %s names %s, which discharged no slot of its round", record.Short(stamp), record.Short(d))
				}
			}
			resumed := *h.Resume
			resumed.Stamp = stamp
			h = resumed
		default:
			next, err := machine.Narrow(h, rec, stamp)
			if err != nil {
				return none, err
			}
			// One instruction of several: an invoke is forked, a fault holds
			// a slot of its own. The release its parent committed says which
			// it was.
			if rec.State != record.Return {
				if alt, _ := machine.Fanned(h, rec, stamp); tree.Released(w.root, alt) {
					next = alt
				}
			}
			h = next
		}
		w.seen[stamp] = h
	}
	return h, nil
}
