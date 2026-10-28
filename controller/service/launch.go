package service

import (
	"bytes"
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
	"sync"
	"syscall"
	"time"

	"skyra/controller/core/record"
)

// Program is one long-lived program the service host may run, as its
// registration declares it. What it needs to run is declared here, not in
// code: its home, its egress, its login. A model CLI names a Provider, which
// writes the CLI's own config; a program without one gets only its home.
type Program struct {
	Path string   `json:"path"` // what runs: the CLI itself, or a wrapper that execs it
	Args []string `json:"args,omitempty"`
	// Provider is a registered Provider's name, or "".
	Provider string `json:"provider,omitempty"`
	// CLI is the model CLI. When it is Path, the program is the CLI, run
	// from where it was installed (its package is beside it). Otherwise the
	// program is a wrapper, run from its verified copy, that execs CLI.
	CLI string `json:"cli,omitempty"`
	// Reads are files of the CLI's own package it reads at startup.
	Reads []string `json:"reads,omitempty"`
	// Auth is how the user's existing login is reached. "file": AuthPath,
	// linked read-only into the home. "keychain": the one generic-password
	// item Keychain, served from outside the sandbox by Helper, whose copy
	// inside the sandbox is the only `security` there.
	Auth     string `json:"auth,omitempty"`
	AuthPath string `json:"auth_path,omitempty"`
	Keychain string `json:"keychain,omitempty"`
	Helper   string `json:"helper,omitempty"`
	// ProxyHosts are the exact HTTPS CONNECT names the program may reach,
	// through the egress proxy at Proxy.
	ProxyHosts []string `json:"proxy_hosts,omitempty"`
	Proxy      string   `json:"proxy,omitempty"`

	// Pinned when the host loads the registration: a changed file is
	// refused at open.
	sha, cliSHA, helperSHA, proxySHA string
}

// Provider is one model CLI's part of a program's launch.
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

// Run is one session's launch, as its Provider sees it.
type Run struct {
	Exe, CLI    string
	Args        []string
	Dir         string // this run's private directory
	Paths       Paths  // the names the profile is rendered with
	ProxyURL    string
	Env         []string
	ProfileFile string
}

// Paths are names the host allocates for one session. They are never
// supplied by a client.
type Paths struct {
	Home, Scratch, Work, Cache string
	Config                     string // the host-written config in Home: readable, never writable
	Catalog                    string // a host-written file beside Home the CLI reads, or ""
	Auth                       string // file auth: the resolved credential file
	Helper                     string // keychain auth: the verified copy of the helper, named security
	Broker                     string // keychain auth: the broker's socket
	ProxyPort                  int
	BridgePort                 int // a loopback port the program serves its own tools on, or 0
}

var (
	mu        sync.RWMutex
	providers = map[string]Provider{}
)

// Register makes a provider known by the name a program's registration
// gives it.
func Register(name string, p Provider) {
	mu.Lock()
	defer mu.Unlock()
	providers[name] = p
}

func lookup(name string) Provider {
	mu.RLock()
	defer mu.RUnlock()
	return providers[name]
}

