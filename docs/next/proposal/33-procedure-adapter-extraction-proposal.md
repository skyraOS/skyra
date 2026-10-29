# 33 — Conversation workflows through interpreter procedures

Date: 2026-10-03

Status: proposed; not implemented. Revised after the human clarified the
scope: all application operations enter through procedures exposed by the
interpreter actor. Further revised to make a conversation an invocation,
and reconstruct its state from a retained root invocation and the complete
set of accepted leaf observations. This supersedes the earlier
observation-free forward-read design. A separate conversation-state database
and caching are removed from the initial implementation.

Builds on [32](../../design/procedures.md) and
32's implementation proposal.
The earlier extraction of duplicated procedure-tool helpers remains useful,
but is a supporting refactor within this larger change.

## The intended application

The conversation list at the left of the screen comes from a procedure.
Selecting an entry invokes another procedure that loads that conversation's
state. Sending a message invokes a procedure that continues it, using the
Codex operations underneath. Creating, closing and deleting conversations
have procedure entry points too.

The human invokes contracts on the interpreter actor. The interpreter runs
procedures describing future work, calling provider and host capabilities
as needed. Each actor can reconstruct its own state through the Boundary's
history syscall; that is independent of procedure execution.
Codex supplies its exported thread and turn operations; their schemas are
implementation inputs to these workflows, rather than the human-facing API.

```text
desktop / browser / command-line caller
                  |
       interpreter application contracts
                  |
       installed procedure bodies
          /        |         \
 Boundary strand Codex       host / operator
 reads          operations  operations
```

This must result in real procedure bodies with meaningful wiring and
workflow decisions. A procedure that calls the old monolithic user actor
and returns its answer is only a migration bridge.

UI-only behavior such as draft editing, hover, scrolling and rendering stays
local. Human application actions enter through procedure contracts.
Reconstruction remains an actor's call to the Boundary, even when a load
procedure requests that actor's reconstructed view. Protocol decoding,
bounded event handling, filesystem effects and custody retain their owners.

## A conversation is an invocation; its state is its strand

The human's `create-thread` conversation invocation is the thing represented
by a conversation box in the sidebar. The box identifies that particular
invocation, not merely a contract, provider thread ID or cached transcript.

The invocation must enclose all of its turns. The user adapter retains
the conversation root and its accepted observation leaves. Procedures
invoked beneath this root describe future operations, including provider
thread creation and subsequent turns. A procedure's completion does not
terminate the enclosing user conversation.

Do not equate Codex's completed `thread/start` operation with the entire
long-lived application invocation. Provider operations belong beneath the
root represented by the conversation box. The user adapter can continue
waiting for input using its existing continuation/parking pattern; this
does not require an interpreter-owned conversation loop.

Loading state is a syscall made by an actor to its Boundary. The Boundary
returns recorded work bounded by root and leaves; the requesting user
adapter derives transcript, calls, faults and recorded provider state.
The interpreter's own use of that same syscall is for compiling procedures
from past work. Neither caller owns the syscall or its execution history.

Recorded state is authoritative. Rebuild views on each read initially.
Caching, summary indexes and materialized conversation state are later
optimizations, not prerequisites. Non-recorded resources such as live
descriptors still require an owner; rebuilding a view does not recreate them.

## Actor reconstruction rule: retain the root and accepted leaves

For each independently reconstructable actor instance/conversation, retain
a minimal checkpoint:

```json
{
  "root_invocation": "<immutable conversation anchor>",
  "leaves": ["<accepted branch endpoint>", "<another accepted branch endpoint>"]
}
```

A conversation reference is the lookup key. The root identifies its work;
the leaves form its accepted frontier. One actor serving several
conversations needs one checkpoint per conversation. A single last
observation is sufficient only when the accepted history has one endpoint.

A leaf is an accepted record stamp, not necessarily an invoke or a completed
computation. It can be an invocation, return, fault or resolve. Outstanding
and parked calls are represented by their latest accepted records. At
initialization, the root itself is the sole leaf.

Checkpoint advancement follows execution relationships:

- Sequential progress replaces the covered branch endpoint with its
  accepted successor.
- Fan-out retains every accepted branch endpoint. A child advancing does
  not discard an unfinished sibling; a partially published round cannot
  claim that all children have been accepted.
- An accepted fan-in resolve replaces the branch endpoints it joins.
  Those leaves remain reconstructable through the resolve's recorded
  dependencies. Unrelated open branches remain in the set.
