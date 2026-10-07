// Command ideashost commits the operator-selected repository's ideas folder.
// The write procedure parks its observation here after writing. Git runs on
// the host side; sandboxed actors gain no Git execution or metadata grants.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"skyra/controller/hosts/journal"
	"strings"
	"syscall"
	"time"

	"skyra/controller/runtime/custody"
	"skyra/controller/runtime/park"
)

const ideasPath = ":(top,literal)ideas"

type committer struct{ repo, git, index, session string }

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
		go serveIdeasConnection(g, conn.(*net.UnixConn), skyra)
	}
}

func serveIdeasConnection(g committer, conn *net.UnixConn, skyra string) {
	f, err := park.Receive(conn)
	if err != nil {
		park.Answer(conn, err.Error())
		conn.Close()
		return
	}
	err = park.Answer(conn, "")
	conn.Close()
	if err != nil {
		f.Close()
		return
	}
	g.session = f.Root
	out := g.commit(f.Payload)
	event, _ := json.Marshal(map[string]any{"to": f.Stamp, "payload": out})
	if err := park.Resume(skyra, f, event, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "ideashost: recovery required:", err)
	}
}

// commit preserves the write observation and adds commit or commit_error.
// A refused write never triggers Git. A private index and an isolated ref
// prevent notes from changing the checked-out branch or user staging area.
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

	const branch = "skyra/ideas"
	ref := "refs/heads/" + branch
	r["branch"] = branch
	parent, _ := g.run("rev-parse", "--verify", ref)
	base := parent
	if base == "" {
		base, err = g.run("rev-parse", "--verify", "HEAD")
		if err != nil {
			return fail(err)
		}
	}
	index, err := os.CreateTemp(gitdir, "skyra-ideas-index-")
	if err != nil {
		return fail(err)
	}
	index.Close()
	os.Remove(index.Name())
	defer os.Remove(index.Name())
	g.index = index.Name()
	if _, err := g.run("read-tree", base); err != nil {
		return fail(err)
	}
	blob, err := g.runInput(bytes.NewReader(data), "hash-object", "-w", "--stdin")
	if err != nil {
		return fail(err)
	}
	mode := "100644"
	st, err := os.Stat(actual)
	if err != nil {
		return fail(err)
	}
	if st.Mode().Perm()&0111 != 0 {
		mode = "100755"
	}
	file := "ideas/" + filepath.ToSlash(rel)
	if _, err := g.run("update-index", "--add", "--cacheinfo", mode, blob, file); err != nil {
		return fail(err)
	}
	tree, err := g.run("write-tree")
	if err != nil {
		return fail(err)
	}
	oldTree, _ := g.run("rev-parse", base+"^{tree}")
	hash := parent
	if tree == oldTree && parent != "" {
		r["commit_unchanged"] = true
	} else {
		args := []string{"commit-tree", tree, "-p", base, "-m", "ideas: save agent write"}
		hash, err = g.run(args...)
		if err != nil {
			return fail(err)
		}
		previous := parent
		if previous == "" {
			previous = strings.Repeat("0", 40)
		}
		if _, err := g.run("update-ref", ref, hash, previous); err != nil {
			return fail(err)
		}
	}
	r["commit"] = hash
	if g.session != "" {
		log, e := journal.Load(gitdir, g.session)
		if e != nil {
			return fail(e)
		}
		before, e := g.run("show", base+":"+file)
		existed := e == nil
		existsInLog := false
		for _, entry := range log.Files {
			if entry.Path == file {
				existsInLog = true
			}
		}
		log.Write(file, []byte(before), existed, data)
		if !existsInLog {
			for i := range log.Files {
				if log.Files[i].Path == file {
					if actualExists, ok := r["before_exists"].(bool); ok {
						log.Files[i].Existed = actualExists
						if !actualExists {
							log.Files[i].Before = ""
							log.Files[i].Content = ""
						}
					}
					if beforeHash, ok := r["before_sha256"].(string); ok && beforeHash != "" {
						log.Files[i].Before = beforeHash
					}
				}
			}
		}

		log.Commit(hash, branch)
		if e := journal.Save(gitdir, g.session, log); e != nil {
			return fail(e)
		}
	}

	return r
}

func digestIdeas(data []byte) string                   { s := sha256.Sum256(data); return hex.EncodeToString(s[:]) }
func (g committer) run(args ...string) (string, error) { return g.runInput(nil, args...) }
func (g committer) runInput(input io.Reader, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	base := []string{"-C", g.repo, "-c", "core.hooksPath=/dev/null", "-c", "commit.gpgsign=false", "-c", "core.fsmonitor=false"}
	cmd := exec.CommandContext(ctx, g.git, append(base, args...)...)
	cmd.Stdin = input
	// Do not inherit alternate repositories, indexes, or injected Git config.
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GIT_") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	if g.index != "" {
		cmd.Env = append(cmd.Env, "GIT_INDEX_FILE="+g.index)
	}
	if len(args) > 0 && args[0] == "commit-tree" {
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_NAME=Skyra agent", "GIT_AUTHOR_EMAIL=skyra@localhost")
	}
	var diagnostics bytes.Buffer
	cmd.Stderr = &diagnostics
	b, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(string(b)+diagnostics.String()), err)
	}
	if len(args) > 0 && args[0] == "show" {
		return string(b), nil
	}
	return strings.TrimSpace(string(b)), nil
}
