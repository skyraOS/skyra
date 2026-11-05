package main

import (
	"encoding/json"
	"strings"
	"testing"
)

const view = `{"contracts":[{"ref":"sha256:` + "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" + `","target":"read","input":{"type":"string"},"scope":{}},{"ref":"sha256:` + "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" + `","target":"read","input":{"type":"number"},"scope":{}},{"ref":"sha256:` + "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" + `","target":"echo","input":{"type":"string"},"scope":{}}]}`

func TestProposeResolvesToolsFromTheViewedContracts(t *testing.T) {
	result, err := propose([]byte(`{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":["read"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Goal         string   `json:"goal"`
		Capabilities []string `json:"capabilities"`
		To           any      `json:"to"`
	}
	if err = json.Unmarshal(result, &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Goal != "work" || spec.To != nil || len(spec.Capabilities) != 2 || !strings.Contains(spec.Capabilities[0], "aaaa") || !strings.Contains(spec.Capabilities[1], "bbbb") {
		t.Fatalf("a tool's contracts were not resolved from the view: %s", result)
	}
}

func TestProposeRefusesWhatTheViewDoesNotHold(t *testing.T) {
	for name, raw := range map[string]string{
		"no profile":      `{"propose_spec":{"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":["read"]}}`,
		"act bound of 0":  `{"propose_spec":{"profile":{"act":0,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":["read"]}}`,
		"unviewed tool":   `{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":["write"]}}`,
		"not a view":      `{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":{"tools":["read"]},"capabilities":["read"]}}`,
		"undeclared":      `{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":[],"grant":true}}`,
		"repeated tool":   `{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":["read","read"]}}`,
		"refs, not tools": `{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"text":"work"}],"inputs":[],"catalogue":` + view + `,"capabilities":["sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"]}}`,
	} {
		if _, err := propose([]byte(raw)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}

// Every word of the goal is accounted for: quotes of at least three words from
// an observation, by its id, are cited; text, and shorter quotes, are Plan's.
func TestCoverageCountsEveryWordOfTheGoal(t *testing.T) {
	id := strings.Repeat("d", 64)
	result, err := propose([]byte(`{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":[{"quote":"   12| echo the whole\n   13| request","of":"` + id + `"},{"text":"then stop."},{"quote":"just this","of":"` + id + `"}],"inputs":[],"catalogue":` + view + `,"capabilities":["echo"]}}`))
	if err != nil {
		t.Fatal(err)
	}
	var spec struct {
		Goal     string `json:"goal"`
		Segments []struct{ Text, Source, Of string }
		Coverage struct{ Words, Cited, Plan, Human int }
	}
	if err = json.Unmarshal(result, &spec); err != nil {
		t.Fatal(err)
	}
	if spec.Goal != "echo the whole\nrequest then stop. just this" || spec.Coverage.Words != 8 || spec.Coverage.Cited != 4 || spec.Coverage.Plan != 4 || spec.Coverage.Human != 0 {
		t.Fatalf("coverage: %s", result)
	}
	if spec.Segments[0].Source != "cited" || spec.Segments[0].Of != id || spec.Segments[2].Source != "plan" || spec.Segments[2].Of != "" {
		t.Fatalf("segments: %s", result)
	}
	for name, goal := range map[string]string{
		"quote of nothing":  `[{"quote":"echo the whole request"}]`,
		"quote and text":    `[{"quote":"echo the whole request","of":"` + id + `","text":"x"}]`,
		"text claiming one": `[{"text":"echo the whole request","of":"` + id + `"}]`,
		"empty segment":     `[{}]`,
	} {
		if _, err := propose([]byte(`{"propose_spec":{"profile":{"act":1,"turns":8,"depth":3,"fanout":32,"budget":64},"goal":` + goal + `,"inputs":[],"catalogue":` + view + `,"capabilities":["echo"]}}`)); err == nil {
			t.Errorf("%s accepted", name)
		}
	}
}
