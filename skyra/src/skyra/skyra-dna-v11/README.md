# skyra-dna-v11 — the execution model: cold path, hot path

**Date:** 2026-08-24
**Status: INTERFACES + THROWAWAY.** A fresh, boundary-finding build of the v11 execution model. Design
authority is `skyra/specs/2026-08-24/skyra-dna-v11/` (on `2026-08-23/`). Docs win over code.

## The central question

> **Can a request run the two paths — COLD (THINK: the kernel compiles it into a Body and mints a
> capability) then HOT (ACT: the user presents the capability and the CTU executes that Body as a
> Context) — with Body (compiled) and Context (live) split, the Move barrier and its fd/socket tables,
> the horizon carried as a capability, and the CTU nearly-stateless behind a kernel that owns the table?**

```text
COLD (new request)     move(entity=kernel, …) → THINK compiles a Body → kernel PROVISIONS + MINTS a
                       capability → runs it → REMEMBERs. returns the capability the user now holds.
HOT (remembered)       move(entity=engram, +capability) → CTU VERIFIES the capability offline →
                       MATERIALIZE a Context from the cached Body → ACT (Drive). no kernel round-trip.
```

## The one idea

**Two things were fused in v10 and split here: the Body (compiled, static, content-addressed — THINK's
output) and the Context (live, mutable, per-run — what the CTU drives).** MATERIALIZE is the loader
between them. A request is COLD the first time (through the kernel: compile + provision + mint) and HOT
thereafter (direct to the CTU with a held capability, verified offline). REMEMBER is the bridge. The
horizon is a capability the *user* carries — so the CTU stays stateless and the hot path needs no
kernel. It is OAuth + macaroons at the request level.

## The packages (my proposals — rejectable)

```text
causal/   the crossing + core types: Body vs Context, Move/Transfer, Capability (macaroon) + Horizon,
          Entity→Host→Address, Outcome incl. FAULT.
think/    THINK — Compile(request) → Body. the cold-path compiler (proposes; never executes/mints).
kernel/   the AUTHORITY — Table (Mint/Provision/Remember), Barrier (egress/ingress + socket table),
          Identity, Handshake(=acquire). owns the full table; provisions views.
ctu/      the CTU — Move (entry), Drive (ACT), Materialize (loader), Verify (offline cap), Drivers.
```

## Reading order

```text
1  README.md              (this) — the question, the one idea, the layout.
2  interface-rationale.md  every type, the term it comes from, and the seams already known.
3  causal/causal.go        the core types, cited to the v11 docs inline.
4  think ▸ kernel ▸ ctu    the cold-path compiler, the authority, the hot-path core.
5  claude-ideas.md         the SHAPE is a set of proposals — attributed, every one rejectable.
```

## What is mine vs yours

The **terms** are from the v11 docs. The **shape** — the Body/Context split, Capability-as-macaroon,
Materialize as the loader, Move as the CTU entry, the cold/hot split — is a set of proposals, each
attributed and rejectable in **`claude-ideas.md`**. Reshape any; if you do, say why. The implementer
fills bodies + writes the traces and reports the trail in `boundary-crossings.md`.

## What to prove — the target trail

```text
1  COLD RUNS        move(entity=kernel) → THINK compiles a Body → kernel provisions + MINTS a capability
                    → materialize → Drive → REMEMBER. returns the held capability + a result.
2  HOT RUNS         the SAME request → move(entity=engram, +capability) → CTU VERIFIES offline →
                    materialize from the cached Body → Drive. NO kernel call (assert it).
3  FAULT PROVISIONS a move to a target NOT in the horizon FAULTs → kernel provisions (or denies) → re-drive.
4  BODY ≠ CONTEXT   two HOT runs of one Body produce two Contexts; the Body is unchanged/shared.
5  BARRIER          a cut parks a cut-ID (fd) in the possibility's fd table; the socket table is kernel-side;
                    a Response pops by fd; a Request establishes a return route (no fd).
```

The morning read is `boundary-crossings.md` + `execution-trace.md` — see `overnight.md`.
