package main

import (
	"encoding/json"
	"testing"
)

func TestLinkedSpecIsContextRatherThanChatOrTreeInput(t *testing.T) {
	transcript := []entry{{Role: "user", Text: "Discuss the next step."}}
	without, _ := json.Marshal(workContext(transcript, nil))
	expected, _ := json.Marshal(transcript)
	if string(without) != string(expected) {
		t.Fatal("legacy chat shape changed")
	}
	spec := json.RawMessage(`{"spec":"Read the README","revision":2,"treeId":"tree-c"}`)
	context := workContext(transcript, spec).(map[string]any)
	if string(context["spec"].(json.RawMessage)) != string(spec) || len(transcript) != 1 || transcript[0].Text != "Discuss the next step." {
		t.Fatal("spec was folded into the transcript")
	}
}
