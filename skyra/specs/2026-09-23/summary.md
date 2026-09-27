# Summary: the stateless controller, where it stands

Date: 2026-09-25, updated 2026-09-26  
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

**Latest (2026-09-25):** two shifts. First, **invoke, return, fault and
interrupt are now derived, not stated**: an emitter supplies only operands, and
the runtime derives the state with one pure function. A return is invoking a
single-use **reply capability**. Second, the project is focusing: for
September to December the target is **the Horizon and the authority model**,
with two node types (continuation = LLM, leaf = tool), framed as a research
question ([`research-question.md`](research-question.md)).

## Document map

| Document | Holds | Status |
|---|---|---|
| [`stateless-controller/proposal.md`](stateless-controller/proposal.md) | the original spec: decisions D1–D9 | agreed; **lags** the later work below |
| [`stateless-controller/thoughts/01–14`](stateless-controller/thoughts/) | the reasoning behind each decision, in order | agreed per note |
| [`minimal-shape/`](minimal-shape/) | runnable demo: boot to echo | **built**, tested |
| [`improvements.md`](improvements.md) | separate the custody envelope from the reductions | proposed |
| [`io-buckets.md`](io-buckets.md) | every I/O operation bucketed; primitives; handlers by state; the join rule | proposed |
| [`faults.md`](faults.md) | faults (runtime-proven) and interrupts (memory misses) | agreed in outline |
| [`durable-storage.md`](durable-storage.md) | memory of contracts and procedures; snapshots | agreed in outline |
| [`research-question.md`](research-question.md) | the semester research question, metrics, hypotheses, plan | draft for discussion |
| [`tomorrows-agenda.md`](tomorrows-agenda.md) | semantics first: the transition checklist, `interpret`, the reply capability | agenda for `transitions.md` |
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
| D9 | States: **resolve → invoke → return / fault**, named by "the system is in a state of ___". Since thought 14 these are **derived by the runtime**, not stated by the emitter (Part II §7) |

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

## 7. Derived, not stated ([thoughts/14](stateless-controller/thoughts/14-interpreted-not-stated.md))

The LLM does not declare whether its output is an invoke or a return. **The
runtime derives it**, from state (memory, the stack, the Horizon) and operands.
This is D1 taken all the way: the emitter does not even name what its output is.

- **The address decides.** Output sent to a capability in reach is an invoke.
  Output sent to the parent is a return. The runtime already knows the parent
  from the custody chain, so the LLM does not need to.
- **Return is invoking the reply capability.** The parent hands each child a
  single-use `reply`, a handle to the slot it waits on. This is
  continuation-passing style, and E's promise resolvers.
- **Unaddressed output is a reply.** An unspecified target falls to the least
  authority available, and `reply` can only discharge one slot, once. So a leaf's
  output is always a reply: tools just print their data.
- **States are derived, not stated.** The emitter supplies `(capability?,
  payload)`. `interpret(state, output)` derives invoke, return, fault or
  interrupt, and the record stores the result. Nothing an untrusted party emits
  can put the system into a state.

```text
interpret(state, output):
  no output                          → fault (custody)
  not parseable                      → fault (malformed move)
  names a capability not in reach    → fault (horizon)
  names a capability in reach, or none (→ reply):
      payload fits its input contract   → invoke, or return if it is reply
      fits nothing known / ambiguous    → interrupt (novelty)
```

What falls out: **one primitive** (`invoke(capability, payload)`); the model
**cannot forge control flow**; **exactly-once returns** (reply is single use);
**free tail calls** (pass your reply on); and **one contract check**, because the
reply capability's input contract *is* the return contract. Targets never emit a
Move of any kind.

**Names are notation.** A state is a predicate over system state; its name is a
label, like a bound variable. New states cost nothing on the wire. A state is real
only if it has a derivation rule, invariants, and enforcement.

## 8. Narrowing versus enforcement

Two questions that must not be folded into one "security":

