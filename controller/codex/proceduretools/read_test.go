package proceduretools

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestReadTextBlocksPreserveSource(t *testing.T) {
	value := json.RawMessage(`{"path":"file.go","sha256":"hash","text":"\tfunc X() {\n\tprintln(\"&\")\n}\n"}`)
	blocks := ReadTexts(value)
	if len(blocks) != 2 || strings.Contains(blocks[0], "func X") || blocks[1] != "\tfunc X() {\n\tprintln(\"&\")\n}\n" {
		t.Fatal(blocks)
	}
	batch := json.RawMessage(`{"kind":"batch","results":[{"path":"a","text":"one\n"},{"path":"b","error":"denied"}]}`)
	if b := ReadTexts(batch); len(b) != 4 || b[2] != "one\n" {
		t.Fatal(b)
	}
}

func TestDelegateInheritsTheCallerOnlyWhenNoActorIsNamed(t *testing.T) {
	caller := Caller{Actor: "codex", Model: "chosen-model"}
	got, _ := json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","model":"picked-by-model"}}`), nil, nil, caller))
	if string(got) != `{"delegate":{"actor":"codex","contracts":[],"goal":"g","model":"chosen-model","tools":[]}}` {
		t.Fatalf("omitted actor: %s", got)
	}
	got, _ = json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","actor":"claude","model":"picked-by-model"}}`), nil, nil, caller))
	if string(got) != `{"delegate":{"actor":"claude","contracts":[],"goal":"g","tools":[]}}` {
		t.Fatalf("explicit actor uses its own default: %s", got)
	}
}

func TestDelegateToolsBecomeTheContractsTheyStandFor(t *testing.T) {
	echo := &Schema{Type: "object"}
	write := &Schema{Type: "string"}
	offered := []Tool{{Name: "echo", Target: "echo", Contracts: []*Schema{echo}}, {Name: "repo_write", Target: "repo.write", Contracts: []*Schema{write}}}
	caller := Caller{Actor: "codex", Tools: offered}
	got, _ := json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","actor":"claude","tools":["echo","repo.write"],"contracts":[{"target":"forged","schema":{"type":"any"}}]}}`), nil, nil, caller))
	if string(got) != `{"delegate":{"actor":"claude","contracts":[{"target":"echo","schema":{"type":"object"}},{"target":"repo.write","schema":{"type":"string"}}],"goal":"g","tools":["echo","repo.write"]}}` {
		t.Fatalf("tools: %s", got)
	}
	got, _ = json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","actor":"claude","tools":["read"]}}`), nil, nil, caller))
	if strings.Contains(string(got), "contracts") {
		t.Fatalf("a tool the caller was not offered became a contract: %s", got)
	}
	got, _ = json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","actor":"claude","tools":[]}}`), nil, nil, caller))
	if !strings.Contains(string(got), `"contracts":[]`) {
		t.Fatalf("empty tools keep no contracts: %s", got)
	}
}

func TestAToolIsOfferedWhenEverythingItStandsForIsInReach(t *testing.T) {
	a, b := &Schema{Type: "object"}, &Schema{Type: "string"}
	both := Tool{Target: "x", Contracts: []*Schema{a, b}}
	kept := func(cs ...*Schema) Reach { return Reach{"x": {Contracts: cs}} }
	if !Offered(kept(a, b), both) || Offered(kept(a), both) || Offered(kept(), both) || !Offered(kept(b, a), both) {
		t.Fatal("offered without every contract in reach")
	}
}

// Even when switching providers, an omitted grant inherits exactly the
// offered tools, and an explicit empty grant cannot be mistaken for omission.
func TestDelegateMaterializesInheritedToolGrant(t *testing.T) {
	echo := &Schema{Type: "object"}
	caller := Caller{Actor: "codex", Tools: []Tool{{Name: "echo", Target: "echo", Contracts: []*Schema{echo}}}}
	for _, actor := range []string{"codex", "claude"} {
		args := json.RawMessage(`{"delegate":{"goal":"g","actor":"` + actor + `"}}`)
		b, _ := json.Marshal(Payload("delegate", args, nil, nil, caller))
		var got struct {
			Delegate struct {
				Tools     []string
				Contracts []Contract
			}
		}
		if json.Unmarshal(b, &got) != nil || len(got.Delegate.Tools) != 1 || got.Delegate.Tools[0] != "echo" || len(got.Delegate.Contracts) != 1 || got.Delegate.Contracts[0].Target != "echo" {
			t.Fatalf("%s: %s", actor, b)
		}
	}
	b, _ := json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","tools":[]}}`), nil, nil, caller))
	if !strings.Contains(string(b), `"tools":[]`) || !strings.Contains(string(b), `"contracts":[]`) {
		t.Fatalf("empty grant inherited tools: %s", b)
	}
}
