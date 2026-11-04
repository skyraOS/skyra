// Step, driven by events alone: no shell, no model. Each test plays the
// shell's part, answering the effects Step asks for.
package tree

import (
	"reflect"
	"strings"
	"testing"
)

// start is a tree of the test apprentice, started: the root is self, with
// the reach and kind given; self is also the actor that runs every child.
func start(t *testing.T, k Kind, reach []Contract, depth int) (Tree, []Effect) {
	t.Helper()
	return Step(NewTree(app, depth), Started{Self: self, Reach: reach, Kind: k, Actor: self, Endowment: shapeS})
}

func turns(es []Effect) []Turn {
	var out []Turn
	for _, e := range es {
		if x, ok := e.(Execute); ok {
			for _, a := range x.Acts {
				if tn, ok := a.(Turn); ok {
					out = append(out, tn)
				}
			}
		}
	}
	return out
}

func dispatches(es []Effect) []InvokeContract {
	var out []InvokeContract
	for _, e := range es {
		if x, ok := e.(Execute); ok {
			for _, a := range x.Acts {
				if ic, ok := a.(InvokeContract); ok {
					out = append(out, ic)
				}
			}
		}
	}
	return out
}

func records(es []Effect, kind string) []RecordWrite {
	var out []RecordWrite
	for _, e := range es {
		if c, ok := e.(Commit); ok {
			for _, w := range c.Writes {
				if rw, ok := w.(RecordWrite); ok && rw.Kind == kind {
					out = append(out, rw)
				}
			}
		}
	}
	return out
}

func reads(es []Effect) []CELKey {
	var out []CELKey
	for _, e := range es {
		if rd, ok := e.(Read); ok {
			for _, k := range rd.Keys {
				out = append(out, k.(CELKey))
			}
		}
	}
	return out
}

func turnFor(t *testing.T, es []Effect, id AgentID) Turn {
	t.Helper()
	for _, tn := range turns(es) {
		if tn.Agent == id {
			return tn
		}
	}
	t.Fatalf("no turn for %s in %v", id, es)
	return Turn{}
}

func agent(t *testing.T, tr Tree, id AgentID) Agent {
	t.Helper()
	a, ok := tr.Agents[id]
	if !ok {
		t.Fatalf("no agent %s", id)
	}
	return *a
}

func reply(tr Tree, id AgentID, out string) (Tree, []Effect) {
	return Step(tr, Replied{Agent: id, Out: []byte(out)})
}

