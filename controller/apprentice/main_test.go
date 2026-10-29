package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"skyra/apprentice/proc"
	"skyra/apprentice/remember"
	"skyra/apprentice/types"
)

var agentIn = &proc.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*proc.Schema{"text": {Type: "string"}}}

func fixture(t *testing.T) (remember.Trace, string) {
	t.Helper()
	b, err := os.ReadFile("remember/testdata/relay.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Observation string         `json:"observation"`
		Trace       remember.Trace `json:"trace"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f.Trace, f.Observation
}

func ask(obs, target string) json.RawMessage {
	b, _ := json.Marshal(map[string]any{"remember": map[string]any{"observation": obs, "contract": map[string]any{"target": target, "schema": agentIn}}})
	return b
}

func ret(v any) proc.Step {
	b, _ := json.Marshal(v)
	return proc.Step{State: "return", Value: b}
}

func tools() map[string]*proc.Schema {
	return map[string]*proc.Schema{"agent": agentIn, "echo": agentIn, "apprentice": rememberContract()}
}

// remember's body: the trace, the compile, the extension, then what was
// remembered. The first step is a syscall; the apprentice extends itself.
func TestRememberRunsItsBody(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	st, obs := fixture(t)
	e := envelope{Frame: strings.Repeat("1", 64), Input: ask(obs, "agent"), Reach: heldReach(tools(), nil)}
	out, err := runRemember(store, e)
	if err != nil || !strings.Contains(string(out), `"to":"system"`) || !strings.Contains(string(out), `"trace"`) {
		t.Fatalf("remember's first step is the trace syscall: %s %v", out, err)
	}
	e.Steps = []proc.Step{ret(st)}
	out, err = runRemember(store, e)
	if err != nil || !strings.Contains(string(out), `"export"`) || !store.Has(agentIn) {
		t.Fatalf("then it stores the procedure and extends itself: %s %v", out, err)
	}
	e.Steps = append(e.Steps, ret(map[string]bool{"ok": true}))
	out, err = runRemember(store, e)
	if err != nil || string(out) != `{"payload":{"remembered":["agent"]},"to":"`+e.Frame+`"}` {
		t.Fatalf("then it returns what it remembered: %s %v", out, err)
	}
}

// A second remember through the same contract is refused, from the same
// Boundary and from another (the store is the apprentice's, not a Boundary's),
// before anything is stored or extended. So is remembering a run of the
// procedure itself: a second representation of the same contract.
func TestRememberRefusesASecondRepresentation(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	st, obs := fixture(t)
	first := envelope{Frame: strings.Repeat("1", 64), Input: ask(obs, "agent"), Reach: heldReach(tools(), nil), Steps: []proc.Step{ret(st)}}
	if _, err := runRemember(store, first); err != nil || !store.Has(agentIn) {
		t.Fatal(err)
	}
	other := tools()
	delete(other, "echo") // another Boundary
	again := envelope{Frame: strings.Repeat("2", 64), Input: ask(obs, "agent"), Reach: heldReach(other, nil), Steps: []proc.Step{ret(st)}}
	out, err := runRemember(store, again)
	if err != nil || !strings.Contains(string(out), "one representation per contract") || strings.Contains(string(out), "export") {
		t.Fatalf("a second remember is refused before export: %s %v", out, err)
	}

	// A run of the procedure itself: the trace ends at the apprentice's
	// invoke of the contract it already implements.
	own := remember.Trace{Chain: st.Chain, Records: map[string]remember.Record{}, Opaque: st.Opaque}
	cause := st.Chain[len(st.Chain)-1]
	for k, r := range st.Records {
		if k == cause {
			r.To = "apprentice"
		}
		own.Records[k] = r
	}
	tl := tools()
	tl["apprentice"] = rememberContract()
	e := envelope{Frame: strings.Repeat("3", 64), Input: ask(obs, "apprentice"), Reach: heldReach(tl, map[string][]*proc.Schema{"apprentice": {rememberContract(), agentIn}}), Steps: []proc.Step{ret(own)}}
	out, err = runRemember(store, e)
	if err != nil || !strings.Contains(string(out), "one representation per contract") {
		t.Fatalf("remembering a procedure's own output is refused: %s %v", out, err)
	}
}

// A trace the kernel refused comes back as remember's refusal.
func TestRememberPassesOnATraceRefusal(t *testing.T) {
	e := envelope{Frame: strings.Repeat("4", 64), Input: ask(strings.Repeat("a", 64), "agent"), Reach: heldReach(tools(), nil),
		Steps: []proc.Step{{State: "fault", Fault: &proc.Fault{Kind: "boundary", Cause: "trace: no invoke of agent through that contract caused this observation"}}}}
	out, err := runRemember(proc.Store{Dir: t.TempDir()}, e)
	if err != nil || !strings.Contains(string(out), `"refused":"trace: no invoke`) {
		t.Fatalf("%s %v", out, err)
	}
}

