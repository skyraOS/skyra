# Register decisions — historical inventory for the next design pass

> **Subsequent accepted shape:** [register-frame.md](register-frame.md) replaces the single-value model
> with one register per fd containing a schema-only Contract and ordered child observations. The review
> below preserves the history; its latest-status claims predate that decision.

**Date:** 2026-09-06  
**Status: REVIEW / no new decisions retired.** Collects the register history from the August 18/23
designs, v11, August 27 revisions, September brainstorm, and current ownership/Body decisions.
This is an inspection of local specs and code, not an implementation or a claim that all historical
invariants are enforced. Source links preserve the original reasoning, including superseded claims.

## Reading the statuses

**Subsequent refinement:** [parent/child contracts](body-registers-and-return.md#parent-and-child-contracts-on-the-engram)
now have explicit durable fields on Engram, sharing [Contract = Action + Schema](body-registers-and-return.md#contract--action--schema).
This supersedes the intermediate shared-Return proposal; actual returned values remain Observations.
The inventory below captures the preceding history; open placement questions should
be read with that refinement, whose concrete fields and register bindings still need design.

- **Current:** supported by the latest conversation-backed specs.
- **Superseded:** a later explicit decision already replaced the claim.
- **Earlier proposal:** useful prior work, not automatically promoted to a current requirement.
- **Open/reconcile:** the current register design needs to resolve or restate it.

The latest anchors are [Body/registers/return](body-registers-and-return.md),
[Focus](focus-and-move.md), [kernel/Memory/CTU ownership](engram-loading-and-fault-return.md), and
[scheduler lifecycle](ingress-and-fire.md). The latter's kernel/Memory ownership has itself been
updated by the loading decision. File dates and old "SETTLED" banners alone do not establish precedence.

## Actual code surfaces

### v11: a contract/value cell

From [v11 causal.go](../../src/skyra/skyra-dna-v11/causal/causal.go):

```go
type Register struct {
    Contract Contract
    Value    *Observation // nil means not yet landed
}

// Context.Registers: map[RegKey]Register
// MoveTemplate: Reads []RegKey, Writes RegKey, Contract Contract, ...
// Body: Moves []MoveTemplate, Return RegKey, ...
```

The [v11 CTU implementation](../../src/skyra/skyra-dna-v11/ctu/ctu.go) supplies useful concrete evidence:
`operands` follows Reads order and checks map presence plus non-nil Value; `land` fills a contract if
absent and assigns Value; `flipDependents` scans Waiting moves in the matching context; `setResult`
reads Body.Return. These helpers do not establish the latest lifetime/contract design. In particular,
`land` itself does not reject overwrites or validate structural types. The source contains no register
deletion implementing consume/free. Existing [tests](../../src/skyra/skyra-dna-v11/ctu/ctu_test.go) cover
earlier local/dependency/remote paths, not the new Body and shared Memory proposal. No tests were rerun
for this documentation review.

### v12: landed values plus keys; no Register struct

From [v12 causal.go](../../src/skyra/skyra-dna-v12/causal/causal.go):

```go
type RegKey struct {
    Name Name
    Tag  Tag
}
type Tag struct{ Iter int }

type MoveID struct {
    Possibility PossibilityID
    Writes      RegKey
}

type Observation struct {
    Content Content
    Segment Segment
}
type Content struct{ Bytes []byte }
type Segment struct{ By Name }
type Contract struct{ Name Name }

// Context.Registers: map[RegKey]Observation
// Context.Result: *Observation
// Context.ReturnFd: *MoveID
// Body.Children: map[RegKey]BodyID
// Body.Contract: Contract
```

These are current code shapes, not the completed new design. There is no structural type schema,
discovered-contract record, reference-counted value store, or implemented new matching/consume protocol
in these definitions. [Engine](../../src/skyra/skyra-ctu-engine/engine.go) still imports v12 Context and
RegKey; its execution algorithms are proposal stubs and its ownership has not caught up with the specs.

## How we got here

| Stage | Register decision / shape | Status now |
|---|---|---|
| Aug 18 | Typed assignment names a destination register and contract; landing is binding, not REMEMBER | Naming and binding/persistence distinction remain useful; cold structural expectations were later corrected |
| Aug 23 / v11 | Register = Contract plus optional Value; explicit Reads/Writes; one writer per register identity | Historical representation; single-assignment intent still relevant, new representation open |
| Aug 27 | First call names the child; return discovers shape; REMEMBER records the contract; warm return guards it | Current, refined by Body's declared roles/partially known requirements |
| Aug 28 | Reads/Writes collapse into child edges: children supply inputs, parent slot receives output | Later exposed missing request/literal/argument bindings; blanket ban on explicit requirements no longer fits Body |
| Aug 31–Sep 1 | Delete Register/RegState; keep only landed values; scheduling inferred from Work/InFlight/Registers membership | Value-map baseline remains in v12; scheduling/table-membership rule superseded |
| Sep 2 brainstorm | Named binding, trace retention, new tags for correction, per-edge slots/shared values, every input produced by a move | Earlier proposals with open details; do not treat all ledger entries as implemented or final |
| Sep 5–6 | One scheduler lifecycle map; Memory owns shared resident structures; kernel coordinates/provisions; CTU executes | Current |
| Latest Body discussion | Body is register requirements plus return specification; live values remain in Memory; Op fulfills return | Current direction; concrete register and contract fields remain open |

Sources: [typed assignment / continuation](../2026-08-18/continuation-language.md),
[live possibility](../2026-08-23/live-possibility.md),
[move graph](../2026-08-23/possibility-graph.md),
[contracts](../2026-08-27/contracts-and-remember.md),
[move/possibility collapse](../2026-08-27/move-and-possibility.md),
[register lifecycle](../2026-08-27/register-lifecycle-and-transition.md),
[brainstorm ledger](execution-brainstorm/decision.md).

## Decisions to have in front of us

| Topic | What we previously said | Assessment for this pass |
|---|---|---|
| Role versus type | A slot is named for its role. Two inputs can share a type, e.g. original/revised Documents or minuend/subtrahend numbers | **Current.** Match by execution/slot identity; observation type alone cannot choose a slot |
| Register identity | RegKey = Name + Tag; MoveID scopes its output slot with PossibilityID; child's return echoes parent's fd | **Code baseline.** Exact frame scope and root output convention need reconciliation |
| Single assignment | One producer writes a given register version once; semantic rework creates a new version instead of overwriting | **Earlier invariant worth retaining.** Duplicate identical returns/conflicting writes and enforcement remain open |
| Missing values | v11 used nil Value; v12 uses absence from Registers, minting a value entry only at LAND | **Current code baseline.** Declared slot requirements may exist before values. Lazy value allocation does not mean the requirement is absent |
| Register-state enum | RegState/Register were removed when state became table membership | **Partly superseded rationale.** Scheduler has its own explicit states now. This does not automatically require or prohibit a future register specification struct |
| Readiness | All registers read by an ordinary consumer must be present | **Current baseline, refine.** Body asks what is needed to return; presence, type acceptance, semantic fulfillment, and conditional obligations need definitions |
| Wakeup | Aug 23 stored backward dependencies and rebuilt a forward waiters index; v11 instead scans Waiting moves | **Reopened.** Later seam B and latest ownership leave index ownership/query-vs-event open; an old index sketch is not a current commitment |
| Ordering | Dependencies establish execution order; source/text order does not. D4 proposes nominal binding, marshaled into positional Driver.Perform inputs at the edge | **Named roles fit current design.** Driver argument order and canonical digest order still need separate rules |
| Input origins | Child returns alone lost leaf inputs/request data. Alternatives: literals on Body, const-producing moves, runtime args; D10 leans toward every slot having a writer move | **Open.** Boundary pre-seeding versus authority-producing move was explicitly left unresolved; new Body framing does not solve input injection automatically |
| Contracts | Call names; cold return discovers; REMEMBER records; warm return guards | **Current.** Slot-level versus Body-level durable contract fields, type registry, composition, and mismatch behavior need definition |
| Value lifetime | A landed value waits for its consumer; it is not freed merely because its producer completed | **Current obligation.** Exact consume/free point remains open, especially across faults and asynchronous execution |
| Fan-out | D12 proposes one fd/slot per consumer edge, referencing one shared content-addressed value; free each slot independently and reclaim value after last reference | **Earlier leaning proposal.** Not represented by v12's inline Observation map; confirm before adopting value references/refcounts |
| Trace lifetime | D7 retains who-produced/consumed-what until REMEMBER, independently of value release | **Earlier proposal/current persistence obligation.** Trace representation and exact payload retention remain open |
| Fault retry | Missing backing preserves operands and same MoveID; Faulted is not a value returned to the parent | **Current.** No separate Attempt ID adopted; recovery completion cannot consume the inputs |
| Semantic failure | Later specs describe errno as an Observation; earlier Aug 18 mismatch withheld binding entirely | **Evolution to reconcile.** A returned error being present does not automatically satisfy a success-observation requirement |
| Residency | Context-not-resident and operand-not-yet-produced were separate waits; historical latch Absent→Materializing→Resident deduplicates loading | **Distinction current.** New exact-EngramID loading and live-frame identity require separate keys; concrete latch/waiter structure open |
| Ownership | Registers moved from travelling Possibility to local Context, then CTU-owned state, now shared Memory | **Current:** Memory owns storage; CTU writes through it; kernel provisions; scheduler owns lifecycle |
| Root return | v11 Body.Return explicitly selects output; v12 removed it as derivable from DAG sink, but no precise root fd/result convention followed | **Open.** New Body return specification must make this concrete |

Detailed source anchors: [D4 nominal binding](execution-brainstorm/decision.md#d4--bind-operands-nominally-no-internal-calling-convention),
[operand options](operand-model.md), [args proposal](execution-brainstorm/args-decision.md),
[D7 trace](execution-brainstorm/decision.md#d7--opcodes-are-transitions-two-pops-the-trace-remember-compiles),
[D8 correction](execution-brainstorm/decision.md#d8--effectful-leaves-compute-not-fire-staged-at-a-kernel-confirm-gate-fault),
[D10 input sources](execution-brainstorm/decision.md#d10--the-source-collapse-every-fd-is-written-by-a-move-fault-is-a-move-to-kernel-opcode),
[D12 fan-out](execution-brainstorm/decision.md#d12--per-edge-fd-1-fd--1-register-context-fan-out-via-a-content-addressed-refcounted-value),
[identity/races](../2026-08-27/context-identity-and-races.md),
[gap inspection](../../src/skyra/skyra-dna-v12/gaps.md),
[current lifetime/transition obligations](closed-loop-outstanding.md).

## Tags: three meanings we must not silently combine

[Execution-cycle §5](../2026-08-23/execution-cycle.md) proposed loop unfolding with
`Tag{Ctx uint64, Iter uint32}`: a new iteration gets new register identities; nested loop entry/exit maps
contexts. v11/v12 only retain `Tag{Iter int}`. Nested-loop semantics were not implemented by that type.

[Register-lifecycle seam H](../2026-08-27/register-lifecycle-and-transition.md) and D8 later proposed
bumping Tag.Iter for deoptimization or human correction. This is a new semantic value/version, not
automatically the same thing as loop iteration.

The latest Materialize trace reissues the **same MoveID** after loading missing backing, with preserved
inputs and no separate Attempt ID. Since MoveID contains RegKey, changing Tag also changes MoveID in the
current code. We must decide when semantic rework creates a new version and how it routes to consumers;
do not apply the old "bump tag on reissue" wording to every storage fault.

## Claims already superseded, versus candidates to retire

Already superseded by later decisions (no new retirement made by this review):

- Scheduling state must be represented by separate Work/InFlight/Registers tables. Scheduler now owns
  one MoveID→State map; values and lifecycle records have distinct lifetimes.
- Completion means deleting scheduling records; quiescence follows merely from empty work tables.
  Completed records now remain until a separate retirement policy, and outstanding recovery matters.
- CTU/opcodes directly mutate scheduling maps; CTU separately invokes an involuntary fault handler.
  CTU returns events; the kernel scheduler interprets them and routes recovery.
- The kernel cannot install backing in execution memory. It now provisions through shared Memory.
- Children as both possible choices and mandatory data requirements. Focus and Dependencies separate them.
- Caller must already know an unseen child's structural return contract. Discovery precedes REMEMBER.
- The old travelling register frame as the default remote mechanism. Later Move/Context decisions keep
  frames local and route values; wire details remain incomplete.

Candidates to revise explicitly in the upcoming register design, not decisions taken here:

1. **"No Register struct" as a blanket ban.** Deleting an old Contract/Value/state wrapper does not forbid
   a durable register requirement now that Body explicitly records those requirements.
2. **"All reads are derivable from child edges."** Keep one source of binding truth, but make request,
   literal, bound-input, and return roles expressible rather than assuming they need no representation.
3. **"Free all inputs when the parent fires."** Replace with a precise lifetime rule covering async work,
   failed checks/handoffs, recovery, and fan-out before deleting any live value.
4. **"Any reissue increments Tag.Iter."** Distinguish semantic versions and loop occurrences from a
   Materialize retry of the same invocation.
5. **"LAND is CAS against the durable engram."** The Aug 28 proposal mixes content-addressed persistence
   with live register publication. Current Memory owns live publication; choose its actual synchronization
   contract separately. A content digest is not itself a concurrency protocol.
6. **"PossibilityID = hash(body + inputs) is the universal live identity."** Earlier reuse proposals and
   v12 comments say this; latest discussion separates reusable EngramID from execution identity. Shared
   code/inputs do not by themselves settle whether two intentional executions share register state.
   Also reconcile earlier per-Body PossibilityID versus latest whole-chain wording before choosing frame keys.

## Recommended next walkthrough

Use one return and two same-type roles:

```text
Return: Comparison
Registers: original : Document, revised : Document
```

Walk declaration → one value arrives → second arrives → fulfill return → REMEMBER → warm run.
At each step record only what we actually need: slot identity/role, nominal requirement, discovered
contract, producer binding, and live value. Then repeat with missing backing before execution and with
an observation that fails a warm guard. This should expose which facts belong durably on Body versus
in live Memory, without prematurely adding another binding table or status enum.

The first concrete decision is what a **declared register** contains while its value and full structural
contract may both be unknown. The next is what exactly makes it **satisfied**. This inventory supplies
the history for those decisions; it does not preselect their structs.
