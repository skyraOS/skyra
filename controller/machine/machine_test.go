package machine

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"skyra/controller/contract"
	"skyra/controller/horizon"
	"skyra/controller/record"
)

var (
	text   = &contract.Schema{Type: "string"}
	echoIn = &contract.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*contract.Schema{"text": text}}
	echoS  = horizon.Surface{Path: "/bin/echo-target", SHA256: "e", Input: echoIn, Output: text}
	agentS = horizon.Surface{Path: "/bin/agent", SHA256: "a", Input: echoIn, Output: text, Reach: []string{"echo"}}
	both   = map[string]horizon.Surface{"echo": echoS, "agent": agentS}
	// Horizons as the emitter holds them: the operator (no surface), and a leaf.
	operator = horizon.Horizon{Stamp: "g0", Reach: both}
	leaf     = horizon.Horizon{Stamp: "r1", Surface: &echoS, Reach: map[string]horizon.Surface{}}
)

// intakeRecord runs the Intake phase and returns what it commits.
func intakeRecord(h horizon.Horizon, in string) record.Record {
	_, effs := Step(Instance{Phase: Intake}, Received{Under: h, In: []byte(in)})
	return Committed(effs)
}

// The emitter supplies operands; interpret derives the state.
func TestInterpret(t *testing.T) {
	numeric := leaf
	numeric.Surface = &horizon.Surface{Path: "/bin/n", Input: echoIn, Output: &contract.Schema{Type: "number"}}
	cases := []struct {
		name    string
		under   horizon.Horizon
		in      string
		state   record.State
		to      string // invoke: the capability
		payload string // invoke or return: the canonical payload
		kind    string // fault: its kind
		cause   string // fault: a substring of its cause
	}{
		{"addressed, in reach, fits", operator, `{"to":"echo","payload":{"text":"hi"}}`, record.Invoke, "echo", `{"text":"hi"}`, "", ""},
		{"unaddressed text is a reply", leaf, `hello`, record.Return, "", `"hello"`, "", ""},
		{"explicit reply", leaf, `{"to":"reply","payload":"hi"}`, record.Return, "", `"hi"`, "", ""},
		{"unaddressed JSON is still just text", leaf, `{"text":"hi"}`, record.Return, "", `"{\"text\":\"hi\"}"`, "", ""},
		{"a leaf addresses a capability", leaf, `{"to":"echo","payload":{"text":"x"}}`, record.Fault, "", "", record.Horizon, `"echo" is not in reach`},
		{"out of reach", operator, `{"to":"ghost","payload":{}}`, record.Fault, "", "", record.Horizon, "not in reach"},
		{"the operator has no reply", operator, `hello`, record.Fault, "", "", record.Horizon, "operator holds no reply"},
		{"payload does not fit", operator, `{"to":"echo","payload":{"txt":"x"}}`, record.Fault, "", "", record.Contract, `unknown field "txt"`},
		{"payload missing a field", operator, `{"to":"echo","payload":{}}`, record.Fault, "", "", record.Contract, `missing required field "text"`},
		{"reply does not fit", numeric, `hello`, record.Fault, "", "", record.Contract, "reply: want a number"},
		{"address without payload", operator, `{"to":"echo"}`, record.Fault, "", "", record.Malformed, "exactly"},
		{"address with extra fields", operator, `{"to":"echo","payload":{},"x":1}`, record.Fault, "", "", record.Malformed, "exactly"},
		{"address is not a name", operator, `{"to":7,"payload":{}}`, record.Fault, "", "", record.Malformed, "exactly"},
		{"no output", leaf, ``, record.Fault, "", "", record.Custody, "no output"},
		{"not text", leaf, "\xff\xfe", record.Fault, "", "", record.Malformed, "UTF-8"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, effs := Step(Instance{Phase: Intake}, Received{Under: c.under, In: []byte(c.in)})
			r := Committed(effs)
			if s.Phase != Release || s.Stamp != record.Stamp(r) || s.State != r.State {
				t.Fatalf("intake must derive the state, commit, stamp and wait for the release: %+v", s)
			}
			if !reflect.DeepEqual(effs, []Effect{StoreInput{[]byte(c.in)}, Commit{r}, SendStamp{s.Stamp}}) {
				t.Fatalf("intake effects: %#v", effs)
			}
			if r.Predecessor != c.under.Stamp || r.Input != record.Sum([]byte(c.in)) {
				t.Fatalf("record does not bind predecessor and input: %+v", r)
			}
			if r.State != c.state || r.To != c.to || string(r.Payload) != c.payload {
				t.Fatalf("got %s to=%q payload=%s fault=%+v; want %s to=%q payload=%s", r.State, r.To, r.Payload, r.Fault, c.state, c.to, c.payload)
			}
			if c.kind != "" && (r.Fault == nil || r.Fault.Kind != c.kind || !strings.Contains(r.Fault.Cause, c.cause)) {
				t.Fatalf("fault %+v, want %s mentioning %q", r.Fault, c.kind, c.cause)
			}
		})
	}
}

