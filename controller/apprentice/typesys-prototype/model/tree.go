package model

// The tree, as Model.lean defines it. Every decision is one edge: an invoke
// or a boot pushes one frame, a return pops one. A frame decides only while
// open and with nothing pending; what it pushes is the kind its operator
// pushes; and a return lands its observation in exactly the frame its
// dynamic link names.

import (
	"encoding/json"
	"fmt"
	"slices"
)

// A Tree is every frame, linked by dynamic links, and the frames that have
// returned, latest first.
type Tree struct {
	Frames []Frame   `json:"frames"`
	Closed []InstLoc `json:"closed"`
}

// An EdgeTag says which of the three edges a decision takes.
type EdgeTag uint8

const (
	EdgeInvoke EdgeTag = iota // pushes a frame of Scope by Op, under the decider's program
	EdgeBoot                  // pushes the frame Boot gave, under the program passed
	EdgeReturn                // pops, landing Obs in the decider's parent
)

// An Edge is a decision's edge.
type Edge struct {
	Tag      EdgeTag
	Op       Op
	Scope    Scope
	Bindings []Member
	Model    *DefLoc // the model a delegation chose
	Boot     Frame
	Obs      Observation
}

// Pushes is direction as the edge's form, as it is the operator's.
func (e Edge) Pushes() bool { return e.Tag != EdgeReturn }

// Reaches gives what a program's operators reach, by the program's
// location and the operator's name: Program.Reach, through a frame's static
// link.
type Reaches func(program DefLoc, name string) []DefLoc

// Find is the first frame with id.
func (t Tree) Find(id InstLoc) (Frame, bool) {
	for _, f := range t.Frames {
		if f.ID == id {
			return f, true
		}
	}
	return Frame{}, false
}

// IsOpen is whether a frame has not returned.
func (t Tree) IsOpen(id InstLoc) bool { return !slices.Contains(t.Closed, id) }

// Pending is whether a frame waits: a frame it invoked has not returned.
func (t Tree) Pending(id InstLoc) bool {
	return slices.ContainsFunc(t.Frames, func(c Frame) bool {
		return c.Parent != nil && *c.Parent == id && t.IsOpen(c.ID)
	})
}

// Invokes is how many frames a frame has invoked: its next invoke's index.
func (t Tree) Invokes(id InstLoc) int {
	n := 0
	for _, c := range t.Frames {
		if c.Parent != nil && *c.Parent == id {
			n++
		}
	}
	return n
}

// Push is invoke: a frame pushes its nth child, of scope, bound to
// bindings, consulting model. Its static link is its parent's program, its
// dynamic link its parent, and it starts with no state.
func Push(ids Ids, parent Frame, n int, s Scope, bindings []Member, model *DefLoc) Frame {
	p := parent.ID
	return Frame{ID: ids.Child(parent.ID, n, s), Program: parent.Program, Parent: &p, Scope: s,
		Bindings: bindings, Model: model}
}

// modelOK is whether an invoke names its model rightly: a delegation one
// that fits delegate's first position, any other invoke none.
func modelOK(models []DefLoc, o Op, m *DefLoc) bool {
	if o == OpDelegate {
		return m != nil && slices.Contains(models, *m)
	}
	return m == nil
}

func subset(xs, ys []DefLoc) bool {
	for _, x := range xs {
		if !slices.Contains(ys, x) {
			return false
		}
	}
	return true
}

// Admits is what the deciding frame may emit: an operator its kind emits,
// pushing the kind that operator pushes, a delegation passing a subset of
// its definitions, narrowing tool types and naming one of its program's
// models; a boot only from a kind that boots, of a root linked to the
// decider.
func Admits(k int, reaches Reaches, f Frame, e Edge) bool {
	switch e.Tag {
	case EdgeInvoke:
		act := reaches(f.Program, "act")
		kind, ok := Pushed(k, act, e.Op, e.Scope.Entries)
		if !modelOK(reaches(f.Program, "delegate"), e.Op, e.Model) || e.Op == OpBoot ||
			!slices.Contains(Ops(f.Scope.Kind), e.Op) || !ok || kind != e.Scope.Kind {
			return false
		}
		return e.Op != OpDelegate || subset(Defs(e.Scope.Entries), Defs(f.Scope.Entries)) &&
			len(ToolTypes(act, e.Scope.Entries)) < len(ToolTypes(act, f.Scope.Entries))
	case EdgeBoot:
		return slices.Contains(Ops(f.Scope.Kind), OpBoot) && e.Boot.Parent != nil &&
			*e.Boot.Parent == f.ID && e.Boot.Scope.Kind == KindRoot
	}
	return true
}

