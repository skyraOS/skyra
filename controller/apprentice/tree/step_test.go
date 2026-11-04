// Step, driven by events alone: no shell, no model. A fake shell plays the
// apprentice's loop, answering every read and commit from its own records,
// and checks on every Step that it returned at most one effect and left the
// tree it was handed unchanged.
package tree

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

// shell holds the records that outlive a run: the input records, the guard
// chains, and record/1.
type shell struct {
	t        *testing.T
	history  []string
	chains   map[ReturnBase][]json.RawMessage
	guards   map[RefinementID][]byte
	records  []RecordWrite
	holds    bool // every guard's answer
	conflict bool // every guard write conflicts
}

func newShell(t *testing.T) *shell {
	return &shell{t: t, chains: map[ReturnBase][]json.RawMessage{}, guards: map[RefinementID][]byte{}, holds: true}
}

// step is one event, and every read and commit it leads to, answered. It
// returns every effect Step asked for; the last is an Execute, or nothing.
func (s *shell) step(tr Tree, e Event) (Tree, []Effect) {
	s.t.Helper()
	var all []Effect
	for {
		kept := tr.clone()
		next, es := Step(tr, e)
		if !reflect.DeepEqual(tr, kept) {
			s.t.Fatalf("Step changed the tree it was handed, on %T", e)
		}
		if len(es) > 1 {
			s.t.Fatalf("Step returned %d effects on %T", len(es), e)
		}
		tr, all = next, append(all, es...)
		if len(es) == 0 {
			return tr, all
		}
		switch ef := es[0].(type) {
		case Read:
			e = s.read(ef)
		case Commit:
			e = s.commit(ef)
		case Execute:
			if tr.Awaiting() {
				s.t.Fatal("an execute while awaiting")
			}
			return tr, all
		}
	}
}

func (s *shell) read(r Read) Event {
	var found [][]byte
	for _, k := range r.Keys {
		switch k := k.(type) {
		case HistoryKey:
			if len(s.history) == 0 {
				found = append(found, nil)
			} else {
				found = append(found, mustJSON(s.history))
			}
		case ChainKey:
			if c := s.chains[k.Base]; len(c) == 0 {
				found = append(found, nil)
			} else {
				found = append(found, mustJSON(c))
			}
		case CELKey:
			found = append(found, mustJSON(s.holds))
		}
	}
	return ReadResult{found}
}

