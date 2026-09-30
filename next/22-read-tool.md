# The read tool: the first leaf that observes the world

Date: 2026-09-28  
Status: built (`controller/targets/read/`), verified with the fake provider and the
OS tests. Not yet validated against live Codex. Follows [20](20-boundry-crossing.md)
and [21](21-observation-passing.md).

> **Revised 2026-09-30** ([24](24-the-room.md), [25](25-surface-refactor.md)).
> `read` no longer has an empty reach of its own: it runs under its caller's
> Horizon, with that Horizon's reach and grants, and the profile is rendered
> per Horizon. It returns by addressing its slot (`SKYRA_FRAME`), not
> `reply`; output that addresses nothing is a leaf fault that keeps the
> output. One guarantee is traded: under a shared Horizon, `read` addressing
> `codex` derives an invoke, not a horizon fault. The Horizon is the unit that
> contains a tool; to contain `read` more tightly, its caller narrows it
> through `system.narrow` ([26](26-agents-and-facets.md)). File contents still
> only ever travel inside `payload.text`, never as `read`'s envelope.

## What it is

`read` is a leaf: it has an empty reach, it shares no code with skyra, and it checks
no authority. It reads a range of lines from one UTF-8 file or lists one
directory, and replies with what it observed. The workspace grant in the
registry is the Horizon the human approves once, when they start `skyra` in a
directory. Every read inside that directory then runs with no prompt. Outside
it, the OS says no:

```text
registry   read: authority.read = [{workspace, subpath}]    the grant, approved once
reach      user ⊇ codex ⊇ read                              Narrow: both must list it
profile    (allow file-read* (subpath workspace))           rendered per invocation
           (deny file-read* (subpath <this run's .skyra>))  the floor, always
env        SKYRA_READ=<workspace>                           information only: resolves relative paths
```

## What it takes and what it returns

**Input**: operands Codex authors on the way down. `path` is relative to the
workspace, or absolute. `start_line` and `max_lines` are optional, and the leaf
clamps them because contracts have no bounds.

**Output**: an observation the leaf authors (next/21: a leaf may author its
return). It carries its own provenance:

| Field | Why |
|---|---|
| `path`, `kind` (`file`, `dir`, `other`) | what was observed |
| `sha256`, `size` of the **whole** file | the content address of the source. Later: a `write` that takes `expect_sha256` refuses if the file changed since it was read |
| `start_line`, `end_line`, `total_lines`, `truncated`, `text` | the slice, and how to page. At most 2000 lines or 64 KiB per read |
| `entries` | a directory listing, since Codex's native `ls` faults |
| `error` (`not_found`, `denied`, `not_text`, `too_large`, `not_regular`) | failure is a return, not a fault (design/faults.md). It is the leaf's claim; the profile is the truth |

## One correction found while building

`interpret` gives unaddressed output to `reply` **as a string**. A leaf whose
return contract is an object must therefore address reply explicitly:
`{"to":"reply","payload":{…}}`. The first version printed a bare object, and
intake derived `fault (contract): reply: want an object, got a string`.
Addressing reply keeps the object contract strictly checked at intake, and it
means file contents only ever cross as a string inside `payload.text`. A file
whose contents are `{"to":"echo",…}` is returned as text. Even a leaf that
addressed a capability would get a horizon fault.

## The adapter

`codexresident` no longer names echo. `modelTools` (name → description) is
the adapter's split between model-facing tools and its own `codex.*`
operations (next/20). A tool is offered to Codex only if it is also in the
continuation's reach. Each pending call is invoked under the tool it names and
answered by the step at the same position. A step for a different tool is an
error, not an answer. A tool's fault goes back to Codex as `success:false` in
the same turn.

## Limits, and what is open

- **The floor protects only this run's `.skyra`.** Sibling conversations and
  `~/.skyra/providers/` (which links the Codex login) are not denied by the
  profile. For that reason the `skyra` command refuses a workspace that
  overlaps `~/.skyra` or `~/.codex`, or that is `$HOME` or `/`. A state-root
  floor in `Render` would make that refusal unnecessary.
- **Every read is stored twice and replayed forward.** It is stored in
  `inputs/` and in the record payload, and it appears again in every later
  adapter continuation's `steps`. The per-read cap bounds each read, but not a
  turn.
- **File contents go to the model provider.** That is what granting a
  workspace means, and the startup doc says so.
- **One workspace per `skyra` process.** A per-conversation workspace is a plan
  (a Horizon issued by the user's side), which is out of scope here.
- **Live Codex** has not seen the new tool surface. Rerun
  `controller/experiments/appserver/tool_surface.py`, and confirm that optional
  properties with `additionalProperties:false` are accepted.
