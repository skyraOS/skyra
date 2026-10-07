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
	app := f.String("app", "", "approved macOS bundle ID")
	pid := f.Int("pid", 0, "selected running process")
	window := f.Int("window-id", 0, "selected window")
	socket := f.String("socket", "", "new bridge socket inside a private directory")
	driverSocket := f.String("driver-socket", "", "explicit bounded Cua daemon socket")
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
	if !filepath.IsAbs(*socket) || !filepath.IsAbs(*driverSocket) {
		return errors.New("absolute bridge and driver socket paths are required")
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
	if err := scope.Validate(); err != nil {
		return err
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	probe, stop := context.WithTimeout(ctx, 15*time.Second)
	c, err := observation.Dial(probe, *driverSocket)
	if err == nil {
		err = c.Verify(scope)
		c.Close()
	}
	stop()
	if err != nil {
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
	return serve(ctx, l, *driverSocket, scope)
}

func readManifest(path string, expected []byte) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := io.ReadAll(io.LimitReader(f, 4097))
	if err != nil {
		return nil, err
	}
	if len(b) > 4096 {
		return nil, errors.New("manifest exceeds 4 KiB")
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

func serve(ctx context.Context, l net.Listener, driverSocket string, scope observation.Scope) error {
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
			out := handle(requestCtx, conn, driverSocket, scope)
			json.NewEncoder(conn).Encode(out)
		}()
	}
}

func handle(ctx context.Context, input io.Reader, driverSocket string, scope observation.Scope) observation.Result {
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
		r.PID, r.WindowID = scope.PID, scope.WindowID
	}
	if r.PID != scope.PID || r.WindowID != scope.WindowID {
		return out.Failed("outside_scope", errors.New("request does not name the selected window"))
	}
	c, err := observation.Dial(ctx, driverSocket)
	if err != nil {
		return out.Failed("unavailable", err)
	}
	defer c.Close()
	stop := context.AfterFunc(ctx, func() { c.Close() })
	defer stop()
	return c.Observe(scope, r)
}
