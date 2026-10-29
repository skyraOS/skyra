package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func parse(t *testing.T, out string) (string, json.RawMessage) {
	t.Helper()
	var emission struct {
		To      string          `json:"to"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal([]byte(out), &emission); err != nil {
		t.Fatalf("parse emission: %v: %s", err, out)
	}
	return emission.To, emission.Payload
}

func TestOpenWhenNoIncarnation(t *testing.T) {
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, _ := parse(t, out)
	if to != "claude.open" {
		t.Fatalf("expected claude.open, got %s", to)
	}
}

func TestDeliverConfigAfterOpen(t *testing.T) {
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Reach = reachOf(map[string]*schema{
		"echo": {Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}},
	}, nil)
	c.Steps = []step{{
		To: "claude.open", State: "return",
		Value: json.RawMessage(`{"incarnation":"inc-1"}`),
	}}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, payload := parse(t, out)
	if to != "claude.exchange" {
		t.Fatalf("expected claude.exchange, got %s", to)
	}
	var ex struct {
		Deliver []string `json:"deliver"`
	}
	json.Unmarshal(payload, &ex)
	if len(ex.Deliver) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(ex.Deliver))
	}
	var config struct {
		Prompt string `json:"prompt"`
		Frame  string `json:"frame"`
	}
	json.Unmarshal([]byte(ex.Deliver[0]), &config)
	if config.Prompt != "hello" || config.Frame != "f1" {
		t.Fatalf("config: %+v", config)
	}
}

func TestReadMoreOnStreamEvents(t *testing.T) {
	streamLine := `{"type":"system","subtype":"init","session_id":"s1"}`
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Reach = reachOf(map[string]*schema{"echo": {Type: "object"}}, nil)
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   json.RawMessage(`{"lines":["` + strings.ReplaceAll(streamLine, `"`, `\"`) + `"],"partial":"","delivery":"written","peer_eof":false,"error":""}`)},
	}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, _ := parse(t, out)
	if to != "claude.exchange" {
		t.Fatalf("expected claude.exchange, got %s", to)
	}
}

func TestCompleteTurn(t *testing.T) {
	resultLine := `{"type":"result","subtype":"success","session_id":"s1","result":"done","message":{"content":[{"type":"text","text":"I helped you."}]}}`
	obs, _ := json.Marshal(map[string]any{"lines": []string{resultLine}, "partial": "", "delivery": "", "peer_eof": false, "error": ""})
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   obs},
	}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, payload := parse(t, out)
	if to != "f1" {
		t.Fatalf("expected frame f1, got %s", to)
	}
	var result struct {
		Text    string `json:"text"`
		Session string `json:"session"`
	}
	json.Unmarshal(payload, &result)
	if result.Text != "I helped you." {
		t.Fatalf("text = %q", result.Text)
	}
	if result.Session != "s1" {
		t.Fatalf("session = %q", result.Session)
	}
}

func TestBridgeCallDispatch(t *testing.T) {
	bridgeLine := `{"_bridge":"call","id":"f1:mcp:1","name":"echo","args":{"text":"hi"}}`
	obs, _ := json.Marshal(map[string]any{"lines": []string{bridgeLine}, "partial": "", "delivery": "", "peer_eof": false, "error": ""})
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Reach = reachOf(map[string]*schema{"echo": {Type: "object"}}, nil)
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   json.RawMessage(`{"lines":[],"partial":"","delivery":"written","peer_eof":false,"error":""}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":[]}`),
			Value:   obs},
	}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, payload := parse(t, out)
	if to != "echo" {
		t.Fatalf("expected echo, got %s", to)
	}
	if !strings.Contains(string(payload), `"text"`) {
		t.Fatalf("payload: %s", payload)
	}
}

