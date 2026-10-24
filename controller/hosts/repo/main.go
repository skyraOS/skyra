// repohost owns the operator-selected repository capability.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"skyra/controller/runtime/custody"
	"skyra/controller/runtime/park"
	"skyra/controller/runtime/tree"
)

type repository struct {
	path, git, goRoot, moduleCache, python, session, workspace string
	denyPush                                                   bool
	layout                                                     *checkLayout // what repo.check builds; nil is this repository's own
}

func main() {
	custody.SealInherited()
	repo := flag.String("repo", "", "fixed repository root")
	sock := flag.String("sock", "", "private park socket")
	skyra := flag.String("skyra", "", "runtime binary")
	goRoot := flag.String("go-root", "", "operator-selected Go toolchain root")
	moduleCache := flag.String("module-cache", "", "read-only dependency cache")
	python := flag.String("python", "", "operator-selected Python binary")
	denyPush := flag.Bool("deny-push", false, "disable publishing through Git push")
	layoutFile := flag.String("check-layout", "", "operator-provisioned repo.check layout (JSON); default is this repository's")
	flag.Parse()
	if *repo == "" || *sock == "" || *skyra == "" {
		fmt.Fprintln(os.Stderr, "usage: repohost -repo DIR -sock PATH -skyra BIN")
		os.Exit(2)
	}
	root, err := filepath.EvalSymlinks(*repo)
	if err == nil {
		root, err = filepath.Abs(root)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "repohost:", err)
		os.Exit(1)
	}
	layout, err := loadLayout(*layoutFile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "repohost:", err)
		os.Exit(1)
	}
	g := repository{denyPush: *denyPush, path: root, git: "/usr/bin/git", goRoot: *goRoot, moduleCache: *moduleCache, python: *python, layout: &layout}
	if g.python != "" {
		g.python, err = filepath.EvalSymlinks(g.python)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
	if top, err := g.query("rev-parse", "--show-toplevel"); err != nil || top != root {
		fmt.Fprintln(os.Stderr, "repohost: repo must be the repository root")
		os.Exit(1)
	}
	if err := serve(g, *sock, *skyra); err != nil {
		fmt.Fprintln(os.Stderr, "repohost:", err)
		os.Exit(1)
	}
}

func serve(g repository, socket, skyra string) error {
	l, err := net.Listen("unix", socket)
	if err != nil {
		return err
	}
	defer l.Close()
	if err := os.Chmod(socket, 0600); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return serveRepository(ctx, g, l, skyra)
}

func serveRepository(ctx context.Context, g repository, l net.Listener, skyra string) error {
	go func() { <-ctx.Done(); l.Close() }()
	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		// Taking custody must not wait on another call's Git/build operation
		// or runtime continuation. Senders have only 30 seconds for the ack.
		go g.serveConnection(conn.(*net.UnixConn), skyra)
	}
}

func (g repository) serveConnection(conn *net.UnixConn, skyra string) {
	defer conn.Close()
	f, err := park.Receive(conn)
	if err != nil {
		park.Answer(conn, err.Error())
		return
	}
	err = park.Answer(conn, "")
	conn.Close()
	if err != nil {
		f.Close()
		return
	}
	g.session = f.Root
	var out any
	if record, e := tree.Load(f.Root, f.Stamp); e == nil && record.To == "find" {
		out = g.find(f.Payload)
	} else {
		out = g.handle(f.Payload)
	}
	event, _ := json.Marshal(map[string]any{"to": f.Stamp, "payload": out})
	if err := park.Resume(skyra, f, event, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "repohost: recovery required:", err)
	}
}

func decode(b json.RawMessage, value any) error {
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err := d.Decode(value); err != nil {
		return err
	}
	if d.Decode(new(any)) != io.EOF {
		return fmt.Errorf("trailing JSON")
	}
	return nil
}

func (g repository) handle(b json.RawMessage) any {
	var keys map[string]json.RawMessage
	if json.Unmarshal(b, &keys) != nil {
		return map[string]string{"error": "invalid_request"}
	}
	// Validate before acquiring custody resources: malformed requests must
	// return the same typed refusal even when Git or its lock is unavailable.
	var operation func() any
	var failure func(string) any
	if len(keys) == 0 || keys["changes"] != nil {
		var in struct {
			Changes bool `json:"changes"`
		}
		if e := decode(b, &in); e != nil {
			return map[string]string{"error": "invalid_request: " + e.Error()}
		}
		lock, e := g.lock()
		if e != nil {
			return map[string]string{"error": e.Error()}
		}
		defer lock.Close()
		defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
		return g.roots(in.Changes)
	}
	if _, ok := keys["action"]; ok {
		var in checkInput
		failure = func(s string) any { return checkResult{ExitCode: -1, Error: s} }
		if err := decode(b, &in); err != nil {
			return failure("invalid_request: " + err.Error())
		}
		// Checks lock only while copying source, not while building or testing.
		return g.check(in)
	} else if _, ok := keys["target"]; ok {
		var in execInput
		failure = func(s string) any { return execResult{ExitCode: -1, Error: s} }
		if err := decode(b, &in); err != nil {
			return failure("invalid_request: " + err.Error())
		}
		operation = func() any { return g.execute(in) }
	} else {
		var in writeInput
		failure = func(s string) any { return writeResult{Error: s} }
		if keys["path"] == nil && keys["revert"] == nil || (keys["text"] == nil && keys["old"] == nil && keys["edits"] == nil && keys["revert"] == nil) {
			return failure("invalid_request: path and text or old are required")
		}
		if err := decode(b, &in); err != nil {
			return failure("invalid_request: " + err.Error())
		}
		operation = func() any { return g.write(in) }
	}
	lock, err := g.lock()
	if err != nil {
		return failure(err.Error())
	}
	defer lock.Close()
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	return operation()
}

func (g repository) lock() (*os.File, error) {
	dir, err := g.commonDir()
	if err != nil {
		return nil, err
	}
	lock, err := os.OpenFile(filepath.Join(dir, "skyra-repo.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		lock.Close()
		return nil, err
	}
	return lock, nil
}
