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

	"skyra/controller/core/contract"
	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

var (
	text   = &contract.Schema{Type: "string"}
	echoIn = &contract.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*contract.Schema{"text": text}}
	echoS  = horizon.Surface{Path: "/bin/echo-target", SHA256: "e", Input: echoIn}
	agentS = horizon.Surface{Path: "/bin/agent", SHA256: "a", Input: echoIn, Continuation: true}
	both   = map[string]horizon.Surface{"echo": echoS, "agent": agentS}
	// Horizons as the emitter holds them: the operator (no surface), and a leaf.
	operator = horizon.Horizon{Stamp: "g0", Reach: both, Frame: "g0"}
	leaf     = horizon.Horizon{Stamp: "r1", Surface: &echoS, Reach: map[string]horizon.Surface{}, Frame: "r1", Resume: &operator}
)

func TestPrimaryBundleContractIsPassedToContinuation(t *testing.T) {
	surface := horizon.Surface{Input: echoIn, Continuation: true, BundlePath: "/write-bundle.json"}
	if got := against(surface, json.RawMessage(`{"text":"hi"}`)); got != echoIn {
		t.Fatalf("bundle's primary contract must be named in its envelope: %+v", got)
	}
	surface.BundlePath = ""
	if got := against(surface, json.RawMessage(`{"text":"hi"}`)); got != nil {
		t.Fatalf("ordinary single-contract actor envelope changed: %+v", got)
	}
}

// intakeRecord runs the Intake phase and returns what it commits.
func intakeRecord(h horizon.Horizon, in string) record.Record {
	_, effs := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(h), In: []byte(in)})
	return CommittedRecord(effs)
}

// The emitter supplies operands; interpret derives the state.
func TestInterpret(t *testing.T) {
	cases := []struct {
		name    string
		under   horizon.Horizon
		in      string
		state   record.State
		to      string // invoke: the capability
		payload string // the canonical payload: an invoke's input, a return's value, a fault's observation
		kind    string // fault: its kind
		cause   string // fault: a substring of its cause
	}{
		{"addressed, in reach, fits", operator, `{"to":"echo","payload":{"text":"hi"}}`, record.Invoke, "echo", `{"text":"hi"}`, "", ""},
		{"addresses its own slot", leaf, `{"to":"r1","payload":"hi"}`, record.Return, "", `"hi"`, "", ""},
		{"unaddressed text is a leaf fault, kept", leaf, `hello`, record.Fault, "", `"hello"`, record.Leaf, "addresses nothing"},
		{"unaddressed JSON is a leaf fault, kept as text", leaf, `{"text":"hi"}`, record.Fault, "", `"{\"text\":\"hi\"}"`, record.Leaf, "addresses nothing"},
		{"reply is not a name", leaf, `{"to":"reply","payload":"hi"}`, record.Fault, "", `{"payload":"hi","to":"reply"}`, record.Horizon, `"reply" is not in reach`},
		{"another frame's slot is not in reach", leaf, `{"to":"g0","payload":"hi"}`, record.Fault, "", `{"payload":"hi","to":"g0"}`, record.Horizon, `"g0" is not in reach`},
		{"a leaf addresses a capability", leaf, `{"to":"echo","payload":{"text":"x"}}`, record.Fault, "", `{"payload":{"text":"x"},"to":"echo"}`, record.Horizon, `"echo" is not in reach`},
		{"out of reach", operator, `{"to":"ghost","payload":{}}`, record.Fault, "", `{"payload":{},"to":"ghost"}`, record.Horizon, "not in reach"},
		{"the operator has no slot to return to", operator, `{"to":"g0","payload":"hi"}`, record.Fault, "", `{"payload":"hi","to":"g0"}`, record.Horizon, `"g0" is not in reach`},
		{"payload does not fit", operator, `{"to":"echo","payload":{"txt":"x"}}`, record.Fault, "", `{"payload":{"txt":"x"},"to":"echo"}`, record.Contract, `unknown field "txt"`},
		{"payload missing a field", operator, `{"to":"echo","payload":{}}`, record.Fault, "", `{"payload":{},"to":"echo"}`, record.Contract, `missing required field "text"`},
		{"a return is any JSON: its type is computed", leaf, `{"to":"r1","payload":{"any":["thing"]}}`, record.Return, "", `{"any":["thing"]}`, "", ""},
		{"address without payload", operator, `{"to":"echo"}`, record.Fault, "", `"{\"to\":\"echo\"}"`, record.Malformed, "exactly"},
		{"address with extra fields", operator, `{"to":"echo","payload":{},"x":1}`, record.Fault, "", `"{\"to\":\"echo\",\"payload\":{},\"x\":1}"`, record.Malformed, "exactly"},
		{"address is not a name", operator, `{"to":7,"payload":{}}`, record.Fault, "", `"{\"to\":7,\"payload\":{}}"`, record.Malformed, "exactly"},
		{"no output", leaf, ``, record.Fault, "", "", record.Custody, "no output"},
		{"not text", leaf, "\xff\xfe", record.Fault, "", "", record.Malformed, "UTF-8"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s, effs := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(c.under), In: []byte(c.in)})
			r := CommittedRecord(effs)
			if s.Phase != Stamping || s.Stamp() != record.Stamp(r) || s.State() != r.State {
				t.Fatalf("intake must derive the state and commit: %+v", s)
			}
			if !reflect.DeepEqual(effs, []Effect{Commit{Writes: []Write{ObjectWrite{"inputs", []byte(c.in)}, recordWrite(r)}}}) {
				t.Fatalf("intake commits exactly its input, then its record: %#v", effs)
			}
			s, effs = Step(s, created(effs))
			if s.Phase != Release || !reflect.DeepEqual(effs, []Effect{Execute{Act: SendStamp{s.Stamp()}}}) {
				t.Fatalf("then it sends its stamp and waits for the release: %s %#v", s.Phase, effs)
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

// The state picks the program: the same release runs a target for an
// invoke and settles a slot for a return or a fault.
func TestStateIsTheTopLevelBranch(t *testing.T) {
	for _, c := range []struct {
		under  horizon.Horizon
		in     string
		state  record.State
		settle bool // the first thing after the release is the discharge, not a run
	}{
		{operator, `{"to":"echo","payload":{"text":"x"}}`, record.Invoke, false},
		{leaf, `{"to":"r1","payload":"x"}`, record.Return, true},
		{operator, `{"to":"ghost","payload":{}}`, record.Fault, true},
	} {
		s, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(c.under), In: []byte(c.in)})
		if s.State() != c.state {
			t.Fatalf("%s: state %s, want %s", c.in, s.State(), c.state)
		}
		self, _ := Narrow(c.under, s.R, s.Stamp())
		_, effs := released(s, self)
		if c.settle {
			if w := writes(t, effs); len(w) != 1 || w[0] != (ResultWrite{Stamp: s.Stamp(), Frame: self.Frame}) {
				t.Fatalf("%s (%s): first a discharge: %#v", c.in, c.state, effs)
			}
		} else if a := acts(effs); len(a) != 1 {
			t.Fatalf("%s (%s): first a run: %#v", c.in, c.state, effs)
		} else if _, ok := a[0].(Run); !ok {
			t.Fatalf("%s (%s): first a run: %#v", c.in, c.state, effs)
		}
	}

	// A terminal state never executes, even if a target's output arrives.
	s, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(leaf), In: []byte(`{"to":"r1","payload":"x"}`)})
	s.Phase = Running
	_, effs := Step(s, Executed{Obs: []byte("x")})
	if ok, err := exited(effs); !ok || err == "" {
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
		Received{Under: HeldHorizon(operator), In: []byte(`{"to":"echo","payload":{"text":"hi"}}`)})
	return s
}

