package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func value(v any) json.RawMessage { b, _ := json.Marshal(v); return b }
func emitted(t *testing.T, c continuation) (string, map[string]json.RawMessage) {
	t.Helper()
	out, err := decide(c)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err = json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatal(err)
	}
	var to string
	if err = json.Unmarshal(m["to"], &to); err != nil {
		t.Fatal(err)
	}
	return to, m
}

// An idle server is reclaimed by its holder's backoff (proposal 38); the
// adapter finds out when its inspect sees the server gone, and starts another.
func TestExistingThreadInspectsBeforeTurnAndRestartsWhenGone(t *testing.T) {
	c := continuation{Frame: "frame"}
	c.Input.Text = "next question"
	c.Input.Thread = "thread-1"
	c.Input.Incarnation = "old-process"
	c.Input.ActivityAt = 1000
	to, m := emitted(t, c)
	if to != "codex.inspect" {
		t.Fatalf("first operation %s", to)
	}
	var payload map[string]any
	json.Unmarshal(m["payload"], &payload)
	if payload["op"] != "inspect" {
		t.Fatal(payload)
	}
	c.Steps = append(c.Steps, step{To: to, Payload: m["payload"], State: "return", Value: value(map[string]any{
		"lines": []string{`{"id":"frame:inspect","result":{"data":[]}}` + "\n"}, "delivery": "written", "read": "observed", "observed_at": 1200, "peer_eof": false, "error": "", "partial": "",
	})})
	to, _ = emitted(t, c)
	if to != "codex.exchange" {
		t.Fatalf("recent thread should start a turn, got %s", to)
	}
	c.Steps[0].Value = value(map[string]any{
		"lines": []string{}, "delivery": "not_sent", "read": "not_started", "observed_at": 1801000, "peer_eof": true, "error": "", "partial": "",
	})
	to, m = emitted(t, c)
	if to != "codex.stop" {
		t.Fatalf("a reclaimed server should be stopped and replaced, got %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, Payload: m["payload"], State: "return", Value: value(map[string]any{
		"lines": []string{}, "delivery": "shutdown", "read": "observed", "observed_at": 1801001, "peer_eof": true, "error": "", "partial": "",
	})})
	to, _ = emitted(t, c)
	if to != "codex.open" {
		t.Fatalf("after stop should open a fresh session, got %s", to)
	}
}

func TestOnlyModelFacingToolsAreOfferedToCodex(t *testing.T) {
	c := continuation{Frame: "frame", Tools: map[string]*schema{
		"echo":       {Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}},
		"codex.stop": {Type: "object"}, "codex.exchange": {Type: "object"},
	}}
	c.Input.Text = "hello"
	c.Steps = []step{
		{To: "codex.open", State: "return", Value: value(map[string]any{"incarnation": "actor-1", "pid": 1, "profile": "hash"})},
		{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{rpc("frame:initialize", "initialize", map[string]any{})}}),
			Value: value(map[string]any{"lines": []string{`{"id":"frame:initialize","result":{"userAgent":"fake"}}` + "\n"}, "delivery": "written", "read": "observed", "observed_at": 1, "peer_eof": false, "error": "", "partial": ""})},
	}
	to, m := emitted(t, c)
	if to != "codex.exchange" {
		t.Fatalf("expected thread setup, got %s", to)
	}
	var q struct {
		Deliver []string `json:"deliver"`
	}
	if err := json.Unmarshal(m["payload"], &q); err != nil {
		t.Fatal(err)
	}
	if len(q.Deliver) != 2 || !strings.Contains(q.Deliver[1], `"name":"echo"`) || strings.Contains(q.Deliver[1], `"name":"codex.stop"`) || strings.Contains(q.Deliver[1], `"name":"codex.exchange"`) {
		t.Fatalf("offered tools: %v", q.Deliver)
	}
}

