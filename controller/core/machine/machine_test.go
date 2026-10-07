package machine

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"slices"
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
	_, effs := Step(Instance{Phase: Intake}, Received{Under: h, In: []byte(in)})
	return Committed(effs)
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
		{leaf, `{"to":"r1","payload":"x"}`, record.Return, Settle{}},
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
	s, _ := Step(Instance{Phase: Intake}, Received{Under: leaf, In: []byte(`{"to":"r1","payload":"x"}`)})
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

// A connection is an edge between two actors (proposal 38): the adapter that
// invoked a connection surface, and whoever invoked the adapter. Its socket is
// derived from those names and the floor, never from a frame, so every turn
// reaches the same resident; and a connection surface is granted exactly it.
func TestConnectionsAreEdges(t *testing.T) {
	io := horizon.Surface{Path: "/bin/codexio", SHA256: "h", Connection: true}
	under := func(frame, caller string) horizon.Horizon {
		return horizon.Horizon{Surface: &io, Name: "codex.exchange", Frame: frame,
			Resume: &horizon.Horizon{Name: "codex", Frame: "adapter-" + frame, Resume: &horizon.Horizon{Name: caller}}}
	}
	edge := func(h horizon.Horizon, floor string) string {
		t.Helper()
		socket, err := edgeSocket(h, floor, "/private/tmp/skyra-501")
		if err != nil {
			t.Fatal(err)
		}
		return socket
	}
	first := edge(under("turn-1", "interpreter"), "/c1/.skyra")
	if second := edge(under("turn-2", "interpreter"), "/c1/.skyra"); second != first {
		t.Fatalf("two turns of one edge reach different residents: %s %s", first, second)
	}
	if edge(under("turn-1", "user"), "/c1/.skyra") == first || edge(under("turn-1", "interpreter"), "/c2/.skyra") == first {
		t.Fatal("another caller, or another lineage, shares the edge")
	}
	if !strings.HasPrefix(first, "/private/tmp/skyra-501/") || len(first) > 100 {
		t.Fatalf("socket path: %s", first)
	}
	run := runOf(under("turn-1", "interpreter"), nil, "/c1/.skyra", first)
	if run.Resident != first || !slices.Contains(run.Env, "SKYRA_RESIDENT="+first) || !strings.Contains(run.Profile, first) {
		t.Fatalf("a connection surface is granted its edge's socket: %+v", run)
	}
	if strings.Contains(runOf(under("turn-1", "interpreter"), nil, "/c1/.skyra", "").Profile, "/private/tmp/skyra-501") {
		t.Fatal("a surface with no edge was granted the socket directory")
	}
	orphan := horizon.Horizon{Surface: &io, Name: "codex.exchange", Resume: &horizon.Horizon{}}
	if _, err := edgeSocket(orphan, "/c1/.skyra", "/private/tmp/skyra-501"); err == nil {
		t.Fatal("a connection surface not invoked by an adapter has no edge")
	}
	if _, err := edgeSocket(under("turn-1", "interpreter"), "/c1/.skyra", ""); err == nil {
		t.Fatal("no socket directory was accepted")
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
	// The operator's own emission is refused: its fault discharges frame 0,
	// which only the operator waits on.
	s, _ := Step(Instance{Phase: Intake}, Received{Under: operator, In: []byte(`{"to":"ghost","payload":{}}`)})
	s, effs := released(s, selfOf(t, s))
	if !reflect.DeepEqual(effs, []Effect{Settle{s.Stamp}, Exit{}}) {
		t.Fatalf("with only the operator waiting: %#v", effs)
	}
	s, _ = Step(Instance{Phase: Intake}, Received{Under: leaf, In: []byte(`{"to":"r1","payload":"x"}`)})
	s, effs = released(s, selfOf(t, s))
	if !reflect.DeepEqual(effs, []Effect{Settle{s.Stamp}, GatherWaiter{}}) || s.Phase != Gather {
		t.Fatalf("with a continuation waiting: %#v", effs)
	}
}

// The child re-derives the Horizon it looked up: one with the right stamp
// but more reach than its commit derives is refused, and nothing runs.
func TestReleaseMustBeWhatTheCommitDerives(t *testing.T) {
	s := awaiting(t)
	forged := selfOf(t, s)
	forged.Reach = map[string]horizon.Surface{"echo": echoS, "agent": agentS, "shell": echoS}
	_, effs := released(s, forged)
	if x, ok := effs[0].(Exit); !ok || !strings.Contains(x.Err, "not what this commit derives") {
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
	out := []byte(`{"to":"` + s.Self.Frame + `","payload":"hi"}`)
	s, effs := Step(s, Emitted{Out: out})
	if !reflect.DeepEqual(effs, []Effect{Spawn{Under: s.Self, Out: out}}) {
		t.Fatalf("the emission goes to a successor, uninterpreted: %#v", effs)
	}

	// The successor commits exactly what a real successor would.
	succ := intakeRecord(s.Self, string(out))
	s2, effs := Step(s, Stamped{Stamp: record.Stamp(succ), R: &succ})
	if succ.State != record.Return {
		t.Fatalf("the emission addressed the frame's slot: %+v", succ)
	}
	want := horizon.Horizon{Stamp: record.Stamp(succ), Reach: s.Self.Reach, Authority: s.Self.Authority, Frame: s.Self.Frame, Resume: s.Self.Resume}
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
	effs = publishedRound(t, effs, call, a.Self.Stamp)
	if effs[0] != (OpenDep{Round: a.Self.Stamp, Child: b.Stamp}) || !effs[1].(SendHorizon).NewFrame {
		t.Fatalf("an invoke's slot is opened before it is released: %#v", effs)
	}

	// Echo returns to its slot; the return discharges it and wakes the agent's frame.
	b, _ = released(b, effs[1].(SendHorizon).Child)
	echoReturn := []byte(`{"to":"` + b.Self.Frame + `","payload":"hi"}`)
	b, _ = Step(b, Emitted{Out: echoReturn})
	r, _ := Step(Instance{Phase: Intake}, Received{Under: b.Self, In: echoReturn})
	_, effs = Step(b, Stamped{Stamp: r.Stamp, R: &r.R})
	r, effs = released(r, effs[0].(SendHorizon).Child)
	if !reflect.DeepEqual(effs, []Effect{Settle{r.Stamp}, GatherWaiter{}}) {
		t.Fatalf("the return discharges echo's slot, then reads the agent's round: %#v", effs)
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

	// Every slot discharged: commit a resolve naming every discharge, then
	// claim. Nothing is looked up between gather and resolve (next/32).
	g, effs := Step(r, round(true))
	if g.Phase != Claim || len(effs) != 3 {
		t.Fatalf("gather resolves at once: %#v", effs)
	}
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

	// Its return discharges the agent's own slot, and wakes the operator.
	final, _ := Step(Instance{Phase: Intake}, Received{Under: g.Self, In: []byte(`{"to":"` + a.Stamp + `","payload":"echo said \"hi\""}`)})
	child, _ := Narrow(g.Self, final.R, final.Stamp)
	if final.State != record.Return || child.Frame != a.Stamp || child.Resume.Frame != operator.Frame {
		t.Fatalf("the agent's return discharges its own frame, waited on by the operator: %+v", child)
	}
}

func TestOutOfPhase(t *testing.T) {
	_, effs := Step(Instance{Phase: Intake}, Emitted{Out: []byte("x")})
	if x, ok := effs[0].(Exit); !ok || !strings.Contains(x.Err, "not valid in phase") {
		t.Fatalf("an event in the wrong phase must stop the instance: %#v", effs)
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
		"a decision after gather":           {Instance{State: record.Return, Phase: Claim}, Emitted{Out: []byte("x")}},
		"a route looked up after gather":    {Instance{State: record.Return, Phase: Claim}, Settled{}},
		"rounds read by an invoke's facet":  {Instance{State: record.Invoke, Phase: Execute}, Settled{}},
		"a retest of a recorded run":        {Instance{State: record.Invoke, Phase: Handoff}, Emitted{Out: []byte("x")}},
		"a return's gather in a dead phase": {Instance{State: record.Fault, Phase: Done}, Emitted{Out: []byte("x")}},
	} {
		if _, effs := Step(c.s, c.e); len(effs) != 1 || !strings.Contains(effs[0].(Exit).Err, "not valid in phase") {
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
	s, effs = Step(s, Emitted{Out: []byte(`{"to":"` + s.Self.Frame + `","payload":"hi"}`)})
	if _, ok := effs[0].(Spawn); !ok {
		t.Fatalf("a resumed instance hands the event off like any output: %#v", effs)
	}
}

func TestRunTellsATargetItsSlotAndReadGrants(t *testing.T) {
	s := horizon.Surface{Path: "/bin/tool"}
	if env := runOf(horizon.Horizon{Surface: &s, Frame: "f1"}, nil, "/floor", "").Env; !reflect.DeepEqual(env, []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=f1"}) {
		t.Fatalf("a target is told its slot; no read grant, no SKYRA_READ: %v", env)
	}
	grants := horizon.Authority{Read: []horizon.PathGrant{{Path: "/work", Scope: "subpath"}, {Path: "/etc/x", Scope: "literal"}}}
	env := runOf(horizon.Horizon{Surface: &s, Authority: grants}, nil, "/floor", "").Env
	if env[len(env)-1] != "SKYRA_READ=/work:/etc/x" {
		t.Fatalf("%v", env)
	}
}

// fanOut has the agent emit several instructions at once and returns the
// agent's instance after its handoff, each successor's intake, and the
// handoff's effects.
func fanOut(t *testing.T, instructions ...string) (Instance, []Instance, []Effect) {
	t.Helper()
	a, _ := Step(Instance{Phase: Intake}, Received{Under: operator, In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
	a, _ = released(a, selfOf(t, a))
	emission := []byte("[" + strings.Join(instructions, ",") + "]")
	a, effs := Step(a, Emitted{Out: emission})
	var kids []Instance
	for i, in := range instructions {
		spawn, ok := effs[0].(Spawn)
		if !ok || string(spawn.Out) != in {
			t.Fatalf("instruction %d is spawned with exactly its bytes: %#v", i, effs)
		}
		k, _ := Step(Instance{Phase: Intake}, Received{Under: a.Self, In: spawn.Out})
		kids = append(kids, k)
		a, effs = Step(a, Stamped{Stamp: k.Stamp, R: &k.R})
		if i < len(instructions)-1 {
			for _, e := range effs {
				if _, released := e.(SendHorizon); released {
					t.Fatalf("a successor is released before every instruction is stamped: %#v", effs)
				}
			}
		}
	}
	effs = publishedRound(t, effs, a.Out, a.Self.Stamp)
	return a, kids, effs
}

// Fan-out (next/32, H2): one emission, N successors, each taking one
// instruction; every slot is opened before any successor is released, and
// the waiter resumes once, with every step in the order it was emitted.
func TestFanOutResumesTheWaiterOnceWithEveryStepInOrder(t *testing.T) {
	first, second := `{"to":"echo","payload":{"text":"b"}}`, `{"to":"echo","payload":{"text":"a"}}`
	a, kids, effs := fanOut(t, first, second)
	want := []Effect{
		OpenDep{Round: a.Self.Stamp, Child: kids[0].Stamp, Pos: 0},
		OpenDep{Round: a.Self.Stamp, Child: kids[1].Stamp, Pos: 1},
	}
	if len(effs) != 5 || !reflect.DeepEqual(effs[:2], want) {
		t.Fatalf("every slot opens before any release: %#v", effs)
	}
	for i := range kids {
		sh := effs[2+i].(SendHorizon)
		if sh.Index != i || !sh.NewFrame || sh.Child.Frame != kids[i].Stamp || sh.Child.Resume.Stamp != a.Self.Stamp {
			t.Fatalf("successor %d is released into a frame of its own, in the agent's round: %#v", i, sh)
		}
	}
	if effs[4] != (Exit{}) {
		t.Fatalf("then the emitter exits: %#v", effs)
	}

	// Both return. Gather sees the whole round; only once both are
	// discharged does it resolve, naming both in instruction order.
	var returns []Instance
	var children []Child
	for i, k := range kids {
		self := effs[2+i].(SendHorizon).Child
		k, _ = released(k, self)
		out := []byte(`{"to":"` + self.Frame + `","payload":"` + []string{"b", "a"}[i] + `"}`)
		k, _ = Step(k, Emitted{Out: out})
		r, _ := Step(Instance{Phase: Intake}, Received{Under: k.Self, In: out})
		_, hand := Step(k, Stamped{Stamp: r.Stamp, R: &r.R})
		r, _ = released(r, hand[0].(SendHorizon).Child)
		returns = append(returns, r)
		children = append(children, Child{Stamp: kids[i].Stamp, Invoke: kids[i].R, Result: r.Stamp, Discharge: &r.R})
	}
	open := Settled{Frame: a.R, Rounds: []Round{{Stamp: a.Self.Stamp, Children: []Child{children[0], {Stamp: kids[1].Stamp, Invoke: kids[1].R}}}}}
	if _, effs := Step(returns[0], open); !reflect.DeepEqual(effs, []Effect{Exit{}}) {
		t.Fatalf("a round with an open slot does not resolve: %#v", effs)
	}
	full := Settled{Frame: a.R, Rounds: []Round{{Stamp: a.Self.Stamp, Children: children}}}
	g, effs := Step(returns[1], full)
	res := effs[1].(Commit).R
	if !reflect.DeepEqual(res.Deps, []string{returns[0].Stamp, returns[1].Stamp}) {
		t.Fatalf("the resolve names every discharge, in instruction order: %+v", res)
	}
	g, effs = Step(g, Claimed{Won: true})
	var c Continuation
	if err := json.Unmarshal(effs[1].(Run).Stdin, &c); err != nil || len(c.Steps) != 2 || string(c.Steps[0].Value) != `"b"` || string(c.Steps[1].Value) != `"a"` {
		t.Fatalf("the agent resumes once, with both steps in the order it emitted them: %s", effs[1].(Run).Stdin)
	}
}

// One instruction out of reach faults that one, not its sibling: the fault
// holds a slot of its own in the round and discharges it.
func TestFanOutFaultsOnlyTheInstructionOutOfReach(t *testing.T) {
	a, kids, effs := fanOut(t, `{"to":"echo","payload":{"text":"a"}}`, `{"to":"shell","payload":{"text":"a"}}`)
	if kids[0].State != record.Invoke || kids[1].State != record.Fault || kids[1].R.Fault.Kind != record.Horizon {
		t.Fatalf("only the instruction out of reach faults: %+v %+v", kids[0].R, kids[1].R)
	}
	if len(effs) != 5 || effs[1] != (OpenDep{Round: a.Self.Stamp, Child: kids[1].Stamp, Pos: 1}) {
		t.Fatalf("the fault holds a slot of its own: %#v", effs)
	}
	self := effs[3].(SendHorizon).Child
	f, effs := released(kids[1], self)
	if !reflect.DeepEqual(effs, []Effect{Settle{kids[1].Stamp}, GatherWaiter{}}) || f.Self.Resume.Frame != a.Self.Frame {
		t.Fatalf("the fault discharges its own slot, then reads the agent's round: %#v", effs)
	}
	// Alone, the same fault discharges the emitter's own slot, as always.
	alone, _ := Narrow(a.Self, kids[1].R, kids[1].Stamp)
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
		ag, _ := Step(Instance{Phase: Intake}, Received{Under: room, In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
		self, _ := Narrow(room, ag.R, ag.Stamp)
		ag, _ = released(ag, self)
		other = strings.ReplaceAll(other, "FRAME", ag.Self.Frame)
		ag, effs := Step(ag, Emitted{Out: []byte("[" + echo + "," + other + "]")})
		for _, in := range []string{echo, other} {
			k, _ := Step(Instance{Phase: Intake}, Received{Under: ag.Self, In: effs[0].(Spawn).Out})
			if string(effs[0].(Spawn).Out) != in {
				t.Fatalf("%s: spawned %s", name, effs[0].(Spawn).Out)
			}
			ag, effs = Step(ag, Stamped{Stamp: k.Stamp, R: &k.R})
		}
		if _, killed := effs[0].(KillSuccessor); !killed {
			t.Fatalf("%s: every successor is killed: %#v", name, effs)
		}
		for _, e := range effs {
			if _, ok := e.(SendHorizon); ok {
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

func publishedRound(t *testing.T, effs []Effect, out []byte, round string) []Effect {
	t.Helper()
	if len(effs) < 2 || !reflect.DeepEqual(effs[0], StoreInput{out}) || effs[1] != (PublishEmission{Round: round, Hash: record.Sum(out)}) {
		t.Fatalf("complete round must be committed before slots and releases: %#v", effs)
	}
	return effs[2:]
}

func TestContinuationIsToldItsCaller(t *testing.T) {
	stdinOf := func(effs []Effect) Continuation {
		t.Helper()
		var c Continuation
		if err := json.Unmarshal(effs[0].(Run).Stdin, &c); err != nil {
			t.Fatal(err)
		}
		return c
	}
	a, _ := Step(Instance{Phase: Intake}, Received{Under: operator, In: []byte(`{"to":"agent","payload":{"text":"hi"}}`)})
	a, effs := released(a, selfOf(t, a))
	if c := stdinOf(effs); c.Caller != "" {
		t.Fatalf("the operator has no name, got caller %q", c.Caller)
	}
	// The agent invokes a continuation: that child is told the agent called it.
	call := []byte(`{"to":"agent","payload":{"text":"again"}}`)
	a, _ = Step(a, Emitted{Out: call})
	b, _ := Step(Instance{Phase: Intake}, Received{Under: a.Self, In: call})
	_, effs = Step(a, Stamped{Stamp: b.Stamp, R: &b.R})
	effs = publishedRound(t, effs, call, a.Self.Stamp)
	_, effs = released(b, effs[1].(SendHorizon).Child)
	if c := stdinOf(effs); c.Caller != "agent" {
		t.Fatalf("caller is derived from the Horizon, got %q", c.Caller)
	}
	// A payload cannot claim a caller: the field is the kernel's, outside
	// input, and a strict contract refuses an unknown one at intake.
	forged, _ := Step(Instance{Phase: Intake}, Received{Under: operator, In: []byte(`{"to":"agent","payload":{"text":"x","caller":"user"}}`)})
	if forged.R.State != record.Fault || forged.R.Fault.Kind != record.Contract {
		t.Fatalf("a payload naming a caller was not refused: %+v", forged.R)
	}
}

func TestContinuationReportsCurrentWorkspaceAndScratch(t *testing.T) {
	h := horizon.Horizon{Authority: horizon.Authority{Read: []horizon.PathGrant{{Path: "/workspace", Scope: "subpath"}}, Scratch: "/notes"}}
	var c Continuation
	if e := json.Unmarshal(continuation(Settled{Frame: record.Record{Payload: json.RawMessage(`{}`)}}, h), &c); e != nil {
		t.Fatal(e)
	}
	if c.Workspace != "/workspace" || c.Scratch != "/notes" {
		t.Fatal(c)
	}
}
