# 34 — Actor constellations: procedure-facing application contracts

Date: 2026-10-03

Status: proposed; not implemented. Revised to agree with
[33 — Conversation workflows](33-procedure-adapter-extraction-proposal.md)
and its [reconstruct-state procedure design](reconstruct-state/README.md).

An actor is a constellation; its input contracts are stars. The human can
inspect the capabilities in a Horizon and choose a contract for future
work. Ordinary application actions address procedures on the interpreter.
Codex's exported thread/turn operations remain underlying capabilities,
visible for inspection rather than the default chat destination.

Builds on [23](../23-desktop-app.md), [25](../25-surface-refactor.md),
[31](../31-the-sky-dreams.md), and [32](../32-procedures.md).

## The ownership boundary

The Horizon executes work, records what happened and owns the history
syscall, `system.strand`. It is available to actors in that Horizon subject
to reach, contract and history-visibility checks. The interpreter is one
caller, not the owner of that capability.

The user adapter retains each conversation's root invocation and complete
accepted observation-leaf set. Those references bound reconstruction of the
recorded tree, including parallel branches. A single latest observation is
not sufficient when multiple branches are open.

The interpreter may expose a `reconstruct-state` procedure that composes
Horizon history reads and user-state projection. It may also compose
future provider operations into meaningful application workflows. This
does not transfer execution/history ownership to the interpreter, and
user recovery may call the same syscall directly.

## What the human can do

The sidebar's conversation box represents an enclosing conversation
invocation. Selecting a box requests reconstruction through a procedure.
Sending a message invokes the application's send procedure, which uses
Codex operations underneath. Create, close, delete and refresh also have
procedure entry points.

| UI action / contract selection | Procedure destination | Behavior |
|---|---|---|
| Open the conversation list | `interpreter → conversation.list` | Discover authorized conversations and arrange their reconstructed views |
| Select a conversation box | `interpreter → conversation.load` | Trigger `reconstruct-state`; update the app through its snapshot contract |
| Send ordinary text | `interpreter → conversation.send` | Run future work under the user conversation; provider IDs are filled internally |
| Create / close / delete | Corresponding interpreter application contract | Compose the required future operations |
| Choose a remembered task | Its input contract on `interpreter` | Run that procedure, including existing refinement/fallback behavior |
| Inspect Codex / other actors | Their visible contract stars | Show underlying capabilities and recorded calls; inspection grants no authority |

Ordinary chat starts with `conversation.send` selected, not Codex's raw text
contract. The human does not supply thread IDs, incarnations, checkpoint
stamps, provider-home paths or raw RPC parameters.

Example future-work emission from the user adapter:

```json
{
  "to": "interpreter",
  "payload": {
    "conversation.send": {
      "id": "c-123",
      "text": "explain this",
      "requestId": "r-456",
      "expectedFrontier": "<opaque checkpoint revision>"
    }
  }
}
```

The final registered schema must settle the exact field spelling. This
illustrates the procedure-facing destination; it is not a change already
implemented in the desktop's current text-only `submit` message.

Direct underlying actor calls can remain an explicit developer/advanced
mode if wanted. They are not the application default and cannot bypass
Horizon intake. Do not make every visible infrastructure contract a
one-click application action merely because it has a star.

## What is already there

- `desktop/Sources/Skyra/Sky/Horizons.swift` draws regions and surface stars.
- `controller/hosts/web/horizons.go` derives historical regions from genesis
  and released Horizons. Its reached members currently contain name and
  surface identity, not all contract schemas.
- Surfaces hold their initial and added input contracts. Intake selects
  exactly one by payload fit; titles are labels rather than dispatch keys.
- `system` is installed at boot and retained through narrowing. The
  interpreter's `runRemember` already calls `system.strand`.
- The desktop and browser currently receive full snapshots. Sidebar
  selection is local only, and the user actor currently calls Codex directly.
- `hosts/web/calls.go` recognizes turns by `To == "codex"`, which must be
  adapted to user requests and procedure descendants.

The reconstruction procedure and new application contracts are design
artifacts in proposal 33, not deployed capabilities. Contract selection
must reflect what is actually installed and reachable in the selected
interaction.

## Contracts belong to an actor in a Horizon

The same executable may be reached under different names and contract
sets. Identify a selectable occurrence by actor address, contract schema
identity and its live context. Use titles or readable shape summaries as
labels; display hashes only as internal identities, not task names.

Layout stays stable as contracts appear. Lines joining contract stars are
visual organization, not a claim that one contract invokes another.
Procedure wiring remains separate from actor/contract membership.

Historical stars are inspectable. Only contracts held in the live emitting
user context are selectable for a call. A reference discovered in another
frame's history is not transferable authority. The host derives live
context from validated runtime/custody evidence, not a client-supplied
Horizon ID or a reconstructed assumption that an old parked frame is live.

## Picking a contract guides input; intake still selects by fit

Keep the kernel emission `{"to": actor, "payload": value}`. Selected schema
identity and context token are UI/user-adapter metadata, not a kernel
dispatch override.

The user adapter rechecks current reach, selected schema identity and unique
fit before submission. A title match is insufficient. Stale selections,
missing contracts and overlapping schemas receive explicit rejection and
refresh. Intake remains the final authority/fit check.

Application procedure contracts use distinct wrappers such as
`conversation.send` or `conversation.load` so they select unambiguously.
A remembered task may still fall back to its original actor outside its
refinement. Selecting its star does not guarantee zero model steps.

## Submission and reconstruction are separate operations

Preserve the user actor as the holder of the conversation root and
observations and the receiver of human events. Procedure completion need
not discharge that enclosing conversation:

