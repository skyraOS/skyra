# The distributed possibility — execute the possibility, ship the slice

**Date:** 2026-08-22
**Status: WORKING.** Settled in conversation on 2026-08-22. The CTU-owned core of the distributed
possibility model. It takes a position on the question `2026-08-18/entity-ctu-boundary/decode-and-run.md`
left open as **"LOCAL vs DISTRIBUTED — THE LARGEST QUESTION"**: a possibility does not run in one
place and message elsewhere — it is **partitioned across CTUs**, and only the slice a CTU needs
travels to it. The two mechanisms this leans on live in their owning components:
`../kernel/remember-gather.md` (commit) and `../fabric/signed-provenance.md` (the cut on the wire).
ORIENT — the third read named below — is `orient.md`.

_Instantiates `2026-08-18/the-quadrant.md` — the cross-CTU case makes **inter-CTU trust** the main event, not an edge._

> **The CTU executes a possibility. The possibility carries its own state, so "run it" needs no
> other argument. When a possibility hits a dependency a CTU cannot resolve locally, the CTU does
> not fetch the answer back — it forwards the slice of the possibility that needs the other CTU,
> under one shared identity. Execution roams by partitioning; a MOVE is the cut between partitions.**

---

## The unit of execution is the possibility

Give the CTU a possibility; it executes it. Nothing else is passed, because a `Possibility` already
holds its own state — frame, the observations in scope, its outstanding set, its provenance. This is
the habit to build: **the CTU is the executor, the possibility is the program-and-state.** It makes
the CTU nearly stateless between possibilities — it holds its dispatch tables
(`2026-08-18/dispatch.md`), its kernel binding, and its signing key, and nothing of the computation
itself. The live state lives **in the possibility, wherever the possibility currently is** — not in
a machine-wide table on the CTU.

---

## Two ways a possibility could distribute — and the one chosen

```text
A  SHIP THE SLICE   one identity names the whole possibility; it exists as FRAGMENTS across CTUs.
   (chosen)         each CTU holds only the instructions and dependencies it must resolve, and
                    forwards to the next CTU only the slice that CTU needs. This is dataflow-graph
                    partitioning: the graph is cut by locality, and each cut edge is a MOVE.

B  SHIP THE WHOLE   the entire possibility travels intact to each CTU; each runs what it can and
   (degenerate)     forwards the whole thing onward. This is the mobile-agent shape.
```

**A is chosen.** B is the **degenerate case of A** — a possibility whose partition happens to be all
of it — kept available for a possibility too small to be worth cutting, or a deliberate whole-state
checkpoint. Choosing A closes no door.

### Why A — and the reason is not bandwidth

Shipping less is a bonus. The deciding reason is **least-knowledge**. Under B, every CTU that touches
the possibility sees the *whole* frame, scope, and provenance — even a CTU doing one small sub-task.
Under A, a CTU sees **only the slice it is asked to execute**, and the horizon that travels with a
slice can be **attenuated to that slice** — which is exactly what the token is for
(`2026-08-18/kernel-ctu-boundary/horizon-and-root.md`: append-only narrowing caveats). Least
authority and least knowledge fall out of the same mechanism. B fights the horizon model; A *is* it.

### The cost lands on commit, which is the right place

```text
                   DISTRIBUTE (hot path)          COMMIT / OBSERVE (rare path)
A  ship the slice  cheap — send only the slice     EXPENSIVE — the cone spans CTUs; REMEMBER must
                                                    GATHER it back to the root's kernel.
B  ship the whole  expensive — ship everything      cheap — the whole cone is already present.
```

REMEMBER is the rare path (root-only, durable); forwarding a slice is the hot path. Optimize the
common case; pay at the uncommon one. The gather mechanism and why content-addressing makes it
tractable are `../kernel/remember-gather.md`.

---

## The distributed possibility IS the possibility tree

A is the possibility tree of `2026-08-18/entity-ctu-boundary/decode-and-run.md` made literal across
hosts:

```text
POSSIBILITY        a causal graph of instructions. its identity names the whole across every CTU.
FORWARDED SLICE    = a child possibility on another CTU, under the same shared identity.
CROSS-CUT EDGE     = a MOVE. the intermediate observation in transit from one partition to the next.
                     its correlation is (possibility identity, edge). detail: ../fabric/signed-provenance.md.
FAN-OUT            a node with two outgoing cut edges. no longer a strain: it is two MOVEs leaving.
FAN-IN (JOIN)      a node whose inputs arrive as MOVEs from two CTUs. the join node STAYS PUT;
                   completed slices RECOGNIZE back to it.
```