func TestBridgeResultDelivery(t *testing.T) {
	bridgeLine := `{"_bridge":"call","id":"f1:mcp:1","name":"echo","args":{"text":"hi"}}`
	obs, _ := json.Marshal(map[string]any{"lines": []string{bridgeLine}, "partial": "", "peer_eof": false, "error": ""})
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Reach = reachOf(map[string]*schema{"echo": {Type: "object"}}, nil)
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   json.RawMessage(`{"lines":[],"partial":"","delivery":"written","peer_eof":false,"error":""}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":[]}`),
			Value:   obs},
		{To: "echo", State: "return", Value: json.RawMessage(`{"echoed":"hi"}`)},
	}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, payload := parse(t, out)
	if to != "claude.exchange" {
		t.Fatalf("expected claude.exchange, got %s", to)
	}
	var ex struct {
		Deliver []string `json:"deliver"`
	}
	json.Unmarshal(payload, &ex)
	if len(ex.Deliver) != 1 {
		t.Fatalf("expected 1 delivery, got %d", len(ex.Deliver))
	}
	var br struct {
		Bridge string          `json:"_bridge"`
		ID     string          `json:"id"`
		Value  json.RawMessage `json:"value"`
	}
	json.Unmarshal([]byte(ex.Deliver[0]), &br)
	if br.Bridge != "result" || br.ID != "f1:mcp:1" {
		t.Fatalf("bridge result: %+v", br)
	}
}

func TestFailedTurn(t *testing.T) {
	resultLine := `{"type":"result","subtype":"error_during_execution","is_error":true,"session_id":"s1"}`
	obs, _ := json.Marshal(map[string]any{"lines": []string{resultLine}, "partial": "", "peer_eof": false, "error": ""})
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   obs},
	}
	_, err := decide(c)
	if err == nil {
		t.Fatal("expected error for failed turn")
	}
	if !strings.Contains(err.Error(), "error_during_execution") {
		t.Fatalf("error: %v", err)
	}
}

func TestPeerEOFBeforeResult(t *testing.T) {
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   json.RawMessage(`{"lines":[],"partial":"","delivery":"written","peer_eof":true,"error":""}`)},
	}
	_, err := decide(c)
	if err == nil || !strings.Contains(err.Error(), "exited before completing") {
		t.Fatalf("expected exit error, got %v", err)
	}
}

func TestToolFaultBecomesError(t *testing.T) {
	bridgeLine := `{"_bridge":"call","id":"f1:mcp:1","name":"echo","args":{}}`
	obs, _ := json.Marshal(map[string]any{"lines": []string{bridgeLine}, "partial": "", "peer_eof": false, "error": ""})
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Reach = reachOf(map[string]*schema{"echo": {Type: "object"}}, nil)
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":["{\"prompt\":\"hello\"}"]}`),
			Value:   json.RawMessage(`{"lines":[],"partial":"","delivery":"written","peer_eof":false,"error":""}`)},
		{To: "claude.exchange", State: "return",
			Payload: json.RawMessage(`{"deliver":[]}`),
			Value:   obs},
		{To: "echo", State: "fault", Fault: &struct{ Kind, Cause string }{"denied", "not in reach"},
			Value: json.RawMessage(`null`)},
	}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	to, payload := parse(t, out)
	if to != "claude.exchange" {
		t.Fatalf("expected claude.exchange, got %s", to)
	}
	var ex struct {
		Deliver []string `json:"deliver"`
	}
	json.Unmarshal(payload, &ex)
	if len(ex.Deliver) != 1 || !strings.Contains(ex.Deliver[0], `"_bridge":"error"`) {
		t.Fatalf("expected bridge error delivery: %v", ex.Deliver)
	}
}

func TestOperationLimit(t *testing.T) {
	c := continuation{Frame: "f1"}
	c.Steps = make([]step, 2001)
	_, err := decide(c)
	if err == nil || !strings.Contains(err.Error(), "operation limit") {
		t.Fatalf("expected operation limit, got %v", err)
	}
}

func TestTurnConfigIncludesSession(t *testing.T) {
	c := continuation{Frame: "f1"}
	c.Input.Text = "hello"
	c.Input.Session = "prev-session"
	c.Reach = reachOf(map[string]*schema{"echo": {Type: "object"}}, nil)
	c.Steps = []step{
		{To: "claude.open", State: "return", Value: json.RawMessage(`{"incarnation":"inc-1"}`)},
	}
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	_, payload := parse(t, out)
	var ex struct {
		Deliver []string `json:"deliver"`
	}
	json.Unmarshal(payload, &ex)
	if !strings.Contains(ex.Deliver[0], `"session":"prev-session"`) {
		t.Fatalf("expected session in config: %s", ex.Deliver[0])
	}
}

func TestPromptUsesCurrentContinuationPaths(t *testing.T) {
	c := continuation{Workspace: "/project", Scratch: "/notes"}
	cfg := turnConfig(c)
	text := cfg["system_prompt"].(string)
	if !strings.Contains(text, "/project") || !strings.Contains(text, "/notes") {
		t.Fatal(text)
	}
}