| | Narrowing | Enforcement |
|---|---|---|
| Governs | **what appears in the frame**: what the model can see | **what is invokable at all**, from this agent, at this invocation |
| Benefit | **focus**: fewer choices, fewer tokens | **a hard stop**, independent of the model |
| Security value | incidental: a model can guess or be told a name | the guarantee |

- **Invariant: frame ⊆ authority.** Everything the model sees is invokable; not
  everything invokable is shown. An attempt outside authority is a horizon fault,
  visible or not.
- Agent libOS separates the same two axes ("model-visible affordances may evolve
  while resource authority changes only through explicit, audited runtime
  primitives").

## 9. The research question ([research-question.md](research-question.md))

> **Can an agent runtime whose authority is object-capability based, and
> enforced by the OS, match a conventional agent runtime on cost and capability,
> while providing strictly stronger containment?**

- **Two node types:** a **continuation** (an LLM, with reach) and a **leaf** (a
  tool, whose reach is only `reply`).
- **"Efficient"** is three measurements: overhead, capability (task success), and
  friction (authority faults per task). **"Better security"** is containment,
  blast radius and auditability.
- **Evaluation as a 2×2:** visibility (full or narrowed) × enforcement
  (session-level sandbox or per-invocation ocap). Narrowing is measured on
  capability and focus; enforcement is measured on containment with full
  visibility, so it stands on its own.
- **Baseline:** the same Codex model with its native Seatbelt sandbox. The
  comparison is per-invocation ocap narrowing rendered into the OS sandbox, versus
  one coarse sandbox for the session.
- **Related work** is active, especially in 2026: CaMeL (77% of AgentDojo tasks
  with guarantees, against 84% undefended), Progent (monotonic narrowing), Agent
  libOS (closest in spirit), ActPlane (1.9–8.4% overhead at the kernel), PORTICO
  (revocable, epoch-bound capabilities), Bounded Agents (narrowing delegation),
  and the AgentDojo benchmark. What looks distinctive here: **per-invocation OS
  enforcement**, **cross-layer faithfulness** (every semantic invariant mapped to
  a mechanism or listed as a gap), a **replayable hash-linked trace**, and
  **memoized procedures** that could make repeated runs cheaper.
- **Plan:** transition semantics and the nested-Seatbelt spike (late September);
  a real continuation with leaf tools under rendered sandboxes (October); audit and
  replay (November); the evaluation and a writeup (December).

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

1. **What an input contract is, concretely.** Both interrupts key on it, and it
   also defines the return contract (the reply capability's input contract). It is
   the most load-bearing undefined term.
2. **`interpret` as a total function:** the addressing format for a
   continuation's output, and exactly what "state" contains.
3. **Where slots live and what a continuation looks like on disk**
   (`frames/<owner>/deps/<index>`, and a Potential in resolve whose holes name
   them).
4. **Nested Seatbelt.** Codex sandboxes its own tools; whether that works, or
   fails open, inside our profile has not been tested (stage 0 of the proposal).
5. **How a child asks for authority** (for example a reserved `escalate`
   surface), precisely enough for intake to confirm the gap.
6. **Reconciling the adapter layer** with dependencies, faults, memory and the
   reply capability.
7. **Evaluation design:** the task set, reusing AgentDojo or building a suite,
   the strength of the baseline, and how escalation works during evaluation.

## Next steps

1. **Semantics first** ([`tomorrows-agenda.md`](tomorrows-agenda.md)): define the
   input contract; write `interpret` as a total function; define the reply
   capability's lifecycle; fill in the transition checklist in `transitions.md`.
2. **Run the nested-Seatbelt spike** against the real Codex app-server.
3. **Update the demo to derived states:** echo prints plain output, and intake
   interprets it as a reply.
4. **Build dependencies in the demo:** `settle → Settled → resolve`, boot as the
   first dependent, the polling loop removed.
5. **Settle the evaluation design** with the collaborator, and consolidate the
   agreed pieces into one current spec so `proposal.md` stops lagging.
