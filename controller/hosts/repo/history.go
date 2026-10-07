package main

import (
	"fmt"
	"path/filepath"
	"skyra/controller/hosts/journal"
)

func (g repository) commonDir() (string, error) {
	dir, e := g.query("rev-parse", "--git-common-dir")
	if e != nil {
		return "", e
	}
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(g.path, dir)
	}
	return filepath.EvalSymlinks(dir)
}
func (g repository) recordWrite(path string, before []byte, exists bool, after []byte) error {
	if g.session == "" {
		return nil
	}
	dir, e := g.commonDir()
	if e != nil {
		return e
	}
	l, e := journal.Load(dir, g.session)
	if e != nil {
		return e
	}
	if g.workspace != "" {
		rel, e := filepath.Rel(g.workspace, filepath.Join(g.path, path))
		if e != nil || !filepath.IsLocal(rel) {
			return fmt.Errorf("history path leaves workspace")
		}
		path = filepath.ToSlash(rel)
	}
	l.Write(path, before, exists, after)
	return journal.Save(dir, g.session, l)
}
func (g repository) original(path string) (journal.Entry, error) {
	if g.session == "" {
		return journal.Entry{}, fmt.Errorf("revert_unavailable: no recorded conversation identity")
	}
	dir, e := g.commonDir()
	if e != nil {
		return journal.Entry{}, e
	}
	l, e := journal.Load(dir, g.session)
	if e != nil {
		return journal.Entry{}, e
	}
	for _, entry := range l.Files {
		if entry.Path == path {
			if entry.Existed && digest([]byte(entry.Content)) != entry.Before {
				return journal.Entry{}, fmt.Errorf("revert_refused: original content was not recorded by repo.write")
			}
			return entry, nil
		}
	}
	return journal.Entry{}, fmt.Errorf("revert_refused: only files written by this conversation can be restored")
}
func (g repository) recordCommit(hash, branch string) error {
	if g.session == "" {
		return nil
	}
	dir, e := g.commonDir()
	if e != nil {
		return e
	}
	l, e := journal.Load(dir, g.session)
	if e != nil {
		return e
	}
	l.Commit(hash, branch)
	return journal.Save(dir, g.session, l)
}
func (g repository) roots(changes bool) any {
	r := map[string]any{"workspace": g.path, "read": []string{g.path}, "scratch": filepath.Join(g.path, "ideas"), "permissions": map[string]any{"repo_write": true, "scratch_write": true, "git_commit_approval": false, "git_merge_approval": false, "git_push_allowed": !g.denyPush, "scratch_commit_ref": "skyra/ideas", "worktrees": "worktrees/<name>", "shell": false, "delete": false}}
	if changes {
		dir, e := g.commonDir()
		if e != nil {
			r["error"] = e.Error()
			return r
		}
		l, e := journal.Load(dir, g.session)
		if e != nil {
			r["error"] = e.Error()
			return r
		}
		r["changes"], r["commits"] = l.Summary(), l.Commits
	}
	return r
}
