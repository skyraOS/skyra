# Proposals: every change the new design asks for

Date: 2026-09-30  
Status: **partly implemented** on branch `design-room-implimentation`
(2026-09-30). Collects what [24](../24-the-room.md),
[25](../25-surface-refactor.md), [26](../26-agents-and-facets.md),
[30](../30-types-by-origin-routes.md) (section F, added 2026-10-01) and
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

| Id | Change | Where | Status |
|---|---|---|---|
| A1 | `interpret` derives return from the slot: `to` equal to the emitter's frame id is a return; a name in reach is an invoke; anything else is a horizon fault | `core/machine/machine.go:285–322` | **Built**. `d49ce18` |
| A2 | Unaddressed output is a `leaf` fault carrying the output, not a string return | `machine.go:292`, `core/record/record.go` (new kind `Leaf`) | **Built**. `d49ce18` |
| A3 | Drop the `"reply"` reservation | `runtime/registry/registry.go:56` | **Built**. `d49ce18` |
| A4 | Tell every target its slot: `SKYRA_FRAME=<frame id>` in the environment (continuations already see `frame` on stdin) | `machine.go:569` (`runOf`) | **Built**. `d49ce18` |
| A5 | Every emitter of `"reply"` addresses its slot instead | `targets/read/main.go:65`, `codex/io/main.go:50`, `codex/resident/main.go:391`, `codex/adapter/main.go:234` (legacy), `runtime/instance/instance.go:380`, `hosts/web/main.go:314` (uses the parked instance's frame), `hosts/web/calls.go:220` | **Built**. `d49ce18`. Also `echo`, `agent`, `user`, which returned unaddressed text |
| A6 | Faults carry their observation: the bytes that did not fit, the address outside the Horizon, what a program printed; the continuation's `StepView` exposes it | `record.FaultInfo`, `machine.go:627` (`StepView`), `machine.go:635` | **Built**. `d49ce18`. The observation is the fault record's `payload`, shown as the step's `value` |
| A7 | Codex's native attempt is returned as an observation (a typed `refusal` in the resident's output contract), not `native.<type>` addressed to trigger a fault | `codex/resident/main.go:367`, `codex/resident/contract.json`, `codex/adapter/main.go:292`, `hosts/web/calls.go`, `hosts/web/calls_test.go` | **Built**. `d49ce18`. The field is `refused` |
| A8 | The fake Codex's `"tool": "reply"` bypass probe becomes a probe of naming `reply` (now a horizon fault) and of guessing a slot | `codex/fake/main.go:133` | **Built**. `d49ce18`. The probe still names `reply`: now an unoffered tool, interrupted and returned as `refused` |
| A9 | Tests: `TestInterpret`, replay, `read_test`, `resume_test`, `lineage_test` | `*_test.go` | **Built**. `d49ce18`, plus a resident test for the refusal |

### B. The Horizon: durable, derived, shared (24)

