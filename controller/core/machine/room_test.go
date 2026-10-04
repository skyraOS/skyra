package machine

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Narrowing through the Horizon (next/26, next/32): a caller invokes system
// with narrow, the Horizon returns a reference and records it in the caller's reach, and invoking
// the reference runs the actor under the narrowed Horizon. The caller's own
// Horizon is unchanged.
func TestNarrowingThroughTheRoom(t *testing.T) {
	root := operator
	root.Reach = map[string]horizon.Surface{"echo": echoS, "agent": agentS}
	root.Reach[System] = Syscalls()
	root.Authority = horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")

	// The facet: an ordinary invoke, checked at intake.
	ask := `{"to":"system","payload":{"narrow":{"actor":"echo","reach":[],"authority":{}}}}`
	s, _ := Step(Instance{Phase: Intake}, Received{Under: agent, In: []byte(ask)})
	if s.State != record.Invoke || s.R.To != System {
		t.Fatalf("narrow is invoked like any capability: %+v", s.R)
	}
	facet := selfOf(t, s)
	ref := "system.ref.1" // a name, never the invoke's stamp: a continuation is shown it (next/30)
	if g := facet.Resume.Refs[ref]; g.Actor != "echo" || len(g.Reach) != 0 || !reflect.DeepEqual(g.Authority, horizon.Authority{}) {
		t.Fatalf("the waiting caller is recorded as holding the reference: %+v", facet.Resume.Refs)
	}
	if len(agent.Refs) != 0 || len(facet.Refs) != 0 {
		t.Fatalf("no one else holds it: %+v %+v", agent.Refs, facet.Refs)
	}

	// Nothing runs: the room derives the result, which returns the reference.
	s, effs := released(s, facet)
	spawn, ok := effs[0].(Spawn)
	if !ok {
		t.Fatalf("a facet's result is handed off like any output: %#v", effs)
	}
	ret, _ := Step(Instance{Phase: Intake}, Received{Under: spawn.Under, In: spawn.Out})
	var got struct{ Reference string }
	if ret.State != record.Return || json.Unmarshal(ret.R.Payload, &got) != nil || got.Reference != ref {
		t.Fatalf("system.narrow returns the reference: %+v", ret.R)
	}

	// The caller resumes holding it, sees it as a tool, and invokes it.
	caller := *facet.Resume
	caller.Stamp = "resolved"
	if _, shown := tools(caller)[ref]; !shown {
		t.Fatalf("a held reference is shown to the continuation: %v", tools(caller))
	}
	call := intakeRecord(caller, `{"to":"`+ref+`","payload":{"text":"hi"}}`)
	if call.State != record.Invoke || call.To != ref {
		t.Fatalf("a held reference is invoked like a name: %+v", call)
	}
	child, err := Narrow(caller, call, "c1")
	if _, sys := child.Reach[System]; err != nil || child.Surface.Path != echoS.Path || len(child.Reach) != 1 || !sys || !reflect.DeepEqual(child.Authority, horizon.Authority{}) || child.Resume.Stamp != "resolved" {
		t.Fatalf("the actor runs under the narrowed Horizon, waited on by the caller: %+v %v", child, err)
	}
	if !reflect.DeepEqual(caller.Authority, root.Authority) || len(caller.Reach) != len(root.Reach) {
		t.Fatalf("the caller's own Horizon is unchanged: %+v", caller)
	}
}

// A narrowing may only drop what is held; anything else is a horizon fault,
// derived at intake. A reference nobody recorded is not in reach.
func TestNarrowingCannotWiden(t *testing.T) {
	root := operator
	root.Reach = map[string]horizon.Surface{"echo": echoS}
	root.Reach[System] = Syscalls()
	root.Authority = horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	for name, c := range map[string]struct{ in, cause string }{
		"authority outside the Horizon": {`{"to":"system","payload":{"narrow":{"actor":"echo","authority":{"read":[{"path":"/etc","scope":"subpath"}]}}}}`, "outside the Horizon"},
		"a name not in reach":           {`{"to":"system","payload":{"narrow":{"actor":"echo","reach":["shell"]}}}`, `keeps "shell"`},
		"an actor not in reach":         {`{"to":"system","payload":{"narrow":{"actor":"shell"}}}`, `"shell" is not an actor`},
		"the Horizon itself":            {`{"to":"system","payload":{"narrow":{"actor":"system"}}}`, "not an actor"},
		"a reference never recorded":    {`{"to":"system.ref.1","payload":{"text":"x"}}`, "not in reach"},
	} {
		r := intakeRecord(root, c.in)
		if r.State != record.Fault || r.Fault.Kind != record.Horizon || !strings.Contains(r.Fault.Cause, c.cause) {
			t.Fatalf("%s: %+v", name, r.Fault)
		}
	}
	// Narrowing to a subset is fine.
	if r := intakeRecord(root, `{"to":"system","payload":{"narrow":{"actor":"echo","authority":{"read":[{"path":"/private/work/notes","scope":"literal"}]}}}}`); r.State != record.Invoke {
		t.Fatalf("a subset of the Horizon may be kept: %+v", r.Fault)
	}
}