func (s *shell) commit(c Commit) Event {
	var out []Outcome
	for _, w := range c.Writes {
		switch w := w.(type) {
		case RecordWrite:
			s.records = append(s.records, w)
			out = append(out, OutcomeCreated)
		case HistoryWrite:
			var id string
			json.Unmarshal(w.Bytes, &id)
			s.history = append(s.history, id)
			out = append(out, OutcomeCreated)
		case GuardWrite:
			old, ok := s.guards[w.ID]
			switch {
			case s.conflict || ok && !bytes.Equal(old, w.Bytes):
				out = append(out, OutcomeConflict)
			case ok:
				out = append(out, OutcomeIdentical)
			default:
				s.guards[w.ID] = w.Bytes
				s.chains[w.Base] = append(s.chains[w.Base], json.RawMessage(w.Bytes))
				out = append(out, OutcomeCreated)
			}
		}
	}
	return Committed{out}
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

// start is a tree of the test apprentice, started: the root is self, with
// the reach and kind given; self is also the actor that runs every child.
func (s *shell) start(k Kind, reach []Contract, depth int) (Tree, []Effect) {
	s.t.Helper()
	return s.step(NewTree(app, depth), Started{Self: self, Reach: reach, Kind: k, Actor: self, Endowment: shapeS})
}

func (s *shell) reply(tr Tree, id AgentID, out string) (Tree, []Effect) {
	s.t.Helper()
	return s.step(tr, Replied{Agent: id, Out: []byte(out)})
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

func cels(es []Effect) []CELKey {
	var out []CELKey
	for _, e := range es {
		if rd, ok := e.(Read); ok {
			for _, k := range rd.Keys {
				if c, ok := k.(CELKey); ok {
					out = append(out, c)
				}
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

// The tree of TestStepRunsATree, by its events after Started.
var aTree = []Event{
	Replied{Agent: "0", Out: []byte(`[{"goal":"summarize the logs"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`)},
	Replied{Agent: "0.0", Out: []byte(`[{"perform_effect":{"contract":0,"input":2}}]`)},
	Returned{Agent: "0.0", Element: 0, Shape: shapeA},
	Replied{Agent: "0.0", Out: []byte(`[{"return":{"members":[3]}}]`)},
	Replied{Agent: "0", Out: []byte(`[{"return":{"members":[2]}}]`)},
}

// A whole tree, end to end: a delegate root makes a goal and delegates it
// with one contract; the act transformation provisions an act child, which
// dispatches its goal through perform_effect and returns what came back;
// the root returns its child's return.
func TestStepRunsATree(t *testing.T) {
	sh := newShell(t)
	tr, es := sh.start(KindDelegate, []Contract{x, y}, 3)
	root := turnFor(t, es, "0")
	if !reflect.DeepEqual(root.Offer, Offered(app, KindDelegate, []Contract{x, y})) || root.Binary != self.binary {
		t.Fatalf("the root's turn: %+v", root)
	}

	tr, es = sh.step(tr, aTree[0])
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

	tr, es = sh.step(tr, aTree[1])
	ic := dispatches(es)
	if len(ic) != 1 || ic[0].Contract != x || ic[0].Input != 2 || ic[0].Element != 0 {
		t.Fatalf("the dispatch: %+v", ic)
	}

	tr, es = sh.step(tr, aTree[2])
	if g := records(es, "guard"); len(g) != 1 || g[0].Body["see"] != "unseen" {
		t.Fatalf("the first return is unseen: %+v", g)
	}
	turnFor(t, es, "0.0")
	came := agent(t, tr, "0.0").Names[3]
	if came.Obs.typ.contract != x {
		t.Fatalf("what came back is not typed by its contract: %+v", came)
	}

	tr, es = sh.step(tr, aTree[3])
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

	tr, es = sh.step(tr, aTree[4])
	if !tr.Done() || agent(t, tr, "0").Phase != PhaseDone {
		t.Fatal("the tree is not done")
	}
	if len(turns(es)) != 0 || len(records(es, "done")) != 1 {
		t.Fatalf("after the root returns: %+v", es)
	}
}

// Nothing leaves the process before its records are durable: every Execute
// comes after the Commit of what decided it, and every input made is in the
// history.
func TestStepCommitsBeforeExecuting(t *testing.T) {
	sh := newShell(t)
	tr, es := sh.start(KindDelegate, []Contract{x, y}, 3)
	all := es
	for _, e := range aTree {
		tr, es = sh.step(tr, e)
		all = append(all, es...)
	}
	for i, e := range all {
		if _, ok := e.(Execute); ok {
			if _, ok := all[i-1].(Commit); !ok {
				t.Fatalf("effect %d, an Execute, does not follow a Commit: %T", i, all[i-1])
			}
		}
	}
	if len(sh.history) != len(tr.History) || len(sh.history) == 0 {
		t.Fatalf("the history: %v, the shell's %v", tr.History, sh.history)
	}
	if len(sh.chains) == 0 {
		t.Fatal("no guard was written")
	}
}

// A second run over the same records: the root's input and every return
// base type are seen, so every return asks its guard.
func TestStepRemembersAcrossRuns(t *testing.T) {
	sh := newShell(t)
	tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3)
	for _, e := range aTree {
		tr, _ = sh.step(tr, e)
	}
	tr, es := sh.start(KindDelegate, []Contract{x, y}, 3)
	if inv := records(es, "invoke"); len(inv) != 1 || inv[0].Body["do"] != "seen" {
		t.Fatalf("the root's input on the second run: %+v", inv)
	}
	sh.holds = false
	asked := 0
	for _, e := range aTree {
		tr, es = sh.step(tr, e)
		asked += len(cels(es))
		for _, g := range records(es, "guard") {
			if g.Body["see"] != "seen" || g.Body["ret"] != "fallback" {
				t.Fatalf("a guard on the second run: %+v", g)
			}
		}
	}
	if asked != 2 || !tr.Done() {
		t.Fatalf("%d guards asked; done %v", asked, tr.Done())
	}
}

// A delegate child: two contracts passed, so create_delegation and the
// packaged reach, and the contracts pass-only.
func TestStepDelegateChild(t *testing.T) {
	sh := newShell(t)
	tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3)
	tr, _ = sh.reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0,1]}}]`)
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
	sh := newShell(t)
	tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3)
	tr2, _ := sh.reply(tr, "0", `[{"goal":"g"}]`)
	turned := agent(t, tr2, "0")
	if turned.Phase != PhaseThinking || len(turned.Names) != 2 {
		t.Fatalf("a call of a goal alone goes back to the model: %+v", turned)
	}
	if tr3, _ := sh.reply(tr2, "0", `[{"return":{"members":[1]}}]`); agent(t, tr3, "0").Phase != PhaseFaulted {
		t.Error("the root returned its own goal")
	}
	if tr3, _ := sh.reply(tr, "0", `[{"return":{"members":[0]}}]`); agent(t, tr3, "0").Phase != PhaseFaulted {
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
		sh := newShell(t)
		tr, _ := sh.start(tc.kind, []Contract{x, y}, 3)
		before, history := agent(t, tr, "0"), len(sh.history)
		tr, es := sh.reply(tr, "0", tc.out)
		a := agent(t, tr, "0")
		if a.Phase != PhaseFaulted || len(records(es, "fault")) != 1 {
			t.Errorf("%s: the root is %v, not faulted", tc.name, a.Phase)
			continue
		}
		if len(a.Names) != len(before.Names) || len(tr.Agents) != 1 || len(turns(es)) != 0 ||
			len(dispatches(es)) != 0 || len(sh.history) != history || len(records(es, "name")) != 0 {
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
		sh := newShell(t)
		tr, _ := sh.step(NewTree(app, tc.depth), Started{Self: self, Reach: []Contract{x, y, z}, Kind: KindDelegate, Actor: tc.actor, Endowment: shapeS})
		tr, es := sh.reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":`+tc.pass+`}}]`)
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
	sh := newShell(t)
	tr, es := sh.start(KindAct, []Contract{x, y}, 3)
	if a := agent(t, tr, "0"); a.Kind != KindAct || !reflect.DeepEqual(a.Offer.Invoke, []Contract{x, y}) {
		t.Fatalf("an act root over two contracts: %+v", a)
	}
	turnFor(t, es, "0")
	if tr, _ := sh.start(KindDelegate, []Contract{x, selfOther}, 3); agent(t, tr, "0").Phase != PhaseFaulted {
		t.Error("a root reaching its own binary")
	}
	if tr, _ := sh.step(NewTree(app, 3), Started{Self: self, Kind: 0, Actor: self, Endowment: shapeS}); agent(t, tr, "0").Phase != PhaseFaulted {
		t.Error("a root with no kind")
	}
}

