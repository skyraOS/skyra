# The execution model and the authority boundary

**Date:** 2026-08-26
**Scope:** whole project state after the 2026-08-24 → 26 design arc. Supersedes the picture in
`2026-08-19-the-ctu-pivot-and-the-eight-primitives.md` on the execution side; that summary's
foundation (the CTU, the eight primitives, the kernel-is-authority correction) still stands and this
builds on it.
**Status:** current picture. v9 ran; v10 (`acquire`) was built and reviewed; v11 (the execution model)
is specified with a fresh interface proposal on the bench.

> **One boundary shows up at every level of the system, and it is the same boundary the *development
> process* runs on: a low-authority PROPOSER hands work to an EXECUTOR across a HUMAN-CONFIRMED
> authority boundary, and the proposer never holds execute or mint power.**
>
> THINK proposes → the human/kernel authorizes → the CTU executes. Everything below is that sentence,
> instantiated.

---

## 1 · What Skyra is

A universe simulator that uses LLMs as one of its base cases. It does not make intelligence; it gives
intelligence somewhere to live, someone to learn from, and a history that compounds. The load-bearing
commitment: **the human is a world-shaping participant, not an input source.** The product wins when
the user can feel the universe change after every conversation — memory, relationship, trust, the map.

That commitment is now *mechanical*, not aspirational: the human is the **principal** — the root of
authority — and the whole authority model is built to keep them there.

---

## 2 · The one picture — what changed since 2026-08-19

```text
2026-08-19 (the CTU pivot)              2026-08-26 (now)
the kernel is authority, not machine  → HELD. and made concrete: kernel owns identity + the full table.
a possibility is a reorder buffer      → SPLIT: a BODY (compiled) + a CONTEXT (live). Materialize loads one into the other.
THINK is a primitive that proposes     → THINK is the COLD path: it COMPILES a request into a Body (via the kernel).
execution is DRIVE                      → ACT: DRIVE is the HOT path (the CTU runs a Context).
the horizon bounds authority            → the horizon IS a provisioned VIEW of the kernel's table (an address space).
authority was abstract                  → authority is a CAPABILITY (a macaroon) the USER carries.
one machine                             → kernel (OS-independent authority) + CTU (OS-abstraction) as separate services.
```

The through-line of the whole arc: **the OS/CPU analogy stopped being a metaphor and became the
architecture.** Sockets, file descriptors, page tables, the MMU, the loader, dynamic linking, the
trap — each turned out to be the right shape for a real Skyra part.

---

## 3 · The foundation that survived

From the CTU pivot, still true and now load-bearing:

```text
CTU        the Causal Transition Unit — the execution core. nearly-stateless "pipeline hardware."
KERNEL     the authority — not the machine. it decides WHAT; it never does the HOW.
MOVE       the one transport verb. a cut carries a narrow child possibility (Request) or returns an
           observation (Response).
POSSIBILITY  a graph of moves over a register file — a reorder buffer, edge-triggered by DRIVE.
           park a placed cut, RECOGNIZE lands it, re-drive. termination is quiescence.
TRUST AXIS  host = BELIEVED (system), entity = CHECKED (user). (privilege is orthogonal — see §5.)
REMEMBER   gather the signed cone by content address; contracts crystallize here.
```

**v9** proved this runs: a DAG fans out across two CTUs, parks two cuts, RECOGNIZE re-drives the join,
REMEMBER gathers the cone, dispatch shows host-BELIEVED / entity-CHECKED / kernel-TRAP, and a SWITCH
move unfolds a two-iteration loop that terminates by quiescence. Throwaway, but it proved the shapes
fit.

---

## 4 · v10 — `acquire`: getting capabilities securely

THINK's first settled child, **built by Codex and reviewed**. `acquire` turns a novel endpoint into a
registered, contract-checked userspace capability:

```text
discover   FIND        search (intent → names) + resolve (name → endpoints). read-only.
test       CHARACTERIZE reach · handshake · challenge · describe → a ConnectionContract. safe probes only.
admit      REGISTER     select the fit slice + request the write. the MINT is the kernel's.
```