// room is the operator's Horizon with the agent, echo and the syscalls.
func room() horizon.Horizon {
	root := operator
	root.Reach = map[string]horizon.Surface{"echo": echoS, "agent": agentS, System: Syscalls()}
	return root
}

const shout = `{"title":"shout","type":"object","required":["shout"],"fields":{"shout":{"type":"string"}}}`

func extend(schema string) string {
	return `{"to":"system","payload":{"extend-reach":{"schema":` + schema + `}}}`
}

// extend-reach (next/32, H3): an actor adds a contract to itself, with no
// approval. The kernel derives the actor from the emitter; the contract joins
// the emitter's reach entry in every frame of its Horizon; invoking it is
// selected by fit and runs the emitter's own program.
func TestAnActorExtendsItself(t *testing.T) {
	root := room()
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")
	if agent.Name != "agent" || agent.Origin != "" {
		t.Fatalf("a frame knows the name it was invoked through, and shares the room's Horizon: %+v", agent)
	}
	s, _ := Step(Instance{Phase: Intake}, Received{Under: agent, In: []byte(extend(shout))})
	if s.State != record.Invoke || s.R.To != System {
		t.Fatalf("extend-reach is an invoke of system: %+v", s.R)
	}
	facet := selfOf(t, s)
	if got := facet.Resume.Reach["agent"].Contracts; len(got) != 1 || got[0].Title != "shout" {
		t.Fatalf("the waiting agent holds the new contract: %+v", facet.Resume.Reach["agent"])
	}
	if below := facet.Resume.Resume; below == nil || len(below.Reach["agent"].Contracts) != 1 {
		t.Fatalf("so does every frame below that shares its Horizon: %+v", below)
	}
	if len(agent.Reach["agent"].Contracts) != 0 || len(root.Reach["agent"].Contracts) != 0 {
		t.Fatal("nothing is written in place: other copies are untouched")
	}

	// Its result is derived, like any syscall's.
	_, effs := released(s, facet)
	ret, _ := Step(Instance{Phase: Intake}, Received{Under: effs[0].(Spawn).Under, In: effs[0].(Spawn).Out})
	if ret.State != record.Return || string(ret.R.Payload) != `{"ok":true}` {
		t.Fatalf("extend-reach returns ok: %+v", ret.R)
	}

	// The agent resumes holding it: it is shown, selected by fit, and runs
	// the agent's own program, executed against the new contract.
	resumed := *facet.Resume
	resumed.Stamp = "resolved"
	if c := contracts(resumed)["agent"]; len(c) != 2 || c[1].Title != "shout" {
		t.Fatalf("a continuation is shown every contract: %v", contracts(resumed))
	}
	call := intakeRecord(resumed, `{"to":"agent","payload":{"shout":"hi"}}`)
	if call.State != record.Invoke {
		t.Fatalf("the new contract is invokable: %+v", call)
	}
	child, _ := Narrow(resumed, call, "c1")
	if child.Surface.Path != agentS.Path {
		t.Fatalf("it runs the emitter's own program: %+v", child.Surface)
	}
	c := against(*child.Surface, call.Payload)
	if c == nil || c.Title != "shout" {
		t.Fatalf("the program is executed against the contract its input fits: %+v", c)
	}
	if again := intakeRecord(resumed, `{"to":"agent","payload":{"text":"hi"}}`); again.State != record.Invoke || against(agentS, again.Payload) != nil {
		t.Fatalf("the registered contract still is: %+v", again)
	}
}

