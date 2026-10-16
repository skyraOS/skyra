package proc

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"
)

var greet = &Schema{Title: "greet", Type: "object", Required: []string{"names", "greeting"}, Fields: map[string]*Schema{
	"names": {Type: "array", Items: &Schema{Type: "string"}}, "greeting": {Type: "string"}}}

func raw(s string) json.RawMessage { return json.RawMessage(s) }

// world is a stand-in for the kernel and every target: it answers each
// instruction of an emission, in order, as a step.
func world(t *testing.T, emission []byte) []Step {
	t.Helper()
	var ins []struct {
		To      string          `json:"to"`
		Payload json.RawMessage `json:"payload"`
	}
	if emission[0] == '[' {
		if err := json.Unmarshal(emission, &ins); err != nil {
			t.Fatal(err)
		}
	} else {
		ins = make([]struct {
			To      string          `json:"to"`
			Payload json.RawMessage `json:"payload"`
		}, 1)
		if err := json.Unmarshal(emission, &ins[0]); err != nil {
			t.Fatal(err)
		}
	}
	var steps []Step
	for _, i := range ins {
		st := Step{To: i.To, Payload: i.Payload, State: "return"}
		switch {
		case i.To == "echo" || strings.HasPrefix(i.To, "system.ref."):
			var p struct{ Text string }
			_ = json.Unmarshal(i.Payload, &p)
			st.Value, _ = json.Marshal(p.Text)
		case i.To == "system":
			st.Value = raw(`{"reference":"system.ref.1"}`)
		case i.To == "shell":
			st.State, st.Value, st.Fault = "fault", raw(`{"to":"shell"}`), &Fault{Kind: "horizon", Cause: `"shell" is not in reach`}
		default:
			t.Fatalf("no such target %q", i.To)
		}
		steps = append(steps, st)
	}
	return steps
}

// drive runs a procedure to its end as the kernel would, one resume per
// round, each one starting from nothing but the store and the steps: a new
// process every time. It returns every emission and the return.
func drive(t *testing.T, src Source, input string) ([]string, Next) {
	t.Helper()
	store := Store{Dir: t.TempDir()}
	p, nodes, err := Compile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(p, nodes); err != nil {
		t.Fatal(err)
	}
	const frame = "f000000000000000000000000000000000000000000000000000000000000000"
	var emissions []string
	var steps []Step
	for i := 0; i < 20; i++ {
		p, nodes, err := store.Procedure(src.Contract) // a resume reads everything again
		if err != nil {
			t.Fatal(err)
		}
		pos, err := store.Position(frame)
		if err != nil {
			t.Fatal(err)
		}
		pos.Procedure = Key(src.Contract)
		next, err := Plan(p, nodes, raw(input), pos, steps)
		if err != nil {
			t.Fatal(err)
		}
		if next.Return != nil || next.Stuck != nil {
			return emissions, next
		}
		pos.Rounds = append(pos.Rounds, next.Round)
		if err := store.SetPosition(frame, pos); err != nil {
			t.Fatal(err)
		}
		emissions = append(emissions, string(next.Emission()))
		steps = append(steps, world(t, next.Emission())...)
	}
	t.Fatal("the procedure never returned")
	return nil, Next{}
}

func b(root string, path ...any) json.RawMessage {
	out, _ := json.Marshal(Binding(root, path...))
	return out
}

