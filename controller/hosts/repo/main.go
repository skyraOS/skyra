// repohost owns the operator's repository edit and Git capability. Calls park
// and resume through ordinary custody; no sandboxed actor gains shell exec,
// Git metadata write access, credentials, or a file deletion operation.
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
)

type repository struct{ path, git, goRoot, moduleCache, python string }

func main() {
	custody.SealInherited()
	repo := flag.String("repo", "", "fixed repository root")
	sock := flag.String("sock", "", "private park socket")
	skyra := flag.String("skyra", "", "runtime binary")
	goRoot := flag.String("go-root", "", "operator-selected Go toolchain root")
	moduleCache := flag.String("module-cache", "", "read-only dependency cache")
	python := flag.String("python", "", "operator-selected Python binary")
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
	g := repository{path: root, git: "/usr/bin/git", goRoot: *goRoot, moduleCache: *moduleCache, python: *python}
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
	go func() { <-ctx.Done(); l.Close() }()
	for {
		conn, err := l.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		f, err := park.Receive(conn.(*net.UnixConn))
		if err != nil {
			park.Answer(conn.(*net.UnixConn), err.Error())
			conn.Close()
			continue
		}
		err = park.Answer(conn.(*net.UnixConn), "")
		conn.Close()
		if err != nil {
			f.Close()
			continue
		}
		out := g.handle(f.Payload)
		event, _ := json.Marshal(map[string]any{"to": f.Stamp, "payload": out})
		if err := park.Resume(skyra, f, event, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "repohost: recovery required:", err)
		}
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
	// One host serializes Git and writes across conversations. Other launcher
	// instances share the same advisory lock, outside Git's own index lock.
	dir, err := g.query("rev-parse", "--absolute-git-dir")
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	lock, err := os.OpenFile(filepath.Join(dir, "skyra-repo.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return map[string]string{"error": err.Error()}
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	if _, ok := keys["action"]; ok {
		var in checkInput
		if err := decode(b, &in); err != nil {
			return checkResult{ExitCode: -1, Error: "invalid_request: " + err.Error()}
		}
		return g.check(in)
	}
	if _, ok := keys["target"]; ok {
		var in execInput
		if err := decode(b, &in); err != nil {
			return execResult{Error: "invalid_request: " + err.Error()}
		}
		return g.execute(in)
	}
	var in writeInput
	if keys["path"] == nil || keys["text"] == nil {
		return writeResult{Error: "invalid_request: path and text are required"}
	}
	if err := decode(b, &in); err != nil {
		return writeResult{Error: "invalid_request: " + err.Error()}
	}
	return g.write(in)
}
