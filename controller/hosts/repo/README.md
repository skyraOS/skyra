# Repository edits and Git

When this Skyra repository is the selected workspace, new conversations receive
`repo.write` and `execve` alongside the existing scratch `write` procedure.
`repo.write` creates, replaces, or appends UTF-8 text anywhere in the worktree:

```json
{"path":"controller/example.go","text":"package example\n"}
```

Git's input contract has only `target` and `args`. `target` is `git` (or
`/usr/bin/git`), and `args` is argv, without shell parsing:

```json
{"target":"git","args":["add","--","controller/example.go"]}
{"target":"git","args":["commit","-m","Add example"]}
{"target":"git","args":["push","origin","main"]}
{"target":"git","args":["merge","feature"]}
```

These are recorded host actors using the existing park/custody protocol.
The operator starts one private repository host per backend; it binds both
operations to this repository. Sandboxed model actors still cannot start a
shell, execute arbitrary programs, write Git metadata, or read credentials.
Git uses the operator's configured remotes and authentication. The host's
output is bounded to 64 KiB, commands time out after two minutes, and replay
uses the recorded observations even after the host has stopped.

Commit, push (including explicitly requested force pushes), and merge require
no further approval. File path and ref deletion are unavailable. Git commands
and options are validated before execution: rm, clean, reset, checkout,
branch/tag deletion, deleting push refspecs, mirror/prune pushes, config writes,
arbitrary targets, shell aliases, external drivers and repository overrides
are refused. Writes reject traversal, symlinks, hardlinks, nonregular files,
`.git` components and `.skyra` state. Replacing a file's text is allowed.

Commit/add check tracked and staged deletions. Merge previews a single pinned
commit using [Git merge-tree](https://git-scm.com/docs/git-merge-tree), then
checks that existing file paths survive. Conflicting merges are refused before
worktree changes. Push checks each existing destination's tree, pins the
outgoing object, and adds an exact ref lease to reject a destination that
changed after the check. Renames that remove an old path count as deletion.
Git's own internal lock/ref housekeeping is permitted; the restriction concerns
worktree paths and user branch/tag refs. A normal push still requires a fast
forward; a requested force push may rewrite history while preserving those paths.

The host disables hooks, fsmonitor, signing subprocesses, automatic maintenance,
submodule recursion and merge autostash. Repositories configured with custom
filters or merge drivers are refused. Git and scratch automatic commits share
an advisory lock; external user commands need not take it, so avoid changing
the worktree or index concurrently with an agent operation.

Run `go test ./controller/hosts/repo ./controller/hosts/ideas
./controller/codex/resident` to verify edits, commit/push/merge, refusals, custody
and offline replay. Tests push only to disposable local repositories.

Restart the backend and open a new conversation after installation. Existing
conversations retain the capability set pinned at their creation.

## Build and verification tools

`repo.check` is offered to Think alongside `repo.write` and bounded Git.
Provider names are `repo_check` and `repo_write`; recorded targets retain dots.

```json
{"action":"go.test","package":"./controller/blender"}
{"action":"go.build"}
{"action":"go.format"}
{"action":"swift.test"}
{"action":"swift.build"}
{"action":"python.test"}
```

Go defaults to `./controller/...`. Package selectors stay under `controller/`;
there are no extra argv, environment, command, or script fields. Formatting is
inspection: it reports files requiring gofmt instead of changing them.
Python currently runs Blender adapter unit tests. Swift builds/tests the native
`desktop` package using full Xcode.

The host copies current source into a temporary workspace, omitting ideas,
Git/runtime metadata, symlinks, caches and prior build outputs. The copy is
limited to 128 MiB. Code, tests, build plugins and their subprocesses run under
macOS Seatbelt: they can write to the disposable job, read the snapshot,
installed toolchains/system libraries and dependency cache, and use loopback
networking plus Unix sockets within the job. External network destinations and
original checkout/private home file reads are denied. Environment variables are
constructed from scratch; operator credentials and Skyra tokens are not inherited.
Swift additionally needs Foundation’s temporary replacement directories for atomic
writes. Only new `NSIRD_swift-build_*`, `NSIRD_swift-test_*`, and
`NSIRD_swift-driver_*` folders in the
operator’s temporary pool are granted; pre-existing matching folders are denied.
New `xcrun_db-*` staging files are allowed for SDK discovery; the shared
`xcrun_db` metadata cache is readable but cannot be written.
The runner does not grant the entire system temporary directory.
Subprocess execution is allowed inside this same sandbox for build systems;
there is no tool accepting arbitrary Bash or executable commands. System service
and IOKit access remain broad for native builds; this is a development runner,
not a completed audit of every macOS service boundary.

Each job has a five-minute deadline and process-group termination on timeout.
Output is limited to 64 KiB. Results include `exit_code`, `error`, `output`,
`truncated`, `duration_ms`, and a JSON `report_path` under `ideas/checks/`.
Checks with nonzero exit or a reported error are unsuccessful provider results.
Build executables/caches are discarded; these actions verify compilation rather
than installing or launching the app. Dependencies must already be available;
the runner does not download packages or grant network access to install them.

The app backend automatically registers these tools when this repo is the
selected workspace. Restart the backend and start a new Think conversation to
pick up the new capability set. Existing conversations keep their pinned tools.
