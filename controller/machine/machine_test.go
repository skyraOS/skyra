package machine

import (
	"encoding/json"
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
	operator = horizon.Horizon{Stamp: "g0", Reach: both, Frame: "g0"}
	leaf     = horizon.Horizon{Stamp: "r1", Surface: &echoS, Reach: map[string]horizon.Surface{}, Frame: "r1", Resume: &operator}
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

// released feeds a release carrying the whole frame stack the Horizon needs.
func released(s Instance, self horizon.Horizon) (Instance, []Effect) {
	return Step(s, Released{Self: self, Frames: self.Depth()})
}

// The state picks the program: the same Released event runs a target for an
// invoke and settles a slot for a return or a fault.
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
		self, _ := Narrow(c.under, s.R, s.Stamp)
		_, effs := released(s, self)
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

func TestStopReleasesOnlyItsConnection(t *testing.T) {
	s := Instance{Phase: Execute, State: record.Invoke, Resources: []horizon.Resource{
		{Kind: "connection", Owner: "codex", Incarnation: "one"},
		{Kind: "connection", Owner: "codex", Incarnation: "two"},
	}}
	s, effects := Step(s, Emitted{Out: []byte(`{"to":"reply","payload":{}}`), ReleasedResource: "one"})
	if s.Phase != Handoff || len(s.Resources) != 1 || s.Resources[0].Incarnation != "two" {
		t.Fatalf("resource release: %+v", s)
	}
	if _, ok := effects[0].(Spawn); !ok {
		t.Fatalf("expected handoff, got %#v", effects)
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

func selfOf(t *testing.T, s Instance) horizon.Horizon {
	t.Helper()
	h, err := Narrow(s.Under, s.R, s.Stamp)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestReleaseStartsExecution(t *testing.T) {
	s := awaiting(t)
	s, effs := released(s, selfOf(t, s))
	if s.Phase != Execute || len(effs) != 1 {
		t.Fatalf("phase %s, effects %#v", s.Phase, effs)
	}
	run := effs[0].(Run)
	if !reflect.DeepEqual(run.Surface, echoS) || string(run.Stdin) != `{"text":"hi"}` {
		t.Fatalf("a leaf runs exactly the committed surface with the committed payload: %#v", run)
	}
	if !strings.Contains(run.Profile, `(deny file-read* file-write* (subpath "/state/.skyra"))`) {
		t.Fatalf("rendered profile lacks the .skyra floor:\n%s", run.Profile)
	}
}

func TestReleaseRefusals(t *testing.T) {
	s := awaiting(t)
	good := selfOf(t, s)
	other := good
	other.Stamp = "other"
	for name, ev := range map[string]Released{
		"for another commit":        {Self: other, Frames: 2},
		"without its frame stack":   {Self: good, Frames: 0},
		"missing the waiting frame": {Self: good, Frames: 1},
	} {
		_, effs := Step(s, ev)
		if x, ok := effs[0].(Exit); !ok || x.Err == "" {
			t.Fatalf("%s: a bad release must stop the instance without executing: %#v", name, effs)
		}
	}
}

// A return or fault discharges its slot. If a continuation waits on it, the
// instance goes on to read what that continuation still waits on; if only the
// operator waits, it exits.
func TestReleaseToSettle(t *testing.T) {
	s, _ := Step(Instance{Phase: Intake}, Received{Under: leaf, In: []byte(`x`)})
	s, effs := released(s, horizon.Horizon{Stamp: s.Stamp, Frame: "r1"})
	if !reflect.DeepEqual(effs, []Effect{Settle{s.Stamp}, Exit{}}) {
		t.Fatalf("with only the operator waiting: %#v", effs)
	}
	s, _ = Step(Instance{Phase: Intake}, Received{Under: leaf, In: []byte(`x`)})
	s, effs = released(s, horizon.Horizon{Stamp: s.Stamp, Frame: "r1", Resume: &operator})
	if !reflect.DeepEqual(effs, []Effect{Settle{s.Stamp}, GatherWaiter{}}) || s.Phase != Gather {
		t.Fatalf("with a continuation waiting: %#v", effs)
	}
}

// executing returns an instance whose target is running.
func executing(t *testing.T) Instance {
	t.Helper()
	s := awaiting(t)
	s, _ = released(s, selfOf(t, s))
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
	if !reflect.DeepEqual(effs[2:], []Effect{Settle{record.Stamp(f)}, GatherWaiter{}}) {
		t.Fatalf("the fault discharges this instance's own slot and wakes its waiter: %#v", effs)
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
	want := horizon.Horizon{Stamp: record.Stamp(succ), Reach: map[string]horizon.Surface{}, Frame: s.Self.Frame, Resume: s.Self.Resume}
	if !reflect.DeepEqual(effs, []Effect{SendHorizon{Child: want, NewFrame: false}, Exit{}}) {
		t.Fatalf("a verified return is released to discharge this frame's slot: %#v", effs)
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

// A continuation invokes echo, waits on its slot, and resumes with what came
// back, all without processes: the whole round trip is step.
func TestContinuationInvokesWaitsResumes(t *testing.T) {
	// The operator invokes the agent. It runs with a continuation: no steps yet.
	a, _ := Step(Instance{Phase: Intake}, Received{Under: operator, In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
	a, effs := released(a, selfOf(t, a))
	var c Continuation
	if err := json.Unmarshal(effs[0].(Run).Stdin, &c); err != nil || c.Frame != a.Stamp || string(c.Input) != `{"text":"hi"}` || len(c.Steps) != 0 {
		t.Fatalf("a continuation starts with its frame, its input and no steps: %s", effs[0].(Run).Stdin)
	}

	// It addresses echo: the successor derives an invoke, and the agent opens
	// a slot for it in its own frame, in this round, before releasing it.
	call := []byte(`{"to":"echo","payload":{"text":"hi"}}`)
	a, _ = Step(a, Emitted{Out: call})
	b, _ := Step(Instance{Phase: Intake}, Received{Under: a.Self, In: call})
	_, effs = Step(a, Stamped{Stamp: b.Stamp, R: &b.R})
	if effs[0] != (OpenDep{Round: a.Self.Stamp, Child: b.Stamp}) || !effs[1].(SendHorizon).NewFrame {
		t.Fatalf("an invoke's slot is opened before it is released: %#v", effs)
	}

	// Echo replies; the reply discharges echo's slot and wakes the agent's frame.
	b, _ = released(b, effs[1].(SendHorizon).Child)
	b, _ = Step(b, Emitted{Out: []byte("hi")})
	r, _ := Step(Instance{Phase: Intake}, Received{Under: b.Self, In: []byte("hi")})
	_, effs = Step(b, Stamped{Stamp: r.Stamp, R: &r.R})
	r, effs = released(r, effs[0].(SendHorizon).Child)
	if !reflect.DeepEqual(effs, []Effect{Settle{r.Stamp}, GatherWaiter{}}) {
		t.Fatalf("the reply discharges echo's slot, then reads the agent's round: %#v", effs)
	}

	agentRecord := a.R
	round := func(discharged bool) Settled {
		child := Child{Stamp: b.Stamp, Invoke: b.R}
		if discharged {
			child.Result, child.Discharge = r.Stamp, &r.R
		}
		return Settled{Frame: agentRecord, Rounds: []Round{{Stamp: a.Stamp, Children: []Child{child}}}}
	}

	// Wait for all: an open slot means no resolve, and nothing is committed.
	if _, effs := Step(r, round(false)); !reflect.DeepEqual(effs, []Effect{Exit{}}) {
		t.Fatalf("a round with an open slot must not resolve: %#v", effs)
	}
	// Rounds that do not belong to the waiter are refused.
	wrong := round(true)
	wrong.Rounds[0].Stamp = "elsewhere"
	if _, effs := Step(r, wrong); effs[0].(Exit).Err == "" {
		t.Fatalf("a foreign round must stop the instance: %#v", effs)
	}

	// Every slot discharged: commit a resolve naming every discharge, then claim.
	g, effs := Step(r, round(true))
	res := effs[1].(Commit).R
	if res.State != record.Resolve || res.Frame != a.Stamp || res.Round != a.Stamp || !reflect.DeepEqual(res.Deps, []string{r.Stamp}) || res.Predecessor != r.Stamp {
		t.Fatalf("resolve record: %+v", res)
	}
	if derived, _ := Resolved(r.Stamp, effs[0].(StoreInput).In); record.Stamp(derived) != record.Stamp(res) {
		t.Fatal("a resolve must be derivable from its stored observation")
	}
	if effs[2] != (TryClaim{Round: a.Stamp, Stamp: record.Stamp(res)}) {
		t.Fatalf("then the fan-in claim: %#v", effs)
	}

	// A lost claim exits; the winner resumes the agent with what echo said.
	if _, effs := Step(g, Claimed{Won: false}); !reflect.DeepEqual(effs, []Effect{Exit{}}) {
		t.Fatalf("a lost claim exits: %#v", effs)
	}
	g, effs = Step(g, Claimed{Won: true})
	if g.State != record.Resolve || g.Phase != Execute || g.Self.Frame != a.Stamp || g.Self.Stamp != record.Stamp(res) || effs[0] != (Pop{}) {
		t.Fatalf("the winner takes custody of the resumed agent: %+v %#v", g, effs)
	}
	if err := json.Unmarshal(effs[1].(Run).Stdin, &c); err != nil || len(c.Steps) != 1 || c.Steps[0].To != "echo" || c.Steps[0].State != record.Return || string(c.Steps[0].Value) != `"hi"` {
		t.Fatalf("the agent resumes with the step it took and what came back: %s", effs[1].(Run).Stdin)
	}

	// Its reply discharges the agent's own slot, and wakes the operator.
	final, _ := Step(Instance{Phase: Intake}, Received{Under: g.Self, In: []byte(`echo said "hi"`)})
	child, _ := Narrow(g.Self, final.R, final.Stamp)
	if final.State != record.Return || child.Frame != a.Stamp || child.Resume.Frame != operator.Frame {
		t.Fatalf("the agent's reply discharges its own frame, waited on by the operator: %+v", child)
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

	narrower := horizon.Horizon{Stamp: "g0", Reach: map[string]horizon.Surface{"agent": agentS}, Frame: "g0"}
	h, err = Narrow(narrower, invoke(narrower, "agent"), "r1")
	if err != nil || len(h.Reach) != 0 {
		t.Fatalf("a Horizon must never widen: asked for echo, the parent does not hold it: %+v %v", h, err)
	}

	forged := record.Record{State: record.Invoke, To: "agent"}
	if _, err := Narrow(horizon.Horizon{Stamp: "r1", Reach: map[string]horizon.Surface{}}, forged, "r2"); err == nil {
		t.Fatal("narrow granted a target outside the parent's reach")
	}
}

func TestNarrowAttenuatesDelegatedToolAuthority(t *testing.T) {
	full := echoS
	full.Authority.Read = []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}
	caller := agentS
	caller.Delegate = map[string]horizon.Authority{
		"echo": {Read: []horizon.PathGrant{{Path: "/private/work/one.txt", Scope: "literal"}}},
	}
	root := horizon.Horizon{Stamp: "g0", Frame: "g0", Reach: map[string]horizon.Surface{"agent": caller, "echo": full}}
	invoke := intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`)
	child, err := Narrow(root, invoke, "r1")
	if err != nil {
		t.Fatal(err)
	}
	if len(child.Surface.Authority.Read) != 0 || !reflect.DeepEqual(child.Reach["echo"].Authority, caller.Delegate["echo"]) {
		t.Fatalf("caller must receive only the attenuated echo capability, not echo's direct read grant: %+v", child)
	}
	caller.Delegate["echo"] = horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/workspace/secret", Scope: "literal"}}}
	root.Reach["agent"] = caller
	if _, err := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "r2"); err == nil {
		t.Fatal("accepted a delegated path outside the tool ceiling")
	}
}

// The machine must stay pure. It may import only the standard library's
// non-I/O packages and the pure packages of this module.
func TestMachineIsPure(t *testing.T) {
	allowed := map[string]bool{
		"encoding/json": true, "errors": true, "fmt": true, "unicode/utf8": true,
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

// An async surface is parked with its host instead of run: the host takes
// custody, and a refusal leaves custody here, as a fault on this frame.
func TestAsyncSurfaceParks(t *testing.T) {
	browserS := horizon.Surface{Host: "/tmp/park.sock", Input: echoIn, Output: text}
	root := horizon.Horizon{Stamp: "g0", Frame: "g0", Reach: map[string]horizon.Surface{"browser": browserS}}
	s, _ := Step(Instance{Phase: Intake}, Received{Under: root, In: []byte(`{"to":"browser","payload":{"text":"show me"}}`)})
	s, effs := released(s, selfOf(t, s))
	if s.Phase != Execute || !reflect.DeepEqual(effs, []Effect{Park{Host: "/tmp/park.sock", Payload: []byte(`{"text":"show me"}`)}}) {
		t.Fatalf("an async invoke parks with its host, in Execute: %s %#v", s.Phase, effs)
	}
	if _, effs := Step(s, Parked{}); !reflect.DeepEqual(effs, []Effect{Exit{}}) {
		t.Fatalf("custody taken: the instance exits: %#v", effs)
	}
	_, effs = Step(s, Parked{Err: "no host"})
	if f := effs[1].(Commit).R; f.Fault.Kind != record.Custody || !strings.Contains(f.Fault.Cause, "no host") {
		t.Fatalf("custody refused: the instance faults its own frame: %#v", effs)
	}

	// The host resumes it in Execute with the event: an ordinary handoff.
	s, effs = Step(s, Emitted{Out: []byte(`{"to":"reply","payload":"hi"}`)})
	if _, ok := effs[0].(Spawn); !ok {
		t.Fatalf("a resumed instance hands the event off like any output: %#v", effs)
	}
}
