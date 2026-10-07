# Agent tool improvements

Proposals from a `think-claude` conversation (2026-10-07), based on using the
tools in this workspace to answer questions about the code. They are ordered by
how much each gap actually cost. None of them adds new authority except §6,
which is marked optional.

## Summary

| # | Change | Cost today | Size |
| --- | --- | --- | --- |
| 1 | Content search in `find` | Code questions need guessing and file-by-file reads | Medium |
| 2 | Exact-text edit in `repo_write` | One-line change means rewriting the whole file | Small–medium |
| 3 | Tell the agent where paths resolve | Agent starts out unsure where it is | Small |
| 4 | Plain-text `read` output | Escaped code is harder to read and costs tokens | Small |
| 5 | Update docs on tool grants | Docs disagree with the tools the agent actually has | Small |
| 6 | Read-only Git queries; optional worktree support | Agent can't confirm or use a worktree | Small / optional |
| 7 | Describe or hide `echo` | An unexplained tool | Trivial |

---

## 1. Content search in `find`

**Problem.** `find` matches names, paths and file details only. The Git tool
refuses `grep`. A question like "do the Codex and Claude providers share
code?" needed one directory listing after another, then opening files
to look at their imports. A single search would have answered it.

**Proposal.** Add a content filter to `find`, applied after its existing filters
(`include`/`exclude`, `maxdepth`, `newer`, `larger`). The scope options stay
the same, and only the files that pass them are opened.

New inputs (`controller/targets/find/contract.json`):

| Field | Type | Default | Meaning |
| --- | --- | --- | --- |
| `match` | string | — | Text to look for in each file's lines. Enables content search. |
| `regex` | bool | false | Treat `match` as RE2 (linear-time, no catastrophic backtracking). |
| `case` | bool | true | Case-sensitive matching. |
| `context` | number | 0 | Lines returned before and after each hit (max 10). |
| `max_per_file` | number | 20 | Stop recording hits in a file after this many; count the rest. |
| `files_only` | bool | false | Report matching files without lines (like `grep -l`). |

Result entries when `match` is set (`files_only` false):

```json
{"path":"controller/claude/resident/main.go","kind":"file","line":23,
 "text":"\t\"skyra/controller/codex/proceduretools\"",
 "before":["\t\"skyra/controller/claude/bridge\""],
 "after":["\t\"skyra/controller/codex/prompt\""]}
```

- `line` is 1-based and the same as `read`'s `start_line`, so the next step is
  a direct `read`.
- With `files_only`, entries look like today's, plus `"hits": n`.

Limits and reporting, consistent with the rest of `find`:

- `max_results` counts hits (or files with `files_only`). The description must
  say so.
- The existing `maxOutput`, `maxWalked` and `maxTime` limits still apply, and
  `truncated_by` gains `"per_file"` when a file's hits were cut.
- New `skipped` counts: `binary` (NUL byte in the first 8 KiB), `not_text`
  (invalid UTF-8), `too_large` (above a per-file read limit, e.g. 2 MiB).
  Otherwise "no hits" can't be told apart from "not searched".
- Returned lines are cut at 300 characters with `"cut": true`, so minified files
  and one-line JSON can't fill the whole reply.
- `match` with `kind` `"dir"` or `"any"` returns `bad_kind`. An empty `match`
  string returns `bad_match`. An invalid regex returns `bad_regex: <message>`.
- Reading stays inside the leaf's existing sandbox, so no new authority.
  A file the sandbox refuses counts as `denied`.

**A design note to settle.** The header comment in
`controller/targets/find/main.go` says "a later grep takes the same shape
(proposal 36)", which means a separate `grep` leaf was planned. Putting it in
`find` keeps one tool and one set of scope options, which fits keeping the
agent's tool list small. A separate `grep` would keep each tool's description
shorter. I'd still put it in `find`, but proposal 36 (not in this checkout)
should be checked first.

**Tests** (`main_test.go`, table style like the existing tests): literal hit
with line numbers; regex; case-insensitive; context at the start and end of a file;
`include` narrowing which files are read; `files_only`; `max_per_file` truncation;
binary/non-UTF-8/oversized files counted as skipped; long-line cutting; the three
new errors.