Fan-out — the hard seam when a *whole* possibility roams — stops being a problem: a whole possibility
cannot be in two places at once; a *partitioned* one already is. The return leg (RECOGNIZE) is not
eliminated, only confined to the join, where a completed slice reports its observation to the node
that consumes it. Migration for the spine; a MOVE at every cut. The scheduler's WHERE
(`decode-and-run.md`) gains one decision — **PUSH** (forward the slice to the resource) or **PULL**
(request the resource here); migration is the remote case of one scheduled step, not a new machine.

**A child's parent link IS the cut that created it** (confirmed by the v8 build): a forwarded slice
has no local pointer to its parent — the parent possibility deliberately does not travel — so its
parent link is the **MoveID of the incoming cut**, and a nil parent is structurally the root. This
is cleaner than a tree pointer, and it is *why* arenaID and possibility identity stay off the
response: the one-hop link home is the edge itself, resolved locally at the issuer. (This refines
the arena section below — `Parent` is the immediate link, but across a cut that link is a cut edge,
not a pointer.)

---

## Every possibility carries its arena — the root, denormalized

A possibility is ephemeral; the arena it commits into — the root (`THE ARENA is the ROOT`,
`2026-08-18/.../primitive-details.md`) — is durable. The binding lives **on the possibility**, as an
`arenaID`: its back-reference to its owning root. Two reasons stack:

```text
EPHEMERAL     the CTU is near-stateless (above); the possibility is the only carrier of live state,
              so the arena binding cannot live ambiently on the CTU.
DISTRIBUTED   the root may sit on ANOTHER CTU. "which arena do I belong to?" cannot be DERIVED by
              walking Parent to the top — that walk now crosses hosts. it must be CARRIED.
```

`arenaID` is not `Parent`. `Parent` is the immediate one-hop tree link; `arenaID` is the far end of
the whole Parent chain — the commit anchor. On one CTU you could derive it; across CTUs you
denormalize it onto every fragment so it is answerable locally.

It does **not** travel on the request. The MOVE envelope carries the cut-edge identity (for
correlation and return); it does **not** carry `arenaID`. The home arena stays home: a forwarded
slice never learns it, the return leg rejoins it (RECOGNIZE correlates into the home possibility,
which holds the `arenaID`), and the remote slice could not use it anyway — REMEMBER is root-only and
home. Keep it a **back-reference, not a commit handle**: a non-root holds `arenaID` to know where it
will be gathered and to scope its authority, but REMEMBER stays *structurally* root-only (you REMEMBER
the arena you ARE — `Parent == nil` — not one you merely name). That preserves the house
inexpressibility "a non-root names no arena" in the sense that matters: names-**for-commit** stay
root-only.

---

## Three distributed reads over one possibility

The possibility is partitioned once; everything else is a read over that partition along the same
cut edges. There are three:

```text
EXECUTION        forward the slice        the hot path.                 (this document)
THE CONE         gather toward the root   REMEMBER's subject — the committed PAST.  ../kernel/remember-gather.md
THE OUTSTANDING  read local, pass along   ORIENT's subject — the in-flight FRONTIER.  orient.md
```

REMEMBER and ORIENT are **duals**: REMEMBER gathers the cone (completed causal history) toward the
root to commit; ORIENT reads the outstanding frontier (what is still in flight) out toward the leaves.
Same topology, opposite temporal ends. Each is developed in its own component doc.

---

## What this reframes in the 08-18 set

```text
decode-and-run.md   RESOLVES its largest open item in one direction: the eligible set MAY span hosts;
   "LOCAL vs        the scheduler gains PUSH vs PULL. The residual estimate problem (liveness,
   DISTRIBUTED"     latency) it names is still real; this doc does not price it.
changes.md          (src/skyra/skyra-dna-v7/changes/changes.md, throwaway) its MoveID-two-legs
                    conclusion holds: the two legs are the two directions of a cross-cut edge, and
                    the outbound leg carries a slice, not merely an instruction.
```

No new primitive is introduced. MOVE is still the only transport; a cut is a MOVE. GREET admits an
arriving slice; RECOGNIZE admits a completed slice at a join. The eight are untouched. Shared
vocabulary (possibility identity, slice, arenaID, cone, frontier) is `../causal/terms.md`.

---

## Open (CTU-side)

```text
PUSH vs PULL POLICY   the scheduler's new choice is stated as mechanism, not policy. WHEN to migrate
                      the slice vs pull the resource (size, anchoring, liveness) is deferred.
MOVING JOINS          the join node is taken to STAY PUT. a migrating join makes completed slices
                      chase a moving target (forwarding pointers). not needed yet; a recorded choice.
PARTITION IS LAZY     the cut is discovered as execution flows (a CTU forwards the sub-slice gated on
                      a dependency it cannot resolve), not planned up front. an ahead-of-time cut for
                      a known wide fan-out is unexplored.
THE EDGE CONTRACT     a cut carries an intermediate observation that the join checks against a
                      contract; the typed cut is owned by ../fabric/signed-provenance.md and the
                      set's still-undefined CONTRACT object. the next real interface.
```
