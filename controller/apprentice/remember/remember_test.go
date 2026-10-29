package remember

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"skyra/apprentice/proc"
)

var agentIn = &proc.Schema{Type: "object", Required: []string{"text"}, Fields: map[string]*proc.Schema{"text": {Type: "string"}}}

// relay is a real trace: the scripted agent relaying "relay:hello", as the
// kernel's trace syscall handed it back in run.sh.
func relay(t *testing.T) (Trace, string) {
	t.Helper()
	b, err := os.ReadFile("testdata/relay.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Observation string `json:"observation"`
		Trace       Trace  `json:"trace"`
	}
	if err := json.Unmarshal(b, &f); err != nil {
		t.Fatal(err)
	}
	return f.Trace, f.Observation
}

func reach() Reach {
	return Reach{"agent": {agentIn}, "echo": {agentIn}, "apprentice": {{Title: "remember", Type: "object"}}}
}

func node(t *testing.T, src proc.Source, name string) string {
	t.Helper()
	n, ok := src.Nodes[name]
	if !ok {
		t.Fatalf("no node %s in %v", name, src.Order)
	}
	b, _ := json.Marshal(n)
	return string(b)
}

// A trace from the scripted agent compiles to the nodes it ran: one per
// call, by round, every round after the one before; bindings found by
// provenance; constants kept; the return made of bindings only.
func TestTheAgentsTraceCompiles(t *testing.T) {
	st, obs := relay(t)
	srcs, err := Compile(st, obs, "agent", agentIn, reach(), "apprentice")
	if err != nil {
		t.Fatal(err)
	}
	if len(srcs) != 1 {
		t.Fatalf("one continuation, one procedure: %d", len(srcs))
	}
	src, ev := srcs[0].Source, srcs[0].Evidence
	if len(ev.Rounds) != 2 || len(ev.Rounds[0]) != 2 || ev.Rounds[0][1].Type != "echo → string" || ev.InputType != "input → {text: string}" {
		t.Fatalf("the evidence is the trace as names and presentations: %+v", ev)
	}
	if len(ev.Constants) != 1 || ev.Constants[0] != "relayed" {
		t.Fatalf("the constants the procedure keeps: %v", ev.Constants)
	}
	if src.Original != "agent" || !proc.Same(src.Contract, agentIn) || strings.Join(src.Order, " ") != "r0c0 r0c1 r1c0" {
		t.Fatalf("the agent's contract, remembered from the agent, three calls: %+v", src)
	}
	for name, want := range map[string]string{
		"r0c0": `{"call":"echo","in":{"$":["input"]}}`,                         // the whole input: a parameter
		"r0c1": `{"call":"echo","in":{"text":"relayed"}}`,                      // a constant, kept as recorded
		"r1c0": `{"call":"echo","in":{"$":["input"]},"after":["r0c0","r0c1"]}`, // the next round, after the one before
	} {
		if got := node(t, src, name); got != want {
			t.Errorf("%s:\n got %s\nwant %s", name, got, want)
		}
	}
	if string(src.Return) != `{"again":{"$":["input","text"]},"echoes":[{"$":["input","text"]},{"$":["r0c1"]}]}` {
		t.Fatalf("the return passes up what ran: %s", src.Return)
	}
	if src.Refinement == nil || src.Refinement.Input == "" || len(src.Refinement.Edges) != 3 {
		t.Fatalf("the base types are the default refinement, at entry and every edge: %+v", src.Refinement)
	}
	// It compiles into nodes the apprentice runs.
	if _, _, err := proc.Compile(src); err != nil {
		t.Fatal(err)
	}
}

// What remember refuses: an observation that is not the contract's return, a
// target that did not cause it, and a frame whose return the model authored.
func TestRememberRefuses(t *testing.T) {
	st, obs := relay(t)
	var refused Refused
	if _, err := Compile(st, st.Chain[2], "agent", agentIn, reach(), "apprentice"); !asRefused(err, &refused) || !strings.Contains(string(refused), "not its return") {
		t.Errorf("an echo's return is not the agent's: %v", err)
	}
	if _, err := Compile(st, obs, "echo", agentIn, reach(), "apprentice"); !asRefused(err, &refused) {
		t.Errorf("the trace ends at the agent, not echo: %v", err)
	}
	authored := st
	authored.Records = map[string]Record{}
	for k, r := range st.Records {
		if k == obs {
			r.Payload = json.RawMessage(`"echo said relay:hello"`)
		}
		authored.Records[k] = r
	}
	if _, err := Compile(authored, obs, "agent", agentIn, reach(), "apprentice"); !asRefused(err, &refused) || !strings.Contains(string(refused), "authored") {
		t.Errorf("an authored result is content, not control: %v", err)
	}
}

func asRefused(err error, r *Refused) bool {
	x, ok := err.(Refused)
	if ok {
		*r = x
	}
	return ok
}
