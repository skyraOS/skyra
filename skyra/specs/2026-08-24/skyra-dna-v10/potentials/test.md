# test — poke an endpoint, recover its contract

**Date:** 2026-08-24
**Status: SHAPE, converging.** `test` is a **child of THINK, sibling of `discover`** (not inside it).
It is a safe-probe **READ**, so its privilege is **conditional**: it runs directly on an already-admitted
(known) endpoint, and faults to `ask_permission` (D2) only for **novel first-contact**. The privilege
boundary is *admission of a novel endpoint* — not `test` itself. Its job: given one endpoint candidate,
recover its **ConnectionContract**. Builds on `user-as-privilege-boundary.md`, `../think/discover.md`,
`../research/decisions.md` D1–D2.

---

## Where `test` sits — THINK's second child

```text
THINK
├── discover → candidates (+ novelty verdict)
└── test     → safe-probe → ConnectionContract
       discover ──▶ test                        DIRECT, when the candidate is already admitted (known)
       discover ──▶ ask_permission ──▶ test     when the candidate is NOVEL (first contact = admission)
```

`test` names its target endpoint like any move: the name **RESOLVES** if the endpoint is admitted
(known → run) or **FAULTS** if novel (→ `ask_permission` admits it, then run). Same trap mechanism as
`socket()` / `Interpret`. So the fast path — a known endpoint flowing `discover → test` with no detour —
is not a special case; it is *a name that resolves.*

---

## The one act: provoke-and-read

`test` is not a toolbox of operations. It is **one irreducible act** applied repeatedly:

> **`probe` = send a minimal, SAFE request → read the requirement the endpoint STATES back.**

You do not discover a contract by *satisfying* it; you discover it by making the endpoint **declare**
it. Each rung is stated by the endpoint when you fail to meet it — `connect` states reachability, the
TLS handshake states identity, an unauthenticated request states the auth demand, a schema query
states the shape. (This passes the 08-18 test: the sub-probes share an *irreducible act*, not merely a
consumer — a unification, not a bundle.)

---

## The ladder: four contract fragments to recover

The single act runs up four rungs, each yielding one fragment of the `ConnectionContract`, grounded in
how endpoints really self-describe:

```text
REACH               connect()                       → answers? (up / refused / timeout)
TRANSPORT+IDENTITY  TLS handshake                   → cert identity + protocol / ALPN it speaks
AUTH DEMAND         minimal UNAUTHENTICATED request  → 401 + WWW-Authenticate / OIDC .well-known
                                                       = the credential TYPE wanted (NOT the credential)
SHAPE + FITNESS     safe self-description query       → OpenAPI / GraphQL introspection / gRPC
                                                       reflection / OPTIONS = operations + req/resp schema
```

`REACH` + `TRANSPORT` fold into one "establish," so the irreducible move is *connect, then read three
declarations* — but each rung is a distinct contract field, so keep them nameable. You cannot drop a
rung: each is a distinct thing a real conversation needs.

---

## The leaves — `test`'s possibility structure

The one act specializes into **four leaf probes**, one per contract fragment — distinct graph nodes
because each reads a different declaration and writes a different register, even though all four are the
same `probe` underneath.

```text
reach       connect()                          → reachable? (up / refused / timeout)
handshake   TLS handshake                       → identity (cert) + protocol / ALPN
challenge   minimal UNAUTHENTICATED request      → auth demand (401 / WWW-Authenticate / OIDC .well-known)
describe    safe self-description query          → operations + request/response SHAPE
```

Dependency chain with a fan — the reorder-buffer shape, as in `discover`:

```text
reach ──▶ handshake ──▶ challenge
                    └──▶ describe
```

`handshake` needs `reach`'s channel; `challenge` and `describe` both read `handshake`'s transport and
usually run in **parallel** — `describe` gains an edge from `challenge` only when the self-description
is itself auth-gated. The parent `test` possibility **joins** the four fragments into the
`ConnectionContract`; the leaves only probe — assembly is the parent's (as `discover` assembles the
`RegistrationProposal`).

