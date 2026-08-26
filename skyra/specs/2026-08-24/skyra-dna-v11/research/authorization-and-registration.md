# Authorization and registration

**Date:** 2026-08-24
**Status: IN PROGRESS.** Digging into what authorization and registration actually are, and the mature
patterns to borrow. Builds on `capability-table-and-horizon.md`, `move-and-the-kernel-barrier.md`, D4
(`Horizon = []Grant`), and the Starlark note's "package admission vs horizon admission."

---

## The split — they are different axes

```text
REGISTRATION   a capability comes to EXIST and be DISCOVERABLE in the kernel's full table.
               about the RESOURCE. "this is known / can be named." independent of any principal.
AUTHORIZATION  a registered capability is PROVISIONED into a principal's horizon.
               about the RELATIONSHIP (principal ↔ resource). "you may use this." per-principal, revocable.
```

A capability must be *registered* to be *authorizable*, but registration ≠ authorization: a thing can
exist and be un-usable by you; you can be authorized the moment it's registered. Two axes.

---

## Patterns for AUTHORIZATION

### 1 · Object-capability model (ocap) — the foundation; we are already building it

```text
DESIGNATION = AUTHORITY   naming a thing and using it are the same — HOLDING the capability IS the
                          authorization. (our: a move can only target what is in its horizon.)
NO AMBIENT AUTHORITY      you act only through capabilities you hold. (our: horizon = provisioned view.)
ATTENUATION               delegate by passing a NARROWED capability. (our: Bound only narrows.)
```

Sources: **seL4** (formally-verified capability microkernel), **KeyKOS / EROS**, Mark Miller's
*"Capability Myths Demolished."* Payoff: ocap **solves the confused-deputy problem** — ambient
authority is what lets a deputy be tricked into misusing power; bundling designation + authority makes
that structurally impossible. That is *why* "no move without a horizon entry" is safe, not just tidy.

### 2 · Macaroons (and Biscuit) — the grant FORMAT

Bearer tokens attenuated by **appending caveats** — narrow-only, offline, no contact with the issuer.
That is literally `Bound` (D4). Biscuit is the modern datalog-based version with offline verification.
**A grant is a macaroon:** a capability + a chain of narrowing caveats + a signature.

### 3 · OAuth2 — the FLOW, and it grounds the "API-auth stub"

```text
OAuth2                        SKYRA
resource owner (consents)  →  the HUMAN (owns policy; woken for consent)
authorization server       →  the KERNEL (issues grants under policy)
access token               →  a GRANT in the horizon (a capability)
scopes                     →  caveats / Bound attenuation
consent screen             →  ask_permission (wake human)
token revocation           →  our revocation / view-coherence
refresh / expiry           →  time-bounded grants (open)
```

Tells us grants should likely be **time-bounded and revocable** — which we had flagged open.

### 4 · ACL vs capability — we are a HYBRID

ACL authorizes by looking up identity in the resource's list; capability by holding a token. We are
**both, at two phases**: the **decision** to provision is identity/policy-based (ACL-flavored), the
**enforcement** is capability-based (the horizon — holding = using). **Decide by identity, enforce by
capability.**

---

## Patterns for REGISTRATION

### 5 · Service discovery (Consul / etcd / DNS-SD)

```text
REGISTER = name → endpoint + CONTRACT     (our acquire/admit → the ConnectionContract)
HEALTH CHECK                              (our test — registrations can go STALE)
TTL / LEASE                               a registration EXPIRES and must be re-verified (→ our timeouts)
DEREGISTER                                the inverse of admit (revocation of existence)
```

Lesson: **registration isn't permanent** — it's leased, health-checked, and expires. A stale
registration triggers a re-`test`, exactly like a discovery health-check failing. Also registration:
**dynamic linking** (resolve a symbol, patch the GOT — already used for acquire) and **ocap minting**
(the kernel creates an unforgeable reference).

---

## Definitions (working)

```text
REGISTRATION   the kernel MINTS an unforgeable reference for a resource and records its contract into
               the full table — LEASED, health-checkable, revocable. it makes the capability EXIST.
               patterns: ocap minting · service discovery · dynamic linking.
AUTHORIZATION  the kernel PROVISIONS a registered capability into a principal's horizon — decide by the
               human's policy (identity), enforce by possession (capability), attenuate by caveat,
               revocable. it makes the capability USABLE by a principal.
               patterns: ocap · macaroons · OAuth2.
```

---

## What the patterns hand us for open questions

```text
REVOCATION      ocap does it with a CARETAKER / MEMBRANE (an intermediary you can sever) or by
                indirection — so a horizon holds INDIRECT references the kernel can cut. → view-coherence.
LEASED GRANTS   OAuth expiry + service-discovery TTL → grants are time-bounded / refreshable; folds
                registration-staleness and grant-expiry into the one timeout mechanism.
```

## Anchors

```text
ocap               = THE MODEL (adopt its vocabulary explicitly).
macaroons/biscuit  = the GRANT FORMAT (caveat attenuation = Bound).
OAuth2             = the FLOW (consent + issuance + revocation + expiry).
service discovery  = REGISTRATION (lease + health-check = test).
```

## Open / in progress

```text
· the exact grant / macaroon format and what a caveat is in Skyra terms.
· how identity is established — the authentication stub (barrier §7).
· multi-principal delegation (D3) — SPKI/SDSI, UCAN as further references.
· the registration lease + re-test lifecycle — when does a capability go stale.
```
