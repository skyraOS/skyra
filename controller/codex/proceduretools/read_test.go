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
	got, _ := json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","model":"picked-by-model"}}`), nil, nil, nil, caller))
	if string(got) != `{"delegate":{"actor":"codex","goal":"g","model":"chosen-model"}}` {
		t.Fatalf("omitted actor: %s", got)
	}
	got, _ = json.Marshal(Payload("delegate", json.RawMessage(`{"delegate":{"goal":"g","actor":"claude","model":"picked-by-model"}}`), nil, nil, nil, caller))
	if string(got) != `{"delegate":{"actor":"claude","goal":"g"}}` {
		t.Fatalf("explicit actor uses its own default: %s", got)
	}
}