// Nothing comes from nowhere: an act node makes its input, then dispatches
// it. Values never enter a type: two inputs of different text are one input
// base type; what tells them apart is CEL's.
func TestStepCreateInput(t *testing.T) {
	sh := newShell(t)
	tr, _ := sh.start(KindAct, []Contract{x}, 3)
	tr, es := sh.reply(tr, "0", `[{"create_input":"a"},{"perform_effect":{"contract":0,"input":{"earlier":0}}},{"create_input":"b"},{"perform_effect":{"contract":0,"input":{"earlier":2}}}]`)
	ic := dispatches(es)
	if len(ic) != 2 || ic[0].Input != 1 || ic[1].Input != 2 {
		t.Fatalf("the dispatches: %+v", ic)
	}
	w := agent(t, tr, "0").Waiting
	if len(w) != 2 || w[0].Input != w[1].Input {
		t.Fatalf("two inputs of one type, two input base types: %+v", w)
	}
	// Both must come back before the model goes on.
	tr, es = sh.step(tr, Returned{Agent: "0", Element: 1, Shape: shapeA})
	if len(turns(es)) != 0 {
		t.Fatal("a turn before every dispatch came back")
	}
	// One input base type, one shape back: the second is seen, and its guard
	// is asked before the model goes on.
	sh.holds = false
	_, es = sh.step(tr, Returned{Agent: "0", Element: 3, Shape: shapeA})
	if g := records(es, "guard"); len(cels(es)) != 1 || len(g) != 1 || g[0].Body["ret"] != "fallback" {
		t.Fatalf("the second return: %+v", es)
	}
	turnFor(t, es, "0")
}

