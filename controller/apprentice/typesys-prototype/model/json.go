package model

// The JSON of 61's types, as Vectors.lean writes it. A shape is its
// canonical encoding; a selector is as written after $n; an interface is an
// object naming how it is built.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"skyra/typesys-prototype/tree"
)

// Enc is a selector as it is written after $n: .a.b, or [2:3,5:5].
func (s Sel) Enc() string {
	if !s.Lines {
		var b strings.Builder
		for _, k := range s.Path {
			b.WriteString("." + k)
		}
		return b.String()
	}
	parts := make([]string, len(s.Runs))
	for i, r := range s.Runs {
		parts[i] = strconv.Itoa(r[0]) + ":" + strconv.Itoa(r[1])
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// ParseSel reads a selector's encoding.
func ParseSel(enc string) (Sel, error) {
	if !strings.HasPrefix(enc, "[") {
		if enc == "" {
			return Sel{}, nil
		}
		if !strings.HasPrefix(enc, ".") {
			return Sel{}, fmt.Errorf("selector %q", enc)
		}
		return Sel{Path: strings.Split(enc[1:], ".")}, nil
	}
	if !strings.HasSuffix(enc, "]") {
		return Sel{}, fmt.Errorf("selector %q", enc)
	}
	s := Sel{Lines: true, Runs: [][2]int{}}
	body := enc[1 : len(enc)-1]
	if body == "" {
		return s, nil
	}
	for _, run := range strings.Split(body, ",") {
		a, b, ok := strings.Cut(run, ":")
		x, err1 := strconv.Atoi(a)
		y, err2 := strconv.Atoi(b)
		if !ok || err1 != nil || err2 != nil {
			return Sel{}, fmt.Errorf("selector %q", enc)
		}
		s.Runs = append(s.Runs, [2]int{x, y})
	}
	return s, nil
}

func (s Sel) MarshalJSON() ([]byte, error) { return json.Marshal(s.Enc()) }

func (s *Sel) UnmarshalJSON(b []byte) error {
	var enc string
	if err := json.Unmarshal(b, &enc); err != nil {
		return err
	}
	var err error
	*s, err = ParseSel(enc)
	return err
}

// ParseShape reads a shape's canonical encoding, tree.Shape.Enc's inverse.
func ParseShape(enc string) (tree.Shape, error) {
	p := shapeParser{s: enc}
	sh, err := p.shape()
	if err == nil && p.i != len(p.s) {
		err = fmt.Errorf("shape %q: trailing %q", enc, p.s[p.i:])
	}
	return sh, err
}

type shapeParser struct {
	s string
	i int
}

func (p *shapeParser) eat(lit string) bool {
	if strings.HasPrefix(p.s[p.i:], lit) {
		p.i += len(lit)
		return true
	}
	return false
}

func (p *shapeParser) shape() (tree.Shape, error) {
	for _, k := range []struct {
		lit  string
		kind tree.ShapeKind
	}{{"null", tree.ShapeNull}, {"boolean", tree.ShapeBool}, {"number", tree.ShapeNum},
		{"string", tree.ShapeStr}, {"bytes", tree.ShapeBytes}} {
		if p.eat(k.lit) {
			return tree.Shape{Kind: k.kind}, nil
		}
	}
	switch {
	case p.eat("exit("):
		j := strings.IndexByte(p.s[p.i:], ')')
		if j < 0 {
			return tree.Shape{}, fmt.Errorf("shape %q: exit", p.s)
		}
		n, err := strconv.Atoi(p.s[p.i : p.i+j])
		p.i += j + 1
		return tree.Shape{Kind: tree.ShapeExit, Status: n}, err
	case p.eat("["):
		sh := tree.Shape{Kind: tree.ShapeArr}
		for !p.eat("]") {
			if len(sh.Elems) > 0 && !p.eat(",") {
				return tree.Shape{}, fmt.Errorf("shape %q: array at %d", p.s, p.i)
			}
			e, err := p.shape()
			if err != nil {
				return tree.Shape{}, err
			}
			sh.Elems = append(sh.Elems, e)
		}
		return sh, nil
	case p.eat("{"):
		sh := tree.Shape{Kind: tree.ShapeObj, Fields: []tree.Field{}}
		for !p.eat("}") {
			if len(sh.Fields) > 0 && !p.eat(",") {
				return tree.Shape{}, fmt.Errorf("shape %q: object at %d", p.s, p.i)
			}
			k, err := p.key()
			if err != nil {
				return tree.Shape{}, err
			}
			if !p.eat(":") {
				return tree.Shape{}, fmt.Errorf("shape %q: key at %d", p.s, p.i)
			}
			v, err := p.shape()
			if err != nil {
				return tree.Shape{}, err
			}
			sh.Fields = append(sh.Fields, tree.Field{Key: k, Shape: v})
		}
		return sh, nil
	}
	return tree.Shape{}, fmt.Errorf("shape %q at %d", p.s, p.i)
}

// key reads a key as tree's escape writes it: `"` and `\` escaped, control
// characters as \u00xx.
func (p *shapeParser) key() (string, error) {
	if !p.eat(`"`) {
		return "", fmt.Errorf("shape %q: key at %d", p.s, p.i)
	}
	var b strings.Builder
	for p.i < len(p.s) {
		c := p.s[p.i]
		switch {
		case c == '"':
			p.i++
			return b.String(), nil
		case c == '\\' && p.i+1 < len(p.s) && p.s[p.i+1] == 'u' && p.i+6 <= len(p.s):
			n, err := strconv.ParseUint(p.s[p.i+2:p.i+6], 16, 8)
			if err != nil {
				return "", err
			}
			b.WriteByte(byte(n))
			p.i += 6
		case c == '\\' && p.i+1 < len(p.s):
			b.WriteByte(p.s[p.i+1])
			p.i += 2
		default:
			b.WriteByte(c)
			p.i++
		}
	}
	return "", fmt.Errorf("shape %q: unterminated key", p.s)
}

func (i Iface) MarshalJSON() ([]byte, error) {
	switch i.Tag {
	case IShape:
		return json.Marshal(map[string]any{"shape": i.Shape.Enc()})
	case IDef:
		return json.Marshal(map[string]any{"def": i.Def})
	case IOp:
		return json.Marshal(map[string]any{"op": i.Op, "parts": nonNil(i.Parts)})
	case ISum:
		return json.Marshal(map[string]any{"sum": nonNil(i.Parts)})
	}
	return json.Marshal(map[string]any{"cut": i.Of, "sel": i.Sel})
}

func (i *Iface) UnmarshalJSON(b []byte) error {
	var o struct {
		Shape *string `json:"shape"`
		Def   *DefLoc `json:"def"`
		Op    *Op     `json:"op"`
		Parts []Iface `json:"parts"`
		Sum   []Iface `json:"sum"`
		Cut   *Iface  `json:"cut"`
		Sel   Sel     `json:"sel"`
	}
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}
	switch {
	case o.Shape != nil:
		s, err := ParseShape(*o.Shape)
		*i = ShapeIface(s)
		return err
	case o.Def != nil:
		*i = DefIface(*o.Def)
	case o.Op != nil:
		*i = Iface{Tag: IOp, Op: *o.Op, Parts: o.Parts}
	case o.Sum != nil:
		*i = Iface{Tag: ISum, Parts: o.Sum}
	case o.Cut != nil:
		*i = Iface{Tag: ICut, Of: o.Cut, Sel: o.Sel}
	default:
		return fmt.Errorf("interface %s", b)
	}
	return nil
}

func (s Scope) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"kind": s.Kind, "entries": nonNil(s.Entries)})
}