// A whole tree, end to end: a delegate root makes a goal and delegates it
// with one contract; the act transformation provisions an act child, which
// dispatches its goal through perform_effect and returns what came back;
// the root returns its child's return.
func TestStepRunsATree(t *testing.T) {
	tr, es := start(t, KindDelegate, []Contract{x, y}, 3)
	root := turnFor(t, es, "0")
	if !reflect.DeepEqual(root.Offer, Offered(app, KindDelegate, []Contract{x, y})) || root.Binary != self.binary {
		t.Fatalf("the root's turn: %+v", root)
	}

	tr, es = reply(tr, "0", `[{"goal":"summarize the logs"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`)
	if p := agent(t, tr, "0").Phase; p != PhaseDelegating {
		t.Fatalf("the root is %v, not delegating", p)
	}
	child := agent(t, tr, "0.0")
	if child.Kind != KindAct || !reflect.DeepEqual(child.Offer.Invoke, []Contract{x}) || child.Self != self {
		t.Fatalf("the child: kind %v, offer %+v", child.Kind, child.Offer)
	}
	// The act transformation: the computation, create_act's description, the goal.
	if len(child.Names) != 3 || child.Names[1].Obs.typ.contract != app.CreateAct() ||
		child.Names[2].Obs.typ.contract != app.Goal() || child.Names[0].Obs.typ.contract != app.Delegate() {
		t.Fatalf("the child's names: %+v", child.Names)
	}
	turnFor(t, es, "0.0")

	tr, es = reply(tr, "0.0", `[{"perform_effect":{"contract":0,"input":2}}]`)
	ic := dispatches(es)
	if len(ic) != 1 || ic[0].Contract != x || ic[0].Input != 2 || ic[0].Element != 0 {
		t.Fatalf("the dispatch: %+v", ic)
	}

	tr, es = Step(tr, Returned{Agent: "0.0", Element: 0, Shape: shapeA})
	if g := records(es, "guard"); len(g) != 1 || g[0].Body["see"] != "unseen" {
		t.Fatalf("the first return is unseen: %+v", g)
	}
	turnFor(t, es, "0.0")
	came := agent(t, tr, "0.0").Names[3]
	if came.Obs.typ.contract != x {
		t.Fatalf("what came back is not typed by its contract: %+v", came)
	}

	tr, es = reply(tr, "0.0", `[{"return":{"members":[3]}}]`)
	if p := agent(t, tr, "0.0").Phase; p != PhaseDone {
		t.Fatalf("the child is %v", p)
	}
	turnFor(t, es, "0")
	r := agent(t, tr, "0")
	if len(r.Names) != 3 || !r.Names[2].Obs.Equal(came.Obs) {
		t.Fatalf("the root did not receive its child's return: %+v", r.Names)
	}
	if want, _ := CreateReturn(child.Input, wrapperShape(KindAct)); r.Names[2].Base != want {
		t.Fatalf("the child's return is not under its wrapper's base type")
	}

	tr, es = reply(tr, "0", `[{"return":{"members":[2]}}]`)
	if !tr.Done() || agent(t, tr, "0").Phase != PhaseDone {
		t.Fatal("the tree is not done")
	}
	if len(turns(es)) != 0 || len(records(es, "done")) != 1 {
		t.Fatalf("after the root returns: %+v", es)
	}
}

