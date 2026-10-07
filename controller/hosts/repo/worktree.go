package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Worktrees stay in the workspace's existing read/write boundary. Only a
// fresh directory may be created; removal, moving and pruning remain absent.
func (g repository) worktreePath(name string, create bool) (string, error) {
	parts := strings.Split(name, "/")
	if len(parts) != 2 || parts[0] != "worktrees" || parts[1] == "" || parts[1] == "." || parts[1] == ".." || strings.ContainsAny(parts[1], "\\\x00") || strings.HasPrefix(parts[1], ".") {
		return "", fmt.Errorf("use worktrees/<name> within this repository")
	}
	parent := filepath.Join(g.path, "worktrees")
	st, err := os.Lstat(parent)
	if os.IsNotExist(err) && create {
		err = os.Mkdir(parent, 0755)
		if err != nil {
			return "", err
		}
		st, err = os.Lstat(parent)
	}
	if err != nil {
		return "", err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return "", fmt.Errorf("worktrees must be a directory, not a symlink")
	}
	path := filepath.Join(parent, parts[1])
	st, err = os.Lstat(path)
	if create {
		if !os.IsNotExist(err) {
			return "", fmt.Errorf("worktree destination must not exist")
		}
	} else {
		if err != nil {
			return "", err
		}
		if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("worktree must be a directory, not a symlink")
		}
	}
	return path, nil
}

func (g repository) worktreeArgs(args []string) ([]string, error) {
	if len(args) < 2 || len(args) > 128 {
		return nil, fmt.Errorf("worktree supports add and list")
	}
	if args[1] == "list" {
		for _, a := range args[2:] {
			if !slices.Contains([]string{"--porcelain", "-z", "--verbose", "-v"}, a) {
				return nil, fmt.Errorf("unsupported worktree list option %q", a)
			}
		}
		return args, nil
	}
	if args[1] != "add" {
		return nil, fmt.Errorf("worktree supports only add and list; removal, move and prune are unavailable")
	}
	var options, positions []string
	branch, detach := false, false
	paths := false
	for i := 2; i < len(args); i++ {
		a := args[i]
		if len(a) > 16<<10 || strings.ContainsRune(a, 0) {
			return nil, fmt.Errorf("invalid argument")
		}
		if a == "--" && !paths {
			paths = true
			continue
		}
		if !paths && strings.HasPrefix(a, "-") {
			switch a {
			case "-b":
				if branch || i+1 >= len(args) {
					return nil, fmt.Errorf("worktree add accepts one -b with a new branch name")
				}
				i++
				if _, err := g.query("check-ref-format", "--branch", args[i]); err != nil || strings.HasPrefix(args[i], "-") || strings.Contains(args[i], "@{") {
					return nil, fmt.Errorf("invalid new branch name")
				}
				branch = true
				options = append(options, "-b", args[i])
			case "--detach", "-d":
				detach = true
				options = append(options, "--detach")
			default:
				return nil, fmt.Errorf("unsupported worktree add option %q", a)
			}
			continue
		}
		positions = append(positions, a)
	}
	if (branch && detach) || len(positions) < 1 || len(positions) > 2 {
		return nil, fmt.Errorf("use worktree add [-b branch | --detach] worktrees/<name> [commit-or-branch]")
	}
	// Validate the revision before creating any folders. Default to HEAD,
	// avoiding Git's implicit remote/branch guessing based on the folder name.
	ref := "HEAD"
	if len(positions) == 2 {
		ref = positions[1]
	}
	if strings.HasPrefix(ref, "-") || strings.ContainsAny(ref, "\x00\\") || strings.Contains(ref, "..") || filepath.IsAbs(ref) {
		return nil, fmt.Errorf("unsupported worktree revision")
	}
	if _, err := g.query("rev-parse", "--verify", ref+"^{commit}"); err != nil {
		return nil, err
	}
	path, err := g.worktreePath(positions[0], true)
	if err != nil {
		return nil, err
	}
	return append(append([]string{"worktree", "add"}, options...), "--", path, ref), nil
}

// A -C selector is a scoped checkout selection, never arbitrary Git -C argv.
// Require both registration and matching common metadata, not just a folder
// that happens to contain another repository.
func (g repository) selectWorktree(name string) (repository, error) {
	path, err := g.worktreePath(name, false)
	if err != nil {
		return g, err
	}
	listed, err := g.query("worktree", "list", "--porcelain", "-z")
	if err != nil {
		return g, err
	}
	if !slices.Contains(strings.Split(listed, "\x00"), "worktree "+path) {
		return g, fmt.Errorf("-C requires a registered worktree in this repository")
	}
	selected := g
	if selected.workspace == "" {
		selected.workspace = g.path
	}
	selected.path = path
	top, err := selected.query("rev-parse", "--show-toplevel")
	if err != nil || top != path {
		return g, fmt.Errorf("invalid worktree root")
	}
	common := func(repo repository) (string, error) {
		dir, err := repo.query("rev-parse", "--git-common-dir")
		if err != nil {
			return "", err
		}
		if !filepath.IsAbs(dir) {
			dir = filepath.Join(repo.path, dir)
		}
		return filepath.EvalSymlinks(dir)
	}
	want, err := common(g)
	if err != nil {
		return g, err
	}
	got, err := common(selected)
	if err != nil || got != want {
		return g, fmt.Errorf("worktree belongs to another repository")
	}
	return selected, nil
}
