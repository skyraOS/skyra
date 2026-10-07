# Repository development tools

Repository-selected Think conversations expose the same capabilities to Codex
and Claude. Provider names are `repo_write` and `repo_check`; recorded targets
retain dots. The launcher binds the host to this repository and registers
repository-aware `find` and `roots` alongside `read` and scratch `write`.

## Search and read

```json
{"in":"controller","match":"proceduretools","word":true,"context":2}
{"in":"worktrees/task","changed":true,"match":"TODO","since":"HEAD"}
{"path":"controller/hosts/repo/main.go","around":{"line":100,"context":10}}
{"paths":["go.mod",{"path":"controller/hosts/repo/main.go","start_line":1,"end_line":80}]}
{"path":"controller/hosts/repo/main.go","outline":true}
```

`find` retains name/date/size/depth/include/exclude filtering. Content search
supports RE2 regex, case folding, whole identifiers, context 0..10, per-file
hit limits and `files_only`. `max_results` counts hits, or files when
`files_only`; maximum 500. Modified ordering is newest first with path ties.
`changed` includes tracked differences and untracked files relative to a pinned
commit; it does not fetch. Search walks at most 20,000 entries for five seconds,
reads files up to 2 MiB, and caps result output at 64 KiB. Lines are cut at 300
characters; binary, non-UTF-8, oversized and denied files have separate counts.
Repository-host search excludes Git metadata and runtime state and refuses
symlink scope paths. Ordinary read grants remain enforced by the leaf sandbox.

`read` supports up to 16 paths or range objects, `around`, `numbered`, and
`outline`. Go outlines use the parser; Swift declarations and Markdown headings
use a lightweight scan rather than a language server. Reads cap files at 16 MiB,
source at 64 KiB and 2,000 lines, and directory listings at 1,000 entries.
Batch replies share a 64 KiB budget. Metadata and plain source arrive in separate
provider content blocks; recorded observations keep their original JSON for replay.
Truncation hints explain how to narrow or continue.

## Guarded source edits

```json
{"path":"controller/example.go","old":"old text","new":"new text","count":1,"sha256":"<read hash>"}
{"path":"controller/example.go","edits":[{"old":"first","new":"second"},{"old":"third","new":"fourth"}],"dry_run":true}
{"path":"controller/new.go","text":"package example\n","create_only":true}
{"revert":"controller/example.go"}
```

Whole-file `text` and `append` remain supported. Exact edits require nonempty
`old` with exactly `count` occurrences (default 1); mismatch errors return counts
and line locations. Up to 32 edit pairs validate together before writing.
Optional SHA256 refuses stale content. Files and resulting text are limited to
1 MiB. Writes refuse traversal, symlinks, hardlinks, nonregular files, `.git`
and `.skyra` components. Creation is exclusive when `create_only` is set.

`dry_run` previews existing or absent files without writing or creating folders.
Every write returns a capped unified diff, new hash/size and changed line range.
`revert` restores this conversation's original content only if the current hash
still matches its last recorded write. Other conversations' files and externally
edited files are refused. Undoing creation would delete a file and is unavailable.
The private bounded journal lives under the Git common directory; its identity
comes from the runtime, never model input. `roots` with `changes:true` reports
written paths, before/after hashes, and commits across this conversation's turns.
Dry runs, searches and check reports do not count as source writes.

## Worktrees and bounded Git

```json
{"target":"git","args":["worktree","add","-b","agent/task","worktrees/task"]}
{"target":"git","args":["worktree","list","--porcelain"]}
{"target":"git","args":["-C","worktrees/task","status","--short"]}
{"worktree":"worktrees/task","path":"controller/example.go","text":"package example\n"}
{"worktree":"worktrees/task","action":"go.test","package":"./controller/example","run":"^TestOne$"}
```

New worktrees must be fresh `worktrees/<name>` destinations with a new `-b`
branch or `--detach`; an optional final start point defaults to HEAD. Checkout
uses committed source and preserves the original index and working files.
`repo.write`/`repo.check` accept `worktree`; Git `-C` accepts only registered
same-repository checkouts. Removal, move, prune, repair, force creation and `-B`
are unavailable. Arbitrary repository overrides are refused.

