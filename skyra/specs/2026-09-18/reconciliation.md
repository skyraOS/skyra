# Reconciliation — the 09-18 collapse vs. the 09-16/17 CU/CTU docs

Date: 2026-09-18.

Status: note. Records which claims in the 2026-09-16 and 2026-09-17 proposals are
superseded by the 2026-09-18 "collapse" (one smart component, one writer, one
queue). It changes no runtime; it aligns the reading order of the specs so a later
implementation follows the newer contract and does not re-derive deleted machinery.

## The two drafts

There are two overlapping drafts of the CU/CTU inner loop in `../2026-09-16`:

- **Draft A — two smart components (2026-09-16 → 09-17).**
  [`ownership-boundries of ctu`](../2026-09-16/ownership-boundries%20of%20ctu/README.md),
  [`interpretation`](../2026-09-16/interpretation/README.md),
  [`supervisor`](../2026-09-16/supervisor/README.md),
  [`possibility`](../2026-09-16/possibility/README.md).
  Interpreter **and** Supervisor are both smart; Moves bounce between **two Move
  mailboxes** via a custody transfer; stale snapshots are reconciled with a global
  **revision** (OCC).

- **Draft B — the collapse (2026-09-18).**
  [`cu-cycle`](../2026-09-16/cu-cycle.md),
  [`ctu/minimal-loop-v2`](../2026-09-16/ctu/minimal-loop-v2.md),
  [`MessageBroker`](../2026-09-16/MessageBroker/README.md),
  [`tu`](../2026-09-16/tu/README.md).
  **One smart component** (Interpreter) + **one dumb writer** (StateExecutor /
  StateOwner) + a worker pool; **one inbound queue**; fd-style identifiers minted at
  commit; **no revision on the interpret path**.

**Draft B is the selected direction.** Draft A is retained for the ownership carve-up
it still gets right (below), but the loop mechanics in it are superseded. The Draft A
files have not yet been rewritten, so read them through this note.

## What survives unchanged (still authoritative in Draft A)

These are not affected by the collapse and remain the current contract:

- **StateOwner is the sole canonical writer**, behind restricted typed profiles;
  components get no general read/write and no map aliases.
  ([MessageBroker](../2026-09-16/MessageBroker/README.md))
- **Possibility / Horizon** as pure identity-and-authority data; **no `Dispatch`/
  `Resolve`** on Horizon; authority-check + Surface binding + find-or-register of an
  execution happen in **one StateOwner transaction**.
  ([possibility](../2026-09-16/possibility/README.md))
- **Trap is a Surface**, not a control-flow path; corrective work re-enters the
  normal queue; Trap cannot mutate canonical state or bypass Horizon.
- **Contracts are `(Surface identity, input contract) -> expected return contract`**;
  capabilities are recursive Surfaces with their own globally addressable identities;
  discovery is not authority.
- **Return validation is execution-specific** — matched against the recorded
  invocation, never trusted from a payload claim.
- **Acceptance != output != completion**; a child Invoke can be a response without
  completing its source.
- **Commit is the boundary** between the rollbackable side and the irreversible
  external effect; effects are released only after commit.
- **Operations are not Moves** — operations are a closed vocabulary over state; Moves
  are protocol transitions between Surfaces.

## What is superseded

