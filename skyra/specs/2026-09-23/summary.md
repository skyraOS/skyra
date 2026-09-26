# Summary: the stateless controller, where it stands

Date: 2026-09-25  
Status: a map of this directory. It separates what is built, what is agreed on
paper, and what is still being explored. Every section links to the document
that holds the detail.

## In one paragraph

The long-running DNA-19 host process is being replaced by a **stateless
controller**: one `skyra` binary that runs once per step. Each run takes custody
of one Move, checks it, commits a content-addressed record, runs the target the
record names inside an OS sandbox rendered from its authority (the Horizon), and
hands the target's output to a fresh successor that checks it in turn. All
durable state lives in `.skyra/`. A runnable demo, [`minimal-shape/`](minimal-shape/),
boots to echo and proves the core protocol. The current design work is on what
happens *around* a single call: dependencies between calls, faults and
interrupts, and a durable memory of contracts that lets the system learn.

## Document map

| Document | Holds | Status |
|---|---|---|
| [`stateless-controller/proposal.md`](stateless-controller/proposal.md) | the original spec: decisions D1–D9 | agreed; **lags** the later work below |
| [`stateless-controller/thoughts/01–13`](stateless-controller/thoughts/) | the reasoning behind each decision, in order | agreed per note |
| [`minimal-shape/`](minimal-shape/) | runnable demo: boot to echo | **built**, tested |
| [`improvements.md`](improvements.md) | separate the custody envelope from the reductions | proposed |
| [`io-buckets.md`](io-buckets.md) | every I/O operation bucketed; primitives; handlers by state; the join rule | proposed |
| [`faults.md`](faults.md) | faults (runtime-proven) and interrupts (memory misses) | agreed in outline |
| [`durable-storage.md`](durable-storage.md) | memory of contracts and procedures; snapshots | agreed in outline |
| [`../../summaries/2026-09-23-the-stateless-controller-design.md`](../../summaries/2026-09-23-the-stateless-controller-design.md) | the handoff written mid-session | superseded by this file |

---

# Part I: the current architecture (built)

## The step

One run of `skyra` is an **instance**: trusted code (`skyra`) plus the untrusted
target it runs. Custody passes between the trusted halves of consecutive
instances. The target never takes part.

```text
parent                         instance                               successor
  bytes ─────────────────────▶ INTAKE   parse, check against Horizon, commit R, stamp
        ◀──────────── stamp ──
  verify R, narrow ─ Horizon ▶ RELEASE  accept; render sandbox
                               EXECUTE  run target; frame + hash its output
                               HANDOFF  spawn ──────────────────────▶ INTAKE …
                                        ◀──────────────────── stamp ──
                                        verify, narrow ── Horizon ──▶ RELEASE …
                                        exit
```

**Decisions that hold (D1–D9, [proposal](stateless-controller/proposal.md)):**

| | Decision |
|---|---|
| D1 | Interpret, then spawn. Emissions are unchecked; the receiver checks them. The emitter is never the checker |
| D2 | Custody: every Move is held by exactly one live supervisor or one durable commit. A parent exits only after its successor stamps |
| D3 | The stamp is `hash(record)`, a pointer the parent verifies by reading the tree, not a claim. No secrets |
| D4 | Phases: intake, release, execute, handoff |
| D5 | One commit per handoff. The release writes nothing |
| D6 | The release is the child's Horizon: required to execute; only the parent can issue it |
| D7 | The Seatbelt sandbox is rendered from the Horizon. `.skyra/` is always denied |
| D8 | Leaves (now largely subsumed: a leaf is a Horizon with empty reach) |
| D9 | States of a Move: **resolve → invoke → return / fault**, named by "the system is in a state of ___" |

## The Horizon

Everything a step knows besides the bytes:

```go
type Horizon struct {
	Stamp   string             // this frame's commit; the predecessor of whatever is emitted under it
	Surface *Surface           // what this frame runs (path + content hash)
	Reach   map[string]Surface // what emissions under it may invoke: the registry, as this frame sees it
	Settles string             // set when the frame settles a parent frame
}
```

