package machine

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

var hex64 = regexp.MustCompile(`[0-9a-f]{64}`)

// A route contract is built from a real Horizon, full of hashes, and shows
// none of them (next/30: hashes never reach a model).
func TestTheRouteContractShowsNoHash(t *testing.T) {
	stamp := record.Sum([]byte("a stamp"))
	echo := echoS
	echo.SHA256 = record.Sum([]byte("echo's binary"))
	echo.ProxySHA256 = record.Sum([]byte("a proxy"))
	reach := map[string]horizon.Surface{"echo": echo, RouteRunner: Runner("/bin/route", record.Sum([]byte("runner")))}
	for name, facet := range RoomFacets() {
		reach[name] = facet
	}
	h := horizon.Horizon{Stamp: stamp, Surface: &echo, Reach: reach, Frame: stamp, Resume: &horizon.Horizon{Stamp: stamp, Frame: stamp},
		Refs:      map[string]horizon.Grant{"system.ref.1": {Actor: "echo", Reach: []string{"echo"}}},
		Authority: horizon.Authority{Read: []horizon.PathGrant{{Path: "/private/work", Scope: "subpath"}}, Connect: []string{"/tmp/host.sock"}}}
	b, _ := json.Marshal(routeContract(h))
	if hex64.Match(b) {
		t.Fatalf("a hash reached the route contract: %s", hex64.Find(b))
	}
	var rc RouteContract
	_ = json.Unmarshal(b, &rc)
	if rc.Horizon.Reach["echo"] == nil || rc.Horizon.Reach[RouteFacet] == nil || rc.Horizon.Reach[RouteRunner] != nil ||
		len(rc.Horizon.Refs) != 1 || rc.Horizon.Authority.Read[0].Path != "/private/work" || len(rc.Interrupts) != 3 {
		t.Fatalf("the whole Horizon, by names and contracts: %s", b)
	}

	// A novel step's continuation carries it; one with no novel step does not.
	s, e := routing(t, Round{Stamp: "a0", Children: []Child{echoed("d0", `"hi"`)}})
	var c Continuation
	_ = json.Unmarshal(continuation(e, *s.Self.Resume, &RouteRun{Child: "d0", Decision: Interrupt, Interrupt: Novelty}), &c)
	if c.Route == nil || c.Steps[0].Type != "echo → string" {
		t.Fatalf("a novel step carries the route contract: %+v", c)
	}
	c = Continuation{}
	_ = json.Unmarshal(continuation(e, *s.Self.Resume, &RouteRun{Child: "d0", Decision: Interrupt, Interrupt: ByRoute}), &c)
	if c.Route != nil {
		t.Fatal("only novelty shows the route contract")
	}
}

// writing is a system.route invoke by the agent, released and waiting for its
// caller's rounds: the agent's first round, in which echo returned "hi", and
// its current one, this invoke.
func writing(t *testing.T, payload string) (Instance, Settled) {
	t.Helper()
	w, e := routing(t, Round{Stamp: "a0", Children: []Child{echoed("d0", `"hi"`)}})
	caller := *w.Self.Resume
	caller.Stamp = "r1"
	invoke := intakeRecord(caller, `{"to":"system.route","payload":`+payload+`}`)
	if invoke.State != record.Invoke {
		t.Fatalf("system.route is invoked like any capability: %+v", invoke.Fault)
	}
	s := Instance{Floor: "/floor", State: record.Invoke, Phase: Release, Under: caller, R: invoke, Stamp: record.Stamp(invoke)}
	self, _ := Narrow(caller, invoke, s.Stamp)
	s, effs := released(s, self)
	if s.Phase != Facet || len(effs) != 1 || effs[0] != (GatherWaiter{}) {
		t.Fatalf("system.route reads its caller's rounds: %#v", effs)
	}
	e.Rounds = append(e.Rounds, Round{Stamp: "r1", Children: []Child{{Stamp: s.Stamp, Invoke: invoke}}})
	return s, e
}

// result is what a facet emitted, as its successor would read it.
func result(t *testing.T, effs []Effect) map[string]any {
	t.Helper()
	for _, eff := range effs {
		if sp, ok := eff.(Spawn); ok {
			var out struct{ Payload map[string]any }
			_ = json.Unmarshal(sp.Out, &out)
			return out.Payload
		}
	}
	t.Fatalf("no result: %#v", effs)
	return nil
}

const interrupts = "def route():\n  return skyra.interrupt()"