// While a read is in flight only its answer is taken: anything else leaves
// the tree as it was and asks for nothing; the shell holds it back. A seen
// base type's guard is asked, and with remember off the model goes on.
func TestStepAwaiting(t *testing.T) {
	sh := newShell(t)
	tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3)
	tr, es := sh.reply(tr, "0", `[{"goal":"g"},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}},{"delegate":{"observations":[{"earlier":0}],"contracts":[0]}}]`)
	if inv := records(es, "invoke"); len(inv) != 2 || inv[0].Body["do"] != "unseen" || inv[1].Body["do"] != "seen" {
		t.Fatalf("two children of one input: %+v", inv)
	}
	for _, c := range []AgentID{"0.0", "0.1"} {
		tr, _ = sh.reply(tr, c, `[{"perform_effect":{"contract":0,"input":2}}]`)
	}
	tr, _ = sh.step(tr, Returned{Agent: "0.0", Element: 0, Shape: shapeA})
	tr, es = Step(tr, Returned{Agent: "0.1", Element: 0, Shape: shapeA})
	if k := cels(es); len(k) != 1 || k[0].CEL != "" || !tr.Awaiting() {
		t.Fatalf("a seen base type asks its default guard: %+v", es)
	}
	if held, es := Step(tr, Stopped{Agent: "0.0", Reason: "x"}); !reflect.DeepEqual(held, tr) || es != nil {
		t.Fatal("an event taken while a read is in flight")
	}
	_, es = sh.step(tr, ReadResult{Found: [][]byte{[]byte("true")}})
	if g := records(es, "guard"); len(g) != 1 || g[0].Body["ret"] != "continue" {
		t.Fatalf("the answer: %+v", g)
	}
	turnFor(t, es, "0.1")
	if tr2, _ := sh.step(tr, ReadResult{Found: [][]byte{[]byte("maybe")}}); agent(t, tr2, "0.1").Phase != PhaseFaulted {
		t.Fatal("an answer that is not true or false")
	}
}

