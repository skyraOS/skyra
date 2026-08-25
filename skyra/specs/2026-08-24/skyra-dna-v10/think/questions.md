# THINK — questions

**Status: CARVING.** Deliberately incomplete. Working surface for one reframe: THINK as the **primer
that creates possibilities**, not a re-read of executor machinery. Nothing here is settled unless it
says SETTLED. A wrong line recorded as settled is worse than a blank — leave edges blank rather than
guess.

---

## FOCUS — the minimality question

> **What is the minimum set of responsibilities THINK can hold to enable the creation of ANY request
> the user may have — while never executing the system?**

The governing balance: THINK is the medium the user talks to the system through, so it must **KNOW**
the system's capabilities; it must not **EXECUTE** them. The whole carving is drawing that know/do
line at the network and syscall level.

**Working hypothesis for the minimal set (NOT settled):**

```text
1  PERCEIVE the capability surface   read what the system can do (internal).            read-only
2  RECALL + assess by history        read memory: seen-before? track record? = SIGNAL.  read-only
3  BOUND the possibility space        input + surface + history → bounded candidate graph. authoring
4  PROPOSE                            return an unpaved tree + risk signals; hand off.   no run/probe
```

All four run under a **read-only horizon**: THINK may perform idempotent READS — including discovery
via *admitted* resolvers (search / resolve) — but can never MINT or mutate. Universality is carried by
**(3) bounding-then-composing + the kernel's MINT at paving** — NOT by THINK holding broad reach.

### Sub-question 1 · What networks must THINK reach?

```text
INTERNAL FABRIC   read the capability surface — what entities/hosts/remotes the system can serve.
MEMORY / HISTORY  read experience — the risk signal (seen-before + track record). REMEMBER-read.
USER CHANNEL      THINK is the user's medium: receive unstructured input; (OPEN) clarify by asking.
OUTBOUND WORLD    idempotent READS via ADMITTED resolvers only (search / resolve), under a read-only
                  horizon. a NEW / unadmitted endpoint is OUTSIDE the horizon — uncontacted — until
                  registration admits it (the accept gate). the EXECUTOR makes effectful contact.
```

### Sub-question 2 · What syscalls / kernel methods must THINK hold?

```text
PRINCIPLE            THINK holds only KNOW methods, never DO / ORIGINATE methods.
HOLDS (know)         ORIENT (live topology) · REMEMBER-read (experience + history/risk) ·
                     capability-registry READ (what is registered) · BOUND (narrow only).
NEVER (do/originate) MOVE (issue/transport) · MINT (authority origination) ·
                     ACTIVATE/MATERIALIZE (pave) · REMEMBER-commit (durable write) · TRAP-for-effect.
OPEN                 is the capability-registry read an existing projection (Interpret's tables) handed
                     to THINK as an observation, or a new kernel READ method THINK holds?
```

---

## RISK — the system owns the CATEGORY, the user owns the TOLERANCE

The system has **no opinion on the user's risk tolerance.** It **does** have an opinion on what the
**categories** of risk are. The category is structural and shipped as a default; it is **not frozen —
the user can edit it.** Tolerance is pure configuration over those categories.

```text
SYSTEM   names the risk CATEGORIES (the taxonomy).   shipped default, user-EDITABLE, not frozen.
USER     sets TOLERANCE per category.                 pure config; the system has no opinion here.
THINK    computes the risk SIGNAL — a READ, never a probe.
PAVING   applies the user's POLICY to the signal: admit / gate / require approval.   (kernel)
```

Endpoint risk is drawn **binary, from history — not from a live test:**

```text
HAVE WE SEEN THIS BEFORE?   novel vs known   — dimension one
WHAT IS ITS HISTORY?        track record     — dimension two
```

Both are **REMEMBER-reads.** Novelty is not *tested*, it is *remembered* — THINK never contacts the
endpoint UNDER ASSESSMENT to judge it (that would be a probe). It may still READ admitted resolvers to
DISCOVER candidates; assessing their risk stays a history read, not a live test. Liveness of an
accepted endpoint is learned by the executor at dispatch (08-13: reachability is learned by attempting).

### Thought experiment — "can you get me water"

THINK must know **what** water is, **where** it is, **how** to find it. It must NOT resolve the exact
tap. Its job is to **bound the possibility space** of the request, not to resolve or execute within
it — the executor resolves the concrete. This is exactly why THINK does not test endpoints: bounding
the space is a read over capability + memory; resolving the exact is the executor's, at dispatch.

---

## The dividing line — RESOLVED by horizon

> **Testing an endpoint is not proposing.** True — so THINK does not test. The contradiction in
> "read-only + propose" dissolves under a **read-only horizon**: THINK may READ (including discovery
> via admitted resolvers) and PROPOSE, but can never MINT or mutate, and a novel endpoint sits OUTSIDE
> its horizon — uncontacted — until registration admits it. The seam is read-only not because nothing
> leaves THINK, but because nothing THINK does can originate authority or change state. **The horizon
> boundary is the accept gate.**

---

## Paving — how a proposal becomes runnable (still holds)

Independent of the probing question, the handoff shape survives:

```text
THINK tree      read (capability surface + memory + admitted resolvers) + bound + propose
                → read-only horizon (idempotent reads only). returns an UNPAVED executor tree (an
                  engram body); endpoints ride as HISTORY VERDICTS (seen/novel + track record), not
                  bindings, not liveness claims.
KERNEL (pave)   admit + MINT + bind + wire, applying the user's risk POLICY.
                a novel/never-seen endpoint enters the runtime ONLY here (the trap) — and only within
                the user's tolerance (auto-admit known-good / gate / require approval).
EXECUTOR tree   resolve deps + dispatch + effect. where EFFECTFUL contact and first contact with a
                newly-admitted endpoint happen; liveness is learned here.
```

Why this is law-consistent: admitting a never-seen endpoint is authority origination = **MINT**,
kernel-only. THINK cannot MINT (only Bound/narrow), so a proposal cannot enter an endpoint into the
runtime. Paving = **MATERIALIZE** walking the unpaved tree (lands the v9 ENGRAM BODY gap).

---

## Open — the edges still being carved

```text
THINK TREE SURVIVES — RESOLVED
    yes, on composition grounds. THINK is a top-level possibility with `discover` as a child; see
    `discover.md`. it does NOT collapse to a leaf proposer.

THE RISK TAXONOMY
    where do the risk CATEGORIES live, in what form, and how are they user-editable without becoming
    user opinion the system enforces? the system ships the taxonomy; the user edits tolerance (and
    maybe membership). mechanism unwritten. is binary novelty + history enough, or dimension one of N?

IS THINK STILL A PRIMITIVE? (leaning: no)
    THINK is a top-level POSSIBILITY (discover its child), so the CTU `Think` method likely dissolves
    into "run a think-possibility." confirm, and decide where the propose-verb leaf lives.

THE HANDOFF ARTIFACT
    what exactly is the returned root — an engram body? a move graph? how much is bound vs left for the
    executor to resolve? the THINK/executor line IS the bound-space / resolve-concrete line; draw it.

UNSTRUCTURED INPUT → BOUNDED SPACE
    the front-door mechanism: prose → what/where/how → a bounded candidate graph. unwritten.
```