func selfOf(t *testing.T, s Instance) horizon.Horizon {
	t.Helper()
	h, err := Narrow(s.Under, s.R, s.Stamp())
	if err != nil {
		t.Fatal(err)
	}
	return h
}

func TestReleaseStartsExecution(t *testing.T) {
	s := awaiting(t)
	s, effs := released(s, selfOf(t, s))
	if s.Phase != Running || len(effs) != 1 || !effs[0].(Execute).Force {
		t.Fatalf("phase %s, effects %#v: one run, after everything before it is durable", s.Phase, effs)
	}
	run := effs[0].(Execute).Act.(Run)
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
	for name, c := range map[string]struct {
		self   Object
		frames int
	}{
		"for another commit":        {HeldHorizon(other), 2},
		"without its frame stack":   {HeldHorizon(good), 0},
		"missing the waiting frame": {HeldHorizon(good), 1},
	} {
		_, effs := releasedWith(s, c.self, c.frames)
		a := acts(effs)
		if ok, err := exited(effs); !ok || err == "" || len(a) != 2 || a[0].(Acknowledge).Err != err {
			t.Fatalf("%s: a bad release is refused to the parent and stops the instance without executing: %#v", name, effs)
		}
	}
}

// A return or fault discharges its slot. If a continuation waits on it, the
// instance goes on to read what that continuation still waits on; if only the
// operator waits, it exits.
func TestReleaseToSettle(t *testing.T) {
	// The operator's own emission is refused: its fault discharges frame 0,
	// which only the operator waits on.
	s, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(operator), In: []byte(`{"to":"ghost","payload":{}}`)})
	s, effs := released(s, selfOf(t, s))
	if w := writes(t, effs); w[0].(ResultWrite).Stamp != s.Stamp() {
		t.Fatalf("the fault discharges its slot: %#v", effs)
	}
	if _, effs = Step(s, created(effs)); !cleanExit(effs) {
		t.Fatalf("with only the operator waiting: %#v", effs)
	}
	s, _ = Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(leaf), In: []byte(`{"to":"r1","payload":"x"}`)})
	s, effs = released(s, selfOf(t, s))
	s, effs = Step(s, created(effs))
	want := Read{Keys: []Key{ObjectKey{"records", operator.Frame}, SlotsKey{operator.Stamp}, EmissionKey{operator.Frame, operator.Stamp}}}
	if !reflect.DeepEqual(effs, []Effect{want}) || s.Phase != Gather {
		t.Fatalf("with a continuation waiting, it reads the waiter's frame and round: %#v", effs)
	}
	// A slot already discharged is not discharged again.
	s, _ = Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(leaf), In: []byte(`{"to":"r1","payload":"x"}`)})
	s, _ = released(s, selfOf(t, s))
	if _, effs := Step(s, Committed{Outcomes: []Outcome{Lost}}); !strings.Contains(exitError(effs), "already settled") {
		t.Fatalf("a lost discharge: %#v", effs)
	}
}

func exitError(effs []Effect) string {
	_, err := exited(effs)
	return err
}