// A delegate child: two contracts passed, so create_delegation and the
// packaged reach, and the contracts pass-only.
func TestStepDelegateChild(t *testing.T) {
	tr, _ := start(t, KindDelegate, []Contract{x, y}, 3)
	tr, _ = reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0,1]}}]`)
	c := agent(t, tr, "0.0")
	if c.Kind != KindDelegate || !reflect.DeepEqual(c.Offer.PassOnly, []Contract{x, y}) || len(c.Offer.Invoke) != 0 {
		t.Fatalf("the child: %v %+v", c.Kind, c.Offer)
	}
	if len(c.Names) != 4 || c.Names[1].Obs.typ.contract != app.CreateDelegation() ||
		c.Names[2].Obs.typ.contract != app.Reach() || c.Names[3].Obs.typ.contract != app.Goal() {
		t.Fatalf("the child's names: %+v", c.Names)
	}
}

// A goal made by the model is never one of its returns; what it was passed
// is never returned either.
func TestStepReturnRules(t *testing.T) {
	tr, _ := start(t, KindDelegate, []Contract{x, y}, 3)
	tr2, _ := reply(tr, "0", `[{"goal":"g"}]`)
	turned := agent(t, tr2, "0")
	if turned.Phase != PhaseThinking || len(turned.Names) != 2 {
		t.Fatalf("a call of a goal alone goes back to the model: %+v", turned)
	}
	if tr3, _ := reply(tr2, "0", `[{"return":{"members":[1]}}]`); agent(t, tr3, "0").Phase != PhaseFaulted {
		t.Error("the root returned its own goal")
	}
	if tr3, _ := reply(tr, "0", `[{"return":{"members":[0]}}]`); agent(t, tr3, "0").Phase != PhaseFaulted {
		t.Error("the root returned its endowment, what it was passed")
	}
}

// Every call that is not one the Offer allows is a fault, never an exit
// status, and a faulted call leaves nothing behind.
func TestStepFaults(t *testing.T) {
	many := "[" + strings.Repeat(`{"goal":"g"},`, MaxFanOut) + `{"goal":"g"}]`
	for _, tc := range []struct {
		name string
		kind Kind
		out  string
	}{
		{"not json", KindDelegate, `not json`},
		{"an empty call", KindDelegate, `[]`},
		{"more than 32", KindDelegate, many},
		{"an unknown instruction", KindDelegate, `[{"bogus":1}]`},
		{"two keys", KindDelegate, `[{"goal":"a","exit":"b"}]`},
		{"an unknown field", KindDelegate, `[{"return":{"members":[0],"extra":1}}]`},
		{"goal from an act node", KindAct, `[{"goal":"g"}]`},
		{"create_input from a delegate", KindDelegate, `[{"create_input":"i"}]`},
		{"perform_effect from a delegate", KindDelegate, `[{"perform_effect":{"contract":0,"input":0}}]`},
		{"delegate from an act node", KindAct, `[{"delegate":{"observations":[0],"contracts":[0]}}]`},
		{"return not alone", KindDelegate, `[{"goal":"g"},{"return":{"members":[0]}}]`},
		{"exit not alone", KindDelegate, `[{"goal":"g"},{"exit":"x"}]`},
		{"a later reference", KindDelegate, `[{"delegate":{"observations":[{"earlier":1}],"contracts":[0]}},{"goal":"g"}]`},
		{"an earlier reference to nothing made", KindAct, `[{"perform_effect":{"contract":0,"input":0}},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`},
		{"a name not in scope", KindDelegate, `[{"goal":"g"},{"delegate":{"observations":[99],"contracts":[0]}}]`},
		{"a contract not held", KindDelegate, `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[7]}}]`},
		{"a contract it may not invoke", KindAct, `[{"perform_effect":{"contract":3,"input":0}}]`},
		{"a delegate with no goal", KindDelegate, `[{"delegate":{"observations":[0],"contracts":[0]}}]`},
		{"one observation twice", KindDelegate, `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0},{"earlier":0}],"contracts":[0]}}]`},
		{"an empty return", KindDelegate, `[{"return":{"members":[]}}]`},
	} {
		tr, _ := start(t, tc.kind, []Contract{x, y}, 3)
		before := agent(t, tr, "0")
		tr, es := reply(tr, "0", tc.out)
		a := agent(t, tr, "0")
		if a.Phase != PhaseFaulted || len(records(es, "fault")) != 1 {
			t.Errorf("%s: the root is %v, not faulted", tc.name, a.Phase)
			continue
		}
		if len(a.Names) != len(before.Names) || len(tr.Agents) != 1 || len(turns(es)) != 0 || len(dispatches(es)) != 0 {
			t.Errorf("%s: the faulted call left something behind: %d names, %d agents", tc.name, len(a.Names), len(tr.Agents))
		}
	}
}

// A child is refused, never its parent: one passed nothing, or the actor in
// its own reach, or past the depth bound. The parent's model goes on.
func TestStepChildRefused(t *testing.T) {
	for _, tc := range []struct {
		name  string
		actor Contract
		depth int
		pass  string
	}{
		{"nothing passed", self, 3, `[]`},
		{"the actor in its reach", z, 3, `[2]`},
		{"past the depth bound", self, 0, `[0]`},
	} {
		tr, _ := Step(NewTree(app, tc.depth), Started{Self: self, Reach: []Contract{x, y, z}, Kind: KindDelegate, Actor: tc.actor, Endowment: shapeS})
		tr, es := reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":`+tc.pass+`}}]`)
		if agent(t, tr, "0.0").Phase != PhaseFaulted {
			t.Errorf("%s: the child is not refused", tc.name)
		}
		if agent(t, tr, "0").Phase != PhaseThinking {
			t.Errorf("%s: the parent does not go on", tc.name)
		}
		turnFor(t, es, "0")
	}
}