- A terminal return/fault becomes the endpoint of the work it discharges.
  Distinguish a child fault from termination of the whole conversation.
- Normalize the set by removing duplicates and endpoints already covered
  by another retained endpoint's dependency closure. Sort stamps for
  stable checkpoint encoding; execution order comes from recorded
  dependencies, not that sorting.

Reconstruct the recorded tree by walking backward from every retained leaf
to the root and following resolve dependencies. Merge the closures and
deduplicate shared ancestors. The records retain branch topology and
emission order; the checkpoint need not copy that tree.

Keep the root/frontier durably, not a second transcript, phase or provider
mapping database. Summary/view caching remains deferred. This reconstructs
accepted recorded work through the complete frontier. It does not recover
unobserved external effects or prove that a saved live resource still exists.

A recorded child result can appear in the tree before its parent has joined
it. Preserve that distinction in projection: show the child outcome, but
do not treat it as input already consumed by the parent's next decision.

## Use `system.strand` for bounded tree reconstruction

Today `core/machine/system.go` requires one `observation` and `contract`.
`runtime/replay.ReadStrand` climbs from that observation to its requesting
invoke and collects predecessor chains and resolve discharges.

Keep the existing single-observation form for remembering. Add a
root-plus-leaves form for reconstructing an actor's accepted tree:

```json
{
  "to": "system",
  "payload": {
    "strand": {
      "invocation": "<root invocation>",
      "observations": ["<accepted leaf>", "<another accepted leaf>"],
      "contract": {"target": "interpreter", "schema": "<actual create contract object>"}
    }
  }
}
```

The illustrative schema string is an object on the wire. Require exactly
one form: existing `observation`, or `invocation` plus a nonempty
`observations` set. Do not accept an observation-free descendant search.
Verify every supplied leaf belongs beneath the anchor and that the anchor
was invoked through the named contract.

Reconstruction requirements:

- Preserve requester-Boundary visibility and narrowed-child opacity on
  every branch. Knowing stamps grants no authority. Cross-root dispatch
  establishes scoped read context, not an arbitrary caller-supplied path.
- Reuse predecessor/resolve-discharge traversal and replay validation
  across all leaves, with shared memoization and merged record closures.
  A multi-leaf result cannot pretend to be one linear `Chain`; define a
  versioned tree result with root, frontier, records and opacity.
- Support a frontier containing only the root at initialization and
  nonterminal frontiers for open conversations.
- Validate acceptance, ancestry and compatibility of all supplied leaves.
  Checkpoints come from accepted runtime events, not newest matching files
  or losing resolve candidates. Retaining leaves does not bypass acceptance
  or the rule that a fan-in has one winning continuation.
- Stop at the retained frontier while newer work continues. Reconstruct
  the same tree from the same root/frontier; derive causal and emission
  order from records, not file times or sorted hashes.
- Keep existing observation-mode `remember` output compatible. Conversation
  projection belongs in the requesting actor (the user adapter for this
  conversation), not the interpreter. Traversal stays provider-independent.

Update `core/machine/system.go`, `ReadStrand` effect/dispatch in
`core/machine/machine.go` and `runtime/instance/instance.go`, and
`runtime/replay/strand.go` for this additional reconstruction form.
The user adapter owns the conversation's root/leaves slot. It records all
accepted observations it receives and calls `system.strand` to reconstruct
its own state. Hosts resolve UI references to authorized roots without
inventing leaves. Provision actor-owned storage if that slot must survive
process/host replacement; do not move its bookkeeping into the interpreter.

Listing needs scoped discovery of conversation references/checkpoints.
Reconstruct summaries from each root/frontier. A checkpoint directory is
required recovery metadata, not a materialized conversation index or
transcript cache.

## Verified syscall ownership in the current code

- `controller/interp/main.go: runRemember` emits
  `{"to":"system","payload":{"strand":...}}` on its first step. On resume
  it unmarshals the returned strand and calls `remember.Compile`.
- `controller/core/machine/system.go: Syscalls` defines the Boundary's
  `system` surface with `narrow`, `extend-reach` and `strand`.
- `controller/runtime/boot/boot.go` adds that surface to genesis reach.
  `system.go: kept` retains `system` even when a child Boundary narrows reach.
- `controller/core/machine/machine.go: release` recognizes the system
  call and emits `ReadStrand`, using the requester's Boundary context.
