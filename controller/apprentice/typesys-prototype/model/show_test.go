package model

import (
	"reflect"
	"testing"
)

// An instruction read and shown again is itself, and reads back to the
// same template.
func TestInstrRoundTrip(t *testing.T) {
	for _, s := range []string{
		"return 0 $4",
		"return 1 $5 $3.stderr",
		"$2 $1",
		"$2 $1[2] '-maxdepth' '1'",
		"$2 $3.stdout[1:2,4]",
		"delegate(goal($1), $2)",
		"delegate(goal($1[2]), $3, $4'/'$5.stdout[1])",
	} {
		tpl, err := ParseInstr(s)
		if err != nil {
			t.Fatalf("%s: %v", s, err)
		}
		if got := tpl.Instr(); got != s {
			t.Fatalf("%s shows as %s", s, got)
		}
		back, err := ParseInstr(tpl.Instr())
		if err != nil || !reflect.DeepEqual(back, tpl) {
			t.Fatalf("%s does not read back: %v", s, err)
		}
	}
}
