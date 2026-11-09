package model

// Cuts, as Model.lean defines them. A cut divides an interface by a
// selector. The key keeps I / sel undivided; reducing it to a shape is a
// function, and agrees with the shape of the value the cut takes. Cutting a
// value is 58's (tree.Cut).

import (
	"unicode/utf8"

	"skyra/typesys-prototype/tree"
)

// ToCut is the selector as 58 cuts with it.
func (s Sel) ToCut() tree.Sel {
	if s.Lines {
		return tree.Sel{Kind: tree.SelLines, Runs: s.Runs}
	}
	return tree.Sel{Kind: tree.SelField, Path: s.Path}
}

// CutAll is a value cut by selectors in order.
func CutAll(v any, sels []Sel) (any, bool) {
	for _, s := range sels {
		var ok bool
		if v, ok = tree.Cut(v, s.ToCut()); !ok {
			return nil, false
		}
	}
	return v, true
}

// FieldShape is a field path through a shape.
func FieldShape(s tree.Shape, path []string) (tree.Shape, bool) {
	for _, k := range path {
		if s.Kind != tree.ShapeObj {
			return tree.Shape{}, false
		}
		found := false
		for _, f := range s.Fields {
			if f.Key == k {
				s, found = f.Shape, true
				break
			}
		}
		if !found {
			return tree.Shape{}, false
		}
	}
	return s, true
}

// ReduceShape is a shape divided by a selector: a field path's shape, or
// text's lines, which are text. An empty selector divides nothing, as
// tree.Cut refuses it.
func ReduceShape(sel Sel, s tree.Shape) (tree.Shape, bool) {
	if !sel.Lines {
		if len(sel.Path) == 0 {
			return tree.Shape{}, false
		}
		return FieldShape(s, sel.Path)
	}
	if len(sel.Runs) == 0 || s.Kind != tree.ShapeStr && s.Kind != tree.ShapeBytes {
		return tree.Shape{}, false
	}
	return tree.Shape{Kind: tree.ShapeStr}, true
}

// CutIface is an interface divided by selectors in order, kept undivided.
func CutIface(i Iface, sels []Sel) Iface {
	for _, s := range sels {
		of := i
		i = Iface{Tag: ICut, Of: &of, Sel: s}
	}
	return i
}

// Reduce is an interface with its cuts divided: a cut of a shape is that
// shape divided, a cut of a cut is divided in turn, and nothing else is a
// cut.
func (i Iface) Reduce() (Iface, bool) {
	if i.Tag != ICut {
		return i, true
	}
	of, ok := i.Of.Reduce()
	if !ok || of.Tag != IShape {
		return Iface{}, false
	}
	s, ok := ReduceShape(i.Sel, of.Shape)
	return ShapeIface(s), ok
}

// textOf is a value's text: a string, or bytes that are UTF-8.
func textOf(v any) (string, bool) {
	switch v := v.(type) {
	case string:
		return v, true
	case []byte:
		if utf8.Valid(v) {
			return string(v), true
		}
	}
	return "", false
}