// The child re-derives the Horizon it looked up: one with the right stamp
// but more reach than its commit derives is refused, and nothing runs.
func TestReleaseMustBeWhatTheCommitDerives(t *testing.T) {
	s := awaiting(t)
	forged := selfOf(t, s)
	forged.Reach = map[string]horizon.Surface{"echo": echoS, "agent": agentS, "shell": echoS}
	_, effs := released(s, forged)
	if !strings.Contains(exitError(effs), "not what this commit derives") {
		t.Fatalf("a Horizon the commit does not derive must stop the instance: %#v", effs)
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
	s2, effs := Step(s, Executed{})
	obs := inputOf(t, effs)
	f := committedRecords(t, effs)[0]
	if !f.Supervisor || f.State != record.Fault || f.Fault.Kind != record.Custody || f.Predecessor != s.Stamp() || !strings.Contains(f.Fault.Cause, "without emitting") {
		t.Fatalf("supervisor fault: %+v", f)
	}
	if f.Input != record.Sum(obs) || record.Stamp(Observed(s.Stamp(), obs)) != record.Stamp(f) {
		t.Fatal("a supervisor fault must be derivable from its stored observation")
	}
	if w := writes(t, effs); len(w) != 3 || w[2] != (ResultWrite{Stamp: record.Stamp(f), Frame: s.Self.Frame}) {
		t.Fatalf("the fault discharges this instance's own slot in the same commit: %#v", effs)
	}
	if s2, effs = Step(s2, created(effs)); s2.Phase != Gather {
		t.Fatalf("then it wakes its waiter: %#v", effs)
	}
}

func TestHandoff(t *testing.T) {
	s := executing(t)
	out := []byte(`{"to":"` + s.Self.Frame + `","payload":"hi"}`)
	s, effs := Step(s, Executed{Obs: out})
	if !reflect.DeepEqual(effs, []Effect{Commit{Writes: []Write{horizonWrite(s.Self)}}}) {
		t.Fatalf("the Horizon the successor intakes under is committed first: %#v", effs)
	}
	s, effs = Step(s, created(effs))
	if !reflect.DeepEqual(effs, []Effect{Execute{Act: Spawn{Under: HeldHorizon(s.Self).Key, Out: out}}}) {
		t.Fatalf("the emission goes to a successor, uninterpreted: %#v", effs)
	}

	// The successor commits exactly what a real successor would.
	succ := intakeRecord(s.Self, string(out))
	s2, effs := stampedBy(s, succ)
	if succ.State != record.Return {
		t.Fatalf("the emission addressed the frame's slot: %+v", succ)
	}
	want := horizon.Horizon{Stamp: record.Stamp(succ), Reach: s.Self.Reach, Authority: s.Self.Authority, Frame: s.Self.Frame, Resume: s.Self.Resume}
	if !reflect.DeepEqual(effs, []Effect{Commit{Writes: []Write{horizonWrite(want)}}}) {
		t.Fatalf("a verified return opens no slot: only its Horizon is committed: %#v", effs)
	}
	s2, effs = Step(s2, created(effs))
	if !reflect.DeepEqual(effs, []Effect{Execute{Act: ReleaseChild{Horizon: HeldHorizon(want).Key, Stamp: record.Stamp(succ)}}, Execute{Act: Exit{}}}) {
		t.Fatalf("a verified return is released to discharge this frame's slot: %#v", effs)
	}
	if s2.Child() != record.Stamp(succ) {
		t.Fatalf("child stamp not recorded: %+v", s2)
	}
}

func TestHandoffRefusals(t *testing.T) {
	s := executing(t)
	out := []byte(`hi`)
	s, _ = emitted(s, out)
	good := intakeRecord(s.Self, string(out))

	wrongParent := good
	wrongParent.Predecessor = "someone-else"
	wrongBytes := intakeRecord(s.Self, `other`)

	for name, c := range map[string]struct {
		stamp string
		raw   []byte
		died  string
	}{
		"no record":     {stamp: "nothing"},
		"wrong parent":  {stamp: record.Stamp(wrongParent), raw: HeldRecord(wrongParent).Raw},
		"wrong bytes":   {stamp: record.Stamp(wrongBytes), raw: HeldRecord(wrongBytes).Raw},
		"died silently": {died: "EOF"},
		// A lying shell (41, V1): what it read under the stamp is not the record.
		"flipped bytes":       {stamp: record.Stamp(good), raw: flipped(HeldRecord(good)).Raw},
		"non-canonical bytes": {stamp: reencoded(HeldRecord(good)).Key, raw: reencoded(HeldRecord(good)).Raw},
		"another record":      {stamp: record.Stamp(good), raw: HeldRecord(wrongBytes).Raw},
	} {
		var k Instance
		var effs []Effect
		if c.died != "" {
			k, effs = Step(s, Executed{Err: c.died})
		} else {
			k, effs = stampedWith(s, c.stamp, c.raw)
			if a := acts(effs); len(a) != 1 || a[0] != (KillSuccessors{}) {
				t.Fatalf("%s: an unverified successor must be killed first: %#v", name, effs)
			}
			k, effs = Step(k, Executed{})
		}
		f := committedRecords(t, effs)
		if len(f) != 1 || !f[0].Supervisor {
			t.Fatalf("%s: expected a supervisor fault, got %#v", name, effs)
		}
		_, effs = Step(k, created(effs))
		for _, a := range acts(effs) {
			if _, ok := a.(ReleaseChild); ok {
				t.Fatalf("%s: released a successor that failed verification", name)
			}
		}
	}
}

// A continuation invokes echo, waits on its slot, and resumes with what came
// back, all without processes: the whole round trip is step.
func TestContinuationInvokesWaitsResumes(t *testing.T) {
	// The operator invokes the agent. It runs with a continuation: no steps yet.
	a, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(operator), In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
	a, effs := released(a, selfOf(t, a))
	var c Continuation
	if err := json.Unmarshal(runIn(t, effs).Stdin, &c); err != nil || c.Frame != a.Stamp() || string(c.Input) != `{"text":"hi"}` || len(c.Steps) != 0 {
		t.Fatalf("a continuation starts with its frame, its input and no steps: %s", runIn(t, effs).Stdin)
	}

	// It addresses echo: the successor derives an invoke, and the agent opens
	// a slot for it in its own frame, in this round, before releasing it.
	call := []byte(`{"to":"echo","payload":{"text":"hi"}}`)
	a, _ = emitted(a, call)
	b, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(a.Self), In: call})
	a, effs = stampedBy(a, b.R)
	rest := publishedRound(t, effs, call, a.Self.Stamp)
	echoSelf := a.Kids[0].Horizon
	if rest[0] != (SlotWrite{Round: a.Self.Stamp, Child: b.Stamp()}) || rest[1] != (FrameWrite{Stamp: b.Stamp()}) {
		t.Fatalf("an invoke's slot and frame are committed before it is released: %#v", rest)
	}
	if _, effs := Step(a, created(effs)); acts(effs)[0].(ReleaseChild).NewFrame != b.Stamp() {
		t.Fatalf("it is released into its own frame: %#v", effs)
	}

	// Echo returns to its slot; the return discharges it and wakes the agent's frame.
	b, _ = released(b, echoSelf)
	echoReturn := []byte(`{"to":"` + b.Self.Frame + `","payload":"hi"}`)
	b, _ = emitted(b, echoReturn)
	r, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(b.Self), In: echoReturn})
	b, _ = stampedBy(b, r.R)
	r, effs = released(r, b.Kids[0].Horizon)
	if w := writes(t, effs); len(w) != 1 || w[0].(ResultWrite).Stamp != r.Stamp() {
		t.Fatalf("the return discharges echo's slot: %#v", effs)
	}
	r, _ = Step(r, created(effs))
	if r.Phase != Gather {
		t.Fatalf("then reads the agent's round: %s", r.Phase)
	}

	agentRecord := a.R
	round := func(discharged bool) Settled {
		child := Child{Invoke: HeldRecord(b.R)}
		if discharged {
			d := HeldRecord(r.R)
			child.Discharge = &d
		}
		return Settled{Frame: HeldRecord(agentRecord), Rounds: []Round{{Stamp: a.Stamp(), Children: []Child{child}}}}
	}
	gatherRead := []Effect{Read{Keys: r.Reads}}

	// Wait for all: an open slot means no resolve, and nothing is committed.
	if _, effs := gatheredFrom(t, r, gatherRead, round(false)); !cleanExit(effs) {
		t.Fatalf("a round with an open slot must not resolve: %#v", effs)
	}
	// A lying shell: another record's bytes under the waiter's frame.
	wrong := round(true)
	wrong.Frame = Object{Key: a.Stamp(), Raw: HeldRecord(b.R).Raw}
	if _, effs := gatheredFrom(t, r, gatherRead, wrong); !strings.Contains(exitError(effs), "does not hash") {
		t.Fatalf("a foreign frame must stop the instance: %#v", effs)
	}

	// Every slot discharged: commit a resolve naming every discharge, and
	// claim in the same commit. Nothing is looked up between gather and
	// resolve (next/32).
	g, effs := gatheredFrom(t, r, gatherRead, round(true))
	w := writes(t, effs)
	if g.Phase != Claim || len(w) != 3 {
		t.Fatalf("gather resolves at once: %#v", effs)
	}
	res := committedRecords(t, effs)[0]
	if res.State != record.Resolve || res.Frame != a.Stamp() || res.Round != a.Stamp() || !reflect.DeepEqual(res.Deps, []string{r.Stamp()}) || res.Predecessor != r.Stamp() {
		t.Fatalf("resolve record: %+v", res)
	}
	if derived, _ := Resolved(r.Stamp(), inputOf(t, effs)); record.Stamp(derived) != record.Stamp(res) {
		t.Fatal("a resolve must be derivable from its stored observation")
	}
	if w[2] != (ClaimWrite{Round: a.Stamp(), Stamp: record.Stamp(res), Frame: a.Stamp()}) {
		t.Fatalf("then the fan-in claim: %#v", effs)
	}

	// A lost claim exits; the winner resumes the agent with what echo said.
	if _, effs := Step(g, Committed{Outcomes: []Outcome{Created, Created, Lost}}); !cleanExit(effs) {
		t.Fatalf("a lost claim exits: %#v", effs)
	}
	g, effs = Step(g, created(effs))
	if g.State() != record.Resolve || g.Phase != Running || g.Self.Frame != a.Stamp() || g.Self.Stamp != record.Stamp(res) || acts(effs)[0] != (Pop{}) {
		t.Fatalf("the winner takes custody of the resumed agent: %+v %#v", g, effs)
	}
	if err := json.Unmarshal(runIn(t, effs[1:]).Stdin, &c); err != nil || len(c.Steps) != 1 || c.Steps[0].To != "echo" || c.Steps[0].State != record.Return || string(c.Steps[0].Value) != `"hi"` {
		t.Fatalf("the agent resumes with the step it took and what came back: %s", runIn(t, effs[1:]).Stdin)
	}

	// Its return discharges the agent's own slot, and wakes the operator.
	final, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(g.Self), In: []byte(`{"to":"` + a.Stamp() + `","payload":"echo said \"hi\""}`)})
	child, _ := Narrow(g.Self, final.R, final.Stamp())
	if final.State() != record.Return || child.Frame != a.Stamp() || child.Resume.Frame != operator.Frame {
		t.Fatalf("the agent's return discharges its own frame, waited on by the operator: %+v", child)
	}
}

