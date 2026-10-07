package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"
)

type execInput struct {
	Target string   `json:"target"`
	Args   []string `json:"args"`
}
type execResult struct {
	Target    string   `json:"target,omitempty"`
	Args      []string `json:"args,omitempty"`
	ExitCode  int      `json:"exit_code"`
	Output    string   `json:"output,omitempty"`
	Truncated bool     `json:"truncated,omitempty"`
	Error     string   `json:"error,omitempty"`
}

// value options consume the following argument; flag options do not. Unknown
// options are refused rather than letting a new Git feature bypass the policy.
var flags = map[string]string{
	"status":    "--short -s --branch -b --porcelain --porcelain=v1 --porcelain=v2 --untracked-files --ignored",
	"diff":      "--cached --staged --stat --numstat --shortstat --name-only --name-status --check --quiet --exit-code --patch -p --no-renames --no-ext-diff --no-textconv --binary --color=never",
	"log":       "--oneline --graph --all --decorate --stat --name-only --name-status --patch -p --no-merges --merges --reverse --no-ext-diff --no-textconv",
	"show":      "--stat --name-only --name-status --patch -p --no-ext-diff --no-textconv",
	"rev-parse": "--show-toplevel --absolute-git-dir --verify --short --symbolic-full-name --abbrev-ref --is-inside-work-tree",
	"ls-files":  "--stage -s --cached --modified --deleted --others --exclude-standard -z",
	"ls-tree":   "-r -t --name-only --long -l -z",
	"branch":    "--list --show-current --all -a --remotes -r --verbose -v --no-track",
	"tag":       "--list -l --annotate -a --no-sign",
	"add":       "--all -A --update -u --force -f --dry-run -n --verbose -v --intent-to-add -N",
	"commit":    "--all -a --amend --no-edit --allow-empty --allow-empty-message --only -o --include -i --signoff -s --reset-author --no-verify -n --dry-run --verbose -v --no-gpg-sign",
	"push":      "--force -f --force-with-lease --force-if-includes --tags --all --branches --follow-tags --set-upstream -u --dry-run -n --porcelain --atomic --no-verify --verbose -v --quiet -q",
	"fetch":     "--all --tags --no-tags --no-prune --no-prune-tags --verbose -v --quiet -q --dry-run",
	"merge":     "--no-ff --ff --ff-only --squash --no-squash --no-commit --commit --no-edit --no-gpg-sign --no-verify --stat --no-stat --quiet -q --verbose -v",
}
var values = map[string]string{
	"diff":   "--diff-filter --unified -U",
	"log":    "--max-count -n --format --pretty --since --until --author --grep",
	"show":   "--format --pretty",
	"tag":    "--message -m",
	"commit": "--message -m --author --date --cleanup --fixup --squash",
	"push":   "--force-with-lease --push-option -o",
	"merge":  "--message -m",
}

func parse(args []string) (positions []string, err error) {
	if len(args) == 0 || len(args) > 128 {
		return nil, fmt.Errorf("args must contain a Git subcommand and at most 128 arguments")
	}
	allowed, ok := flags[args[0]]
	if !ok {
		return nil, fmt.Errorf("Git subcommand %q is not allowed; deletion, checkout, reset, config, shell aliases and maintenance are unavailable", args[0])
	}
	for _, a := range args {
		if len(a) > 16<<10 || strings.ContainsRune(a, 0) {
			return nil, fmt.Errorf("invalid argument")
		}
	}
	paths := false
	for i := 1; i < len(args); i++ {
		a := args[i]
		if a == "--" && !paths {
			paths = true
			continue
		}
		if !paths && strings.HasPrefix(a, "-") {
			key, _, equal := strings.Cut(a, "=")
			if slices.Contains(strings.Fields(allowed), a) {
				continue
			}
			if slices.Contains(strings.Fields(values[args[0]]), key) {
				if !equal {
					i++
					if i >= len(args) {
						return nil, fmt.Errorf("%s needs a value", key)
					}
				}
				continue
			}
			return nil, fmt.Errorf("Git option %q is not allowed", a)
		}
		// File pathspecs and revision expressions cannot name an outside path.
		if strings.HasPrefix(a, "/") || strings.Contains(a, "..") || strings.Contains(a, "\\") {
			return nil, fmt.Errorf("outside_repo or unsupported revision: %q", a)
		}
		positions = append(positions, a)
	}
	return positions, nil
}