func TestSystemRouteRecordsARoute(t *testing.T) {
	s, e := writing(t, `{"step":0,"source":"`+strings.ReplaceAll(interrupts, "\n", `\n`)+`"}`)
	s, effs := Step(s, e)
	load, ok := effs[0].(LoadRuns)
	want, _ := TypeOf(*s.Self.Resume, e.Rounds[0].Children[0])
	if !ok || s.Phase != Cases || load.Key != (RouteKey{Parent: horizon.Identity(*s.Self.Resume.Surface), Binding: "echo", Type: want.Type}) {
		t.Fatalf("the step is resolved to its key, under the caller's identity: %#v", effs)
	}
	_, effs = Step(s, RunsLoaded{})
	put, ok := effs[0].(PutRoute)
	if !ok || string(put.Source) != interrupts || put.Entry.Author != load.Key.Parent || put.Entry.Source != record.Sum([]byte(interrupts)) {
		t.Fatalf("with no recorded runs the route is recorded, authored by the caller: %#v", effs)
	}
	if r := result(t, effs); r["ok"] != true {
		t.Fatalf("system.route returns ok: %v", r)
	}

	for name, c := range map[string]struct{ payload, cause string }{
		"no such step":           {`{"step":5,"source":"x"}`, "no step 5"},
		"a step that is a facet": {`{"step":1,"source":"x"}`, "no step 1"},
		"a step that is not int": {`{"step":0.5,"source":"x"}`, "indexes"},
	} {
		s, e := writing(t, c.payload)
		_, effs := Step(s, e)
		if r := result(t, effs); r["ok"] != false || !strings.Contains(r["error"].(string), c.cause) {
			t.Errorf("%s: %v", name, r)
		}
	}
}

// The log is the test suite (F8): the new source reruns every recorded case;
// a changed decision is a flip, refused unless the caller accepts every flip.
func TestARewrittenRouteIsTestedAgainstTheLog(t *testing.T) {
	recordedIn, _ := json.Marshal(RunnerInput{Source: "old", Return: json.RawMessage(`"hi"`), Targets: []string{"echo"}})
	cases := RunsLoaded{Runs: []RecordedRun{{Input: recordedIn, Output: []byte(`{"to":"d0","payload":{"interrupt":true}}`)}}}
	src := `"source":"` + strings.ReplaceAll(interrupts, "\n", `\n`) + `"`

	run := func(payload string, now string) []Effect {
		s, e := writing(t, payload)
		s, _ = Step(s, e)
		s, effs := Step(s, cases)
		r, ok := effs[0].(Run)
		var in RunnerInput
		if !ok || s.Phase != Retest || json.Unmarshal(r.Stdin, &in) != nil || in.Source != interrupts || string(in.Return) != `"hi"` {
			t.Fatalf("the recorded case reruns under the new source: %#v", effs)
		}
		_, effs = Step(s, Emitted{Out: []byte(`{"to":"` + s.Stamp + `","payload":` + now + `}`)})
		return effs
	}

	if effs := run(`{"step":0,`+src+`}`, `{"interrupt":true}`); result(t, effs)["ok"] != true {
		t.Fatalf("the same decision is no flip: %#v", effs)
	}
	effs := run(`{"step":0,`+src+`}`, `{"fault":"no"}`)
	r := result(t, effs)
	flips, _ := r["flips"].([]any)
	if r["ok"] != false || len(flips) != 1 || flips[0].(map[string]any)["case"] != 0.0 || flips[0].(map[string]any)["return"] != "hi" {
		t.Fatalf("a flip is refused, shown by case and return: %v", r)
	}
	for _, eff := range effs {
		if _, ok := eff.(PutRoute); ok {
			t.Fatal("a refused route is not recorded")
		}
	}
	effs = run(`{"step":0,`+src+`,"accept":[0]}`, `{"fault":"no"}`)
	put, ok := effs[0].(PutRoute)
	if !ok || len(put.Entry.Accepted) != 1 || string(put.Entry.Accepted[0].Was) != `{"interrupt":true}` || string(put.Entry.Accepted[0].Now) != `{"fault":"no"}` {
		t.Fatalf("an accepted flip is recorded with the route: %#v", effs)
	}
	if r := result(t, run(`{"step":0,`+src+`,"accept":[0]}`, `{"interrupt":true}`)); r["ok"] != false {
		t.Fatalf("accepting a case that does not flip is refused: %v", r)
	}
}
