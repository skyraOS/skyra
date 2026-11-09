package model

// The machine: 61's runtime, composed of what Model.lean defines. It is pure:
// each Step takes one input and gives one effect for the shell to perform
// (run a leaf, ask the model, or end), and changes nothing it was handed.
//
// Genesis pushes the first frame: the goal as $1, then the program's tools,
// its kind counted under k (61, open question 8: the shell is the user,
// outside the tree). At each frame its key is decided: an unseen key is
// hesitated, recorded with nothing committed; the tally over the templates
// stored there selects one, or startles the model, whose guard is committed
// as one more implementation and its own template run. Every $n a frame
// holds is in its key, so every guard is stored: nothing runs unlearned.
// What the model writes is typed as 61 types it: a cut is an elimination,
// a literal an introduction, refused at commit when it copies text the
// frame holds, so what can be cut is cut. Running a template records each
// in the deciding frame's state, a $n of its own, whether the template was
// written or trusted; the template is stored inlined (Frame.Inline), its
// references by their positions in the key, which holds what was seen and
// never what was eliminated or introduced. So a stored $n means the same
// at every frame with that key, however its entries are numbered.
// A template's operator is one tree edge: an act pushes a leaf the world
// runs, a delegation a child, a return pops into the parent. A bound
// exhausted pops too, with an error observation.

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"

	"skyra/typesys-prototype/tree"
)

// Inputs.
type (
	// Answered is what a leaf's process gave: stdout, exit status and
	// stderr. The goal arrives as one, exit 0.
	Answered struct {
		Bytes  []byte
		Exit   int
		Stderr []byte
	}
	// Wrote is the guard the model wrote when asked.
	Wrote struct{ Guard string }
)

// Effects.
type (
	// RunLeaf is a leaf for the world: Argv is the tool, then each
	// argument's value as text. What it gives comes back as Answered.
	RunLeaf struct {
		Tool DefLoc
		Argv []string
	}
	// Ask startles the model at a frame's key: its kind (what it may
	// write), how deep it is (the first frame is 0), its $n, the model its
	// delegation chose, and why the last guard was refused, if one was.
	Ask struct {
		Key    Key
		Kind   Kind
		Depth  int
		Scope  []Shown
		Model  *DefLoc
		Reason string
	}
	// Finished is the first frame's return.
	Finished struct{ Value any }
	// Exhausted is the first frame's outcome when one of its bounds is
	// exhausted: an error observation.
	Exhausted struct{ Value map[string]any }
)

// A Role is what a $n is to its frame: an agent's $1 is its goal; a
// definition is a tool; everything else is context.
type Role uint8

const (
	RoleGoal Role = iota
	RoleTool
	RoleContext
)

// Shown is one $n as an ask shows it: its role, its tool or value, and,
// for an elimination or an introduction, where it came from.
type Shown struct {
	Role   Role
	Tool   DefLoc
	Value  any
	Origin string
}

// ErrAsking is an input the machine is not waiting for.
var ErrAsking = errors.New("the machine is not waiting for this input")

type phase uint8

const (
	phaseStart phase = iota // waiting on the goal
	phaseAsk                // waiting on the model's guard
	phaseRun                // waiting on the leaf the world runs
	phaseDone
)

// A Machine is one run: its program, memory (which outlives it), and the
// tree.
type Machine struct {
	Ids     Ids
	Program Program
	Memory  Memory
	Tree    Tree

	root  DefLoc // memory's root: the program without its constants
	loc   DefLoc // the program's location, every frame's static link
	at    InstLoc
	phase phase
	asks  int             // the asks of the transition in progress
	taken map[InstLoc]int // each frame's transitions
}

// DefaultConstants are a program's when its owner sets none: their values
// are not settled (58, open question 6).
var DefaultConstants = Constants{Act: 1, Transitions: 64, Refinements: 3, Inhibition: [2]int{0, 1}, Floor: 2, Quorum: 1}

