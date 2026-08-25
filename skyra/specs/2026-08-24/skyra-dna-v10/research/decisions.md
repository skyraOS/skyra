# Decisions — discovery

**Date:** 2026-08-24
**Status: DECISIONS.** Settled calls that came out of the discovery research (`pipeline.md`,
`socket-triple-to-fd.md`). Only firm decisions live here; open threads stay in `../think/questions.md`.

---

## D1 · `getaddrinfo` is a pillar of discovery — SETTLED

We build discovery on `getaddrinfo`. It is the operation that takes a **human-readable name** (a
`.com`, a service, an identity) and returns a **bounded set of candidate endpoints** — the shape
THINK is meant to produce.

```c
int getaddrinfo(const char *hostname, const char *servname,
                const struct addrinfo *hints, struct addrinfo **res);
// *res → linked list of struct addrinfo, each = (ai_family, ai_socktype, ai_protocol) + ai_addr
```

**Why it earns pillar status:**

```text
NAME → ADDRESS       it turns the human-meaningful into the machine-addressable — the front of discovery.
BOUNDED SET          it returns MANY candidates, not one — a possibility space, not an answer.
TRIPLE + ADDRESS     each result carries the capability triple AND the location in one record — the
                     exact shape of a potential move.
```

### Refinement: `getaddrinfo` is STAGE ONE, not the middle

Earlier `pipeline.md` placed it between `socket` and `connect`. Reordered: **name resolution comes
first.** Discovery starts from a name, not from a nameless capability. Working order:

```text
getaddrinfo(name, service) → candidates    RESOLVE the name → bounded (triple + address) set   [DISCOVERY]
socket(triple)             → fd            take a candidate's triple → a nameless handle
connect(fd, address)       → outcome       test/bind the candidate's address (executor, at dispatch)
```

`pipeline.md` should be updated to this order when we next work it.

**Scope of this decision:** it fixes `getaddrinfo` as a *pillar abstraction* — the resolver shape
discovery is built on. It does NOT yet settle: which horizon THINK holds (static vs live), the
`hints` input contract, who owns the live network resolve, or the novelty/admission gate. Those remain
open in `../think/questions.md`.

---

## D2 · Permissions are a horizon-gated `ask_permission(obs)`, not recursion to root — SETTLED

When a possibility needs a permission it lacks, it makes a **direct call to the resident privilege
entity** via a capability `ask_permission(obs)` — it does **not** bubble the need up the parent chain
to root. Detail in `../potentials/user-as-privilege-boundary.md`.

```text
CHOSEN  (B)   ask_permission(obs) — a trap-MOVE naming the resident privilege entity. present on some
              (privilege-bearing) levels; ABSENT from THINK's read-only horizon. parks like any cut.
REJECTED (A)  recurse the need up the parent chain to root, grant flows back down.
```

**Why B:**

```text
ALREADY THE MACHINE   naming a privileged resource already TRAPS (v9 Interpret); horizon decides if the
                      name resolves. A needs NEW machinery.
NO WAITING PARENTS    A re-introduces the synchronous parent chain v9 killed (rationale seam #1): every
                      intermediate parent forwards a permission it has no stake in. B is one cut, one park.
HORIZON-UNIFORM       authority is already horizon-gated; "ask for authority" is the same rule, not a new one.
POSITION-INDEPENDENT  trap from any depth, like a syscall — authority is not coupled to call depth.
```

### The scarcity guardrail (part of the decision)

`ask_permission` must stay **scarce** — held by the user-root and only levels it *explicitly* grants
it to; never an ambient default. This recovers A's one real benefit (confinement: a chokepoint that
sees every request) via *scarcity + a governing entity + `obs` carrying context*, instead of via tree
position. Scarcity is nearly free: **`Bound` only narrows**, so a child can never acquire
`ask_permission` it wasn't given — the guardrail is simply "don't seed it broadly."

**Still open:** how the privilege entity is addressed from any depth; scope-of-grant (accepted X for P,
does Q re-ask?); concurrency; poke economics. See `../potentials/user-as-privilege-boundary.md`.

---

## D3 · Privilege (principal) and kernel (mechanism) stay SEPARATE — SETTLED

The authority boundary factors into two components that **do not collapse**, even for a single user:

```text
PRINCIPAL   privilege — DECIDES: novelty + the user's policy + wake-human. the user's resident agent.
MECHANISM   kernel    — MINTS: authority origination, trap-reached, NEUTRAL (bakes in no user policy).
```

**Decision: keep them separate, always** — not a single-user optimization to fuse.

**Why:** the boundary between *deciding* and *originating* is real. Collapsing them into one component
would teach a reader — or a **model reasoning over the system** — that the boundary does not exist, when
it does. A visible boundary is load-bearing: it marks *where the user's authority is exercised* apart
from *where neutral origination happens*, and those are different acts with different owners. Keep the
seam legible.

This closes the v10 collapse question (seam S8 / claude-idea 4): **split, not a seam.** Admitting a
capability is kernel domain (the MINT); privilege only decides and traps to it.

---

## D4 · A horizon IS its issued grants (`Horizon = []Grant`) — SETTLED (folds the v10 reshape)

The v10 build broke the predicted seam #3: `Horizon.AskPermission bool` could not express authority. A
bool cannot distinguish *poke endpoint X for purpose P* from *register X*, nor stop a grant being
replayed against endpoint Y, and it wrongly implied THINK holds an ask-switch.

**A horizon is nothing but its issued grants.** Horizon and grant are the same thing:

```text
Horizon = []Grant           each Grant.Scope names { endpoint, purpose, action(Poke|Register) }.
SCARCITY IS STRUCTURAL      a possibility starts with ZERO grants (THINK holds none). Bound still only
                            NARROWS — it can drop grants, never add one. authority enters only by a
                            grant the boundary issues.
```

This supersedes the `AskPermission` bool: the horizon carries the exact scoped authority a possibility
was issued, and "read-only THINK" is simply "starts with no effect grants."

---

## D5 · Privilege is the TRAP AXIS, not a third block — SETTLED

Trust **blocks** are exactly two: **Host** (BELIEVED) and **Entity** (CHECKED). **Privilege is
orthogonal** — the vertical **trap axis**: a well-known trap address (`skyra://privilege/user`),
reached by a fault / name-miss, **never a peer resource the scheduler selects.**

```text
TRUST-BLOCK AXIS   Host | Entity     — who runs it, and whether its output is checked. (dispatch.md)
TRAP AXIS          Privilege          — orthogonal; crossed by a fault, not chosen by placement.
```

Corrects a v10 code proposal that made `Block{Host, Entity, Privilege}`; aligns with
`2026-08-23/execution-cycle.md`. Note this is the **trust-structure** axis — distinct from D3's
**authority-owner** axis (principal *privilege* decides, mechanism *kernel* mints). Two axes; both hold.
