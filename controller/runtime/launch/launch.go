// Package launch prepares a provider resident (proposal 47): its private run
// directory, its egress proxy and, for keychain auth, its keychain broker,
// and its rendered profile. Every provider takes this one path. What one
// model CLI needs beyond it, its pinned version, its environment and config,
// and any step it takes once its profile is rendered, is a Provider its
// adapter registers. The runtime knows none of them by name.
package launch

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"skyra/controller/core/horizon"
	"skyra/controller/core/machine"
	"skyra/controller/core/record"
	"skyra/controller/runtime/hold"
)

// Provider is one model CLI's part of a resident's launch.
type Provider interface {
	// Pinned is the one version of the CLI the profile is verified for:
	// a field of its --version output.
	Pinned() string
	// Configure writes the CLI's config under the run's paths, names it in
	// r.Paths.Config (and r.Paths.Catalog, r.Paths.BridgePort if it uses
	// them), and sets r.Env. It runs before the profile is rendered.
	Configure(r *Run) error
	// Ready runs once the profile is rendered and written to r.ProfileFile.
	Ready(r *Run) error
}

// Run is one provider resident's launch, as its Provider sees it.
type Run struct {
	Surface     horizon.Surface
	Dir         string                // this run's private directory
	Paths       horizon.ResidentPaths // the names the profile is rendered with
	ProxyURL    string
	Env         []string
	ProfileFile string
}

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

// Register makes a provider known by the name a registry entry's provider
// block gives it.
func Register(name string, p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[name] = p
}

// Known reports whether a provider is registered under name.
func Known(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := providers[name]
	return ok
}

func lookup(name string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[name]
}

// Installed reports whether a provider is its own resident: it runs from
// where it was installed, checked but not copied (its package is beside
// it), and its pinned profile allows exec of that path alone.
func Installed(s horizon.Surface) bool {
	return s.Provider != nil && s.Provider.Program == s.Path
}

// Resident is a prepared provider resident, ready to start.
type Resident struct {
	Profile     string
	ProfileFile string
	Env         []string
	Work        string
	Watch       *os.File    // the write end of the helpers' watch pipe: the resident's fd 3
	helpers     []*exec.Cmd // the egress proxy and, for keychain auth, the keychain broker
}

// Abort stops the helpers of a resident that did not start.
func (p *Resident) Abort() {
	if p == nil {
		return
	}
	if p.Watch != nil {
		p.Watch.Close()
		p.Watch = nil
	}
	for _, h := range p.helpers {
		if h.Process != nil {
			h.Process.Kill()
			h.Wait()
		}
	}
	p.helpers = nil
}

// Started hands the helpers to the resident: they hold the read end of the
// watch pipe, the resident now holds its only write end, and they exit when
// it does.
func (p *Resident) Started() {
	if p == nil {
		return
	}
	if p.Watch != nil {
		p.Watch.Close()
		p.Watch = nil
	}
	for _, h := range p.helpers {
		h.Process.Release()
	}
	p.helpers = nil
}

// Prepare makes one provider resident's private run directory, starts its
// egress proxy and, for keychain auth, its keychain broker, has its provider
// write its config, and renders its profile (proposal 47).
func Prepare(x machine.StartResident, root string) (p *Resident, err error) {
	s := x.Run.Surface
	v := s.Provider
	if s.ResidentProfile != "darwin-provider-v1" || v == nil {
		return nil, errors.New("unknown resident profile")
	}
	provider := lookup(v.Name)
	if provider == nil {
		return nil, fmt.Errorf("unknown provider %q", v.Name)
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
	pinned := provider.Pinned()
	version, err := exec.Command(v.Program, "--version").Output()
	if err != nil || pinned == "" || !slices.Contains(strings.Fields(string(version)), pinned) {
		found := strings.TrimSpace(string(version))
		if err != nil {
			found = err.Error()
		}
		return nil, fmt.Errorf("resident profile is verified for %s %s, found %q", v.Name, pinned, found)
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
	// Home and work outlive the run: the CLI keeps its sessions under home,
	// keyed by the working directory, and the next start resumes them there.
	// Scratch stays per run.
	home, work, tmp := filepath.Join(base, v.Name, "home"), filepath.Join(base, v.Name, "work"), filepath.Join(run, "tmp")
	for _, dir := range []string{home, work} {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	if err = os.Mkdir(tmp, 0700); err != nil {
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
	paths := horizon.ResidentPaths{Home: home, Scratch: tmp, Work: work, Cache: cache}
	switch v.Auth {
	case "file":
		if paths.Auth, err = filepath.EvalSymlinks(v.AuthPath); err != nil {
			return nil, err
		}
		link := filepath.Join(home, filepath.Base(v.AuthPath))
		if err = os.Remove(link); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if err = os.Symlink(paths.Auth, link); err != nil {
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
	p = &Resident{Work: work, Watch: write}
	cleanup := p
	defer func() {
		read.Close()
		if err != nil {
			cleanup.Abort()
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
		broker.Args[0] = "keychain-broker" // as security, it would be the shim
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
	r := &Run{Surface: s, Dir: run, Paths: paths, ProxyURL: "http://127.0.0.1:" + strconv.Itoa(paths.ProxyPort), ProfileFile: filepath.Join(run, "actor.sb")}
	if err = provider.Configure(r); err != nil {
		return nil, err
	}
	p.Env = r.Env
	p.Profile, err = horizon.RenderResident(x.Horizon, x.Floor, r.Paths)
	if err != nil {
		return nil, err
	}
	p.ProfileFile = r.ProfileFile
	if err = os.WriteFile(p.ProfileFile, []byte(p.Profile), 0600); err != nil {
		return nil, err
	}
	if err = provider.Ready(r); err != nil {
		return nil, err
	}
	return p, nil
}

// FreePort is a loopback port nothing listens on now, for a resident to
// serve its own tools on: the profile names that one port, not every local
// service.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