func (s *Scope) UnmarshalJSON(b []byte) error {
	var o struct {
		Kind    Kind    `json:"kind"`
		Entries []Iface `json:"entries"`
	}
	err := json.Unmarshal(b, &o)
	*s = Scope{Kind: o.Kind, Entries: o.Entries}
	return err
}

// Text is a scope's compact JSON, keys sorted, nothing escaped that JSON
// does not require: the text a vector's child id holds.
func (s Scope) Text() string { return compact(s) }

func compact(v any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		panic(err)
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func (l Lit) MarshalJSON() ([]byte, error) {
	if l.IsExit {
		return json.Marshal(map[string]any{"exit": l.Exit})
	}
	return json.Marshal(map[string]any{"str": l.Str})
}

func (l *Lit) UnmarshalJSON(b []byte) error {
	var o struct {
		Str  *string `json:"str"`
		Exit *int    `json:"exit"`
	}
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}
	switch {
	case o.Str != nil:
		*l = Lit{Str: *o.Str}
	case o.Exit != nil:
		*l = Lit{IsExit: true, Exit: *o.Exit}
	default:
		return fmt.Errorf("literal %s", b)
	}
	return nil
}

func (o Observation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"value": o.Value, "iface": o.Iface, "loc": o.Loc})
}

func (o *Observation) UnmarshalJSON(b []byte) error {
	var x struct {
		Value any     `json:"value"`
		Iface Iface   `json:"iface"`
		Loc   InstLoc `json:"loc"`
	}
	err := json.Unmarshal(b, &x)
	*o = Observation{Value: x.Value, Iface: x.Iface, Loc: x.Loc}
	return err
}