// runIn is the one Run effs holds.
func runIn(t *testing.T, effs []Effect) Run {
	t.Helper()
	for _, a := range acts(effs) {
		if r, ok := a.(Run); ok {
			return r
		}
	}
	t.Fatalf("no run: %#v", effs)
	return Run{}
}

func TestOutOfPhase(t *testing.T) {
	_, effs := Step(Instance{Phase: Intake}, Executed{Obs: []byte("x")})
	if !strings.Contains(exitError(effs), "not valid in phase") {
		t.Fatalf("an event in the wrong phase must stop the instance: %#v", effs)
	}
	s := awaiting(t)
	if _, effs := Step(s, ReadResult{Found: [][]byte{nil}}); !strings.Contains(exitError(effs), "answered with 1") {
		t.Fatalf("a read answered with more than it asked stops the instance: %#v", effs)
	}
	if _, effs := Step(s, Committed{Outcomes: []Outcome{Created, Conflict}}); !strings.Contains(exitError(effs), "integrity") {
		t.Fatalf("a conflict is an integrity fault: %#v", effs)
	}
}

// The phases routes had are gone (next/32, H1): where the route runner's
// decision, system.route's rounds and its retests were taken, Step now
// refuses them like any event out of phase.
func TestRemovedRoutePhasesAreRefused(t *testing.T) {
	for name, c := range map[string]struct {
		s Instance
		e Event
	}{
		"a decision after gather":           {Instance{R: record.Record{State: record.Return}, Phase: Claim}, Executed{Obs: []byte("x")}},
		"a route looked up after gather":    {Instance{R: record.Record{State: record.Return}, Phase: Claim}, ReadResult{}},
		"rounds read by an invoke's facet":  {Instance{R: record.Record{State: record.Invoke}, Phase: Running}, ReadResult{}},
		"a retest of a recorded run":        {Instance{R: record.Record{State: record.Invoke}, Phase: Handoff}, ReadResult{}},
		"a return's gather in a dead phase": {Instance{R: record.Record{State: record.Fault}, Phase: Done}, Executed{Obs: []byte("x")}},
	} {
		if _, effs := Step(c.s, c.e); len(effs) != 1 || !strings.Contains(exitError(effs), "not valid in phase") {
			t.Errorf("%s: %#v", name, effs)
		}
	}
}