| Id | Change | Where | Status |
|---|---|---|---|
| B1 | The `Horizon` struct holds authority only (reach, grants). `Stamp`, `Frame` and `Resume` move out | `core/horizon/horizon.go:126` | **In part**. `a275748`. Authority is on the Horizon; `Stamp`, `Frame` and `Resume` are still in the struct |
| B2 | Horizon id = hash(authority, the invoke that created it, its input contract) | `horizon`, `record` | **In part**. `dad1bf0`. The id is the hash of the Horizon's encoding; the (authority, invoke, contract) form waits for E |
| B3 | A Horizon is committed durably at creation, and that commit is the release | `runtime/tree` (`horizons/`), `machine.handoff` (`:413`) | **Built**. `dad1bf0`. `.skyra/horizons/` |
| B4 | The child is handed the Horizon's hash; its first act is a lookup, and it re-derives what it found | `machine.accept` (`:330`), `instance` `SendHorizon` (`instance.go:177`), `handshake` | **Built**. `dad1bf0`. `accept` refuses a Horizon its commit does not derive |
| B5 | The stack lives with the Horizon in durable state; steps read it and write write-once entries. The `Resume` copy goes | `machine.gather`/`claim` (`:509`, `:532`), `instance.gather` (`instance.go:242`), `Depth()` | Not built. The Horizon is committed with its stack, but the `Resume` copy still travels and `gather`/`claim` still read it |
| B6 | Every call shares its caller's Horizon; a Horizon is created only by a narrowing, a registration, or a call into another agent | `machine.Narrow` (`:456`) | **Built**. `a275748`. Narrowing and references in `8ba6f4c`; registration and calls into another agent are not built |
| B7 | The Seatbelt profile is rendered per Horizon, plus the exclusions: the floor, and `process-exec` of the running program only. Exec can never be registered | `horizon.Render` (`:157`), `resident.go:21` | **Built**. `a275748`. Exec is no longer expressible as a grant; OS test for exec denial |
| B8 | The room: the derivation's root. Genesis is its creation, authored by the operator and carrying a nonce | `runtime/boot/boot.go:47` | **Built**. `dad1bf0`. Author `operator`, a 16-byte nonce |
| B9 | Crash recovery by reading state: a record no derivation accepted is an orphan; an accepted step with no successor runs again; effects are at least once | `boot`, `webhost` (retires the `533d98d` rule) | Not built. The webhost's sibling rule stays until item 8 defines an accepted record |
| B10 | Resume re-derives the instance from the tree; the host holds a hash, not authority (closes 20's finding 1) | `skyra/main.go:107`, `runtime/park` | **Built**. `dad1bf0`. Refuses work that is not parked, a wrong frame count, and a second resume. Closes 20's finding 1 |
| B11 | Replay checks that every mark is backed by a record (improvements 8, direction 1); after that, frame fds can become hashes | `runtime/replay` | Not built. Improvements item 8, direction 1 |

### C. The surface: identity plus contracts (25)

| Id | Change | Where | Status |
|---|---|---|---|
| C1 | Delete `Surface.Reach`, `Authority` and `Delegate`; they are the Horizon's | `horizon.go:85`, `registry.go:73,116,148,151`, `machine.Narrow`, the `"reach"` line in four `contract.json` files | **Built**. `a275748`. Contracts say `"continuation": true`; every registry grant joins the room's one authority |
| C2 | Delete `Leaf()`; the stdin format (continuation or bare payload) becomes an explicit property of the surface. **Superseded by [27](../27-a-horizon-is-a-process.md):** one envelope for every worker, so the `continuation` flag goes | `horizon.go:111`, `machine.go:359` | **Built**. `a275748`. `Surface.Continuation`, declared in the contract |
| C3 | Delete `Surface.Args`. **Corrected:** not unused; the launcher passes the real app-server's `--disable` flags through it, so it moves with C4 | `horizon.go`, `registry.go:70,79,148`, `instance.go:361`, `appserver.go:178` | Deferred. Moves with C4 |
| C4 | Later: move the resident and network fields off the surface; network grants to the Horizon, keep-alive to control flow (a loop operator) | `horizon.go:92–100`, `instance.startResident`, `runtime/custody` | Not built. Needs the loop operator |

### D. Agents and facets (26)

| Id | Change | Where | Status |
|---|---|---|---|
| D1 | The room's public facets `system.narrow` and `system.register`, implemented by the runtime and unregisterable; an admin facet in the human's reach | `machine`, `registry` | **In part**. `8ba6f4c`. `system.narrow` built. `system.register` needs a tool's bytes in the tree (improvements item 6); the admin facet needs faults that climb to the human |
| D2 | Narrowing is two steps: `system.narrow` returns a reference, then the caller invokes it | `machine.interpret`, `Narrow` | **Built**. `8ba6f4c`. Demo scenario: the agent narrows echo, then calls it |
| D3 | Reach becomes a c-list: references carried by a return are recorded by the room and added to the receiver's reach | `machine`, `record` | **In part**. `8ba6f4c`. Only for `system.narrow`: the reference is a name, `system.ref.<n>` (F1; it was the invoke's stamp), recorded in the waiting caller's `Horizon.Refs`. Returns do not yet carry references in general |
| D4 | Registration presents facets; the registrant gets them back as references | `registry` | Not built. Needs the facet declaration format |
| D5 | A call to another agent's facet runs under that agent's Horizon | `machine.Narrow` | Not built. Needs agents with Horizons of their own (registration) |
| D6 | Codex as the first agent with facets (creator: start/exchange/inspect/stop; public: the turn), replacing `ResourceOwnerDepth` | `codex/resident`, `registry-codex.json` | Not built. Needs D4 and C4 |

