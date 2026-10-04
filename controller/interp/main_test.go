package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"skyra/interp/proc"
	"skyra/interp/remember"
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