// Each node shape: a fork and an each in one round, a sequence after it, and
// a return built from what came back.
func TestNodeShapes(t *testing.T) {
	src := Source{Contract: greet, Order: []string{"each", "fork", "then"},
		Nodes: map[string]SourceNode{
			"each": {Call: raw(`"echo"`), In: raw(`{"text":` + string(b("item")) + `}`), Each: b("input", "names")},
			"fork": {Call: raw(`"echo"`), In: raw(`{"text":` + string(b("input", "greeting")) + `}`)},
			"then": {Call: raw(`"echo"`), In: raw(`{"text":` + string(b("fork")) + `}`), After: []string{"fork"}},
		},
		Return: raw(`{"greeting":` + string(b("then")) + `,"names":` + string(b("each")) + `,"fixed":7}`),
	}
	emissions, next := drive(t, src, `{"names":["ada","bo"],"greeting":"hi"}`)
	want := []string{
		`[{"payload":{"text":"ada"},"to":"echo"},{"payload":{"text":"bo"},"to":"echo"},{"payload":{"text":"hi"},"to":"echo"}]`,
		`{"payload":{"text":"hi"},"to":"echo"}`,
	}
	if !reflect.DeepEqual(emissions, want) {
		t.Fatalf("a fork and an each fan out in one round, the sequence follows alone:\n%s", strings.Join(emissions, "\n"))
	}
	if string(next.Return) != `{"fixed":7,"greeting":"hi","names":["ada","bo"]}` {
		t.Fatalf("the return is built from what came back, constants kept: %s", next.Return)
	}

	// An empty each discharges at once, with nothing emitted for it.
	emissions, next = drive(t, src, `{"names":[],"greeting":"hi"}`)
	if len(emissions) != 2 || emissions[0] != `{"payload":{"text":"hi"},"to":"echo"}` || string(next.Return) != `{"fixed":7,"greeting":"hi","names":[]}` {
		t.Fatalf("an empty each: %v %s", emissions, next.Return)
	}
}

// Creating a Horizon inside a procedure (decisions 25, 26): narrow, then a
// node whose call is the reference it returned. A narrow goes alone, even
// when other calls are ready: the kernel refuses it among siblings.
func TestNarrowThenCallTheReference(t *testing.T) {
	src := Source{Contract: greet, Order: []string{"narrow", "other", "call"},
		Nodes: map[string]SourceNode{
			"narrow": {Call: raw(`"system"`), In: raw(`{"narrow":{"actor":"echo","reach":[]}}`)},
			"other":  {Call: raw(`"echo"`), In: raw(`{"text":"meanwhile"}`)},
			"call":   {Call: b("narrow", "reference"), In: raw(`{"text":` + string(b("input", "greeting")) + `}`), After: []string{"narrow"}},
		},
		Return: b("call"),
	}
	emissions, next := drive(t, src, `{"names":[],"greeting":"hi"}`)
	want := []string{
		`{"payload":{"narrow":{"actor":"echo","reach":[]}},"to":"system"}`,
		`[{"payload":{"text":"meanwhile"},"to":"echo"},{"payload":{"text":"hi"},"to":"system.ref.1"}]`,
	}
	if !reflect.DeepEqual(emissions, want) || string(next.Return) != `"hi"` {
		t.Fatalf("narrow alone, then the call through its reference:\n%s\n%s", strings.Join(emissions, "\n"), next.Return)
	}
}

// A call that faults stops the procedure as compiled: the plan says where.
func TestAFaultedCallIsWhereTheProcedureStops(t *testing.T) {
	src := Source{Contract: greet, Nodes: map[string]SourceNode{
		"a": {Call: raw(`"shell"`), In: raw(`{}`)},
		"z": {Call: raw(`"echo"`), In: raw(`{"text":"x"}`), After: []string{"a"}},
	}, Return: b("z")}
	_, next := drive(t, src, `{"names":[],"greeting":"hi"}`)
	if next.Stuck == nil || next.Stuck.Step.Fault.Kind != "horizon" {
		t.Fatalf("stuck at the faulted call: %+v", next)
	}
}

// Steps that do not match where the frame is are refused, never guessed at.
func TestPlanRefusesStepsThatDoNotMatch(t *testing.T) {
	p, nodes, _ := Compile(Source{Contract: greet, Nodes: map[string]SourceNode{"a": {Call: raw(`"echo"`), In: raw(`{"text":"x"}`)}}, Return: b("a")})
	if _, err := Plan(p, nodes, raw(`{}`), Position{}, []Step{{State: "return", Value: raw(`"x"`)}}); err == nil {
		t.Fatal("a step the frame never emitted")
	}
}

