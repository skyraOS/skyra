// Package launch is Codex's part of a provider resident's launch (proposal
// 47): the version its profile is verified for, its environment, its private
// config.toml, and the model catalog it bootstraps once its profile is
// rendered. The runtime's launch does the rest.
package launch

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"skyra/controller/runtime/launch"
)

// Pinned is the one Codex version the resident profile is verified for
// (codex/provider/launch.py pins the same). Changing it means rerunning
// experiments/appserver/tool_surface.py and the OS probes first.
const Pinned = "0.160.0"

// quiet turns off every prompt Codex adds of its own (proposal 38): its
// permissions and approval text, its collaboration mode, its environment
// context, its apps text, and project documents such as AGENTS.md. Skyra's
// header is the whole prompt; the sandbox, not a sentence, keeps the agent in
// its Horizon. These are top-level keys in Codex's own defaults file.
const quiet = "include_permissions_instructions = false\ninclude_apps_instructions = false\ninclude_collaboration_mode_instructions = false\ninclude_environment_context = false\nproject_doc_max_bytes = 0\n"

// Provider is Codex, as the runtime's launch registers it.
type Provider struct{}

func (Provider) Pinned() string { return Pinned }

func (Provider) Configure(r *launch.Run) error {
	home := r.Paths.Home
	r.Paths.Config, r.Paths.Catalog = filepath.Join(home, "config.toml"), filepath.Join(r.Dir, "models.json")
	initial := "cli_auth_credentials_store = \"file\"\nweb_search = \"disabled\"\nsandbox_mode = \"read-only\"\napproval_policy = \"never\"\n"
	if err := os.WriteFile(r.Paths.Config, []byte(initial), 0600); err != nil {
		return err
	}
	r.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + home, "CODEX_HOME=" + home, "TMPDIR=" + r.Paths.Scratch, "HTTPS_PROXY=" + r.ProxyURL}
	return nil
}

// Ready bootstraps the model catalog: one app-server, under the rendered
// profile, initializes and writes its model metadata; the catalog keeps those
// models with every native tool and prompt of Codex's own turned off, and the
// app-server's config names it.
func (Provider) Ready(r *launch.Run) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	s, home := r.Surface, r.Paths.Home
	args := append([]string{"-f", r.ProfileFile, s.Path}, s.Args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", args...)
	cmd.Dir, cmd.Env = r.Paths.Work, r.Env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	log, err := os.OpenFile(filepath.Join(r.Dir, "bootstrap.stderr"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	cmd.Stderr = log
	if err = cmd.Start(); err != nil {
		return err
	}
	defer func() { cmd.Process.Kill(); cmd.Wait() }()
	request := map[string]any{"id": "initialize", "method": "initialize", "params": map[string]any{"clientInfo": map[string]string{"name": "skyra-bootstrap", "version": "1"}, "capabilities": map[string]bool{"experimentalApi": true}}}
	b, _ := json.Marshal(request)
	if _, err = stdin.Write(append(b, '\n')); err != nil {
		return err
	}
	line, err := bufio.NewReader(stdout).ReadBytes('\n')
	if err != nil {
		return fmt.Errorf("bootstrap initialize: %w", err)
	}
	var answer struct {
		ID     string          `json:"id"`
		Result json.RawMessage `json:"result"`
	}
	if json.Unmarshal(line, &answer) != nil || answer.ID != "initialize" || len(answer.Result) == 0 {
		return errors.New("bootstrap refused initialization")
	}
	cacheFile := filepath.Join(home, "models_cache.json")
	deadline := time.Now().Add(10 * time.Second)
	var cache []byte
	for time.Now().Before(deadline) {
		cache, err = os.ReadFile(cacheFile)
		if err == nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err != nil {
		return fmt.Errorf("bootstrap model metadata: %w", err)
	}
	var metadata struct {
		ClientVersion string           `json:"client_version"`
		Models        []map[string]any `json:"models"`
	}
	if json.Unmarshal(cache, &metadata) != nil || metadata.ClientVersion != Pinned || len(metadata.Models) == 0 {
		return errors.New("isolated model metadata does not match pinned Codex")
	}
	for _, model := range metadata.Models {
		model["tool_mode"] = "standard"
		model["multi_agent_version"] = "disabled"
		model["experimental_supported_tools"] = []any{}
		model["apply_patch_tool_type"] = nil  // no native file edits: with an environment, Codex would offer apply_patch
		model["supports_search_tool"] = false // deferred tool discovery stays off
		model["use_responses_lite"] = false   // the lite API omits hosted web search
		model["include_skills_usage_instructions"] = false
		model["include_plugin_usage_instructions"] = false
		model["include_apps_usage_instructions"] = false
	}
	models, _ := json.Marshal(map[string]any{"models": metadata.Models})
	if err = os.WriteFile(filepath.Join(r.Dir, "models.json"), models, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(home, "config.toml"), []byte(residentConfig(r.Dir)), 0600)
}

// residentConfig is the app-server's private config.toml: its catalog, its
// own sandbox and approvals, none of its own prompts, and no agents, user
// input, plan or bundled skills. Top-level keys come before any [section].
func residentConfig(run string) string {
	return fmt.Sprintf("model_catalog_json = %q\n", filepath.Join(run, "models.json")) + "cli_auth_credentials_store = \"file\"\nweb_search = \"live\"\nsandbox_mode = \"read-only\"\napproval_policy = \"never\"\n" + quiet + "[agents]\nenabled = false\n[tools.experimental_request_user_input]\nenabled = false\n[tools.update_plan]\nenabled = false\n[skills.bundled]\nenabled = false\n"
}