func (m Member) MarshalJSON() ([]byte, error) {
	if m.Obs != nil {
		return json.Marshal(map[string]any{"obs": m.Obs})
	}
	return json.Marshal(map[string]any{"tool": m.Tool})
}

func (m *Member) UnmarshalJSON(b []byte) error {
	var x struct {
		Tool *DefLoc      `json:"tool"`
		Obs  *Observation `json:"obs"`
	}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	switch {
	case x.Obs != nil:
		*m = Member{Obs: x.Obs}
	case x.Tool != nil:
		*m = Member{Tool: *x.Tool}
	default:
		return fmt.Errorf("member %s", b)
	}
	return nil
}

func (e Entry) MarshalJSON() ([]byte, error) {
	switch {
	case e.Arrived != nil:
		return json.Marshal(map[string]any{"arrived": e.Arrived})
	case e.Introduce != nil:
		return json.Marshal(map[string]any{"introduce": e.Introduce})
	}
	return json.Marshal(map[string]any{"eliminate": e.Eliminate, "sels": nonNil(e.Sels)})
}

func (e *Entry) UnmarshalJSON(b []byte) error {
	var x struct {
		Arrived   *Observation `json:"arrived"`
		Eliminate *int         `json:"eliminate"`
		Sels      []Sel        `json:"sels"`
		Introduce *Lit         `json:"introduce"`
	}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	switch {
	case x.Arrived != nil:
		*e = Entry{Arrived: x.Arrived}
	case x.Introduce != nil:
		*e = Entry{Introduce: x.Introduce}
	case x.Eliminate != nil:
		*e = Entry{Eliminate: *x.Eliminate, Sels: x.Sels}
	default:
		return fmt.Errorf("entry %s", b)
	}
	return nil
}

func (f Frame) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"id": f.ID, "program": f.Program, "parent": f.Parent,
		"scope": f.Scope, "bindings": nonNil(f.Bindings), "state": nonNil(f.State), "model": f.Model})
}

func (f *Frame) UnmarshalJSON(b []byte) error {
	var x struct {
		ID       InstLoc  `json:"id"`
		Program  DefLoc   `json:"program"`
		Parent   *InstLoc `json:"parent"`
		Scope    Scope    `json:"scope"`
		Bindings []Member `json:"bindings"`
		State    []Entry  `json:"state"`
		Model    *DefLoc  `json:"model"`
	}
	err := json.Unmarshal(b, &x)
	*f = Frame{ID: x.ID, Program: x.Program, Parent: x.Parent, Scope: x.Scope,
		Bindings: x.Bindings, State: x.State, Model: x.Model}
	return err
}

func (k Key) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"root": k.Root, "kind": k.Kind, "seen": nonNil(k.Seen)})
}

// nonNil writes an empty list as [], as Lean does, never as null.
func nonNil[T any](xs []T) []T {
	if xs == nil {
		return []T{}
	}
	return xs
}
