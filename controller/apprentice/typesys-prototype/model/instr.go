package model

// Instructions as the model writes them, read into templates. Each form is
// one operator: `return 0 $4 [$5]` returns, `$2 $1 …` runs the tool at $2,
// `delegate(goal($1), $2, …)` delegates. An argument is a $n and any chain
// of selectors, each cutting what the last gave: $3.stdout[1] is line 1 of
// $3's field stdout. Or it is an introduction: a literal in single quotes,
// '-l' or '.version', text no $n holds, constructed whole. Nothing else may
// be written: text that is not one of these is refused, never dropped.
// Whether a literal copies what the frame holds is the machine's to check
// (Frame.Introduce), not the grammar's.

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"skyra/typesys-prototype/tree"
)

// ParseInstr reads an instruction, its $n the frame's.
func ParseInstr(s string) (Template, error) {
	fields, err := words(s)
	if err != nil {
		return Template{}, err
	}
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
	if err == nil && args[0].Lit != nil {
		return Template{}, errors.New("a goal is taken from a parameter, never written: goal($n), or goal of a cut")
	}
	return Template{Op: OpDelegate, Args: args}, err
}

// words splits an instruction at spaces outside single quotes.
func words(s string) ([]string, error) {
	var ws []string
	var cur strings.Builder
	quoted, in := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '\'':
			quoted, in = !quoted, true
			cur.WriteByte(c)
		case !quoted && (c == ' ' || c == '\t' || c == '\n'):
			if in {
				ws, in = append(ws, cur.String()), false
				cur.Reset()
			}
		default:
			cur.WriteByte(c)
			in = true
		}
	}
	if quoted {
		return nil, errors.New("a literal is unterminated: close its single quote")
	}
	if in {
		ws = append(ws, cur.String())
	}
	return ws, nil
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

// argOf reads one argument, all of f: a literal; or a $n, then its
// selectors in order. A literal holds no $ or @, so it is never read as a
// reference.
func argOf(f string) (Arg, error) {
	if len(f) >= 2 && f[0] == '\'' && f[len(f)-1] == '\'' {
		text := f[1 : len(f)-1]
		if strings.ContainsAny(text, "'$@") {
			return Arg{}, fmt.Errorf("a literal holds no ', $ or @: %s does", f)
		}
		return Arg{Lit: &Lit{Str: text}}, nil
	}
	bad := fmt.Errorf("an argument is a $n, a cut of one, or a literal in single quotes, and nothing else: %s is none", f)
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
	depth, start, quoted := 0, 0, false
	for i := 0; i < len(s); i++ {
		if s[i] == '\'' {
			quoted = !quoted
		}
		if quoted {
			continue
		}
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