## 2. Exact-text edit in `repo_write`

**Problem.** `repo_write` replaces a whole file or appends to it. Changing one
line of a 12 KB file means sending all 12 KB back. That costs tokens and
can silently drop content if the model miscopies anything.

**Proposal.** Add an edit mode, either as a new mode of `repo_write` or a
separate `repo_edit`:

```json
{"path":"controller/targets/find/main.go",
 "old":"\tMaxResults *float64 `json:\"max_results\"`\n",
 "new":"\tMaxResults *float64 `json:\"max_results\"`\n\tMatch      string   `json:\"match\"`\n",
 "count":1, "sha256":"00d9d3a0…"}
```

- `old` must appear exactly `count` times (default 1). If not, nothing is
  written and the error gives the actual count and the line numbers of each
  occurrence: `"old_matched: 0"` / `"old_matched: 3 at lines 12, 40, 88"`.
- Optional `sha256` is the hash `read` returned. If the file changed since
  then, the edit is refused (`stale`), so it can't overwrite someone else's change.
- The reply gives the new `sha256`, the size, and the changed line range.
- The same rules as `repo_write` apply: no symlinks, hardlinks, `.git` or
  runtime state.
- Optionally allow several `{old,new}` pairs in one call, all applied or none,
  so a change across one file is a single recorded step.

## 3. Tell the agent where paths resolve

**Problem.** The agent's environment reports its working directory as a
per-turn provider folder (`…/providers/claude-NNN/work`) that "is not a git
repository". The tools actually work on the repository root. I only worked
this out by calling `roots`.

**Proposal.** Pick one:
- Add one line to the think prompt (the `codex/prompt` and Claude bridge setup):
  `Tools resolve paths against <workspace>; scratch is <scratch>.`, which is the
  same information as `roots`; or
- start the provider in the workspace, if the provider sandbox allows that.

The first is simpler and doesn't change isolation.

## 4. Plain-text `read` output

**Problem.** `read` returns file contents inside a JSON string, so code comes
back with `\n`, `\t`, `\"` and `&` escapes. That's harder to read and
uses noticeably more tokens than the raw text.

**Proposal.** Return two MCP content blocks: a small JSON block with the details
(`path`, `start_line`, `end_line`, `total_lines`, `truncated`, `sha256`,
`size`) and a plain-text block with the file contents unchanged. Recorded
replay is unaffected, since the observation is still deterministic. An optional
`numbered: true` could prefix each line with its line number to make
`old`/`line` references easier.

## 5. Update docs on tool grants

**Observation.** `controller/interp/workflows/think/README.md` and the think
bundles describe a read-only set (`echo, read, find, roots, write`).
In a repository-selected conversation the agent also gets `repo_write`,
`execve` (Git, including push and merge without approval) and `repo_check`.
`ideas/skyra-repo-development-tools.md` explains why. The think README doesn't
mention it, and `desktop_observe`, listed there, wasn't in this
conversation's tools.

**Proposal.**
- Have the think README point to the repository development tools doc and say
  the bundle's set is the minimum, not everything the agent can get.
- Check whether `desktop_observe` should have been granted here.
- Reconsider whether **push** should need no approval. Commit and merge are
  local and can be undone; push publishes. Keeping the grants minimal argues for
  requiring approval on push, or leaving it out by default.

## 6. Read-only Git queries; optional worktree support

**Problem.** The agent couldn't confirm whether it was in a linked worktree:
`git worktree list` and `git rev-parse --git-dir` / `--git-common-dir` are
refused. It worked it out by checking that `.git` is a directory.
Separately, the user wanted a change made in a worktree, and the agent couldn't
set one up.

**Proposal (no new authority).** Allow read-only queries:
- `worktree list` (only `list`);
- `rev-parse --git-dir`, `--git-common-dir` and `--is-inside-work-tree` as
  queries only. The `--git-dir=<path>` and `-C` that switch repository stay
  refused.