```text
REACH SEPARATE      reach is kept its own leaf (not folded into handshake) as the CHEAPEST gate — a bare
                    liveness check before committing to a full handshake. fold to 3 leaves
                    (establish = reach+handshake) only if minimal is preferred over the cheap gate.
FITNESS ≠ A LEAF    describe gathers WHAT operations exist; deciding whether one serves the request's
                    PURPOSE is a judgment ABOVE the leaves (and "no operation fits" is the diagnosis that
                    refines the outer search↔test loop). the leaves stay pure probes.
```

---

## The two invariants — what makes it `test` and not `use`

```text
SAFE PROBES ONLY        idempotent, non-committal verbs — HEAD / GET / OPTIONS, reflection,
                        introspection. NEVER an effectful operation. poking must not POST /orders.
DISCOVER, NEVER SATISFY  test learns the auth METHOD; it never obtains a credential or runs the real
                        operation. the auth rung returns "required: <method> @ <where> — UNVERIFIED
                        until execution." satisfying the contract is the executor's, at dispatch.
```

The second invariant is how `test` characterizes auth without holding a credential: it recovers *what
is required* and marks it verify-live. Neither invariant is droppable — together they are the line
between `test` and `use`.

---

## Output: a (possibly partial) ConnectionContract

```text
ConnectionContract
    TRANSPORT   triple + port + TLS/ALPN required
    IDENTITY    the cert / host identity it should present
    AUTH        method + credential TYPE required   (often PARTIAL: required, unverified-until-exec)
    FITNESS     which operation serves the purpose + request/response SHAPE
    SUCCESS     what response CONSTITUTES a successful call
```

Complete on the rungs `test` could read; explicitly **deferred** ("verify live") on rungs that need a
real credential. This contract becomes the endpoint-move's `Move.Contract` in the executor tree — the
rubric the executor is graded against at dispatch.

---

## It is a demand-driven loop, terminating two ways

`probe` is iterative: each read names the next probe (a 401 → go read the auth-discovery doc; the
OpenAPI → read the shape). It ends when the contract is assembled, **or** when a rung is unrecoverable
(no self-description, an auth method we can never satisfy, no fitting operation) — then it emits a
**diagnosis, never silence.** That diagnosis is the signal that feeds the outer `search ↔ test` loop to
pick a different candidate.

---

## Per-protocol dispatch (the parallel to `resolve`)

`test` is a **per-protocol prober** the way `resolve` (D1) is a **per-namespace resolver**: the *rungs
are universal* (every endpoint has reach / transport / auth / shape), but the *probe mechanism is
protocol-specific* (HTTP 401 vs gRPC reflection vs GraphQL introspection). `test` dispatches on
protocol to pick its probe vocabulary, exactly as `resolve` dispatches on namespace. Same shape, one
layer up.

---

## Minimal set, stated

```text
1 ACT        probe = safe provoke → read the stated requirement
4 RUNGS      reach · transport+identity · auth-demand · shape+fitness   (each = one contract fragment)
2 INVARIANTS safe probes only · discover-never-satisfy
OUTPUT       a (possibly partial) ConnectionContract + a diagnosis for any unrecoverable rung
SCOPE        one endpoint at a time; runs directly if admitted, behind ask_permission only for novel (D2)
```

---

## Out of scope (deliberately)

```text
CREDENTIAL ACQUISITION   obtaining/presenting the actual credential — executor + user, not test.
LIVE VERIFY              proving the contract holds with a real credential — executor, at dispatch.
CANDIDATE SELECTION      choosing WHICH endpoint to poke — that is discover / the search↔test loop.
```

---

## Open

```text
POKE-GRANT vs REGISTER-GRANT   for a NOVEL endpoint, does ask_permission grant permission to POKE
                               (learn the contract, then a SECOND decision to register) or full
                               admission upfront? poke-first learns the ConnectionContract BEFORE
                               committing — two-phase, and a narrower grant than register. scope-of-grant,
                               landing specifically on test.
```
