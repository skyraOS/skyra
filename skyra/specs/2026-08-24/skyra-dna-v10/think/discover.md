# discover — the possibility shape

**Date:** 2026-08-24
**Status: SHAPE, converging.** THINK's discovery, as a possibility. `search` and `resolve` are
settled leaves. `test` is NOT here — it moved to the privilege side
(`../potentials/user-as-privilege-boundary.md`). Builds on `../research/decisions.md` D1–D2 and
`../research/pipeline.md`.

---

## The shape

THINK is a **top-level possibility**. `discover` is a **child** of it (one child among possibly
several). `discover` takes an observation and returns a **registration proposal** — the artifact the
user accepts to register found entities/endpoints into the userspace (entity) block.

```text
THINK possibility  (read-only horizon of ADMITTED reads; holds NO ask_permission by default)
├── discover(obs Observation) -> RegistrationProposal
│     ├── search(obs)      → candidate NAMES        leaf · deterministic control · LIVE values
│     └── resolve(names)   → candidate ENDPOINTS    leaf · deterministic control · LIVE values
└── test  → ConnectionContract    (sibling; a safe-probe READ — see ../potentials/test.md)
        discover → test                       DIRECT for an admitted (known) endpoint
        discover → ask_permission → test      for a NOVEL endpoint (first contact = POKE grant; MINT is later, at admit)
```

### It is a dataflow chain, not independent siblings

The three children are ordered by **register dependency**, exactly the reorder-buffer shape:

```text
search   writes  candidates(names)
resolve  reads   candidates(names)   writes  endpoints(triple + address)   ← consumes search
                 endpoints flow to `test` (THINK's sibling child): DIRECT if admitted, via ask_permission if novel
```

So `discover` is a small dependency graph, not a flat list. `resolve` cannot fire until `search`
lands its names.

---

## Leaves: `search` and `resolve`

```text
search    intent / description → candidate NAMES.   "you can't connect to what you don't know exists."
          the WIDE step (a search service, a directory). finds names you do not already hold.
resolve   a NAME → candidate ENDPOINTS.             getaddrinfo (DNS) / ENUM (E.164) / a lookup API.
          the per-namespace resolver (D1): dispatch on the name's namespace to pick the resolver.
```

**"Deterministic leaves" means deterministic CONTROL, not deterministic VALUES.** Each leaf is a
simple dispatched call — no sub-tree. But its *output is live*: a search returns different results
tomorrow; DNS changes. The non-deterministic result must be **captured** (provenance) and fed to the
**novelty / history read**, never assumed stable.

---

## Both leaves run under THINK's READ-ONLY HORIZON

`search` and `resolve` do hit the network — but only as **idempotent reads via ADMITTED resolvers**
(the search service, DNS, a lookup API are already-registered capabilities). This does not break the
read-only seam, because:

```text
ONE ENGINE            DRIVE runs THINK like any possibility — under a read-only / propose HORIZON.
READS, NEVER MINT     the horizon permits idempotent reads; it cannot MINT or mutate.
NOVEL ENDPOINTS OUT   the endpoints `resolve` FINDS are OUTSIDE the horizon — uncontacted — until
                      registration admits them. the horizon boundary IS the accept gate.
```

So the found endpoint is never touched by `discover`. First contact is `test` — THINK's sibling child
(not a leaf of `discover`). `test` runs directly on an admitted endpoint; a NOVEL first contact faults
to `ask_permission` (D2). The privilege boundary is *admission of a novel endpoint*, not `test` itself.

---

## The return: `RegistrationProposal`

```text
RegistrationProposal   a bounded set of candidate (capability + endpoint) records,
                       each tagged with its NOVELTY VERDICT (new vs seen-before) + history,
                       proposed for registration into the userspace (entity) block.
```

This is the `addrinfo`-shaped return from `pipeline.md` made concrete: each candidate carries the
triple + the address (as a verdict, not a live binding). `discover` does NOT register it — it holds no
`ask_permission` (D2). A novel candidate's first contact faults for a **POKE** grant — novelty is
known at the poke, and **poke is NOT the mint**. The **MINT** happens later, at `admit`'s **REGISTER**
grant, which traps to the kernel (D3, D5); that kernel mint is the only path a novel endpoint enters
the runtime.

---

## Deliberately not here

```text
test         THINK's SIBLING child (not a discover leaf). safe-probe READ; privilege is conditional —
             direct if admitted, ask_permission for novel first-contact (D2). see ../potentials/test.md.
accept UX    how the user is shown the proposal and accepts. downstream of the shape.
propose-verb where THINK's leaf-level propose lives if `Think` the primitive dissolves. see questions.md.
```