// Every call shares its caller's Horizon (next/24): the same reach and the
// same grants, whatever the callee is. Only the frame is new.
func TestNarrowSharesTheCallersHorizon(t *testing.T) {
	invoke := func(under horizon.Horizon, target string) record.Record {
		return intakeRecord(under, `{"to":"`+target+`","payload":{"text":"x"}}`)
	}
	root := operator
	root.Authority = horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}}

	for _, target := range []string{"echo", "agent"} {
		h, err := Narrow(root, invoke(root, target), "r1")
		if err != nil || h.Surface.Path != root.Reach[target].Path || !reflect.DeepEqual(h.Reach, root.Reach) || !reflect.DeepEqual(h.Authority, root.Authority) {
			t.Fatalf("%s must run under its caller's reach and grants: %+v %v", target, h, err)
		}
		if h.Frame != "r1" || h.Resume == nil || h.Resume.Stamp != root.Stamp {
			t.Fatalf("%s gets a new frame, waited on by its caller: %+v", target, h)
		}
	}

	// A return or fault runs nothing: it discharges the caller's slot, under
	// the same Horizon.
	leafUnder, _ := Narrow(root, invoke(root, "echo"), "r1")
	ret := intakeRecord(leafUnder, `{"to":"r1","payload":"x"}`)
	h, err := Narrow(leafUnder, ret, "r2")
	if err != nil || h.Frame != "r1" || h.Surface != nil || !reflect.DeepEqual(h.Authority, root.Authority) {
		t.Fatalf("a return discharges the frame it ran in: %+v %v", h, err)
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
		"encoding/json": true, "errors": true, "fmt": true, "unicode/utf8": true, "slices": true, "strings": true,
		"skyra/controller/core/contract": true,
		"skyra/controller/core/horizon":  true,
		"skyra/controller/core/record":   true,
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
	browserS := horizon.Surface{Host: "/tmp/park.sock", Input: echoIn}
	root := horizon.Horizon{Stamp: "g0", Frame: "g0", Reach: map[string]horizon.Surface{"browser": browserS}}
	s, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(root), In: []byte(`{"to":"browser","payload":{"text":"show me"}}`)})
	s, effs := released(s, selfOf(t, s))
	want := []Effect{Execute{Act: Park{Host: "/tmp/park.sock", Payload: []byte(`{"text":"show me"}`), Stamp: s.Stamp(), Frames: 2}, Force: true}}
	if s.Phase != Parking || !reflect.DeepEqual(effs, want) {
		t.Fatalf("an async invoke parks with its host: %s %#v", s.Phase, effs)
	}
	if _, effs := Step(s, Executed{}); !cleanExit(effs) {
		t.Fatalf("custody taken: the instance exits: %#v", effs)
	}
	_, effs = Step(s, Executed{Err: "no host"})
	if f := committedRecords(t, effs)[0]; f.Fault.Kind != record.Custody || !strings.Contains(f.Fault.Cause, "no host") {
		t.Fatalf("custody refused: the instance faults its own frame: %#v", effs)
	}

	// The host resumes it running, with the event: an ordinary handoff.
	s.Phase = Running
	_, effs = emitted(s, []byte(`{"to":"`+s.Self.Frame+`","payload":"hi"}`))
	spawned(t, effs)
}

func TestRunTellsATargetItsSlotAndReadGrants(t *testing.T) {
	s := horizon.Surface{Path: "/bin/tool"}
	if env := runOf(horizon.Horizon{Surface: &s, Frame: "f1"}, nil, "/floor").Env; !reflect.DeepEqual(env, []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=f1"}) {
		t.Fatalf("a target is told its slot; no read grant, no SKYRA_READ: %v", env)
	}
	grants := horizon.Authority{Read: []horizon.PathGrant{{Path: "/work", Scope: "subpath"}, {Path: "/etc/x", Scope: "literal"}}}
	env := runOf(horizon.Horizon{Surface: &s, Authority: grants}, nil, "/floor").Env
	if env[len(env)-1] != "SKYRA_READ=/work:/etc/x" {
		t.Fatalf("%v", env)
	}
}

