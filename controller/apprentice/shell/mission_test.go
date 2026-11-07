package shell

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"skyra/apprentice/proc"
	"skyra/typesys-prototype/tree"
)

func TestImportsEnterTheEndowmentAndPinnedContractsSurviveShellReplay(t *testing.T) {
	dir := t.TempDir()
	frame := "typed-inputs"
	reach := map[string][]*proc.Schema{"model": {{Type: "object"}}, Name: {{Type: "object"}}, "echo": {{Type: "object", Fields: map[string]*proc.Schema{"text": {Type: "string"}}, Required: []string{"text"}}}}
	pins := map[string]string{"model": "sha256:" + strings.Repeat("a", 64), Name: "sha256:" + strings.Repeat("b", 64), "echo": "sha256:" + strings.Repeat("c", 64)}
	ask := Ask{Goal: "Use the evidence", Kind: "act", Actor: "model", Reach: []string{"echo"}, Imports: []Import{{Ref: strings.Repeat("d", 64), Binary: pins[Name], Contract: `{"type":"string"}`, Value: json.RawMessage(`"evidence"`), Source: json.RawMessage(`{"target":"plan","receipt":"verified"}`)}}}
	input, _ := json.Marshal(map[string]any{"tree": ask})
	if _, err := ResumePinned(dir, frame, input, reach, pins, nil); err != nil {
		t.Fatal(err)
	}
	s := &shell{dir: filepath.Join(dir, "tree"), frame: name(frame)}
	if err := s.load(); err != nil {
		t.Fatal(err)
	}
	s.reach, s.pins = s.state.Reach, s.state.Binaries
	if err := s.setup(*s.state.Ask); err != nil {
		t.Fatal(err)
	}
	s.tree = tree.NewTree(s.app, s.profile)
	for _, log := range s.state.Log {
		event, err := s.decode(log)
		if err != nil {
			t.Fatal(err)
		}
		var effects []tree.Effect
		s.tree, effects = tree.Step(s.tree, event)
		s.names(effects)
	}
	if s.self.Binary() != tree.Binary(pins["model"]) {
		t.Fatal("contract still uses a hash of the target name")
	}
	root := s.tree.Agents[tree.AgentID("0")]
	if len(root.Names) != 1 {
		t.Fatal("an import entered the tree as its own observation")
	}
	text, _ := s.values["0"][0].V.(string)
	if !strings.HasPrefix(text, "Use the evidence\n") || !strings.Contains(text, `{"target":"plan","receipt":"verified"}`) || !strings.HasSuffix(text, "\nevidence") {
		t.Fatalf("import value or source lost from the endowment on replay: %q", text)
	}
	if s.state.Log[0].Seen[0] == tree.Digest(text) {
		t.Fatal("the endowment's refinement does not bind the import's reference and source")
	}
	if _, err := View(dir); err != nil {
		t.Fatal("view could not replay pinned typed inputs", err)
	}
}

func TestEqualImportedValuesFromDifferentSourcesAreDifferentEndowments(t *testing.T) {
	seen := func(source string) string {
		dir := t.TempDir()
		reach := map[string][]*proc.Schema{"model": {{Type: "object"}}, Name: {{Type: "object"}}}
		ask := Ask{Goal: "g", Kind: "act", Actor: "model", Imports: []Import{{Ref: strings.Repeat("d", 64), Binary: "sha256:" + strings.Repeat("b", 64), Contract: `{"type":"string"}`, Value: json.RawMessage(`"same"`), Source: json.RawMessage(source)}}}
		input, _ := json.Marshal(map[string]any{"tree": ask})
		if _, err := Resume(dir, "f", input, reach, nil); err != nil {
			t.Fatal(err)
		}
		s := &shell{dir: filepath.Join(dir, "tree"), frame: name("f")}
		if err := s.load(); err != nil {
			t.Fatal(err)
		}
		return s.state.Log[0].Seen[0]
	}
	if seen(`{"target":"plan"}`) == seen(`{"target":"other"}`) {
		t.Fatal("two equal values from different sources share a refinement")
	}
}

func TestMalformedImportRefusesTheTree(t *testing.T) {
	reach := map[string][]*proc.Schema{"model": {{Type: "object"}}, Name: {{Type: "object"}}}
	ask := Ask{Goal: "g", Kind: "act", Actor: "model", Imports: []Import{{Ref: "r", Binary: "sha256:" + strings.Repeat("b", 64), Contract: `{"type": "string"}`, Value: json.RawMessage(`"v"`), Source: json.RawMessage(`{}`)}}}
	input, _ := json.Marshal(map[string]any{"tree": ask})
	out, err := Resume(t.TempDir(), "f", input, reach, nil)
	if err != nil || !strings.Contains(string(out), "refused") {
		t.Fatalf("a non-canonical import contract was accepted: %s %v", out, err)
	}
}
