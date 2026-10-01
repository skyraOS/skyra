package machine

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// returned is a child of a round that invoked to and got payload back.
func returned(to, payload string) Child {
	return Child{Invoke: record.Record{State: record.Invoke, To: to}, Discharge: &record.Record{State: record.Return, Payload: json.RawMessage(payload)}}
}

// The actor is its identity, not its name: re-registering another program as
// echo is a new type for the same payload. One surface and one shape is one
// type, whatever the leaves.
func TestTypeOfAReturn(t *testing.T) {
	other := echoS
	other.SHA256 = "another binary"
	parent := horizon.Horizon{Reach: map[string]horizon.Surface{"echo": echoS}}
	impostor := horizon.Horizon{Reach: map[string]horizon.Surface{"echo": other}}

	a, ok := TypeOf(parent, returned("echo", `{"text":"hi"}`))
	b, _ := TypeOf(impostor, returned("echo", `{"text":"hi"}`))
	if !ok || a.Type == b.Type {
		t.Fatalf("two surfaces named echo must give two types: %+v %+v", a, b)
	}
	c, _ := TypeOf(parent, returned("echo", `{"text":"yo"}`))
	if a.Type != c.Type || a.Presentation != "echo → {text: string}" || a.Binding != "echo" {
		t.Fatalf("one surface and one shape is one type: %+v %+v", a, c)
	}
	if d, _ := TypeOf(parent, returned("echo", `"hi"`)); d.Type == a.Type {
		t.Fatal("another shape is another type")
	}

	// A reference is bound to the actor it stands for, and is that actor's type.
	parent.Refs = map[string]horizon.Grant{"system.ref.1": {Actor: "echo"}}
	r, ok := TypeOf(parent, returned("system.ref.1", `{"text":"hi"}`))
	if !ok || r.Binding != "echo" || r.Type != a.Type {
		t.Fatalf("a reference is typed as its actor: %+v", r)
	}

	// A fault has no type.
	f := returned("echo", `"x"`)
	f.Discharge.State = record.Fault
	if _, ok := TypeOf(parent, f); ok {
		t.Fatal("only a return has a type")
	}
}

// routing is an agent waiting on one round in which echo returned, with a
// route declared for echo and the runner in the room.
func routing(t *testing.T, rounds ...Round) (Instance, Settled) {
	t.Helper()
	router := agentS
	router.Routes = map[string]string{"echo": "declared"}
	reach := map[string]horizon.Surface{"echo": echoS, "agent": router, RouteRunner: Runner("/bin/route", "r")}
	for name, facet := range RoomFacets() {
		reach[name] = facet
	}
	waiter := horizon.Horizon{Stamp: rounds[len(rounds)-1].Stamp, Surface: &router, Reach: reach, Frame: "a1",
		Resume: &horizon.Horizon{Stamp: "g0", Reach: reach, Frame: "g0"}}
	frame := record.Record{State: record.Invoke, To: "agent", Payload: json.RawMessage(`{"text":"hi"}`)}
	waiter.Frame = record.Stamp(frame)
	s := Instance{Floor: "/floor", State: record.Return, Phase: Gather, Stamp: "d1",
		Self: horizon.Horizon{Stamp: "d1", Reach: reach, Frame: "c1", Resume: &waiter}}
	return s, Settled{Frame: frame, Rounds: rounds}
}

func echoed(result, payload string) Child {
	c := returned("echo", payload)
	c.Result = result
	return c
}

