package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"skyra/controller/targets/find/search"
	"sort"
	"strings"
	"time"
)

func (g repository) changedFiles(since string) ([]string, error) {
	if since == "" {
		since = "HEAD"
	}
	if strings.HasPrefix(since, "-") || strings.ContainsAny(since, "\x00\\:") || strings.Contains(since, "..") || filepath.IsAbs(since) {
		return nil, fmt.Errorf("bad_since: use a commit or branch, without repository/path overrides")
	}
	oid, e := g.query("rev-parse", "--verify", since+"^{commit}")
	if e != nil {
		return nil, e
	}
	tracked, e := g.query("diff", "--no-ext-diff", "--no-textconv", "--name-only", "-z", oid, "--")
	if e != nil {
		return nil, e
	}
	untracked, e := g.query("ls-files", "--others", "--exclude-standard", "-z")
	if e != nil {
		return nil, e
	}
	set := map[string]bool{}
	for _, name := range append(strings.Split(tracked, "\x00"), strings.Split(untracked, "\x00")...) {
		if filepath.IsLocal(name) {
			set[name] = true
		}
	}
	out := []string{}
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out, nil
}
func (g repository) find(raw json.RawMessage) search.Result {
	var in search.Input
	failure := func(e error) search.Result {
		return search.Result{In: in.In, Matches: []search.Match{}, Error: e.Error()}
	}
	if e := decode(raw, &in); e != nil {
		return failure(e)
	}
	scratch := filepath.Join(g.path, "ideas")
	safe := g.path
	name := in.In
	if s, ok := strings.CutPrefix(name, "scratch:"); ok {
		safe = scratch
		name = s
	}
	if name == "" {
		name = "."
	}
	if filepath.IsAbs(name) {
		rel, e := filepath.Rel(safe, name)
		if e != nil || !filepath.IsLocal(rel) {
			return failure(fmt.Errorf("outside_repo: find is bound to this workspace and its scratch"))
		}
		name = rel
	}
	if !filepath.IsLocal(name) {
		return failure(fmt.Errorf("outside_repo: use a path within this workspace"))
	}
	root, e := os.OpenRoot(safe)
	if e != nil {
		return failure(e)
	}
	defer root.Close()
	partPath := ""
	for _, component := range strings.Split(filepath.ToSlash(name), "/") {
		if component == ".git" || component == ".skyra" {
			return failure(fmt.Errorf("protected_path: repository metadata and runtime state are unavailable to find"))
		}
		partPath = filepath.Join(partPath, component)
		st, err := root.Lstat(partPath)
		if err == nil && st.Mode()&os.ModeSymlink != 0 {
			return failure(fmt.Errorf("symlink: use a direct path within the workspace"))
		}
	}
	in.SafeRoot, in.SafePath = root, safe
	if in.Changed {
		selected := g
		prefix := ""
		parts := strings.Split(filepath.ToSlash(name), "/")
		if safe == g.path && len(parts) >= 2 && parts[0] == "worktrees" {
			prefix = filepath.ToSlash(filepath.Join(parts[0], parts[1]))
			selected, e = g.selectWorktree(prefix)
			if e != nil {
				return failure(e)
			}
		}
		files, e := selected.changedFiles(in.Since)
		if e != nil {
			return failure(e)
		}
		in.ChangedPaths = map[string]bool{}
		for _, f := range files {
			shown := filepath.ToSlash(filepath.Join(prefix, f))
			if safe == scratch {
				rel, e := filepath.Rel("ideas", shown)
				if e != nil || !filepath.IsLocal(rel) {
					continue
				}
				shown = "scratch:" + filepath.ToSlash(rel)
			}
			if filepath.IsAbs(in.In) {
				shown = filepath.Join(g.path, shown)
			}
			in.ChangedPaths[shown] = true
		}
	}
	return search.Find(g.path, scratch, in, time.Now())
}
