// Package launch is Claude Code's part of a provider resident's launch
// (proposal 47): the version its profile is verified for, its environment,
// its settings, and the loopback port claude-start serves its tools on. The
// runtime's launch does the rest.
package launch

import (
	"os"
	"path/filepath"
	"strconv"

	"skyra/controller/runtime/launch"
)

// Pinned is the one Claude Code version the provider profile is verified for
// (proposal 47). Changing it means rerunning a turn under the rendered
// profile and reading its sandbox denials first.
const Pinned = "2.1.310"

// Provider is Claude Code, as the runtime's launch registers it.
type Provider struct{}

func (Provider) Pinned() string { return Pinned }

func (Provider) Configure(r *launch.Run) error {
	var err error
	if r.Paths.BridgePort, err = launch.FreePort(); err != nil {
		return err
	}
	if r.Paths.Config, err = settings(r.Paths.Home); err != nil {
		return err
	}
	user := os.Getenv("USER")
	if user == "" {
		user = filepath.Base(os.Getenv("HOME"))
	}
	home, tmp := r.Paths.Home, r.Paths.Scratch
	r.Env = []string{"PATH=" + filepath.Dir(r.Paths.Helper) + ":/usr/bin:/bin", "HOME=" + home, "TMPDIR=" + tmp, "CLAUDE_CODE_TMPDIR=" + tmp, "USER=" + user,
		"HTTPS_PROXY=" + r.ProxyURL, "NO_PROXY=127.0.0.1,localhost", "DISABLE_AUTOUPDATER=1",
		"SKYRA_CLAUDE_PROGRAM=" + r.Surface.Provider.Program, "SKYRA_CLAUDE_BRIDGE_PORT=" + strconv.Itoa(r.Paths.BridgePort), "SKYRA_KEYCHAIN_SOCKET=" + r.Paths.Broker}
	return nil
}

func (Provider) Ready(*launch.Run) error { return nil }

// settings gives the resident the user's default Claude Code profile: a copy
// of ~/.claude/settings.json in its private home, which it reads and cannot
// change. Hooks, plugins and MCP stay off by the CLI's own flags.
func settings(home string) (string, error) {
	b := []byte("{}\n")
	if userHome, err := os.UserHomeDir(); err == nil {
		if s, err := os.ReadFile(filepath.Join(userHome, ".claude", "settings.json")); err == nil {
			b = s
		}
	}
	dir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "settings.json")
	return path, os.WriteFile(path, b, 0600)
}