// fanOut has the agent emit several instructions at once and returns the
// agent's instance after its handoff, each successor's intake, and the
// writes of the round's commit after its emission.
func fanOut(t *testing.T, instructions ...string) (Instance, []Instance, []Write) {
	t.Helper()
	a, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(operator), In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
	a, _ = released(a, selfOf(t, a))
	emission := []byte("[" + strings.Join(instructions, ",") + "]")
	a, effs := emitted(a, emission)
	var kids []Instance
	for i, in := range instructions {
		if sp := spawned(t, effs); string(sp.Out) != in {
			t.Fatalf("instruction %d is spawned with exactly its bytes: %#v", i, effs)
		}
		k, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(a.Self), In: []byte(in)})
		kids = append(kids, k)
		a, effs = stampedBy(a, k.R)
		if i < len(instructions)-1 {
			spawned(t, effs) // a successor is never released before every instruction is stamped
		}
	}
	return a, kids, publishedRound(t, effs, a.Out, a.Self.Stamp)
}

// releases answers the round's commit and returns its releases, which must
// be the last acts before the emitter exits.
func releases(t *testing.T, a Instance, w []Write) []ReleaseChild {
	t.Helper()
	_, effs := Step(a, Committed{Outcomes: make([]Outcome, len(w)+2)})
	acted := acts(effs)
	if acted[len(acted)-1] != (Exit{}) {
		t.Fatalf("the emitter exits after releasing: %#v", effs)
	}
	var out []ReleaseChild
	for _, x := range acted[:len(acted)-1] {
		out = append(out, x.(ReleaseChild))
	}
	return out
}

// Fan-out (next/32, H2): one emission, N successors, each taking one
// instruction; every slot is opened before any successor is released, and
// the waiter resumes once, with every step in the order it was emitted.
func TestFanOutResumesTheWaiterOnceWithEveryStepInOrder(t *testing.T) {
	first, second := `{"to":"echo","payload":{"text":"b"}}`, `{"to":"echo","payload":{"text":"a"}}`
	a, kids, w := fanOut(t, first, second)
	want := []Write{
		SlotWrite{Round: a.Self.Stamp, Child: kids[0].Stamp(), Pos: 0},
		SlotWrite{Round: a.Self.Stamp, Child: kids[1].Stamp(), Pos: 1},
	}
	if len(w) != 6 || !reflect.DeepEqual(w[:2], want) {
		t.Fatalf("every slot opens in the same commit, before any release: %#v", w)
	}
	rel := releases(t, a, w)
	for i := range kids {
		self := a.Kids[i].Horizon
		if rel[i].Index != i || rel[i].NewFrame != kids[i].Stamp() || self.Frame != kids[i].Stamp() || self.Resume.Stamp != a.Self.Stamp {
			t.Fatalf("successor %d is released into a frame of its own, in the agent's round: %#v", i, rel[i])
		}
	}

	// Both return. Gather sees the whole round; only once both are
	// discharged does it resolve, naming both in instruction order.
	var returns []Instance
	var children []Child
	for i, k := range kids {
		self := a.Kids[i].Horizon
		k, _ = released(k, self)
		out := []byte(`{"to":"` + self.Frame + `","payload":"` + []string{"b", "a"}[i] + `"}`)
		k, _ = emitted(k, out)
		r, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(k.Self), In: out})
		k, _ = stampedBy(k, r.R)
		r, effs := released(r, k.Kids[0].Horizon)
		r, _ = Step(r, created(effs))
		returns = append(returns, r)
		d := HeldRecord(r.R)
		children = append(children, Child{Invoke: HeldRecord(kids[i].R), Discharge: &d})
	}
	gatherRead := func(r Instance) []Effect { return []Effect{Read{Keys: r.Reads}} }
	open := Settled{Frame: HeldRecord(a.R), Rounds: []Round{{Stamp: a.Self.Stamp, Children: []Child{children[0], {Invoke: HeldRecord(kids[1].R)}}}}}
	if _, effs := gatheredFrom(t, returns[0], gatherRead(returns[0]), open); !cleanExit(effs) {
		t.Fatalf("a round with an open slot does not resolve: %#v", effs)
	}
	full := Settled{Frame: HeldRecord(a.R), Rounds: []Round{{Stamp: a.Self.Stamp, Children: children}}}
	g, effs := gatheredFrom(t, returns[1], gatherRead(returns[1]), full)
	res := committedRecords(t, effs)[0]
	if !reflect.DeepEqual(res.Deps, []string{returns[0].Stamp(), returns[1].Stamp()}) {
		t.Fatalf("the resolve names every discharge, in instruction order: %+v", res)
	}
	_, effs = Step(g, created(effs))
	var c Continuation
	if err := json.Unmarshal(runIn(t, effs).Stdin, &c); err != nil || len(c.Steps) != 2 || string(c.Steps[0].Value) != `"b"` || string(c.Steps[1].Value) != `"a"` {
		t.Fatalf("the agent resumes once, with both steps in the order it emitted them: %s", runIn(t, effs).Stdin)
	}
}

