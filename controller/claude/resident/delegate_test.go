package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDelegationIgnoresAndPreservesSavedParentSession(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("SKYRA_STATE", dir)
	file := filepath.Join(dir, "session")
	if err := os.WriteFile(file, []byte("parent-session"), 0600); err != nil {
		t.Fatal(err)
	}
	c := continuation{Frame: "f1"}
	c.Input.Fresh = true
	c.Input.Text = "goal"
	c.Input.Session = "must-not-resume"
	empty := []string{}
	c.WorkTools = &empty
	config := turnConfig(c)
	if _, ok := config["session"]; ok {
		t.Fatal("delegated child resumed parent session")
	}
	b, _ := json.Marshal(config)
	if !strings.Contains(string(b), `"tools":[]`) {
		t.Fatalf("reasoning-only config: %s", b)
	}
	c.Steps = []step{{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc"}`)},
		{To: "claude.exchange", State: "return", Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"goal\"}"]}`), Value: json.RawMessage(`{"lines":["{\"type\":\"result\",\"subtype\":\"success\",\"session_id\":\"child-session\",\"result\":\"findings\"}"]}`)}}
	if _, err := decide(c); err != nil {
		t.Fatal(err)
	}
	if saved, _ := os.ReadFile(file); string(saved) != "parent-session" {
		t.Fatalf("child replaced parent session: %s", saved)
	}
}

func TestClaudeDispatchesDelegateThroughInterpreter(t *testing.T) {
	remember := &schema{Title: "remember", Type: "object", Required: []string{"remember"}, Fields: map[string]*schema{"remember": {Type: "object"}}}
	delegate := &schema{Title: "delegate", Type: "object", Required: []string{"delegate"}, Fields: map[string]*schema{"delegate": {Type: "object", Required: []string{"goal"}, Fields: map[string]*schema{"goal": {Type: "string"}}}}}
	c := continuation{Frame: "f1", Tools: map[string]*schema{"interpreter": remember}, Contracts: map[string][]*schema{"interpreter": {remember, delegate}}}
	c.Input.Text = "goal"
	names := []string{"delegate"}
	c.WorkTools = &names
	ts := tools(c)
	if len(ts) != 1 || ts[0].Name != "delegate" || ts[0].Target != "interpreter" {
		t.Fatalf("procedure catalog: %+v", ts)
	}
	c.Steps = []step{{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc"}`)},
		{To: "claude.exchange", State: "return", Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"goal\"}"]}`), Value: json.RawMessage(`{"lines":["{\"_bridge\":\"call\",\"id\":\"call\",\"name\":\"delegate\",\"args\":{\"delegate\":{\"goal\":\"inspect\"}}}"]}`)}}
	out, err := decide(c)
	if err != nil || !strings.Contains(out, `"to":"interpreter"`) || !strings.Contains(out, `"goal":"inspect"`) {
		t.Fatalf("procedure dispatch: %s %v", out, err)
	}
	c.Steps = append(c.Steps, step{To: "interpreter", State: "return", Value: json.RawMessage(`{"failed":{"kind":"child","cause":"provider exited"}}`)})
	out, err = decide(c)
	if err != nil || !strings.Contains(out, `\"_bridge\":\"error\"`) || !strings.Contains(out, "provider exited") {
		t.Fatalf("failure delivery: %s %v", out, err)
	}
}

func TestClaudeCanonicalWorkToolHasAValidTransportLabel(t *testing.T) {
	names := []string{"repo_write"}
	c := continuation{WorkTools: &names, Tools: map[string]*schema{"repo.write": {Type: "object"}}}
	ts := tools(c)
	if len(ts) != 1 || ts[0].Name != "repo_write" || ts[0].Target != "repo.write" {
		t.Fatalf("transport name changed ownership: %+v", ts)
	}
}
