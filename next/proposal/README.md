# Proposals: every change the new design asks for

Date: 2026-09-30  
Status: proposal. Collects what [24](../24-the-room.md),
[25](../25-surface-refactor.md), [26](../26-agents-and-facets.md) and
[improvements](../../improvements.md) items 1, 2, 8 and 9 change, in the code
and in every other document under `next/`.

Each `next/` document has a `<name>-proposal.md` here saying what it needs:
code changes where the document owns them, and edits to the document where
the new design supersedes it. This page is the complete list, each change
once, with an id the per-document proposals refer to.

## The design in one paragraph

Authority is derived from state plus proposal at every step. A Horizon is an
agent's execution context: authority only, identified by its authority, the
invoke that created it and its input contract, committed durably and handed
over as a hash the child looks up. Every call shares its caller's Horizon
unless the caller narrows (through the room's `system.narrow` facet) or calls
another agent (whose Horizon it runs under). The room is the root Horizon,
whose Horizon is the registry table. Returns address the slot; unaddressed
output is a leaf fault; faults carry their observation and climb to whoever
can grant. A surface is identity plus contracts. Facets group an agent's
contracts; reach holds facets.

## The complete list

### A. States: returns address the slot (improvements 1, 2)

| Id | Change | Where |
|---|---|---|
| A1 | `interpret` derives return from the slot: `to` equal to the emitter's frame id is a return; a name in reach is an invoke; anything else is a horizon fault | `core/machine/machine.go:285–322` |
| A2 | Unaddressed output is a `leaf` fault carrying the output, not a string return | `machine.go:292`, `core/record/record.go` (new kind `Leaf`) |
| A3 | Drop the `"reply"` reservation | `runtime/registry/registry.go:56` |
| A4 | Tell every target its slot: `SKYRA_FRAME=<frame id>` in the environment (continuations already see `frame` on stdin) | `machine.go:569` (`runOf`) |
| A5 | Every emitter of `"reply"` addresses its slot instead | `targets/read/main.go:65`, `codex/io/main.go:50`, `codex/resident/main.go:391`, `codex/adapter/main.go:234` (legacy), `runtime/instance/instance.go:380`, `hosts/web/main.go:314` (uses the parked instance's frame), `hosts/web/calls.go:220` |
| A6 | Faults carry their observation: the bytes that did not fit, the address outside the Horizon, what a program printed; the continuation's `StepView` exposes it | `record.FaultInfo`, `machine.go:627` (`StepView`), `machine.go:635` |
| A7 | Codex's native attempt is returned as an observation (a typed `refusal` in the resident's output contract), not `native.<type>` addressed to trigger a fault | `codex/resident/main.go:367`, `codex/resident/contract.json`, `codex/adapter/main.go:292`, `hosts/web/calls.go`, `hosts/web/calls_test.go` |
| A8 | The fake Codex's `"tool": "reply"` bypass probe becomes a probe of naming `reply` (now a horizon fault) and of guessing a slot | `codex/fake/main.go:133` |
| A9 | Tests: `TestInterpret`, replay, `read_test`, `resume_test`, `lineage_test` | `*_test.go` |

### B. The Horizon: durable, derived, shared (24)

| Id | Change | Where |
|---|---|---|
| B1 | The `Horizon` struct holds authority only (reach, grants). `Stamp`, `Frame` and `Resume` move out | `core/horizon/horizon.go:126` |
| B2 | Horizon id = hash(authority, the invoke that created it, its input contract) | `horizon`, `record` |
| B3 | A Horizon is committed durably at creation, and that commit is the release | `runtime/tree` (`horizons/`), `machine.handoff` (`:413`) |
| B4 | The child is handed the Horizon's hash; its first act is a lookup, and it re-derives what it found | `machine.accept` (`:330`), `instance` `SendHorizon` (`instance.go:177`), `handshake` |
| B5 | The stack lives with the Horizon in durable state; steps read it and write write-once entries. The `Resume` copy goes | `machine.gather`/`claim` (`:509`, `:532`), `instance.gather` (`instance.go:242`), `Depth()` |
| B6 | Every call shares its caller's Horizon; a Horizon is created only by a narrowing, a registration, or a call into another agent | `machine.Narrow` (`:456`) |
| B7 | The Seatbelt profile is rendered per Horizon, plus the exclusions: the floor, and `process-exec` of the running program only. Exec can never be registered | `horizon.Render` (`:157`), `resident.go:21` |
| B8 | The room: the derivation's root. Genesis is its creation, authored by the operator and carrying a nonce | `runtime/boot/boot.go:47` |
| B9 | Crash recovery by reading state: a record no derivation accepted is an orphan; an accepted step with no successor runs again; effects are at least once | `boot`, `webhost` (retires the `533d98d` rule) |
| B10 | Resume re-derives the instance from the tree; the host holds a hash, not authority (closes 20's finding 1) | `skyra/main.go:107`, `runtime/park` |
| B11 | Replay checks that every mark is backed by a record (improvements 8, direction 1); after that, frame fds can become hashes | `runtime/replay` |

### C. The surface: identity plus contracts (25)

| Id | Change | Where |
|---|---|---|
| C1 | Delete `Surface.Reach`, `Authority` and `Delegate`; they are the Horizon's | `horizon.go:85`, `registry.go:73,116,148,151`, `machine.Narrow`, the `"reach"` line in four `contract.json` files |
| C2 | Delete `Leaf()`; the stdin format (continuation or bare payload) becomes an explicit property of the surface | `horizon.go:111`, `machine.go:359` |
| C3 | Delete `Surface.Args`. **Corrected:** not unused; the launcher passes the real app-server's `--disable` flags through it, so it moves with C4 | `horizon.go`, `registry.go:70,79,148`, `instance.go:361`, `appserver.go:178` |
| C4 | Later: move the resident and network fields off the surface; network grants to the Horizon, keep-alive to control flow (a loop operator) | `horizon.go:92–100`, `instance.startResident`, `runtime/custody` |

### D. Agents and facets (26)

| Id | Change | Where |
|---|---|---|
| D1 | The room's public facets `system.narrow` and `system.register`, implemented by the runtime and unregisterable; an admin facet in the human's reach | `machine`, `registry` |
| D2 | Narrowing is two steps: `system.narrow` returns a reference, then the caller invokes it | `machine.interpret`, `Narrow` |
| D3 | Reach becomes a c-list: references carried by a return are recorded by the room and added to the receiver's reach | `machine`, `record` |
| D4 | Registration presents facets; the registrant gets them back as references | `registry` |
| D5 | A call to another agent's facet runs under that agent's Horizon | `machine.Narrow` |
| D6 | Codex as the first agent with facets (creator: start/exchange/inspect/stop; public: the turn), replacing `ResourceOwnerDepth` | `codex/resident`, `registry-codex.json` |

### E. Identity (improvements 9)

| Id | Change | Where |
|---|---|---|
| E1 | One written canonical encoding for records and declarations | `core/record`, `core/contract` |
| E2 | Domain-separated hashes: type, content, record, frame, Horizon | `record.Stamp` |
| E3 | Two hashes per step: content id and stamp | `record`, `replay` |
| E4 | Ids as the path of handoffs, segmented by frame; the storage layout follows | `runtime/tree` |
| E5 | Types and principals as records; the runtime derives a record's type | `registry`, `record` |
| E6 | The recipient on return and fault records (improvements decision 5, 2b) | `record`, `machine.faulted` |
| E7 | A version for the rules, or a one-time migration of existing `.skyra` trees | `replay` |

## Status

Built on branch `design-room-implimentation`, 2026-09-30. Every Go test and
every `skyra demo` scenario pass, each ending in a byte-identical replay, and a
fake-Codex conversation driven through the webhost (turns, `echo`, `read`, a
native attempt, a close) replays byte-identically.

| Ids | State | Commit | Note |
|---|---|---|---|
| A1–A9 | **built** | `d49ce18` | The fake Codex's probe still names `reply`: an unoffered tool, now interrupted and returned as `refused` |
| B1 | **in part** | `a275748` | Authority is on the Horizon; `Stamp`, `Frame` and `Resume` are still in the struct |
| B2 | **in part** | `dad1bf0` | A Horizon's id is the hash of its encoding, not yet (authority, creating invoke, input contract): that is E |
| B3, B4 | **built** | `dad1bf0` | Committed at release; the child looks it up and refuses one its commit does not derive |
| B5 | not built | | The stack is durable (each Horizon is committed with it), but the `Resume` copy still travels |
| B6, B7 | **built** | `a275748` | Exec is no longer expressible as a grant |
| B8 | **built** | `dad1bf0` | Author `operator`, a 16-byte nonce |
| B9 | not built | | The webhost's sibling rule stays until item 8 defines an accepted record |
| B10 | **built** | `dad1bf0` | Closes 20's finding 1 |
| B11 | not built | | Improvements item 8, direction 1 |
| C1, C2 | **built** | `a275748` | `"continuation": true` replaces `"reach"` in a contract |
| C3 | deferred | | Moves with C4 (see the correction above) |
| C4 | not built | | Needs the loop operator |
| D1 | **in part** | `8ba6f4c` | `system.narrow` built. `system.register` needs a tool's bytes in the tree for replay (improvements item 6); the admin facet needs faults that climb to the human |
| D2, D3 | **built** | `8ba6f4c` | A reference is the narrow invoke's stamp, recorded in the waiting caller's `Horizon.Refs` |
| D4–D6 | not built | | Need the facet declaration format, agents with Horizons of their own, and C4 |
| E1–E7 | not built | | One encoding change for all of it; its open questions (a written canonical encoding, migrate or version) come first |
| Doc edits | **done** | | 14–23, the agenda, the research question, both READMEs. Not applied to `next/README.md`, `think.md` or `act.md`, which hold uncommitted work |

## Order

1. **A** (A1–A9). No change to record encoding except the new fault kind.
2. **C1–C3** and **B1, B6, B7**: the surface shrinks and the Horizon becomes
   authority only, shared, rendered once. These land together.
3. **B2–B5, B8–B10**: the Horizon committed and looked up, the stack durable,
   the room at genesis, crash recovery and resume from the tree.
4. **B11**, then frame hashes.
5. **D1–D5**: the room's facets, the c-list, facets at registration.
6. **E**: identity, one encoding change for all of it.
7. **C4** and **D6**, with improvements item 4: Codex as an agent, the loop
   operator.

## Per document

| Document | Proposal | Kind |
|---|---|---|
| [14](../14-interpreted-not-stated.md) | [14-interpreted-not-stated-proposal](14-interpreted-not-stated-proposal.md) | superseded in part: the `reply` capability |
| [15](../15-adapters.md) | [15-adapters-proposal](15-adapters-proposal.md) | historical; a note |
| [16](../16-ingress.md) | [16-ingress-proposal](16-ingress-proposal.md) | code and doc: park by hash, resume from the tree |
| [17](../17-app-server-custody.md) | [17-app-server-custody-proposal](17-app-server-custody-proposal.md) | its blocking open item is answered |
| [18](../18-app-server-spec-improvements.md) | [18-app-server-spec-improvements-proposal](18-app-server-spec-improvements-proposal.md) | section 7 conflicts; facets replace the tool split |
| [19](../19-appserver-custody-features-not-implimented.md) | [19-…-proposal](19-appserver-custody-features-not-implimented-proposal.md) | statuses change |
| [20](../20-boundry-crossing.md) | [20-boundry-crossing-proposal](20-boundry-crossing-proposal.md) | crossings change; finding 1 closes |
| [21](../21-observation-passing.md) | [21-observation-passing-proposal](21-observation-passing-proposal.md) | three open choices answered |
| [22](../22-read-tool.md) | [22-read-tool-proposal](22-read-tool-proposal.md) | code and doc: address the slot, shared Horizon |
| [23](../23-desktop-app.md) | [23-desktop-app-proposal](23-desktop-app-proposal.md) | cut 3 gets its mechanism |
| [24](../24-the-room.md) | [24-the-room-proposal](24-the-room-proposal.md) | implementation plan |
| [25](../25-surface-refactor.md) | [25-surface-refactor-proposal](25-surface-refactor-proposal.md) | implementation plan |
| [26](../26-agents-and-facets.md) | [26-agents-and-facets-proposal](26-agents-and-facets-proposal.md) | implementation plan |
| [agenda](../agenda.md) | [agenda-proposal](agenda-proposal.md) | `interpret` and `reply` sections superseded |
| [research-question](../research-question.md) | [research-question-proposal](research-question-proposal.md) | the system description and H3 |
| [think](../think.md) | [think-proposal](think-proposal.md) | a Horizon from a narrowing |
| [act](../act.md) | [act-proposal](act-proposal.md) | narrowing per node; interrupts vs faults |
| [README](../README.md) | [README-proposal](README-proposal.md) | the index |