// Nodes are content-addressed: the id covers the call, the input and the
// dependencies, whatever their key order, and a procedure shares nodes with
// any other that has the same calls.
func TestNodesAreContentAddressed(t *testing.T) {
	a := Node{Call: raw(`"echo"`), In: raw(`{"text":"x","n":1}`), After: []string{"b", "a"}}
	b := Node{Call: raw(`"echo"`), In: raw(`{"n":1,"text":"x"}`), After: []string{"a", "b"}}
	if a.ID() != b.ID() {
		t.Fatal("key order and dependency order change nothing")
	}
	c := a
	c.After = []string{"a"}
	if c.ID() == a.ID() {
		t.Fatal("a node's id covers its dependencies")
	}
	if _, _, err := Compile(Source{Contract: greet, Nodes: map[string]SourceNode{
		"a": {Call: raw(`"echo"`), In: raw(`{"text":` + string(b2("z")) + `}`)},
		"z": {Call: raw(`"echo"`), In: raw(`{"text":"x"}`)},
	}, Return: b2("a")}); err == nil {
		t.Fatal("a binding to a node it does not run after")
	}
	if _, _, err := Compile(Source{Contract: greet, Nodes: map[string]SourceNode{
		"a": {Call: raw(`"echo"`), In: raw(`{}`), After: []string{"b"}},
		"b": {Call: raw(`"echo"`), In: raw(`{}`), After: []string{"a"}},
	}, Return: raw(`null`)}); err == nil {
		t.Fatal("a cycle")
	}
}

func b2(root string, path ...any) json.RawMessage { return b(root, path...) }

// One representation per contract (decision 28): a second procedure for the
// same contract is refused.
func TestOneRepresentationPerContract(t *testing.T) {
	store := Store{Dir: t.TempDir()}
	src := Source{Contract: greet, Nodes: map[string]SourceNode{"a": {Call: raw(`"echo"`), In: raw(`{"text":"x"}`)}}, Return: b("a")}
	p, nodes, _ := Compile(src)
	if err := store.Put(p, nodes); err != nil {
		t.Fatal(err)
	}
	if err := store.Put(p, nodes); err == nil || !strings.Contains(err.Error(), "one representation") {
		t.Fatalf("a second representation: %v", err)
	}
	if !store.Has(greet) || store.Has(&Schema{Type: "string"}) {
		t.Fatal("has")
	}
	restated := *greet
	restated.Required = []string{"greeting", "names"}
	if !store.Has(&restated) {
		t.Fatal("the same contract with its required fields in another order is already held")
	}
	p2, nodes2, _ := Compile(Source{Contract: &restated, Nodes: src.Nodes, Return: src.Return})
	if err := store.Put(p2, nodes2); err == nil || !strings.Contains(err.Error(), "one representation") {
		t.Fatalf("a restated contract is a second representation: %v", err)
	}
}

// A contract's identity ignores the order of its required fields, at every
// depth, and repeats in them; its title and its shape still count.
func TestContractIdentityIgnoresRequiredOrder(t *testing.T) {
	obj := func(title string, req ...string) *Schema {
		return &Schema{Title: title, Type: "object", Required: req, Fields: map[string]*Schema{
			"a": {Type: "string"}, "b": {Type: "string"}}}
	}
	same := [][2]*Schema{
		{obj("", "a", "b"), obj("", "b", "a")},
		{obj("", "a", "b"), obj("", "b", "a", "b")},
		{&Schema{Type: "object", Fields: map[string]*Schema{"x": obj("", "a", "b")}},
			&Schema{Type: "object", Fields: map[string]*Schema{"x": obj("", "b", "a")}}},
		{&Schema{Type: "array", Items: obj("", "a", "b")}, &Schema{Type: "array", Items: obj("", "b", "a")}},
	}
	for i, c := range same {
		if !Same(c[0], c[1]) || Key(c[0]) != Key(c[1]) {
			t.Fatalf("%d: %s and %s are one contract", i, c[0].Canonical(), c[1].Canonical())
		}
	}
	different := [][2]*Schema{
		{obj("", "a"), obj("", "a", "b")},
		{obj("greet", "a", "b"), obj("wave", "a", "b")},
	}
	for i, c := range different {
		if Same(c[0], c[1]) || Key(c[0]) == Key(c[1]) {
			t.Fatalf("%d: %s and %s are different contracts", i, c[0].Canonical(), c[1].Canonical())
		}
	}
	in := obj("", "b", "a")
	in.Canonical()
	if !slices.Equal(in.Required, []string{"b", "a"}) {
		t.Fatal("canonicalizing leaves the contract as written")
	}
}