// NewProgram is a program whose act reaches the tools and whose delegate
// reaches the models. Its entry scope is a root holding them all.
func NewProgram(version DefLoc, c Constants, tools, models []DefLoc) Program {
	sum := func(ds []DefLoc) Iface {
		parts := make([]Iface, len(ds))
		for i, d := range ds {
			parts[i] = DefIface(d)
		}
		return Iface{Tag: ISum, Parts: parts}
	}
	var entries []Iface
	for _, d := range append(slices.Clone(tools), models...) {
		entries = append(entries, DefIface(d))
	}
	return Program{Version: version, Constants: c, Entry: Scope{Kind: KindRoot, Entries: entries},
		Global: map[string]Iface{
			"act":      {Tag: IOp, Op: OpAct, Parts: []Iface{sum(tools)}},
			"delegate": {Tag: IOp, Op: OpDelegate, Parts: []Iface{sum(models)}},
		}}
}

// NewMachine is a run of the program over memory, waiting on its goal. The
// memory is not changed: each step works on its own.
func NewMachine(ids Ids, p Program, mem Memory) (Machine, error) {
	if !p.WellFormed() {
		return Machine{}, errors.New("the program is not well formed")
	}
	if len(p.Reach("act")) == 0 {
		return Machine{}, errors.New("a program reaches at least one tool")
	}
	if mem == nil {
		mem = Memory{}
	}
	return Machine{Ids: ids, Program: p, Memory: mem,
		root: ids.Program(Program{Version: p.Version, Global: p.Global}), loc: ids.Program(p)}, nil
}

// Step takes one input and gives the next effect.
func (m Machine) Step(in any) (Machine, any, error) {
	m.Memory, m.taken = maps.Clone(m.Memory), maps.Clone(m.taken)
	if m.taken == nil {
		m.taken = map[InstLoc]int{}
	}
	switch in := in.(type) {
	case Answered:
		switch m.phase {
		case phaseStart:
			v, _ := tree.Arrive(tree.OutputText, in.Bytes)
			return m.genesis(v)
		case phaseRun:
			v, _ := tree.ArriveExit(tree.OutputText, in.Bytes, in.Exit, in.Stderr)
			return m.end(m.at, v, false)
		}
	case Wrote:
		if m.phase == phaseAsk {
			return m.wrote(in.Guard)
		}
	}
	return m, nil, ErrAsking
}

// Frame is the frame deciding, or the leaf the world runs.
func (m Machine) Frame() Frame {
	f, _ := m.Tree.Find(m.at)
	return f
}

// Depth is how deep a frame is: the first frame is 0.
func (m Machine) Depth(id InstLoc) int {
	d := 0
	for f, ok := m.Tree.Find(id); ok && f.Parent != nil; f, ok = m.Tree.Find(*f.Parent) {
		d++
	}
	return d
}

// step is one tree edge. Every frame is the one program's, so what its
// operators reach is that program's.
func (m Machine) step(fid InstLoc, e Edge) (Tree, bool) {
	reaches := func(_ DefLoc, name string) []DefLoc { return m.Program.Reach(name) }
	return m.Tree.Step(m.Ids, m.Program.Constants.Act, reaches, fid, e)
}

func (m Machine) key(f Frame) Key { return f.Key(func(DefLoc) DefLoc { return m.root }) }

// genesis pushes the first frame: the goal, then the tools.
func (m Machine) genesis(goal any) (Machine, any, error) {
	tools := m.Program.Reach("act")
	id := m.Ids.First(m.loc, Grant{Defs: tools})
	o := Arrive(goal, m.Ids.Slot(id, 1))
	entries, bindings := []Iface{o.Iface}, []Member{{Obs: &o}}
	for _, t := range tools {
		entries, bindings = append(entries, DefIface(t)), append(bindings, Member{Tool: t})
	}
	kind, _ := AgentKind(m.Program.Constants.Act, tools, entries)
	f := Frame{ID: id, Program: m.loc, Scope: Scope{Kind: kind, Entries: entries}, Bindings: bindings}
	if models := m.Program.Reach("delegate"); len(models) > 0 {
		f.Model = &models[0]
	}
	m.Tree = Tree{Frames: []Frame{f}}
	return m.decide(id)
}