func (g repository) execute(in execInput) execResult {
	r := execResult{Target: in.Target, Args: in.Args, ExitCode: -1}
	fail := func(err error) execResult { r.Error = "refused: " + err.Error(); return r }
	if in.Target != "git" && in.Target != "/usr/bin/git" {
		return fail(fmt.Errorf("target must be git; no arbitrary executable or shell"))
	}
	positions, err := parse(in.Args)
	if err != nil {
		return fail(err)
	}
	args := slices.Clone(in.Args)
	// Configured external filters/drivers could turn add/merge into arbitrary
	// execution. Refuse such repositories; hooks and fsmonitor are disabled.
	if custom, err := g.query("config", "--get-regexp", `^(filter\..*\.(clean|smudge|process)|merge\..*\.driver)$`); err == nil && custom != "" {
		return fail(fmt.Errorf("custom Git filters or merge drivers must be removed by the operator"))
	} else if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 1 {
			return fail(err)
		}
	}
	switch args[0] {
	case "add", "commit":
		if err := g.noDeletedIndex(); err != nil {
			return fail(err)
		}
	case "branch":
		if len(positions) > 2 {
			return fail(fmt.Errorf("branch accepts at most a new name and a start point"))
		}
	case "tag":
		if len(positions) > 2 {
			return fail(fmt.Errorf("tag accepts at most a new name and an object"))
		}
	case "fetch":
		if len(positions) > 1 {
			return fail(fmt.Errorf("fetch accepts a configured remote, without ref overrides"))
		}
		if len(positions) == 1 {
			if err := g.remote(positions[0]); err != nil {
				return fail(err)
			}
		}
		if len(positions) == 0 {
			remotes, err := g.query("remote")
			if err != nil {
				return fail(err)
			}
			for _, name := range strings.Fields(remotes) {
				if err := g.remote(name); err != nil {
					return fail(err)
				}
			}
		}
		// Override even operator config: never prune refs or recurse submodules.
		args = append([]string{"fetch", "--no-prune", "--no-prune-tags", "--no-recurse-submodules"}, args[1:]...)
	case "merge":
		if len(positions) != 1 {
			return fail(fmt.Errorf("merge requires exactly one commit or branch"))
		}
		if err := g.noDeletedIndex(); err != nil {
			return fail(err)
		}
		// Pin the validated merge input so a moving ref cannot change its tree.
		oid, err := g.query("rev-parse", "--verify", positions[0]+"^{commit}")
		if err != nil {
			return fail(err)
		}
		preview, err := g.query("merge-tree", "--write-tree", "HEAD", oid)
		if err != nil {
			return fail(fmt.Errorf("merge preview failed or conflicts require manual resolution: %w", err))
		}
		if err := g.noRemovedPaths("HEAD", preview); err != nil {
			return fail(err)
		}
		// An option's value may equal the ref. Replace only the positional token.
		args = pinMerge(args, oid)
		args = append([]string{"merge", "--strategy=ort", "--no-edit", "--no-gpg-sign"}, args[1:]...)
	case "push":
		args, err = g.pushArgs(args, positions)
		if err != nil {
			return fail(err)
		}
	}
	if slices.Contains([]string{"diff", "show", "log"}, args[0]) {
		args = append([]string{args[0], "--no-ext-diff", "--no-textconv"}, args[1:]...)
	}
	out, truncated, err := g.run(args...)
	r.Output, r.Truncated, r.ExitCode = strings.TrimSpace(out), truncated, 0
	if err != nil {
		r.ExitCode = -1
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			r.ExitCode = exit.ExitCode()
		}
		r.Error = err.Error()
	}
	if err == nil && in.Args[0] == "push" && !slices.Contains(in.Args, "--dry-run") && !slices.Contains(in.Args, "-n") && (slices.Contains(in.Args, "-u") || slices.Contains(in.Args, "--set-upstream")) {
		if err := g.setUpstream(in.Args, positions); err != nil {
			r.Error = "push succeeded; setting upstream failed: " + err.Error()
		}
	}
	return r
}