- `controller/runtime/instance/instance.go` executes that effect with
  `runtime/replay.ReadStrand` and returns `StrandRead` to the machine.
- `machine.tools/contracts` expose the reached system schemas in actor
  continuation envelopes. No check reserves `strand` for the interpreter;
  any actor in that Boundary can invoke it, subject to contract/reach and
  history-visibility checks.

The interpreter is one syscall client. Procedures handle future work;
the Boundary handles execution and records what happened. User-state
reconstruction requires no procedure-language feature.

## Existing kernel path: not an implementation challenge

Use the ordinary invocation/return path already implemented:

```text
user invokes interpreter.reconstruct-state
    interpreter invokes system.strand in the caller's Boundary
        Boundary returns the tree to the syscall's slot
    kernel resumes the interpreter with that return
    interpreter invokes user projection and receives its return
    interpreter returns the assembled snapshot to its own slot
kernel resumes the waiting user adapter
user adapter delivers the snapshot through the existing app transport
```

`machine.Narrow` retains the caller's reach, references and authority and
stores its waiting parent in `Resume`. `stranded` emits the syscall result
addressed to its own frame. `interpret` derives that as a return;
`settle/gather/resolve/claim` discharge the slot and resume its parent.
`runProcedure` likewise returns to its own frame, delivering its result to
the waiting caller. Parent routing is the kernel's existing responsibility.

Do not introduce a special application router or operator bridge for this
path. The syscall caller is already in its Boundary; a return follows its
parent automatically. Actual app serialization/delivery remains transport
work, not a new kernel routing rule. A separate deployment concern about
enumerating legacy storage roots must not be confused with this call path.

Verified by reading these functions and passing
`go test ./controller/core/machine ./controller/runtime/replay`, including
`TestStrandIsReadThenReturned` and the continuation/fan-in tests. Remaining
work is multi-leaf history input, user checkpoint/projection operations,
procedure registration and app-compatible output, not return routing.

## What the current code actually does

The following is based on reading the controller, both UIs, the launcher,
and schemas exported by the installed, pinned Codex CLI.

| Current behavior | Code | Consequence |
|---|---|---|
| Select a saved conversation locally | `desktop/Sources/Skyra/Views/Sidebar.swift`, `Session.selected` in `Connection.swift`; `frontend/components/Sidebar.tsx`, `selectConversation` in `frontend/lib/store.ts` | Selection sends no request and runs no procedure |
| Load all past transcripts at host startup | `controller/hosts/web/main.go: loadPast` | Reads conversation roots and the latest browser transcript; marks past conversations closed |
| Send every full conversation in a snapshot | webhost `broadcast` | Sidebar listing and loading details are coupled; there is no separate load operation |
| Create a conversation and seed its root | webhost `handle(open-conversation)`, `seed` | Host chooses ID, scratch path, boot arguments and initial target `user` |
| Submit and close | webhost `handle(submit, close-interaction)` | Checks ready state, mutates transcript/phase and resumes a browser frame with text or `/close` |
| Reconstruct transcript and choose the next action | `controller/targets/user/main.go` | User actor invokes Codex directly, carries thread/incarnation/activity state, then parks on browser again |
| Decide process/thread/turn lifecycle | `controller/codex/resident/main.go: decide` | Starts or inspects a process, creates a thread or reuses its ID, starts a turn, runs tools and reports the final result |
| Delete local history | webhost `deleteConversation` | Writes tombstone, removes conversation directory including scratch/provider files, closes parked descriptors |
| Derive live calls and their turn grouping | `controller/hosts/web/calls.go: derive, isTool` | A turn is recognized by an invocation addressed to `codex`; adding interpreter calls changes this assumption |
| Install and execute a procedure | `controller/interp/main.go: install, runProcedure`, `proc/source.go`, `proc/plan.go` | Finite DAGs, bindings and `each`; no conditional branch, repeat or explicit fault handler |
| Keep a resident connection | `core/machine.resourceOwner`, `runtime/custody`, `runtime/park`, `runtime/instance/appserver.go` | Resources belong to an ancestor frame selected by a configured depth; descriptors are not durable JSON state |

The production launcher `skyra` registers `codex` as the resident adapter
and registers `codex.start/exchange/inspect/stop` separately. It registers
the interpreter with only its `remember` contract. There are no installed
application procedure contracts in production.

The legacy `controller/codex/adapter` speaks through the actor host and
remains in `controller/run.sh`. It is not the production conversation path.

