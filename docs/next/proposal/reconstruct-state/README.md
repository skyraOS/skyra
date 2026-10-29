# Reconstruct-state: procedure design for Skyra.app

Status: proposed, 2026-10-03. Design artifacts, not installed application
code. The procedure sources use the current interpreter DAG format, but
depend on the proposed multi-leaf syscall and new actor contracts below.
Compiling them does not prove those operations exist or work.

## What the app accepts today

The receiver is `desktop/Sources/Skyra/Protocol.swift: ServerMessage`,
consumed by `Connection.swift: Session.receive`. It accepts three known
message types: `snapshot`, `accepted` and `rejected`. Unknown types are
decoded but ignored. A generic `state` object or `conversation-update`
message will therefore not update the app.

Reconstruction returns `snapshot`. Its required desktop fields are:

| Object | Fields |
|---|---|
| Snapshot | `type: "snapshot"`, `uiIncarnation: string`, `revision: integer`, `conversations: array` |
| Conversation | `conversation: string`, `phase: ready/pending/failed/closed`, `entries: array`; optional `boundaries` |
| Entry | `role: string`, `text: string`, `ts: integer`; optional/nullable `requestId`, optional `call` |
| Call | `to`, `arg`, `state`, `stamp`: strings; optional `result`, `fault`, `cause`: strings; optional integer `ms` |
| Boundary view | `id: string`, `root: boolean`, `actors: string[]`, `reach: {name, id}[]`, `grants: integer` |

Emit the compatible superset already sent by `webhost.broadcast`: include
`requests` at snapshot level and `pendingRequestId` on each conversation.
The browser's `frontend/lib/protocol/schemas.ts` requires those fields,
although the desktop currently ignores them. Emit `requestId: null` for
entries without a request. Use integer `revision`, `ts` and `ms` so Swift
decoding succeeds even though browser schemas accept general numbers.

Request records contain `requestId`, `status` (accepted/returned/failed/
rejected), integer `ts`, and optional `code`/`message`. Call states use
running/returned/fault. Entry roles use user/agent/call for both clients.

See [snapshot.example.json](snapshot.example.json) for an accepted shape.
Its call stamp is illustrative; real call entries retain the invoke/fault
stamp used by the existing UI. Root/leaves and provider session identifiers
are not exposed as additional application fields.

## Full replacement is a contract requirement

`Session.receive` assigns the entire `conversations` array on a snapshot.
It clears selection if the selected conversation disappears. The browser
store also replaces its conversation map and request map. Neither client
merges a one-conversation snapshot into its existing collection.

Therefore the final `reconstruct-state` output includes every conversation
visible to the authenticated app session, ordered newest first. Selecting
one conversation may trigger reconstruction, but it cannot emit only that
conversation under today's `snapshot` contract. Initially reconstruct all
visible views; caching is deferred. Summary-only listing, paged entries and
single-conversation patches require a separately specified protocol change.
This revises proposal 33's initial assumption that list/load could return
partial data straight to the existing app without changing its receiver.

The existing sidebar derives a title from the first user entry. Emptying
unselected entries to save work would lose titles as well as history.

## Procedure composition

Two finite procedure bodies use ordinary kernel calls and returns:

1. [reconstruct-state.procedure.json](reconstruct-state.procedure.json)
   runs in the authenticated application context. Its input is the current
   UI incarnation, integer publication revision and ordered conversation
   inputs `{conversation, availability}` authorized in the calling context.
   For each it invokes the interpreter's `reconstruct-conversation`
   contract directly. The kernel preserves the caller's Boundary and routes
   the nested procedure's return to the waiting parent.
   It then invokes `user.project-app-state` to assemble the returned views
   into the exact snapshot and returns that snapshot unchanged.
