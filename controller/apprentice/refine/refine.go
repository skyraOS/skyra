// Package refine is refinements in CEL (next/32, H9): predicates the model
// writes at remember, over the base types the apprentice declares, that say
// which inputs and returns a compiled body is valid for. Inside them the
// compiled body runs; outside them the original actor does.
//
// The environment is fixed and small: total, deterministic, with a cost
// bounded before it runs.
//
//	input        the procedure's input
//	input_type   its base type, as a model reads it
//	returns      what each node has returned so far, by its name (r0c0, r1c0, ...)
//	types        the base type of each, as a model reads it
//
// Only pure functions: CEL's standard library, and of its macros has, all,
// exists and exists_one. map and filter are left out: over a map they have no
// defined order, and a refinement is a predicate, which all and exists can say
// in any order. Time is an input, never a clock: CEL has no now(), and none is
// added. The records carry no timestamp yet (the record encoding does not
// change in H), so there is no invoked_at to give either.
package refine

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/checker"
	"cel.dev/cel-go/common/types"
)

// MaxCost bounds a refinement's estimated cost, and its cost at run time.
const MaxCost = 100_000

// sizeBound is how large the estimator takes any list, map or string to be.
const sizeBound = 1000

// Environment is what a model is told it may use.
var Environment = map[string]string{
	"input":      "the procedure's input, as JSON values",
	"input_type": "the input's base type, as text",
	"returns":    "map: each node's name → what it returned (only nodes that have returned)",
	"types":      "map: each node's name → the base type of what it returned, as text",
	"functions":  "CEL's standard library: comparisons, arithmetic, size, string functions, has(), all(), exists(), exists_one(); no map(), no filter(), no clock",
}

var env = func() *cel.Env {
	e, err := cel.NewEnv(
		cel.ClearMacros(), cel.Macros(cel.HasMacro, cel.AllMacro, cel.ExistsMacro, cel.ExistsOneMacro),
		cel.Variable("input", cel.DynType),
		cel.Variable("input_type", cel.StringType),
		cel.Variable("returns", cel.MapType(cel.StringType, cel.DynType)),
		cel.Variable("types", cel.MapType(cel.StringType, cel.StringType)),
	)
	if err != nil {
		panic(err)
	}
	return e
}()

// Program is a checked refinement.
type Program struct{ prg cel.Program }

// Compile checks a refinement: it parses, type-checks against the
// environment to a bool, and its estimated cost is bounded.
func Compile(src string) (Program, error) {
	ast, iss := env.Compile(src)
	if iss != nil && iss.Err() != nil {
		return Program{}, fmt.Errorf("does not check: %v", iss.Err())
	}
	if !ast.OutputType().IsExactType(types.BoolType) {
		return Program{}, fmt.Errorf("is %s, not a predicate (bool)", ast.OutputType())
	}
	est, err := env.EstimateCost(ast, bounded{})
	if err != nil {
		return Program{}, fmt.Errorf("its cost cannot be estimated: %v", err)
	}
	if est.Max > MaxCost {
		return Program{}, fmt.Errorf("may cost %d, more than %d", est.Max, MaxCost)
	}
	prg, err := env.Program(ast, cel.EvalOptions(cel.OptTrackState), cel.CostLimit(MaxCost))
	if err != nil {
		return Program{}, err
	}
	return Program{prg}, nil
}

// bounded is the estimator: anything whose size CEL cannot see is taken to
// be at most sizeBound.
type bounded struct{}

func (bounded) EstimateSize(checker.AstNode) *checker.SizeEstimate {
	return &checker.SizeEstimate{Min: 0, Max: sizeBound}
}
func (bounded) EstimateCallCost(string, string, *checker.AstNode, []checker.AstNode) *checker.CallEstimate {
	return nil
}

// Vars is what a refinement is evaluated over.
type Vars struct {
	Input     any
	InputType string
	Returns   map[string]any
	Types     map[string]string
}

// Eval evaluates a refinement. An error anywhere in it, even one && or ||
// absorbed into a result, is an error: a predicate reading a field that
// never exists must not pass for true.
func (p Program) Eval(v Vars) (bool, error) {
	returns, ts := v.Returns, v.Types
	if returns == nil {
		returns = map[string]any{}
	}
	if ts == nil {
		ts = map[string]string{}
	}
	out, det, err := p.prg.Eval(map[string]any{"input": plain(v.Input), "input_type": v.InputType, "returns": plain(returns), "types": ts})
	if err != nil {
		return false, err
	}
	if det != nil && det.State() != nil {
		st := det.State()
		ids := st.IDs()
		slices.Sort(ids)
		for _, id := range ids {
			if val, ok := st.Value(id); ok && types.IsError(val) {
				return false, fmt.Errorf("an error inside it: %v", val)
			}
		}
	}
	b, ok := out.Value().(bool)
	if !ok {
		return false, errors.New("not a bool")
	}
	return b, nil
}

// plain re-reads a decoded value with float numbers, which CEL takes.
func plain(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if json.Unmarshal(b, &out) != nil {
		return v
	}
	return out
}