var resume = &proc.Schema{Title: "resume", Type: "object", Required: []string{"resume"}, Fields: map[string]*proc.Schema{"resume": {Type: "any"}}}

// remembered stores the procedure the relay trace compiles to.
func remembered(t *testing.T, store proc.Store) {
	t.Helper()
	st, obs := fixture(t)
	if _, err := runRemember(store, envelope{Frame: strings.Repeat("9", 64), Input: ask(obs, "agent"), Reach: heldReach(tools(), nil), Steps: []proc.Step{ret(st)}}); err != nil {
		t.Fatal(err)
	}
}

// emitted reads an emission as its instructions.
func emitted(t *testing.T, out []byte) []struct {
	To      string          `json:"to"`
	Payload json.RawMessage `json:"payload"`
} {
	t.Helper()
	var ins []struct {
		To      string          `json:"to"`
		Payload json.RawMessage `json:"payload"`
	}
	if out[0] != '[' {
		out = append(append([]byte("["), out...), ']')
	}
	if err := json.Unmarshal(out, &ins); err != nil {
		t.Fatalf("%s: %v", out, err)
	}
	return ins
}

// Inside its refinement a remembered procedure runs without its model: the
// same base types, round after round, and no instruction to the agent.
func TestSameBaseTypesNoModel(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	remembered(t, store)
	e := envelope{Frame: strings.Repeat("5", 64), Input: json.RawMessage(`{"text":"relay:new"}`), Contract: agentIn,
		Reach: heldReach(tools(), map[string][]*proc.Schema{"agent": {agentIn, resume}})}
	for i := 0; i < 5; i++ {
		out, err := runProcedure(store, e, agentIn)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(out), `"to":"`+e.Frame+`"`) {
			if string(out) != `{"payload":{"again":"relay:new","echoes":["relay:new","relayed"]},"to":"`+e.Frame+`"}` {
				t.Fatalf("the procedure's return on new input: %s", out)
			}
			return
		}
		for _, in := range emitted(t, out) {
			if in.To != "echo" {
				t.Fatalf("no model runs inside the refinement: %s", out)
			}
			var p struct{ Text string }
			_ = json.Unmarshal(in.Payload, &p)
			v, _ := json.Marshal(p.Text)
			e.Steps = append(e.Steps, proc.Step{To: "echo", Payload: in.Payload, State: "return", Value: v})
		}
	}
	t.Fatal("never returned")
}

// A different base type at entry: the model is invoked from the start, with
// the command, and its answer is the procedure's.
func TestADifferentInputFallsBackFromTheStart(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	loose := &proc.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*proc.Schema{"text": {Type: "string"}, "note": {Type: "string"}}}
	in, _ := types.Decode([]byte(`{"text":"x"}`))
	src := proc.Source{Contract: loose, Original: "agent",
		Refinement: &proc.Refinement{Input: proc.InputType(loose, in), InputShown: proc.InputShown(loose, in)},
		Nodes:      map[string]proc.SourceNode{"a": {Call: json.RawMessage(`"echo"`), In: json.RawMessage(`{"$":["input"]}`)}},
		Return:     json.RawMessage(`{"$":["a"]}`)}
	p, nodes, err := proc.Compile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put(p, nodes); err != nil {
		t.Fatal(err)
	}
	e := envelope{Frame: strings.Repeat("6", 64), Input: json.RawMessage(`{"note":"n","text":"x"}`), Contract: loose, Reach: heldReach(tools(), nil)}
	out, err := runProcedure(store, e, loose)
	ins := emitted(t, out)
	if err != nil || len(ins) != 1 || ins[0].To != "agent" || string(ins[0].Payload) != `{"note":"n","text":"x"}` {
		t.Fatalf("the original actor is invoked with the command: %s %v", out, err)
	}
	e.Steps = []proc.Step{{To: "agent", State: "return", Value: json.RawMessage(`"the agent did it"`)}}
	if out, err = runProcedure(store, e, loose); err != nil || string(out) != `{"payload":"the agent did it","to":"`+e.Frame+`"}` {
		t.Fatalf("its answer is the procedure's: %s %v", out, err)
	}
	// The same input shape runs compiled.
	e = envelope{Frame: strings.Repeat("7", 64), Input: json.RawMessage(`{"text":"y"}`), Contract: loose, Reach: heldReach(tools(), nil)}
	if out, _ := runProcedure(store, e, loose); emitted(t, out)[0].To != "echo" {
		t.Fatalf("inside the refinement it runs compiled: %s", out)
	}
}