// turnWithCalls is a continuation whose turn has started and whose Codex has
// asked for each tool call given, in order.
func turnWithCalls(calls ...string) continuation {
	c := continuation{Frame: "frame", Tools: map[string]*schema{
		"echo": {Type: "object", Fields: map[string]*schema{"text": {Type: "string"}}, Required: []string{"text"}},
		"read": {Type: "object", Fields: map[string]*schema{"path": {Type: "string"}}, Required: []string{"path"}},
	}}
	c.Input.Text = "hello"
	observed := func(deliver []string, lines ...string) step {
		return step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": deliver}),
			Value: value(map[string]any{"lines": lines, "delivery": "written", "read": "observed", "observed_at": 1, "peer_eof": false, "error": "", "partial": ""})}
	}
	turn := []string{`{"id":"frame:turn","result":{"turn":{"id":"turn-1"}}}`}
	for i, call := range calls {
		tool, args, _ := strings.Cut(call, " ")
		turn = append(turn, fmt.Sprintf(`{"id":"call-%d","method":"item/tool/call","params":{"threadId":"thread-1","turnId":"turn-1","tool":%q,"arguments":%s}}`, i, tool, args))
	}
	c.Steps = []step{
		{To: "codex.open", State: "return", Value: value(map[string]any{"incarnation": "actor-1"})},
		observed([]string{rpc("frame:initialize", "initialize", map[string]any{})}, `{"id":"frame:initialize","result":{}}`),
		observed([]string{`{"method":"initialized"}`, rpc("frame:thread", "thread/start", map[string]any{})},
			`{"id":"frame:thread","result":{"thread":{"id":"thread-1"},"approvalPolicy":"never","sandbox":{"type":"readOnly","networkAccess":false}}}`),
		observed([]string{rpc("frame:turn", "turn/start", map[string]any{})}, turn...),
	}
	return c
}

func TestEachToolCallInvokesTheToolItNames(t *testing.T) {
	c := turnWithCalls(`echo {"text":"hi"}`, `read {"path":"README.md"}`)
	to, m := emitted(t, c)
	if to != "echo" || string(m["payload"]) != `{"text":"hi"}` {
		t.Fatalf("first call: %s %s", to, m["payload"])
	}
	c.Steps = append(c.Steps, step{To: "echo", Payload: m["payload"], State: "return", Value: value("hi")})
	to, m = emitted(t, c)
	var q struct{ Deliver []string }
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], `"id":"call-0"`) {
		t.Fatalf("echo's result should answer call-0: %s %v", to, q.Deliver)
	}
	c.Steps = append(c.Steps, step{To: to, Payload: m["payload"], State: "return", Value: value(map[string]any{"lines": []string{}, "delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": ""})})
	to, m = emitted(t, c)
	if to != "read" || string(m["payload"]) != `{"path":"README.md"}` {
		t.Fatalf("second call: %s %s", to, m["payload"])
	}
	c.Steps = append(c.Steps, step{To: "read", Payload: m["payload"], State: "fault", Fault: &struct{ Kind, Cause string }{"contract", "read: unknown field"}})
	to, m = emitted(t, c)
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], `"success":false`) || !strings.Contains(q.Deliver[0], "fault (contract)") {
		t.Fatalf("a tool fault goes back to Codex in the same turn: %s %v", to, q.Deliver)
	}
}

func TestAToolOutsideReachIsAViolation(t *testing.T) {
	c := turnWithCalls(`read {"path":"x"}`)
	delete(c.Tools, "read")
	to, m := emitted(t, c)
	var q struct{ Deliver []string }
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], "turn/interrupt") {
		t.Fatalf("an unoffered tool should interrupt the turn: %s %v", to, q.Deliver)
	}
	if _, err := decide(func() continuation {
		c := turnWithCalls(`read {"path":"x"}`)
		c.Steps = append(c.Steps, step{To: "echo", State: "return", Value: value("hi")})
		return c
	}()); err == nil || !strings.Contains(err.Error(), "asked for read") {
		t.Fatalf("a result from a different tool must not answer the call: %v", err)
	}
}

// A native attempt is reported as what it is: once the turn is interrupted,
// the adapter returns an observation of the attempt to its own slot. It does
// not name an unregistered capability to make the runtime fault.
func TestANativeAttemptIsReturnedAsAnObservation(t *testing.T) {
	c := turnWithCalls(`read {"path":"x"}`)
	delete(c.Tools, "read")
	to, m := emitted(t, c)
	if to != "codex.exchange" {
		t.Fatalf("the attempt interrupts the turn first: %s", to)
	}
	c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{
		"lines":    []string{`{"method":"turn/completed","params":{"turn":{"id":"turn-1","status":"interrupted"}}}`},
		"delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": "",
	})})
	to, m = emitted(t, c)
	var r struct{ Text, Refused, Thread string }
	if err := json.Unmarshal(m["payload"], &r); err != nil || to != "frame" || r.Refused != "native.dynamicTool.read" || r.Thread != "thread-1" {
		t.Fatalf("the attempt is returned to the slot, naming what was attempted: %s %s", to, m["payload"])
	}
}