2. [reconstruct-conversation.procedure.json](reconstruct-conversation.procedure.json)
   runs in that conversation's Boundary. It calls `user.read-checkpoint`,
   calls the Boundary's `system.strand` with root and all observation leaves,
   then calls `user.project-state` with the tree/checkpoint and current
   availability evidence. Its return is an intermediate view bundle:
   `{ "conversation": <ConversationView>, "requests": <RequestRecord[]> }`.
   This bundle is never sent directly to the app.

The verified path is `user → interpreter → system.strand → interpreter
→ user`, with projection calls composed in the middle. `machine.Narrow`
preserves the calling Boundary and waiting parent; `stranded` returns to the
syscall frame; `settle/gather/resolve/claim` resume that parent. The
interpreter returns the completed snapshot through exactly the same path.
No special application router or operator bridge is required for it.

The draft sources no longer depend on an `application` dispatch actor.
The existing host's enumeration of separate legacy conversation roots is
a storage/deployment concern, not evidence that kernel return routing is
missing. Do not grant cross-root reads merely from stamps, or make a new
router a prerequisite for the normal in-Boundary reconstruction path.

The Boundary owns the history syscall and returns the validated tree.
The user adapter owns checkpoints and user projection semantics. The
interpreter composes those operations. A user actor may also call the same
syscall directly for recovery; a reconstruction procedure is convenient
composition, not ownership of the history capability.

The outer procedure's returned value is the app payload. The normal
interpreter return still has `{"to": <its slot>, "payload": <snapshot>}`.
The WebSocket delivery layer unwraps that addressed return and forwards
the snapshot JSON, not the kernel instruction envelope, to Skyra.app.

## First implementation: format directly for the existing app

Decision from the human: let the procedure "cheat" for now and format the
reconstructed data exactly as Skyra.app expects. A general state-projection
framework, reusable view model and formal projection architecture are
deferred. They are not prerequisites for the first working procedure.

Use the existing snapshot field names, display conventions and available
transcript/call helpers directly. Assemble known structure and constants
in procedure return templates where possible. Where the current template
language cannot transform the recorded tree, use the smallest app-specific
formatting helper needed; do not extend the procedure language just to
format JSON. The draft `project-state`/`project-app-state` calls are places
for that pragmatic formatting, not a commitment to a generic subsystem or
new service. Their implementation may simply extract existing formatting
code behind the user adapter's contracts.

The final returned snapshot, not an intermediate generic state object, is
the first-version product. Acceptance is that the app accepts the JSON and
correctly redraws its conversation boxes, messages, calls and status.
Keep the full-collection replacement and revision rules already required
by its receiver. Formatting shortcuts may fill structural defaults such
as null request IDs; they must not invent completed work or live custody.
Formalize projection only after this direct path works.

## Draft operations and formatting helpers

Use disjoint input wrappers and register their schemas in the relevant
Boundaries. `user` may remain one executable with several input contracts.

| Target / operation | Input and result |
|---|---|
| `interpreter` / `reconstruct-conversation` | Receives `{conversation, availability}`; runs the nested read/project procedure through the ordinary kernel invocation path |
| `user` / `read-checkpoint` | Receives `{conversation}`; reads the retained slot and returns `{root_invocation, leaves, contract}` as one coherent snapshot |
| `system` / `strand` | Receives `{invocation, observations, contract}`; returns the authorized merged tree and permitted Boundary metadata; preserves existing singular-observation mode |
| `user` / `project-state` | Receives `{conversation, availability, checkpoint, strand}`; pure projection to `{conversation, requests}` |
| `user` / `project-app-state` | Receives `{uiIncarnation, revision, views}`; validates bundles, preserves conversation ordering, merges request records and returns the complete app snapshot |

The availability argument comes from current host/runtime evidence, not
an inferred live descriptor in the recorded tree. Define it as structured
status such as `{input_live: bool, turn_live: bool, recovery_required: bool}`
in the registered operation schema; the draft procedure accepts it as `any`
so a source artifact does not invent recovery semantics. Projection must
validate it. Missing checkpoint fields or unrecognized availability fail
the read, not produce a fabricated ready conversation.

