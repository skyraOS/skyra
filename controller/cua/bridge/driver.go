package main

// The bounded Cua daemon this bridge observes through, and its replacement.
//
// Cua latches an idle-expired (30 min) or expired (8 h) manifest for the life
// of its process: every later call is refused. The bridge does not keep the
// daemon alive, which would keep an unused grant live; it replaces it on
// demand, as an idle Codex server is reclaimed and started again (proposal
// 38). The replacement loads the same approved manifest bytes, so it holds the
// same hash the registry pinned at genesis. A relaunch that fails backs off
// exponentially, and a request inside the backoff is refused at once.
//
// The bridge ends its daemon when it stops. A bridge killed outright cannot;
// session.py then ends any daemon still serving from its private directory.
// (Cua's stdin parent-liveness mode was tried and did not end a daemon outside
// its embedded mode, whose host trust the bridge does not take on.)

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"skyra/controller/cua/observation"
)

// daemon is one launched Cua daemon: the socket it serves and how to end it.
type daemon struct {
	Socket string
	PID    int
	stop   func()
	once   sync.Once
}

// Stop ends the daemon once; later calls do nothing.
func (d *daemon) Stop() {
	d.once.Do(func() {
		if d.stop != nil {
			d.stop()
		}
	})
}

// upstream owns the current daemon. Replacing it is serialized: a request
// that saw a stale daemon another request already replaced uses the new one.
type upstream struct {
	launch     func(ctx context.Context, gen int) (*daemon, error)
	scope      observation.Scope // the version and manifest hash every daemon must hold
	base       context.Context   // the bridge's lifetime; a launch outlives the request that asked for it
	start, max time.Duration     // relaunch backoff
	now        func() time.Time

	mu       sync.Mutex
	current  *daemon
	gen      int
	failures int
	retryAt  time.Time
	lastErr  error
}

// launchTimeout bounds one launch, from open to a verified daemon.
const launchTimeout = 20 * time.Second

func newUpstream(base context.Context, scope observation.Scope, launch func(context.Context, int) (*daemon, error)) *upstream {
	return &upstream{launch: launch, scope: scope, base: base, start: 2 * time.Second, max: 2 * time.Minute, now: time.Now}
}

// get is the live daemon, launching one if there is none.
func (u *upstream) get() (*daemon, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.current != nil {
		return u.current, nil
	}
	return u.relaunch()
}

// replace retires stale, if it is still the current daemon, and launches the
// next. If another request already replaced it, that daemon is returned.
func (u *upstream) replace(stale *daemon) (*daemon, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.current != nil && u.current != stale {
		return u.current, nil
	}
	if u.current != nil {
		u.current.Stop()
		u.current = nil
	}
	return u.relaunch()
}

// relaunch starts and verifies the next daemon, under the backoff. Caller holds mu.
func (u *upstream) relaunch() (*daemon, error) {
	if wait := u.retryAt.Sub(u.now()); wait > 0 {
		return nil, fmt.Errorf("relaunch backing off %s after %d failed launches: %w", wait.Round(time.Second), u.failures, u.lastErr)
	}
	u.gen++
	ctx, cancel := context.WithTimeout(u.base, launchTimeout)
	defer cancel()
	d, err := u.launch(ctx, u.gen)
	if err == nil {
		if err = verify(ctx, d.Socket, u.scope); err != nil {
			d.Stop()
		}
	}
	if err != nil {
		u.failures++
		wait := u.start << (u.failures - 1)
		if wait <= 0 || wait > u.max {
			wait = u.max
		}
		u.retryAt, u.lastErr = u.now().Add(wait), err
		return nil, err
	}
	u.failures, u.retryAt, u.lastErr = 0, time.Time{}, nil
	u.current = d
	return d, nil
}

// Close ends the current daemon: the bridge is stopping.
func (u *upstream) Close() {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.current != nil {
		u.current.Stop()
		u.current = nil
	}
}

func verify(ctx context.Context, socket string, scope observation.Scope) error {
	c, err := observation.Dial(ctx, socket)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Verify(scope)
}

// stale reports whether a failed observation means the daemon can serve no
// more: its manifest lapsed, or it is gone. A slow capture is not stale.
func stale(r observation.Result) bool {
	if r.Error == nil {
		return false
	}
	if r.Error.Code == "unavailable" {
		return true // the dial failed
	}
	for _, s := range []string{
		"capability manifest idle timeout exceeded",
		"capability manifest expired",
		"daemon does not hold the expected approved bounded manifest",
		"EOF", "broken pipe", "connection reset", "connection refused",
	} {
		if strings.Contains(r.Error.Message, s) {
			return true
		}
	}
	return false
}

// observe runs one observation on the current daemon. If the daemon is stale
// it is replaced and the observation retried once: it is read-only.
func observe(ctx context.Context, up *upstream, scope observation.Scope, r observation.Request) observation.Result {
	out := observation.NewResult(scope)
	d, err := up.get()
	if err != nil {
		return out.Failed("unavailable", fmt.Errorf("desktop driver unavailable: %w", err))
	}
	res := observeOn(ctx, d.Socket, scope, r)
	if !stale(res) {
		return res
	}
	fmt.Fprintf(os.Stderr, "cua-bridge: daemon pid=%d cannot serve (%s); replacing it\n", d.PID, res.Error.Message)
	if d, err = up.replace(d); err != nil {
		return out.Failed("unavailable", fmt.Errorf("desktop driver unavailable: %w", err))
	}
	return observeOn(ctx, d.Socket, scope, r)
}