// The state picks the program: the same Released event runs a target for an
// invoke and settles the parent for a return or a fault.
func TestStateIsTheTopLevelBranch(t *testing.T) {
	for _, c := range []struct {
		under horizon.Horizon
		in    string
		state record.State
		first Effect // the first effect after the release
	}{
		{operator, `{"to":"echo","payload":{"text":"x"}}`, record.Invoke, Run{}},
		{leaf, `x`, record.Return, Settle{}},
		{operator, `{"to":"ghost","payload":{}}`, record.Fault, Settle{}},
	} {
		s, _ := Step(Instance{Phase: Intake}, Received{Under: c.under, In: []byte(c.in)})
		if s.State != c.state {
			t.Fatalf("%s: state %s, want %s", c.in, s.State, c.state)
		}
		self := horizon.Horizon{Stamp: s.Stamp, Surface: &echoS, Reach: map[string]horizon.Surface{}, Settles: c.under.Stamp}
		_, effs := Step(s, Released{Self: self, Node: true})
		if reflect.TypeOf(effs[0]) != reflect.TypeOf(c.first) {
			t.Fatalf("%s (%s): first effect %T, want %T", c.in, c.state, effs[0], c.first)
		}
	}

	// A terminal state never executes, even if an event for Execute arrives.
	s, _ := Step(Instance{Phase: Intake}, Received{Under: leaf, In: []byte(`x`)})
	s.Phase = Execute
	_, effs := Step(s, Emitted{Out: []byte("x")})
	if x, ok := effs[0].(Exit); !ok || x.Err == "" {
		t.Fatalf("a return must not execute: %#v", effs)
	}
}

func TestStepIsDeterministic(t *testing.T) {
	in := `{"payload":{"text":"hi"},"to":"echo"}`
	a, _ := record.Encode(intakeRecord(operator, in))
	b, _ := record.Encode(intakeRecord(operator, in))
	if string(a) != string(b) {
		t.Fatalf("step is not byte-identical across runs:\n%s\n%s", a, b)
	}
}

// awaiting returns an instance that has committed an invoke of echo and is
// waiting for its release.
func awaiting(t *testing.T) Instance {
	t.Helper()
	s, _ := Step(Instance{Floor: "/state/.skyra", Phase: Intake},
		Received{Under: operator, In: []byte(`{"to":"echo","payload":{"text":"hi"}}`)})
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
	if !reflect.DeepEqual(run.Surface, echoS) || string(run.Stdin) != `{"text":"hi"}` {
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
	s, _ := Step(Instance{Phase: Intake}, Received{Under: leaf, In: []byte(`x`)})
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
	obs := effs[0].(StoreInput).In
	f := effs[1].(Commit).R
	if !f.Supervisor || f.State != record.Fault || f.Fault.Kind != record.Custody || f.Predecessor != s.Stamp || !strings.Contains(f.Fault.Cause, "without emitting") {
		t.Fatalf("supervisor fault: %+v", f)
	}
	if f.Input != record.Sum(obs) || record.Stamp(Observed(s.Stamp, obs)) != record.Stamp(f) {
		t.Fatal("a supervisor fault must be derivable from its stored observation")
	}
	if !reflect.DeepEqual(effs[2:], []Effect{Settle{record.Stamp(f)}, Exit{}}) {
		t.Fatalf("the fault settles this instance's own frame: %#v", effs)
	}
}

func TestHandoff(t *testing.T) {
	s := executing(t)
	out := []byte(`hi`)
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
	out := []byte(`hi`)
	s, _ = Step(s, Emitted{Out: out})
	good := intakeRecord(s.Self, string(out))

	wrongParent := good
	wrongParent.Predecessor = "someone-else"
	wrongBytes := intakeRecord(s.Self, `other`)

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
	invoke := func(under horizon.Horizon, target string) record.Record {
		return intakeRecord(under, `{"to":"`+target+`","payload":{"text":"x"}}`)
	}

	h, err := Narrow(operator, invoke(operator, "echo"), "r1")
	if err != nil || len(h.Reach) != 0 || h.Surface.Path != echoS.Path {
		t.Fatalf("a leaf must get its surface and nothing in reach: %+v %v", h, err)
	}

	h, err = Narrow(operator, invoke(operator, "agent"), "r1")
	if err != nil || len(h.Reach) != 1 || !reflect.DeepEqual(h.Reach["echo"], echoS) {
		t.Fatalf("a non-leaf is granted what it asks for and its parent holds: %+v %v", h, err)
	}

	narrower := horizon.Horizon{Stamp: "g0", Reach: map[string]horizon.Surface{"agent": agentS}}
	h, err = Narrow(narrower, invoke(narrower, "agent"), "r1")
	if err != nil || len(h.Reach) != 0 {
		t.Fatalf("a Horizon must never widen: asked for echo, the parent does not hold it: %+v %v", h, err)
	}

	forged := record.Record{State: record.Invoke, To: "agent"}
	if _, err := Narrow(horizon.Horizon{Stamp: "r1", Reach: map[string]horizon.Surface{}}, forged, "r2"); err == nil {
		t.Fatal("narrow granted a target outside the parent's reach")
	}
}

// The machine must stay pure. It may import only the standard library's
// non-I/O packages and the pure packages of this module.
func TestMachineIsPure(t *testing.T) {
	allowed := map[string]bool{
		"errors": true, "fmt": true, "unicode/utf8": true,
		"skyra/controller/contract": true,
		"skyra/controller/horizon":  true,
		"skyra/controller/record":   true,
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