// A failed turn ends with Codex's own reason, such as a usage limit: the human
// needs it, not only the status.
func TestAFailedTurnCarriesCodexsReason(t *testing.T) {
	c := turnWithCalls()
	reason := "You've hit your usage limit. Try again at Oct 3rd, 2026 11:01 PM."
	c.Steps = append(c.Steps, step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{}}),
		Value: value(map[string]any{"lines": []string{
			`{"method":"turn/completed","params":{"threadId":"thread-1","turn":{"id":"turn-1","status":"failed","error":{"message":` + string(value(reason)) + `,"codexErrorInfo":"usageLimitExceeded"}}}}`,
		}, "delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": ""})})
	_, err := decide(c)
	if err == nil || !strings.Contains(err.Error(), "turn ended failed: "+reason) {
		t.Fatalf("got %v", err)
	}
}

// With the interpreter in reach (next/32), Codex is offered remember, which
// names one of its own tool calls by index, procedure, and one tool per
// procedure that exists, by title; read and the rest stay as they were. A
// call is mapped to the instruction the adapter, a worker, derives for it.
func TestRememberAndProceduresAreOffered(t *testing.T) {
	text := &schema{Type: "object", Required: []string{"path"}, Fields: map[string]*schema{"path": {Type: "string"}}}
	remember := &schema{Title: "remember", Type: "object", Required: []string{"remember"}, Fields: map[string]*schema{"remember": {Type: "any"}}}
	look := &schema{Title: "look", Type: "object", Required: []string{"path"}, Fields: map[string]*schema{"path": {Type: "string"}}}
	c := continuation{Frame: "f", Tools: map[string]*schema{"read": text, "interpreter": remember, "codex.exchange": text},
		Contracts: map[string][]*schema{"interpreter": {remember, look}}}
	var names []string
	for _, tl := range tools(c) {
		names = append(names, tl.name)
	}
	if strings.Join(names, " ") != "read remember procedure look" {
		t.Fatalf("tools: %v", names)
	}
	results := []step{{To: "read", Payload: json.RawMessage(`{"path":"a"}`), State: "return", Observation: "obs"}}
	got := toolEmission(c, results, "remember", json.RawMessage(`{"call":0}`))
	want := `{"payload":{"remember":{"contract":{"schema":{"type":"object","fields":{"path":{"type":"string"}},"required":["path"]},"target":"read"},"observation":"obs"}},"to":"interpreter"}`
	if got != want {
		t.Fatalf("remember:\n got %s\nwant %s", got, want)
	}
	if got := toolEmission(c, results, "procedure", json.RawMessage(`{"input":{"path":"b"}}`)); got != `{"payload":{"path":"b"},"to":"interpreter"}` {
		t.Fatalf("procedure: %s", got)
	}
	if got := toolEmission(c, results, "remember", json.RawMessage(`{"call":3}`)); !strings.Contains(got, `{"remember":{}}`) {
		t.Fatalf("a call that does not exist is left for intake to refuse: %s", got)
	}
	// Without the interpreter, nothing changes.
	delete(c.Tools, "interpreter")
	if names := tools(c); len(names) != 1 || names[0].name != "read" {
		t.Fatalf("without the interpreter: %v", names)
	}
}

func TestThreadPromptIsTheHeaderOnly(t *testing.T) {
	for header, want := range map[string]string{
		"think": "You are inside Skyra, a cognitive architecture, fulfilling: think",
		"":      "You are inside Skyra, a cognitive architecture.",
	} {
		c := continuation{Frame: "frame", Tools: map[string]*schema{}}
		c.Input.Text, c.Input.Header = "hello", header
		c.Steps = []step{
			{To: "codex.open", State: "return", Value: value(map[string]any{"incarnation": "actor-1"})},
			{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{rpc("frame:initialize", "initialize", map[string]any{})}}),
				Value: value(map[string]any{"lines": []string{`{"id":"frame:initialize","result":{}}` + "\n"}, "delivery": "written", "read": "observed", "observed_at": 1, "peer_eof": false, "error": "", "partial": ""})},
		}
		_, m := emitted(t, c)
		var q struct {
			Deliver []string `json:"deliver"`
		}
		if err := json.Unmarshal(m["payload"], &q); err != nil || len(q.Deliver) != 2 {
			t.Fatalf("thread setup: %s", m["payload"])
		}
		var start struct {
			Params map[string]any `json:"params"`
		}
		if err := json.Unmarshal([]byte(q.Deliver[1]), &start); err != nil {
			t.Fatal(err)
		}
		if start.Params["baseInstructions"] != want {
			t.Fatalf("header %q: prompt %v", header, start.Params["baseInstructions"])
		}
		if _, ok := start.Params["developerInstructions"]; ok {
			t.Fatal("the old developer line is still sent")
		}
	}
}

