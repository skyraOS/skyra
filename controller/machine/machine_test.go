package machine

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"skyra/controller/horizon"
	"skyra/controller/move"
	"skyra/controller/record"
)

var (
	echoS  = horizon.Surface{Path: "/bin/echo-target", SHA256: "e", Leaf: true}
	agentS = horizon.Surface{Path: "/bin/agent", SHA256: "a", Leaf: false}
	both   = map[string]horizon.Surface{"echo": echoS, "agent": agentS}
)

// intakeRecord runs the Intake phase and returns what it commits.
func intakeRecord(h horizon.Horizon, in string) record.Record {
	_, effs := Step(Instance{Phase: Intake}, Received{Under: h, In: []byte(in)})
	return Committed(effs)
}

func TestIntake(t *testing.T) {
	root := horizon.Horizon{Stamp: "g0", Reach: both}
	leaf := horizon.Horizon{Stamp: "r1", Reach: map[string]horizon.Surface{}}
	cases := []struct {
		name   string
		under  horizon.Horizon
		in     string
		state  move.State
		reason string // substring, for faults decided at intake
	}{
		{"invoke in reach", root, `{"invoke":{"target":"echo","args":{"text":"hi"}}}`, move.StateInvoke, ""},
		{"return", leaf, `{"return":{"value":"hi"}}`, move.StateReturn, ""},
		{"emitted fault", leaf, `{"fault":{"reason":"boom"}}`, move.StateFault, ""},
		{"leaf invokes", leaf, `{"invoke":{"target":"echo","args":{}}}`, move.StateFault, "a leaf may only return"},
		{"target out of reach", root, `{"invoke":{"target":"ghost","args":{}}}`, move.StateFault, "not in reach"},
		{"unknown field", root, `{"return":{"value":"hi","extra":1}}`, move.StateFault, "unknown field"},
		{"two variants", root, `{"return":{"value":"a"},"fault":{"reason":"b"}}`, move.StateFault, "exactly one"},
		{"no variant", root, `{}`, move.StateFault, "exactly one"},
		{"garbage", root, `not json`, move.StateFault, "invalid"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, effs := Step(Instance{Phase: Intake}, Received{Under: c.under, In: []byte(c.in)})
			r := Committed(effs)
			if s.Phase != Release || s.Stamp != record.Stamp(r) || s.State != r.State {
				t.Fatalf("intake must decide the state, commit, stamp and wait for the release: %+v", s)
			}
			if !reflect.DeepEqual(effs, []Effect{StoreInput{[]byte(c.in)}, Commit{r}, SendStamp{s.Stamp}}) {
				t.Fatalf("intake effects: %#v", effs)
			}
			if r.Predecessor != c.under.Stamp || r.Input != record.Sum([]byte(c.in)) {
				t.Fatalf("record does not bind predecessor and input: %+v", r)
			}
			if r.State != c.state {
				t.Fatalf("state %q, want %q (reason %q)", r.State, c.state, r.Reason)
			}
			if c.reason != "" && !strings.Contains(r.Reason, c.reason) {
				t.Fatalf("reason %q, want it to mention %q", r.Reason, c.reason)
			}
		})
	}
}

// The state picks the program: the same Released event runs a target for an
// invoke and settles the parent for a return or a fault.
func TestStateIsTheTopLevelBranch(t *testing.T) {
	for _, c := range []struct {
		in    string
		state move.State
		first Effect // the first effect after the release
	}{
		{`{"invoke":{"target":"echo","args":{}}}`, move.StateInvoke, Run{}},
		{`{"return":{"value":"x"}}`, move.StateReturn, Settle{}},
		{`{"fault":{"reason":"x"}}`, move.StateFault, Settle{}},
		{`{"invoke":{"target":"ghost","args":{}}}`, move.StateFault, Settle{}}, // rejected at intake
	} {
		s, _ := Step(Instance{Phase: Intake}, Received{Under: horizon.Horizon{Stamp: "p", Reach: both}, In: []byte(c.in)})
		if s.State != c.state {
			t.Fatalf("%s: state %s, want %s", c.in, s.State, c.state)
		}
		self := horizon.Horizon{Stamp: s.Stamp, Surface: &echoS, Reach: map[string]horizon.Surface{}, Settles: "p"}
		_, effs := Step(s, Released{Self: self, Node: true})
		if reflect.TypeOf(effs[0]) != reflect.TypeOf(c.first) {
			t.Fatalf("%s (%s): first effect %T, want %T", c.in, c.state, effs[0], c.first)
		}
	}

	// A terminal state never executes, even if an event for Execute arrives.
	s, _ := Step(Instance{Phase: Intake}, Received{Under: horizon.Horizon{Stamp: "p"}, In: []byte(`{"return":{"value":"x"}}`)})
	s.Phase = Execute
	_, effs := Step(s, Emitted{Out: []byte("x")})
	if x, ok := effs[0].(Exit); !ok || x.Err == "" {
		t.Fatalf("a return must not execute: %#v", effs)
	}
}

