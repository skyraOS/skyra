package model

// Instructions as the model writes them, read into templates. Each form is
// one operator: `return 0 $4 [$5]` returns, `$2 $1 …` runs the tool at $2,
// `delegate(goal($1), $2, …)` delegates. An argument is a $n and any chain
// of selectors, each cutting what the last gave: $3.stdout[1] is line 1 of
// $3's field stdout. Nothing else may be written: text that is not one of
// these is refused, never dropped.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"skyra/typesys-prototype/tree"
)

// ParseInstr reads an instruction, its $n the frame's.
func ParseInstr(s string) (Template, error) {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return Template{}, errors.New("an empty instruction")
	}
	head := fields[0]
	switch {
	case head == "return":
		if (len(fields) != 3 && len(fields) != 4) || (fields[1] != "0" && fields[1] != "1") {
			return Template{}, errors.New("return takes an exit status, 0 when the goal is done or 1 when it is " +
				"not, then what it returns, a $n or a cut of one, and optionally why, another")
		}
		exit := Lit{IsExit: true, Exit: int(fields[1][0] - '0')}
		args, err := argsOf(fields[2:])
		return Template{Op: OpReturn, Args: append([]Arg{{Lit: &exit}}, args...)}, err
	case strings.HasPrefix(head, "$"):
		tool, err := argOf(head)
		if err != nil {
			return Template{}, err
		}
		if len(tool.Sels) > 0 {
			return Template{}, errors.New("a tool is run whole: the first position is a $n, not a cut")
		}
		args, err := argsOf(fields[1:])
		return Template{Op: OpAct, Args: append([]Arg{tool}, args...)}, err
	}
	inner, ok := strings.CutPrefix(strings.TrimSpace(s), "delegate(")
	if inner, ok = strings.CutSuffix(inner, ")"); !ok {
		return Template{}, errors.New("this is not an instruction: write one of the forms under Instructions")
	}
	parts := splitArgs(inner)
	g, ok := strings.CutPrefix(parts[0], "goal(")
	if g, ok = strings.CutSuffix(g, ")"); !ok {
		return Template{}, errors.New("a delegation's first argument is the goal: goal($n), or goal of a cut")
	}
	args, err := argsOf(append([]string{g}, parts[1:]...))
	return Template{Op: OpDelegate, Args: args}, err
}

func argsOf(fs []string) ([]Arg, error) {
	args := make([]Arg, len(fs))
	for i, f := range fs {
		var err error
		if args[i], err = argOf(f); err != nil {
			return nil, err
		}
	}
	return args, nil
}

// argOf reads one argument, all of f: a $n, then its selectors in order.
func argOf(f string) (Arg, error) {
	bad := fmt.Errorf("an argument is a $n or a cut of one, and nothing else: %s is neither", f)
	w := 1
	for w < len(f) && f[w] >= '0' && f[w] <= '9' && w < 10 {
		w++
	}
	if len(f) < 2 || f[0] != '$' || w == 1 {
		return Arg{}, bad
	}
	n, _ := strconv.Atoi(f[1:w])
	a := Arg{Var: n}
	for rest := f[w:]; rest != ""; {
		s, sw, err := tree.ReadSel(rest)
		if err != nil {
			return Arg{}, err
		}
		if sw == 0 {
			return Arg{}, bad
		}
		if s.Kind == tree.SelLines {
			a.Sels = append(a.Sels, Sel{Lines: true, Runs: s.Runs})
		} else {
			a.Sels = append(a.Sels, Sel{Path: s.Path})
		}
		rest = rest[sw:]
	}
	return a, nil
}

// splitArgs splits at the commas between arguments, not those inside a
// cut's brackets or a goal's parentheses, each trimmed.
func splitArgs(s string) []string {
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '[', '(':
			depth++
		case ']', ')':
			depth--
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(s[start:]))
}