// One instruction out of reach faults that one, not its sibling: the fault
// holds a slot of its own in the round and discharges it.
func TestFanOutFaultsOnlyTheInstructionOutOfReach(t *testing.T) {
	a, kids, w := fanOut(t, `{"to":"echo","payload":{"text":"a"}}`, `{"to":"shell","payload":{"text":"a"}}`)
	if kids[0].State() != record.Invoke || kids[1].State() != record.Fault || kids[1].R.Fault.Kind != record.Horizon {
		t.Fatalf("only the instruction out of reach faults: %+v %+v", kids[0].R, kids[1].R)
	}
	if len(w) != 6 || w[1] != (SlotWrite{Round: a.Self.Stamp, Child: kids[1].Stamp(), Pos: 1}) {
		t.Fatalf("the fault holds a slot of its own: %#v", w)
	}
	self := a.Kids[1].Horizon
	f, effs := released(kids[1], self)
	if w := writes(t, effs); w[0].(ResultWrite).Stamp != kids[1].Stamp() || f.Self.Resume.Frame != a.Self.Frame {
		t.Fatalf("the fault discharges its own slot: %#v", effs)
	}
	if f, _ = Step(f, created(effs)); f.Phase != Gather {
		t.Fatalf("then reads the agent's round: %s", f.Phase)
	}
	// Alone, the same fault discharges the emitter's own slot, as always.
	alone, _ := Narrow(a.Self, kids[1].R, kids[1].Stamp())
	if alone.Frame != a.Self.Frame {
		t.Fatalf("a lone fault discharges the emitter's slot: %+v", alone)
	}
}

// What cannot share an emission is refused before any successor is
// released: a return, and an invoke that changes the waiter's Horizon.
func TestFanOutRefusesAReturnAndANarrowing(t *testing.T) {
	room := operator
	room.Reach = map[string]horizon.Surface{"echo": echoS, "agent": agentS}
	room.Reach[System] = Syscalls()
	echo := `{"to":"echo","payload":{"text":"a"}}`
	for name, other := range map[string]string{
		"a return":    `{"to":"FRAME","payload":"done"}`,
		"a narrowing": `{"to":"system","payload":{"narrow":{"actor":"echo"}}}`,
	} {
		ag, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(room), In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
		self, _ := Narrow(room, ag.R, ag.Stamp())
		ag, _ = released(ag, self)
		other = strings.ReplaceAll(other, "FRAME", ag.Self.Frame)
		ag, effs := emitted(ag, []byte("["+echo+","+other+"]"))
		for _, in := range []string{echo, other} {
			sp := spawned(t, effs)
			if string(sp.Out) != in {
				t.Fatalf("%s: spawned %s", name, sp.Out)
			}
			k, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(ag.Self), In: sp.Out})
			ag, effs = stampedBy(ag, k.R)
		}
		if a := acts(effs); len(a) != 1 || a[0] != (KillSuccessors{}) {
			t.Fatalf("%s: every successor is killed: %#v", name, effs)
		}
		ag, effs = Step(ag, Executed{})
		_, effs = Step(ag, created(effs))
		for _, a := range acts(effs) {
			if _, ok := a.(ReleaseChild); ok {
				t.Fatalf("%s: nothing is released: %#v", name, effs)
			}
		}
	}
}

// An emission that is a JSON array is that many instructions; anything else
// is one. An empty array, or too many, is no emission at all.
func TestInstructions(t *testing.T) {
	for in, n := range map[string]int{
		`{"to":"echo","payload":{}}`:   1,
		`[{"to":"a"}, 7 ,"x"]`:         3,
		`[{"to":"echo","payload":{}}]`: 1,
		`[not json`:                    1,
		`hello`:                        1,
	} {
		outs, err := Instructions([]byte(in))
		if err != nil || len(outs) != n {
			t.Errorf("%s: %d instructions, %v", in, len(outs), err)
		}
	}
	if outs, _ := Instructions([]byte(`[{"to": "a"},  "x"]`)); string(outs[0]) != `{"to": "a"}` || string(outs[1]) != `"x"` {
		t.Fatalf("each instruction is its exact bytes: %q", outs)
	}
	if _, err := Instructions([]byte(`[]`)); err == nil {
		t.Fatal("an empty array is no instruction")
	}
	if _, err := Instructions([]byte(`[{"to":"a"},{"to":"a"}]`)); err == nil {
		t.Fatal("the same instruction twice would be one record")
	}
	if _, err := Instructions([]byte(`[{"to":"a"},  {"to":"a"} ]`)); err == nil {
		t.Fatal("whitespace outside an instruction is not part of it: still the same")
	}
	if outs, err := Instructions([]byte(`[{"to":"a"},{"to":"a" }]`)); err != nil || len(outs) != 2 {
		t.Fatal("instructions are compared as bytes: whitespace inside one makes two records")
	}
	if _, err := Instructions([]byte("[" + strings.Repeat("1,", MaxFanOut) + "1]")); err == nil {
		t.Fatal("more than MaxFanOut instructions")
	}
}

func publishedRound(t *testing.T, effs []Effect, out []byte, round string) []Write {
	t.Helper()
	w := writes(t, effs)
	if len(w) < 2 || !reflect.DeepEqual(w[0], ObjectWrite{"inputs", out}) || w[1] != (EmissionWrite{Round: round, Hash: record.Sum(out)}) {
		t.Fatalf("complete round must be committed before slots and releases: %#v", effs)
	}
	return w[2:]
}