func TestStepIsDeterministic(t *testing.T) {
	under := horizon.Horizon{Stamp: "g0", Reach: both}
	in := `{"invoke":{"target":"echo","args":{"b":"2","a":"1"}}}`
	a, _ := record.Encode(intakeRecord(under, in))
	b, _ := record.Encode(intakeRecord(under, in))
	if string(a) != string(b) {
		t.Fatalf("step is not byte-identical across runs:\n%s\n%s", a, b)
	}
}

// awaiting returns an instance that has committed an invoke of echo and is
// waiting for its release.
func awaiting(t *testing.T) Instance {
	t.Helper()
	s, _ := Step(Instance{Floor: "/state/.skyra", Phase: Intake},
		Received{Under: horizon.Horizon{Stamp: "g0", Reach: both}, In: []byte(`{"invoke":{"target":"echo","args":{"text":"hi"}}}`)})
	return s
}

func TestReleaseStartsExecution(t *testing.T) {
	s := awaiting(t)
	self := horizon.Horizon{Stamp: s.Stamp, Surface: &echoS, Reach: map[string]horizon.Surface{}}
	s, effs := Step(s, Released{Self: self, Node: true})
	if s.Phase != Execute || len(effs) != 1 {
		t.Fatalf("phase %s, effects %#v", s.Phase, effs)
	}
	run := effs[0].(Run)
	if run.Surface != echoS || string(run.Stdin) != `{"text":"hi"}` {
		t.Fatalf("must run exactly the committed surface with the committed args: %#v", run)
	}
	if !strings.Contains(run.Profile, `(deny file-read* file-write* (subpath "/state/.skyra"))`) {
		t.Fatalf("rendered profile lacks the .skyra floor:\n%s", run.Profile)
	}
}

func TestReleaseRefusals(t *testing.T) {
	s := awaiting(t)
	for name, ev := range map[string]Released{
		"for another commit": {Self: horizon.Horizon{Stamp: "other"}, Node: true},
		"without a node":     {Self: horizon.Horizon{Stamp: s.Stamp}, Node: false},
	} {
		_, effs := Step(s, ev)
		if x, ok := effs[0].(Exit); !ok || x.Err == "" {
			t.Fatalf("%s: a bad release must stop the instance without executing: %#v", name, effs)
		}
	}
}

func TestReleaseToSettle(t *testing.T) {
	s, _ := Step(Instance{Phase: Intake}, Received{Under: horizon.Horizon{Stamp: "r1"}, In: []byte(`{"return":{"value":"x"}}`)})
	s, effs := Step(s, Released{Self: horizon.Horizon{Stamp: s.Stamp, Settles: "r1"}, Node: true})
	if !reflect.DeepEqual(effs, []Effect{Settle{s.Stamp}, Exit{}}) {
		t.Fatalf("a settling release writes the result and exits: %#v", effs)
	}
}

// executing returns an instance whose target is running.
func executing(t *testing.T) Instance {
	t.Helper()
	s := awaiting(t)
	s, _ = Step(s, Released{Self: horizon.Horizon{Stamp: s.Stamp, Surface: &echoS, Reach: map[string]horizon.Surface{}}, Node: true})
	return s
}

func TestSilentTargetFaults(t *testing.T) {
	s := executing(t)
	_, effs := Step(s, Emitted{})
	f := effs[0].(Commit).R
	if !f.Supervisor || f.State != move.StateFault || f.Predecessor != s.Stamp || !strings.Contains(f.Reason, "without emitting") {
		t.Fatalf("supervisor fault: %+v", f)
	}
	if !reflect.DeepEqual(effs[1:], []Effect{Settle{record.Stamp(f)}, Exit{}}) {
		t.Fatalf("the fault settles this instance's own frame: %#v", effs)
	}
}