// A route that sends: its envelope is handed off as if the waiter emitted
// it, under the waiter's Horizon, and opens the waiter's next round.
func TestRoutingSendsUnderTheParentsHorizon(t *testing.T) {
	s, e := routing(t, Round{Stamp: "a0", Children: []Child{echoed("d1", `"hi"`)}})
	s, effs := Step(s, e)
	look, ok := effs[0].(LookupRoute)
	if !ok || look.Declared != "declared" || look.Key.Parent != horizon.Identity(*s.Self.Resume.Surface) || look.Key.Binding != "echo" {
		t.Fatalf("the route is looked up by parent, binding and type, with the declared one: %#v", effs)
	}
	source := []byte("def route():\n  return skyra.send(\"echo\", {\"text\": skyra.get(\"$\") + \"!\"})")
	s, effs = Step(s, RouteLooked{Declared: source})
	run, ok := effs[1].(Run)
	if !ok || s.Phase != Route || run.Surface.System != "runner" || len(run.Env) != 2 || run.Env[1] != "SKYRA_FRAME=d1" {
		t.Fatalf("the runner runs under an empty Horizon, returning to the discharge: %#v", effs)
	}
	var in RunnerInput
	if json.Unmarshal(run.Stdin, &in) != nil || in.Source != string(source) || string(in.Return) != `"hi"` || string(in.Previous) != "null" ||
		!reflect.DeepEqual(in.Targets, []string{"agent", "echo"}) {
		t.Fatalf("runner input: %s", run.Stdin)
	}
	s, effs = Step(s, Emitted{Out: []byte(`{"to":"d1","payload":{"send":{"to":"echo","payload":{"text":"hi!"}}}}`)})
	if _, ok := effs[1].(RecordRun); !ok {
		t.Fatalf("the run is indexed under its key: %#v", effs)
	}
	res := effs[3].(Commit).R
	var obs Observation
	if json.Unmarshal(effs[2].(StoreInput).In, &obs) != nil || obs.Route.Decision != Send || string(obs.Route.Send) != `{"payload":{"text":"hi!"},"to":"echo"}` {
		t.Fatalf("the resolve observes the decision: %s", effs[2].(StoreInput).In)
	}
	s, effs = Step(s, Claimed{Won: true})
	if !reflect.DeepEqual(effs, []Effect{Pop{}, Spawn{Under: s.Self, Out: obs.Route.Send}}) || s.Phase != Handoff || s.Self.Stamp != record.Stamp(res) {
		t.Fatalf("the winner emits the send for the waiter: %#v", effs)
	}
	next, _ := Step(Instance{Phase: Intake}, Received{Under: s.Self, In: obs.Route.Send})
	_, effs = Step(s, Stamped{Stamp: next.Stamp, R: &next.R})
	if next.State != record.Invoke || effs[0] != (OpenDep{Round: record.Stamp(res), Child: next.Stamp}) {
		t.Fatalf("the send is the waiter's next round: %+v %#v", next.R, effs)
	}
}

func TestRouteDecisions(t *testing.T) {
	s, _ := routing(t, Round{Stamp: "a0"})
	parent := *s.Self.Resume
	run := RouteRun{Child: "d1", Binding: "echo"}
	for name, c := range map[string]struct {
		out, failed           string
		decision, kind, cause string
	}{
		"interrupt":             {`{"to":"d1","payload":{"interrupt":true}}`, "", Interrupt, "", ""},
		"fault":                 {`{"to":"d1","payload":{"fault":"truncated"}}`, "", Refuse, ByRoute, "truncated"},
		"send":                  {`{"to":"d1","payload":{"send":{"to":"echo","payload":{"text":"x"}}}}`, "", Send, "", ""},
		"send outside reach":    {`{"to":"d1","payload":{"send":{"to":"shell","payload":{}}}}`, "", Refuse, record.Horizon, `"shell" is not in reach`},
		"send off contract":     {`{"to":"d1","payload":{"send":{"to":"echo","payload":{"txt":"x"}}}}`, "", Refuse, record.Contract, `unknown field "txt"`},
		"send to the room":      {`{"to":"d1","payload":{"send":{"to":"system.narrow","payload":{"actor":"echo"}}}}`, "", Refuse, record.Horizon, "the room's"},
		"send to its own slot":  {`{"to":"d1","payload":{"send":{"to":"` + parent.Frame + `","payload":"x"}}}`, "", Refuse, record.Horizon, "cannot return"},
		"two decisions":         {`{"to":"d1","payload":{"interrupt":true,"fault":"x"}}`, "", Refuse, ByRoute, "exactly one"},
		"another discharge":     {`{"to":"d2","payload":{"interrupt":true}}`, "", Refuse, ByRoute, "not a decision for this return"},
		"the runner failed":     {"", "exited", Refuse, ByRoute, "route runner: exited"},
		"the runner emits junk": {`hello`, "", Refuse, ByRoute, "not a decision"},
	} {
		got := Decided(parent, run, []byte(c.out), c.failed)
		if got.Decision != c.decision || (c.kind != "" && (got.Fault == nil || got.Fault.Kind != c.kind || !strings.Contains(got.Fault.Cause, c.cause))) {
			t.Errorf("%s: %+v %+v", name, got, got.Fault)
		}
		if c.decision == Interrupt && got.Interrupt != ByRoute {
			t.Errorf("%s: an interrupt the route asked for is of kind route: %+v", name, got)
		}
	}
}

