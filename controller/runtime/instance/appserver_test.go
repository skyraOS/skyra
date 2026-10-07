package instance

import (
	"os"
	"strings"
	"testing"
)

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
	launcher, err := os.ReadFile("../../codex/provider/launch.py")
	if err != nil {
		t.Fatal(err)
	}
	py := strings.NewReplacer("('", "", "'\n         '", "", "'\n", "", "')", "", `\n`, "\n").Replace(string(launcher))
	if !strings.Contains(py, "QUIET = "+quiet) {
		t.Fatal("launch.py's QUIET and appserver.go's quiet differ")
	}
}