// A wait that ends with nothing to read (service-io reports Go's "i/o timeout")
// is silence during a turn, a failure before one, and a lost server when it
// was a health check.
func TestQuietWaits(t *testing.T) {
	quietly := func(to string, deliver []string) step {
		return step{To: to, State: "return", Payload: value(map[string]any{"deliver": deliver}),
			Value: value(map[string]any{"lines": []string{}, "delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "read unix ->: i/o timeout", "partial": ""})}
	}
	// During a turn: keep waiting on Codex.
	c := turnWithCalls()
	c.Steps = append(c.Steps, quietly("codex.exchange", nil))
	if to, _ := emitted(t, c); to != "codex.exchange" {
		t.Fatalf("silence during a turn should wait again, got %s", to)
	}
	// Before a turn: the server should have answered initialize.
	c = continuation{Frame: "frame", Tools: map[string]*schema{}}
	c.Input.Text = "hello"
	c.Steps = []step{
		{To: "codex.open", State: "return", Value: value(map[string]any{"incarnation": "actor-1"})},
		quietly("codex.exchange", []string{rpc("frame:initialize", "initialize", map[string]any{})}),
	}
	if _, err := decide(c); err == nil || !strings.Contains(err.Error(), "i/o timeout") {
		t.Fatalf("a silent server before the turn must fail, got %v", err)
	}
	// A health check nobody answers: stop the server and start another.
	c = continuation{Frame: "frame"}
	c.Input.Text, c.Input.Thread, c.Input.Incarnation, c.Input.ActivityAt = "next", "thread-1", "old-process", 1000
	c.Steps = []step{quietly("codex.inspect", []string{rpc("frame:inspect", "thread/loaded/list", map[string]any{})})}
	if to, _ := emitted(t, c); to != "codex.stop" {
		t.Fatalf("an unanswered health check should stop the server, got %s", to)
	}
}

// An open that returns no session is the turn's failure, with the service
// host's reason.
func TestAnOpenWithoutASessionFails(t *testing.T) {
	c := continuation{Frame: "frame", Tools: map[string]*schema{}}
	c.Input.Text, c.Input.Header = "hello", "think"
	c.Steps = []step{{To: "codex.open", State: "return", Value: value(map[string]any{"lines": []string{}, "error": "no program \"codex\""})}}
	if _, err := decide(c); err == nil || !strings.Contains(err.Error(), `no program "codex"`) {
		t.Fatalf("an open without a session: %v", err)
	}
}

// A thread lasts one invocation (proposal 38), so a turn's memory is the
// context its invoker passed: given to Codex ahead of the message, which is
// always the turn's last input.
func TestContextPrecedesTheMessage(t *testing.T) {
	items := turnInput(value([]map[string]string{{"role": "user", "text": "tool:hello"}, {"role": "agent", "text": "hi"}}), "next")
	if len(items) != 2 {
		t.Fatalf("items: %v", items)
	}
	earlier := items[0].(map[string]string)["text"]
	if !strings.Contains(earlier, "user: tool:hello") || !strings.Contains(earlier, "agent: hi") || items[1].(map[string]string)["text"] != "next" {
		t.Fatalf("context then message: %v", items)
	}
	for _, none := range []string{``, `null`, `[]`, `{}`} {
		if items := turnInput(json.RawMessage(none), "only"); len(items) != 1 {
			t.Fatalf("context %q added an item: %v", none, items)
		}
	}
	if got := turnInput(value(map[string]string{"spec": "v1"}), "go")[0].(map[string]string)["text"]; !strings.Contains(got, `{"spec":"v1"}`) {
		t.Fatalf("other context is passed as recorded: %s", got)
	}
}

// Codex has no read of images (proposal 49, decision 2): an image view
// interrupts the turn, like any other native item but web search.
func TestViewingAnImageIsAViolation(t *testing.T) {
	c := turnWithCalls()
	c.Steps = append(c.Steps, step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{}}), Value: value(map[string]any{
		"lines":    []string{`{"method":"item/started","params":{"item":{"type":"imageView","path":"/w/a.png"},"threadId":"thread-1","turnId":"turn-1"}}`},
		"delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": "",
	})})
	to, m := emitted(t, c)
	var q struct{ Deliver []string }
	json.Unmarshal(m["payload"], &q)
	if to != "codex.exchange" || len(q.Deliver) != 1 || !strings.Contains(q.Deliver[0], "turn/interrupt") {
		t.Fatalf("an image view must interrupt the turn: %s %s", to, m["payload"])
	}
}

