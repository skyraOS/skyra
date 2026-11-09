// Proposals 58 and 59: one base's versioned refinement, as Refine.lean
// defines it. A version holds branches, each a CEL condition selecting an
// instruction. The model writes a guard, a chain `c₀ ? "i₀" : c₁ ? "i₁" :
// null`, and never sees the existing branches. The chain is stored whole,
// as branches under one voter; when values arrive, the votes are tallied
// (tally.go). A transition is one instruction's life, ending in trust; a
// frame takes at most its transition bound of them, and an exhausted frame
// returns an error observation.
//
// Everything here is pure. Which version is current, the store, and the
// scope table are not in this file.
package tree

import (
	"errors"
	"slices"
	"strings"

	"cel.dev/cel-go/cel"
	"cel.dev/cel-go/common/ast"
	"cel.dev/cel-go/common/operators"
	"cel.dev/cel-go/common/types"
	"cel.dev/cel-go/ext"
)

// --- branches and versions ---

// A Branch is one condition and the instruction it selects. ID is its local
// id, unique only within its base's refinement; Rank is the local id it was
// first committed under, which orders it: a lower rank is older; Voter is
// the local id its chain was committed under (59).
type Branch struct {
	ID    int    `json:"id"`
	Rank  int    `json:"rank"`
	Voter int    `json:"voter"`
	Cond  string `json:"cond"`  // CEL: the effective condition
	Instr string `json:"instr"` // the instruction template it selects
}

// A Version is one version of a base's refinement (Refine.Refinement): its
// branches, unordered, and the next local id. Ids start at 1 (b1).
type Version struct {
	Next     int      `json:"next"`
	Branches []Branch `json:"branches"`
}

// EmptyVersion is a new base's refinement: no branches, so every arrival
// is null.
func EmptyVersion() Version { return Version{Next: 1, Branches: []Branch{}} }

// A Link is one condition of the model's chain and its instruction, as
// written: not yet effective.
type Link struct {
	Cond  string
	Instr string
}

// The condition forms the Lean names, written as CEL. The Go writes exactly
// these strings, so a branch's condition is the same text on both sides.
func and(p, q string) string { return "(" + p + ") && (" + q + ")" }
func not(p string) string    { return "!(" + p + ")" }

// notAny is that no condition of the list holds: `true` for none.
func notAny(cs []string) string {
	if len(cs) == 0 {
		return "true"
	}
	return and(not(cs[0]), notAny(cs[1:]))
}

// --- guards ---

// A Position is where an observation sits in its base: the arrival, or the
// k-th observation its invoke was passed. It is what a frame's `$n` means
// in a stored branch (the scope table's `at`).
type Position int

// Arrived is the value that arrived under the base; Passed(k) is the k-th
// observation passed to the invoke it answers.
const Arrived Position = -1

func Passed(k int) Position { return Position(k) }

// cel is the position as a CEL variable: arrived, passed0, passed1, …
func (p Position) cel() string {
	if p == Arrived {
		return "arrived"
	}
	return "passed" + itoa(int(p))
}

// ref is the position as an instruction reference: @arrived, @passed[k].
func (p Position) ref() string {
	if p == Arrived {
		return "@arrived"
	}
	return "@passed[" + itoa(int(p)) + "]"
}