// seen is each $n a frame has seen (a tool, a binding, an arrival) by its
// position in the key, from 1, and back. Eliminations and introductions
// have none.
func seen(f Frame) (toKey, toItem map[int]int) {
	toKey, toItem = map[int]int{}, map[int]int{}
	for i, it := range f.Items() {
		if it.Tag == ItemTool || it.Tag == ItemSeen {
			s := len(toKey) + 1
			toKey[i+1], toItem[s] = s, i+1
		}
	}
	return toKey, toItem
}

// vals are what a frame has seen as a guard reads it, the key's position s
// as passed{s-1}: a tool as its name, every other $n as its value.
func vals(f Frame) map[string]any {
	out := map[string]any{}
	toKey, _ := seen(f)
	for n, s := range toKey {
		it, _ := f.Item(n)
		out["passed"+strconv.Itoa(s-1)] = itemValue(f, n, it)
	}
	return out
}

func itemValue(f Frame, n int, it Item) any {
	if it.Tag == ItemTool {
		return string(it.Tool)
	}
	v, _ := f.Read(n, nil)
	return v
}

// tables are a frame's $n as positions: for conditions, each $n seen as
// its key position's, with its shape; for instructions, every $n as
// passed{n-1}, both ways, read back to the frame's $n before inlining.
func tables(f Frame) (conds, instrs map[int]tree.Position, refs map[tree.Position]int,
	shapes map[tree.Position]tree.Shape) {
	conds, instrs, refs, shapes = map[int]tree.Position{}, map[int]tree.Position{}, map[tree.Position]int{},
		map[tree.Position]tree.Shape{}
	toKey, _ := seen(f)
	for i, it := range f.Items() {
		p := tree.Passed(i)
		instrs[i+1], refs[p] = p, i+1
		if s, ok := toKey[i+1]; ok {
			conds[i+1] = tree.Passed(s - 1)
			shapes[tree.Passed(s-1)] = tree.ShapeOf(itemValue(f, i+1, it))
		}
	}
	return conds, instrs, refs, shapes
}

// stored is a template as memory stores it: inlined, so it reads only what
// was seen, and each reference its key position.
func stored(f Frame, t Template) Template {
	toKey, _ := seen(f)
	out := Template{Op: t.Op}
	for _, a := range t.Args {
		if a.Lit == nil {
			a = f.Inline(a.Var, a.Sels)
		}
		if a.Lit == nil {
			a.Var = toKey[a.Var]
		}
		out.Args = append(out.Args, a)
	}
	return out
}

// bound is a stored template in a frame: each key position its $n there.
func bound(f Frame, t Template) (Template, error) {
	_, toItem := seen(f)
	out := Template{Op: t.Op}
	for _, a := range t.Args {
		if a.Lit == nil {
			n, ok := toItem[a.Var]
			if !ok {
				return Template{}, fmt.Errorf("the key has no position %d", a.Var)
			}
			a.Var = n
		}
		out.Args = append(out.Args, a)
	}
	return out, nil
}

// recorded is what running a template records: each cut an elimination,
// each literal an introduction, in its arguments' order.
func recorded(t Template) []Entry {
	var es []Entry
	for _, a := range t.Args {
		switch {
		case a.Lit != nil:
			l := *a.Lit
			es = append(es, Entry{Introduce: &l})
		case len(a.Sels) > 0:
			es = append(es, Entry{Eliminate: a.Var, Sels: a.Sels})
		}
	}
	return es
}

func (m Machine) tally() tree.Tally {
	c := m.Program.Constants
	return tree.Tally{Inhibition: c.Inhibition[0] * 1000 / c.Inhibition[1], Floor: c.Floor, Group: TemplateGroup}
}

// decide starts a frame's transition at its key: hesitate if unseen, then
// the tally; a template selected is trusted, null or a conflict startles.
func (m Machine) decide(id InstLoc) (Machine, any, error) {
	m.at, m.asks = id, 0
	f := m.Frame()
	k := m.key(f)
	if _, seen := m.Memory[k.Text()]; !seen {
		m.Memory.Apply(Event{Key: k})
	}
	t := m.tally()
	o, err := m.Memory.Decide(t.Inhibition, t.Floor, m.Program.Constants.Quorum, k, vals(f))
	if err != nil {
		return m, nil, err
	}
	if o.Kind == tree.SelectOne {
		return m.trust(f, o.Template)
	}
	return m.ask("")
}

