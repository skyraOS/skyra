package main

import "testing"

// Whatever the procedure returns is the agent's line (proposal 38): its text,
// a refusal said as one, and otherwise its JSON, never dropped.
func TestAnswerIsShownByConvention(t *testing.T) {
	for in, want := range map[string]string{
		`{"text":"It says Skyra.","thread":"thread-1"}`:         "It says Skyra.",
		`{"text":"","refused":"native.commandExecution"}`:       "⚠ the agent attempted native.commandExecution; the turn was interrupted",
		`{"refused":"system.ref.1 returned a fault (custody)"}`: "⚠ system.ref.1 returned a fault (custody)",
		`"plain"`:       "plain",
		`{"answer":42}`: `{"answer":42}`,
		`[1,2]`:         `[1,2]`,
	} {
		if got := answer([]byte(in)); got != want {
			t.Fatalf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestTheTableRunsThinkOnTheInterpreter(t *testing.T) {
	if p, ok := procedures["think"]; !ok || p.Actor != "interpreter" {
		t.Fatalf("think: %+v", p)
	}
	if _, ok := procedures["codex"]; ok {
		t.Fatal("codex is reached only through a procedure, never directly")
	}
}