It ran end to end and answered its question: a novel endpoint faults, the privilege boundary decides,
the kernel mints, the acquired capability is dispatched as a **checked entity**, and reacquisition
takes the fast path. The build reshaped four of my proposals — the most important being that a
`bool` horizon couldn't express scoped authority.

### The decisions it settled (v10)

```text
D1  getaddrinfo is the pillar of discovery — name → a bounded set of candidate (triple + address) pairs.
D2  ask_permission is a horizon-gated capability (a direct trap), not recursion to root — and SCARCE.
D3  privilege (PRINCIPAL, decides) and kernel (MECHANISM, mints) stay SEPARATE — a collapsed boundary
    reads as no boundary, and it is real.
D4  a horizon IS its issued grants: Horizon = []Grant, each scoped { endpoint, purpose, action }.
D5  privilege is the TRAP AXIS, not a third trust block. blocks are exactly host + entity.
```

### The ownership model (v10 → v11)

```text
HUMAN   owns POLICY (risk tolerance). the principal; the escalation target. never executes.
KERNEL  EXECUTES the human's policy; owns the separation handlers, the MINT, and the full table.
CTU     owns EXECUTION; reads its provisioned VIEW to dispatch; RAISES traps. never mints.
```

The Skyra inversion, stated plainly: an OS bakes policy into the kernel; **Skyra hands policy to the
human and the kernel merely executes it.** That is the North Star made mechanical.

---

## 5 · v11 — the execution model (the current frontier)

The v11 line answers *"how is a possibility created, and where do THINK and execute live in the CTU?"*
It is specified in `specs/2026-08-24/skyra-dna-v11/` with a fresh interface proposal in
`src/skyra/skyra-dna-v11/`. The pieces:

### Body vs Context (the CPU split)

```text
BODY     the compiled move graph. STATIC, content-addressed, shareable, read-only. THINK's output; the
         "compiled artifact." (an ELF's .text.)
CONTEXT  the live run — register file, in-flight fds, retirement watermark, carried horizon. MUTABLE,
         per-run. the "process." two runs of one Body = two Contexts.
MATERIALIZE  the LOADER: Body + Horizon → a fresh Context.
```

### THINK / ACT — the cold and hot paths

```text
COLD (new request)     move(entity=kernel, …) → THINK COMPILES a Body → kernel PROVISIONS + MINTS a
                       capability → runs it → REMEMBERs. returns the capability the user now HOLDS.
HOT (remembered)       move(entity=engram, +capability) → CTU VERIFIES the capability OFFLINE →
                       Materialize from the cached Body → ACT (Drive). NO kernel round-trip.
```

REMEMBER is the bridge (cold → hot); it is the dynamic linker patching the GOT. And the hot path is
kernel-free because **the user carries the horizon** (a macaroon) and the CTU verifies it offline —
OAuth + macaroons at the request level.

### The Move barrier and the fd/socket split

```text
CTU owns MOVE          the single entry. ingress calls it; the CTU calls egress to send. Header dispatches
                       Request (push a child) / Response (land a return). subsumes v9's Greet/Recognize/Arrive.
KERNEL BARRIER         ingress/egress = the kernel barrier (custody: sign on egress, verify on ingress).
KERNEL owns the SOCKET TABLE   in-flight crossings (routing, custody, timeouts). client-side, flat.
POSSIBILITY holds FDs  cut-IDs = file descriptors (cut-ID → register). the frontier is the open-fd set.
```

Inbound: a **Response** pops an fd; a **Request** has no fd — it establishes a *return route* and, if
unsolicited, passes an identity gate (principal → self-auth; foreign → policy; stubbed).

### The capability table and the horizon (the MMU)

```text
kernel owns IDENTITY + the FULL TABLE   (authorization = identity × capability; a self-owned table = self-authorization.)
horizon = the provisioned SUBSET        a view of the table, carried with the work. []Grant given a home.
MMU model                               horizon = address space; grant = page-table entry; an unprovisioned
                                        reach = a PAGE FAULT → the kernel provisions (map) or denies (segfault).
view, not copy                          revocation reflected (TLB-shootdown-like). only the kernel widens; Bound narrows.
```

### Authorization vs registration (two axes)

