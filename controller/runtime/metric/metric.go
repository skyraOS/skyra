// Package metric is the research metric (next/32, H8), read from the records
// alone: procedures remembered per task, and calls to them that ran with zero
// approvals and zero model steps.
//
// It reads the tree as replay does, deriving every Horizon, and claims
// nothing the records do not say:
//
//	a task                  an invoke the operator emitted (a seed)
//	a remember              an invoke of an actor that keeps state, with {"remember": ...},
//	                        whose return names what it remembered
//	a call to a procedure   an invoke selected by a contract its target did not have at
//	                        genesis: one added by extend-reach
//	a model step            an invoke, beneath that call, of a continuation other than the
//	                        procedure's own program
//	an approval             an invoke, beneath that call, of an async surface: a human
package metric

import (
	"encoding/json"
	"fmt"
	"slices"

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/replay"
	"skyra/controller/runtime/tree"
)

// Metric is what a tree says.
type Metric struct {
	Tasks      int      `json:"tasks"`
	Remembers  int      `json:"remembers"`
	Remembered []string `json:"remembered"` // what each remember returned it remembered
	Calls      int      `json:"calls"`      // calls to remembered procedures
	Unattended int      `json:"unattended"` // of them, returned with zero model steps and zero approvals
}

func (m Metric) String() string {
	per := 0.0
	if m.Tasks > 0 {
		per = float64(len(m.Remembered)) / float64(m.Tasks)
	}
	return fmt.Sprintf("metric: %d procedure(s) remembered %v in %d remember(s), %.2f per task (%d task(s)); %d call(s) to them, %d returned with zero model steps and zero approvals",
		len(m.Remembered), m.Remembered, m.Remembers, per, m.Tasks, m.Calls, m.Unattended)
}

type invoke struct {
	stamp  string
	r      record.Record
	under  horizon.Horizon // the Horizon it was emitted under
	target horizon.Surface
	name   string // the name in reach it resolved to (a reference's actor)
}

// Read reads the metric from a tree.
func Read(root string) (Metric, error) {
	stamps, err := tree.Stamps(root)
	if err != nil {
		return Metric{}, err
	}
	w := replay.NewWalker(root)
	var genesis record.Record
	var invokes []invoke
	for _, s := range stamps {
		r, err := tree.Load(root, s)
		if err != nil {
			return Metric{}, err
		}
		if r.Predecessor == "" {
			genesis = r
			continue
		}
		if r.State != record.Invoke {
			continue
		}
		under, err := w.At(r.Predecessor)
		if err != nil {
			continue // not on any lineage replay accepts: never released
		}
		name := r.To
		if g, ok := under.Refs[r.To]; ok {
			if _, named := under.Reach[r.To]; !named {
				name = g.Actor
			}
		}
		target, ok := under.Reach[name]
		if !ok {
			continue
		}
		invokes = append(invokes, invoke{s, r, under, target, name})
	}

	var m Metric
	for _, in := range invokes {
		if in.under.Surface == nil {
			m.Tasks++
		}
		if in.target.State != "" && hasKey(in.r.Payload, "remember") {
			m.Remembers++
			if d, ok := discharge(root, in.stamp); ok && d.State == record.Return {
				var got struct{ Remembered []string }
				if json.Unmarshal(d.Payload, &got) == nil {
					m.Remembered = append(m.Remembered, got.Remembered...)
				}
			}
			continue
		}
		if !added(genesis, in) {
			continue
		}
		m.Calls++
		d, ok := discharge(root, in.stamp)
		if !ok || d.State != record.Return {
			continue
		}
		attended := false
		for _, beneath := range invokes {
			if beneath.stamp == in.stamp || !within(beneath.under, in.stamp) {
				continue
			}
			model := beneath.target.Continuation && beneath.target.Path != in.target.Path
			if model || beneath.target.Host != "" {
				attended = true
				break
			}
		}
		if !attended {
			m.Unattended++
		}
	}
	slices.Sort(m.Remembered)
	return m, nil
}

// added reports whether an invoke was selected by a contract its target did
// not hold at genesis.
func added(genesis record.Record, in invoke) bool {
	v, err := contract.Decode(in.r.Payload)
	if err != nil {
		return false
	}
	c, err := machine.Selected(in.target, v)
	if err != nil {
		return false
	}
	want := string(contract.Canonical(c))
	for _, g := range genesis.Reach[in.name].All() {
		if string(contract.Canonical(g)) == want {
			return false
		}
	}
	return true
}

// within reports whether a Horizon is in frame's subtree: frame is its own
// frame or one waiting below it.
func within(h horizon.Horizon, frame string) bool {
	for f := &h; f != nil; f = f.Resume {
		if f.Frame == frame {
			return true
		}
	}
	return false
}

func discharge(root, frame string) (record.Record, bool) {
	s, err := tree.Result(root, frame)
	if err != nil {
		return record.Record{}, false
	}
	r, err := tree.Load(root, s)
	return r, err == nil
}

func hasKey(payload json.RawMessage, key string) bool {
	var o map[string]json.RawMessage
	if json.Unmarshal(payload, &o) != nil {
		return false
	}
	_, ok := o[key]
	return ok
}