// ask startles the model, or, the transition's asks spent, ends the frame.
func (m Machine) ask(reason string) (Machine, any, error) {
	f := m.Frame()
	if m.asks >= m.Program.Constants.Refinements {
		return m.end(f.ID, errorValue(fmt.Sprintf("refinement bound exhausted: the model was asked %d times at this key "+
			"without a guard that runs", m.asks)), true)
	}
	m.asks++
	m.phase = phaseAsk
	var scope []Shown
	for i, it := range f.Items() {
		switch {
		case it.Tag == ItemTool:
			scope = append(scope, Shown{Role: RoleTool, Tool: it.Tool})
		case i == 0:
			scope = append(scope, Shown{Role: RoleGoal, Value: itemValue(f, 1, it)})
		case it.Tag == ItemElim:
			scope = append(scope, Shown{Role: RoleContext, Value: itemValue(f, i+1, it),
				Origin: "cut from " + selText(Arg{Var: it.N, Sels: it.Sels})})
		case it.Tag == ItemIntro:
			scope = append(scope, Shown{Role: RoleContext, Value: itemValue(f, i+1, it), Origin: "introduced"})
		default:
			scope = append(scope, Shown{Role: RoleContext, Value: itemValue(f, i+1, it)})
		}
	}
	return m, Ask{Key: m.key(f), Kind: f.Scope.Kind, Depth: m.Depth(f.ID), Scope: scope, Model: f.Model,
		Reason: reason}, nil
}

// wrote stores the model's guard at the frame's key, checks it by shape and
// each template against the frame, commits it, and runs its own template
// for what the frame holds now.
func (m Machine) wrote(guard string) (Machine, any, error) {
	f := m.Frame()
	conds, instrs, refs, shapes := tables(f)
	chain, err := tree.StoreConds(guard, conds, instrs)
	var u tree.Unstorable
	if errors.As(err, &u) && u.N >= 1 && u.N <= len(instrs) {
		return m.ask(fmt.Sprintf("$%d was made in this frame, a cut or a literal: a condition reads what was seen, "+
			"so read what it was made from", u.N))
	}
	if errors.As(err, &u) {
		return m.ask(fmt.Sprintf("$%d is not in scope: the scope is $1 to $%d", u.N, len(instrs)))
	}
	if err != nil {
		return m.ask(err.Error())
	}
	if err := tree.CheckGuard(chain, shapes); err != nil {
		return m.ask(err.Error())
	}
	links := make([]Link, len(chain))
	for i, l := range chain {
		t, err := templateOf(l.Instr, refs)
		if err == nil {
			err = m.fits(f, t)
		}
		if err != nil {
			return m.ask(err.Error())
		}
		for _, a := range t.Args {
			if a.Lit == nil {
				continue
			}
			if _, n, ok := f.Introduce(*a.Lit); !ok {
				return m.ask(fmt.Sprintf("'%s' copies text $%d holds: take that part of $%d instead of writing it",
					a.Lit.Str, n, n))
			}
		}
		t = stored(f, t)
		links[i], chain[i].Instr = Link{Cond: l.Cond, Template: t}, t.Text()
	}
	// The chain's branches are cast on what the frame holds now, as every
	// later decision here casts them: a guard that does not evaluate on the
	// values it was written for is refused, never committed.
	if _, err := tree.Cast(tree.CommitChain(tree.EmptyVersion(), chain).Branches, vals(f)); err != nil {
		return m.ask("the guard does not evaluate on these values: " + err.Error())
	}
	k := m.key(f)
	m.Memory.Apply(Event{Key: k, Startle: true, Impl: links})
	after, err := tree.AfterStartle(chain, m.Memory[k.Text()].Branches, vals(f), m.tally())
	if err != nil {
		return m, nil, err
	}
	if after.Startle {
		return m.ask("the guard does not select what arrived")
	}
	t, err := ParseTemplate(after.Instr)
	if err != nil {
		return m, nil, err
	}
	return m.trust(f, t)
}