```text
contract star / sidebar action
    → UI request → authenticated webhost → user context
    → interpreter procedure → underlying actor operations
    → recorded return/fault → retained user observation leaves

reconstruct-state procedure
    → user checkpoint read → Horizon strand syscall → user projection
    → complete app snapshot → WebSocket delivery → app update
```

The user adapter can also invoke the history syscall directly to recover
its own state. Neither path needs an interpreter conversation-loop feature
for reconstruction. Procedures express/combine operations; the Horizon
executes them and records the results.

Invocation and parent-return routing already exist in the kernel:
`user → interpreter procedure → system.strand → interpreter → user`.
Calls share the caller's Horizon; the syscall returns to its slot and the
kernel resumes the waiting parent. The procedure returns its app snapshot
through the same mechanism. No special application routing layer is needed
for reconstruction. Treat this as verified existing plumbing, not a build
step or unresolved architecture question; see 33's kernel-path section.

Structured submissions carry the application destination, selected
contract identity, context token, request ID and JSON payload. Ordinary
text chat is translated to the default send procedure without making the
human fill internal continuity fields. Text that looks like JSON remains
text unless structured composition was explicitly selected.

Do not reuse the text-only browser contract for structured returns without
updating its schema. Preserve the existing async/custody validation and
recheck request/context at actual admission.

## The procedure's output matches Skyra.app's input

The current app accepts a `snapshot` with `uiIncarnation`, integer
`revision` and a full `conversations` array. Each conversation supplies
its ID, phase, entries and optional Horizons. Emit the browser-compatible
`requests` array and `pendingRequestId` as well.

The app replaces its conversation collection rather than merging one
conversation. Therefore a selected-conversation load ultimately publishes
a complete snapshot of all visible conversations. A one-conversation
snapshot would remove the other sidebar boxes. No partial update protocol
is assumed in the first implementation; caches and paginated UI messages
can follow later.

The reconstruct-state design uses ordinary nested kernel invocations,
Horizon reads and user projection operations, then returns the
assembled snapshot. The delivery layer unwraps the interpreter's addressed
return before forwarding the payload to the app.

For the first implementation, let the procedure format that snapshot
directly using the current app fields and existing helpers. A small
app-specific formatting helper is sufficient where templates cannot do
the transformation. A general projection framework is deferred; getting
accepted JSON onto the screen is the immediate goal.

Keep current `uiIncarnation` and publication revision supplied by the
delivery context. They are not a provider incarnation or a historical
checkpoint hash. Reject superseded reconstruction results so an older
tree cannot overwrite a newer view. A new delivery incarnation handles
host replacement as today.

Adding selectable contract stars requires an explicit protocol extension
beyond today's snapshot fields. Propose an optional per-conversation live
destination object containing a context token and actors with address and
contract identity/title/input schema. Keep it separate from historical
`horizons[].actors`. Add decoder/models and browser schemas before relying
on it: extra JSON fields being ignored is not dynamic contract rendering.

The user projection operation builds that extended snapshot from authorized
history plus current live capability evidence. The reconstruction
procedure's final output must match whichever app input version is
installed. Keep old snapshot fields compatible during rollout.

## Composer and sky changes

- Show the selected application procedure and input contract beside send.
  Ordinary text targets `conversation.send`.
- Use schema-guided fields for structured procedure inputs and a JSON
  editor for nested values/`any`. Never use a model to invent arguments
  implicitly.
- Provide keyboard-accessible contract selection alongside the sky.
- Keep historical/underlying provider contracts visually inspectable and
  distinguish them from live selectable application contracts.
- Sending is disabled while the conversation cannot admit another action.
  Draft editing and inspection remain available.
- Animate recorded calls under their actual actor/contract stars, including
  internal Codex operations beneath interpreter procedure calls. Distinguish
  plumbing from user task/tool work instead of hiding its existence.

Group transcript calls by the admitted human request and its descendants,
including fan-out and resolves. `To == "codex"` is no longer the only turn
boundary. A reconstruction/read operation itself does not start a model
turn or recursively become part of the state it reconstructs.

## Build order and acceptance

| Step | Work | Acceptance |
|---|---|---|
| 1 | Land root/leaves reconstruction and the app-compatible procedure snapshot from 33 | Existing app decoder accepts the result; all sidebar boxes survive load; no provider inference on read |
| 2 | Route ordinary actions through installed interpreter contracts, keeping user checkpoints/custody | Chat invokes send procedure; list/load/create/close/delete follow the same entry boundary |
| 3 | Extend snapshot protocol with authorized live destination schemas/context | Desktop and browser decode/retain the fields; historical contexts do not become callable |
| 4 | Add structured submission, schema-guided composer and request grouping | Selected payload uniquely fits; stale/ambiguous/missing contracts are rejected; procedure descendants attach to the right request |
| 5 | Draw contract constellations and animate actual operation paths | Stable identities/layout, keyboard access, reduced motion and inspect-only stars work |

Verify multi-leaf state reconstruction, simultaneous branch completions,
accepted joins, host replacement, lost custody, stale publication, duplicate
actions, narrowed reference scope, contract overlap and remembered-task
fallback. Keep existing single-observation remembering behavior and replay
checks. Do not rewrite pinned historical roots during migration.

## Open presentation choices

- Whether contracts expand around a focused actor or remain always visible.
  Expansion is the starting recommendation for a crowded sky.
- How much schema appears on hover versus in the composer.
- Whether direct underlying actor invocation is offered in an explicit
  developer mode. Infrastructure visibility alone is not authorization.
- How the current input-context token and optional destination fields are
  versioned across desktop/browser deployments.

This proposal changes how future calls are chosen and drawn. The Horizon
continues to own execution and history; the user adapter retains root/leaves;
interpreter procedures compose capabilities and return app-compatible data.