func observeOn(ctx context.Context, socket string, scope observation.Scope, r observation.Request) observation.Result {
	c, err := observation.Dial(ctx, socket)
	if err != nil {
		return observation.NewResult(scope).Failed("unavailable", err)
	}
	defer c.Close()
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer stop()
	return c.Observe(scope, r)
}

// cua launches the installed Cua app as a bounded daemon in the bridge's
// private directory, under the approved manifest.
type cua struct {
	App, Manifest, Dir string
}

func (c cua) binary() string { return filepath.Join(c.App, "Contents/MacOS/cua-driver") }

func (c cua) Launch(ctx context.Context, gen int) (*daemon, error) {
	name := func(ext string) string { return filepath.Join(c.Dir, fmt.Sprintf("driver-%d.%s", gen, ext)) }
	socket, pidfile, logfile := name("sock"), name("pid"), name("log")
	cleanup := func() {
		os.Remove(socket)
		os.Remove(pidfile)
	}
	open := exec.CommandContext(ctx, "open", "-n", "-g", "-a", c.App,
		"--stdout", logfile, "--stderr", logfile, "--env", "CUA_TELEMETRY_ENABLED=false",
		"--args", "serve", "--socket", socket, "--pid-file", pidfile, "--permission-mode", "bounded",
		"--capability-manifest", c.Manifest, "--approve-capability-manifest", "--no-overlay")
	if out, err := open.CombinedOutput(); err != nil {
		cleanup()
		return nil, fmt.Errorf("open Cua: %v: %s", err, strings.TrimSpace(string(out)))
	}
	if err := waitDaemonFiles(ctx, socket, pidfile); err != nil {
		// It may still start; session.py's sweep ends it if it does.
		cleanup()
		return nil, err
	}
	pid, err := c.identify(ctx, socket, pidfile)
	if err != nil {
		cleanup()
		return nil, err
	}
	d := &daemon{Socket: socket, PID: pid}
	d.stop = func() {
		c.shutdown(socket, pid)
		cleanup()
	}
	fmt.Fprintf(os.Stderr, "cua-bridge: daemon %d pid=%d socket=%s\n", gen, pid, socket)
	return d, nil
}

// Cua publishes its listening socket before its PID file. Socket existence
// alone does not mean startup is complete. Identity verification still runs
// afterwards and must match the process serving the socket.
func waitDaemonFiles(ctx context.Context, socket, pidfile string) error {
	tick := time.NewTicker(100 * time.Millisecond)
	defer tick.Stop()
	for {
		_, socketErr := os.Stat(socket)
		pid, pidErr := os.ReadFile(pidfile)
		if socketErr == nil && pidErr == nil && len(strings.TrimSpace(string(pid))) > 0 {
			return nil
		}
		for _, err := range []error{socketErr, pidErr} {
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("Cua startup files: %w", err)
			}
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("Cua did not finish startup; check its macOS permission dialog: %w", ctx.Err())
		case <-tick.C:
		}
	}
}

// identify checks that the socket's daemon is the process its pid file names,
// running the installed executable.
func (c cua) identify(ctx context.Context, socket, pidfile string) (int, error) {
	b, err := os.ReadFile(pidfile)
	if err != nil {
		return 0, fmt.Errorf("daemon pid file: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(b)))
	if err != nil || pid <= 0 {
		return 0, errors.New("daemon pid file is not a pid")
	}
	reported, err := metadataPID(ctx, socket)
	if err != nil {
		return 0, err
	}
	if reported != pid || !c.runs(pid) {
		return 0, errors.New("socket peer does not identify the installed Cua app")
	}
	return pid, nil
}

func metadataPID(ctx context.Context, socket string) (int, error) {
	conn, err := (&net.Dialer{}).DialContext(ctx, "unix", socket)
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(3 * time.Second))
	if err := json.NewEncoder(conn).Encode(map[string]string{"method": "metadata"}); err != nil {
		return 0, err
	}
	line, err := observation.ReadLine(conn, observation.MaxResponse)
	if err != nil {
		return 0, err
	}
	var m struct {
		OK     bool `json:"ok"`
		Result struct {
			PID int `json:"pid"`
		} `json:"result"`
	}
	if json.Unmarshal(line, &m) != nil || !m.OK || m.Result.PID <= 0 {
		return 0, errors.New("daemon metadata does not name its process")
	}
	return m.Result.PID, nil
}

// runs reports whether pid is a live process of the installed executable, so
// a pid the system has reused is never signalled.
func (c cua) runs(pid int) bool {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=").Output()
	if err != nil {
		return false
	}
	got, err1 := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
	want, err2 := filepath.EvalSymlinks(c.binary())
	return err1 == nil && err2 == nil && got == want
}

// shutdown asks the daemon to stop, bound to its pid; then SIGTERM, then
// SIGKILL, each only while pid is still the installed executable.
func (c cua) shutdown(socket string, pid int) {
	if conn, err := net.DialTimeout("unix", socket, time.Second); err == nil {
		conn.SetDeadline(time.Now().Add(2 * time.Second))
		json.NewEncoder(conn).Encode(map[string]any{"method": "shutdown_if_pid", "args": map[string]int{"expected_pid": pid}})
		observation.ReadLine(conn, 4096)
		conn.Close()
	}
	for _, sig := range []syscall.Signal{0, syscall.SIGTERM, syscall.SIGKILL} {
		if sig != 0 && c.runs(pid) {
			syscall.Kill(pid, sig)
		}
		for deadline := time.Now().Add(2 * time.Second); time.Now().Before(deadline); time.Sleep(50 * time.Millisecond) {
			if syscall.Kill(pid, 0) != nil {
				return
			}
		}
	}
	fmt.Fprintf(os.Stderr, "cua-bridge: daemon pid=%d did not stop\n", pid)
}
