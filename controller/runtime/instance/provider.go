package instance

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/hold"
)

type preparedResident struct {
	profile, work string
	env           []string
	watch         *os.File
	helpers       []*exec.Cmd // the egress proxy and, for keychain auth, the keychain broker
}

// quiet turns off every prompt Codex adds of its own (proposal 38): its
// permissions and approval text, its collaboration mode, its environment
// context, its apps text, and project documents such as AGENTS.md. Skyra's
// header is the whole prompt; the sandbox, not a sentence, keeps the agent in
// its Horizon. These are top-level keys in Codex's own defaults file.
const quiet = "include_permissions_instructions = false\ninclude_apps_instructions = false\ninclude_collaboration_mode_instructions = false\ninclude_environment_context = false\nproject_doc_max_bytes = 0\n"

// PinnedCodex is the one Codex version the resident profile is verified for
// (codex/provider/launch.py pins the same). Changing it means rerunning
// experiments/appserver/tool_surface.py and the OS probes first.
const PinnedCodex = "0.160.0"

// PinnedClaude is the one Claude Code version the provider profile is
// verified for (proposal 47). Changing it means rerunning a turn under the
// rendered profile and reading its sandbox denials first.
const PinnedClaude = "2.1.309"

var pinned = map[string]string{"codex": PinnedCodex, "claude": PinnedClaude}

func (p *preparedResident) abort() {
	if p == nil {
		return
	}
	if p.watch != nil {
		p.watch.Close()
		p.watch = nil
	}
	for _, h := range p.helpers {
		if h.Process != nil {
			h.Process.Kill()
			h.Wait()
		}
	}
	p.helpers = nil
}

// actorStarted hands the helpers to the resident: they hold the read end of
// the watch pipe, the resident now holds its only write end, and they exit
// when it does.
func (p *preparedResident) actorStarted() {
	if p == nil {
		return
	}
	if p.watch != nil {
		p.watch.Close()
		p.watch = nil
	}
	for _, h := range p.helpers {
		h.Process.Release()
	}
	p.helpers = nil
}