// Every record read is verified: a corrupt input record faults the root, a
// corrupt guard chain the agent that read it, and a conflicting write whose
// write it was.
func TestStepVerifiesWhatItReads(t *testing.T) {
	sh := newShell(t)
	sh.history = []string{"sha256:short"}
	if tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3); agent(t, tr, "0").Phase != PhaseFaulted {
		t.Error("a corrupt input record")
	}

	// One honest run, for a real chain, then the chain corrupted.
	honest := newShell(t)
	tr, _ := honest.start(KindDelegate, []Contract{x, y}, 3)
	for _, e := range aTree[:3] {
		tr, _ = honest.step(tr, e)
	}
	var base ReturnBase
	for b := range honest.chains {
		base = b
	}
	def := honest.chains[base][0]
	written := guardBytes(base, WrittenGuard("x"), RefinementOf(base, DefaultGuard()))
	for _, tc := range []struct {
		name  string
		chain []json.RawMessage
	}{
		{"not canonical: its keys out of order", []json.RawMessage{json.RawMessage(`{"replaces":"","guard":"default","base":"` + base.id + `"}`)}},
		{"another base type's", []json.RawMessage{guardBytes(ReturnBase{"sha256:" + strings.Repeat("9", 64)}, DefaultGuard(), "")}},
		{"a written guard first", []json.RawMessage{guardBytes(base, WrittenGuard("x"), "")}},
		{"a broken link", []json.RawMessage{def, guardBytes(base, WrittenGuard("x"), "sha256:nope")}},
		{"a second default", []json.RawMessage{def, def}},
	} {
		sh := newShell(t)
		sh.chains[base] = tc.chain
		tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3)
		for _, e := range aTree[:3] {
			tr, _ = sh.step(tr, e)
		}
		if agent(t, tr, "0.0").Phase != PhaseFaulted {
			t.Errorf("%s: the chain was taken", tc.name)
		}
	}
	// An honest chain with a rewrite: its written guard is the one asked.
	sh = newShell(t)
	sh.chains[base] = []json.RawMessage{def, written}
	tr, _ = sh.start(KindDelegate, []Contract{x, y}, 3)
	for _, e := range aTree[:2] {
		tr, _ = sh.step(tr, e)
	}
	tr, _ = Step(tr, aTree[2])
	_, es := Step(tr, ReadResult{Found: [][]byte{mustJSON(sh.chains[base])}})
	if k := cels(es); len(k) != 1 || k[0].CEL != "x" {
		t.Errorf("a rewritten chain: %+v", es)
	}

	sh = newShell(t)
	sh.conflict = true
	tr, _ = sh.start(KindDelegate, []Contract{x, y}, 3)
	for _, e := range aTree[:3] {
		tr, _ = sh.step(tr, e)
	}
	if agent(t, tr, "0.0").Phase != PhaseFaulted {
		t.Error("a conflicting guard write was taken")
	}
}

// What comes back must be well formed, and a failure or a stop is a fault.
func TestStepBackFaults(t *testing.T) {
	for _, tc := range []struct {
		name string
		e    Event
	}{
		{"a shape not well formed", Returned{Agent: "0", Element: 1, Shape: ""}},
		{"an error", Returned{Agent: "0", Element: 1, Shape: shapeA, Err: "boom"}},
		{"nothing in flight", Returned{Agent: "0", Element: 5, Shape: shapeA}},
		{"stopped", Stopped{Agent: "0", Reason: "timeout"}},
	} {
		sh := newShell(t)
		tr, _ := sh.start(KindAct, []Contract{x}, 3)
		tr, _ = sh.reply(tr, "0", `[{"create_input":"a"},{"perform_effect":{"contract":0,"input":{"earlier":0}}}]`)
		if tr, _ := sh.step(tr, tc.e); agent(t, tr, "0").Phase != PhaseFaulted {
			t.Errorf("%s: not a fault", tc.name)
		}
	}
}

// A faulted child gives its parent nothing; the parent goes on.
func TestStepChildFaultGoesUp(t *testing.T) {
	sh := newShell(t)
	tr, _ := sh.start(KindDelegate, []Contract{x, y}, 3)
	tr, _ = sh.step(tr, aTree[0])
	tr, es := sh.step(tr, Stopped{Agent: "0.0", Reason: "crash"})
	if agent(t, tr, "0.0").Phase != PhaseFaulted || len(agent(t, tr, "0").Names) != 2 {
		t.Fatal("the fault did not go up as nothing")
	}
	turnFor(t, es, "0")
}

// The same events over the same records give the same tree and the same
// effects: replay re-derives it.
func TestStepReplays(t *testing.T) {
	play := func() (Tree, [][]Effect) {
		sh := newShell(t)
		tr, es := sh.start(KindDelegate, []Contract{x, y}, 3)
		all := [][]Effect{es}
		for _, e := range aTree {
			tr, es = sh.step(tr, e)
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