func TestHandoff(t *testing.T) {
	s := executing(t)
	out := []byte(`{"return":{"value":"hi"}}`)
	s, effs := Step(s, Emitted{Out: out})
	if !reflect.DeepEqual(effs, []Effect{Spawn{Under: s.Self, Out: out}}) {
		t.Fatalf("the emission goes to a successor, uninterpreted: %#v", effs)
	}

	// The successor commits exactly what a real successor would.
	succ := intakeRecord(s.Self, string(out))
	s2, effs := Step(s, Stamped{Stamp: record.Stamp(succ), R: &succ})
	want := horizon.Horizon{Stamp: record.Stamp(succ), Reach: map[string]horizon.Surface{}, Settles: s.Stamp}
	if !reflect.DeepEqual(effs, []Effect{SendHorizon{Child: want, NewFrame: false}, Exit{}}) {
		t.Fatalf("a verified Return is released to settle this frame: %#v", effs)
	}
	if s2.Child != record.Stamp(succ) {
		t.Fatalf("child stamp not recorded: %+v", s2)
	}
}

func TestHandoffRefusals(t *testing.T) {
	s := executing(t)
	out := []byte(`{"return":{"value":"hi"}}`)
	s, _ = Step(s, Emitted{Out: out})
	good := intakeRecord(s.Self, string(out))

	wrongParent := good
	wrongParent.Predecessor = "someone-else"
	wrongBytes := intakeRecord(s.Self, `{"return":{"value":"other"}}`)

	for name, ev := range map[string]Stamped{
		"no record":     {Stamp: "nothing"},
		"wrong parent":  {Stamp: record.Stamp(wrongParent), R: &wrongParent},
		"wrong bytes":   {Stamp: record.Stamp(wrongBytes), R: &wrongBytes},
		"died silently": {Err: "EOF"},
	} {
		_, effs := Step(s, ev)
		var fault *record.Record
		for _, e := range effs {
			if _, ok := e.(SendHorizon); ok {
				t.Fatalf("%s: released a successor that failed verification", name)
			}
			if c, ok := e.(Commit); ok {
				fault = &c.R
			}
		}
		if fault == nil || !fault.Supervisor {
			t.Fatalf("%s: expected a supervisor fault, got %#v", name, effs)
		}
		if name != "died silently" {
			if _, ok := effs[0].(KillSuccessor); !ok {
				t.Fatalf("%s: an unverified successor must be killed first: %#v", name, effs)
			}
		}
	}
}

func TestOutOfPhase(t *testing.T) {
	_, effs := Step(Instance{Phase: Intake}, Emitted{Out: []byte("x")})
	if x, ok := effs[0].(Exit); !ok || !strings.Contains(x.Err, "not valid in phase") {
		t.Fatalf("an event in the wrong phase must stop the instance: %#v", effs)
	}
}

func TestNarrow(t *testing.T) {
	parent := horizon.Horizon{Stamp: "g0", Reach: both}
	invoke := func(under horizon.Horizon, target string) record.Record {
		return intakeRecord(under, `{"invoke":{"target":"`+target+`","args":{}}}`)
	}

	h, err := Narrow(parent, invoke(parent, "echo"), "r1")
	if err != nil || len(h.Reach) != 0 || h.Surface.Path != echoS.Path {
		t.Fatalf("a leaf must get its surface and nothing in reach: %+v %v", h, err)
	}

	h, err = Narrow(parent, invoke(parent, "agent"), "r1")
	if err != nil || len(h.Reach) != 2 {
		t.Fatalf("a non-leaf reaches at most what its parent reaches: %+v %v", h, err)
	}

	narrower := horizon.Horizon{Stamp: "g0", Reach: map[string]horizon.Surface{"agent": agentS}}
	h, err = Narrow(narrower, invoke(narrower, "agent"), "r1")
	if err != nil || len(h.Reach) != 1 || h.Reach["agent"] != agentS {
		t.Fatalf("a horizon.Horizon must never widen: %+v %v", h, err)
	}

	forged := record.Record{State: move.StateInvoke, Move: &move.Move{Invoke: &move.Invoke{Target: "agent"}}}
	if _, err := Narrow(horizon.Horizon{Stamp: "r1", Reach: map[string]horizon.Surface{}}, forged, "r2"); err == nil {
		t.Fatal("narrow granted a target outside the parent's reach")
	}
}

// The machine must stay pure. It may import only the standard library's
// non-I/O packages and the pure packages of this module.
func TestMachineIsPure(t *testing.T) {
	allowed := map[string]bool{
		"encoding/json": true, "errors": true, "fmt": true, "maps": true,
		"skyra/controller/horizon": true,
		"skyra/controller/move":    true,
		"skyra/controller/record":  true,
	}
	files, _ := filepath.Glob("*.go")
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		pf, err := parser.ParseFile(token.NewFileSet(), f, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, imp := range pf.Imports {
			p, _ := strconv.Unquote(imp.Path.Value)
			if !allowed[p] {
				t.Errorf("%s imports %q: package machine must not touch the OS, the tree or the network", f, p)
			}
		}
	}
}
