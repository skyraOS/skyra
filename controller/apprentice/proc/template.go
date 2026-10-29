package proc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
)

// Templates. A node's call, its in and its each, and a procedure's return,
// are JSON in which a binding stands for a value found by provenance (32,
// "Bindings by provenance"):
//
//	{"$": ["input", "doc"]}       a parameter: a field of what the procedure was invoked with
//	{"$": ["<node>", "parts"]}    a binding: a field of what an earlier node returned
//	{"$": ["item"]}               in a node with each: the element it is called for
//	{"$const": <JSON>}            a constant, written so only when it would read as a binding
//
// Everything else is a constant: the procedure uses it on every run.

// Env is what a template is evaluated in.
type Env struct {
	Input   any
	Values  map[string]any // node → what it returned (a node with each: every element's, in order)
	Item    any
	HasItem bool
}

// Eval evaluates a template.
func Eval(t json.RawMessage, env Env) (any, error) {
	v, err := decode(t)
	if err != nil {
		return nil, err
	}
	return eval(v, env)
}

func decode(b []byte) (any, error) {
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	var v any
	err := d.Decode(&v)
	return v, err
}

func eval(v any, env Env) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		if path, ok := binding(x); ok {
			return resolve(path, env)
		}
		if c, ok := x["$const"]; ok && len(x) == 1 {
			return c, nil
		}
		out := make(map[string]any, len(x))
		for k, e := range x {
			r, err := eval(e, env)
			if err != nil {
				return nil, err
			}
			out[k] = r
		}
		return out, nil
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			r, err := eval(e, env)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil
	}
	return v, nil
}

// binding reads {"$": [segment, ...]}: the first segment a string, the rest
// object keys or array indexes.
func binding(o map[string]any) ([]any, bool) {
	p, ok := o["$"].([]any)
	if !ok || len(o) != 1 || len(p) == 0 {
		return nil, false
	}
	if _, ok := p[0].(string); !ok {
		return nil, false
	}
	return p, true
}

// Binding is a template that stands for one value: a path from a root.
func Binding(root string, path ...any) map[string]any {
	return map[string]any{"$": append([]any{root}, path...)}
}

// Const writes a constant so it cannot read as a binding.
func Const(v any) any {
	if o, ok := v.(map[string]any); ok && len(o) == 1 {
		if _, b := o["$"]; b {
			return map[string]any{"$const": v}
		}
		if _, c := o["$const"]; c {
			return map[string]any{"$const": v}
		}
	}
	return v
}

func resolve(path []any, env Env) (any, error) {
	var v any
	switch root := path[0].(string); root {
	case "input":
		v = env.Input
	case "item":
		if !env.HasItem {
			return nil, fmt.Errorf("item outside a node with each")
		}
		v = env.Item
	default:
		got, ok := env.Values[root]
		if !ok {
			return nil, fmt.Errorf("binding to %s, which has not returned", short(root))
		}
		v = got
	}
	for _, seg := range path[1:] {
		switch s := seg.(type) {
		case string:
			o, ok := v.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("binding: %q of a value that is not an object", s)
			}
			if v, ok = o[s]; !ok {
				return nil, fmt.Errorf("binding: no field %q", s)
			}
		case json.Number:
			i, err := strconv.Atoi(s.String())
			a, ok := v.([]any)
			if err != nil || !ok || i < 0 || i >= len(a) {
				return nil, fmt.Errorf("binding: no element %s", s)
			}
			v = a[i]
		default:
			return nil, fmt.Errorf("binding: a segment is a key or an index")
		}
	}
	return v, nil
}

// Refs lists the nodes a template binds to, sorted.
func Refs(t json.RawMessage) []string {
	v, err := decode(t)
	if err != nil {
		return nil
	}
	var refs []string
	var walk func(v any)
	walk = func(v any) {
		switch x := v.(type) {
		case map[string]any:
			if p, ok := binding(x); ok {
				if r := p[0].(string); r != "input" && r != "item" && !slices.Contains(refs, r) {
					refs = append(refs, r)
				}
				return
			}
			if _, ok := x["$const"]; ok && len(x) == 1 {
				return
			}
			for _, e := range x {
				walk(e)
			}
		case []any:
			for _, e := range x {
				walk(e)
			}
		}
	}
	walk(v)
	slices.Sort(refs)
	return refs
}

// rename rewrites the roots of a template's bindings.
func rename(t json.RawMessage, names map[string]string) (json.RawMessage, error) {
	if len(t) == 0 {
		return nil, nil
	}
	v, err := decode(t)
	if err != nil {
		return nil, err
	}
	var walk func(v any) any
	walk = func(v any) any {
		switch x := v.(type) {
		case map[string]any:
			if p, ok := binding(x); ok {
				if id, ok := names[p[0].(string)]; ok {
					p = append([]any{id}, p[1:]...)
				}
				return map[string]any{"$": p}
			}
			if _, ok := x["$const"]; ok && len(x) == 1 {
				return x
			}
			out := map[string]any{}
			for k, e := range x {
				out[k] = walk(e)
			}
			return out
		case []any:
			out := make([]any, len(x))
			for i, e := range x {
				out[i] = walk(e)
			}
			return out
		}
		return v
	}
	return json.Marshal(walk(v))
}

func short(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
