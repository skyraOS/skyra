package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"go.starlark.net/syntax"

	"skyra/controller/core/contract"
)

// MaxSteps bounds one run of a route. It is the backstop: a route has no
// while, no recursion and no for, so it cannot loop on its own.
const MaxSteps = 100000

// Input is what the runtime hands the runner on stdin.
type Input struct {
	Source   string          `json:"source"`
	Return   json.RawMessage `json:"return"`
	Previous json.RawMessage `json:"previous"` // null: no earlier return of this type
	Targets  []string        `json:"targets"`  // the names a send may address: the parent's reach
}

// Decision is what a route decides, exactly one of three (next/30, with
// "wake" named interrupt: the runtime raises it, never the parent).
type Decision struct {
	Interrupt bool      `json:"interrupt,omitempty"` // deliver the return to the parent
	Send      *Envelope `json:"send,omitempty"`      // send an observation to a target in the parent's reach
	Fault     string    `json:"fault,omitempty"`     // refuse, carrying the return; also how a broken route is reported
}

type Envelope struct {
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
}

// decision is a Decision as a Starlark value: what route() must return.
type decision struct{ d Decision }

func (d *decision) String() string        { return "decision" }
func (d *decision) Type() string          { return "decision" }
func (d *decision) Freeze()               {}
func (d *decision) Truth() starlark.Bool  { return starlark.True }
func (d *decision) Hash() (uint32, error) { return 0, errors.New("unhashable: decision") }

// Run runs a route once over a return and decides. Every failure of the
// route itself (it does not compile, uses for, runs too long, decides twice
// or not at all) is a fault decision, never a crash.
func Run(in Input) Decision {
	d, err := run(in)
	if err != nil {
		return Decision{Fault: "route: " + err.Error()}
	}
	return d
}

func run(in Input) (Decision, error) {
	ret, err := contract.Decode(in.Return)
	if err != nil {
		return Decision{}, fmt.Errorf("return: %w", err)
	}
	var prev any
	hasPrev := len(in.Previous) > 0 && string(in.Previous) != "null"
	if hasPrev {
		if prev, err = contract.Decode(in.Previous); err != nil {
			return Decision{}, fmt.Errorf("previous: %w", err)
		}
	}

	// No while, no recursion, no top-level control (Starlark's defaults), and
	// no for or comprehensions: branching is if.
	opts := &syntax.FileOptions{}
	f, err := opts.Parse("route.star", in.Source, 0)
	if err != nil {
		return Decision{}, err
	}
	var loop error
	syntax.Walk(f, func(n syntax.Node) bool {
		switch n.(type) {
		case *syntax.ForStmt, *syntax.Comprehension:
			if loop == nil {
				start, _ := n.Span()
				loop = fmt.Errorf("%s: a route branches with if; for is not allowed", start)
			}
		}
		return loop == nil
	})
	if loop != nil {
		return Decision{}, loop
	}

	var decided []Decision
	decide := func(d Decision) (starlark.Value, error) {
		decided = append(decided, d)
		return &decision{d}, nil
	}
	pathArg := func(b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (string, error) {
		var path string
		err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &path)
		return path, err
	}
	skyra := &starlarkstruct.Module{Name: "skyra", Members: starlark.StringDict{
		"get": starlark.NewBuiltin("get", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			path, err := pathArg(b, args, kwargs)
			if err != nil {
				return nil, err
			}
			v, ok, err := at(ret, path)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, fmt.Errorf("get: the return has no %s", path)
			}
			return toStarlark(v)
		}),
		"previous": starlark.NewBuiltin("previous", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
				return nil, err
			}
			if !hasPrev {
				return starlark.None, nil
			}
			return toStarlark(prev)
		}),
		"changed": starlark.NewBuiltin("changed", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			path, err := pathArg(b, args, kwargs)
			if err != nil {
				return nil, err
			}
			now, inNow, err := at(ret, path)
			if err != nil {
				return nil, err
			}
			if !hasPrev {
				return starlark.True, nil // nothing to compare with: everything is new
			}
			was, inWas, _ := at(prev, path)
			if !inNow && !inWas {
				return nil, fmt.Errorf("changed: neither return has %s", path)
			}
			return starlark.Bool(inNow != inWas || contract.Content(now) != contract.Content(was)), nil
		}),
		"diff": starlark.NewBuiltin("diff", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
				return nil, err
			}
			if !hasPrev {
				if contract.Diff(ret, ret).Leaves == 0 {
					return starlark.Float(0), nil
				}
				return starlark.Float(1), nil // every leaf is new
			}
			return starlark.Float(contract.Diff(prev, ret).Ratio), nil
		}),
		"interrupt": starlark.NewBuiltin("interrupt", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 0); err != nil {
				return nil, err
			}
			return decide(Decision{Interrupt: true})
		}),
		"send": starlark.NewBuiltin("send", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var to string
			var payload starlark.Value
			if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 2, &to, &payload); err != nil {
				return nil, err
			}
			found := false
			for _, t := range in.Targets {
				found = found || t == to
			}
			if !found {
				return nil, fmt.Errorf("send: %q is not in the parent's reach", to)
			}
			v, err := fromStarlark(payload)
			if err != nil {
				return nil, fmt.Errorf("send: %w", err)
			}
			return decide(Decision{Send: &Envelope{To: to, Payload: contract.Canonical(v)}})
		}),
		"fault": starlark.NewBuiltin("fault", func(_ *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
			var cause string
			if err := starlark.UnpackPositionalArgs(b.Name(), args, kwargs, 1, &cause); err != nil {
				return nil, err
			}
			if cause == "" {
				return nil, errors.New("fault: a cause is required")
			}
			return decide(Decision{Fault: cause})
		}),
	}}
	predeclared := starlark.StringDict{"skyra": skyra}
	prog, err := starlark.FileProgram(f, predeclared.Has)
	if err != nil {
		return Decision{}, err
	}
	thread := &starlark.Thread{Name: "route"} // Load is nil: load() fails
	thread.SetMaxExecutionSteps(MaxSteps)
	globals, err := prog.Init(thread, predeclared)
	if err != nil {
		return Decision{}, err
	}
	fn, ok := globals["route"].(*starlark.Function)
	if !ok || fn.NumParams() != 0 {
		return Decision{}, errors.New("the source must define def route():")
	}
	decided = nil // a decision made while loading the file is not the route's
	got, err := starlark.Call(thread, fn, nil, nil)
	if err != nil {
		return Decision{}, err
	}
	d, ok := got.(*decision)
	switch {
	case !ok:
		return Decision{}, fmt.Errorf("route() returned %s, not a decision", got.Type())
	case len(decided) != 1:
		return Decision{}, fmt.Errorf("route() made %d decisions; a route returns exactly one", len(decided))
	}
	return d.d, nil
}