Reach only narrows (`narrow`). It is computed, never stored, so replay re-derives
it. The frame's directory travels beside it as an fd over `SCM_RIGHTS`, and that
fd is the actual capability.

## The code: front end, controller, back end

```text
step   (front end)   pure: (state, event) → (state, effects). Decides everything
drive  (controller)  performs effects, turns outcomes into the next event. Decides nothing
tree, handshake, …   (back end) the I/O services
```

- **Event:** a fact, reported to the step, about something that already
  happened. Processed by `step`.
- **Effect:** a request, produced by the step, for something to happen. Processed
  by `drive`.
- `step` switches on the Move's **state** first (the top-level decision), then on
  the phase.

## The demo (`minimal-shape/`)

- One package per owner, imports running one way. `machine` is pure, enforced by
  `TestMachineIsPure`.
- Three scenarios: `echo hello`; echo probing `.skyra/` from inside its sandbox
  (read and write denied, with a control run showing both allowed without it); a
  leaf trying to invoke (committed and stamped as a fault).
- **Replay:** every scenario re-derives its lineage from genesis with the pure
  step and requires byte-identical records.
- 15 tests. Stamps are deterministic, and have stayed byte-identical through
  every refactor except the one that deliberately changed the record format.

---

# Part II: ideas being explored (on paper)

## 1. Custody versus reductions ([improvements.md](improvements.md))

A step does two kinds of work that the code does not yet separate:

- **Custody:** the fixed protocol around every step (commit, stamp, release,
  verify). 7 of 9 effects and 3 of 4 events belong to it.
- **Reductions:** what the Move actually does: **Resolve**, **Apply**,
  **Return**, **Fault** (plus **Admit**, the boundary that turns bytes into a
  Move).

Proposal: custody becomes a fixed envelope around one reduction:
`custody in → reduce → perform → custody out`.

## 2. I/O buckets, primitives and handlers ([io-buckets.md](io-buckets.md))

Every one of the 20 I/O operations is sorted into a bucket:
**bookkeeping** (`tree`), **transport: control** (the socket), **transport:
data** (stdin/stdout), **process**, **execute**, **observation**. Each bucket
exposes primitives. `drive` bundles primitives into **handlers**. `step` names
the handler; `drive` calls it. **A handler never decides**: it runs a fixed
sequence up to the next wait and returns an event.

| State | Handler | Returns |
|---|---|---|
| invoke, return, fault | `acceptCustody` | `Released` |
| invoke | `apply` | `Emitted` |
| invoke | `startSuccessor` | `Stamped` |
| invoke | `release` | — |
| return, fault | `settle` | — (to become `Settled`) |
| invoke | `fault` | — |

What lines up: `acceptCustody` is universal; invoke owns most of the rest;
return and fault are the same program.

Open direction: the base primitives reduce to **read** and **write** on a
location, plus **claim** (first write wins) and **publish** (atomic rename),
plus process lifecycle. Existing primitives become wrappers.

## 3. Dependencies and the join rule ([thoughts/13](stateless-controller/thoughts/13-dependencies.md))

- **A dependency** is something an invocation needs before it can be returned.
  An **Invoke opens** them (one slot per child); a **Return or Fault
  discharges** exactly one; **Resolve** is when the last is discharged.
  Dependencies are the edges of the tree; records are the nodes.
- **Join rule:** all children settle before the parent resumes. Wait for all, not
  fail fast. How the parent resumes is decided by its continuation.
- **A settle is a fact.** It produces a `Settled` event; `step` reads the parent's
  open dependencies and, if none remain, a fan-in claim lets exactly one sibling
  resolve the parent. This replaces boot's polling in the demo.
- **The join rule is enforced by hashing.** A parent's completion record must
  include every child's discharge hash, so it cannot be written early. The
  lineage has hashes in two directions: top-down provenance (child names parent)
  and bottom-up completion (parent names children). One completion hash names,
  verifies and can memoize a whole finished subtree.