// What extend-reach is not: anything that names another actor, a duplicate,
// a schema the kernel cannot check, an emitter that is no actor, or a branch
// of a fan-out.
func TestExtendReachRefuses(t *testing.T) {
	root := room()
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")
	forked := agent
	forked.Forked = true
	for name, c := range map[string]struct {
		under       horizon.Horizon
		in          string
		kind, cause string
	}{
		"a payload naming an actor":   {agent, `{"to":"system","payload":{"extend-reach":{"schema":` + shout + `,"actor":"echo"}}}`, record.Contract, `unknown field "actor"`},
		"a duplicate (actor, schema)": {agent, extend(`{"type":"object","required":["text"],"fields":{"text":{"type":"string"}}}`), record.Contract, "already holds"},
		"a schema it cannot check":    {agent, extend(`{"type":"blob"}`), record.Contract, "unknown type"},
		"a schema with unknown parts": {agent, extend(`{"type":"string","pattern":"x"}`), record.Contract, `unknown field "pattern"`},
		"the operator":                {root, extend(shout), record.Horizon, "only an actor"},
		"one of several siblings":     {forked, extend(shout), record.Horizon, "siblings"},
	} {
		r := intakeRecord(c.under, c.in)
		if r.State != record.Fault || r.Fault.Kind != c.kind || !strings.Contains(r.Fault.Cause, c.cause) {
			t.Errorf("%s: %+v", name, r.Fault)
		}
	}
}

// An actor invoked through a reference runs in a Horizon of its own, and
// extends itself there only: its caller's Horizon is untouched. An actor not
// in its own Horizon's reach has no entry to extend.
func TestAnActorExtendsItselfInItsOwnHorizonOnly(t *testing.T) {
	root := room()
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")
	through := func(keep string) horizon.Horizon {
		ask := intakeRecord(agent, `{"to":"system","payload":{"narrow":{"actor":"echo","reach":`+keep+`}}}`)
		facet, _ := Narrow(agent, ask, "n1")
		caller := *facet.Resume
		caller.Stamp = "resolved"
		echo, err := Narrow(caller, intakeRecord(caller, `{"to":"system.ref.1","payload":{"text":"hi"}}`), "e1")
		if err != nil || echo.Name != "echo" || echo.Origin != "e1" {
			t.Fatalf("a reference's actor runs in a Horizon of its own: %+v %v", echo, err)
		}
		return echo
	}
	echo := through(`["echo"]`)
	r := intakeRecord(echo, extend(shout))
	facet, err := Narrow(echo, r, "x1")
	if r.State != record.Invoke || err != nil {
		t.Fatalf("echo extends itself: %+v %v", r, err)
	}
	if len(facet.Resume.Reach["echo"].Contracts) != 1 {
		t.Fatalf("in its own Horizon: %+v", facet.Resume.Reach["echo"])
	}
	if below := facet.Resume.Resume; len(below.Reach["echo"].Contracts) != 0 {
		t.Fatalf("and nowhere else: the caller's Horizon is untouched: %+v", below.Reach["echo"])
	}
	if r := intakeRecord(through(`[]`), extend(shout)); r.State != record.Fault || r.Fault.Kind != record.Horizon || !strings.Contains(r.Fault.Cause, "own Horizon") {
		t.Fatalf("an actor outside its own reach has nothing to extend: %+v", r.Fault)
	}
}

// A payload that fits two contracts of one target is ambiguous: a contract
// fault at invoke. The kernel never reasons about overlap ahead of time.
func TestAPayloadFittingTwoContractsIsRefused(t *testing.T) {
	root := room()
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")
	loose := `{"type":"object","fields":{"text":{"type":"string"},"n":{"type":"number"}}}`
	facet, _ := Narrow(agent, intakeRecord(agent, extend(loose)), "x1")
	resumed := *facet.Resume
	resumed.Stamp = "resolved"
	if r := intakeRecord(resumed, `{"to":"agent","payload":{"text":"hi"}}`); r.State != record.Fault || r.Fault.Kind != record.Contract || !strings.Contains(r.Fault.Cause, "ambiguous") {
		t.Fatalf("a payload fitting both is refused: %+v", r.Fault)
	}
	if r := intakeRecord(resumed, `{"to":"agent","payload":{"n":1}}`); r.State != record.Invoke {
		t.Fatalf("one fitting only the new contract is selected: %+v", r.Fault)
	}
}