// The root's kind is given, never counted; its self check still holds.
func TestStepRootKind(t *testing.T) {
	tr, es := start(t, KindAct, []Contract{x, y}, 3)
	if a := agent(t, tr, "0"); a.Kind != KindAct || !reflect.DeepEqual(a.Offer.Invoke, []Contract{x, y}) {
		t.Fatalf("an act root over two contracts: %+v", a)
	}
	turnFor(t, es, "0")
	if tr, _ := start(t, KindDelegate, []Contract{x, selfOther}, 3); agent(t, tr, "0").Phase != PhaseFaulted {
		t.Error("a root reaching its own binary")
	}
	if tr, _ := Step(NewTree(app, 3), Started{Self: self, Kind: 0, Actor: self, Endowment: shapeS}); agent(t, tr, "0").Phase != PhaseFaulted {
		t.Error("a root with no kind")
	}
}

// Nothing comes from nowhere: an act node makes its input, then dispatches
// it. Values never enter a type: two inputs of different text are one input
// base type; what tells them apart is CEL's.
func TestStepCreateInput(t *testing.T) {
	tr, _ := start(t, KindAct, []Contract{x}, 3)
	tr, es := reply(tr, "0", `[{"create_input":"a"},{"perform_effect":{"contract":0,"input":{"earlier":0}}},{"create_input":"b"},{"perform_effect":{"contract":0,"input":{"earlier":2}}}]`)
	ic := dispatches(es)
	if len(ic) != 2 || ic[0].Input != 1 || ic[1].Input != 2 {
		t.Fatalf("the dispatches: %+v", ic)
	}
	w := agent(t, tr, "0").Waiting
	if len(w) != 2 || w[0].Input != w[1].Input {
		t.Fatalf("two inputs of one type, two input base types: %+v", w)
	}
	// Both must come back before the model goes on.
	tr, es = Step(tr, Returned{Agent: "0", Element: 1, Shape: shapeA})
	if len(turns(es)) != 0 {
		t.Fatal("a turn before every dispatch came back")
	}
	// One input base type, one shape back: the second is seen, and its guard
	// is asked before the model goes on.
	tr, es = Step(tr, Returned{Agent: "0", Element: 3, Shape: shapeA})
	if len(reads(es)) != 1 || len(turns(es)) != 0 {
		t.Fatalf("the second return: %+v", es)
	}
	_, es = Step(tr, ReadResult{Found: [][]byte{[]byte("false")}})
	if g := records(es, "guard"); len(g) != 1 || g[0].Body["ret"] != "fallback" {
		t.Fatalf("the guard broke: %+v", g)
	}
	turnFor(t, es, "0")
}