**Optional (adds authority, so decide deliberately).** Let the agent work in a
worktree without widening Git much:
- allow `worktree add worktrees/<name> -b <new-branch>` only: path under
  `worktrees/`, new branch only, no `--force`, no removal;
- let `repo_write` and `repo_check` take a `worktree` field that changes which
  folder they act on to `worktrees/<name>`, with the same rules. Without this
  second part, a worktree is useless to the agent: the checker only runs the
  main checkout, and Go treats the worktree's `go.mod` as a separate module.

If that's too much, the current workaround is for the user to create the worktree
and for the agent to keep its edits in folders the user isn't changing. That
only works when the change doesn't overlap anything uncommitted.

## 7. Describe or hide `echo`

`echo` takes `misbehave` and `probe` parameters and is described only as
"Skyra echo". It looks like a test fixture. Either describe when an agent should
use it, or leave it out of the think scope so the tool list contains only tools
the agent might use.

---

## Suggested order

1. §1 content search and §2 exact-text edit: the largest effect on how often the
   agent gets it right first time.
2. §3 and §5: one-line prompt and doc fixes.
3. §4 plain-text read: small change to the reply format, saves tokens on every read.
4. §6 read-only queries, then decide on the optional worktree support.
5. §7 whenever convenient.

---

# Part 2: Further features

## Principle

No general shell. Each capability is an **optional parameter on a tool that
already exists**, limited by that tool's existing sandbox and rules. A new
parameter must not reach anything the tool couldn't already reach. Where one
would, it's marked **adds authority** and needs a deliberate decision. New
tools only where no existing tool fits.

Today's tool set is enough: `read`, `find`, `write`, `repo_write`,
`repo_check`, `execve` (Git), `roots`.

## 8. Agent writes create commits without saying so

**Observation.** The scratch `write` tool (writing the first part of this doc)
committed the file as `0f9b6c3 "ideas: save agent write"` on the user's
current feature branch, under the user's name. The commit contained only that
file and nothing was pushed, but the tool's description doesn't mention
committing, and the commit lands on whatever branch is checked out.

**Proposal.**
- Say so in the description, and return `commit` *and* `branch` in the
  result (it currently returns only `commit`).
- Better: record agent writes somewhere other than the user's branch, e.g. a
  dedicated `skyra/ideas` ref updated with Git's low-level commands (no
  checkout), or simply don't commit and let the user decide. Agent notes
  shouldn't mix with feature work and end up in a PR by accident.
- Mark the agent as the author (e.g. `Skyra (think-claude) <…>`) with the
  user as committer, so history shows who wrote what.

## 9. `read`

| Parameter | Effect | Why |
| --- | --- | --- |
| `paths: [..]` | Read several files (or `{path,start_line,end_line}` ranges) in one call, each with its own details. | Answering most questions means reading 3–6 related files; one round trip instead of six. |
| `around: {line, context}` | Return `context` lines either side of `line`. | Pairs with `find match` results: jump straight to a hit. |
| `outline: true` | For Go (via `go/parser`) and Swift/Markdown (headings), return top-level declarations with line ranges instead of the contents. | See the layout of a 400-line file for a few hundred tokens, then read only the part needed. |
| `numbered: true` | Prefix each line with its line number (from §4). | Exact line references for edits and for the user. |

No added authority: all of these read the same paths `read` already can.

## 10. `find`

In addition to §1:

| Parameter | Effect | Why |
| --- | --- | --- |
| `word: true` | Match `match` only as a whole identifier. | Finding the uses of `Tool` without `Tools`, `ToolName`… gets close to "find references" with no language server. |
| `sort: "modified"` | Most recently modified first. | "What changed recently here" is a common first question. |
| `changed: true` | Only files changed relative to `HEAD` (or a `since` branch/commit), via a read-only `git status`/`diff --name-only`. | Limits a search or a check to the work actually in progress. Also lets the agent avoid the user's uncommitted files, the problem that came up with worktrees. |

`changed` reads Git metadata. It must use the same read-only Git path and
never fetch.

## 11. `repo_write`

In addition to §2 (exact-text edit):

