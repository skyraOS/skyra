// Command ideashost commits the operator-selected repository's ideas folder.
// The write procedure parks its observation here after writing. Git runs on
// the host side; sandboxed actors gain no Git execution or metadata grants.
package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"skyra/controller/runtime/custody"
	"skyra/controller/runtime/park"
)

const ideasPath = ":(top,literal)ideas"

type committer struct{ repo, git string }

func main() {
	custody.SealInherited()
	repo := flag.String("repo", "", "fixed repository whose ideas directory is committed")
	sock := flag.String("sock", "", "park socket")
	skyra := flag.String("skyra", "", "runtime binary")
	flag.Parse()
	if *repo == "" || *sock == "" || *skyra == "" {
		fmt.Fprintln(os.Stderr, "usage: ideashost -repo DIR -sock PATH -skyra BIN")
		os.Exit(2)
	}
	if err := serve(committer{repo: *repo, git: "/usr/bin/git"}, *sock, *skyra); err != nil {
		fmt.Fprintln(os.Stderr, "ideashost:", err)
		os.Exit(1)
	}
}

func serve(g committer, sock, skyra string) error {
	l, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	defer l.Close()
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
		out := g.commit(f.Payload)
		event, _ := json.Marshal(map[string]any{"to": f.Stamp, "payload": out})
		if err := park.Resume(skyra, f, event, os.Stderr); err != nil {
			fmt.Fprintln(os.Stderr, "ideashost: recovery required:", err)
		}
	}
}

// commit preserves the write observation and adds commit or commit_error.
// A refused write never triggers Git. Every Git mutation has a fixed ideas
// pathspec; --only prevents unrelated staged changes entering the commit.
func (g committer) commit(payload json.RawMessage) map[string]any {
	var request struct {
		Write map[string]any `json:"write"`
	}
	if err := json.Unmarshal(payload, &request); err != nil || request.Write == nil {
		return map[string]any{"commit_error": "invalid write observation"}
	}
	r := request.Write
	if why, _ := r["error"].(string); why != "" {
		return r
	}
	fail := func(err error) map[string]any { r["commit_error"] = err.Error(); return r }
	name, _ := r["path"].(string)
	want, _ := r["sha256"].(string)
	if !filepath.IsLocal(name) || len(want) != 64 || r["written"] == nil {
		return fail(errors.New("not a successful write observation"))
	}
	// Serialize across launcher instances, without taking Git's own index lock.
	gitdir, err := g.run("rev-parse", "--absolute-git-dir")
	if err != nil {
		return fail(err)
	}
	lock, err := os.OpenFile(filepath.Join(gitdir, "skyra-repo.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return fail(err)
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		return fail(err)
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	root, err := g.run("rev-parse", "--show-toplevel")
	if err != nil {
		return fail(err)
	}
	realRepo, err := filepath.EvalSymlinks(g.repo)
	if err != nil {
		return fail(err)
	}
	if root != realRepo {
		return fail(errors.New("configured directory is not the repository root"))
	}
	ideas := filepath.Join(root, "ideas")
	actual, err := filepath.EvalSymlinks(filepath.Join(ideas, name))
	if err != nil {
		return fail(err)
	}
	rel, err := filepath.Rel(ideas, actual)
	if err != nil || !filepath.IsLocal(rel) {
		return fail(errors.New("written file leaves ideas"))
	}
	data, err := os.ReadFile(actual)
	if err != nil {
		return fail(err)
	}
	sum := sha256.Sum256(data)
	if hex.EncodeToString(sum[:]) != want {
		return fail(errors.New("written file changed before commit"))
	}
	if _, err := g.run("add", "-A", "--", ideasPath); err != nil {
		return fail(err)
	}
	_, diffErr := g.run("diff", "--cached", "--quiet", "--", ideasPath)
	if diffErr == nil {
		r["commit_unchanged"] = true
		return r
	}
	var exit *exec.ExitError
	if !errors.As(diffErr, &exit) || exit.ExitCode() != 1 {
		return fail(diffErr)
	}
	if _, err := g.run("commit", "--only", "-m", "ideas: save agent write", "--", ideasPath); err != nil {
		return fail(err)
	}
	hash, err := g.run("rev-parse", "HEAD")
	if err != nil {
		return fail(err)
	}
	r["commit"] = hash
	return r
}

func (g committer) run(args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base := []string{"-C", g.repo, "-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false", "-c", "core.fsmonitor=false"}
	cmd := exec.CommandContext(ctx, g.git, append(base, args...)...)
	// Do not inherit alternate repositories, indexes, or injected Git config.
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GIT_") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	b, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(b)), err)
	}
	return strings.TrimSpace(string(b)), nil
}
