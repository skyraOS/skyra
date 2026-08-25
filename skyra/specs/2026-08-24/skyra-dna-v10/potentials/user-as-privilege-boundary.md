# User as the privilege boundary

**Date:** 2026-08-24
**Status: SHAPE, converging.** How authority is originated in the discovery/potential flow. Builds on
`../think/discover.md` and `../research/decisions.md`. The core move: the privilege boundary is a
**resident, addressable entity**, not a return path.

---

## The claim

> **The user is the privilege boundary — the single origin of authority (MINT / register /
> first-contact).** It is a **system-resident, addressable ENTITY**, reached by a CALL when a
> possibility needs a permission it lacks — NOT by recursing back to the request's root.

---

## Not the return path

A user request starts a possibility. That possibility does **not** have to unwind back to root to get
permissions. When *any* possibility, at *any* depth, hits a permission it doesn't hold, it **MOVES to
the privilege entity directly** — the way you trap to the kernel from wherever you are, rather than
returning to `main()` to make a syscall.

```text
REJECTED   escalate UP to root → root grants → return DOWN.    couples authority to CALL DEPTH.
THIS       call the privilege entity directly, from any depth.  authority is AMBIENT + ADDRESSABLE.
```

The privilege entity is resident **in the system and user blocks**, so it is reachable from either
side without special control flow. Giving a request is not the same as having to route grants through
the request's root.

---

## It is just a cut that may park

A permission request is the capability **`ask_permission(obs)`** (D2) — no new machinery, just an
ordinary **cut** naming the privilege entity:

```text
possibility needs a permission it lacks
   → MOVE (cut) to the privilege entity
   → PARKS in InFlight   (if a human is needed, it simply parks ON the human)
   → grant returns → RECOGNIZE lands it → DRIVE re-drives, now holding the permission
```

This is exactly the reorder-buffer's park/recognize path. A grant that waits on a human is just a cut
with a slow response leg — no different in shape from a slow remote move.

---

## Fast path / slow path — the same novelty gate

```text
call the privilege entity:
   ALREADY GRANTED / KNOWN  → auto-confirm, NO human wake, immediate grant          (fast path)
   NOVEL                    → wake the human → accept → MINT / register → grant      (slow path)
```

Same gate as "the model won't wake on known types": known proceeds silently by checking the CTU
registration; only the novel wakes the human. Endpoint-novelty and type-novelty are one gate.

---

## `ask_permission` must stay scarce

`ask_permission` is **horizon-gated** and must be **scarce** — held by the user-root and only levels it
*explicitly* grants it to, never an ambient default. This is what recovers the one benefit of the
rejected recurse-to-root model (a chokepoint that sees every request): confinement comes from
**scarcity + a governing entity + `obs` carrying context**, not from tree position.

Scarcity is nearly free: **`Bound` only narrows a horizon**, so a child can never acquire
`ask_permission` it was not given. The guardrail is one line — *don't seed `ask_permission` broadly;
seed it at root and pass it down on purpose.*

---

## Principal vs mechanism (keep separate)

```text
PRINCIPAL   the privilege entity — the USER's resident agent; the authority on whose behalf we act;
            it DECIDES (auto-confirm known / wake human on novel).
MECHANISM   the KERNEL — MINT / MATERIALIZE / the registry write; HOW origination physically happens.
```

The privilege entity is kernel-*like* (you call it for authority) but it is the **principal**, and it
traps to the kernel for the actual grant.

**Keep them separate, always — SETTLED (D3).** Not a single-user optimization to fuse. The boundary
between *deciding* and *originating* is real; collapsing them into one component would teach a reader —
or a model reasoning over the system — that the boundary does not exist, when it does. It stays visible
because it marks where the user's authority is exercised apart from where neutral origination happens.
(Also required the moment there is more than one principal — multiple users, delegated authority.)

---

## Where `test` sits, and where the boundary really is

`test` is **THINK's sibling child** of `discover`, not a separate location. It is a safe-probe READ,
so its privilege is **conditional**: it runs directly on an admitted (known) endpoint and faults to
`ask_permission` only for **novel first-contact**. So the privilege boundary is *admission of a novel
endpoint* — not `test` wholesale. `discover` (search + resolve) stays read-only; `test` crosses the
boundary via `ask_permission` only when its target is novel. See `test.md`.

---

## Open

```text
ADDRESSING        the well-known name / target of the privilege entity. how is it resolved from any depth?
SCOPE OF GRANT    accepted endpoint X for purpose P — does using it for Q re-call? grants carry scope;
                  a use outside scope re-escalates.
CONCURRENCY       re-entrancy: many possibilities calling the one privilege entity at once.
POKE ECONOMICS    a novel poke can wake a human. bias the search↔test loop toward cheap unprivileged
                  discovery; escalate a novel poke only when discover is confident. (fast path saves
                  KNOWN pokes; the first NOVEL poke is the expensive one.)
SYSTEM-ADMINISTERED?  ask_permission's GRANT decision may be a SYSTEM possibility, not a model call.
                  the grant is mechanical — check the registration, apply the user's policy/tolerance,
                  known→auto / novel→wake human — so the SYSTEM can administer it; THINK / the model
                  never decides permissions. explore when we return.
```