// A different base type at an edge: the model resumes there, through its
// resume contract, with the command, what already ran, and which edge
// refused and why, as names and presentations. Without one, it is given the
// command and starts over.
func TestADifferentEdgeResumesTheModelThere(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	remembered(t, store)
	e := envelope{Frame: strings.Repeat("8", 64), Input: json.RawMessage(`{"text":"relay:new"}`), Contract: agentIn,
		Reach: heldReach(tools(), map[string][]*proc.Schema{"agent": {agentIn, resume}})}
	out, _ := runProcedure(store, e, agentIn)
	ins := emitted(t, out)
	// echo answers the first call with a number: not the trace's base type.
	e.Steps = []proc.Step{
		{To: "echo", Payload: ins[0].Payload, State: "return", Value: json.RawMessage(`42`)},
		{To: "echo", Payload: ins[1].Payload, State: "return", Value: json.RawMessage(`"relayed"`)},
	}
	out, err := runProcedure(store, e, agentIn)
	fell := emitted(t, out)
	if err != nil || len(fell) != 1 || fell[0].To != "agent" {
		t.Fatalf("the original actor is invoked: %s %v", out, err)
	}
	var r struct {
		Resume struct {
			Command json.RawMessage
			Ran     []struct {
				To, Type string
				Value    json.RawMessage
			}
			Refused struct{ At, Why string }
		}
	}
	if err := json.Unmarshal(fell[0].Payload, &r); err != nil {
		t.Fatal(err)
	}
	if string(r.Resume.Command) != `{"text":"relay:new"}` || len(r.Resume.Ran) != 2 || r.Resume.Ran[0].Type != "echo → number" ||
		r.Resume.Refused.At != "echo" || !strings.Contains(r.Resume.Refused.Why, "echo returned echo → number; this procedure was compiled for echo → string") {
		t.Fatalf("it is shown the command, what ran and why the edge refused: %s", fell[0].Payload)
	}
	if strings.Contains(string(fell[0].Payload), e.Frame) || strings.Contains(string(fell[0].Payload), "\"hash\"") {
		t.Fatal("no hash reaches the model")
	}

	// Without a resume contract, the actor is given the command.
	e.Frame, e.Reach = strings.Repeat("c", 64), heldReach(tools(), nil)
	e.Steps = nil
	out, _ = runProcedure(store, e, agentIn)
	ins = emitted(t, out)
	e.Steps = []proc.Step{
		{To: "echo", Payload: ins[0].Payload, State: "return", Value: json.RawMessage(`42`)},
		{To: "echo", Payload: ins[1].Payload, State: "return", Value: json.RawMessage(`"relayed"`)},
	}
	out, _ = runProcedure(store, e, agentIn)
	if fell := emitted(t, out); fell[0].To != "agent" || string(fell[0].Payload) != `{"text":"relay:new"}` {
		t.Fatalf("without resume, the command: %s", out)
	}
}

// A call that faulted is an edge that refused, too.
func TestAFaultFallsBack(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	remembered(t, store)
	e := envelope{Frame: strings.Repeat("d", 64), Input: json.RawMessage(`{"text":"relay:new"}`), Contract: agentIn, Reach: heldReach(tools(), nil)}
	out, _ := runProcedure(store, e, agentIn)
	ins := emitted(t, out)
	fault := &proc.Fault{Kind: "boundary", Cause: `"echo" is not in reach`}
	e.Steps = []proc.Step{{To: "", State: "fault", Fault: fault, Payload: ins[0].Payload}, {State: "fault", Fault: fault, Payload: ins[1].Payload}}
	out, err := runProcedure(store, e, agentIn)
	if fell := emitted(t, out); err != nil || fell[0].To != "agent" {
		t.Fatalf("a faulted call falls back to the original actor: %s %v", out, err)
	}
}

var refineContract = &proc.Schema{Title: "refine", Type: "object", Required: []string{"refine"}, Fields: map[string]*proc.Schema{"refine": {Type: "any"}}}

