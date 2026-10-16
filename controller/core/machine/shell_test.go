package machine

import (
	"encoding/json"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// The tests drive Step as the shell would, in memory and with no OS. These
// helpers are the shell's honest answers: every write is new, and reads are
// answered from what the test says the tree holds.

// created answers a commit: every write is new.
func created(effs []Effect) Committed {
	c := effs[len(effs)-1].(Commit)
	out := make([]Outcome, len(c.Writes))
	return Committed{Outcomes: out} // Created is the zero Outcome
}

// committing answers the commit effs ends with, if it ends with one.
func committing(s Instance, effs []Effect) (Instance, []Effect) {
	if len(effs) > 0 {
		if _, ok := effs[len(effs)-1].(Commit); ok {
			return Step(s, created(effs))
		}
	}
	return s, effs
}

// writes is every write of the commit effs ends with.
func writes(t *testing.T, effs []Effect) []Write {
	t.Helper()
	if len(effs) == 0 {
		t.Fatal("no effects")
	}
	c, ok := effs[len(effs)-1].(Commit)
	if !ok {
		t.Fatalf("not a commit: %#v", effs)
	}
	return c.Writes
}

// committedRecords decodes every record a commit writes.
func committedRecords(t *testing.T, effs []Effect) []record.Record {
	t.Helper()
	var rs []record.Record
	for _, w := range writes(t, effs) {
		if o, ok := w.(ObjectWrite); ok && o.Kind == "records" {
			var r record.Record
			if err := json.Unmarshal(o.Bytes, &r); err != nil {
				t.Fatal(err)
			}
			rs = append(rs, r)
		}
	}
	return rs
}

// inputOf is the bytes a commit stores as an input.
func inputOf(t *testing.T, effs []Effect) []byte {
	t.Helper()
	for _, w := range writes(t, effs) {
		if o, ok := w.(ObjectWrite); ok && o.Kind == "inputs" {
			return o.Bytes
		}
	}
	t.Fatalf("the commit stores no input: %#v", effs)
	return nil
}

// acts is the actions of the executes effs holds, in order.
func acts(effs []Effect) []Action {
	var out []Action
	for _, e := range effs {
		if x, ok := e.(Execute); ok {
			out = append(out, x.Act)
		}
	}
	return out
}

// exited reports whether effs ends this instance, and with what error.
func exited(effs []Effect) (bool, string) {
	a := acts(effs)
	if len(a) == 0 {
		return false, ""
	}
	x, ok := a[len(a)-1].(Exit)
	return ok, x.Err
}

// cleanExit reports whether effs is exactly a clean exit.
func cleanExit(effs []Effect) bool {
	ok, err := exited(effs)
	return ok && err == "" && len(effs) == 1
}

// released answers an instance's intake commit and its release: the parent
// names self and passes its whole frame stack. What comes back is what the
// instance does once it has acknowledged custody (or why it refused).
func released(s Instance, self horizon.Horizon) (Instance, []Effect) {
	return releasedWith(s, HeldHorizon(self), self.Depth())
}

// releasedWith is released with the bytes the shell read and the number of
// descriptors that arrived, so a test can lie about either.
func releasedWith(s Instance, self Object, frames int) (Instance, []Effect) {
	if s.Phase == Stamping {
		s, _ = Step(s, Committed{Outcomes: []Outcome{Created, Created}})
	}
	s, _ = Step(s, Executed{Obs: []byte(self.Key), Handles: frames})
	s, effs := Step(s, ReadResult{Found: [][]byte{self.Raw}})
	if s.Phase != Acknowledging {
		return s, effs
	}
	s, _ = Step(s, Committed{Outcomes: []Outcome{Created}})
	return Step(s, Executed{})
}

// emitted reports a target's output and answers the commit of the Horizon
// its first successor intakes under: what comes back is that spawn.
func emitted(s Instance, out []byte) (Instance, []Effect) {
	return committing(Step(s, Executed{Obs: out}))
}

// spawned is the Spawn effs ends with.
func spawned(t *testing.T, effs []Effect) Spawn {
	t.Helper()
	a := acts(effs)
	if len(a) != 1 {
		t.Fatalf("not one spawn: %#v", effs)
	}
	sp, ok := a[0].(Spawn)
	if !ok {
		t.Fatalf("not a spawn: %#v", effs)
	}
	return sp
}

// stamped answers a spawn with the successor's stamp, and the read of its
// record with the record's bytes. A successor still to spawn is spawned.
func stampedBy(s Instance, r record.Record) (Instance, []Effect) {
	o := HeldRecord(r)
	return stampedWith(s, o.Key, o.Raw)
}

func stampedWith(s Instance, stamp string, raw []byte) (Instance, []Effect) {
	s, _ = Step(s, Executed{Obs: []byte(stamp)})
	s, effs := Step(s, ReadResult{Found: [][]byte{raw}})
	if s.Phase == Spawning {
		return committing(s, effs)
	}
	return s, effs
}

// gatheredFrom answers gather's reads from what the tree holds, given as the
// rounds it would read: their records, their slots, and the discharges.
func gatheredFrom(t *testing.T, s Instance, effs []Effect, tree Settled) (Instance, []Effect) {
	t.Helper()
	objects := map[string][]byte{tree.Frame.Key: tree.Frame.Raw}
	slots, emissions, results := map[string][]string{}, map[string]string{}, map[string]string{}
	for _, r := range tree.Rounds {
		emissions[r.Stamp] = r.Emission
		for _, c := range r.Children {
			objects[c.Invoke.Key] = c.Invoke.Raw
			slots[r.Stamp] = append(slots[r.Stamp], c.Invoke.Key)
			if c.Discharge != nil {
				objects[c.Discharge.Key] = c.Discharge.Raw
				results[c.Invoke.Key] = c.Discharge.Key
			}
		}
	}
	for {
		rd, ok := effs[len(effs)-1].(Read)
		if !ok {
			return s, effs
		}
		found := make([][]byte, len(rd.Keys))
		for i, k := range rd.Keys {
			switch k := k.(type) {
			case ObjectKey:
				found[i] = objects[k.Hash]
			case SlotsKey:
				if l, ok := slots[k.Round]; ok {
					found[i], _ = json.Marshal(l)
				}
			case EmissionKey:
				if e, ok := emissions[k.Round]; ok && e != "" {
					found[i] = []byte(e)
				}
			case ResultKey:
				if r, ok := results[k.Frame]; ok {
					found[i] = []byte(r)
				}
			default:
				t.Fatalf("gather read %T", k)
			}
		}
		s, effs = Step(s, ReadResult{Found: found})
	}
}
