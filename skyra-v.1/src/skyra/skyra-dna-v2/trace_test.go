package dnav2

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"skyra-v1/src/skyra"
)

func readTraceEvents(t *testing.T, dir string) []TraceEvent {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("trace dir entries=%v err=%v, want exactly one run file", entries, err)
	}
	raw, err := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	if err != nil {
		t.Fatal(err)
	}
	var events []TraceEvent
	for _, line := range strings.Split(strings.TrimSpace(string(raw)), "\n") {
		var e TraceEvent
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			t.Fatalf("bad JSONL line %q: %v", line, err)
		}
		events = append(events, e)
	}
	return events
}

// One traced board run must capture full inference content, the rejection that
// caused a retry (with its parse error), step spans, and worker notes — the
// exact record the first live run lacked.
func TestBoardRunTraceCapturesContentRejectionsAndSpans(t *testing.T) {
	d := testDNA()
	WireGenes(d)
	base := gradeMock("<match>none</match>")
	decomposeCalls := 0
	d.SetInfer(func(system, present string) (string, error) {
		if strings.Contains(system, "decompose one observed utterance") {
			decomposeCalls++
			if decomposeCalls == 1 {
				return "hello there skyra", nil // unannotated — must be rejected
			}
			return "michael[1] greets[1→2] skyra[2]", nil
		}
		return base(system, present)
	})
	d.SetEmit(func(skyra.Impulse) {})
	o, _ := newOrch(d)
	dir := t.TempDir()
	o.SetTraces(NewTraceStore(dir))

	fireTurn(t, o, "X1", "", "hello skyra")

	events := readTraceEvents(t, dir)
	kinds := map[string]int{}
	var sawDecomposeContent, sawReject, sawSensorySpan, sawNote bool
	for _, e := range events {
		kinds[e.Kind]++
		if e.Kind == "infer" && e.Stage == "decompose" &&
			strings.Contains(e.Present, "hello skyra") && e.Response != "" {
			sawDecomposeContent = true
		}
		if e.Kind == "reject" && e.Stage == "decompose" &&
			e.Response == "hello there skyra" && e.Err != "" {
			sawReject = true
		}
		if e.Kind == "span" && e.Step == string(StepSensoryIntake) && e.Condition == "external" {
			sawSensorySpan = true
		}
		if e.Kind == "note" && e.Stage == "observation" && strings.HasPrefix(e.Note, "persisted causal:") {
			sawNote = true
		}
	}
	if !sawDecomposeContent {
		t.Fatalf("no full-content decompose infer event: %+v", kinds)
	}
	if !sawReject {
		t.Fatal("the rejected first decomposition (and its parse error) was not traced")
	}
	if !sawSensorySpan {
		t.Fatal("sensory span missing from trace")
	}
	if !sawNote {
		t.Fatal("observation note missing from trace")
	}
	if kinds["infer"] < 4 {
		t.Fatalf("expected the turn's inference calls in the trace, got %d", kinds["infer"])
	}
	for i, e := range events {
		if e.Seq != i+1 {
			t.Fatalf("event seq %d at position %d — ordering broken", e.Seq, i)
		}
	}
}
