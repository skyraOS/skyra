# Skyra repository development tools

Repository-selected Think conversations offer these tools to both providers:

| Tool | Capability |
| --- | --- |
| `read` | Line ranges, batch reads, around-hit reads, outlines and numbered source; plain text plus metadata. |
| `find` | File and bounded content search, identifier/regex matching, context, modified ordering and changed-file filtering. |
| `repo_write` | Source creation/replacement/append, exact guarded edits, previews, diffs and conversation-scoped restoration. |
| `execve` with `target:git` | Bounded queries and local Git operations, plus independent worktrees under `worktrees/`. |
| `repo_check` | Go/Swift/Python fixed checks, test filters, changed Go packages, vet, guarded format fixes and failure locations. |
| `write` | Scratch notes recorded on the separate `skyra/ideas` ref with agent authorship; preserves feature HEAD/index. |
| `roots` | Workspace, scratch, active permissions and `changes:true` for conversation-written files/hashes/commits. |

Commit and merge remain local operations without approval. The launcher disables
push by default; the operator can enable its existing policy with
`SKYRA_ALLOW_PUSH=1`. Nothing pushes scratch notes automatically.
`desktop_observe` is conditional on a configured connected private desktop bridge
(`SKYRA_DESKTOP_CONFIG`), rather than part of every conversation.

```json
{"in":"controller","match":"serveConnection","word":true,"context":2}
{"path":"controller/hosts/repo/main.go","around":{"line":100,"context":10}}
{"path":"controller/example.go","old":"before","new":"after","sha256":"<read hash>","dry_run":true}
{"target":"git","args":["worktree","add","-b","agent/task","worktrees/task"]}
{"action":"go.test","package":"changed","run":"^TestOne$","worktree":"worktrees/task"}
{"changes":true}
```

No general shell, arbitrary executable recipes, deletion or arbitrary repository
switching is added. Checks execute a temporary source snapshot inside macOS
Seatbelt without inherited credentials or external network. Build outputs are
discarded; actual exit status, bounded logs, failure locations and report paths
are returned. Existing configured-remote Git fetch remains available separately.
Revert refuses externally changed files; undoing newly created files would require
deletion and remains unavailable.

The [host reference](../controller/hosts/repo/README.md) documents limits, recipes,
Git guards, worktree selectors, caveats and installation. The proposal and final
implementation mapping are in [agent-tool-improvements](agent-tool-improvements.md).
Rebuild and restart the backend, then start a new conversation: existing ones keep
pinned tools. This implementation does not commit the development changes.