No projection operation invokes Codex, updates leaves, retries work or
modifies the conversation. The procedure can run on an empty reference
list: the current `each` planner supplies an empty views array and the
assembler returns a valid empty snapshot.

## Projection rules

- Derive human entries, provider answers and faults from accepted records.
  Associate requests using IDs recorded on actions/results. Reconstruct
  tool entries and their running/returned/fault states under their actual
  turn ancestry, adapting `hosts/web/calls.go` helpers rather than treating
  every interpreter invocation as a model tool.
- Renumber display entries with integer `ts` in causal/emission order,
  matching the existing `stamp` helper's sequence-number behavior. Do not
  invent wall-clock timestamps or `ms` durations from hashes. Omit `ms`
  unless reliable recorded timing is available.
- A joined result is consumed parent state; a completed unjoined child is
  still visible recorded child work. Preserve that distinction when
  deriving pending work and request completion.
- `ready` requires a live usable input continuation and no outstanding
  admitted turn. `pending` requires known active work. A terminal close
  yields `closed`; unrecoverable work follows `failed` semantics. Saved
  records alone cannot certify a live ready state. Until recovery is
  implemented, preserve the current closed/read-only behavior for lost
  historical sessions, rather than misrepresenting them as live pending.
- Set `pendingRequestId` to the known outstanding request, otherwise null.
  Request status follows recorded admission/discharge, never socket ACK
  alone. Ambiguous progress must not be labeled returned.
- Reuse Boundary projection semantics from `hosts/web/boundaries.go`. Current
  `ReadStrand` returns chain/records/opacity, not the released Boundary data
  needed by that projector. Include authorized derived Boundary metadata
  in the proposed reconstruction result or a scoped read companion.
  Never expose internals of opaque narrowed children. An empty `boundaries`
  array is acceptable only when no authorized views are available, not
  as a silent replacement of the existing sky visualization.

## Revision and delivery rules

`uiIncarnation` describes the current delivery session/host. `revision`
is its monotonically increasing integer publication sequence. They are
supplied to the reconstruction procedure, not reconstructed from an old
conversation's records or replaced with a checkpoint hash.

Both clients reject lower revisions within the same incarnation; equal
revisions are currently accepted. Allocate a new revision for every distinct
publication. Serialize reconstruction/publication or discard superseded
requests so an older checkpoint set cannot receive a newer revision and
overwrite a newer view. On host restart use a new incarnation, as today.
Transport generation checks must also reject results from a retired host.

If a conversation read fails, keep the last valid snapshot and report the
correlated operation failure. Do not silently omit that conversation, send
a partial replacement or pretend it was deleted. The existing interpreter
may return a refusal for an unhandled fault; the delivery layer must check
for a valid snapshot before publishing it. Failure publication is separate
from successful reconstruction and must not fabricate a new view.

## Verification and implementation scope

The JSON sources can be compiled with the current `interp install` because
they use only existing bindings, dependencies and `each`. Actual execution
still requires multi-leaf `strand`, the actor contracts, checkpoint storage,
registration of both procedure contracts and projection implementations.

Design checks performed: both sources compiled/installed into a temporary
store using the existing `controller/bin/interp` (two nodes for the outer
procedure, three for the per-conversation procedure). The example snapshot
passed the actual browser `ServerMessageSchema`. A check using the actual
Swift `Protocol.swift` decoder could not compile because the local Swift
toolchain/SDK reported incompatible Foundation modules; desktop runtime
decoding remains unverified. The desktop field design above comes from
reading that decoder and its existing protocol tests.

Before implementation is accepted, feed generated snapshots into the real
Swift `ServerMessage` decoder and browser `ServerMessageSchema`; test a
complete two-conversation replacement, request/call entries, stale revisions,
empty collections, partial read failure, host replacement and lost custody.
Keep the existing protocol tests as compatibility checks. No receiver
change is required for full snapshots; sidebar selection still needs the
new procedure invocation request, and partial updates would require a new
protocol explicitly.