// prepareProvider makes one provider resident's private run directory, starts
// its egress proxy and, for keychain auth, its keychain broker, writes its
// config and renders its profile (proposal 47). Codex and Claude take this
// one path; what differs is the surface's Provider and the config each CLI
// reads.
func prepareProvider(x machine.StartResident, root string) (p *preparedResident, err error) {
	s := x.Run.Surface
	v := s.Provider
	if s.ResidentProfile != "darwin-provider-v1" || v == nil {
		return nil, errors.New("unknown resident profile")
	}
	proxyBytes, err := os.ReadFile(s.ProxyPath)
	if err != nil {
		return nil, err
	}
	if record.Sum(proxyBytes) != s.ProxySHA256 {
		return nil, errors.New("registered proxy changed")
	}
	if v.Program != s.Path {
		program, err := os.ReadFile(v.Program)
		if err != nil {
			return nil, err
		}
		if record.Sum(program) != v.ProgramSHA256 {
			return nil, errors.New("registered provider program changed")
		}
	}
	version, err := exec.Command(v.Program, "--version").Output()
	if err != nil || pinned[v.Name] == "" || !slices.Contains(strings.Fields(string(version)), pinned[v.Name]) {
		return nil, fmt.Errorf("resident profile is verified for %s %s", v.Name, pinned[v.Name])
	}
	base := filepath.Join(filepath.Dir(root), "providers")
	if err = os.MkdirAll(base, 0700); err != nil {
		return nil, err
	}
	run, err := os.MkdirTemp(base, v.Name+"-")
	if err != nil {
		return nil, err
	}
	run, err = filepath.EvalSymlinks(run)
	if err != nil {
		return nil, err
	}
	home, tmp, work := filepath.Join(run, "home"), filepath.Join(run, "tmp"), filepath.Join(run, "work")
	for _, dir := range []string{home, tmp, work} {
		if err = os.Mkdir(dir, 0700); err != nil {
			return nil, err
		}
	}
	cacheOut, err := exec.Command("/usr/bin/getconf", "DARWIN_USER_CACHE_DIR").Output()
	if err != nil {
		return nil, err
	}
	cache, err := filepath.EvalSymlinks(strings.TrimSpace(string(cacheOut)))
	if err != nil {
		return nil, err
	}
	paths := horizon.ResidentPaths{Home: home, Scratch: tmp, Work: work, Cache: cache}
	switch v.Auth {
	case "file":
		if paths.Auth, err = filepath.EvalSymlinks(v.AuthPath); err != nil {
			return nil, err
		}
		if err = os.Symlink(paths.Auth, filepath.Join(home, filepath.Base(v.AuthPath))); err != nil {
			return nil, err
		}
	case "keychain":
		// The helper's copy is the only `security` the sandbox can run: on
		// PATH, in a directory the resident cannot write.
		helper, err := os.ReadFile(v.Helper)
		if err != nil {
			return nil, err
		}
		if record.Sum(helper) != v.HelperSHA256 {
			return nil, errors.New("registered keychain helper changed")
		}
		paths.Helper = filepath.Join(run, "bin", "security")
		if err = os.Mkdir(filepath.Dir(paths.Helper), 0700); err != nil {
			return nil, err
		}
		if err = os.WriteFile(paths.Helper, helper, 0500); err != nil {
			return nil, err
		}
		// A socket path stays under 104 bytes on macOS, so the broker's is in
		// the user's private socket directory, not the run directory.
		dir, err := hold.Dir()
		if err != nil {
			return nil, err
		}
		paths.Broker = filepath.Join(dir, "kc-"+filepath.Base(run)+".sock")
	default:
		return nil, fmt.Errorf("unknown provider auth %q", v.Auth)
	}
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	p = &preparedResident{work: work, watch: write}
	cleanup := p
	defer func() {
		read.Close()
		if err != nil {
			cleanup.abort()
		}
	}()
	log, err := os.OpenFile(filepath.Join(run, "proxy.stderr"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer log.Close()
	journal, ready := filepath.Join(run, "proxy.jsonl"), filepath.Join(run, "proxy.port")
	args := []string{s.ProxyPath, "--journal", journal, "--ready", ready, "--watch-fd", "3"}
	for _, host := range s.ProxyHosts {
		args = append(args, "--allow", host)
	}
	proxy := exec.Command("/usr/bin/python3", args...)
	proxy.Stdout, proxy.Stderr = log, log
	proxy.ExtraFiles = []*os.File{read}
	if err = proxy.Start(); err != nil {
		return nil, err
	}
	p.helpers = append(p.helpers, proxy)
	if v.Auth == "keychain" {
		broker := exec.Command(paths.Helper, "serve", "-socket", paths.Broker, "-service", v.Keychain, "-watch-fd", "3")
		broker.Args[0] = "claude-keychain" // as security, it would be the shim
		broker.Stdout, broker.Stderr = log, log
		broker.ExtraFiles = []*os.File{read}
		if err = broker.Start(); err != nil {
			return nil, err
		}
		p.helpers = append(p.helpers, broker)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if b, e := os.ReadFile(ready); e == nil {
			paths.ProxyPort, e = strconv.Atoi(strings.TrimSpace(string(b)))
			if e == nil && paths.ProxyPort > 0 && paths.ProxyPort < 65536 {
				break
			}
		}
		if e := proxy.Process.Signal(syscall.Signal(0)); e != nil {
			return nil, fmt.Errorf("egress proxy exited: %w", e)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if paths.ProxyPort == 0 {
		return nil, errors.New("egress proxy did not bind")
	}
	for paths.Broker != "" {
		if info, e := os.Stat(paths.Broker); e == nil && info.Mode()&os.ModeSocket != 0 {
			break
		}
		if time.Now().After(deadline) {
			return nil, errors.New("keychain broker did not listen")
		}
		time.Sleep(20 * time.Millisecond)
	}
	proxyURL := "http://127.0.0.1:" + strconv.Itoa(paths.ProxyPort)
	switch v.Name {
	case "codex":
		paths.Config, paths.Catalog = filepath.Join(home, "config.toml"), filepath.Join(run, "models.json")
		initial := "cli_auth_credentials_store = \"file\"\nweb_search = \"disabled\"\nsandbox_mode = \"read-only\"\napproval_policy = \"never\"\n"
		if err = os.WriteFile(paths.Config, []byte(initial), 0600); err != nil {
			return nil, err
		}
		p.env = []string{"PATH=/usr/bin:/bin", "HOME=" + home, "CODEX_HOME=" + home, "TMPDIR=" + tmp, "HTTPS_PROXY=" + proxyURL}
	case "claude":
		if paths.BridgePort, err = freePort(); err != nil {
			return nil, err
		}
		if paths.Config, err = claudeSettings(home); err != nil {
			return nil, err
		}
		user := os.Getenv("USER")
		if user == "" {
			user = filepath.Base(os.Getenv("HOME"))
		}
		p.env = []string{"PATH=" + filepath.Dir(paths.Helper) + ":/usr/bin:/bin", "HOME=" + home, "TMPDIR=" + tmp, "CLAUDE_CODE_TMPDIR=" + tmp, "USER=" + user,
			"HTTPS_PROXY=" + proxyURL, "NO_PROXY=127.0.0.1,localhost", "DISABLE_AUTOUPDATER=1",
			"SKYRA_CLAUDE_PROGRAM=" + v.Program, "SKYRA_CLAUDE_BRIDGE_PORT=" + strconv.Itoa(paths.BridgePort), "SKYRA_KEYCHAIN_SOCKET=" + paths.Broker}
	default:
		return nil, fmt.Errorf("unknown provider %q", v.Name)
	}
	p.profile, err = horizon.RenderResident(x.Horizon, x.Floor, paths)
	if err != nil {
		return nil, err
	}
	if err = os.WriteFile(filepath.Join(run, "actor.sb"), []byte(p.profile), 0600); err != nil {
		return nil, err
	}
	if v.Name == "codex" {
		if err = bootstrapCatalog(s, p, run); err != nil {
			return nil, err
		}
	}
	return p, nil
}

// claudeSettings gives the resident the user's default Claude Code profile:
// a copy of ~/.claude/settings.json in its private home, which it reads and
// cannot change. Hooks, plugins and MCP stay off by the CLI's own flags.
func claudeSettings(home string) (string, error) {
	settings := []byte("{}\n")
	if userHome, err := os.UserHomeDir(); err == nil {
		if b, err := os.ReadFile(filepath.Join(userHome, ".claude", "settings.json")); err == nil {
			settings = b
		}
	}
	dir := filepath.Join(home, ".claude")
	if err := os.Mkdir(dir, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "settings.json")
	return path, os.WriteFile(path, settings, 0600)
}

// freePort is a loopback port nothing listens on now, for a resident to
// serve its own tools on: the profile names that one port, not every local
// service.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func bootstrapCatalog(s horizon.Surface, p *preparedResident, run string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	args := append([]string{"-f", filepath.Join(run, "actor.sb"), s.Path}, s.Args...)
	cmd := exec.CommandContext(ctx, "/usr/bin/sandbox-exec", args...)
	cmd.Dir, cmd.Env = p.work, p.env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	log, err := os.OpenFile(filepath.Join(run, "bootstrap.stderr"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
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
	cacheFile := filepath.Join(run, "home", "models_cache.json")
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
	if json.Unmarshal(cache, &metadata) != nil || metadata.ClientVersion != PinnedCodex || len(metadata.Models) == 0 {
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
	if err = os.WriteFile(filepath.Join(run, "models.json"), models, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(run, "home", "config.toml"), []byte(residentConfig(run)), 0600)
}

// residentConfig is the app-server's private config.toml: its catalog, its
// own sandbox and approvals, none of its own prompts, and no agents, user
// input, plan or bundled skills. Top-level keys come before any [section].
func residentConfig(run string) string {
	return fmt.Sprintf("model_catalog_json = %q\n", filepath.Join(run, "models.json")) + "cli_auth_credentials_store = \"file\"\nweb_search = \"live\"\nsandbox_mode = \"read-only\"\napproval_policy = \"never\"\n" + quiet + "[agents]\nenabled = false\n[tools.experimental_request_user_input]\nenabled = false\n[tools.update_plan]\nenabled = false\n[skills.bundled]\nenabled = false\n"
}
