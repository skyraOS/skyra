package machine

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// Narrowing through the room (next/26): a caller invokes system.narrow, the
// room returns a reference and records it in the caller's reach, and invoking
// the reference runs the actor under the narrowed Horizon. The caller's own
// Horizon is unchanged.
func TestNarrowingThroughTheRoom(t *testing.T) {
	root := operator
	root.Reach = map[string]horizon.Surface{"echo": echoS, "agent": agentS}
	for name, facet := range RoomFacets() {
		root.Reach[name] = facet
	}
	root.Authority = horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	agent, _ := Narrow(root, intakeRecord(root, `{"to":"agent","payload":{"text":"x"}}`), "a1")

	// The facet: an ordinary invoke, checked at intake.
	ask := `{"to":"system.narrow","payload":{"actor":"echo","reach":[],"authority":{}}}`
	s, _ := Step(Instance{Phase: Intake}, Received{Under: agent, In: []byte(ask)})
	if s.State != record.Invoke || s.R.To != NarrowFacet {
		t.Fatalf("system.narrow is invoked like any capability: %+v", s.R)
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
	if err != nil || child.Surface.Path != echoS.Path || len(child.Reach) != 0 || !reflect.DeepEqual(child.Authority, horizon.Authority{}) || child.Resume.Stamp != "resolved" {
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
	for name, facet := range RoomFacets() {
		root.Reach[name] = facet
	}
	root.Authority = horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}}
	for name, c := range map[string]struct{ in, cause string }{
		"authority outside the Horizon": {`{"to":"system.narrow","payload":{"actor":"echo","authority":{"read":[{"path":"/etc","scope":"subpath"}]}}}`, "outside the Horizon"},
		"a name not in reach":           {`{"to":"system.narrow","payload":{"actor":"echo","reach":["shell"]}}`, `keeps "shell"`},
		"an actor not in reach":         {`{"to":"system.narrow","payload":{"actor":"shell"}}`, `"shell" is not an actor`},
		"the room itself":               {`{"to":"system.narrow","payload":{"actor":"system.narrow"}}`, "not an actor"},
		"a reference never recorded":    {`{"to":"system.ref.1","payload":{"text":"x"}}`, "not in reach"},
	} {
		r := intakeRecord(root, c.in)
		if r.State != record.Fault || r.Fault.Kind != record.Horizon || !strings.Contains(r.Fault.Cause, c.cause) {
			t.Fatalf("%s: %+v", name, r.Fault)
		}
	}
	// Narrowing to a subset is fine.
	if r := intakeRecord(root, `{"to":"system.narrow","payload":{"actor":"echo","authority":{"read":[{"path":"/private/work/notes","scope":"literal"}]}}}`); r.State != record.Invoke {
		t.Fatalf("a subset of the Horizon may be kept: %+v", r.Fault)
	}
}

// The route runner is the runtime's: in reach, so every Horizon carries it,
// but never invocable, never shown as a tool, and kept by every narrowing.
func TestTheRouteRunnerIsTheRuntimes(t *testing.T) {
	root := operator
	root.Reach = map[string]horizon.Surface{"echo": echoS, RouteRunner: Runner("/bin/route", "r")}
	for name, facet := range RoomFacets() {
		root.Reach[name] = facet
	}
	if r := intakeRecord(root, `{"to":"system.route-runner","payload":{}}`); r.State != record.Fault || r.Fault.Kind != record.Horizon {
		t.Fatalf("the runner cannot be invoked: %+v", r)
	}
	if _, shown := tools(root)[RouteRunner]; shown {
		t.Fatal("the runner is no one's tool")
	}
	if r := intakeRecord(root, `{"to":"system.narrow","payload":{"actor":"system.route-runner"}}`); r.State != record.Fault {
		t.Fatalf("the runner cannot be narrowed to: %+v", r)
	}
	if k := kept(root.Reach, nil); k[RouteRunner].System != "runner" || len(k) != 1 {
		t.Fatalf("a narrowing keeps the runner: %v", k)
	}
}