// Land is where a return lands: the frame its dynamic link names, with the
// observation appended as an arrival. No other frame changes.
func (t Tree) Land(p InstLoc, o Observation) []Frame {
	fs := make([]Frame, len(t.Frames))
	for i, g := range t.Frames {
		if g.ID == p {
			g.State = append(slices.Clone(g.State), Entry{Arrived: &o})
		}
		fs[i] = g
	}
	return fs
}

// Step is one decision's edge from the frame fid, or false when it is
// refused: the frame is unknown, closed, pending, or the edge not admitted;
// or a return has no parent in the tree.
func (t Tree) Step(ids Ids, k int, reaches Reaches, fid InstLoc, e Edge) (Tree, bool) {
	f, ok := t.Find(fid)
	if !ok || !t.IsOpen(fid) || t.Pending(fid) || !Admits(k, reaches, f, e) {
		return Tree{}, false
	}
	closed := slices.Clone(t.Closed)
	switch e.Tag {
	case EdgeInvoke:
		c := Push(ids, f, t.Invokes(fid), e.Scope, e.Bindings, e.Model)
		return Tree{Frames: append(slices.Clone(t.Frames), c), Closed: closed}, true
	case EdgeBoot:
		return Tree{Frames: append(slices.Clone(t.Frames), e.Boot), Closed: closed}, true
	}
	if f.Parent == nil {
		return Tree{}, false
	}
	if _, ok := t.Find(*f.Parent); !ok {
		return Tree{}, false
	}
	return Tree{Frames: t.Land(*f.Parent, e.Obs), Closed: append([]InstLoc{fid}, closed...)}, true
}

// Replay is the tree as a fold of its edges, or false at the first refused.
func (t Tree) Replay(ids Ids, k int, reaches Reaches, log []LogEdge) (Tree, bool) {
	for _, x := range log {
		var ok bool
		if t, ok = t.Step(ids, k, reaches, x.Frame, x.Edge); !ok {
			return Tree{}, false
		}
	}
	return t, true
}

// A LogEdge is one edge in the log: the frame that decided, and its edge.
type LogEdge struct {
	Frame InstLoc
	Edge  Edge
}

func (t Tree) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"frames": nonNil(t.Frames), "closed": nonNil(t.Closed)})
}

func (e Edge) MarshalJSON() ([]byte, error) {
	switch e.Tag {
	case EdgeInvoke:
		return json.Marshal(map[string]any{"invoke": e.Op, "scope": e.Scope,
			"bindings": nonNil(e.Bindings), "model": e.Model})
	case EdgeBoot:
		return json.Marshal(map[string]any{"boot": e.Boot})
	}
	return json.Marshal(map[string]any{"return": e.Obs})
}

func (e *Edge) UnmarshalJSON(b []byte) error {
	var x struct {
		Invoke   *Op          `json:"invoke"`
		Scope    Scope        `json:"scope"`
		Bindings []Member     `json:"bindings"`
		Model    *DefLoc      `json:"model"`
		Boot     *Frame       `json:"boot"`
		Return   *Observation `json:"return"`
	}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	switch {
	case x.Invoke != nil:
		*e = Edge{Tag: EdgeInvoke, Op: *x.Invoke, Scope: x.Scope, Bindings: x.Bindings, Model: x.Model}
	case x.Boot != nil:
		*e = Edge{Tag: EdgeBoot, Boot: *x.Boot}
	case x.Return != nil:
		*e = Edge{Tag: EdgeReturn, Obs: *x.Return}
	default:
		return fmt.Errorf("edge %s", b)
	}
	return nil
}
