// Command cua-bridge serves observation requests for one operator-selected window.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	"skyra/controller/cua/observation"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cua-bridge:", err)
		os.Exit(1)
	}
}

func run() error {
	manifestOnly := len(os.Args) > 1 && os.Args[1] == "manifest"
	args := os.Args[1:]
	if manifestOnly {
		args = args[1:]
	}
	f := flag.NewFlagSet("cua-bridge", flag.ContinueOnError)
	app := f.String("app", "", "approved macOS bundle IDs, comma-separated")
	windowsFile := f.String("windows", "", "operator-selected window inventory JSON")
	pid := f.Int("pid", 0, "selected running process")
	window := f.Int("window-id", 0, "selected window")
	socket := f.String("socket", "", "new bridge socket inside a private directory")
	driverApp := f.String("driver-app", "/Applications/CuaDriver.app", "installed Cua app the bridge launches its daemons from")
	driverVersion := f.String("driver-version", "", "expected exact Cua version")
	manifestFile := f.String("manifest", "", "exact generated manifest loaded by Cua")
	if err := f.Parse(args); err != nil {
		return err
	}
	if f.NArg() != 0 {
		return errors.New("unexpected arguments")
	}
	expected, err := observation.Manifest(*app)
	if err != nil {
		return err
	}
	if manifestOnly {
		_, err := os.Stdout.Write(expected)
		return err
	}
	if !filepath.IsAbs(*socket) || !filepath.IsAbs(*driverApp) || !filepath.IsAbs(*manifestFile) {
		return errors.New("absolute bridge socket, driver app and manifest paths are required")
	}
	info, err := os.Stat(filepath.Dir(*socket))
	if err != nil {
		return err
	}
	if !info.IsDir() || info.Mode().Perm()&0o077 != 0 {
		return errors.New("bridge directory must be private (0700)")
	}
	manifest, err := readManifest(*manifestFile, expected)
	if err != nil {
		return err
	}
	scope := observation.Scope{DriverVersion: *driverVersion, ManifestHash: observation.ManifestHash(manifest), BundleID: *app, PID: *pid, WindowID: *window}
	scopes := []observation.Scope{scope}
	if *windowsFile != "" {
		scopes, err = readScopes(*windowsFile, scope)
		if err != nil {
			return err
		}
		scope = scopes[0]
	}
	for _, selected := range scopes {
		if err := selected.Validate(); err != nil {
			return err
		}
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	// The bridge launches its daemons in its own private directory, and ends
	// the current one when it stops; a daemon that outlives it reads end of
	// file on its liveness pipe and stops itself.
	up := newUpstream(ctx, scope, cua{App: *driverApp, Manifest: *manifestFile, Dir: filepath.Dir(*socket)}.Launch)
	defer up.Close()
	if _, err := up.get(); err != nil {
		return fmt.Errorf("bounded daemon verification: %w", err)
	}
	// Never unlink an existing endpoint: it may belong to a live bridge.
	l, err := net.Listen("unix", *socket)
	if err != nil {
		return err
	}
	defer l.Close()
	if err := os.Chmod(*socket, 0o600); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "cua-bridge: observation-only pid=%d window=%d manifest=%s socket=%s\n", scope.PID, scope.WindowID, scope.ManifestHash, *socket)
	return serve(ctx, l, up, scopes, *windowsFile != "")
}

func readManifest(path string, expected []byte) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 65537))
	if err != nil {
		return nil, err
	}
	if len(b) > 65536 {
		return nil, errors.New("manifest exceeds 64 KiB")
	}
	var actualValue, expectedValue any
	if err := json.Unmarshal(b, &actualValue); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(expected, &expectedValue); err != nil {
		return nil, err
	}
	a, _ := json.Marshal(actualValue)
	e, _ := json.Marshal(expectedValue)
	if !bytes.Equal(a, e) {
		return nil, errors.New("manifest differs from generated observation-only scope")
	}
	return b, nil
}

func serve(ctx context.Context, l net.Listener, up *upstream, scopes []observation.Scope, inventory bool) error {
	stop := context.AfterFunc(ctx, func() { l.Close() })
	defer stop()
	slots := make(chan struct{}, 4)
	var workers sync.WaitGroup
	defer workers.Wait()
	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		select {
		case slots <- struct{}{}:
		default:
			conn.Close()
			continue
		}
		workers.Add(1)
		go func() {
			defer workers.Done()
			defer func() { <-slots }()
			defer conn.Close()
			requestCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()
			stop := context.AfterFunc(requestCtx, func() { conn.Close() })
			defer stop()
			conn.SetDeadline(time.Now().Add(15 * time.Second))
			out := handleScopes(requestCtx, conn, up, scopes, inventory)
			json.NewEncoder(conn).Encode(out)
		}()
	}
}

func handle(ctx context.Context, input io.Reader, up *upstream, scope observation.Scope) observation.Result {
	return handleScopes(ctx, input, up, []observation.Scope{scope}, false)
}

func handleScopes(ctx context.Context, input io.Reader, up *upstream, scopes []observation.Scope, inventory bool) observation.Result {
	scope := scopes[0]
	out := observation.NewResult(scope)
	b, err := observation.ReadLine(input, observation.MaxRequest)
	if err != nil {
		return out.Failed("invalid_request", err)
	}
	r, err := observation.DecodeRequest(b)
	if err != nil {
		return out.Failed("invalid_request", err)
	}
	if r.PID == 0 && r.WindowID == 0 {
		if inventory {
			out.Windows = scopes
			out.Text = "Available windows for this app session. Call desktop_observe with a listed pid and window_id to read text and capture that window."
			return out
		}
		r.PID, r.WindowID = scope.PID, scope.WindowID
	}
	for _, selected := range scopes {
		if selected.PID == r.PID && selected.WindowID == r.WindowID {
			scope = selected
			out = observation.NewResult(scope)
			break
		}
	}
	if r.PID != scope.PID || r.WindowID != scope.WindowID {
		return out.Failed("outside_scope", errors.New("request does not name the selected window"))
	}
	return observe(ctx, up, scope, r)
}

func readScopes(path string, common observation.Scope) ([]observation.Scope, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 256*1024+1))
	if err != nil {
		return nil, err
	}
	if len(b) > 256*1024 {
		return nil, errors.New("window inventory exceeds limit")
	}
	var scopes []observation.Scope
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(&scopes); err != nil {
		return nil, err
	}
	if d.Decode(new(any)) != io.EOF {
		return nil, errors.New("trailing inventory data")
	}
	if len(scopes) == 0 || len(scopes) > 128 {
		return nil, errors.New("window inventory must have 1..128 entries")
	}
	ids := strings.Split(common.BundleID, ",")
	seen := map[[2]int]bool{}
	for i := range scopes {
		s := &scopes[i]
		if !slices.Contains(ids, s.BundleID) {
			return nil, errors.New("window app is outside manifest")
		}
		s.DriverVersion = common.DriverVersion
		s.ManifestHash = common.ManifestHash
		if err := s.Validate(); err != nil {
			return nil, err
		}
		key := [2]int{s.PID, s.WindowID}
		if seen[key] {
			return nil, errors.New("duplicate window identity")
		}
		seen[key] = true
	}
	return scopes, nil
}
