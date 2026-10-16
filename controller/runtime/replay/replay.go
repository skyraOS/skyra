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
)

// Store is where replay reads a lineage from: the tree on disk (Replay), or
// bytes handed over whole, as the website does in a visitor's browser. A
// store is not trusted: everything read from one must hash to its name.
type Store interface {
	Get(kind, hash string) ([]byte, error)
}

// Lineage is Replay over any store: it walks from head back to genesis, then
// forward, and returns the number of records re-derived: the chain, and every
// sibling a resolve names.
func Lineage(s Store, head string) (int, error) {
	w := walker{s: s, seen: map[string]horizon.Horizon{}}
	if _, err := w.at(head); err != nil {
		return 0, err
	}
	return len(w.seen), nil
}

// Walker re-derives Horizons for many records, each record once: for a
// reader that asks about a whole tree (skyra metric).
type Walker struct{ w walker }

// At is the Horizon the record at stamp released, as HorizonAt.
func (w *Walker) At(stamp string) (horizon.Horizon, error) { return w.w.at(stamp) }

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

// released reports whether a Horizon was committed: the release a parent
// committed for a child.
func released(s Store, h horizon.Horizon) bool {
	b, err := json.Marshal(h)
	if err != nil {
		return false
	}
	_, err = get(s, "horizons", record.Sum(b))
	return err == nil
}

// walker re-derives records, each once: seen holds the Horizon each record
// released, so siblings that share a chain share its re-derivation.
type walker struct {
	s    Store
	seen map[string]horizon.Horizon
}

// at re-derives every record from genesis (or from the last record already
// re-derived) to head, and returns the Horizon head released.
func (w *walker) at(head string) (horizon.Horizon, error) {
	var none horizon.Horizon
	if head == "" {
		return none, errors.New("empty lineage head")
	}
	var chain []string
	s := head
	for s != "" {
		if _, done := w.seen[s]; done {
			break
		}
		r, err := load(w.s, s)
		if err != nil {
			return none, err
		}
		chain = append(chain, s)
		s = r.Predecessor
	}
	slices.Reverse(chain)
	var h horizon.Horizon
	if s != "" {
		h = w.seen[s]
	} else {
		g, err := load(w.s, chain[0])
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
		rec, err := load(w.s, stamp)
		if err != nil {
			return none, err
		}
		in, err := get(w.s, "inputs", rec.Input)
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
			_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: machine.HeldHorizon(h), In: in})
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
			slots := make([]string, 0, len(o.Discharges))
			used := map[string]bool{}
			for _, d := range o.Discharges {
				if used[d] {
					return none, fmt.Errorf("resolve %s repeats a discharge", record.Short(stamp))
				}
				used[d] = true
				discharge, err := load(w.s, d)
				if err != nil {
					return none, err
				}
				if discharge.State != record.Return && discharge.State != record.Fault {
					return none, fmt.Errorf("resolve %s names an unfinished dependency", record.Short(stamp))
				}
				sib := h
				if d != rec.Predecessor {
					sib, err = w.at(d)
					if err != nil {
						return none, fmt.Errorf("resolve sibling: %w", err)
					}
				}
				if sib.Resume == nil || sib.Resume.Frame != o.Frame || sib.Resume.Stamp != o.Round || slices.Contains(slots, sib.Frame) {
					return none, fmt.Errorf("resolve %s names an unrelated or repeated slot", record.Short(stamp))
				}
				slots = append(slots, sib.Frame)
			}
			if err := w.verifyEmission(o, slots); err != nil {
				return none, err
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
				if alt, _ := machine.Fanned(h, rec, stamp); released(w.s, alt) {
					next = alt
				}
			}
			h = next
		}
		w.seen[stamp] = h
	}
	return h, nil
}

// verifyEmission reconstructs every expected slot from the immutable emission,
// in order. Version 2 histories must carry this evidence; legacy histories
// still get terminal-state, uniqueness and ownership checks.
func (w *walker) verifyEmission(o machine.Observation, slots []string) error {
	under, err := w.at(o.Round)
	if err != nil {
		return err
	}
	outer := under
	for outer.Resume != nil {
		outer = *outer.Resume
	}
	genesis, err := load(w.s, outer.Frame)
	if err != nil {
		return err
	}
	if o.Emission == "" {
		if genesis.Version >= 2 {
			return errors.New("resolve has no complete round emission")
		}
		return nil
	}
	out, err := get(w.s, "inputs", o.Emission)
	if err != nil {
		return err
	}
	parts, err := machine.Instructions(out)
	if err != nil {
		return err
	}
	if len(parts) != len(slots) {
		return errors.New("resolve does not cover every emitted instruction")
	}
	for i, part := range parts {
		_, effs := machine.Step(machine.Instance{Phase: machine.Intake}, machine.Received{Under: machine.HeldHorizon(under), In: part})
		r := machine.Committed(effs)
		if r.State != record.Invoke && !(len(parts) > 1 && r.State == record.Fault) {
			return errors.New("resolve emission did not open a slot")
		}
		if record.Stamp(r) != slots[i] {
			return errors.New("resolve discharges do not match the emitted slots in order")
		}
	}
	return nil
}