## Codex operations to build upon

The controller pins `codex-cli 0.160.0`
(`runtime/instance/appserver.go: PinnedCodex`). Its installed binary was
used to generate JSON Schema with:

```sh
codex app-server generate-json-schema --experimental --out <temporary-directory>
```

Inspected exports include `ThreadStartParams`, `ThreadListParams`,
`ThreadReadParams`, `ThreadResumeParams`, `ThreadTurnsListParams`,
`ThreadItemsListParams`, `ThreadUnsubscribeParams`, `ThreadArchiveParams`,
and turn-start/interrupt and dynamic-tool call schemas.

The [official app-server documentation](https://learn.chatgpt.com/docs/app-server)
describes the underlying thread/turn interface. Implementation must use
schemas generated from the pinned version, with fixtures and version
metadata, rather than assuming current documentation exactly matches it.

Important facts from those local exports:

- `thread/read` requires `threadId`; `includeTurns` is optional. Its
  description recommends paginated turns/items for full history.
- `thread/resume` requires `threadId`. It can override approval and sandbox
  policy. It does not expose a `dynamicTools` field in this pinned export.
- Resume's `history` field is explicitly marked for Codex Cloud, not local
  use. Its rollout `path` alternative is unstable. Do not use either as a
  generic shortcut for restoring local conversation state.
- `thread/list` is provider-thread discovery. It is not the application's
  list of conversations: application history also includes faults,
  procedure calls, scratch and Boundary context.
- `thread/unsubscribe`, archive, local deletion and process stop have
  distinct meanings. Closing a Skyra conversation must not accidentally
  archive or delete its provider thread.

Generated JSON Schema uses unions, references and other features absent
from Skyra's simple `contract.Schema`. Add supported, versioned projections
of the required operations and validate their actual provider payloads in
the Codex actor. Do not blindly load the whole schema bundle as kernel input
contracts or widen the kernel schema language for this migration.

## Human-facing procedure contracts

The concrete reconstruction design and draft DAG sources are in
[Reconstruct-state for Skyra.app](reconstruct-state/README.md). Its final
return matches the existing app `snapshot` input: full conversation
collection, UI incarnation and integer revision, plus browser-compatible
request fields. The existing app replaces snapshots wholesale; partial
list/load results are internal until a separate partial-update protocol
is implemented.

First-version decision: the reconstruction procedure may format directly
for Skyra.app's current snapshot shape, reusing existing formatting code
and a small helper only where templates need it. A general state-projection
framework is deferred. The immediate acceptance test is valid JSON that
updates the app correctly; the draft projection calls do not require a
new generic subsystem.

Register disjoint, titled input contracts on `interpreter`. Use an
operation wrapper so selection by structural fit remains unambiguous:

```json
{"to":"interpreter","payload":{"conversation.load":{"id":"c-123"}}}
```

The title names the operation for display; the payload wrapper makes its
contract distinct. Human payloads contain conversation IDs and ordinary
input, never provider IDs, OS paths, resident descriptors or frame stamps.

| Procedure | Input inside its wrapper | Return / behavior |
|---|---|---|
| `conversation.list` | optional cursor and limit | Reconstructed summary page: invocation reference, title, phase, frontier and continuation availability |
| `conversation.load` | ID, optional history cursor | Reconstructed strand projection and retained frontier; no inference turn |
| `conversation.create` | optional title/configuration | Establishes the enclosing invocation/root/scratch; publishes reference/progress while awaiting human input |
| `conversation.send` | ID, text, request ID, expected frontier | Routes an action into the enclosing workflow; its next turn and outcome remain descendants |
| `conversation.close` | ID, request ID, expected frontier | Ends application interaction and releases its live session when appropriate; preserves readable history |
| `conversation.delete` | ID, request ID, expected frontier | Preserves the existing local-deletion meaning and tombstone protections |
| `conversation.refresh` | ID and last frontier | Returns committed changes for an active view; does not submit another turn |

List/load/refresh are reads. Mutations carry request IDs and an expected
frontier; admission is serialized by the enclosing workflow and runtime/host
mechanisms. Use accepted action records to recognize duplicates. A
transport acknowledgement means receipt, not successful execution. Do not
introduce a separate durable ledger merely to cache those results.

Selecting a sidebar item always invokes `conversation.load`, even if the
UI can immediately show cached content. A late response for a previously
selected item must not replace the currently selected view.

Loading and continuing are separate. An old conversation can be readable
even when its provider session cannot be recovered. If continuation is
unavailable, show the reason; sending must not silently create a fresh
conversation with an empty provider context. Explicitly closed conversations
stay closed unless a future reopen operation is specified.

## Concrete workflow bodies

`conversation.list`: request authorized conversation references and
reconstructed summaries from the user adapter/host-facing actor. That actor
uses its root/leaves slots and the Boundary syscall. The procedure binds the
returned summaries into the UI result; no summary cache is required.

`conversation.load`: trigger `reconstruct-state`, whose per-conversation
procedure calls `user.read-checkpoint`, the Boundary's `system.strand`, and
`user.project-state`. The outer procedure assembles all visible views into
the app's full snapshot through `user.project-app-state`. The Boundary owns
history access, the user adapter owns projection/checkpoint semantics, and
the interpreter composes these operations. The adapter may also reconstruct
directly for its own recovery.

`conversation.create`: request scoped root/scratch provisioning and
establish the user adapter's enclosing conversation invocation. Invoke
provider-thread creation through the installed procedure beneath that root.
The procedure can complete while the user adapter remains waiting for input.

`conversation.send`: route the structured action into the user adapter's
conversation. It reconstructs state as needed via its Boundary, then invokes
the appropriate future-work procedure for start/resume/turn operations.
Provider results become observations retained by the user adapter.

`conversation.close`: request future session release operations through a
procedure and discharge the user adapter's conversation when closing is
complete. Recorded closure is reconstructed like other historical state.

`conversation.delete`: authorize the exact reference, refuse active work,
request resource release and local cleanup, and preserve tombstone/late-park
protections. A deletion tombstone is control metadata, not a view cache.

`conversation.refresh`: request the adapter's view at its latest retained
frontier. Hosts forward notifications or trigger refresh for subscribed
views. The actor uses the same syscall independently of procedure wiring.

A reconstructed read may be composed into a human-facing procedure, but
must remain directly usable by the actor itself for recovery. It does not
need to invoke the interpreter to learn what happened.

## State and host boundaries

Remove the previously proposed `controller/targets/conversations` state
actor, persistent conversation index, duplicate transcript/phase store and
mandatory mutation-ledger database from this implementation plan.

The interpreter owns procedure definitions and its own execution positions.
The user adapter owns the conversation root/leaves slot and reconstructs
its own state through the Boundary syscall from invocation records,
returns, faults and workflow inputs. Ensure required facts are recorded:
provider thread/session identity, scoped provider-home reference, catalog
identity, workflow version, and human action request IDs. If a fact is not
currently in those records, add it to the owning actor's operation payload
or result rather than hiding it in a parallel conversation database.

Live-resource ownership, exclusion of concurrent turns and request
deduplication still need explicit rules. Use the enclosing workflow's
serialized event intake, checkpoint advancement and accepted records/runtime
claims where possible.
Do not assume history projection itself provides an atomic admission lock.
A host's in-memory request set is not durable proof of a completed action.

Retain a narrowly scoped operator/session service for root/scratch
provisioning, invocation-reference discovery, exact cleanup and live
descriptor custody. It must not own conversation workflow policy or
materialized transcript state. It may initially be implemented by
refactoring the existing host instead of introducing a new daemon.

The web host retains authentication, WebSocket framing, generation checks,
request correlation, delivery and host-held parked-frame custody. It routes
actions through interpreter contracts and publishes procedure progress.
It no longer independently chooses conversation lifecycle transitions.

Today `prepareAppserver` allocates a fresh private home under
`providers/actor-*` for each process. Record and reuse the scoped provider
home needed to resume its saved thread. A thread ID reconstructed from
lineage alone does not make an empty new home contain the provider history.

A management root/bootstrap context is still needed for list/create/load
before a conversation exists. It has scoped discovery/read/provisioning
capabilities. Dispatching work into a conversation must use that
conversation's Boundary; management grants never flow into model work.

Reconstructing a conversation after a host restart makes it readable. It
does not automatically revive its enclosing invocation or old descriptors.
Continuation requires a tested runtime recovery/rebind mechanism that
preserves the invocation anchor and accepted ancestry. Report unavailable
continuation until that mechanism exists; do not substitute a fresh
unrelated conversation or silently replay an uncertain turn.

## Interpreter changes needed for future-work composition only

Reconstruction is not an interpreter feature. Do not add tree-reading,
user-state checkpoint ownership or a conversation loop to `interp/proc`
to solve it.

The current finite DAG supports fixed future-work sequences, bindings,
fan-out and `each`. Use that existing capability first for installed
operation procedures. The user adapter retains its event/parking loop and
invokes finite procedures beneath the conversation root.

Required provisioning: installed procedure bundles with manifests,
declared contracts, content hashes and startup versioning. `interp install`
currently installs one source into its store; it does not register that
contract in a Boundary. Bundle installation and published contracts must
agree before a human entry point is exposed.

Conditional operation selection and cleanup may eventually require
workflow branches or explicit fault handlers. Identify a concrete future
procedure body needing them before extending the language. These would
control future work, not reconstruct the past. Preserve current remembered
DAG IDs and learned-refinement behavior when adding such features.

Hand-installed workflows already exist in `interp/examples` and `run.sh`.
Support authored application procedures alongside remembered ones, and
update 32's remembering-only wording accordingly. Keep Codex's bounded
protocol/event handling inside its owning adapter.

## Codex adapter changes

Expose supported provider operation contracts, provisionally
`codex.thread.start/read/resume/list`, `codex.thread.turns/items`,
`codex.thread.unsubscribe`, `codex.turn.run/interrupt`, and session
initialize/inspect/release operations. These are internal capabilities
called by procedures, not the desktop's application contracts.

Split `codex/resident/main.go` by responsibility:

- Extract request/response schema projections, validation, IDs and RPC
  correlation into a provider protocol package.
- Retain bounded turn handling: provider events, dynamic calls, tool-result
  correlation, terminal confirmation and native-attempt interruption.
- Move the choices now in `decide` about idle replacement, starting versus
  reusing/resuming a thread, and returning/persisting session identity into
  the application procedure bodies and session service.
- Preserve `delivery: unknown` as an uncertain effect requiring
  reconciliation. Procedure retries must not resend that turn blindly.
- Share the existing `remember`/procedure discovery and translation code
  between legacy and production paths as originally proposed. Retain model
  call indexing independently of transport operations.

Freeze tool definitions and name-to-target bindings per provider thread.
The pinned resume schema cannot supply a new dynamic-tool catalog.
A resumed thread's usable catalog must be recovered and checked against
current reach; intake still checks each emitted instruction. A catalog
mismatch returns a defined continuation-unavailable result until a tested
migration/fork operation exists. Do not invent tools at resume.

## Change, remove, add inventory

| Location | Change / removal | Addition or retained responsibility |
|---|---|---|
| `controller/interp/main.go`, `interp/proc/` | Provision installed procedures; add control only when a concrete future-work body requires it | Bundle installation; no user-state reconstruction or mandatory conversation repeat feature |
| New `controller/interp/workflows/conversations/` | — | Authored procedure sources, manifest and public contracts for list/load/create/send/close/delete/refresh |
| `core/machine/system.go`, `machine.go`, `runtime/instance/instance.go`, `runtime/replay/strand.go`, `runtime/tree` | Preserve single-observation mode; add root plus nonempty observations set | Multi-leaf reconstruction, merged closures, acceptance/boundary validation and checkpoint integration |
| Operator/session host, provisionally `controller/hosts/application/` or refactored existing host | No separate conversation-state actor/database | Management bootstrap, scoped invocation discovery/dispatch, provisioning and resource custody |
| `controller/hosts/web/main.go` | Remove `loadPast`, `seed`, `deleteConversation` and the lifecycle branches in `handle` after cutover; replace all-history `broadcast` | Generic interpreter invocation gateway, correlated results and subscriptions; retain authentication/transport/custody mechanics |
| `controller/hosts/web/calls.go`, `boundaries.go` | Move reusable projection code; remove `To == codex` as the sole turn delimiter | Project in the user adapter from Boundary-supplied recorded work, with workflow/system calls distinguishable from model tools |
| `controller/targets/user/main.go`, `contract.json`, `targets/browser/contract.json` | Replace direct Codex orchestration with future-work procedure calls; introduce structured actions | Retain user continuation/parking; add root/leaves slot and direct Boundary syscall reconstruction with actor-owned state access |
| `controller/codex/resident/main.go`, `contract.json` | Split monolithic text-input orchestration; remove migrated lifecycle choices once procedures own them | Versioned exported operation projections and bounded turn actor |
| `controller/codex/io/main.go`, `contract.json` | Adapt only as needed for session-service access | Keep bounded I/O, partial lines, uncertain delivery and EOF semantics |
| New `controller/codex/protocol/`, shared procedure-tool helper | Remove duplicate schema/tool-binding helpers from adapters | Pure schema projection and model-facing procedure binding |
| `controller/codex/adapter/`, `hosts/actor/`, `run.sh` | Migrate shared helpers; later remove legacy conversation orchestration only after fixtures are replaced | Keep legacy transport demos until production replacement passes equivalent tests |
| `controller/runtime/instance/appserver.go` | Replace always-fresh provider-home allocation with validated session-home provisioning | Preserve pinned version, proxy verification and sandbox construction |
| `controller/core/machine/machine.go`, `core/boundary/`, `runtime/custody`, `runtime/park` | Audit ancestry-based resource ownership across added interpreter/service frames | Preserve validated custody; change owner targeting only if required by the session design |
| Root `skyra` launcher and `controller/runtime/registry` | Stop registering only `remember` for the interpreter; provision user root/checkpoint storage and procedure call path | Provision bundle/contracts before genesis, pin all binaries and manifests, register scoped service capabilities |
| `desktop/Sources/Skyra/{Connection,Protocol}.swift`, `Views/{Sidebar,Composer,ConversationView}.swift` | Remove local-only sidebar load and closed/failed → silently new conversation behavior | Procedure invocation requests, list summaries, loaded detail, explicit unavailable/error state and correlated create result |
| `frontend/lib/{store,ws-client}.ts`, `protocol/{types,schemas}.ts`, `components/{Sidebar,Conversation}.tsx` | Same application routing migration | Explicit load on selection, cached views keyed by ID/revision, late-response protection |
| `frontend/mock/`, controller web/Codex tests, desktop protocol/live tests | Replace assumptions about direct Codex submission and full-history snapshots | Procedure contracts and end-to-end workflow fixtures |
| `controller/README.md`, 23/32 and proposal index | Update ownership, bootstrap, authored-procedure and history-loading descriptions | Explain procedure-facing application and remaining low-level actors |

Do not blanket-delete the resident adapter, user actor or legacy host at
the start. Delete production routes and migrated branches once their
replacement is verified; remove whole legacy binaries only when no pinned
conversation, demo or test still depends on them.

## Custody and authority are implementation gates

Adding interpreter wrappers changes the ancestor chain.
`machine.resourceOwner` walks `Surface.ResourceOwnerDepth`;
`skyra` computes that depth for the current chat/ask layout. Simply adding
one frame or incrementing a constant everywhere is insufficient when
list/load/send and service dispatch have different shapes.

The proposed session service owns live descriptors and accepts only its
scoped session handles. Specify how a procedure invocation obtains access
to that owned session before moving lifecycle code. Prefer host-mediated
operations using existing async/custody mechanisms; if this requires
replacing depth-based ownership with an explicit validated owner
capability, document and test that kernel change independently.

Never put a descriptor in interpreter state, share one live connection in
fan-out, revive a dead parked frame from its stored stamp, or give a model
the management service's cross-root grants. Listing/loading does not resume
a parked frame. Continuing after a host restart must recover/rebind the same enclosing
conversation invocation through validated runtime mechanisms; uncertain
in-flight turns require reconciliation rather than assumed recovery.

## Remaining implementation unknown: frontier publication and recovery

Root plus all accepted leaves supplies bounded reconstruction without
discovering a newest record. The remaining question is advancing the whole
leaf set durably in step with runtime acceptance.

Parallel children must not overwrite one another's frontier updates. A
join removes only the leaves covered by its accepted dependency closure;
a stale child writer must not restore a leaf superseded by that join.
The user adapter is the owner of this slot and retains both observations
when two branches finish together. Serialize its slot updates or use a
compare-and-update rule for concurrent adapter instances; do not assign
conversation checkpoint ownership to the runtime or interpreter.

A crash can occur after acceptance but before checkpoint publication.
Atomic replacement of the checkpoint file does not make it atomic with
record publication or provider effects. Recover any accepted progress
missing from the saved frontier using runtime acceptance/claim evidence,
or report unresolved work rather than blindly repeat it.

The first experiment tests sequential advancement, concurrent fan-out,
partial rounds, child discharges, winning joins, stale writers and crashes
between acceptance and publication. This is recovery bookkeeping, not
view caching. Reconstructing recorded progress still does not revive live
descriptors, parked frames or provider connections.

## Delivery sequence and proof

1. **Checkpoint/reconstruction experiment.** Define root invocation and
   complete accepted-frontier storage, advancement and reconciliation. Test bootstrap,
   live/parked reads, fan-out joins, stale writers, orphan candidates and
   crashes between acceptance and checkpoint publication.
2. **List/load through the interpreter.** Add scoped discovery/bootstrap,
   the user adapter's Boundary syscall and actor-owned conversation projection.
   Both sidebars call list/load procedures. Reconstruct on demand with no
   database or persistent cache; prove reads start no provider turn.
3. **Enclosing user invocation and procedure provisioning.** Preserve the
   user adapter's input/parking loop, root/leaves slot and direct syscall
   reconstruction. Install finite future-work procedures beneath that root;
   sends/turns are descendants and close discharges the user frame.
4. **Provider operation extraction.** Split resident operations, preserve
   bounded turn events and safety checks, record provider-home/thread and
   catalog identities needed for reconstruction.
5. **Send and live continuation.** Route actions through procedures and the
   enclosing frame. Prove multiple turns share its anchor/provider thread,
   concurrent/duplicate actions do not run duplicate turns, and refresh
   shows faults or completed work from records.
6. **Recovery and cleanup.** Test authorized rebind of the same conversation
   after host/process loss, saved provider-home access, unavailable history,
   uncertain delivery, close/delete and late-park tombstone protections.
7. **Remove old production paths.** Retire duplicate webhost lifecycle
   branches, direct user → Codex submission and obsolete UI messages.
   Migrate legacy fixtures before deleting their binaries.

Acceptance tests must also cover:

- Missing/empty observations or mixed single/multi forms are refused.
  Each leaf's acceptance, root ancestry and contract are checked. The
  initial root-only frontier and nonterminal reads work; remembering stays
  compatible with the existing single-observation form.
- Cross-Boundary reads are refused and narrowed child internals stay opaque.
  Cross-root references require scoped authorization.
- Multi-leaf reconstruction excludes losing/orphan branches, deduplicates
  shared ancestors and preserves emission order. A slow/open branch stays
  represented when a sibling advances. Child completion and parent join
  remain distinct in the projection.
- Repeated load reconstructs the same tree at the same root/frontier.
  Accepted joins replace their covered leaves without losing unrelated
  branches; discarded leaves remain reachable through join dependencies.
  Load does not include itself.
- Frontier writes cannot lose concurrent branches, regress after joins or
  publish unaccepted work. Crash-gap recovery does not blindly repeat an
  already accepted action.
- The UI, command-line and test clients call the same interpreter entry
  contracts; late load results cannot change the selected conversation.
- Turn faults, native attempts, provider EOF, catalog/policy mismatch and
  uncertain delivery produce recorded outcomes without blind retries.
- Old roots load through a read-only projection adapter without rewriting
  pinned records. Cached/indexed state is not needed to recover the view.
- Added interpreter frames preserve authority and descriptor ownership.
  Reads alone never claim a dead parked frame is live.
- User conversation actions and future-work procedures stay beneath the
  user root. User reconstruction works through `system.strand` without
  calling the interpreter. Existing remembered DAGs/refinements still pass.

Run controller and separate interpreter suites, fake-provider integration,
desktop protocol/live and browser mock tests as each stage lands. Inspect
individual `run.sh` scenarios because the script uses `|| true`.
A real pinned-Codex restart/resume test is required before claiming provider
continuation works; generated schemas prove only the interface.

Caching is explicitly deferred. Optimize discovery and reconstruction only
after the uncached flow is correct and measured.

## Relationship to the other proposals

This revises H8 from merely offering remembered procedures to making
procedures the application's entry points. J22/J23's missing model
`resume`/`refine` contracts concern learned-procedure fallback and
refinement authoring; they are different from Codex `thread/resume` and
are not resolved by this change.

[34 — Actor constellations](34-actor-constellations-proposal.md) is aligned
with this proposal: the default application destination is the interpreter's
`conversation.send` contract, and sidebar actions address its other procedure
contracts. Reconstruction returns the existing full app snapshot, with an
explicit protocol extension needed to render selectable contract schemas.
Constellations may still expose actor capabilities for explicit inspection
or developer use; the application workflow must not depend on direct human
submission to the Codex actor.
