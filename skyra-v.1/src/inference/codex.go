package inference

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const codexTimeout = 600 * time.Second // agent runs are slower than a raw API call

// CallCodex runs inference through the Codex CLI (`codex exec`) instead of an HTTP
// API. codex is invoked locally and authenticated via the ChatGPT login, so calls
// are covered by the Codex subscription — no per-token API billing. It is a drop-in
// for Call: the same (system, present) -> answer seam, usable anywhere an InferFunc
// is wanted.
//
// The system prompt (the cell's "header") and the present (the frame — the rendered
// children) are joined and piped to codex on stdin. The agent's FINAL message is
// captured with --output-last-message, so we read a clean answer (the tag the cell
// asked for, e.g. <label>…</label>), not codex's event stream on stdout.
//
// The call is hermetic and side-effect-free: a read-only sandbox (never writes or
// executes), --ephemeral (no persisted session files), and a throwaway working dir
// so the agent never crawls the surrounding repo.
//
// Model selection follows the user's codex config; set CODEX_MODEL to override.
func CallCodex(system, present string) (string, error) {
	work, err := os.MkdirTemp("", "skyra-codex-")
	if err != nil {
		return "", fmt.Errorf("codex: workdir: %w", err)
	}
	defer os.RemoveAll(work)
	lastPath := filepath.Join(work, "last.txt")

	args := []string{
		"exec",
		"--ephemeral",           // no persisted session files
		"--skip-git-repo-check", // the throwaway workdir is not a repo
		"-s", "read-only",       // pure inference — never write or execute
		"--color", "never",
		"--output-last-message", lastPath,
	}
	if model := os.Getenv("CODEX_MODEL"); model != "" {
		args = append(args, "-m", model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), codexTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Dir = work // the agent's working root: empty, so it has nothing to crawl
	cmd.Stdin = strings.NewReader(system + "\n\n" + present)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("codex exec: %w (%s)", err, strings.TrimSpace(stderr.String()))
	}

	out, err := os.ReadFile(lastPath)
	if err != nil {
		return "", fmt.Errorf("codex: read last message: %w", err)
	}
	answer := strings.TrimSpace(string(out))
	if answer == "" {
		return "", fmt.Errorf("codex: empty answer (model returned nothing)")
	}
	return answer, nil
}
