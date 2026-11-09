package model

// Templates and interfaces as a person reads them: a template as the
// instruction the model wrote, which ParseInstr reads back to the same
// template; an interface as its shape, definition, or composition.

import (
	"strconv"
	"strings"
)

// Instr is the template as an instruction: `return 0 $4`, `$2 $1 '-l'`,
// `delegate(goal($1), $2)`.
func (t Template) Instr() string {
	args := make([]string, len(t.Args))
	for i, a := range t.Args {
		args[i] = a.text()
	}
	switch t.Op {
	case OpReturn:
		return strings.Join(append([]string{"return"}, args...), " ")
	case OpDelegate:
		if len(args) > 0 {
			args[0] = "goal(" + args[0] + ")"
		}
		return "delegate(" + strings.Join(args, ", ") + ")"
	}
	return strings.Join(args, " ")
}

func (a Arg) text() string {
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
				continue
			}
			b.WriteString(Arg{Var: p.Var, Sels: p.Sels}.text())
		}
		return b.String()
	case a.Pieces != nil:
		var b strings.Builder
		for _, p := range a.Pieces {
			b.WriteString(p.text())
		}
		return b.String()
	}
	var b strings.Builder
	b.WriteString("$" + strconv.Itoa(a.Var))
	for _, s := range a.Sels {
		b.WriteString(s.text())
	}
	return b.String()
}

// text is a selector as written: Enc, a run of one line its number alone.
func (s Sel) text() string {
	if !s.Lines {
		return s.Enc()
	}
	parts := make([]string, len(s.Runs))
	for i, r := range s.Runs {
		parts[i] = strconv.Itoa(r[0])
		if r[1] != r[0] {
			parts[i] += ":" + strconv.Itoa(r[1])
		}
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// Text is the interface as a person reads it: a shape, a definition by
// name, an operator over its parts, alternatives by |, or what a cut
// divides then its selector.
func (i Iface) Text() string {
	switch i.Tag {
	case IShape:
		return i.Shape.Enc()
	case IDef:
		return string(i.Def)
	case IOp, ISum:
		parts := make([]string, len(i.Parts))
		for n, p := range i.Parts {
			parts[n] = p.Text()
		}
		if i.Tag == ISum {
			return "(" + strings.Join(parts, " | ") + ")"
		}
		return string(i.Op) + "(" + strings.Join(parts, ", ") + ")"
	case ICut:
		if i.Of == nil {
			return "cut" + i.Sel.text()
		}
		return i.Of.Text() + i.Sel.text()
	}
	return "?"
}
