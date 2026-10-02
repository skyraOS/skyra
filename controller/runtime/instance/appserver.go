package instance

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
)

type preparedResident struct {
	profile, work string
	env           []string
	watch         *os.File
	proxy         *exec.Cmd
}

// PinnedCodex is the one Codex version the resident profile is verified for
// (codex/provider/launch.py pins the same). Changing it means rerunning
// experiments/appserver/tool_surface.py and the OS probes first.
const PinnedCodex = "0.160.0"

func (p *preparedResident) abort() {
	if p == nil {
		return
	}
	if p.watch != nil {
		p.watch.Close()
		p.watch = nil
	}
	if p.proxy != nil && p.proxy.Process != nil {
		p.proxy.Process.Kill()
		p.proxy.Wait()
		p.proxy = nil
	}
}
func (p *preparedResident) actorStarted() {
	if p == nil {
		return
	}
	if p.watch != nil {
		p.watch.Close()
		p.watch = nil
	}
	if p.proxy != nil {
		p.proxy.Process.Release()
		p.proxy = nil
	}
}

func prepareAppserver(x machine.StartResident, root string) (p *preparedResident, err error) {
	s := x.Run.Surface
	if s.ResidentProfile != "darwin-appserver-v1" {
		return nil, errors.New("unknown resident profile")
	}
	proxyBytes, err := os.ReadFile(s.ProxyPath)
	if err != nil {
		return nil, err
	}
	if record.Sum(proxyBytes) != s.ProxySHA256 {
		return nil, errors.New("registered proxy changed")
	}
	version, err := exec.Command(s.Path, "--version").Output()
	if err != nil || strings.TrimSpace(string(version)) != "codex-cli "+PinnedCodex {
		return nil, errors.New("resident profile is verified for Codex " + PinnedCodex)
	}
	base := filepath.Join(filepath.Dir(root), "providers")
	if err = os.MkdirAll(base, 0700); err != nil {
		return nil, err
	}
	run, err := os.MkdirTemp(base, "actor-")
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
	userHome, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	auth, err := filepath.EvalSymlinks(filepath.Join(userHome, ".codex", "auth.json"))
	if err != nil {
		return nil, err
	}
	if err = os.Symlink(auth, filepath.Join(home, "auth.json")); err != nil {
		return nil, err
	}
	initial := "cli_auth_credentials_store = \"file\"\nweb_search = \"disabled\"\nsandbox_mode = \"read-only\"\napproval_policy = \"never\"\n"
	if err = os.WriteFile(filepath.Join(home, "config.toml"), []byte(initial), 0600); err != nil {
		return nil, err
	}
	cacheOut, err := exec.Command("/usr/bin/getconf", "DARWIN_USER_CACHE_DIR").Output()
	if err != nil {
		return nil, err
	}
	cache, err := filepath.EvalSymlinks(strings.TrimSpace(string(cacheOut)))
	if err != nil {
		return nil, err
	}
	journal, ready := filepath.Join(run, "proxy.jsonl"), filepath.Join(run, "proxy.port")
	log, err := os.OpenFile(filepath.Join(run, "proxy.stderr"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer log.Close()
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	args := []string{s.ProxyPath, "--journal", journal, "--ready", ready, "--watch-fd", "3"}
	for _, host := range s.ProxyHosts {
		args = append(args, "--allow", host)
	}
	proxy := exec.Command("/usr/bin/python3", args...)
	proxy.Stdout, proxy.Stderr = log, log
	proxy.ExtraFiles = []*os.File{read}
	if err = proxy.Start(); err != nil {
		read.Close()
		write.Close()
		return nil, err
	}
	read.Close()
	p = &preparedResident{work: work, watch: write, proxy: proxy}
	cleanup := p
	defer func() {
		if err != nil {
			cleanup.abort()
		}
	}()
	deadline := time.Now().Add(10 * time.Second)
	var port int
	for time.Now().Before(deadline) {
		if b, e := os.ReadFile(ready); e == nil {
			port, e = strconv.Atoi(strings.TrimSpace(string(b)))
			if e == nil && port > 0 && port < 65536 {
				break
			}
		}
		if e := proxy.Process.Signal(syscall.Signal(0)); e != nil {
			return nil, fmt.Errorf("egress proxy exited: %w", e)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if port == 0 {
		return nil, errors.New("egress proxy did not bind")
	}
	p.profile, err = horizon.RenderResident(x.Horizon, x.Floor, horizon.ResidentPaths{Home: home, Scratch: tmp, Work: work, Auth: auth, Cache: cache, ProxyPort: port})
	if err != nil {
		return nil, err
	}
	p.env = []string{"PATH=/usr/bin:/bin", "HOME=" + home, "CODEX_HOME=" + home, "TMPDIR=" + tmp, "HTTPS_PROXY=http://127.0.0.1:" + strconv.Itoa(port)}
	if err = os.WriteFile(filepath.Join(run, "actor.sb"), []byte(p.profile), 0600); err != nil {
		return nil, err
	}
	if err = bootstrapCatalog(s, p, run); err != nil {
		return nil, err
	}
	return p, nil
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
		model["supports_search_tool"] = false
		model["include_skills_usage_instructions"] = false
		model["include_plugin_usage_instructions"] = false
		model["include_apps_usage_instructions"] = false
	}
	models, _ := json.Marshal(map[string]any{"models": metadata.Models})
	if err = os.WriteFile(filepath.Join(run, "models.json"), models, 0600); err != nil {
		return err
	}
	config := fmt.Sprintf("model_catalog_json = %q\n", filepath.Join(run, "models.json")) + "cli_auth_credentials_store = \"file\"\nweb_search = \"disabled\"\nsandbox_mode = \"read-only\"\napproval_policy = \"never\"\n[agents]\nenabled = false\n[tools.experimental_request_user_input]\nenabled = false\n[tools.update_plan]\nenabled = false\n[skills.bundled]\nenabled = false\n"
	return os.WriteFile(filepath.Join(run, "home", "config.toml"), []byte(config), 0600)
}
