# Why Skyra — the grips, the invariants, the thesis

**Date:** 2026-09-01
**Status: MOTIVATION (stable).** This is the WHY the rest of the specs assume. Every other doc names a
mechanism; this one names the PAIN each mechanism pays down, so the build stays pointed at a wound and not at
an aesthetic. Written from repeated, first-hand time inside Codex / Claude Code — the failures below are
observed, not theorized.

> **THE THESIS.** Operating systems assumed ONE actor per user. That assumption is no longer true. A user now
> runs N semi-autonomous agents at once, and the OS's principal abstraction — the process, owned by one
> trusted human — has no answer for it. Skyra's principal abstraction is the **bounded, auditable agent**, not
> the process. That is the whole move; everything below is a consequence.

**The move is COUPLING, not invention.** Permissions exist. Spec/plan negotiation exists. Skyra invents
neither. It makes them FIRST-CLASS and welds them INTO the execution model, because they were only ever
bolted on beside it — a permissions prompt here, an ignorable plan-mode checklist there, neither an edge in
the graph that actually runs. Coupling them to execution is the contribution. Claiming less makes the claim
truer and the pitch stronger.

## The five grips → three invariants

```text
GRIP (the observed failure of the current meta)                          INVARIANT it forces
─────────────────────────────────────────────────────────────────────── ──────────────────────────────────
1  skills do too much. a skill is a .md doc RE-READ every turn, wired to  I  COGNITION IS COMPILED AND
   its tools with ZERO coupling. the pairing (tool + prompt) is never a      MEMOIZED, NOT RE-DERIVED.
   durable object — you re-engage the whole interface each time.
2  re-deriving the same workflow burns the same inference over and over.
   long skill docs + code recalls re-sent every step. it costs real money.

3  spec back-and-forth is not first-class. plan mode is a checklist the    II THE PLAN IS A TYPED CONTRACT;
   model can silently ignore. a FAILED PLAN should be a first-class event,    DEVIATION SURFACES AS A FAULT.
   not a line quietly skipped.

4  bash = velocity bought with ambient authority + an unreadable audit     III AUTHORITY IS CAPABILITY-BOUNDED
   trail. one `rm -rf` from disaster, and no legible record of what ran.       AND PROVENANCE IS STRUCTURAL.
5  you cannot track 5+ agents by hand. "read all the code" is a lie — it       ONE LIST.
   is impossible at 1M+ lines. it is fine to lose fine-grained control (you
   are not the one coding now) — but you MUST keep a HIGH-LEVEL view of what
   each agent is doing and, above all, what it CANNOT touch.
```

## The invariants, made concrete

```text
I   COMPILED, MEMOIZED COGNITION.  the unit is an ENGRAM = a fused (tool + prompt), not a re-read markdown
    doc decoupled from its tools. THINK compiles a request into engrams once; ACT runs them; REMEMBER persists
    them. a repeated workflow pages in WARM instead of re-deriving. → Body / engram / the WARM path.

II  THE PLAN IS A TYPED CONTRACT.  THINK negotiates a contract with the human (what I want + the horizon it may
    touch); the human ACCEPTS. the plan cannot be "ignored" because the contract is a TYPE, and the world
    violating it is a TYPE ERROR — a surfaced fault, not silent drift. deviation is not forbidden (that would
    kill adaptivity); deviation is UNTYPED-drift-made-impossible: it either fits the contract or it faults.
    → Contract / Outcome{Fault} / the deopt.

III CAPABILITY-BOUNDED AUTHORITY, STRUCTURAL PROVENANCE, ONE LIST.  an agent fires nothing without a horizon
    handed to it. the audit trail is a CONSEQUENCE of signed moves, not a log bolted on after. and there is
    exactly ONE list — the HORIZON. what an agent CANNOT touch is the horizon's complement, read off the same
    list for free; a second list is the exact overload this project exists to delete. across N agents: N
    horizons, and the only composed question a human asks is where two OVERLAP (both hold the same cap). that
    is set-intersection over one primitive, not a new artifact. → Horizon / Capability / signed Observation.
```

## The execution loop (the mechanism that couples all three)

```text
THINK   human ↔ model + spec + horizon  →  PROPOSES a contract, transforms it into multiple ENGRAMS.
ACT     compilation — a JIT. the model IS present here (≈160 tokens/engram-act + calls to sibling engrams).
        hierarchical, top-down; the LEAVES touch the world. freezes on a remembered observation.
USER    ACCEPTS or DENIES → kicks back to THINK or ACT to renegotiate. if THINK needs MORE than the horizon
        grants, it does NOT widen — it proposes a NEW contract and returns to the human. escalation is
        re-negotiation, never a silent grab.
FAULT   the world returns a NEW TYPE → the contract's guard trips → deopt → back to negotiation.
```

Note the apparent contradiction and its resolution: "execute without the model" (vision.md) vs "model during
compilation" are NOT in tension — it is a JIT. The model is present at COMPILE (cold, and again on deopt);
the WARM/HOT run executes on the compiled guards, model-free, until a new type from the world trips one.

## Recurrence is the resource

REMEMBER is the one machine that turns "you did this before" into BOTH cheaper and safer — the same coin,
two faces:

```text
FIRST run (cold)   expensive (full compile) AND loosely guarded (contract is a NAME; the structural type is
                   still being discovered) — which is exactly when a human most wants to be watching.
Nth run (warm)     cheap (skip the compile, page in the engram) AND tightly guarded (the remembered contract
                   is now a structural type; any violation faults sharply).
```

Cost and safety are not two features; they are one property that ACCRUES with repetition. This is the
economic and the safety engine of the whole system, in one sentence.

## Scope honesty — what is pain-grounded vs. what is a bet

Every grip above is SINGLE-MACHINE, MULTI-AGENT. The one-actor-per-user thesis is a single-box claim; none
of the five wounds needs distribution to bleed. So:

```text
PAIN-GROUNDED (load-bearing)   the multi-actor single-machine substrate: compiled cognition, the typed
                               contract, bounded authority + one-list audit. protect this with your life.
A BET (layered on top)         the distributed tier (offline cross-host cap verification, moves crossing the
                               barrier, content-addressed DAG replication). wanted, not yet driven by a grip.
                               allowed to lag; do not let it front-run the grounded core.
```

Ranking, not cutting. Keep the whole vision — but when time is short, the pain-grounded core is what ships.

## Status of the code

The thesis leads; the code follows. That is the ONLY order that builds anything — you do not discover a thesis
by running code, you commit to it and make the code meet it. The CTU inner core is mid-clean; the THINK/ACT
split is downstream of it. What this doc buys is a FIXED target. The risk was never a thesis ahead of the code
— it was a thesis that kept MOVING (v7→v12), so the code never got a still target to converge on. This doc
freezes it. The next revision to the thesis should come from the COMPILER fighting back, not from a nicer
ontology. Now the code part.