## 4. Faults and interrupts ([faults.md](faults.md))

| | Faults | Interrupts |
|---|---|---|
| Means | something the runtime **proves** broke | something **durable memory does not have** |
| Terminal? | yes | no: resumable |
| Routed to | the parent (discharges its dependency) | an LLM (produces what is missing) |
| Kinds | **custody** (protocol broke), **malformed move** (bytes), **horizon** (tried to reach outside), **authority** (cannot fulfil its contract under its Horizon) | **novelty** (a return shape with no contract), **procedure** (a reachable Potential with no input contract) |

- **Failure is not a fault.** A target that could not do its work returns; its
  return is checked against the contract like any other, and an unfamiliar shape
  is a novelty interrupt.
- Faults are named by cause. From the parent, every fault looks the same: a
  dependency discharged with a fault. The continuation decides; the default is
  to propagate, so escalation is just propagation up the stack.
- An interrupt puts the Potential back in **resolve** with one more open
  dependency (the LLM's answer). No fifth state; it survives crashes.

## 5. Durable storage: memory ([durable-storage.md](durable-storage.md))

Three places of state:

| Place | Holds | Scope |
|---|---|---|
| Event log | what happened (the bytes each step received) | one lineage |
| Stack | open dependencies | one lineage |
| **Memory** | what the system knows: contracts and procedures | **all lineages** |

- **Schemas are JSON types.** Fit is structural and strict: unknown fields do not
  fit.
- **Input contracts are the roots.** Return contracts and procedures have
  `parent: <input contract hash>`.
- **Snapshots:** memory grows by writing a new manifest (never editing one); its
  hash is the root. **Every record names the root it read**, so replay reproduces
  verdicts even after memory has grown. Advancing `head` is a claim.
- Memory grows at registration (seeded contracts), on accepted novelty
  interrupts, and on resolved procedure interrupts. Stored next to the lineages.

## 6. Stateful actors ([thoughts/12](stateless-controller/thoughts/12-stateless-adapters-stateful-actors.md))

The runtime sees only stateless executables. Stateful actors (the Codex
app-server, the browser socket) live outside it, kept in step by adapters. The
latest proposal: the adapter repeats the instance pattern one layer down
(read, commit, execute against the actor), and tail calls with holes carry
mid-turn dependencies back to the runtime. Not yet reconciled with sections 3–5.

---

# Part III: where it goes

## Prior art

Merkle traces of dependency execution are well established: *Build Systems à la
Carte* (verifying and constructive traces), Bazel and Buck2 (Merkle input trees
and an action cache), Nix (hashed derivations), Unison (hash-named code),
Temporal (deterministic replay). What looks new here is the combination:
completion hashes as **traces** (not only cache keys), **nondeterministic** LLM
steps recorded before use, a cache miss that **escalates to an LLM and is
learned** rather than recomputed, and **custody and OS-enforced authority** woven
into every step.

## Most important open questions

1. **What an input contract is, concretely.** Both interrupts key on it; it is
   the most load-bearing undefined term.
2. **Where slots live and what a continuation looks like on disk**
   (`frames/<owner>/deps/<index>`, and a Potential in resolve whose holes name
   them).
3. **Nested Seatbelt.** Codex sandboxes its own tools; whether that works, or
   fails open, inside our profile has not been tested (stage 0 of the proposal).
4. **How a child asks for authority** (for example a reserved `escalate`
   surface), precisely enough for intake to confirm the gap.
5. **Reconciling the adapter layer** with dependencies, faults and memory.

## Next steps

1. **Define the input contract** concretely, with its JSON-types schema.
2. **Build dependencies in the demo:** `settle → Settled → resolve`, boot as the
   first dependent, the polling loop removed. The smallest step that puts the new
   ideas into running code.
3. **Separate custody from reductions** in the demo, keeping stamps identical.
4. **Consolidate:** fold the agreed pieces back into one current spec, so
   `proposal.md` stops lagging.
5. **Run the nested-Seatbelt spike** against the real Codex app-server.