### E. Identity (improvements 9)

| Id | Change | Where | Status |
|---|---|---|---|
| E1 | One written canonical encoding for records and declarations | `core/record`, `core/contract` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |
| E2 | Domain-separated hashes: type, content, record, frame, Horizon | `record.Stamp` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |
| E3 | Two hashes per step: content id and stamp | `record`, `replay` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |
| E4 | Ids as the path of handoffs, segmented by frame; the storage layout follows | `runtime/tree` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |
| E5 | Types and principals as records; the runtime derives a record's type | `registry`, `record` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |
| E6 | The recipient on return and fault records (improvements decision 5, 2b) | `record`, `machine.faulted` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |
| E7 | A version for the rules, or a one-time migration of existing `.skyra` trees | `replay` | Not built. The encoding's open questions come first: a written canonical encoding, migrate or version |

### F. Types by origin, and routes (30)

The implementation plan, with the decisions to confirm first, is
[30-types-by-origin-routes-proposal](30-types-by-origin-routes-proposal.md).

| Id | Change | Where | Status |
|---|---|---|---|
| F1 | Audit where hashes reach a model or an external actor. The invariant: hashes go to workers, never to models | `machine.runOf`, `Continuation`, `StepView`, `targets/agent`, `codex/adapter`, `codex/resident`, `hosts/web` | Built (`12cf4f6`). One leak fixed: references are `system.ref.<n>`. Two exceptions accepted, recorded in 30 |
| F2 | `Shape`, `Content`, `Diff` and `Present` over decoded JSON: two hashes per node, a pinned injective encoding | `core/contract/shape.go` | Built (`4f7db10`) |
| F3 | The type of a return, `H(actor identity, contract, shape)`, derived on demand, never stored in a record | `core/contract`, `core/machine` | Built (`f496bbd`) |
| F4 | The Starlark route runner, its own module, under an empty Horizon; no `while`, no recursion, a step cap; the `skyra` package | `targets/route` (new module), `core/machine/room.go` | Built (`3b1932a`). `if` only (decision 7); the runner imports `core/contract` |
| F5 | Routes declared at registration, per binding, for deterministic parents | `runtime/registry`, `horizon.Surface.Routes` | Built (`82515e1`) |
| F6 | Routing at discharge: interrupt, send or fault, at gather; a send is an invoke under the parent's Horizon | `core/machine` (`gather`, a `RunRoute` effect) | Built (`5256b2c`). "Wake" is named interrupt (kinds `novelty`, `route`, `depth`); no `RunRoute` effect: `LookupRoute`, then `Run` |
| F7 | The novelty interrupt (`novel` and a presentation in `StepView`) and the room's `system.route` facet | `core/machine`, `room.go`, `targets/agent` | Built (`5b1d0f3`) |
| F8 | The log as the test suite: a rewritten route runs against its recorded cases; flips are declared | `core/machine`, `runtime/tree` | Built (`5b1d0f3`), inside `system.route` |
| F9 | The declared `"output"` contract becomes optional, then goes | `contract.json` files, `registry`, `interpret`, `horizon.Surface` | Built (`00db79a`, `2d974a0`, `8cce174`): `Surface.Output` is gone |
| F10 | **A route never runs with more authority than it was written with.** Store the author's authority in `RouteEntry` at `system.route`; at lookup, run a route only if the current Horizon's authority and reach are within the author's, else the return is novel. Today routes are keyed by `horizon.Identity(surface)`, so a route written under a narrowed Horizon, or under a prompt injection, fires silently in every later frame of that surface, including broader Horizons | `core/machine/write.go` (`RouteEntry`, `recorded`), `route.go` (`lookedUp`), `runtime/instance` (`lookupRoute`), `runtime/tree` | Not built. Review finding, 2026-10-01 |
| F11 | Replay re-runs each route: the runner over the stored input, and the decision compared. Also check the run's `previous` and `targets` against the frame. Today replay re-derives the decision from the runner's stored output only, and 30 says replay runs it again | `runtime/replay` (`routed`) | Not built. Review finding, 2026-10-01 |
| F12 | `system.route`'s `accept` names flipped cases by their runner input's hash, not by index: indexes come from a directory listing and shift when another frame records a run between the refusal and the retry | `core/machine/write.go` (`FlipView`, `recorded`), `routeIn` | Not built. Review finding, 2026-10-01 |
| F13 | Replay checks `route-table/` against the `system.route` records that wrote it, so a route lookup (novel or a hit) is verifiable. With E | `runtime/replay`, `runtime/tree` | Not built. Review finding, 2026-10-01 |
| F14 | Say which invariant holds: frame ids reach the Codex app-server (an external actor) in JSON-RPC ids, an accepted exception. Either the rule is "never to a model", and 30 says so, or the exception is fixed with 29's actor state | `docs/next/30-types-by-origin-routes.md`, `codex/adapter`, `codex/resident` | Not built. Review finding, 2026-10-01 |