// Load reads the programs the host may run, resolves their paths and pins
// every file by content hash. Relative paths are relative to the file.
func Load(file string) (map[string]*Program, error) {
	var config struct {
		Programs map[string]*Program `json:"programs"`
	}
	b, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&config); err != nil {
		return nil, fmt.Errorf("%s: %w", file, err)
	}
	dir := filepath.Dir(file)
	for name, p := range config.Programs {
		if name == "" || strings.ContainsAny(name, "/\x00") || p == nil {
			return nil, fmt.Errorf("invalid program %q", name)
		}
		if err := p.pin(dir); err != nil {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	return config.Programs, nil
}

func (p *Program) pin(dir string) error {
	pin := func(file string) (string, string, error) {
		if file == "" {
			return "", "", errors.New("missing file")
		}
		if !filepath.IsAbs(file) {
			file = filepath.Join(dir, file)
		}
		resolved, err := filepath.EvalSymlinks(file)
		if err != nil {
			return "", "", err
		}
		b, err := os.ReadFile(resolved)
		if err != nil {
			return "", "", err
		}
		return resolved, record.Sum(b), nil
	}
	var err error
	if p.Path, p.sha, err = pin(p.Path); err != nil {
		return err
	}
	if p.Provider == "" {
		if p.CLI != "" || len(p.Reads) != 0 || p.Auth != "" || p.AuthPath != "" || p.Keychain != "" || p.Helper != "" || len(p.ProxyHosts) != 0 || p.Proxy != "" {
			return errors.New("only a provider program declares a CLI, a login or egress")
		}
		return nil
	}
	if lookup(p.Provider) == nil {
		return fmt.Errorf("unknown provider %q", p.Provider)
	}
	if p.CLI, p.cliSHA, err = pin(p.CLI); err != nil {
		return fmt.Errorf("cli: %w", err)
	}
	for i, read := range p.Reads {
		if !filepath.IsAbs(read) {
			return fmt.Errorf("read %q is not absolute", read)
		}
		p.Reads[i] = filepath.Clean(read)
	}
	switch p.Auth {
	case "file":
		if !filepath.IsAbs(p.AuthPath) || p.Keychain != "" || p.Helper != "" {
			return errors.New("file auth takes an absolute auth_path and nothing else")
		}
		p.AuthPath = filepath.Clean(p.AuthPath)
	case "keychain":
		if p.Keychain == "" || strings.ContainsAny(p.Keychain, "\"\\\n\r\x00") || p.AuthPath != "" {
			return errors.New("keychain auth takes a plain service name and a helper")
		}
		if p.Helper, p.helperSHA, err = pin(p.Helper); err != nil {
			return fmt.Errorf("helper: %w", err)
		}
	default:
		return fmt.Errorf("unknown auth %q", p.Auth)
	}
	if len(p.ProxyHosts) == 0 {
		return errors.New("a provider program declares the hosts it may reach")
	}
	for _, host := range p.ProxyHosts {
		if host == "" || host != strings.ToLower(host) || strings.ContainsAny(host, "/:@ \n\r\x00") {
			return fmt.Errorf("invalid proxy host %q", host)
		}
	}
	if p.Proxy, p.proxySHA, err = pin(p.Proxy); err != nil {
		return fmt.Errorf("proxy: %w", err)
	}
	return nil
}

// installed reports whether a provider program is the CLI itself: it runs
// from where it was installed, checked but not copied, and its profile
// allows exec of that path alone.
func (p *Program) installed() bool { return p.Provider != "" && p.CLI == p.Path }

// launch is a prepared session, ready to start.
type launch struct {
	exe, profile, profileFile, work, log string
	args, env                            []string
	watch                                *os.File    // the write end of the helpers' watch pipe: the program's fd 3
	helpers                              []*exec.Cmd // the egress proxy and, for keychain auth, the keychain broker
}

// abort stops the helpers of a session that did not start.
func (l *launch) abort() {
	if l == nil {
		return
	}
	if l.watch != nil {
		l.watch.Close()
		l.watch = nil
	}
	for _, h := range l.helpers {
		if h.Process != nil {
			h.Process.Kill()
			h.Wait()
		}
	}
	l.helpers = nil
}

// started hands the helpers to the program: they hold the read end of the
// watch pipe, the program now holds its only write end, and they exit when
// it does.
func (l *launch) started() {
	if l.watch != nil {
		l.watch.Close()
		l.watch = nil
	}
	for _, h := range l.helpers {
		h.Process.Release()
	}
	l.helpers = nil
}

// prepare checks a program's pinned files, makes its run directory, starts
// its egress proxy and, for keychain auth, its keychain broker, has its
// provider write its config, and renders its profile.
func prepare(name string, p *Program, state string) (l *launch, err error) {
	program, err := os.ReadFile(p.Path)
	if err != nil {
		return nil, err
	}
	if record.Sum(program) != p.sha {
		return nil, errors.New("registered program changed")
	}
	exe := p.Path
	if !p.installed() {
		// The bytes checked are the bytes that run (43 D5): a copy named by
		// its hash, in a directory only the host writes.
		if exe, err = verifiedCopy(filepath.Join(state, "bin"), p.sha, program); err != nil {
			return nil, err
		}
	}
	base := filepath.Join(state, name)
	// Home and work outlive the session: a CLI keeps its sessions under
	// home, keyed by the working directory, and a later session resumes
	// them there. Scratch is per run.
	home, work := filepath.Join(base, "home"), filepath.Join(base, "work")
	for _, dir := range []string{home, work, filepath.Join(base, "runs")} {
		if err = os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	run, err := os.MkdirTemp(filepath.Join(base, "runs"), "run-")
	if err != nil {
		return nil, err
	}
	if run, err = filepath.EvalSymlinks(run); err != nil {
		return nil, err
	}
	tmp := filepath.Join(run, "tmp")
	if err = os.Mkdir(tmp, 0700); err != nil {
		return nil, err
	}
	paths := Paths{Home: home, Scratch: tmp, Work: work}
	l = &launch{exe: exe, args: p.Args, work: work, log: filepath.Join(run, "stderr.log"), profileFile: filepath.Join(run, "actor.sb")}
	defer func() {
		if err != nil {
			l.abort()
		}
	}()
	r := &Run{Exe: exe, CLI: p.CLI, Args: p.Args, Dir: run, ProfileFile: l.profileFile}
	if p.Provider != "" {
		if err = l.helpersFor(name, p, run, &paths); err != nil {
			return nil, err
		}
		r.Paths, r.ProxyURL = paths, "http://127.0.0.1:"+strconv.Itoa(paths.ProxyPort)
		provider := lookup(p.Provider)
		if err = provider.Configure(r); err != nil {
			return nil, err
		}
		paths = r.Paths
	} else {
		r.Env = []string{"PATH=/usr/bin:/bin", "HOME=" + home, "TMPDIR=" + tmp}
	}
	l.env = r.Env
	if l.profile, err = render(p, exe, paths); err != nil {
		return nil, err
	}
	if err = os.WriteFile(l.profileFile, []byte(l.profile), 0600); err != nil {
		return nil, err
	}
	if p.Provider != "" {
		if err = lookup(p.Provider).Ready(r); err != nil {
			return nil, err
		}
	}
	return l, nil
}

// helpersFor checks a provider program's CLI and its version, reaches its
// login, and starts its egress proxy and, for keychain auth, its keychain
// broker.
func (l *launch) helpersFor(name string, p *Program, run string, paths *Paths) error {
	if !p.installed() {
		cli, err := os.ReadFile(p.CLI)
		if err != nil {
			return err
		}
		if record.Sum(cli) != p.cliSHA {
			return errors.New("registered CLI changed")
		}
	}
	pinned := lookup(p.Provider).Pinned()
	version, err := exec.Command(p.CLI, "--version").Output()
	if err != nil || pinned == "" || !slices.Contains(strings.Fields(string(version)), pinned) {
		found := strings.TrimSpace(string(version))
		if err != nil {
			found = err.Error()
		}
		return fmt.Errorf("%s's profile is verified for %s %s, found %q", name, p.Provider, pinned, found)
	}
	proxyBytes, err := os.ReadFile(p.Proxy)
	if err != nil {
		return err
	}
	if record.Sum(proxyBytes) != p.proxySHA {
		return errors.New("registered proxy changed")
	}
	cacheOut, err := exec.Command("/usr/bin/getconf", "DARWIN_USER_CACHE_DIR").Output()
	if err != nil {
		return err
	}
	if paths.Cache, err = filepath.EvalSymlinks(strings.TrimSpace(string(cacheOut))); err != nil {
		return err
	}
	switch p.Auth {
	case "file":
		if paths.Auth, err = filepath.EvalSymlinks(p.AuthPath); err != nil {
			return err
		}
		link := filepath.Join(paths.Home, filepath.Base(p.AuthPath))
		if err = os.Remove(link); err != nil && !os.IsNotExist(err) {
			return err
		}
		if err = os.Symlink(paths.Auth, link); err != nil {
			return err
		}
	case "keychain":
		// The helper's copy is the only `security` the sandbox can run: on
		// PATH, in a directory the program cannot write.
		helper, err := os.ReadFile(p.Helper)
		if err != nil {
			return err
		}
		if record.Sum(helper) != p.helperSHA {
			return errors.New("registered keychain helper changed")
		}
		paths.Helper = filepath.Join(run, "bin", "security")
		if err = os.Mkdir(filepath.Dir(paths.Helper), 0700); err != nil {
			return err
		}
		if err = os.WriteFile(paths.Helper, helper, 0500); err != nil {
			return err
		}
		// A socket path stays under 104 bytes on macOS, so the broker's is in
		// the user's private socket directory, not the run directory.
		dir, err := Dir()
		if err != nil {
			return err
		}
		paths.Broker = filepath.Join(dir, "kc-"+filepath.Base(run)+".sock")
	}
	read, write, err := os.Pipe()
	if err != nil {
		return err
	}
	defer read.Close()
	l.watch = write
	log, err := os.OpenFile(filepath.Join(run, "proxy.stderr"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer log.Close()
	journal, ready := filepath.Join(run, "proxy.jsonl"), filepath.Join(run, "proxy.port")
	args := []string{p.Proxy, "--journal", journal, "--ready", ready, "--watch-fd", "3"}
	for _, host := range p.ProxyHosts {
		args = append(args, "--allow", host)
	}
	proxy := exec.Command("/usr/bin/python3", args...)
	proxy.Stdout, proxy.Stderr = log, log
	proxy.ExtraFiles = []*os.File{read}
	if err = proxy.Start(); err != nil {
		return err
	}
	l.helpers = append(l.helpers, proxy)
	if p.Auth == "keychain" {
		broker := exec.Command(paths.Helper, "serve", "-socket", paths.Broker, "-service", p.Keychain, "-watch-fd", "3")
		broker.Args[0] = "keychain-broker" // as security, it would be the shim
		broker.Stdout, broker.Stderr = log, log
		broker.ExtraFiles = []*os.File{read}
		if err = broker.Start(); err != nil {
			return err
		}
		l.helpers = append(l.helpers, broker)
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
			return fmt.Errorf("egress proxy exited: %w", e)
		}
		time.Sleep(50 * time.Millisecond)
	}
	if paths.ProxyPort == 0 {
		return errors.New("egress proxy did not bind")
	}
	for paths.Broker != "" {
		if info, e := os.Stat(paths.Broker); e == nil && info.Mode()&os.ModeSocket != 0 {
			break
		}
		if time.Now().After(deadline) {
			return errors.New("keychain broker did not listen")
		}
		time.Sleep(20 * time.Millisecond)
	}
	return nil
}

// verifiedCopy publishes exactly the bytes checked, at dir/<sha256>, and
// returns that path. A copy already there is checked again, not trusted.
func verifiedCopy(dir, sum string, bin []byte) (string, error) {
	exe := filepath.Join(dir, sum)
	if have, err := os.ReadFile(exe); err == nil && record.Sum(have) == sum {
		return exe, nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	tmp, err := os.CreateTemp(dir, ".copy-")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmp.Name()) // a no-op once renamed
	_, err = tmp.Write(bin)
	if err == nil {
		err = tmp.Chmod(0o500)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", err
	}
	// Published whole or not at all: a reader never sees part of a copy.
	return exe, os.Rename(tmp.Name(), exe)
}

// FreePort is a loopback port nothing listens on now, for a program to serve
// its own tools on: the profile names that one port, not every local
// service.
func FreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
