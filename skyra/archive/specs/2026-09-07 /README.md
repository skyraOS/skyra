# Current design notes — reviewed carry-over

> **Start with the current loop:** [2026-09-08 architecture synthesis](../../summaries/2026-09-08-the-main-loop-and-execution-boundaries.md).
> Drivers execute; Scheduler owns registers and continuation state; Continuation Controller is stateless; kernel is the root
> driver. Every Move, including Ret, is a new invocation. Older ownership descriptions below remain to
> be reconciled with that model.

**Date:** 2026-09-07  
**Status:** Selective port reviewed against the current execution decisions. No source document was
copied unchanged. These notes carry surviving concepts, not historical implementation instructions.

Latest boundary: [Working context, compilation, and Remember](host-entity-address-horizon/execution-history-and-live-work.md).
Working context separates history from live work; live work is compiled. Latest successful compile is
current for new invocations, while user-initiated Remember advances the committed version separately.
This revises older descriptions tying compilation directly to persistence.

Read in this order:

1. [Ownership and drivers](ownership-and-drivers.md) — human policy, kernel coordination, shared Memory, CTU execution, driver mechanisms.
2. [Authority and Focus](authority-and-focus.md) — permission, visible choices, and committed dependencies.
3. [Move, register, and return](move-register-return.md) — current live shapes, portable Move, ordered observations, and return-path release.
4. [Engrams and resolution](engrams-and-resolution.md) — remembered behavior, filesystem-driver Resolve, and durable storage.

## The runtime loop (later 2026-09-07 session)

The execution model reframed as a router, not an executor — decided this session, **not yet in `engine.go`**.
These supersede CTU-as-executor, kernel-contains-scheduler, and the Issued/Landed event vocabulary.

- [The main loop](main-loop.md) — Scheduler-owned continuation state, stateless Continuation Controller,
  authorization, dispatch outcomes, and the kernel as root driver.
- [Drivers and the boundary](drivers-and-the-boundary.md) — a driver is an addressed bidirectional transducer;
  the receiver owns the format; all driver egress produces Move(s); `Local`+`Egress` collapse into one address→driver table.
- [The inference substrate](inference-substrate.md) — the LLM as a driver; Codex/subscription rationale; the
  tool encoding (name = opaque handle, description = Action, schema = Params) and `callId ↔ MoveID`.

The ongoing [egress/entity/host thread](host-entity-address-horizon/README.md) stays in its own directory.
It defines Host as stable environment identity and places addresses on entity–host bindings. No old
host-as-address or host-as-privilege-block taxonomy is imported here.

## Settled in the entity/host thread (2026-09-07)

- **Move vs Engram placement (DECIDED):** entity, address, and host live on the Move; the executable
  thing lives on the Engram. What "the executable thing" is remains to be defined.
- **Address owns representation (DECIDED):** traversal is Entity → EntityAddress → HostAddress → Host;
  the EntityAddress owns the parser/representation contract for its endpoint (an LLM endpoint and a
  compiled-execution endpoint may represent history and live work differently). No routing algorithm or
  transport is accepted yet.
- **Host is identity, not address:** a host identifies a stable hosting environment; continuity across
  host replacement comes from entity bindings, engrams, and drivers — not from preserving a HostID.
- **Engram carries a per-run audit log:** `EngramID → InvocationID → ordered WorkUnits{timestamp,
  workArtifact, operatingEntity, change}`. InvocationID is one run's lifecycle; the log is the durable
  record of what that run did. It is a distinct layer from Compile and Remember.
- **Settled identity:** `EngramID = MoveID = fd = hash(Action + ordered Params)` — that action+params is
  the *entire* hash preimage. The audit log, and the current compiled artifact, are **associated data
  keyed by that immutable ID, never hashed into it** (the git-notes pattern): the same action + ordered
  params always yields the same ID while its logs and current compile change beside it. Distinct from
  **Compile** (latest successful compile = current executable) and **Remember** (user-initiated commit).

## Review and precedence

The [source review ledger](source-review.md) accounts for every file under the three requested dated
spec directories: 59 reviewed, 10 used selectively, 49 excluded. Each ported document identifies its
sources and the parts left behind. Originals remain historical references; links to them are provenance,
not an instruction to inherit all their decisions.

These notes apply the latest [checkpoint](../../src/skyra/skyra-ctu-engine/checkpoint.md) and subsequent entity/host decisions.
They supersede conflicting prose in their linked historical sources. Later explicit decisions continue
to take precedence. The broader [surface proposal](../../src/skyra/skyra-ctu-engine/surface-proposal.md) contains older unapproved
suggestions; those additions have not been promoted by this port.

## What was deliberately left out

No old three-table reactor, Claim/Issued/Landed lifecycle, CTU-side result publication, separate backing
identity, Seed/Message wire format, no-forward-input rule, name-only contract, or parent contract-version
manager was carried over. Kernel-does-not-coordinate and kernel-owned decoder descriptions are replaced
by the accepted ownership and filesystem-driver Resolve boundary.

Research analogies, old boot recipes, speculative compiler/interpreter structures, and implementation-
language selections remain at their source. They are not requirements for the next local-loop handoff.
Full dependency/contract checking remains a stub boundary; concrete routing, identity encoding/revisions,
root delivery, and error/lifetime protocols remain open.

## Implementation status

The engine contains shared-Memory boot wiring, fd-entry shapes, ordered landing, and limited legacy
scheduler component code. It does not execute an end-to-end local loop. Pending, Move-return processing,
unified MoveID/EngramID, and filesystem-specific Resolve still require code-surface reconciliation.
A passing component test does not establish those behaviors. This port changes documentation only.
