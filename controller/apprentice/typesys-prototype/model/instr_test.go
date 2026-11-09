package model

import (
	"reflect"
	"strings"
	"testing"
)

func v(n int, sels ...Sel) Arg { return Arg{Var: n, Sels: sels} }

func lit(s string) Arg { return Arg{Lit: &Lit{Str: s}} }

// Each form reads into its template, a selector chain cutting in order;
// text that is not an instruction is refused, never dropped.
func TestParseInstr(t *testing.T) {
	exit := func(n int) Arg { return Arg{Lit: &Lit{IsExit: true, Exit: n}} }
	field := func(k ...string) Sel { return Sel{Path: k} }
	line := func(a, b int) Sel { return Sel{Lines: true, Runs: [][2]int{{a, b}}} }
	for s, want := range map[string]Template{
		"$2 $1":                          {Op: OpAct, Args: []Arg{v(2), v(1)}},
		"$2 $3.stdout[1]":                {Op: OpAct, Args: []Arg{v(2), v(3, field("stdout"), line(1, 1))}},
		"$2 $1[2] $4.a.b[1:2].c":         {Op: OpAct, Args: []Arg{v(2), v(1, line(2, 2)), v(4, field("a", "b"), line(1, 2), field("c"))}},
		"return 0 $4":                    {Op: OpReturn, Args: []Arg{exit(0), v(4)}},
		"return 1 $3.stdout $3.stderr":   {Op: OpReturn, Args: []Arg{exit(1), v(3, field("stdout")), v(3, field("stderr"))}},
		"delegate(goal($1[2]), $2, $4)":  {Op: OpDelegate, Args: []Arg{v(1, line(2, 2)), v(2), v(4)}},
		"$2 '.version' $1[2]":            {Op: OpAct, Args: []Arg{v(2), lit(".version"), v(1, line(2, 2))}},
		"$2 '-n' 'a b' $1":               {Op: OpAct, Args: []Arg{v(2), lit("-n"), lit("a b"), v(1)}},
		"delegate(goal($1), $2, 'x, y')": {Op: OpDelegate, Args: []Arg{v(1), v(2), lit("x, y")}},
		"return 1 $3 'not found'":        {Op: OpReturn, Args: []Arg{exit(1), v(3), lit("not found")}},
	} {
		got, err := ParseInstr(s)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%s: %+v %v", s, got, err)
		}
	}
	for s, why := range map[string]string{
		"$2 -la":                  "nothing else",
		"$2 $1 extra":             "nothing else",
		"$2 $1)":                  "nothing else",
		"$2.x $1":                 "run whole",
		"return 2 $1":             "exit status",
		"return 0 $1 $2 $3":       "exit status",
		"delegate($1, $2)":        "goal",
		"ls $1":                   "not an instruction",
		"$2 $1[x]":                "a cut is",
		"$2 '-l":                  "unterminated",
		"$2 'a$1'":                "holds no",
		"'ls' $1":                 "not an instruction",
		"delegate(goal('x'), $2)": "never written",
	} {
		if _, err := ParseInstr(s); err == nil || !strings.Contains(err.Error(), why) {
			t.Errorf("%s: %v, want %q", s, err, why)
		}
	}
}