### G. Every return is routed (30)

The rule, agreed 2026-10-02 ([30, "Every return is routed"](../30-types-by-origin-routes.md#every-return-is-routed)):
when the runtime derives a return, it computes the return's type and runs the
receiver's route for that type; the route's decision is the next transition;
with no route, the default route interrupts, of kind novelty. **F10 comes
first:** without it, the rule routes every return through F10's hole.

| Id | Change | Where | Status |
|---|---|---|---|
| G1 | One lookup: the receiver's route for the type, else the default route (interrupt, kind `novelty`). Declared routes (F5) are recorded into the same table at registration, so "recorded before declared" goes | `route.go` (`lookedUp`), `runtime/instance` (`lookupRoute`), `runtime/registry` (`LoadWithRoutes`), `runtime/tree` | Not built. After F10 |
| G2 | Delete `routable`'s exceptions: a facet's return and a return at the operator are routed, and reach the default. The operator's default hands the return to the human, as today. `MaxChain` stays the one gate outside routes | `route.go` (`routable`), `core/machine` (`gather`) | Not built |
| G3 | A fault is routed: `TypeOf` types a fault discharge from what the route is given (kind, cause, observation), and a route branches on its kind. Faults that climb to the human become a route that sends upward | `route.go` (`TypeOf`, `Decided`), `targets/route` (the return a fault presents), `replay` | Not built. Open: whether the fault's kind is part of its type |
| G4 | Fan-out: each child's return is routed on its own; the round ends when every route has decided, and the parent is woken once if any interrupted, with every interrupt | `core/machine` (`gather`, `resolve`) | Not built. Waits for fan-out rounds (one output per step today) |

## Status

**Implemented:** A (all), B3, B4, B6, B7, B8, B10, C1, C2, D2, F1–F9.
**In part:** B1, B2, D1, D3.
**Not built:** B5, B9, B11, C4, D4, D5, D6, E1–E7, F10–F14 (review findings; F10 first), G1–G4. **Deferred:** C3.

Every Go test and every `skyra demo` scenario pass, each ending in a
byte-identical replay. A fake-Codex conversation driven through the webhost
(turns, `echo`, `read`, a native attempt, a close) replays byte-identically.
Real Codex was not run. The Status column above says what each change became;
the per-document proposals say the same for their documents.

| Commit | What |
|---|---|
| `d49ce18` | A: returns address the slot; faults carry their observation |
| `a275748` | B1 (in part), B6, B7, C1, C2: one authority per Horizon; the surface is identity and contracts |
| `dad1bf0` | B2 (in part), B3, B4, B8, B10: Horizons committed and looked up; resume from the tree; authored genesis |
| `8ba6f4c` | D1 (narrow), D2, D3 (in part): the room's `system.narrow`; references join the caller's reach |
| `a24fdfb` | the document edits below |

## Order

1. ✅ **A** (A1–A9). No change to record encoding except the new fault kind.
2. ✅ **C1, C2**, **B6, B7**; ◐ **B1**; C3 deferred: the surface shrinks and the Horizon becomes
   authority only, shared, rendered once. These land together.
3. ✅ **B3, B4, B8, B10**; ◐ **B2**; ✗ **B5, B9**: the Horizon committed and looked up, the stack durable,
   the room at genesis, crash recovery and resume from the tree.
4. ✗ **B11**, then frame hashes.
5. ✅ **D2**; ◐ **D1** (narrow only), **D3** (narrow's reference only); ✗ **D4, D5**: the room's facets, the c-list, facets at registration.
6. ✗ **E**: identity, one encoding change for all of it.
7. ✗ **C4** and **D6**, with improvements item 4: Codex as an agent, the loop
   operator.
8. ✓ **F** (2026-10-01, branch `self-extension`): F1 first; then F2–F3 (types), F4–F5 (routes), F6 (routing), F7
   (the interrupt), F8, F9. Self-extension (28, `system.register`) follows.
9. ✗ **G** (agreed 2026-10-02): F10 first; then G1 (one lookup, the default
   route), G2 (no exceptions), G3 (faults routed); G4 with fan-out.

## Per document

| Document | Proposal | Kind | Implemented |
|---|---|---|---|
| [17](../17-app-server-custody.md) | [17-app-server-custody-proposal](17-app-server-custody-proposal.md) | its blocking open item is answered | **Done**: revision note; code: B7, C1 built. The per-thread narrowing at `start` is not wired |
| [18](../18-app-server-spec-improvements.md) | [18-app-server-spec-improvements-proposal](18-app-server-spec-improvements-proposal.md) | section 7 conflicts; facets replace the tool split | **Done**: section 7 reconciled; code: D6 not built |
| [23](../23-desktop-app.md) | [23-desktop-app-proposal](23-desktop-app-proposal.md) | cut 3 gets its mechanism | **Done**: cut 3's mechanism named; code: the admin facet is not built |
| [24](../24-the-room.md) | [24-the-room-proposal](24-the-room-proposal.md) | implementation plan | **In part**: status line added; see B |
| [25](../25-surface-refactor.md) | [25-surface-refactor-proposal](25-surface-refactor-proposal.md) | implementation plan | **In part**: item 1 built; item 2 corrected and deferred; item 3 not built |
| [26](../26-agents-and-facets.md) | [26-agents-and-facets-proposal](26-agents-and-facets-proposal.md) | implementation plan | **In part**: `system.narrow` and its references built; see D |
| [30](../30-types-by-origin-routes.md) | [30-types-by-origin-routes-proposal](30-types-by-origin-routes-proposal.md) | new; supersedes 28's checks and handlers | **Built**: see F |
| [think](../think.md) | [think-proposal](think-proposal.md) | a Horizon from a narrowing | Not applied: `think.md` is uncommitted work in the tree |
| [act](../act.md) | [act-proposal](act-proposal.md) | narrowing per node; interrupts vs faults | Not applied: `act.md` is uncommitted work in the tree |
| [README](../README.md) | [README-proposal](README-proposal.md) | the index | Applied in the working tree, uncommitted |

Removed on 2026-09-30, with their per-document proposals, because they were
retired or fully done: 14 (interpreted, not stated), 15 (adapters), 16
(ingress), 19 (the app-server gap inventory), 22 (the read tool) and the
agenda. The proposals for 20, 21 and the research question were removed too:
their edits are applied, and those documents stay. All are in git history.