func TestReadingTheWebIsNotAViolation(t *testing.T) {
	for _, action := range []string{
		`{"type":"search","queries":["Skyra"],"query":"Skyra"}`,
		`{"type":"openPage","url":"https://example.com"}`,
		`{"type":"findInPage","url":"https://example.com","pattern":"Skyra"}`,
	} {
		t.Run(action, func(t *testing.T) {
			c := turnWithCalls()
			c.Steps = append(c.Steps, step{To: "codex.exchange", State: "return", Payload: value(map[string]any{"deliver": []string{}}), Value: value(map[string]any{
				"lines": []string{
					fmt.Sprintf(`{"method":"item/started","params":{"item":{"type":"webSearch","id":"web-1","action":%s},"threadId":"thread-1","turnId":"turn-1"}}`, action),
					fmt.Sprintf(`{"method":"item/completed","params":{"item":{"type":"webSearch","id":"web-1","action":%s},"threadId":"thread-1","turnId":"turn-1"}}`, action),
				},
				"delivery": "written", "read": "observed", "observed_at": 2, "peer_eof": false, "error": "", "partial": "",
			})})
			to, m := emitted(t, c)
			var poll struct{ Deliver []string }
			if err := json.Unmarshal(m["payload"], &poll); err != nil || to != "codex.exchange" || len(poll.Deliver) != 0 {
				t.Fatalf("a web read must wait for the turn, without interrupting: %s %s", to, m["payload"])
			}
			c.Steps = append(c.Steps, step{To: to, State: "return", Payload: m["payload"], Value: value(map[string]any{
				"lines": []string{
					`{"method":"item/completed","params":{"item":{"type":"agentMessage","text":"Here is what I found: https://example.com"},"threadId":"thread-1","turnId":"turn-1"}}`,
					`{"method":"turn/completed","params":{"turn":{"id":"turn-1","status":"completed"}}}`,
				},
				"delivery": "written", "read": "observed", "observed_at": 3, "peer_eof": false, "error": "", "partial": "",
			})})
			to, m = emitted(t, c)
			var answer struct{ Text, Refused string }
			if err := json.Unmarshal(m["payload"], &answer); err != nil || to != "frame" || answer.Text != "Here is what I found: https://example.com" || answer.Refused != "" {
				t.Fatalf("a turn that read the web should answer: %s %s", to, m["payload"])
			}
		})
	}
}

func TestDesignOperationsAreOfferedToParent(t *testing.T) {
	c := continuation{Tools: map[string]*schema{"design": {Type: "object"}, "design.tools": {Type: "object"}}}
	c.Input.Header = "think"
	if ts := tools(c); len(ts) != 2 || ts[1].name != "design_tools" || ts[1].target != "design.tools" {
		t.Fatalf("think tools: %v", ts)
	}
	c.Input.Header = "design: create a draft"
	if ts := tools(c); len(ts) != 2 {
		t.Fatalf("design tools: %v", ts)
	}
}

func TestBlenderOperationsAreOfferedToParent(t *testing.T) {
	c := continuation{Tools: map[string]*schema{"blender": {Type: "object"}, "blender.tools": {Type: "object"}}}
	if ts := tools(c); len(ts) != 2 || ts[1].name != "blender_tools" || ts[1].target != "blender.tools" {
		t.Fatalf("think tools: %v", ts)
	}
	c.Input.Header = "blender: inspect the scene"
	ts := tools(c)
	if len(ts) != 2 || ts[1].name != "blender_tools" || ts[1].target != "blender.tools" {
		t.Fatalf("blender tools: %v", ts)
	}
}

func TestReadResultHasPlainSource(t *testing.T) {
	source := "\tfunc X() { println(\"&\") }\n"
	value, _ := json.Marshal(map[string]any{"path": "a.go", "text": source})
	reply := toolResult(json.RawMessage(`1`), step{To: "read", State: "return", Value: value})
	result := reply["result"].(map[string]any)
	items := result["contentItems"].([]any)
	if len(items) != 2 || items[1].(map[string]string)["text"] != source {
		t.Fatal(reply)
	}
}
