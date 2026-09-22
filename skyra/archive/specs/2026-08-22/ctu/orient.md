# ORIENT — the distributed frontier read

**Date:** 2026-08-22
**Status: WORKING.** ORIENT under partitioning. The possibility is sharded across CTUs
(`possibility.md`), so its outstanding set is sharded too. ORIENT is the third of the three
distributed reads — the dual of REMEMBER's cone gather (`../kernel/remember-gather.md`).

> **ORIENT does not observe a distributed possibility from outside. ORIENT is distributed: one
> partial read per fragment, composed by propagation along edges the possibility already has.
> "Read own outstanding set" is the local term; "pass it along to the next runtime" is the recursion.**

---

## Local by construction

`2026-08-18/.../primitive.md` gives ORIENT as "read own outstanding set." Under partitioning, "own"
is inherently *this fragment's* outstanding — there is no global observer of a whole possibility
without composition. That is consistent with observable causality: no global clock, no global instant.

## The recursion

```text
ORIENT(me) = my local in-flight  ∪  ORIENT(each slice I am outstanding on)
```

Read own → the local term. Pass it along to the next runtime → the recursion over outstanding edges.
A whole-possibility ORIENT is that fold over the outstanding-edge tree, mirroring the cone's fold over
`Causes` — same recursion, different lens (the cone is the committed past; the frontier is the
in-flight future).

## What it is, formally — and why that is good news

"Read local, pass it along" is a **distributed snapshot** (the pass-along is marker propagation), and
"is the whole possibility done?" — outstanding globally empty — is **distributed termination
detection**. Both are tractable here because the possibility is a **diffusing computation**: a tree
that fans out via MOVE and folds back via RECOGNIZE. That is the exact shape **Dijkstra–Scholten**
termination detection is built for — a node is quiescent when it is locally done AND every child has
acknowledged, acks flowing up the tree. So the whole-view is not new machinery: it is **ORIENT
composing with RECOGNIZE**, the same acks that fold results back up the tree.

## The honest limit: a cut, not an instant

A propagated ORIENT is a consistent CUT, not an instant. With no global clock, a MOVE in flight sits
on a channel, and a naive "every local set is empty ⇒ done" is wrong while work is still in transit.
Fidelity is per use:

```text
APPROXIMATE FRONTIER  scheduling, "roughly what am I waiting on"  — cheap local propagation suffices.
CONSISTENT CUT        global termination / deadlock safety        — needs the ack discipline above
                      ("is this possibility TRULY done?")           (in-flight counted, not just local).
```

## Lead: this may be the escalation substrate

`2026-08-18/README.md` calls ESCALATION "the largest gap" — a failed return once woke the parent;
bulk teardown removed the mechanism and nothing replaced it. This is the first mechanism to touch it
from the mechanism side: a stalled or dead subtree shows up as a **hole in the propagated ORIENT** —
a slice that never acks — and the Dijkstra–Scholten ack-up-the-tree is a natural place to hang "a
return failed, wake the parent." Recorded as a thread to pull, not a resolution.

---

## Open

```text
ORIENT FIDELITY   approximate frontier vs consistent cut. which one each caller needs — scheduling
                  wants approximate, termination/deadlock wants the cut — is not enumerated, and the
                  ack discipline for the cut is sketched, not specified.
ESCALATION HANG   whether "wake the parent" actually hangs on the ack path, and what it wakes into,
                  is a lead, not a mechanism. owned jointly with the escalation gap in the 08-18 set.
```