// Resuming from state after a simulated crash: the process that planned the
// first round is gone; a new one, with nothing but the state directory and
// the steps the kernel delivers, decides exactly what the old one would have.
func TestResumeFromStateAfterACrash(t *testing.T) {
	src := Source{Contract: greet, Nodes: map[string]SourceNode{
		"a": {Call: raw(`"echo"`), In: raw(`{"text":"one"}`)},
		"z": {Call: raw(`"echo"`), In: raw(`{"text":` + string(b("a")) + `}`), After: []string{"a"}},
	}, Return: b("z")}
	dir := t.TempDir()
	const frame = "a000000000000000000000000000000000000000000000000000000000000000"
	{ // the first process: plans round one, records where it is, emits, and dies
		store := Store{Dir: dir}
		p, nodes, _ := Compile(src)
		if err := store.Put(p, nodes); err != nil {
			t.Fatal(err)
		}
		next, err := Plan(p, nodes, raw(`{}`), Position{}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := store.SetPosition(frame, Position{Procedure: Key(greet), Rounds: [][]Call{next.Round}}); err != nil {
			t.Fatal(err)
		}
	}
	store := Store{Dir: dir} // the next process
	p, nodes, err := store.Procedure(greet)
	if err != nil {
		t.Fatal(err)
	}
	pos, err := store.Position(frame)
	if err != nil || len(pos.Rounds) != 1 {
		t.Fatalf("where the frame is survives the process: %+v %v", pos, err)
	}
	next, err := Plan(p, nodes, raw(`{}`), pos, []Step{{To: "echo", State: "return", Value: raw(`"one"`)}})
	if err != nil || string(next.Emission()) != `{"payload":{"text":"one"},"to":"echo"}` {
		t.Fatalf("it goes on where the last one stopped: %s %v", next.Emission(), err)
	}
}

// An each over repeated items calls each of them: identical instructions in
// one round are made distinct bytes, so each is its own record.
func TestRepeatedCallsAreEachTheirOwn(t *testing.T) {
	src := Source{Contract: greet, Nodes: map[string]SourceNode{
		"each": {Call: raw(`"echo"`), In: raw(`{"text":` + string(b("item")) + `}`), Each: b("input", "names")},
	}, Return: b("each")}
	emissions, next := drive(t, src, `{"names":["a","a","a"],"greeting":"hi"}`)
	var ins []json.RawMessage
	if err := json.Unmarshal([]byte(emissions[0]), &ins); err != nil || len(ins) != 3 {
		t.Fatalf("%s", emissions[0])
	}
	if string(ins[0]) == string(ins[1]) || string(ins[1]) == string(ins[2]) || string(ins[0]) == string(ins[2]) {
		t.Fatalf("every call is its own bytes: %s", emissions[0])
	}
	if string(next.Return) != `["a","a","a"]` {
		t.Fatalf("and every call ran: %s", next.Return)
	}
}

// The kernel's normal form and this one are the same (42): both reproduce
// the cases in the kernel's testdata/normal.json.
func TestCanonicalIsTheKernelsNormalForm(t *testing.T) {
	b, err := os.ReadFile("../../core/contract/testdata/normal.json")
	if err != nil {
		t.Fatal(err)
	}
	var cases []struct {
		Case       string
		In, Normal *Schema
	}
	if err := json.Unmarshal(b, &cases); err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		want, _ := json.Marshal(c.Normal)
		if string(c.In.Canonical()) != string(want) {
			t.Errorf("%s: %s, want %s", c.Case, c.In.Canonical(), want)
		}
	}
}