Read-only status/diff/log/show/ls-files/ls-tree/rev-parse/merge-base queries are
available, including `log -N`, `branch --contains`, `diff A...B`, and Git directory
metadata queries. Git accepts argv, never shell text. Add, commit, merge,
branch/tag creation and configured-remote fetch retain their existing allowlists.
Commit and merge need no further approval. **The launcher disables push by
default**; an operator can enable the existing bounded publishing capability with
`SKYRA_ALLOW_PUSH=1`. `roots` reports the active policy. This is an operator setting,
not a model-controlled switch or a per-call approval mechanism.

Delete/reset/clean/checkout, deleting refs/refspecs, mirror/prune pushes, config
writes, shell aliases, hooks, signing subprocesses, custom filters/merge drivers,
external drivers and arbitrary executable targets are refused. Commit/add detect
tracked/staged deletions. Merge previews one pinned commit and refuses conflicts
or removed paths. Enabled pushes pin the outgoing commit, check destination paths
and use exact leases; normal pushes still require fast forwards. Renames that
remove an old path count as deletion. Existing explicitly requested force-push
behavior applies only when publishing is operator-enabled.

Git and writes share an advisory lock in the common Git directory. External
editors and user Git commands do not take it; SHA guards are not an atomic lock
against those commands. Host connections receive custody independently, so
checks and runtime continuations cannot block other conversations' handoffs.
Git stdout is separate from diagnostic stderr, preventing messages corrupting
machine-readable queries. Output caps at 64 KiB; Git has a two-minute timeout.

## Checks

Fixed recipes: `go.test`, `go.build`, `go.vet`, `go.format`, `swift.test`,
`swift.build`, and `python.test` (Blender unit tests). Go defaults to
`./controller/...`; explicit packages stay under `controller/`, and
`package:"changed"` selects packages containing changed Go files. `run` compiles
as a regex and is passed as one Go `-run` or Swift `--filter` argument. Go
`count` is 1..20, default 1 to disable cached test results.

Checks snapshot at most 128 MiB of build inputs under the repository lock, then
release it before running. The worker uses macOS Seatbelt, a constructed environment,
installed toolchains and cached dependencies. It can write only to its disposable
job and narrowly selected Swift replacement/SDK staging files. Original source,
private home files, credentials, external network and external Unix sockets are
unavailable. Localhost fixtures are allowed. No recipe accepts arbitrary commands,
extra argv, environment or scripts. Native system-service grants are a development
runner boundary, not a complete macOS service audit.

Formatting normally reports files needing gofmt; `fix:true` validates all reported
paths/hashes and applies formatted files through repository writes. An external
edit or I/O failure can still prevent application; returned `formatted` names the
files actually changed. Jobs have a five-minute deadline and process-group timeout
termination. Results include actual exit status, output capped at 64 KiB,
structured Go test/compiler and Swift XCTest/compiler failures, elapsed time,
truncation hints, and JSON `report_path` under `ideas/checks/` in the selected
checkout. Swift Testing failures without XCTest report entries retain raw output.
Build outputs are discarded; checks do not install or launch the application.
Dependencies must already be cached.

Repo-host tests using nested sandboxes or outer host sockets explicitly skip
inside `repo.check`; run the normal external test suite to exercise them.

## Scratch notes and installation

Repository scratch `write` saves one file under `ideas/` and automatically records
only that file on `skyra/ideas`, with agent author and configured user committer.
The user's current HEAD and staging area are preserved; nothing is pushed.
The reply includes commit, branch, or commit_error. See the
[write workflow](../../interp/workflows/write/README.md).
Other workspace selections retain plain scratch writes.

Run `go test ./controller/...` for contracts, guards, local disposable Git
repositories, sandbox recipes, custody, provider parity and offline replay.
Rebuild with `./skyra build`, restart the backend, and start a new conversation.
Existing conversations retain their pinned contracts and binaries.
