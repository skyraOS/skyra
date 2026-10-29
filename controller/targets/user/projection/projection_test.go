package projection

import (
	"encoding/json"
	"strings"
	"testing"
)

func testTree() Tree {
	return Tree{Version: 1, Root: "root", Frontier: []string{"joined"}, Order: []string{"root", "browser", "human", "turn", "read", "read-result", "answer", "joined"}, Records: map[string]Record{
		"root":        {State: "invoke", To: "user", Payload: json.RawMessage(`{"conversation":"c-1"}`)},
		"browser":     {Predecessor: "root", State: "invoke", To: "browser"},
		"human":       {Predecessor: "browser", State: "return", Payload: json.RawMessage(`{"Text":"hi","requestId":"req"}`)},
		"turn":        {Predecessor: "human", State: "invoke", To: "codex", Payload: json.RawMessage(`{"text":"hi"}`)},
		"read":        {Predecessor: "turn", State: "invoke", To: "read", Payload: json.RawMessage(`{"path":"README.md"}`)},
		"read-result": {Predecessor: "read", State: "return", Payload: json.RawMessage(`{"text":"hi"}`)},
		"answer":      {Predecessor: "read-result", State: "return", Payload: json.RawMessage(`{"text":"hello","thread":"provider-only"}`)},
		"joined":      {Predecessor: "answer", State: "resolve", Frame: "root", Deps: []string{"answer"}},
	}, Boundaries: map[string]Boundary{
		"root":  {Frame: "root", Name: "user", Reach: map[string]json.RawMessage{"read": json.RawMessage(`{"path":"/read"}`)}, Authority: json.RawMessage(`{"read":[{"path":"/workspace","scope":"subpath"}]}`)},
		"human": {Frame: "browser"}, "read-result": {Frame: "read"}, "answer": {Frame: "turn"}, "joined": {Frame: "root"},
	}}
}
func TestSnapshotProjectsAcceptedWorkAndFitsAppShape(t *testing.T) {
	tr := testTree()
	view, err := ConversationState("c-1", "root", tr, Availability{InputLive: true})
	if err != nil {
		t.Fatal(err)
	}
	if view.Conversation.Phase != "ready" || len(view.Conversation.Entries) != 3 || view.Conversation.PendingRequestID != nil || view.Requests[0].Status != "returned" {
		t.Fatalf("view: %+v", view)
	}
	call := view.Conversation.Entries[1].Call
	if call == nil || call.To != "read" || call.State != "returned" {
		t.Fatalf("call: %+v", call)
	}
	if len(view.Conversation.Boundaries) == 0 || !view.Conversation.Boundaries[0].Root {
		t.Fatal("authorized Boundary visualization disappeared")
	}
	snap, err := AppState("webhost-1", 3, []View{view})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(snap)
	if strings.Contains(string(b), "provider-only") {
		t.Fatal("provider state leaked to the app")
	}
	var obj map[string]any
	json.Unmarshal(b, &obj)
	if obj["type"] != "snapshot" || obj["revision"] != float64(3) || obj["requests"] == nil {
		t.Fatalf("invalid app snapshot: %s", b)
	}
	for i, e := range view.Conversation.Entries {
		if e.TS != i {
			t.Fatal("noncausal display ordering")
		}
	}
}
func TestChildCompletionDoesNotClaimParentConsumedIt(t *testing.T) {
	tr := testTree()
	delete(tr.Records, "joined")
	tr.Order = tr.Order[:len(tr.Order)-1]
	tr.Frontier = []string{"answer"}
	view, err := ConversationState("c-1", "root", tr, Availability{TurnLive: true})
	if err != nil {
		t.Fatal(err)
	}
	if view.Requests[0].Status != "accepted" || view.Conversation.PendingRequestID == nil || view.Conversation.Phase != "pending" {
		t.Fatalf("unjoined answer marked consumed: %+v", view)
	}
	// Lost custody does not turn historical pending work into a live turn.
	lost, err := ConversationState("c-1", "root", tr, Availability{})
	if err != nil || lost.Conversation.Phase != "closed" {
		t.Fatalf("lost custody: %+v %v", lost, err)
	}
}
func TestProjectionRejectsPartialAndInvalidPublications(t *testing.T) {
	tr := testTree()
	if _, err := ConversationState("c-1", "different-root", tr, Availability{}); err == nil {
		t.Fatal("wrong anchor accepted")
	}
	if _, err := ConversationState("c-1", "root", tr, Availability{TurnLive: true, InputLive: true}); err == nil {
		t.Fatal("incompatible live state")
	}
	view, _ := ConversationState("c-1", "root", tr, Availability{})
	if _, err := AppState("host", 1, []View{view, view}); err == nil {
		t.Fatal("duplicate views accepted")
	}
	empty, err := AppState("host", 0, nil)
	if err != nil || empty.Conversations == nil || empty.Requests == nil {
		t.Fatal("empty replacement must use arrays")
	}
}

func TestRefusedToolEmissionIsProjectedFromVisibleObservedInput(t *testing.T) {
	tr := testTree()
	tr.Order = []string{"root", "browser", "human", "turn", "refused", "joined"}
	tr.Records["refused"] = Record{Predecessor: "turn", State: "fault", Fault: &Fault{Kind: "boundary", Cause: "outside reach"}}
	tr.Records["joined"] = Record{Predecessor: "refused", State: "resolve", Deps: []string{"refused"}}
	tr.Boundaries["refused"] = Boundary{Frame: "turn"}
	tr.Inputs = map[string][]byte{"refused": []byte(`{"to":"forbidden","payload":{"text":"attempt"}}`)}
	view, err := ConversationState("c-1", "root", tr, Availability{RecoveryRequired: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Conversation.Entries) != 2 || view.Requests[0].Status != "failed" || view.Conversation.PendingRequestID != nil {
		t.Fatalf("refused turn: %+v", view)
	}
	c := view.Conversation.Entries[1].Call
	if c == nil || c.To != "forbidden" || c.State != "fault" || c.Fault != "boundary" {
		t.Fatalf("refused tool: %+v", c)
	}
}