// A seen return base type asks for its guard's answer, and the model waits
// on it; the answer is recorded, and with remember off the model goes on.
func TestStepSeenAsksTheGuard(t *testing.T) {
	tr, _ := start(t, KindDelegate, []Contract{x, y}, 3)
	tr, es := reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`)
	inv := records(es, "invoke")
	if len(inv) != 2 || inv[0].Body["do"] != "unseen" || inv[1].Body["do"] != "seen" {
		t.Fatalf("two children of one input: %+v", inv)
	}
	for _, c := range []AgentID{"0.0", "0.1"} {
		tr, _ = reply(tr, c, `[{"perform_effect":{"contract":0,"input":2}}]`)
	}
	tr, es = Step(tr, Returned{Agent: "0.0", Element: 0, Shape: shapeA})
	if len(reads(es)) != 0 {
		t.Fatal("an unseen base type asked for a guard")
	}
	tr, es = Step(tr, Returned{Agent: "0.1", Element: 0, Shape: shapeA})
	rd := reads(es)
	if len(rd) != 1 || rd[0].CEL != "" || len(turns(es)) != 0 {
		t.Fatalf("a seen base type: reads %+v, turns %d", rd, len(turns(es)))
	}
	if _, last := es[len(es)-1].(Read); !last {
		t.Fatal("the read is not the last effect")
	}
	// While the read is in flight, nothing else is taken.
	if tr2, es2 := Step(tr, Stopped{Agent: "0.0", Reason: "x"}); agent(t, tr2, "0.0").Phase == PhaseFaulted || len(records(es2, "ignored")) != 1 {
		t.Fatal("an event taken while a read is in flight")
	}
	tr2, es := Step(tr, ReadResult{Found: [][]byte{[]byte("true")}})
	if g := records(es, "guard"); len(g) != 1 || g[0].Body["ret"] != "continue" {
		t.Fatalf("the answer: %+v", g)
	}
	turnFor(t, es, "0.1")
	if _, es := Step(tr, ReadResult{Found: [][]byte{[]byte("maybe")}}); len(records(es, "fault")) != 1 {
		t.Fatal("an answer that is not true or false")
	}
	_ = tr2
}

// What comes back must be well formed, and a failure or a stop is a fault.
func TestStepBackFaults(t *testing.T) {
	for _, tc := range []struct {
		name string
		e    Event
	}{
		{"a shape not well formed", Returned{Agent: "0", Element: 0, Shape: ""}},
		{"an error", Returned{Agent: "0", Element: 0, Shape: shapeA, Err: "boom"}},
		{"nothing in flight", Returned{Agent: "0", Element: 5, Shape: shapeA}},
		{"stopped", Stopped{Agent: "0", Reason: "timeout"}},
	} {
		tr, _ := start(t, KindAct, []Contract{x}, 3)
		tr, _ = reply(tr, "0", `[{"create_input":"a"},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
		e := tc.e
		if r, ok := e.(Returned); ok && r.Element == 0 {
			r.Element = 1
			e = r
		}
		if tr, _ := Step(tr, e); agent(t, tr, "0").Phase != PhaseFaulted {
			t.Errorf("%s: not a fault", tc.name)
		}
	}
}

// A faulted child gives its parent nothing; the parent goes on.
func TestStepChildFaultGoesUp(t *testing.T) {
	tr, _ := start(t, KindDelegate, []Contract{x, y}, 3)
	tr, _ = reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`)
	tr, es := Step(tr, Stopped{Agent: "0.0", Reason: "crash"})
	if agent(t, tr, "0.0").Phase != PhaseFaulted || len(agent(t, tr, "0").Names) != 2 {
		t.Fatal("the fault did not go up as nothing")
	}
	turnFor(t, es, "0")
}

// Step never changes the tree it is handed, and the same events give the
// same tree and the same effects: replay re-derives it.
func TestStepIsPureAndReplays(t *testing.T) {
	events := []Event{
		Replied{Agent: "0", Out: []byte(`[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`)},
		Replied{Agent: "0.0", Out: []byte(`[{"perform_effect":{"contract":0,"input":2}}]`)},
		Returned{Agent: "0.0", Element: 0, Shape: shapeA},
		Replied{Agent: "0.0", Out: []byte(`[{"return":{"members":[3]}}]`)},
		Replied{Agent: "0", Out: []byte(`[{"return":{"members":[2]}}]`)},
	}
	play := func() (Tree, [][]Effect) {
		tr, es := start(t, KindDelegate, []Contract{x, y}, 3)
		all := [][]Effect{es}
		for _, e := range events {
			kept := tr.clone()
			next, es := Step(tr, e)
			if !reflect.DeepEqual(tr, kept) {
				t.Fatalf("Step changed the tree it was handed, on %T", e)
			}
			tr = next
			all = append(all, es)
		}
		return tr, all
	}
	t1, e1 := play()
	t2, e2 := play()
	if !t1.Done() || !reflect.DeepEqual(t1, t2) || !reflect.DeepEqual(e1, e2) {
		t.Fatal("the same events gave a different tree or different effects")
	}
}
