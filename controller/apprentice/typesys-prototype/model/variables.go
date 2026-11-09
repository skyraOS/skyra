package model

// Variables, as Model.lean defines them. A frame's $n are its bindings, then
// its state, in order, from $1. The system instantiates the scope's and each
// arrival; the agent eliminates (cuts something held) or introduces
// (constructs from nothing).

import (
	"slices"
	"strings"

	"skyra/typesys-prototype/tree"
)

// An ItemTag says what a $n holds.
type ItemTag uint8

const (
	ItemTool  ItemTag = iota // a tool, which has no value
	ItemSeen                 // a value seen, bound from the scope or arrived
	ItemIntro                // an introduction
	ItemElim                 // an elimination of an earlier $n
)

// An Item is what a $n holds.
type Item struct {
	Tag   ItemTag
	Tool  DefLoc
	Value any // ItemSeen
	Lit   Lit // ItemIntro
	N     int // ItemElim: the $n eliminated
	Sels  []Sel
}

// Items is a frame's $n in order: its bindings, then its state.
func (f Frame) Items() []Item {
	var is []Item
	for _, m := range f.Bindings {
		if m.Obs != nil {
			is = append(is, Item{Tag: ItemSeen, Value: m.Obs.Value})
		} else {
			is = append(is, Item{Tag: ItemTool, Tool: m.Tool})
		}
	}
	for _, e := range f.State {
		switch {
		case e.Arrived != nil:
			is = append(is, Item{Tag: ItemSeen, Value: e.Arrived.Value})
		case e.Introduce != nil:
			is = append(is, Item{Tag: ItemIntro, Lit: *e.Introduce})
		default:
			is = append(is, Item{Tag: ItemElim, N: e.Eliminate, Sels: e.Sels})
		}
	}
	return is
}

// Item is $n, from $1.
func (f Frame) Item(n int) (Item, bool) {
	is := f.Items()
	if n < 1 || n > len(is) {
		return Item{}, false
	}
	return is[n-1], true
}

// Value is what a literal constructs.
func (l Lit) Value() any {
	if l.IsExit {
		return tree.ExitStatus(l.Exit)
	}
	return l.Str
}

// Read is $n cut by sels. An elimination reads the earlier $n it
// eliminates; one of a later $n reads nothing. An introduction is never
// cut: what is constructed from nothing is constructed whole. A tool has no
// value.
func (f Frame) Read(n int, sels []Sel) (any, bool) {
	it, ok := f.Item(n)
	if !ok {
		return nil, false
	}
	switch it.Tag {
	case ItemSeen:
		return CutAll(it.Value, sels)
	case ItemIntro:
		if len(sels) > 0 {
			return nil, false
		}
		return it.Lit.Value(), true
	case ItemElim:
		if it.N >= n {
			return nil, false
		}
		return f.Read(it.N, append(slices.Clone(it.Sels), sels...))
	}
	return nil, false
}

// An Arg is an argument of a stored template: a reference, a $n cut by
// selectors or a global by name; or an introduction's literal.
type Arg struct {
	Var  int
	Sels []Sel
	Name string // a global, when not empty
	Lit  *Lit
}

// Inline is $n cut by sels as it is stored at commit: each elimination
// replaced by its definition and each introduction by its literal, so a
// stored implementation reads only what was seen.
func (f Frame) Inline(n int, sels []Sel) Arg {
	it, ok := f.Item(n)
	switch {
	case ok && it.Tag == ItemElim && it.N < n:
		return f.Inline(it.N, append(slices.Clone(it.Sels), sels...))
	case ok && it.Tag == ItemIntro && len(sels) == 0:
		l := it.Lit
		return Arg{Lit: &l}
	}
	return Arg{Var: n, Sels: sels}
}

// Eval is an argument's value in a frame. A global names a definition,
// which has no value.
func (f Frame) Eval(a Arg) (any, bool) {
	switch {
	case a.Lit != nil:
		return a.Lit.Value(), true
	case a.Name != "":
		return nil, false
	}
	return f.Read(a.Var, a.Sels)
}

// valueStrings is every string a value holds, at any depth, keys included.
func valueStrings(v any) []string {
	switch v := v.(type) {
	case string:
		return []string{v}
	case []any:
		var ss []string
		for _, x := range v {
			ss = append(ss, valueStrings(x)...)
		}
		return ss
	case map[string]any:
		var ss []string
		for k, x := range v {
			ss = append(ss, k)
			ss = append(ss, valueStrings(x)...)
		}
		return ss
	}
	return nil
}

// Holds is whether $n holds a value with s in its text.
func (f Frame) Holds(s string, n int) bool {
	it, ok := f.Item(n)
	if !ok || it.Tag != ItemSeen {
		return false
	}
	ts := valueStrings(it.Value)
	if t, ok := textOf(it.Value); ok {
		ts = append(ts, t)
	}
	return slices.ContainsFunc(ts, func(t string) bool { return strings.Contains(t, s) })
}

// Introduce is the entry an introduction appends, or the first $n it would
// copy: a string a held value contains is refused, with that $n, since the
// cut takes it instead. An exit status is never a copy.
func (f Frame) Introduce(l Lit) (Entry, int, bool) {
	if !l.IsExit {
		for n := 1; n <= len(f.Items()); n++ {
			if f.Holds(l.Str, n) {
				return Entry{}, n, false
			}
		}
	}
	return Entry{Introduce: &l}, 0, true
}