// At remember the model that was there writes the refinements (H9): it is
// shown its command, rounds, result, base types and constants; an answer that
// fails a check is refused and shown to it, and it tries again; a valid one
// replaces the base types, before the apprentice extends itself.
func TestTheModelWritesTheRefinements(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	st, obs := fixture(t)
	contracts := map[string][]*proc.Schema{"agent": {agentIn, refineContract}}
	e := envelope{Frame: strings.Repeat("e", 64), Input: ask(obs, "agent"), Reach: heldReach(tools(), contracts), Steps: []proc.Step{ret(st)}}
	out, err := runRemember(store, e)
	ins := emitted(t, out)
	if err != nil || ins[0].To != "agent" || !strings.Contains(string(ins[0].Payload), `"refine"`) {
		t.Fatalf("after the compile, the agent is asked for its refinements: %s %v", out, err)
	}
	var shown struct {
		Refine struct {
			Rounds    [][]struct{ Node, To, Type string }
			Constants []any
			InputType string `json:"input_type"`
			Refused   string
		}
	}
	_ = json.Unmarshal(ins[0].Payload, &shown)
	if len(shown.Refine.Rounds) != 2 || shown.Refine.Rounds[0][0].Node != "r0c0" || shown.Refine.InputType != "input → {text: string}" ||
		len(shown.Refine.Constants) != 1 || shown.Refine.Refused != "" || strings.Contains(string(ins[0].Payload), obs) {
		t.Fatalf("it is shown names and presentations, never a hash: %s", ins[0].Payload)
	}
	for _, bad := range []struct{ answer, why string }{
		{`{"header":"input.txt.startsWith(\"relay:\")"}`, "error on its own trace"},
		{`{"header":"input.text.startsWith(\"other:\")"}`, "refuses the trace it was compiled from"},
	} {
		e.Steps = append(e.Steps, ret(json.RawMessage(bad.answer)))
		out, _ = runRemember(store, e)
		_ = json.Unmarshal(emitted(t, out)[0].Payload, &shown)
		if !strings.Contains(shown.Refine.Refused, bad.why) {
			t.Fatalf("%s: refused and shown: %s", bad.answer, out)
		}
	}
	e.Steps = append(e.Steps, ret(json.RawMessage(`{"header":"input.text.startsWith(\"relay:\")","edges":{"r0c1":"returns.r0c1 == \"relayed\""}}`)))
	out, _ = runRemember(store, e)
	if !strings.Contains(string(out), `"export"`) {
		t.Fatalf("a valid answer is kept, then the apprentice extends itself: %s", out)
	}
	p, _, _ := store.Procedure(agentIn)
	if ref := p.Header.Refinement; ref.Header != `input.text.startsWith("relay:")` || len(ref.On) != 1 || ref.Input == "" {
		t.Fatalf("the refinement replaces the base types where present: %+v", ref)
	}

	// Inside the refinement it runs compiled; outside it, the agent.
	run := envelope{Frame: strings.Repeat("f", 64), Input: json.RawMessage(`{"text":"relay:new"}`), Contract: agentIn, Reach: heldReach(tools(), contracts)}
	if out, _ := runProcedure(store, run, agentIn); emitted(t, out)[0].To != "echo" {
		t.Fatalf("inside its refinement, compiled: %s", out)
	}
	run.Frame, run.Input = strings.Repeat("0", 64), json.RawMessage(`{"text":"plain"}`)
	out, _ = runProcedure(store, run, agentIn)
	if fell := emitted(t, out); fell[0].To != "agent" || string(fell[0].Payload) != `{"text":"plain"}` {
		t.Fatalf("outside it, the agent from the start: %s", out)
	}
	// At an edge: the second call returned something else.
	run.Frame, run.Input, run.Steps = strings.Repeat("1", 64), json.RawMessage(`{"text":"relay:new"}`), nil
	out, _ = runProcedure(store, run, agentIn)
	ins = emitted(t, out)
	run.Steps = []proc.Step{
		{To: "echo", Payload: ins[0].Payload, State: "return", Value: json.RawMessage(`"relay:new"`)},
		{To: "echo", Payload: ins[1].Payload, State: "return", Value: json.RawMessage(`"diverted"`)},
	}
	out, _ = runProcedure(store, run, agentIn)
	if fell := emitted(t, out); fell[0].To != "agent" {
		t.Fatalf("an edge refinement that is false falls back: %s", out)
	}
}

// An answer that never passes leaves the base types after MaxAttempts.
func TestAModelThatCannotRefineKeepsTheBaseTypes(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	st, obs := fixture(t)
	e := envelope{Frame: strings.Repeat("e", 64), Input: ask(obs, "agent"), Reach: heldReach(tools(), map[string][]*proc.Schema{"agent": {agentIn, refineContract}}), Steps: []proc.Step{ret(st)}}
	runRemember(store, e)
	var out []byte
	for i := 0; i < MaxAttempts; i++ {
		e.Steps = append(e.Steps, ret(json.RawMessage(`{"header":"size(input.text)"}`)))
		out, _ = runRemember(store, e)
	}
	if !strings.Contains(string(out), `"export"`) {
		t.Fatalf("after %d refusals, on to export: %s", MaxAttempts, out)
	}
	if p, _, _ := store.Procedure(agentIn); p.Header.Refinement.Header != "" || p.Header.Refinement.Input == "" {
		t.Fatalf("with its base types: %+v", p.Header.Refinement)
	}
}
