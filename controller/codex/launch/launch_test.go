package launch

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Both launch paths must offer live web reads with the same read-only local
// policy. Compare their actual generated configs to catch launcher drift.
func TestResidentConfigMatchesProviderLauncher(t *testing.T) {
	config := residentConfig("/run")
	for _, setting := range []string{`web_search = "live"`, `sandbox_mode = "read-only"`, `approval_policy = "never"`} {
		if !strings.Contains(config, setting+"\n") {
			t.Fatalf("missing %s in resident config", setting)
		}
	}
	cmd := exec.Command("python3", "-B", "-c", `import sys
from pathlib import Path
sys.path.insert(0, '../provider')
from launch import configuration
sys.stdout.write(configuration(Path('/run')))`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("provider configuration: %v: %s", err, out)
	}
	if string(out) != config {
		t.Fatalf("launcher configs differ:\nGo:\n%s\nPython:\n%s", config, out)
	}
}

// Codex adds none of its own prompts (proposal 38): every switch is off, as a
// top-level key, before the first [section] where TOML would nest it. The
// Python launcher writes the same switches.
func TestCodexAddsNoPromptsOfItsOwn(t *testing.T) {
	config := residentConfig("/run")
	top := config[:strings.Index(config, "\n[")+1]
	for _, line := range []string{
		"include_permissions_instructions = false",
		"include_apps_instructions = false",
		"include_collaboration_mode_instructions = false",
		"include_environment_context = false",
		"project_doc_max_bytes = 0",
	} {
		if !strings.Contains(top, line+"\n") {
			t.Fatalf("%q is not a top-level key:\n%s", line, config)
		}
	}
	launcher, err := os.ReadFile("../provider/launch.py")
	if err != nil {
		t.Fatal(err)
	}
	py := strings.NewReplacer("('", "", "'\n         '", "", "'\n", "", "')", "", `\n`, "\n").Replace(string(launcher))
	if !strings.Contains(py, "QUIET = "+quiet) {
		t.Fatal("launch.py's QUIET and launch.go's quiet differ")
	}
}