func TestContinuationIsToldItsCaller(t *testing.T) {
	stdinOf := func(effs []Effect) Continuation {
		t.Helper()
		var c Continuation
		if err := json.Unmarshal(runIn(t, effs).Stdin, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	a, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(operator), In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
	a, effs := released(a, selfOf(t, a))
	if c := stdinOf(effs); c.Caller != "" {
		t.Fatalf("the operator has no name, got caller %q", c.Caller)
	}
	// The agent invokes a continuation: that child is told the agent called it.
	call := []byte(`{"to":"agent","payload":{"text":"again"}}`)
	a, _ = emitted(a, call)
	b, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(a.Self), In: call})
	a, effs = stampedBy(a, b.R)
	publishedRound(t, effs, call, a.Self.Stamp)
	_, effs = released(b, a.Kids[0].Horizon)
	if c := stdinOf(effs); c.Caller != "agent" {
		t.Fatalf("caller is derived from the Horizon, got %q", c.Caller)
	}
	// A payload cannot claim a caller: the field is the kernel's, outside
	// input, and a strict contract refuses an unknown one at intake.
	forged, _ := Step(Instance{Phase: Intake}, Received{Under: HeldHorizon(operator), In: []byte(`{"to":"agent","payload":{"text":"x","caller":"user"}}`)})
	if forged.R.State != record.Fault || forged.R.Fault.Kind != record.Contract {
		t.Fatalf("a payload naming a caller was not refused: %+v", forged.R)
	}
}

func TestContinuationReportsCurrentWorkspaceAndScratch(t *testing.T) {
	// The workspace is the one named, not the first read grant.
	h := horizon.Horizon{Authority: horizon.Authority{Read: []horizon.PathGrant{{Path: "/cache", Scope: "subpath"}, {Path: "/workspace", Scope: "subpath"}}, Workspace: "/workspace", Scratch: "/notes"}}
	var c Continuation
	if e := json.Unmarshal(continuation(record.Record{Payload: json.RawMessage(`{}`)}, nil, h), &c); e != nil {
		t.Fatal(e)
	}
	if c.Workspace != "/workspace" || c.Scratch != "/notes" {
		t.Fatal(c)
	}
}

// flipped is an object whose bytes were changed after it was read: it no
// longer hashes to its key.
func flipped(o Object) Object {
	b := append([]byte(nil), o.Raw...)
	b[len(b)/2] ^= 1
	return Object{Key: o.Key, Raw: b}
}

// reencoded is the same value in another encoding, under the hash of those
// bytes: it hashes to its key, but it is not the canonical form.
func reencoded(o Object) Object {
	var v map[string]any
	_ = json.Unmarshal(o.Raw, &v)
	b, _ := json.MarshalIndent(v, "", " ")
	return Object{Key: record.Sum(b), Raw: b}
}

// A lying shell (41, V1): every content-addressed read is verified by the
// core, so bytes that do not hash to their key, or are not the canonical
// encoding, stop the instance before anything acts on them.
func TestTheShellCannotPassOffContentAddressedBytes(t *testing.T) {
	refused := func(name string, effs []Effect) {
		t.Helper()
		if ok, err := exited(effs); !ok || err == "" || len(effs) != 1 {
			t.Fatalf("%s: a lying read must stop the instance: %#v", name, effs)
		}
	}
	in := []byte(`{"to":"echo","payload":{"text":"hi"}}`)
	for name, under := range map[string]Object{
		"flipped":       flipped(HeldHorizon(operator)),
		"non-canonical": reencoded(HeldHorizon(operator)),
		"missing":       {Key: HeldHorizon(operator).Key},
	} {
		_, effs := Step(Instance{Phase: Intake}, Received{Under: under, In: in})
		refused("intake under a "+name+" Horizon", effs)
	}

	s := awaiting(t)
	good := HeldHorizon(selfOf(t, s))
	for name, self := range map[string]Object{
		"flipped":       flipped(good),
		"non-canonical": reencoded(good),
		"missing":       {Key: good.Key},
	} {
		_, effs := releasedWith(s, self, 2)
		if a := acts(effs); len(a) != 2 || a[0].(Acknowledge).Err == "" {
			t.Fatalf("a %s release is refused to the parent: %#v", name, effs)
		}
		if _, err := exited(effs); err == "" {
			t.Fatalf("a %s release must stop the instance: %#v", name, effs)
		}
	}
}

// Replay's derivation is intake's: the record Intaken derives is the one
// intake commits.
func TestIntakenIsWhatIntakeCommits(t *testing.T) {
	for _, in := range []string{`{"to":"echo","payload":{"text":"hi"}}`, `{"to":"ghost","payload":{}}`, `hello`} {
		if a, b := record.Stamp(Intaken(operator, []byte(in))), record.Stamp(intakeRecord(operator, in)); a != b {
			t.Fatalf("%s: replay and intake derive different records", in)
		}
	}
}

// A narrowing that drops the read grant under the workspace drops the
// workspace with it; one that keeps the grant keeps the workspace, and none
// can name another.
func TestNarrowingKeepsTheWorkspaceOnlyWithItsGrant(t *testing.T) {
	h := horizon.Horizon{Reach: map[string]horizon.Surface{"echo": echoS}, Authority: horizon.Authority{
		Read: []horizon.PathGrant{{Path: "/cache", Scope: "subpath"}, {Path: "/work", Scope: "subpath"}}, Workspace: "/work"}}
	for payload, want := range map[string]string{
		`{"actor":"echo"}`:                    "/work",
		`{"actor":"echo","without":["read"]}`: "",
		`{"actor":"echo","authority":{"read":[{"path":"/cache","scope":"subpath"}]}}`:                    "",
		`{"actor":"echo","authority":{"read":[{"path":"/work","scope":"subpath"}],"workspace":"/work"}}`: "/work",
	} {
		g, err := narrowing(h, json.RawMessage(payload))
		if err != nil || g.Authority.Workspace != want {
			t.Errorf("%s: workspace %q, want %q (%v)", payload, g.Authority.Workspace, want, err)
		}
	}
	if _, err := narrowing(h, json.RawMessage(`{"actor":"echo","authority":{"read":[{"path":"/work","scope":"subpath"}],"workspace":"/elsewhere"}}`)); err == nil {
		t.Fatal("a narrowing named another workspace")
	}
}
