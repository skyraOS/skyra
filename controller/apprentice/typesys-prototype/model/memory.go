package model

// Memory and decisions, as Model.lean defines them. Memory is a fold of the
// log onto 59's refinements, one per key: a hesitate records a key never
// seen, a startle commits an implementation there. The decision at a key is
// 59's tally over the templates stored there. A branch's instruction is its
// template's canonical JSON, so commit and casting are tree's own.

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"skyra/typesys-prototype/tree"
)

// A Template is an operator and its arguments, made entirely of variables.
type Template struct {
	Op   Op
	Args []Arg
}

// Text is a template's canonical JSON: the instruction a branch stores.
func (t Template) Text() string { return compact(t) }

// ParseTemplate reads a template from its text.
func ParseTemplate(text string) (Template, error) {
	var t Template
	err := json.Unmarshal([]byte(text), &t)
	return t, err
}

// A Link is one link of an implementation's guard chain: a condition, as
// CEL, and the template it selects.
type Link struct {
	Cond     string
	Template Template
}

// An Event is what the log records at a key: a hesitate, a startle
// committing an implementation, one guard chain, or a forget, taking the
// key back to unseen. The machine never forgets; a person does, through
// the shell.
type Event struct {
	Key     Key
	Startle bool
	Impl    []Link
	Forget  bool
}

// Text is a key's canonical JSON: memory's index.
func (k Key) Text() string { return compact(k) }

// Memory is each seen key's refinement, by the key's text.
type Memory map[string]tree.Version

// Apply is memory after one event: a hesitate keeps what is there, or
// records the key, committed to nothing; a startle commits its chain
// whole; a forget erases the key, everything committed there gone.
func (m Memory) Apply(e Event) {
	k := e.Key.Text()
	if e.Forget {
		delete(m, k)
		return
	}
	v, seen := m[k]
	if !seen {
		v = tree.EmptyVersion()
	}
	if e.Startle {
		links := make([]tree.Link, len(e.Impl))
		for i, l := range e.Impl {
			links[i] = tree.Link{Cond: l.Cond, Instr: l.Template.Text()}
		}
		v = tree.CommitChain(v, links)
	}
	m[k] = v
}

// Refuted classes are kept in memory beside the keys, outside any key: a
// key's text is a JSON object, and these are not.
const refutedPrefix = "refuted:"

// Refute records a class refuted.
func (m Memory) Refute(c DefLoc) { m[refutedPrefix+string(c)] = tree.EmptyVersion() }

// Refuted is whether a class has been refuted.
func (m Memory) Refuted(c DefLoc) bool {
	_, ok := m[refutedPrefix+string(c)]
	return ok
}

// RefutedClasses are the classes refuted, sorted.
func (m Memory) RefutedClasses() []DefLoc {
	var cs []DefLoc
	for k := range m {
		if c, ok := strings.CutPrefix(k, refutedPrefix); ok {
			cs = append(cs, DefLoc(c))
		}
	}
	slices.Sort(cs)
	return cs
}

// Replay is memory from the log.
func Replay(log []Event) Memory {
	m := Memory{}
	for _, e := range log {
		m.Apply(e)
	}
	return m
}

// TemplateGroup is a template's group at level k, as 59 groups an
// instruction: its first k parts, its operator then its arguments, or the
// whole template at the exact level. A floor of 2 never merges two
// operators, or two heads.
func TemplateGroup(k int, instr string) string {
	if k >= tree.Exact {
		return instr
	}
	t, err := ParseTemplate(instr)
	if err != nil {
		return instr
	}
	parts := []string{compact(t.Op)}
	for _, a := range t.Args {
		parts = append(parts, compact(a))
	}
	return strings.Join(parts[:min(k, len(parts))], "\x00")
}

// An Outcome is what a decision gives: null (the path to startle), a
// template, or a conflict.
type Outcome struct {
	Kind     tree.SelectionKind
	Template Template
}

