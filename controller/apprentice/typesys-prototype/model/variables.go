package model

// Variables, as Model.lean defines them. A frame's $n are its bindings, then
// its state, in order, from $1. The system instantiates the scope's and each
// arrival; the agent eliminates (cuts something held) or introduces
// (constructs from nothing).

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"skyra/typesys-prototype/tree"
)

// An ItemTag says what a $n holds.
type ItemTag uint8

const (
	ItemTool  ItemTag = iota // a tool, which has no value
	ItemSeen                 // a value seen, bound from the scope or arrived
	ItemIntro                // an introduction
	ItemJoin                 // an introduction joined from parts
	ItemElim                 // an elimination of an earlier $n
)

// An Item is what a $n holds.
type Item struct {
	Tag   ItemTag
	Tool  DefLoc
	Value any    // ItemSeen
	Lit   Lit    // ItemIntro
	Parts []Part // ItemJoin
	N     int    // ItemElim: the $n eliminated
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
		case e.Join != nil:
			is = append(is, Item{Tag: ItemJoin, Parts: e.Join})
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
	case ItemJoin:
		if len(sels) > 0 {
			return nil, false
		}
		s, ok := f.JoinText(it.Parts)
		return s, ok
	case ItemElim:
		if it.N >= n {
			return nil, false
		}
		return f.Read(it.N, append(slices.Clone(it.Sels), sels...))
	}
	return nil, false
}

// An Arg is an argument of a stored template: a reference, a $n cut by
// selectors or a global by name; or an introduction, a literal or a join.
// Pieces is a join as the model writes it, literals and $n, before
// JoinParts makes it; it is never stored.
type Arg struct {
	Var    int
	Sels   []Sel
	Name   string // a global, when not empty
	Lit    *Lit
	Join   []Part // a join, when not nil
	Pieces []Arg
}

// ReadPart is a part's text: as written, or the text a cut of a $n seen
// reads. A part reads nothing else.
func (f Frame) ReadPart(p Part) (string, bool) {
	if p.Str != nil {
		return *p.Str, true
	}
	it, ok := f.Item(p.Var)
	if !ok || it.Tag != ItemSeen {
		return "", false
	}
	v, ok := CutAll(it.Value, p.Sels)
	if !ok {
		return "", false
	}
	return textOf(v)
}

// JoinText is a join's text: its parts', in order, or none when one has
// none.
func (f Frame) JoinText(ps []Part) (string, bool) {
	var b strings.Builder
	for _, p := range ps {
		s, ok := f.ReadPart(p)
		if !ok {
			return "", false
		}
		b.WriteString(s)
	}
	return b.String(), true
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
	case ok && it.Tag == ItemJoin && len(sels) == 0:
		return Arg{Join: it.Parts}
	}
	return Arg{Var: n, Sels: sels}
}

// Eval is an argument's value in a frame. A global names a definition,
// which has no value.
func (f Frame) Eval(a Arg) (any, bool) {
	switch {
	case a.Lit != nil:
		return a.Lit.Value(), true
	case a.Name != "" || a.Pieces != nil:
		return nil, false
	case a.Join != nil:
		s, ok := f.JoinText(a.Join)
		return s, ok
	}
	return f.Read(a.Var, a.Sels)
}

// candidates are the cuts to try of a value: the whole; each run of its
// lines, when it is text; each field's cuts, through the field.
func candidates(v any) [][]Sel {
	out := [][]Sel{nil}
	switch x := v.(type) {
	case string, []byte:
		if s, ok := textOf(x); ok {
			n := len(tree.Lines(s))
			for i := range n {
				for j := range n - i {
					out = append(out, []Sel{{Lines: true, Runs: [][2]int{{i + 1, i + 1 + j}}}})
				}
			}
		}
	case map[string]any:
		for _, k := range slices.Sorted(maps.Keys(x)) {
			for _, c := range candidates(x[k]) {
				out = append(out, append([]Sel{{Path: []string{k}}}, c...))
			}
		}
	}
	return out
}

// Holds is whether a cut of $n, a value seen, reads exactly s.
func (f Frame) Holds(s string, n int) bool {
	it, ok := f.Item(n)
	if !ok || it.Tag != ItemSeen {
		return false
	}
	return slices.ContainsFunc(candidates(it.Value), func(sels []Sel) bool {
		v, ok := CutAll(it.Value, sels)
		t, tok := textOf(v)
		return ok && tok && t == s
	})
}

// Introduce is the entry an introduction appends, or the first $n it would
// copy: a string a cut of a held value reads is refused, with that $n,
// since the cut takes it instead. An exit status is never a copy.
func (f Frame) Introduce(l Lit) (Entry, int, bool) {
	if !l.IsExit {
		if n, ok := f.copies(l.Str); ok {
			return Entry{}, n, false
		}
	}
	return Entry{Introduce: &l}, 0, true
}

func (f Frame) copies(s string) (int, bool) {
	for n := 1; n <= len(f.Items()); n++ {
		if f.Holds(s, n) {
			return n, true
		}
	}
	return 0, false
}

// A JoinRefusal is why a join is refused: a written part a cut of $Copies
// takes, or a part, Untext, that reads no text.
type JoinRefusal struct {
	Copies int
	Untext *Arg
}

func (r JoinRefusal) Error() string {
	if r.Untext != nil {
		return fmt.Sprintf("%s reads no text: a join is made of text, written or cut", argSource(*r.Untext))
	}
	return fmt.Sprintf("a cut of $%d takes text the join writes: cut it instead of writing it", r.Copies)
}

// JoinParts is a join's parts, from what the model wrote, literals and $n:
// each literal checked as an introduction is, each $n inlined, a join it
// reads spliced in.
func (f Frame) JoinParts(as []Arg) ([]Part, error) {
	ps := []Part{}
	for _, a := range as {
		switch {
		case a.Lit != nil && !a.Lit.IsExit:
			if n, ok := f.copies(a.Lit.Str); ok {
				return nil, JoinRefusal{Copies: n}
			}
			s := a.Lit.Str
			ps = append(ps, Part{Str: &s})
		case a.Lit == nil && a.Name == "" && a.Join == nil && a.Pieces == nil:
			v, ok := f.Read(a.Var, a.Sels)
			if _, tok := textOf(v); !ok || !tok {
				return nil, JoinRefusal{Untext: &a}
			}
			switch in := f.Inline(a.Var, a.Sels); {
			case in.Lit != nil && !in.Lit.IsExit:
				s := in.Lit.Str
				ps = append(ps, Part{Str: &s})
			case in.Join != nil:
				ps = append(ps, in.Join...)
			case in.Lit == nil:
				ps = append(ps, Part{Var: in.Var, Sels: in.Sels})
			default:
				return nil, JoinRefusal{Untext: &a}
			}
		default:
			return nil, JoinRefusal{Untext: &a}
		}
	}
	return ps, nil
}

// argSource is an argument as the model writes it.
func argSource(a Arg) string {
	switch {
	case a.Lit != nil && a.Lit.IsExit:
		return strconv.Itoa(a.Lit.Exit)
	case a.Lit != nil:
		return "'" + a.Lit.Str + "'"
	case a.Name != "":
		return a.Name
	case a.Join != nil:
		var b strings.Builder
		for _, p := range a.Join {
			if p.Str != nil {
				b.WriteString("'" + *p.Str + "'")
			} else {
				b.WriteString(argSource(Arg{Var: p.Var, Sels: p.Sels}))
			}
		}
		return b.String()
	}
	s := "$" + strconv.Itoa(a.Var)
	for _, x := range a.Sels {
		s += x.Enc()
	}
	return s
}
