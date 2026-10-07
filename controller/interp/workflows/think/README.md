# think

The continuation procedure ([proposal 38](../../../../docs/next/proposal/38-think-procedure-proposal.md))
attaches a minimum tool scope for this invocation, then invokes Codex or Claude
on the user's text. The minimum tools are `echo`, `read`, `find`, `roots`, and
scratch `write`, plus provider operations. `echo` is a protocol test fixture.
The prompt identifies the workspace and scratch roots; provider working folders
are isolated and do not determine where tool paths resolve.

The bundle is the minimum capability set. Repository-selected conversations
also receive `repo_write`, bounded Git through `execve`, and `repo_check`.
Repository-aware `find` adds changed-file searches; `roots` adds permissions
and conversation change summaries. See [repository development tools](../../../../ideas/skyra-repo-development-tools.md)
and the [host reference](../../../hosts/repo/README.md).
Scratch `write` records notes on the separate `skyra/ideas` ref. Push is disabled
by default and requires operator configuration (`SKYRA_ALLOW_PUSH=1`).

`desktop_observe` is added only when the operator configures and connects the
private desktop bridge with `SKYRA_DESKTOP_CONFIG`. Its absence from an ordinary
conversation is expected. Actors gain only capabilities in that invocation;
there is no general shell or deletion tool.

Install by naming the bundle in the registry. Registry changes require a backend
restart and a new conversation; existing conversations keep their pinned tools.
