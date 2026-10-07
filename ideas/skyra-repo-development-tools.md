# Skyra repository development tools

When the Tony Stark repository is selected, a new Think conversation receives:

| Tool | Purpose |
| --- | --- |
| `repo_write` / recorded `repo.write` | Create, replace, or append UTF-8 source files in the repository. |
| `execve` with `target: git` | Inspect status/diffs/history; stage, commit, merge, fetch or push through the existing bounded Git policy. |
| `repo_check` / recorded `repo.check` | Run named Go, Swift and Python build/test checks, plus a Go formatting check. |
| `read`, `find`, `roots` | Read source, locate files and inspect granted paths. |
| `desktop_observe` | Inspect app accessibility text and screenshots without controlling it. |

Examples:

```json
{"action":"go.test","package":"./controller/hosts/repo"}
{"action":"go.build"}
{"action":"go.format"}
{"action":"swift.build"}
{"action":"swift.test"}
{"action":"python.test"}
```

There is no arbitrary Bash, executable, argv, script, environment or working
folder input for checks. Tests and build scripts execute inside a macOS sandbox
in a temporary source copy. The original checkout and private home files are
not readable/writable from the worker. Toolchains and cached dependency sources
are readable. Writes are limited to the job and narrowly selected Swift/SDK
temporary files. Network access is limited to loopback and job-local Unix sockets.
Operator environment variables and credentials are not inherited. This is a
bounded development runner; its native system-service grants have not undergone
a complete production isolation audit.

Repository writes stay within the worktree, refuse symlinks/hardlinks and Git/
runtime metadata, and do not provide deletion. The Git tool retains its command
and option allowlists. Checks return actual exit status, output, errors, elapsed
time, truncation, and a readable report under `ideas/checks/`. Failed checks are
marked unsuccessful to Codex. Jobs time out after five minutes. Build outputs
are temporary; checking compilation does not install or launch a new app build.

Go/Python execution and isolation tests pass. The Swift release build and native
test suite also pass inside the runner. One existing native test requiring a live
agent is skipped. Recorded acceptance calls replay byte-identically; see
`ideas/checks/acceptance.json` for results. The attachment test now creates its
fixture in a unique directory under the runner's TMPDIR.

The backend must load the updated registry. Existing conversations keep their
pinned tools; use a new Think conversation for the new capabilities. No changes
were committed as part of this task.

Implementation details: `controller/hosts/repo/README.md`.