// templateOf is a stored instruction as a template: bound back to the
// frame's $n and read (ParseInstr).
func templateOf(stored string, refs map[tree.Position]int) (Template, error) {
	bound, err := tree.Bind(stored, refs)
	if err != nil {
		return Template{}, err
	}
	return ParseInstr(bound)
}

// fits is whether a template can run in a frame, whatever its values: its
// operator is one the frame's kind emits, a tool is run from a $n holding
// one, only values are arguments, and a delegation passes at least one
// tool and fewer than the frame holds.
func (m Machine) fits(f Frame, t Template) error {
	if !slices.Contains(Ops(f.Scope.Kind), t.Op) {
		switch t.Op {
		case OpAct:
			return errors.New("this frame delegates: it does not run tools itself")
		case OpDelegate:
			return errors.New("this frame runs its tools: it does not delegate")
		}
		return fmt.Errorf("this frame does not %s", t.Op)
	}
	items := f.Items()
	isTool := func(a Arg) bool {
		return a.Lit == nil && a.Var >= 1 && a.Var <= len(items) && items[a.Var-1].Tag == ItemTool
	}
	for i, a := range t.Args {
		if a.Lit != nil {
			continue
		}
		if a.Var < 1 || a.Var > len(items) {
			return fmt.Errorf("$%d is not in scope: the scope is $1 to $%d", a.Var, len(items))
		}
		switch {
		case t.Op == OpAct && i == 0 && !isTool(a):
			return fmt.Errorf("a tool is run from a $n holding one: $%d is not a tool", a.Var)
		case isTool(a) && len(a.Sels) > 0:
			return fmt.Errorf("$%d is a tool: it has no value to cut", a.Var)
		case isTool(a) && !(t.Op == OpAct && i == 0) && !(t.Op == OpDelegate && i > 0):
			return fmt.Errorf("$%d is a tool: only a delegation passes tools", a.Var)
		}
	}
	if t.Op == OpDelegate {
		if len(m.Program.Reach("delegate")) == 0 {
			return errors.New("this program has no model to delegate to")
		}
		held := ToolTypes(m.Program.Reach("act"), f.Scope.Entries)
		var passed []DefLoc
		for _, a := range t.Args[1:] {
			if isTool(a) && !slices.Contains(passed, items[a.Var-1].Tool) {
				passed = append(passed, items[a.Var-1].Tool)
			}
		}
		if len(passed) == 0 || len(passed) >= len(held) {
			return fmt.Errorf("a delegation passes at least one tool and fewer than this frame holds: "+
				"it passes %d of %d", len(passed), len(held))
		}
	}
	return nil
}

