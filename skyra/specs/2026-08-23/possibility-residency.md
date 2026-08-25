# The possibility is the page — residency, faulting, and materialization

**Date:** 2026-08-23
**Status: WORKING PROPOSAL.** Settled in conversation on 2026-08-23. Three joined changes to the
CTU↔kernel picture: (1) the `Slice` type dissolves into `Possibility`; (2) ACTIVATE is reframed as a
page-fault handler, with the arena as the address space; (3) kernel materialization splits into LOAD
and MINT along the EVENT/REQUEST entry axis. It builds on `2026-08-22/ctu/possibility.md` and
`2026-08-18/kernel-ctu-boundary/boundaries.md`, and reads with `2026-08-18/CPU-Research/`.

_Instantiates `2026-08-18/the-quadrant.md` — the CTU (system, unprivileged) faults up; the kernel
(system, privileged) materializes. The fault is the trap the quadrant already named ("traps up for
durable I/O")._

> **One executable object crosses every boundary: the possibility. A CTU runs a possibility that is
> RESIDENT. When a reference does not resolve, the CTU faults to the kernel, which MATERIALIZES a
> possibility — LOADING an existing engram, or MINTING a brand-new one. The CTU cannot mint
> authority, only attenuate it, so origination is the kernel's by construction.**

---

## 0 · Terms

Doc terms used as-is: Possibility, Arena, Horizon, Engram, Caveat, Move, Root, ACTIVATE, and the
EVENT / REQUEST entry distinction (`2026-08-18/CPU-Research/kernel-entry-point.md`). `MINT` is the
verb of the kernel-held `HorizonMinter` already named in the v8 interfaces. Declared here because
they are load-bearing and not yet in the vocabulary:

```text
RESIDENT / RESIDENCY   borrowed from the CPU-Research correspondence (a page is present or not). a
                       possibility is RESIDENT on a CTU when it is live in memory there and its
                       references resolve locally. not a new mechanism — a name for a state the
                       model already has.
FAULT                  the CTU's trap to the kernel on an unresolvable reference. the quadrant's
                       "trap up for durable I/O", named. it is the EVENT entry, not the REQUEST one.
MATERIALIZE            the kernel act that turns a name/request into a live possibility. its two
                       modes are LOAD and MINT. named because "ACTIVATE" now covers only the LOAD
                       mode; MATERIALIZE is the umbrella over LOAD + MINT.
```

No new primitive is introduced. Nothing here adds a "process" concept.

---

## 1 · One object crosses every boundary — dissolve the Slice

`2026-08-22/ctu/possibility.md` already states it: **"a forwarded slice = a child possibility on
another CTU, under the same shared identity."** The v8 build proves it — `Greet` receives a `Slice`
and immediately reconstructs a `*Possibility` from it, discarding the wrapper. `Slice` is a
`Possibility` minus one field. There is no second thing to slice.

So: **`Slice` is dissolved into `Possibility`. The object a MOVE carries on its request leg is a
child possibility.** The possibility is the unit of execution *and* of transport, with no second
wire type.

### What must NOT be lost with it — slicing is not the Slice

The `Slice` type was doing security work *by its shape*: it had **no `Arena` field**, which made
leaking the home address space **inexpressible**, not merely forbidden. Dissolving the type must not
demote that guarantee to "remember to strip it." Two structural rules preserve it:

```text
ARENA NEVER TRAVELS   a possibility in flight carries no home arena. the RECEIVER stamps a LOCAL
                      arena at admission. (the v8 Greet already does this.) shape the type so a
                      non-root / in-flight possibility STRUCTURALLY cannot carry the parent's arena —
                      then the leak stays inexpressible without a Slice type.

SHIP A CHILD, NEVER   what crosses is a narrow CHILD possibility, constructed with only the operands
THE PARENT            it needs and an ATTENUATED horizon (Bound). it never held the parent's scope,
                      so it cannot leak it. the child IS the slice — as a first-class possibility.
```

**The cut does not disappear; it moves from "build a Slice" to "build the child possibility."** The
partitioning decision of `2026-08-22` is unchanged and least-knowledge is still the deciding reason.
Only the wire form changed.

### The other side

Shipping a `Possibility` where a minimal `Slice` used to go is a real regression risk: `Possibility`
is the fat type (arena, full frame, frontier), and a naive send re-introduces the option-B
least-knowledge leak `2026-08-22` chose against. The mitigation is the discipline above — a *narrow
child*, never the parent, and no arena on the wire. If a future build ever ships the parent
possibility for a sub-task, this change has regressed `2026-08-22`, and that is the thing to catch in
review. The simplification is only sound while "what crosses is a narrow child" holds.

---

## 2 · Residency, and the fault

A CTU executes a possibility that is **resident** — live in its memory, references resolving locally.
This is the common case and touches no kernel; it is the bottom-left quadrant cell running
unprivileged.

When a reference does **not** resolve — a name, an engram, or an arena that is not resident — the CTU
cannot proceed and cannot invent it (it holds no durable store and cannot mint authority). It
**faults** to the kernel. This is the page fault, and the correspondence is exact:

```text
CPU                                        SKYRA
────────────────────────────────────────  ──────────────────────────────────────────────────
reference a virtual page                   execution references a possibility / name / arena
page present & mapped        → run         RESIDENT possibility            → execute, no kernel
page not present → PAGE FAULT (EVENT)      reference does not resolve → FAULT (trap up), EVENT entry
  kernel: alloc frame, load from disk,       kernel: MATERIALIZE — bring a live possibility into
  update page table, resume                  memory, resume the faulting execution
CR3 / address-space id (ASID)              ARENA id — a root possibility's address space
```

The fault is the **EVENT** entry of `kernel-entry-point.md`, not the REQUEST one: the execution did
not *ask* to enter the kernel, it *touched something absent*. It is involuntary. That is what makes
it a fault and not a call.

**The arena is the address space.** The arena back-reference a possibility carries is the ASID / page
table base of its root. A reference to an arena not currently resident is a fault on that whole
address space; the kernel maps it in.

---

## 3 · Materialization splits: LOAD vs MINT

Not everything the kernel materializes already exists. Two modes, and they divide on the same
EVENT/REQUEST axis:

```text
                        BACKING                    ENTRY              KERNEL ACT
LOAD an existing        an engram exists durably   FAULT / EVENT      ACTIVATE — load engram → live
  engram                (a file-backed page)       (involuntary)      possibility. durable → live.

MINT a brand-new        nothing backs it yet       REQUEST / syscall  create identity + MINT a fresh
  possibility           (demand-zero / a fresh     (voluntary)        ROOT horizon. new authority
                        mmap)                                         enters here.
```

- **LOAD** is `ACTIVATE` unchanged (`2026-08-18/primitive.md`): the blueprint is on "disk," the
  kernel pages it in. Involuntary — reached by a fault.
- **MINT** is origination of something with *no* engram. It is voluntary — an explicit request to
  create — so it is the REQUEST entry, a syscall, not a fault.

**Why MINT is forced through the kernel — the strong reason.** Not merely "it is a privileged write."
The `HorizonMinter` is held *only* by the kernel; the CTU can only `Bound` (attenuate, narrow-only).
A brand-new possibility needs *fresh* root authority, and **the CTU structurally cannot mint
authority** — it can only narrow authority it already holds. Origination is the kernel's *by
construction*, not by policy. This is the strongest kind of boundary: the CTU cannot cross it even if
it tried. (Top-right of the quadrant: the account/pin mints the root the kernel then honors.)

---

## 4 · The request-ingress discriminator, corrected

This project flagged a needed split on the request side — origination vs continuation — and proposed
"a request that doesn't have a known possibility ID" as the trigger. **Once one object crosses every
boundary, that trigger is wrong:** a legitimate cross-boundary *continuation* also arrives with an ID
the receiver has never seen. Unknown ID cannot distinguish them.

The discriminator that survives is **residency / resolvability**, not identity novelty:

```text
Does the request arrive carrying a RUNNABLE, AUTHORIZED possibility?
  YES — a whole child possibility, horizon chains to a known root   → CONTINUATION
        admit + execute. NO kernel. (a "present page")
  NO  — names something not resident but durable (an engram)        → FAULT   → kernel LOAD (ACTIVATE)
  NO  — asks to create something brand-new, no engram               → REQUEST → kernel MINT
```

A continuation carries its own resolved state and a verifiable authority chain — it is a present
page, admitted and run without the kernel. Origination and page-fault are the *same* shape from the
ingress view — "a name that does not resolve to a resident possibility" — and the kernel resolves it
by **loading** or **minting**. The signal is an **unresolvable reference**, never an unfamiliar ID.

### Where this leaves the receiving ports

```text
CONTINUATION   a runnable child possibility arrives and is admitted (authority + provenance checked),
               then executed. no kernel crossing. (this is the request-side admission the build had
               been doing inside Greet — now it receives a possibility, not a Slice.)
RETURNING      a response to a cut THIS CTU issued lands at its join. RECOGNIZE, unchanged.
MATERIALIZED   a fault (LOAD) or a create-request (MINT) produces a possibility via the kernel, which
               then executes like any resident one.
```

RECOGNIZE (returning responses) is untouched. The novelty is entirely on the request/origination
side, and it is now keyed on residency.

---

## 5 · What this changes in the interfaces (v8)

```text
causal/     remove `Slice`. the request leg of `Move` carries a `*Possibility` (a narrow child).
            ensure an in-flight / non-root `Possibility` cannot carry a home `Arena` on the wire.
ctu/        Greet receives and admits a child possibility (not a Slice). add the FAULT path: an
            unresolvable reference traps to the kernel rather than failing. Execute's local-miss
            becomes a fault, not a `Failure`.
kernel/     ACTIVATE is the LOAD mode of materialization (unchanged surface). add MINT: create a new
            root possibility + mint its root horizon. MINT is the only origination of fresh authority.
```

These are interface-shape notes, not a build order. The build is the v8 tree's job.

---

## 6 · Open

```text
FAULT MECHANICS        the exact hand-off: how a faulting Execute suspends, how the kernel resumes it
                       with the materialized possibility, and whether resume is re-entry to Execute or
                       a continuation. the CPU resumes the faulting instruction; the CTU analogue is
                       unspecified.

MINT AS A REQUEST      if MINT is the REQUEST (syscall) entry, what is the request's shape, and who
                       may issue it? origination from outside is the authenticated principal
                       (external ingress, top-right); origination from a running possibility is a
                       different caller. same MINT, two callers — unresolved.

RESIDENCY EVICTION     a page model implies eviction. what evicts a resident possibility, and does an
                       evicted-then-refaulted possibility keep identity? (REMEMBER tears down at the
                       root; ordinary eviction is not REMEMBER. the relationship is unwritten.)

ARENA AS ASID          how far the address-space analogy holds: is an arena a flat space, and what is
                       the equivalent of a page within it? the correspondence is suggestive, not yet
                       load-bearing.

CONTINUATION NAMING    with Slice gone, the request-side admission of a child possibility still shares
                       the name GREET with the (semantic) newcomer case. whether continuation
                       admission and origination want distinct names, or one port branching on
                       residency, is left to the CTU-lockdown pass.
```
