package model

// Programs and boot, as Model.lean defines them. A program is the
// definition a process is instantiated from; boot(program, grant) pushes a
// process's first frame, or refuses. It runs from a root, as its nth
// invoke, or at genesis from nothing.

import (
	"encoding/json"
	"slices"

	"skyra/typesys-prototype/tree"
)

// Constants are a program's: k, the transition and refinement bounds, ι as
// a ratio, the tally's floor, and the quorum: how many implementations a
// key needs before its tally is trusted. Quorum is how many independent
// sightings before anything is trusted, ι how much they must agree once
// there are enough.
type Constants struct {
	Act         int    `json:"act"`
	Transitions int    `json:"transitions"`
	Refinements int    `json:"refinements"`
	Inhibition  [2]int `json:"inhibition"`
	Floor       int    `json:"floor"`
	Quorum      int    `json:"quorum"`
}

// A Program is the language version, the constants, the global scope by
// name, each name's interface, and the entry scope a grant must have.
type Program struct {
	Version   DefLoc           `json:"version"`
	Constants Constants        `json:"constants"`
	Global    map[string]Iface `json:"global"`
	Entry     Scope            `json:"entry"`
}

// Reach is the definitions an operator reaches: the alternatives of the
// sum at its interface's first position. act's are the program's tools,
// delegate's its models.
func (p Program) Reach(name string) []DefLoc {
	i, ok := p.Global[name]
	if !ok || i.Tag != IOp || len(i.Parts) == 0 || i.Parts[0].Tag != ISum {
		return nil
	}
	return Defs(i.Parts[0].Parts)
}

// WellFormed is boot's first check. Global names are each once by
// construction, a map's keys.
func (p Program) WellFormed() bool {
	c := p.Constants
	if c.Act < 1 || c.Transitions < 1 || c.Refinements < 1 ||
		c.Inhibition[1] <= 0 || c.Inhibition[0] > c.Inhibition[1] || c.Floor < 1 || c.Floor > 4 || c.Quorum < 1 {
		return false
	}
	models := p.Reach("delegate")
	for _, t := range p.Reach("act") {
		if slices.Contains(models, t) {
			return false
		}
	}
	return true
}

// A Grant is what boot hands the first frame: definitions, tools and
// models held by the booting frame; and context values.
type Grant struct {
	Defs    []DefLoc
	Context []any
}

func (g *Grant) UnmarshalJSON(b []byte) error {
	var x struct {
		Defs    []DefLoc `json:"defs"`
		Context []any    `json:"context"`
	}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	g.Defs = x.Defs
	g.Context = make([]any, len(x.Context))
	for i, v := range x.Context {
		g.Context[i] = DecodeValue(v)
	}
	return nil
}

// Ids are the hashes: a child by its parent, invoke index and scope; a first
// frame by its program and grant; a slot by its frame and $n; a program's
// location. The core computes none of them itself.
type Ids interface {
	Child(parent InstLoc, n int, s Scope) InstLoc
	First(program DefLoc, g Grant) InstLoc
	Slot(frame InstLoc, n int) InstLoc
	Program(p Program) DefLoc
}

// A Booter is the frame a boot runs from, and which of its invokes it is.
type Booter struct {
	Frame  Frame
	Invoke int
}

// Held is the definitions a frame holds as tools.
func (f Frame) Held() []DefLoc {
	var ds []DefLoc
	for _, m := range f.Bindings {
		if m.Obs == nil {
			ds = append(ds, m.Tool)
		}
	}
	return ds
}

// Arrive is a value as it arrives at a location: its interface its shape.
func Arrive(v any, loc InstLoc) Observation {
	return Observation{Value: v, Iface: ShapeIface(tree.ShapeOf(v)), Loc: loc}
}

// Equal is interface equality: by encoding, which is canonical.
func (i Iface) Equal(j Iface) bool { return compact(i) == compact(j) }

// Boot is the first frame of a process, or false when a check refuses:
// the program is well formed, the grant's interface is the entry scope
// position by position, every definition it passes fits an operator, and
// it only narrows, passing only what its booter holds. At genesis nothing
// holds tools, and what may be passed is open (61's question 8).
func Boot(ids Ids, p Program, g Grant, from *Booter) (Frame, bool) {
	var id InstLoc
	if from == nil {
		id = ids.First(ids.Program(p), g)
	} else {
		id = ids.Child(from.Frame.ID, from.Invoke, p.Entry)
	}
	var bindings []Member
	var ifaces []Iface
	for _, d := range g.Defs {
		bindings = append(bindings, Member{Tool: d})
		ifaces = append(ifaces, DefIface(d))
	}
	for i, v := range g.Context {
		o := Arrive(v, ids.Slot(id, len(g.Defs)+1+i))
		bindings = append(bindings, Member{Obs: &o})
		ifaces = append(ifaces, o.Iface)
	}
	if !p.WellFormed() || p.Entry.Kind != KindRoot ||
		!slices.EqualFunc(ifaces, p.Entry.Entries, Iface.Equal) {
		return Frame{}, false
	}
	models, tools := p.Reach("delegate"), p.Reach("act")
	for _, d := range g.Defs {
		if !slices.Contains(models, d) && !slices.Contains(tools, d) {
			return Frame{}, false
		}
		if from != nil && !slices.Contains(from.Frame.Held(), d) {
			return Frame{}, false
		}
	}
	f := Frame{ID: id, Program: ids.Program(p), Scope: p.Entry, Bindings: bindings}
	if from != nil {
		parent := from.Frame.ID
		f.Parent = &parent
	}
	return f, true
}