// trust runs a template, ending the transition and counting it against the
// frame's bound: a return pops, an act pushes a leaf for the world, a
// delegation pushes a child, which decides. A template whose arguments do
// not read in this frame is refused, and the model asked.
func (m Machine) trust(f Frame, st Template) (Machine, any, error) {
	t, err := bound(f, st)
	if err != nil {
		return m, nil, err
	}
	if err := m.fits(f, t); err != nil {
		return m.ask(err.Error())
	}
	values := make([]any, len(t.Args))
	for i, a := range t.Args {
		if a.Lit == nil && f.Items()[a.Var-1].Tag == ItemTool {
			continue
		}
		v, ok := f.Eval(a)
		if !ok {
			return m.ask(fmt.Sprintf("%s does not apply to $%d's value", selText(a), a.Var))
		}
		values[i] = v
	}
	m.taken[f.ID]++
	if bound := m.Program.Constants.Transitions; m.taken[f.ID] > bound {
		return m.end(f.ID, errorValue(fmt.Sprintf("transition bound exhausted: %d transitions", bound)), true)
	}
	if es := recorded(t); len(es) > 0 {
		m.Tree = m.Tree.Record(f.ID, es)
		f, _ = m.Tree.Find(f.ID)
	}
	switch t.Op {
	case OpReturn:
		return m.end(f.ID, ReturnValue(values[0].(tree.ExitStatus), values[1], values[2:]), false)
	case OpAct:
		tool := f.Items()[t.Args[0].Var-1].Tool
		argv := []string{string(tool)}
		entries, bindings := []Iface{DefIface(tool)}, []Member{{Tool: tool}}
		for _, v := range values[1:] {
			argv = append(argv, argText(v))
			entries = append(entries, Arrive(v, "").Iface)
		}
		s := Scope{Kind: KindLeaf, Entries: entries}
		leaf := m.Ids.Child(f.ID, m.Tree.Invokes(f.ID), s)
		for i, v := range values[1:] {
			o := Arrive(v, m.Ids.Slot(leaf, i+2))
			bindings = append(bindings, Member{Obs: &o})
		}
		next, ok := m.step(f.ID, Edge{Tag: EdgeInvoke, Op: OpAct, Scope: s, Bindings: bindings})
		if !ok {
			return m, nil, errors.New("the tree refused an act the frame fits")
		}
		m.Tree, m.at, m.phase = next, leaf, phaseRun
		return m, RunLeaf{Tool: tool, Argv: argv}, nil
	}
	var entries []Iface
	var passed []func(InstLoc, int) Member
	for i, a := range t.Args {
		if v := values[i]; i == 0 || a.Lit != nil || f.Items()[a.Var-1].Tag != ItemTool {
			entries = append(entries, Arrive(v, "").Iface)
			passed = append(passed, func(id InstLoc, n int) Member {
				o := Arrive(v, m.Ids.Slot(id, n))
				return Member{Obs: &o}
			})
			continue
		}
		tool := f.Items()[a.Var-1].Tool
		entries = append(entries, DefIface(tool))
		passed = append(passed, func(InstLoc, int) Member { return Member{Tool: tool} })
	}
	kind, _ := AgentKind(m.Program.Constants.Act, m.Program.Reach("act"), entries)
	s := Scope{Kind: kind, Entries: entries}
	child := m.Ids.Child(f.ID, m.Tree.Invokes(f.ID), s)
	bindings := make([]Member, len(passed))
	for i, p := range passed {
		bindings[i] = p(child, i+1)
	}
	model := f.Model
	if model == nil {
		model = &m.Program.Reach("delegate")[0]
	}
	next, ok := m.step(f.ID, Edge{Tag: EdgeInvoke, Op: OpDelegate, Scope: s, Bindings: bindings, Model: model})
	if !ok {
		return m, nil, errors.New("the tree refused a delegation the frame fits")
	}
	m.Tree = next
	return m.decide(child)
}

// end pops a frame with a value: the first frame's is the run's outcome;
// any other's lands in its parent, which decides. exhausted marks an
// exhausted bound's error observation.
func (m Machine) end(id InstLoc, v any, exhausted bool) (Machine, any, error) {
	f, _ := m.Tree.Find(id)
	if f.Parent == nil {
		m.phase = phaseDone
		if exhausted {
			return m, Exhausted{Value: v.(map[string]any)}, nil
		}
		return m, Finished{Value: v}, nil
	}
	p, _ := m.Tree.Find(*f.Parent)
	next, ok := m.step(id, Edge{Tag: EdgeReturn, Obs: Arrive(v, m.Ids.Slot(p.ID, len(p.Items())+1))})
	if !ok {
		return m, nil, errors.New("the tree refused a return")
	}
	m.Tree = next
	return m.decide(p.ID)
}

// ReturnValue is what a return gives its caller, built by the machine,
// never written: its exit status, the value it returns as stdout, and,
// when it names one, its why as stderr. Each keeps its own type: text
// returned is text, an object keeps its fields. The exit status is a shape
// no JSON value has, so a return is never mistaken for one.
func ReturnValue(exit tree.ExitStatus, stdout any, why []any) map[string]any {
	v := map[string]any{"exit": exit, "stdout": stdout}
	if len(why) > 0 {
		v["stderr"] = why[0]
	}
	return v
}

func errorValue(msg string) map[string]any { return map[string]any{"error": []byte(msg)} }

func selText(a Arg) string {
	s := "$" + strconv.Itoa(a.Var)
	for _, x := range a.Sels {
		s += x.Enc()
	}
	return s
}

// argText is a value as an argument: a string or bytes as themselves,
// anything else as JSON.
func argText(v any) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	}
	b, _ := json.Marshal(v)
	return string(b)
}