| # | Draft A claim (where) | Superseded by Draft B | Consequence |
|---|---|---|---|
| 1 | **Two smart components.** Interpreter *and* Supervisor make decisions. ([ownership §Central pattern](../2026-09-16/ownership-boundries%20of%20ctu/README.md); [supervisor](../2026-09-16/supervisor/README.md)) | One smart component. Supervision's *decisions* fold into interpretation; only its *worker machinery* remains, triggered by commit. ([minimal-loop-v2 §What v2 changes](../2026-09-16/ctu/minimal-loop-v2.md)) | Supervisor is no longer an interpreting service. Dispatch is a post-commit fan-out, not a decision stage. |
| 2 | **Two Move mailboxes** (Interpretation + Supervisor) with a custody bounce between them. ([ownership §Mailboxes](../2026-09-16/ownership-boundries%20of%20ctu/README.md)) | **One inbound queue.** "Which stage owns this Move" is no longer a writer-held fact; the writer does not know pipeline topology. ([minimal-loop-v2 §What v2 changes](../2026-09-16/ctu/minimal-loop-v2.md)) | The Supervisor mailbox and the second custody list are deleted. |
| 3 | **Custody transfer between components** as an outer handoff protocol (`Transfer(moveID, from, to)`). ([ownership §Component custody](../2026-09-16/ownership-boundries%20of%20ctu/README.md); [interpretation](../2026-09-16/interpretation/README.md)) | Custody-as-handoff retired with the second mailbox. `tu` already restated `Transfer` as a stateless `Transition(request, from, to)` profile; under one queue there is no cross-stage custody bounce to protect. ([tu](../2026-09-16/tu/README.md)) | `CustodyState.Transfer` and the "one outer handoff loop" section no longer describe the runtime. |
| 4 | **`from`/`to` routing** carried on the handoff. ([ownership](../2026-09-16/ownership-boundries%20of%20ctu/README.md); [tu](../2026-09-16/tu/README.md)) | Routing is expressed by the **entry point / opcode jump table**; `from`/`to` is retired. ([minimal-loop-v2 §What v2 changes](../2026-09-16/ctu/minimal-loop-v2.md)) | Route is `decode(opcode) -> handler`, not a bound source/destination pair. |
| 5 | **Global revision / OCC on the interpret path**; detached snapshots reconciled at commit; stale proposals re-read and re-evaluated. ([ownership §Revision-based reconciliation](../2026-09-16/ownership-boundries%20of%20ctu/README.md)) | **No revision on the interpret path.** Read and commit are one uninterrupted turn of the single service; nothing mutates state in between, so no snapshot goes stale. ([minimal-loop-v2 §One consumer, one turn](../2026-09-16/ctu/minimal-loop-v2.md)) | Revision/OCC becomes a *later* tuning knob, needed only if interpretation is made concurrent. It is not in the core. |
| 6 | **`claimed` / pending-acceptance / per-move processing status** and the stale-revision **re-interpret** loop. ([interpretation §interpretAndTransfer](../2026-09-16/interpretation/README.md); [ownership](../2026-09-16/ownership-boundries%20of%20ctu/README.md)) | Deleted. One consumer means `Take` is the only claim; a resumed waiting move is authorized exactly once by the single-threaded turn that satisfies its last dependency. ([minimal-loop-v2 §Traps, faults, and waiting](../2026-09-16/ctu/minimal-loop-v2.md)) | No claim map, no per-move status, no re-interpret retry. |
| 7 | **Resume via re-enqueue** — a satisfied dependency puts the runnable Move back through the mailbox. (implied by [ownership §Sequential interpretation](../2026-09-16/ownership-boundries%20of%20ctu/README.md)) | **Resume happens inside interpret.** Interpreting the Move that satisfies the last dependency resolves the now-runnable Move in the same turn; its identifier is minted at that commit and it dispatches with the rest. ([minimal-loop-v2 §Traps, faults, and waiting](../2026-09-16/ctu/minimal-loop-v2.md)) | No second trip through decode; no second waiting map. |
| 8 | **Commit returns success/ack**; execution registered as a later Supervisor step. ([ownership §Dispatch binds](../2026-09-16/ownership-boundries%20of%20ctu/README.md)) | **Commit mints an fd-style identifier** per authorized execution; the identifier *is* the execution (no separate registration step); scope/liveness live in the writer's table. ([cu-cycle §The identifier is a file descriptor](../2026-09-16/cu-cycle.md)) | Correlation, confinement, and settlement are structural via the identifier, not a registry lookup. |
| 9 | **Single transaction across Interpreter + Supervisor** (earlier idea; already being walked back in `tu`). | **Per-stage atomic transactions**, no transaction spanning components: "work reaches the next stage only after the current stage's operations commit." ([tu](../2026-09-16/tu/README.md)) | The cross-component transaction is explicitly rejected. |

## Reading order after this note

1. [`cu-cycle`](../2026-09-16/cu-cycle.md) — the shape (read -> commit -> execute; identifier as fd).
2. [`ctu/minimal-loop-v2`](../2026-09-16/ctu/minimal-loop-v2.md) — the loop (supersedes `minimal-loop.md` v1).
3. [`MessageBroker`](../2026-09-16/MessageBroker/README.md) and [`tu`](../2026-09-16/tu/README.md) — StateOwner profiles and the stateless transition.
4. [`possibility`](../2026-09-16/possibility/README.md) — identity/authority record (unchanged by the collapse).
5. [`ownership-boundries of ctu`](../2026-09-16/ownership-boundries%20of%20ctu/README.md), [`interpretation`](../2026-09-16/interpretation/README.md), [`supervisor`](../2026-09-16/supervisor/README.md) — read for the surviving ownership carve-up only; treat their loop mechanics (two mailboxes, custody bounce, revision/OCC, re-interpret) as superseded per the table above.

## Still open after the collapse

These were open before and remain open; the collapse does not resolve them:

- The closed **operation vocabulary**, split into system-plane and execution-plane sets, and which operations may co-commit.
- **Identifier encoding** (unforgeable opaque reference vs. writer-table key) and the scope granularity `stateVia(id)` keys to it.
- **Horizon-revocation policy for a live identifier mid-execution** — honor or reject in-flight requests after the grant is pulled, ideally as one liveness check covering revocation and settlement together.
- The **waiting-record contract**: how interpret identifies the moves a satisfied dependency unblocks and authorizes them in the same commit.
- The **single-operation rule** for execution-plane read-modify-write, so no worker holds a snapshot across turns (the condition that keeps "no revision" honest).
- **Naming**: `state.Commit` vs. the earlier `cu.Execute`; reserve "execute" for the world phase.
- **Queue capacity, cancellation, and shutdown** coordination between the interpreter loop and the worker pool.
- **Durability across whole-process loss** — in-process records are not a crash-recovery contract.

See [`../2026-09-16/cu-cycle.md`](../2026-09-16/cu-cycle.md) and
[`../2026-09-16/ctu/minimal-loop-v2.md`](../2026-09-16/ctu/minimal-loop-v2.md).