```text
REGISTRATION   a capability EXISTS + is discoverable (the resource). patterns: ocap minting, service discovery, dynamic linking.
AUTHORIZATION  a principal may USE it (provisioned into a horizon). patterns: ocap, macaroons, OAuth2.
anchors:  ocap = the model · macaroons = the grant format (caveats = Bound) · OAuth2 = the flow · service discovery = registration.
```

### Bootstrap (single host)

```text
main       a thin out-of-band bootloader: seed+start the kernel, start the CTU, let them handshake, EXIT.
           mechanical, NOT an authority — legitimacy flows human → seed → kernel.
kernel     FIRST. loaded (not provisioned), seeded by the human. OS-independent authority (seL4-shaped).
CTU        SECOND. provisioned by the kernel. the OS-abstraction via the DRIVER seam (hosted first, baremetal-capable).
           carries inert drivers; NO ambient authority — never fires one without a horizon (CHERI-shaped).
handshake  = the kernel running ACQUIRE on the CTU (discover/test/admit → drivers into the sysblock).
           so acquire is the UNIVERSAL onboarding primitive: THINK acquires endpoints; the kernel acquires CTUs.
one kernel per host (distributed kernel DEFERRED — it is SIP-federation-shaped).
```

---

## 6 · The recurring shape, and why the OS analogy fits

The same boundary keeps appearing because it is one problem: **a low-authority proposer hands work to
an executor across a human-confirmed authority boundary.**

```text
THINK (propose, read-only)  ⟷  the executor (run)          across  the privilege boundary (human policy).
capability (designation)    ⟷  the mechanism (mint)         across  the kernel trap.
horizon (a view)            ⟷  the full table               across  the MMU / page fault.
the user's request          ⟷  the system                   across  authentication (the principal).
```

And it is **self-similar with the development process** (see the project README): *Claude proposes →
human confirms → Codex executes* IS *THINK proposes → the user confirms → the executor runs*. Where the
process chafes, look for the seam in the system, and vice versa.

The OS/CPU grounding is not decoration — each part found its true shape there: socket → capability
lookup, fd → cut-ID handle, page table → capability table, address space → horizon, loader →
Materialize, dynamic linker → REMEMBER, trap → the privilege/kernel crossing, SIP (not PSTN) → smart
entities coordinating under a dumb authority.

---

## 7 · How the work is run

Two lanes on separate git worktrees (a live build must not share a tree with active spec iteration —
we learned that the hard way mid-arc):

```text
SPEC LANE   skyra/specs/<date>/ — dated specs, decisions.md (numbered), questions.md, grouped by topic
            and by generation (skyra-dna-vN). docs win over code.
IMP LANE    skyra/src/skyra/skyra-dna-vN/ — a FRESH build per generation (no imports). the work order
            (overnight.md) + interface/struct proposals (rejectable) + the two deliverables
            (boundary-crossings.md, execution-trace.md).
ROLES       Claude proposes (rejectable) · the HUMAN confirms (the gate; decisions recorded only after) · Codex executes.
```

---

## 8 · Where it wants to go

```text
NEXT BUILD        v11: prove the two paths — COLD (kernel compiles + mints) then HOT (CTU verifies offline
                  + executes), Body/Context split, the barrier's fd/socket tables. interface proposal is ready.
THINK/ACT         the co-authored signature (THINK writes the runnable half; ACT negotiates the type) is
                  still DIRECTION, not confirmed — the biggest open question of execution.
CAPABILITIES      concretize the grant/macaroon (what a Caveat is), revocation (view-coherence), leased grants.
BOOTSTRAP         the minimal handshake channel; authentication internals (stubbed today).
DISTRIBUTED KERNEL  deferred: "the kernel is the distributed part of the system" — SIP-federation across hosts.
COMPUTATION       Starlark as the hermetic COMPUTE leaf (D1 of v11) — separate computation from dataflow.
```

The floor is solid: the authority model (human/kernel/CTU), the capability/horizon (MMU), the cold/hot
paths, and `acquire` all hold together and map onto proven systems. The frontier is making the
execution model *run* — turning the v11 interface proposal into the next throwaway that finds where
these boundaries actually cross.