// Implementations is how many implementations a refinement holds: its
// voters, each once.
func Implementations(v tree.Version) int {
	voters := map[int]bool{}
	for _, b := range v.Branches {
		voters[b.Voter] = true
	}
	return len(voters)
}

// Decide is the decision at a key, given the values it reads: null for a
// key unseen, never committed, or holding fewer implementations than the
// quorum; else the tally, grouping by template.
func (m Memory) Decide(inhibition, floor, quorum int, k Key, vals map[string]any) (Outcome, error) {
	v, seen := m[k.Text()]
	if !seen || Implementations(v) < quorum {
		return Outcome{Kind: tree.SelectNull}, nil
	}
	votes, err := tree.Cast(v.Branches, vals)
	if err != nil {
		return Outcome{}, err
	}
	s := tree.Tally{Inhibition: inhibition, Floor: floor, Group: TemplateGroup}.Decide(votes)
	if s.Kind != tree.SelectOne {
		return Outcome{Kind: s.Kind}, nil
	}
	t, err := ParseTemplate(s.Instr)
	return Outcome{Kind: tree.SelectOne, Template: t}, err
}

func (t Template) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]any{"op": t.Op, "args": nonNil(t.Args)})
}

func (t *Template) UnmarshalJSON(b []byte) error {
	var x struct {
		Op   Op    `json:"op"`
		Args []Arg `json:"args"`
	}
	err := json.Unmarshal(b, &x)
	*t = Template{Op: x.Op, Args: x.Args}
	return err
}

func (a *Arg) UnmarshalJSON(b []byte) error {
	var x struct {
		Var  *int    `json:"var"`
		Sels []Sel   `json:"sels"`
		Name *string `json:"name"`
		Lit  *Lit    `json:"lit"`
		Join *[]Part `json:"join"`
	}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	switch {
	case x.Join != nil:
		*a = Arg{Join: append([]Part{}, *x.Join...)}
	case x.Lit != nil:
		*a = Arg{Lit: x.Lit}
	case x.Name != nil:
		*a = Arg{Name: *x.Name}
	case x.Var != nil:
		*a = Arg{Var: *x.Var, Sels: x.Sels}
	default:
		return fmt.Errorf("argument %s", b)
	}
	return nil
}

func (e Event) MarshalJSON() ([]byte, error) {
	if e.Forget {
		return json.Marshal(map[string]any{"forget": e.Key})
	}
	if !e.Startle {
		return json.Marshal(map[string]any{"hesitate": e.Key})
	}
	impl := make([]map[string]any, len(e.Impl))
	for i, l := range e.Impl {
		impl[i] = map[string]any{"cond": l.Cond, "template": l.Template}
	}
	return json.Marshal(map[string]any{"startle": e.Key, "impl": impl})
}

func (e *Event) UnmarshalJSON(b []byte) error {
	var x struct {
		Hesitate *Key `json:"hesitate"`
		Startle  *Key `json:"startle"`
		Forget   *Key `json:"forget"`
		Impl     []struct {
			Cond     string   `json:"cond"`
			Template Template `json:"template"`
		} `json:"impl"`
	}
	if err := json.Unmarshal(b, &x); err != nil {
		return err
	}
	switch {
	case x.Hesitate != nil:
		*e = Event{Key: *x.Hesitate}
	case x.Forget != nil:
		*e = Event{Key: *x.Forget, Forget: true}
	case x.Startle != nil:
		*e = Event{Key: *x.Startle, Startle: true}
		for _, l := range x.Impl {
			e.Impl = append(e.Impl, Link{Cond: l.Cond, Template: l.Template})
		}
	default:
		return fmt.Errorf("event %s", b)
	}
	return nil
}

func (o Outcome) MarshalJSON() ([]byte, error) {
	switch o.Kind {
	case tree.SelectOne:
		return json.Marshal(map[string]any{"select": o.Template})
	case tree.SelectConflict:
		return json.Marshal("conflict")
	}
	return []byte("null"), nil
}
