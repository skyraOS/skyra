package dnav3

import "testing"

const goodEmission = `
<framing>Michael wants the auth refactor to compile with tests passing.</framing>

<act id="a1" target="file:/src/auth.go" primitive="write">func Authenticate() {}</act>

<act id="a2" target="build:/src" primitive="compile"
     requires="exact file:/src/auth.go exists true"/>

<act id="a3" target="being:michael" primitive="engage"
     requires="exact test:auth_test.go status fail">The refactor broke two tests — revert?</act>

<expect of="a1">
  exact      0.95  1   file:/src/auth.go   exists    true
  exact      0.90  1   file:/src/auth.go   contains  "func Authenticate"
</expect>
<expect of="a2">  exact      0.85  3   build:/src          status    ok         </expect>
<expect of="a3">  exact      0.99  2   being:michael       received  "revert?"  </expect>
<expect of="a3">  interpret  0.40  8   being:michael       state     satisfied  </expect>
<expect of="a1 a2">  exact  0.60  5   test:auth_test.go   status    pass       </expect>
`

func TestParse(t *testing.T) {
	res, err := Parse(goodEmission)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(res.Rejects) != 0 {
		t.Fatalf("unexpected rejects: %+v", res.Rejects)
	}
	e := res.Expression

	if e.Framing == "" {
		t.Error("framing not parsed")
	}
	if len(e.Acts) != 3 {
		t.Fatalf("got %d acts, want 3", len(e.Acts))
	}
	if len(e.Expects) != 6 {
		t.Fatalf("got %d expects, want 6", len(e.Expects))
	}

	// payload survives, and it is the only interpretive surface
	if e.Acts[0].Payload != "func Authenticate() {}" {
		t.Errorf("payload = %q", e.Acts[0].Payload)
	}
	// self-closing act
	if e.Acts[1].Primitive != "compile" || e.Acts[1].Payload != "" {
		t.Errorf("self-closing act = %+v", e.Acts[1])
	}
	// requires parsed as a triple
	if len(e.Acts[1].Requires) != 1 || e.Acts[1].Requires[0].Subject != "file:/src/auth.go" {
		t.Errorf("requires = %+v", e.Acts[1].Requires)
	}
	// quoted object with spaces
	if e.Expects[1].Object != "func Authenticate" {
		t.Errorf("quoted object = %q", e.Expects[1].Object)
	}
	// within
	if e.Expects[2].Within != 3 {
		t.Errorf("within = %d, want 3", e.Expects[2].Within)
	}
	// interpret mode
	if e.Expects[4].Mode != ModeInterpret {
		t.Errorf("mode = %q, want interpret", e.Expects[4].Mode)
	}
	// JOINT: cardinality on `of` is the edge's source cardinality
	joint := e.Expects[5]
	if len(joint.Of) != 2 || joint.Of[0] != "a1" || joint.Of[1] != "a2" {
		t.Errorf("joint of = %v, want [a1 a2]", joint.Of)
	}

	if err := e.Validate(); err != nil {
		t.Errorf("Validate: %v", err)
	}
}

// A malformed line costs ONE FACT, not the document. This is the property the whole grammar exists
// for: fewer predictions is acceptable, no action is not.
func TestParseDegradesGracefully(t *testing.T) {
	res, err := Parse(`
<framing>partly broken</framing>
<act id="a1" target="file:/x" primitive="write">body</act>
<expect of="a1">
  exact   0.95  1   file:/x   exists   true
  exact   BOGUS 1   file:/x   size     10
  exact   0.90  0   file:/x   mode     644
  banana  0.90  1   file:/x   owner    me
  exact   0.80  2   file:/x   perms    rw
</expect>
`)
	if err != nil {
		t.Fatalf("Parse should not fail on bad LINES: %v", err)
	}
	if got := len(res.Expression.Expects); got != 2 {
		t.Errorf("kept %d expects, want 2 (the two good lines)", got)
	}
	if got := len(res.Rejects); got != 3 {
		t.Fatalf("got %d rejects, want 3", got)
	}
	// and each rejection says why — a model repeating the same bad line is telling you something
	for _, r := range res.Rejects {
		if r.Reason == "" {
			t.Errorf("reject with no reason: %+v", r)
		}
	}
}

// Truncation is free: run out of output tokens and you get fewer predictions, not a corrupt
// document. This is why the envelope is XML and not JSON.
func TestParseSurvivesTruncation(t *testing.T) {
	truncated := `
<framing>ran out of tokens</framing>
<act id="a1" target="file:/x" primitive="write">body</act>
<expect of="a1">  exact  0.95  1  file:/x  exists  true  </expect>
<act id="a2" target="file:/y" primitive="wr`

	res, err := Parse(truncated)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if len(res.Expression.Acts) != 1 {
		t.Errorf("got %d acts, want 1 (the complete one)", len(res.Expression.Acts))
	}
	if len(res.Expression.Expects) != 1 {
		t.Errorf("got %d expects, want 1", len(res.Expression.Expects))
	}
}

// LAW 5, at the parser: an act that projects nothing is unfalsifiable, therefore unlearnable.
func TestValidateRejectsUnfalsifiableAct(t *testing.T) {
	res, err := Parse(`
<act id="a1" target="file:/x" primitive="write">body</act>
<act id="a2" target="file:/y" primitive="write">body</act>
<expect of="a1">  exact  0.95  1  file:/x  exists  true  </expect>
`)
	if err != nil {
		t.Fatal(err)
	}
	if err := res.Expression.Validate(); err == nil {
		t.Error("Validate accepted an act that projects no expected world state")
	}
}

// A bet with no deadline is never settled, therefore unlearnable.
func TestParseRejectsBetWithNoDeadline(t *testing.T) {
	res, _ := Parse(`
<act id="a1" target="file:/x" primitive="write">body</act>
<expect of="a1">  exact  0.95  0  file:/x  exists  true  </expect>
`)
	if len(res.Expression.Expects) != 0 {
		t.Error("parser accepted an expect with within=0")
	}
	if len(res.Rejects) != 1 {
		t.Fatalf("got %d rejects, want 1", len(res.Rejects))
	}
}
