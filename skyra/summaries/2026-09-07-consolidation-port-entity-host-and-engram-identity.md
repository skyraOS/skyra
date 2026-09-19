# The consolidation port, the entity/host model, and Engram identity

> **Current loop synthesis:** [2026-09-08 — main loop and execution boundaries](2026-09-08-the-main-loop-and-execution-boundaries.md)
> supersedes this summary where ownership, continuation storage, and Move/Ret invocation semantics differ.

**Date:** 2026-09-07
**What this is:** a portable digest of the 09-07 session — a large documentation consolidation, the
compile/Remember decoupling, the new entity/host/address routing axis, and the settled Engram-identity
rule. Readable without the code in front of you. Current docs live in `specs/2026-09-07 /` (note: the
directory name has a trailing space — likely accidental, worth `mv`-ing). The engine sketch is
`src/skyra/skyra-ctu-engine/`.

---

## Where this sits (the bridge from prior days)

The engine is a PROPOSAL/probe, not a runtime. The shape carried in from 09-05/09-06: the **kernel
contains the scheduler**; **Memory is shared** between kernel and CTU; work is handed off as **`Fd`
entries** (`Fire([]*Fd)`); a move's operands are its children's returns collected in an ordered
**register frame**; and execution reports **events** rather than writing scheduler state. Today was
mostly consolidation and two forward axes (entity/host routing, Engram identity) — not new engine code.
The code still builds/vets/tests but cannot yet produce an end-to-end root observation.

## 1. The consolidation port

- 59 files across `2026-08-24` / `08-27` / `09-01` were reviewed; **10 ported selectively, 49 excluded**
  (kept as historical references, not deleted). A `source-review.md` ledger accounts for every file.
- Result: four clean docs — `ownership-and-drivers`, `authority-and-focus`, `move-register-return`,
  `engrams-and-resolution` — plus the entity/host thread in its own subdirectory.
- Deliberately shed as scaffolding: the CPU-analogy research corpus, the D1–D13 execution-brainstorm
  ledger, the three-table reactor, and — in the *target* model — the `Claim/Issued/Landed` event
  lifecycle. Ladder, not house.

## 2. Compile ≠ Remember (the decoupling)

Three surfaces are now separated (git-shaped: working tree / build / commit):

```text
Working context   ← edits, observations, ongoing work        (history + live work)
Current executable ← the most recent SUCCESSFUL compile        (automatic-on-success)
Committed version  ← user-initiated REMEMBER                   (a deliberate commit)
```

Compile advances "current" automatically; **Remember is a separate, user-initiated commit**. This
supersedes the older "REMEMBER compiles the trace / emits the executable at persistence" framing — the
two must not be one automatic event. An in-flight invocation keeps the executable version it started with.

## 3. The entity/host/address model (new routing axis)

- **Host = stable environment identity, not an address** (a device independent of its OS; an AWS account
  independent of its VMs). We want continuity across host replacement **without** preserving a HostID —
  continuity comes from entity bindings, engrams, and drivers.
- **Address belongs to the entity–host binding:** traversal `Entity → EntityAddress → HostAddress → Host`.
- **EntityAddress owns the representation/parser contract** — an LLM endpoint and a compiled-execution
  endpoint may parse history/live-work differently; compatibility is defined at the address, not globally.
- **Placement (DECIDED):** entity, address, and host live on the **Move**; the executable thing lives on
  the **Engram**.
- Nothing routing-algorithm or transport level is accepted yet. Structs are lean (`Host{ID}`,
  `HostAddress{Address, Host}`, `EntityAddress{Address, HostAddresses}`); `History`/`LiveWork`/
  `WorkArtifact` are placeholders.

## 4. Engram as a per-run audit log

```text
EngramID → InvocationID → ordered WorkUnits{ timestamp, workArtifact, operatingEntity, change }
```

- **InvocationID = the lifecycle of one run** of a move. When the run completes, its events are saved to
  the engram under that InvocationID.
- It is the **audit log of a live move** — making explicit the event-tracking we already do implicitly on
  the register. Each WorkUnit is a change (who/when/which opcode/resulting artifact); there is no separate
  events list.
- It is a **distinct layer from Compile and Remember** — not versioning, not commit, just "what this run
  did," attributable.

## 5. The settled identity rule (the key clarification of the day)

- **`EngramID = MoveID = fd = hash(Action + ordered Params)`.** Same action + ordered params → same ID;
  change either → a different ID.
- That action+params is the **entire hash preimage**. The audit log (Invocations/WorkUnits) — and the
  current compiled artifact — are **associated data keyed by the immutable ID, never hashed into it**.
  This is the **git-notes pattern**: an immutable content-addressed object with mutable annotations
  attached, none of which move its hash.
- Consequence: one engram accumulates **many** invocations without its ID ever moving; two runs of the
  same move are two distinct `InvocationID` entries, not a collision over one slot.
- This corrected a wrong worry raised mid-session (that a growing history would drift the content hash):
  it doesn't, because the history is not in the preimage. Pinned into `work-unit-and-invocation.md` and
  `proposal-engram.md`.

## What's still open (the reconciliation debt)

```text
1. LIFECYCLE. the target model (Pending + opcode-specific hold/release) diverges from the IMPLEMENTED
   Claim/Issued/Landed + Apply. code and target must be reconciled before the loop is finished.
2. THE EXECUTABLE HALF OF ENGRAM. "what constitutes the executable thing" is undefined; only the audit
   half of the struct exists. when it's added, the hash preimage must stay action+params ONLY.
3. ROUTING. entity–host binding structures, engram availability vs bindings, Horizon vs bindings, and
   egress ownership are all open. no transport chosen.
4. THE LOCAL LOOP. the engine still cannot produce an end-to-end root observation. install → ready →
   claim → fire → publish → compose → root is the next concrete milestone (a pure Ret/Compose tree).
```

## One-line takeaway

Consolidated the design to four clean docs plus an entity/host axis; **decoupled compile (automatic,
"current") from Remember (user-initiated commit)**; and settled that **`EngramID = MoveID =
hash(Action + ordered Params)`** is an immutable content-address with a **per-run audit log (and the
current compile) hanging off it as associated data** — a distinct layer from Compile and Remember.
