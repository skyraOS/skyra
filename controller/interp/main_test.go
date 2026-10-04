package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"skyra/interp/proc"
	"skyra/interp/remember"
	"skyra/interp/types"
)

var agentIn = &proc.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*proc.Schema{"text": {Type: "string"}}}

func fixture(t *testing.T) (remember.Strand, string) {
	t.Helper()
	b, err := os.ReadFile("remember/testdata/relay.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Observation string          `json:"observation"`
		Strand      remember.Strand `json:"strand"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f.Strand, f.Observation
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
	return map[string]*proc.Schema{"agent": agentIn, "echo": agentIn, "interpreter": rememberContract()}
}

// remember's body: the strand, the compile, the extension, then what was
// remembered. The first step is a syscall; the interpreter extends itself.
func TestRememberRunsItsBody(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	st, obs := fixture(t)
	e := envelope{Frame: strings.Repeat("1", 64), Input: ask(obs, "agent"), Tools: tools()}
	out, err := runRemember(store, e)
	if err != nil || !strings.Contains(string(out), `"to":"system"`) || !strings.Contains(string(out), `"strand"`) {
		t.Fatalf("remember's first step is the strand syscall: %s %v", out, err)
	}
	e.Steps = []proc.Step{ret(st)}
	out, err = runRemember(store, e)
	if err != nil || !strings.Contains(string(out), `"extend-reach"`) || !store.Has(agentIn) {
		t.Fatalf("then it stores the procedure and extends itself: %s %v", out, err)
	}
	e.Steps = append(e.Steps, ret(map[string]bool{"ok": true}))
	out, err = runRemember(store, e)
	if err != nil || string(out) != `{"payload":{"remembered":["agent"]},"to":"`+e.Frame+`"}` {
		t.Fatalf("then it returns what it remembered: %s %v", out, err)
	}
}

// A second remember through the same contract is refused, from the same
// Horizon and from another (the store is the interpreter's, not a Horizon's),
// before anything is stored or extended. So is remembering a run of the
// procedure itself: a second representation of the same contract.
func TestRememberRefusesASecondRepresentation(t *testing.T) {
	store := proc.Store{Dir: t.TempDir()}
	st, obs := fixture(t)
	first := envelope{Frame: strings.Repeat("1", 64), Input: ask(obs, "agent"), Tools: tools(), Steps: []proc.Step{ret(st)}}
	if _, err := runRemember(store, first); err != nil || !store.Has(agentIn) {
		t.Fatal(err)
	}
	other := tools()
	delete(other, "echo") // another Horizon
	again := envelope{Frame: strings.Repeat("2", 64), Input: ask(obs, "agent"), Tools: other, Steps: []proc.Step{ret(st)}}
	out, err := runRemember(store, again)
	if err != nil || !strings.Contains(string(out), "one representation per contract") || strings.Contains(string(out), "extend-reach") {
		t.Fatalf("a second remember is refused before extend-reach: %s %v", out, err)
	}

	// A run of the procedure itself: the strand ends at the interpreter's
	// invoke of the contract it already implements.
	own := remember.Strand{Chain: st.Chain, Records: map[string]remember.Record{}, Opaque: st.Opaque}
	cause := st.Chain[len(st.Chain)-1]
	for k, r := range st.Records {
		if k == cause {
			r.To = "interpreter"
		}
		own.Records[k] = r
	}
	tl := tools()
	tl["interpreter"] = rememberContract()
	e := envelope{Frame: strings.Repeat("3", 64), Input: ask(obs, "interpreter"), Tools: tl,
		Contracts: map[string][]*proc.Schema{"interpreter": {rememberContract(), agentIn}}, Steps: []proc.Step{ret(own)}}
	out, err = runRemember(store, e)
	if err != nil || !strings.Contains(string(out), "one representation per contract") {
		t.Fatalf("remembering a procedure's own output is refused: %s %v", out, err)
	}
}

// A strand the kernel refused comes back as remember's refusal.
func TestRememberPassesOnAStrandRefusal(t *testing.T) {
	e := envelope{Frame: strings.Repeat("4", 64), Input: ask(strings.Repeat("a", 64), "agent"), Tools: tools(),
		Steps: []proc.Step{{State: "fault", Fault: &proc.Fault{Kind: "horizon", Cause: "strand: no invoke of agent through that contract caused this observation"}}}}
	out, err := runRemember(proc.Store{Dir: t.TempDir()}, e)
	if err != nil || !strings.Contains(string(out), `"refused":"strand: no invoke`) {
		t.Fatalf("%s %v", out, err)
	}
}

var resume = &proc.Schema{Title: "resume", Type: "object", Required: []string{"resume"}, Fields: map[string]*proc.Schema{"resume": {Type: "any"}}}

// remembered stores the procedure the relay strand compiles to.
func remembered(t *testing.T, store proc.Store) {
	t.Helper()
	st, obs := fixture(t)
	if _, err := runRemember(store, envelope{Frame: strings.Repeat("9", 64), Input: ask(obs, "agent"), Tools: tools(), Steps: []proc.Step{ret(st)}}); err != nil {
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
		Tools: tools(), Contracts: map[string][]*proc.Schema{"agent": {agentIn, resume}}}
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
	e := envelope{Frame: strings.Repeat("6", 64), Input: json.RawMessage(`{"note":"n","text":"x"}`), Contract: loose, Tools: tools()}
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
	e = envelope{Frame: strings.Repeat("7", 64), Input: json.RawMessage(`{"text":"y"}`), Contract: loose, Tools: tools()}
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
		Tools: tools(), Contracts: map[string][]*proc.Schema{"agent": {agentIn, resume}}}
	out, _ := runProcedure(store, e, agentIn)
	ins := emitted(t, out)
	// echo answers the first call with a number: not the strand's base type.
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
	e.Frame, e.Contracts = strings.Repeat("c", 64), nil
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
	e := envelope{Frame: strings.Repeat("d", 64), Input: json.RawMessage(`{"text":"relay:new"}`), Contract: agentIn, Tools: tools()}
	out, _ := runProcedure(store, e, agentIn)
	ins := emitted(t, out)
	fault := &proc.Fault{Kind: "horizon", Cause: `"echo" is not in reach`}
	e.Steps = []proc.Step{{To: "", State: "fault", Fault: fault, Payload: ins[0].Payload}, {State: "fault", Fault: fault, Payload: ins[1].Payload}}
	out, err := runProcedure(store, e, agentIn)
	if fell := emitted(t, out); err != nil || fell[0].To != "agent" {
		t.Fatalf("a faulted call falls back to the original actor: %s %v", out, err)
	}
}