| Parameter | Effect | Why |
| --- | --- | --- |
| `dry_run: true` | Validate and return the diff, write nothing. | Show the user a change before applying it. |
| reply `diff` | Every write returns the diff it applied (capped). | The agent can check what landed without a separate `read`. |
| `revert: <path>` | Restore a file to the content it had before **this conversation's** first write to it. Only files this conversation wrote, nothing else. | A safe undo. Writes can't delete, so a bad change can only be fixed by writing again; this makes that one call and stops it touching others' work. |
| `create_only: true` | Fail if the file already exists. | Stops a "new file" from silently replacing an existing one. |

`revert` needs the runtime to keep the starting hash/content of each file the
conversation wrote. Its record of steps should already have this.

## 12. `repo_check`

| Parameter | Effect | Why |
| --- | --- | --- |
| `run: "<regex>"` | Go `-run` / Swift `--filter` for a single test. | Running the whole package to check one test wastes minutes. Still a fixed recipe; the value is passed as one argument, never through a shell. |
| `package: "changed"` | Check only the packages containing files changed from `HEAD`. | Fast inner loop. |
| action `go.vet` | `go vet` on the package. | Catches a class of bugs tests miss. Same sandbox. |
| `fix: true` on `go.format` | Apply `gofmt` to the files it reports, through `repo_write`'s rules. | Today it only reports. The agent then has to rewrite whole files by hand to fix whitespace. |
| structured `failures` | Alongside raw `output`: `[{test, package, file, line, message}]`, parsed from `go test -json` / Swift's report. | The agent goes straight to the failing line instead of digging through 64 KiB of logs, which gets truncated. |
| `count: 1` (Go) | Turn off the test cache for the run. | Know that a pass was actually run. |

`run`/`package` values are validated (regex compiles; package is `./…` inside
the repo) and passed as single arguments. No added authority.

## 13. Git (`execve`)

Read-only queries, in addition to §6:

- `log -n N` / `-N`. Today `-2` and `-3` are refused, but `--max-count=2` is
  allowed. Same effect, so allow both forms, or say in the refusal which form
  is allowed.
- `branch --contains <commit>`, `merge-base A B`, `show <commit>:<path>`,
  `diff --stat A...B`. All read-only, all useful to answer "where does this
  commit live?" or "what does this branch change?"
- Refusal messages that name the allowed alternative
  (`"use --max-count=N"`), not just the refusal.

**Adds authority, so decide deliberately:** push currently needs no approval.
Commit and merge are local and can be undone; push publishes. Requiring
approval for push (§5) is the one *reduction* proposed here.

## 14. Across all tools

- **What this conversation has changed.** An optional `changes: true` on
  `roots` (or a field always present in its reply): the files this
  conversation wrote, with before/after hashes, plus any commits it made.
  Makes handing back to the user accurate, and lets the agent check
  what it changed before saying it's done.
- **Consistent truncation hints.** `read` already says how to continue. `find`
  and `repo_check` should too: `truncated_by: "results"` plus
  `"hint":"narrow with include/in, or raise max_results (≤500)"`.
- **Refusals that say what's allowed.** Every policy refusal names the
  permitted form or the setting that would change it, so the agent
  adapts in one step instead of guessing.
- **Agent-visible permission summary.** One line in the prompt or `roots`:
  which tools are writable, whether commits/pushes need approval, and where
  writes land. The agent shouldn't learn its permissions by trial and error.

## 15. Deliberately left out

- **Shell / arbitrary commands.** By design. Every case above is covered by a
  limited parameter.
- **Delete / rename.** Kept out. `revert` covers undoing the agent's own
  mistakes. Deletion stays a user action.
- **Network fetch.** Documentation lookup would help, but it adds egress
  authority. If it's ever added, it should be a read-only fetch limited to an
  approved list of domains (e.g. pkg.go.dev, developer.apple.com) and
  recorded like every other observation.

## Updated order

1. §1 content search, §2 exact-text edit, §8 commit behaviour.
2. §12 `run` + structured `failures`, §9 batch/outline `read`.
3. §3, §5, §14 permission summary and hints: prompt and description changes only.
4. §10, §11, §13: smaller parameter additions.
5. §4, §6, §7 as convenient.