// isPosition reports whether a CEL identifier names a position: the only
// variables a stored guard may read.
func isPosition(name string) bool {
	if name == "arrived" {
		return true
	}
	digits, ok := strings.CutPrefix(name, "passed")
	if !ok || digits == "" || (len(digits) > 1 && digits[0] == '0') {
		return false
	}
	for _, c := range digits {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// Unstorable is a guard or instruction reading an observation with no
// position in the base: it is valid, and runs once, unlearned (58).
type Unstorable struct{ N int }

func (u Unstorable) Error() string {
	return "$" + itoa(u.N) + " is not in the base: it runs once, unlearned"
}

// StoreGuard translates the model's guard into its stored, base-relative
// form. The guard reads observations as `$n` and their fields by key
// (`$1.size`, `$1["content-type"]`); scope maps each number the frame holds
// in the base to its position. Outside string literals, `$n` becomes its
// position's CEL variable; in a returned template, its instruction
// reference. A `$n` inside a condition's string literal is text.
func StoreGuard(guard string, scope map[int]Position) ([]Link, error) {
	src, err := substitute(guard, func(n int) (string, error) {
		p, ok := scope[n]
		if !ok {
			return "", Unstorable{n}
		}
		return p.cel(), nil
	})
	if err != nil {
		return nil, err
	}
	chain, err := ParseGuard(src)
	if err != nil {
		return nil, err
	}
	for i, l := range chain {
		instr, err := StoreInstr(l.Instr, scope)
		if err != nil {
			return nil, err
		}
		chain[i].Instr = instr
	}
	return chain, nil
}

// substitute replaces each `$n` outside CEL string literals. It knows CEL's
// quoting: single or double quotes, tripled, and the raw prefix r, under
// which a backslash escapes nothing.
func substitute(src string, to func(int) (string, error)) (string, error) {
	var out strings.Builder
	for i := 0; i < len(src); {
		c := src[i]
		if c == '"' || c == '\'' {
			q := string(c)
			if strings.HasPrefix(src[i:], q+q+q) {
				q = q + q + q
			}
			raw := i > 0 && (src[i-1] == 'r' || src[i-1] == 'R')
			j := i + len(q)
			for ; j < len(src); j++ {
				if !raw && src[j] == '\\' {
					j++
					continue
				}
				if strings.HasPrefix(src[j:], q) {
					break
				}
			}
			if j >= len(src) {
				return "", errors.New("the guard has an unterminated string")
			}
			end := j + len(q)
			out.WriteString(src[i:end])
			i = end
			continue
		}
		if n, w := number(src[i:]); w > 0 {
			s, err := to(n)
			if err != nil {
				return "", err
			}
			out.WriteString(s)
			i += w
			continue
		}
		out.WriteByte(c)
		i++
	}
	return out.String(), nil
}

// references replaces each `$n` in an instruction template. How a literal
// `$` is escaped in a template is open (58, question 8).
func references(tmpl string, to func(int) (string, error)) (string, error) {
	var out strings.Builder
	for i := 0; i < len(tmpl); {
		if n, w := number(tmpl[i:]); w > 0 {
			s, err := to(n)
			if err != nil {
				return "", err
			}
			out.WriteString(s)
			i += w
			continue
		}
		out.WriteByte(tmpl[i])
		i++
	}
	return out.String(), nil
}

// number reads `$n` at the start of s, giving n and its width, or width 0.
func number(s string) (int, int) {
	if len(s) < 2 || s[0] != '$' {
		return 0, 0
	}
	n, w := 0, 1
	for w < len(s) && s[w] >= '0' && s[w] <= '9' && w < 10 {
		n = n*10 + int(s[w]-'0')
		w++
	}
	if w == 1 {
		return 0, 0
	}
	return n, w
}

// ParseGuard reads a stored guard: a chain of conditions, each returning an
// instruction as a string literal, ending in null.
func ParseGuard(src string) ([]Link, error) {
	parsed, iss := parseEnv.Parse(src)
	if iss != nil && iss.Err() != nil {
		return nil, errors.New("the guard does not parse: " + iss.Err().Error())
	}
	info := parsed.NativeRep().SourceInfo()
	var chain []Link
	for e := parsed.NativeRep().Expr(); ; {
		if e.Kind() == ast.LiteralKind && e.AsLiteral() == types.NullValue {
			return chain, nil
		}
		if e.Kind() != ast.CallKind || e.AsCall().FunctionName() != operators.Conditional {
			return nil, errors.New(`a guard is a chain c ? "instruction" : … : null`)
		}
		args := e.AsCall().Args()
		lit := args[1]
		s, ok := literalString(lit)
		if !ok {
			return nil, errors.New("a guard returns string literals or null: an instruction is never built from values")
		}
		c, err := cel.ExprToString(args[0], info)
		if err != nil {
			return nil, err
		}
		chain = append(chain, Link{Cond: c, Instr: s})
		e = args[2]
	}
}

func literalString(e ast.Expr) (string, bool) {
	if e.Kind() != ast.LiteralKind {
		return "", false
	}
	s, ok := e.AsLiteral().(types.String)
	return string(s), ok
}

// parseEnv parses guards and conditions; they are compiled against their
// own variables when evaluated. It tracks macro calls, so a condition using
// exists or all is written back as the model wrote it.
var parseEnv = func() *cel.Env {
	e, err := cel.NewEnv(cel.EnableMacroCallTracking())
	if err != nil {
		panic(err)
	}
	return e
}()

// guardLibs are the libraries a guard may call beyond CEL's standard ones:
// the strings extension (split, trim, lowerAscii, indexOf, substring, …),
// which models reach for when reading text.
var guardLibs = []cel.EnvOption{ext.Strings()}

// --- evaluation ---

// holds evaluates one condition over the values. An evaluation error is an
// error, never false. Each condition is compiled where it is evaluated: the
// core keeps nothing between steps.
func holds(cond string, vals map[string]any) (bool, error) {
	prg, err := compileCond(cond)
	if err != nil {
		return false, err
	}
	out, _, err := prg.Eval(celValues(vals))
	if err != nil {
		return false, err
	}
	b, ok := out.Value().(bool)
	if !ok {
		return false, errors.New(cond + " is not a predicate")
	}
	return b, nil
}

// compileCond compiles a stored condition in its base's environment: the
// positions it reads, arrived and passedK, and no other variable, so a
// condition naming anything else fails to compile. The positions are dyn
// until the core infers shapes (58, question 4); typed from the shape, a
// key the shape lacks will fail to compile too.
func compileCond(cond string) (cel.Program, error) {
	parsed, iss := parseEnv.Parse(cond)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}
	opts := slices.Clone(guardLibs)
	for _, name := range idents(parsed.NativeRep().Expr()) {
		if isPosition(name) {
			opts = append(opts, cel.Variable(name, cel.DynType))
		}
	}
	env, err := cel.NewEnv(opts...)
	if err != nil {
		return nil, err
	}
	checked, iss := env.Compile(cond)
	if iss != nil && iss.Err() != nil {
		return nil, iss.Err()
	}
	return env.Program(checked)
}

// celValues is the values as CEL reads them: an exit status as an int.
func celValues(vals map[string]any) map[string]any {
	var conv func(any) any
	conv = func(v any) any {
		switch v := v.(type) {
		case ExitStatus:
			return int64(v)
		case []any:
			out := make([]any, len(v))
			for i, x := range v {
				out[i] = conv(x)
			}
			return out
		case map[string]any:
			out := make(map[string]any, len(v))
			for k, x := range v {
				out[k] = conv(x)
			}
			return out
		}
		return v
	}
	out := make(map[string]any, len(vals))
	for k, v := range vals {
		out[k] = conv(v)
	}
	return out
}

func idents(e ast.Expr) []string {
	var out []string
	var walk func(ast.Expr)
	walk = func(e ast.Expr) {
		switch e.Kind() {
		case ast.IdentKind:
			if !slices.Contains(out, e.AsIdent()) {
				out = append(out, e.AsIdent())
			}
		case ast.CallKind:
			if t := e.AsCall().Target(); t != nil && t.Kind() != ast.UnspecifiedExprKind {
				walk(t)
			}
			for _, a := range e.AsCall().Args() {
				walk(a)
			}
		case ast.SelectKind:
			walk(e.AsSelect().Operand())
		case ast.ComprehensionKind:
			c := e.AsComprehension()
			for _, x := range []ast.Expr{c.IterRange(), c.AccuInit(), c.LoopCondition(), c.LoopStep(), c.Result()} {
				walk(x)
			}
		}
	}
	walk(e)
	return out
}

func selected(bs []Branch, vals map[string]any) ([]Branch, error) {
	var out []Branch
	for _, b := range bs {
		ok, err := holds(b.Cond, vals)
		if err != nil {
			return nil, errors.Join(errors.New("b"+itoa(b.ID)), err)
		}
		if ok {
			out = append(out, b)
		}
	}
	return out, nil
}

// A Selection is a tally's outcome: null, one instruction, or a conflict.
type Selection struct {
	Kind  SelectionKind
	Instr string // for SelectOne
	Level int    // for SelectOne from a tally: the level it was decided at
}

type SelectionKind uint8

const (
	SelectNull SelectionKind = iota
	SelectOne
	SelectConflict
)

// Distinguish selects: null when nothing selects the values, one
// instruction when every selected branch agrees on it, otherwise a
// conflict. Two branches selecting one instruction are not a conflict. The
// outcome does not depend on the branches' order. It is the tally at ι = 0,
// at the exact level only (Lean: distinguish_tally).
func Distinguish(bs []Branch, vals map[string]any) (Selection, error) {
	sel, err := selected(bs, vals)
	if err != nil {
		return Selection{}, err
	}
	if len(sel) == 0 {
		return Selection{Kind: SelectNull}, nil
	}
	for _, b := range sel[1:] {
		if b.Instr != sel[0].Instr {
			return Selection{Kind: SelectConflict}, nil
		}
	}
	return Selection{Kind: SelectOne, Instr: sel[0].Instr}, nil
}

// --- commit ---

// CommitGuard is startle's commit: the new version is the current branches,
// unchanged, plus the guard's, whole, each with its effective condition.
func CommitGuard(v Version, guard string) (Version, error) {
	chain, err := ParseGuard(guard)
	if err != nil {
		return Version{}, err
	}
	return CommitChain(v, chain), nil
}

// CommitChain commits a stored chain (StoreGuard's) as one new voter, named
// by the local id of its first branch. Nothing is narrowed: a branch that
// overlaps an older one with a different instruction is a dissenting vote,
// counted when values arrive.
func CommitChain(v Version, chain []Link) Version {
	out := Version{Next: v.Next + len(chain), Branches: slices.Clone(v.Branches)}
	var prev []string
	for k, l := range chain {
		out.Branches = append(out.Branches, Branch{ID: v.Next + k, Rank: v.Next + k, Voter: v.Next,
			Cond: and(l.Cond, notAny(prev)), Instr: l.Instr})
		prev = append(prev, l.Cond)
	}
	if out.Branches == nil {
		out.Branches = []Branch{}
	}
	return out
}

// --- after startle ---

// After is what a step does once startle has committed.
type After struct {
	Startle bool   // the chain did not cover the values: startle again
	Instr   string // otherwise, the instruction trust runs
	Learned bool   // whether the committed version selects it for these values
}

// AfterStartle evaluates over the version the step committed, never the
// live HEAD. It runs the instruction its own model's chain gives the values,
// whatever that version's tally selects: other voters may outvote it, and
// the step runs its own instruction once, unlearned.
func AfterStartle(chain []Link, committed []Branch, vals map[string]any, t Tally) (After, error) {
	for _, l := range chain {
		ok, err := holds(l.Cond, vals)
		if err != nil {
			return After{}, err
		}
		if !ok {
			continue
		}
		votes, err := Cast(committed, vals)
		if err != nil {
			return After{}, err
		}
		s := t.Decide(votes)
		return After{Instr: l.Instr, Learned: s.Kind == SelectOne && s.Instr == l.Instr}, nil
	}
	return After{Startle: true}, nil
}

// --- transitions and the transition bound ---

// A State is one thing the system does with an instruction.
type State uint8

const (
	_ State = iota
	StateReject
	StateHesitate
	StateStartle
	StateTally // a count that did not settle the instruction run: a conflict, or a commit counted
	StateTrust
)

// A Transition is the life of one instruction: the states it passed
// through. It ends in trust running the instruction, exactly once; the
// outcome starts the next transition.
type Transition []State

// ErrTransition is a transition that does not end in exactly one trust.
var ErrTransition = errors.New("a transition ends in trust, exactly once")

// Check is whether t is a whole transition.
func (t Transition) Check() error {
	if len(t) == 0 || t[len(t)-1] != StateTrust || slices.Contains(t[:len(t)-1], StateTrust) {
		return ErrTransition
	}
	return nil
}

// A TransitionBound counts the transitions a frame takes, whether or not they
// call the model: a learned cycle that never calls the model still ends.
// Each transition runs one instruction, so it bounds the frame's effects
// too. Cycles are neither forbidden nor detected; the bound is what makes a
// frame terminate. It is not yet a Profile field: the runtime has no trust
// state to end a transition with.
type TransitionBound struct {
	Transitions int // the most transitions the frame may take
	taken       int
}

// ErrTransitionBound is a frame's transition bound, exhausted.
var ErrTransitionBound = errors.New("the frame's transition bound is exhausted")

// Take counts one whole transition, or refuses it once the bound is
// reached: the frame then ends, returning Exhausted.
func (b *TransitionBound) Take(t Transition) error {
	if err := t.Check(); err != nil {
		return err
	}
	if b.taken >= b.Transitions {
		return ErrTransitionBound
	}
	b.taken++
	return nil
}

// Taken is how many transitions the frame has taken.
func (b *TransitionBound) Taken() int { return b.taken }

// Exhausted is what an exhausted frame returns to its caller: an error
// observation, {error: bytes}, holding the system's own message: not a
// decode error, so a base of its own. The caller
// receives it as any outcome: its base has no branch for it, so it is null,
// and the caller's model is startled with it.
func (b *TransitionBound) Exhausted() map[string]any {
	return map[string]any{"error": []byte("transition bound exhausted after " + itoa(b.taken) + " transitions")}
}