// strand (next/32, H4) is read from the tree: intake checks what the
// requester names, release asks the shell to walk, and the walk's strand is
// the syscall's result. A refusal is a horizon fault of the syscall's frame,
// derived again by replay from its observation.
func TestStrandIsReadThenReturned(t *testing.T) {
	root := room()
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")
	obs := strings.Repeat("ab", 32)
	ask := func(observation, target, schema string) string {
		return `{"to":"system","payload":{"strand":{"observation":"` + observation + `","contract":{"target":"` + target + `","schema":` + schema + `}}}}`
	}
	echoContract := `{"type":"object","required":["text"],"fields":{"text":{"type":"string"}}}`
	for name, c := range map[string]struct{ in, cause string }{
		"not a stamp":            {ask("x", "echo", echoContract), "not a stamp"},
		"a target not in reach":  {ask(obs, "shell", echoContract), `"shell" is not an actor`},
		"a contract it lacks":    {ask(obs, "echo", shout), "no such contract"},
		"the Horizon's syscalls": {ask(obs, "system", shout), "not an actor"},
	} {
		if r := intakeRecord(agent, c.in); r.State != record.Fault || r.Fault.Kind != record.Horizon || !strings.Contains(r.Fault.Cause, c.cause) {
			t.Errorf("%s: %+v", name, r.Fault)
		}
	}

	s, _ := Step(Instance{Phase: Intake}, Received{Under: agent, In: []byte(ask(obs, "echo", echoContract))})
	if s.State != record.Invoke {
		t.Fatalf("a strand of a contract the requester holds is an invoke: %+v", s.R.Fault)
	}
	s, effs := released(s, selfOf(t, s))
	read, ok := effs[0].(ReadStrand)
	if !ok || read.Observation != obs || read.Target != "echo" || read.Schema != echoS.Input || read.Origin != "" {
		t.Fatalf("release asks the shell to walk the tree: %#v", effs)
	}

	_, effs = Step(s, StrandRead{Strand: []byte(`{"chain":["c"]}`)})
	spawn := effs[0].(Spawn)
	if ret := intakeRecord(spawn.Under, string(spawn.Out)); ret.State != record.Return || string(ret.Payload) != `{"chain":["c"]}` {
		t.Fatalf("the strand is the syscall's return: %+v %s", ret, spawn.Out)
	}

	_, effs = Step(s, StrandRead{Refused: "no invoke"})
	f := effs[1].(Commit).R
	if f.State != record.Fault || f.Fault.Kind != record.Horizon || f.Fault.Cause != "strand: no invoke" || !f.Supervisor {
		t.Fatalf("a refusal is a horizon fault of the syscall's frame: %+v", f.Fault)
	}
	if again := Observed(f.Predecessor, effs[0].(StoreInput).In); record.Stamp(again) != record.Stamp(f) {
		t.Fatal("replay derives the refusal's kind from its observation")
	}
	if forged := Observed("p", []byte(`{"refused":{"kind":"horizon","cause":"x"},"extra":1}`)); forged.Fault.Kind != record.Custody {
		t.Fatalf("only a refusal exactly as the runtime writes it carries a kind: %+v", forged.Fault)
	}
}

func TestStrandReconstructionForms(t *testing.T) {
	h := room()
	anchor := strings.Repeat("ab", 32)
	leaf := strings.Repeat("cd", 32)
	schema := `{"type":"object","required":["text"],"fields":{"text":{"type":"string"}}}`
	contract := `"contract":{"target":"agent","schema":` + schema + `}`
	cases := []struct {
		args  string
		valid bool
	}{
		{`"observation":"` + leaf + `",` + contract, true},
		{`"invocation":"` + anchor + `","observations":["` + anchor + `"],` + contract, true},
		{`"invocation":"` + anchor + `","observations":["` + leaf + `","` + anchor + `"],` + contract, true},
		{contract, false},
		{`"invocation":"` + anchor + `",` + contract, false},
		{`"observations":["` + leaf + `"],` + contract, false},
		{`"invocation":"` + anchor + `","observations":[],` + contract, false},
		{`"observation":"` + leaf + `","invocation":"` + anchor + `","observations":["` + leaf + `"],` + contract, false},
		{`"observation":"` + leaf + `","observations":[],` + contract, false},
		{`"invocation":"bad","observations":["` + leaf + `"],` + contract, false},
		{`"invocation":"` + anchor + `","observations":["bad"],` + contract, false},
	}
	for _, c := range cases {
		r := intakeRecord(h, `{"to":"system","payload":{"strand":{`+c.args+`}}}`)
		if (r.State == record.Invoke) != c.valid {
			t.Errorf("%s: %+v", c.args, r)
		}
	}
	args := json.RawMessage(`{"invocation":"` + anchor + `","observations":["` + leaf + `"],` + contract + `}`)
	read, err := stranding(h, args)
	if err != nil || read.Invocation != anchor || !reflect.DeepEqual(read.Observations, []string{leaf}) || read.Room != h.Frame {
		t.Fatalf("tree effect: %+v %v", read, err)
	}
}