// at reads a path ($, .key, ["key"], [i]) in a decoded value. ok is false
// when the path does not exist; err is for a path that does not parse.
func at(v any, path string) (any, bool, error) {
	rest, found := strings.CutPrefix(path, "$")
	if !found {
		return nil, false, fmt.Errorf("path %q does not start with $", path)
	}
	for rest != "" {
		switch {
		case rest[0] == '.':
			end := 1
			for end < len(rest) && rest[end] != '.' && rest[end] != '[' {
				end++
			}
			key := rest[1:end]
			if key == "" {
				return nil, false, fmt.Errorf("path %q: empty key", path)
			}
			o, isObject := v.(map[string]any)
			if v, found = o[key]; !isObject || !found {
				return nil, false, nil
			}
			rest = rest[end:]
		case strings.HasPrefix(rest, `["`):
			d := json.NewDecoder(strings.NewReader(rest[1:]))
			var key string
			if err := d.Decode(&key); err != nil {
				return nil, false, fmt.Errorf("path %q: %w", path, err)
			}
			after := rest[1+int(d.InputOffset()):]
			if !strings.HasPrefix(after, "]") {
				return nil, false, fmt.Errorf("path %q: unclosed [", path)
			}
			o, isObject := v.(map[string]any)
			if v, found = o[key]; !isObject || !found {
				return nil, false, nil
			}
			rest = after[1:]
		case rest[0] == '[':
			end := strings.IndexByte(rest, ']')
			i, err := strconv.Atoi(rest[1:max(end, 1)])
			if end < 0 || err != nil || i < 0 {
				return nil, false, fmt.Errorf("path %q: bad index", path)
			}
			a, isArray := v.([]any)
			if !isArray || i >= len(a) {
				return nil, false, nil
			}
			v, rest = a[i], rest[end+1:]
		default:
			return nil, false, fmt.Errorf("path %q: expected . or [", path)
		}
	}
	return v, true, nil
}

// toStarlark converts a decoded value to a frozen Starlark value.
func toStarlark(v any) (starlark.Value, error) {
	switch x := v.(type) {
	case nil:
		return starlark.None, nil
	case bool:
		return starlark.Bool(x), nil
	case string:
		return starlark.String(x), nil
	case json.Number:
		if i, ok := new(big.Int).SetString(string(x), 10); ok {
			return starlark.MakeBigInt(i), nil
		}
		f, err := strconv.ParseFloat(string(x), 64)
		if err != nil {
			return nil, err
		}
		return starlark.Float(f), nil
	case []any:
		elems := make([]starlark.Value, len(x))
		for i, e := range x {
			var err error
			if elems[i], err = toStarlark(e); err != nil {
				return nil, err
			}
		}
		l := starlark.NewList(elems)
		l.Freeze()
		return l, nil
	case map[string]any:
		d := starlark.NewDict(len(x))
		for k, e := range x {
			sv, err := toStarlark(e)
			if err != nil {
				return nil, err
			}
			if err := d.SetKey(starlark.String(k), sv); err != nil {
				return nil, err
			}
		}
		d.Freeze()
		return d, nil
	}
	return nil, fmt.Errorf("cannot convert %T", v)
}

// fromStarlark converts a Starlark value to a JSON value a send can carry.
func fromStarlark(v starlark.Value) (any, error) {
	switch x := v.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(x), nil
	case starlark.String:
		return string(x), nil
	case starlark.Int:
		return json.Number(x.String()), nil
	case starlark.Float:
		f := float64(x)
		if math.IsInf(f, 0) || math.IsNaN(f) {
			return nil, errors.New("a float that is not finite is not JSON")
		}
		return json.Number(strconv.FormatFloat(f, 'g', -1, 64)), nil
	case *starlark.List:
		out := make([]any, x.Len())
		for i := range x.Len() {
			var err error
			if out[i], err = fromStarlark(x.Index(i)); err != nil {
				return nil, err
			}
		}
		return out, nil
	case starlark.Tuple:
		out := make([]any, len(x))
		for i, e := range x {
			var err error
			if out[i], err = fromStarlark(e); err != nil {
				return nil, err
			}
		}
		return out, nil
	case *starlark.Dict:
		out := map[string]any{}
		for _, kv := range x.Items() {
			k, ok := kv[0].(starlark.String)
			if !ok {
				return nil, fmt.Errorf("a dict key must be a string, not %s", kv[0].Type())
			}
			var err error
			if out[string(k)], err = fromStarlark(kv[1]); err != nil {
				return nil, err
			}
		}
		return out, nil
	}
	return nil, fmt.Errorf("a %s is not JSON", v.Type())
}