func (g repository) setUpstream(args, positions []string) error {
	remote := "origin"
	if len(positions) > 0 {
		remote, positions = positions[0], positions[1:]
	}
	if slices.Contains(args, "--all") || slices.Contains(args, "--branches") {
		refs, err := g.query("for-each-ref", "--format=%(refname)", "refs/heads")
		if err != nil {
			return err
		}
		positions = strings.Fields(refs)
	}
	if len(positions) == 0 && !slices.Contains(args, "--tags") {
		branch, err := g.query("symbolic-ref", "--short", "HEAD")
		if err != nil {
			return err
		}
		dest := "refs/heads/" + branch
		if oldRemote, _ := g.query("config", "--get", "branch."+branch+".remote"); oldRemote == remote {
			if oldDest, _ := g.query("config", "--get", "branch."+branch+".merge"); oldDest != "" {
				dest = oldDest
			}
		}
		positions = []string{branch + ":" + dest}
	}
	for _, spec := range positions {
		source, dest, explicit := strings.Cut(strings.TrimPrefix(spec, "+"), ":")
		local, err := g.query("rev-parse", "--symbolic-full-name", source)
		if err != nil || !strings.HasPrefix(local, "refs/heads/") {
			continue
		}
		if !explicit {
			dest = local
		}
		if !strings.HasPrefix(dest, "refs/") {
			dest = "refs/heads/" + dest
		}
		branch := strings.TrimPrefix(local, "refs/heads/")
		for key, value := range map[string]string{"remote": remote, "merge": dest} {
			if _, err := g.query("config", "--local", "branch."+branch+"."+key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func pinMerge(args []string, oid string) []string {
	for i := 1; i < len(args); i++ {
		if args[i] == "--" {
			continue
		}
		if strings.HasPrefix(args[i], "-") {
			key, _, equal := strings.Cut(args[i], "=")
			if !equal && slices.Contains(strings.Fields(values["merge"]), key) {
				i++
			}
			continue
		}
		args[i] = oid
		break
	}
	return args
}

func (g repository) pushArgs(args, positions []string) ([]string, error) {
	remote := "origin"
	if len(positions) > 0 {
		remote, positions = positions[0], positions[1:]
	}
	if err := g.remote(remote); err != nil {
		return nil, err
	}
	url, err := g.query("remote", "get-url", "--push", "--all", remote)
	if err != nil {
		return nil, err
	}
	if strings.Contains(url, "\n") {
		return nil, fmt.Errorf("push requires one configured push URL so deletion checks cover its destination")
	}
	all, tags, follow := slices.Contains(args, "--all") || slices.Contains(args, "--branches"), slices.Contains(args, "--tags"), slices.Contains(args, "--follow-tags")
	if follow {
		return nil, fmt.Errorf("use --tags or explicit tags so every pushed ref is checked")
	}
	if all || tags {
		if len(positions) > 0 {
			return nil, fmt.Errorf("--all/--tags cannot be combined with explicit refspecs")
		}
		prefixes := []string{}
		if all {
			prefixes = append(prefixes, "refs/heads")
		}
		if tags {
			prefixes = append(prefixes, "refs/tags")
		}
		refs, err := g.query(append([]string{"for-each-ref", "--format=%(refname)"}, prefixes...)...)
		if err != nil {
			return nil, err
		}
		positions = strings.Fields(refs)
	}
	if len(positions) == 0 {
		branch, err := g.query("symbolic-ref", "--quiet", "--short", "HEAD")
		if err != nil {
			return nil, fmt.Errorf("push without a ref needs a checked-out branch: %w", err)
		}
		dest := branch
		if upstreamRemote, _ := g.query("config", "--get", "branch."+branch+".remote"); upstreamRemote == remote {
			if upstream, _ := g.query("config", "--get", "branch."+branch+".merge"); upstream != "" {
				dest = upstream
			}
		}
		positions = []string{branch + ":" + dest}
	}
	remoteRefs, err := g.query("ls-remote", url)
	if err != nil {
		return nil, err
	}
	old := map[string]string{}
	for _, line := range strings.Split(remoteRefs, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			old[fields[1]] = fields[0]
		}
	}
	options := []string{"push", "--no-mirror", "--no-prune", "--no-follow-tags", "--recurse-submodules=no"}
	requestedForce := slices.Contains(args, "--force") || slices.Contains(args, "-f")
	for _, a := range args {
		if strings.HasPrefix(a, "--force-with-lease") {
			requestedForce = true
		}
	}
	// Preserve approved options while replacing refspecs with checked OIDs.
	for i := 1; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			break
		}
		if !strings.HasPrefix(a, "-") {
			continue
		}
		if slices.Contains([]string{"--all", "--branches", "--tags", "--force", "-f"}, a) {
			continue
		}
		options = append(options, a)
		key, _, equal := strings.Cut(a, "=")
		if !equal && !slices.Contains(strings.Fields(flags["push"]), a) && slices.Contains(strings.Fields(values["push"]), key) {
			i++
			options = append(options, args[i])
		}
	}
	options = append(options, remote)
	for _, spec := range positions {
		force := strings.HasPrefix(spec, "+")
		spec = strings.TrimPrefix(spec, "+")
		source, dest, hasDest := strings.Cut(spec, ":")
		if source == "" || strings.ContainsAny(spec, "*?[^~ ") || (hasDest && dest == "") {
			return nil, fmt.Errorf("deletion or wildcard refspecs are not allowed")
		}
		if !hasDest {
			dest, err = g.query("rev-parse", "--symbolic-full-name", source)
			if err != nil {
				return nil, err
			}
		}
		if !strings.HasPrefix(dest, "refs/") {
			dest = "refs/heads/" + dest
		}
		if !strings.HasPrefix(dest, "refs/heads/") && !strings.HasPrefix(dest, "refs/tags/") {
			return nil, fmt.Errorf("push destinations must be branches or tags")
		}
		if _, err := g.query("check-ref-format", dest); err != nil {
			return nil, err
		}
		oid, err := g.query("rev-parse", "--verify", source+"^{object}")
		if err != nil {
			return nil, err
		}
		if previous := old[dest]; previous != "" {
			if _, err := g.query("cat-file", "-e", previous+"^{tree}"); err != nil {
				if _, err := g.query("fetch", "--no-tags", "--no-prune", "--no-prune-tags", "--no-write-fetch-head", "--no-recurse-submodules", url, previous); err != nil {
					return nil, fmt.Errorf("cannot verify remote tree: %w", err)
				}
			}
			if err := g.noRemovedPaths(previous, oid); err != nil {
				return nil, err
			}
			if !requestedForce && !force && previous != oid {
				if strings.HasPrefix(dest, "refs/tags/") {
					return nil, fmt.Errorf("existing tags require an explicit force option")
				}
				if _, err := g.query("merge-base", "--is-ancestor", previous, oid); err != nil {
					return nil, fmt.Errorf("non-fast-forward push requires an explicit force option")
				}
			}
			// Never update a destination that moved after its deletion preflight.
			options = append(options, "--force-with-lease="+dest+":"+previous)
		} else {
			options = append(options, "--force-with-lease="+dest+":")
		}
		options = append(options, oid+":"+dest)
	}
	return options, nil
}

func (g repository) noDeletedIndex() error {
	for _, args := range [][]string{{"ls-files", "--deleted", "-z"}, {"diff", "--cached", "--no-renames", "--diff-filter=D", "--name-only", "-z"}} {
		out, err := g.query(args...)
		if err != nil {
			return err
		}
		if out != "" {
			return fmt.Errorf("deletion is not allowed: tracked or staged paths are missing")
		}
	}
	return nil
}

func (g repository) noRemovedPaths(old, next string) error {
	out, err := g.query("diff", "--no-renames", "--no-ext-diff", "--no-textconv", "--diff-filter=D", "--name-only", "-z", old, next, "--")
	if err != nil {
		return err
	}
	if out != "" {
		return fmt.Errorf("deletion is not allowed: result would remove existing file paths")
	}
	return nil
}

func (g repository) remote(name string) error {
	if name == "" || strings.ContainsAny(name, "/: \\~") || strings.HasPrefix(name, "-") {
		return fmt.Errorf("use a configured remote name")
	}
	urls, err := g.query("remote", "get-url", "--all", name)
	if err != nil {
		return err
	}
	pushURLs, err := g.query("remote", "get-url", "--push", "--all", name)
	if err != nil {
		return err
	}
	for _, u := range strings.Split(urls+"\n"+pushURLs, "\n") {
		// ext:: runs commands; other custom remote helpers are also excluded.
		if strings.Contains(u, "::") {
			return fmt.Errorf("external remote helpers are not allowed")
		}
		if scheme, _, ok := strings.Cut(u, "://"); ok && !slices.Contains([]string{"https", "http", "ssh", "git", "file"}, scheme) {
			return fmt.Errorf("remote protocol is not allowed")
		}
	}
	return nil
}

type boundedOutput struct {
	data      []byte
	truncated bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	n := len(p)
	left := (64 << 10) - len(b.data)
	if len(p) > left {
		b.truncated = true
		p = p[:left]
	}
	b.data = append(b.data, p...)
	return n, nil
}
func (g repository) run(args ...string) (string, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	base := []string{"--no-pager", "-C", g.path, "-c", "core.hooksPath=/dev/null", "-c", "core.fsmonitor=false", "-c", "commit.gpgSign=false", "-c", "tag.gpgSign=false", "-c", "push.gpgSign=false", "-c", "gc.auto=0", "-c", "maintenance.auto=false", "-c", "protocol.ext.allow=never", "-c", "submodule.recurse=false", "-c", "merge.autoStash=false"}
	if len(args) > 0 && args[0] == "merge" {
		if branch, err := g.query("symbolic-ref", "--short", "HEAD"); err == nil {
			base = append(base, "-c", "branch."+branch+".mergeOptions=")
		}
	}
	if len(args) > 0 && args[0] == "push" {
		// Git reapplies a remote's mirror setting after parsing --no-mirror.
		// Override the config as well, and never use its configured push refs.
		for i := 1; i < len(args); i++ {
			a := args[i]
			if strings.HasPrefix(a, "-") {
				key, _, equal := strings.Cut(a, "=")
				if !equal && slices.Contains(strings.Fields(values["push"]), key) && !slices.Contains(strings.Fields(flags["push"]), a) {
					i++
				}
				continue
			}
			base = append(base, "-c", "remote."+a+".mirror=false")
			break
		}
		args = append([]string{"push", "--receive-pack=git-receive-pack"}, args[1:]...)
	}
	if len(args) > 0 && args[0] == "fetch" {
		args = append([]string{"fetch", "--upload-pack=git-upload-pack"}, args[1:]...)
	}
	cmd := exec.CommandContext(ctx, g.git, append(base, args...)...)
	cmd.Dir = g.path
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, "GIT_") && !strings.HasPrefix(v, "SKYRA_") {
			cmd.Env = append(cmd.Env, v)
		}
	}
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0", "GIT_EDITOR=false", "GIT_SEQUENCE_EDITOR=false", "GIT_ATTR_NOSYSTEM=1")
	var out boundedOutput
	cmd.Stdout, cmd.Stderr = &out, &out
	err := cmd.Run()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	return string(out.data), out.truncated, err
}
func (g repository) query(args ...string) (string, error) {
	out, truncated, err := g.run(args...)
	if err != nil {
		return "", fmt.Errorf("git %s: %s: %w", args[0], strings.TrimSpace(out), err)
	}
	if truncated {
		return "", fmt.Errorf("Git preflight output exceeds 64 KiB; refusing incomplete evidence")
	}
	return strings.TrimSpace(out), nil
}