// The chain ends at its depth: after MaxChain sends in a row the next return
// interrupts, without running the route. previous is the last earlier return
// of the same binding and type in the frame.
func TestChainDepthAndPrevious(t *testing.T) {
	rounds := []Round{{Stamp: "a0", Children: []Child{echoed("d0", `"first"`)}}}
	for i := 1; i <= MaxChain; i++ {
		rounds = append(rounds, Round{Stamp: fmt.Sprint("a", i), Children: []Child{echoed(fmt.Sprint("d", i), `"again"`)},
			Opened: &Observation{Route: &RouteRun{Decision: Send}}})
	}
	s, e := routing(t, rounds...)
	if run := routable(*s.Self.Resume, e); run == nil || run.Decision != Interrupt || run.Interrupt != Depth {
		t.Fatalf("at the chain's depth the return interrupts: %+v", run)
	}
	s, e = routing(t, rounds[:MaxChain]...)
	run := routable(*s.Self.Resume, e)
	if run == nil || run.Decision != "" {
		t.Fatalf("below the depth the route runs: %+v", run)
	}
	if p := previous(*s.Self.Resume, e.Rounds, *run); string(p) != `"again"` {
		t.Fatalf("previous is the last earlier return of this type: %s", p)
	}
	other := []Round{{Stamp: "a0", Children: []Child{echoed("d0", `{"text":"x"}`)}}, {Stamp: "a1", Children: []Child{echoed("d1", `"y"`)}}}
	s, e = routing(t, other...)
	if p := previous(*s.Self.Resume, e.Rounds, *routable(*s.Self.Resume, e)); p != nil {
		t.Fatalf("another shape is another type, so not previous: %s", p)
	}
}

// The continuation is shown how its route took each step: by names and
// presentations.
func TestTheContinuationSeesItsRoutes(t *testing.T) {
	sent, _ := json.Marshal(map[string]any{"to": "echo", "payload": map[string]string{"text": "hi!"}})
	rounds := []Round{
		{Stamp: "a0", Children: []Child{echoed("d0", `"hi"`)}},
		{Stamp: "a1", Children: []Child{echoed("d1", `"hi!"`)}, Opened: &Observation{Route: &RouteRun{Child: "d0", Decision: Send, Send: sent}}},
	}
	s, e := routing(t, rounds...)
	var c Continuation
	_ = json.Unmarshal(continuation(e, *s.Self.Resume, &RouteRun{Child: "d1", Decision: Interrupt, Interrupt: ByRoute}), &c)
	if c.Steps[0].Sent != "echo" || c.Steps[0].Interrupt != "" || c.Steps[1].By != "route" || c.Steps[1].Interrupt != ByRoute || c.Steps[1].Type != "echo → string" {
		t.Fatalf("steps: %+v", c.Steps)
	}
	_ = json.Unmarshal(continuation(e, *s.Self.Resume, &RouteRun{Child: "d1", Decision: Refuse, Fault: &record.FaultInfo{Kind: ByRoute, Cause: "no"}}), &c)
	if c.Steps[1].State != record.Fault || c.Steps[1].Fault.Cause != "no" || string(c.Steps[1].Value) != `"hi!"` {
		t.Fatalf("a refused return is a fault carrying the return: %+v", c.Steps[1])
	}
}

// Faults, fan-out rounds, the room's facets and the operator are not routed.
func TestWhatIsNotRouted(t *testing.T) {
	f := echoed("d0", `"x"`)
	f.Discharge.State = record.Fault
	narrow := returned(NarrowFacet, `{"reference":"system.ref.1"}`)
	for name, round := range map[string]Round{
		"a fault":         {Stamp: "a0", Children: []Child{f}},
		"a fan-out round": {Stamp: "a0", Children: []Child{echoed("d0", `"x"`), echoed("d1", `"y"`)}},
		"a facet":         {Stamp: "a0", Children: []Child{narrow}},
	} {
		s, e := routing(t, round)
		if run := routable(*s.Self.Resume, e); run != nil {
			t.Errorf("%s is routed: %+v", name, run)
		}
	}
	s, e := routing(t, Round{Stamp: "a0", Children: []Child{echoed("d0", `"x"`)}})
	if run := routable(*s.Self.Resume.Resume, e); run != nil {
		t.Fatal("the operator is not routed")
	}
}
