# Horizon and root

**Date:** 2026-08-21
**Status: WORKING.** The concept settled in conversation on 2026-08-21. The token *structure* in
§4 is explicitly PRELIMINARY and NOT CONFIRMED. Implementation (crypto, wire format) is out of
scope here on purpose — this document is about what the thing IS, not how it is built.

_Instantiates `the-quadrant.md` — this is the **USER column and the elevator** (the pin, the account, the token)._

> **The token is the horizon. And you create a horizon in order to create a root — the two are
> born together, not granted one to the other.**

Reached from both directions independently, which is the reason it is written down: the horizon
was derived as "what a MOVE may touch," and the account capability was derived as "what the pin
unlocks," and they turned out to be the same object.

---

## 1 · The token is the horizon

One structure, and it already appears in three primitives — it was never a new thing to add:

```text
MOVE       the token authorizes the TARGET.        the fabric refuses against it.
THINK      the token filters the NAMESPACE.        you may only propose names it can reach.
REMEMBER   the token names the ARENA.              holding it is what makes the arena nameable.
```

All three are "what you may touch," and "what you may touch" is the horizon. So there is no
separate capability system bolted on beside the horizon — **the horizon was the capability all
along.** BOUND, the primitive that creates a horizon, is therefore the primitive that mints a
token.

---

## 2 · Create a horizon to create a root

A root is what makes REMEMBER available — it names an arena, and a non-root names none. The
account capability that the pin unlocks IS a root token, and a root token IS a horizon. So:

```text
CREATING A ROOT   =   minting its top horizon. the un-attenuated token at the head of the tree.
```

The root and its horizon are not two acts. **The horizon's existence is the root.** You do not
make a root and then grant it authority; making the authority is making the root. Every horizon
below it is a NARROWING of a parent horizon, and the root is simply the one with nothing above it.

This closes the bootstrap. The circle was: every BOUND needs an authority, which APPROACH
establishes, which needs a horizon. It breaks the standard way — **the seed horizon is handed in
from outside the instruction set, by the pin.** The first horizon is the root mint, authorized by
the pin; every narrowing after it needs no one's permission.

---

## 2a · BOUND is two acts — ATTENUATE (local) and DEFINE (privileged)

The earlier resolution over-collapsed this. BOUND does two jobs, and they land on opposite sides
of the boundary:

```text
ATTENUATE   narrow a HELD horizon in-flight, to ride in a MOVE envelope. a local, unprivileged
            append that cannot widen by construction. MUST be local — every MOVE narrows, so if
            narrowing trapped, every MOVE would enter kernel mode and the MMU-in-hardware property
            would be gone. NON-CROSSING, non-negotiable.
DEFINE      create a NAMED, durable, account-level horizon — reusable over and over. this is a
            durable write, and where the written 08-18 "BOUND is a syscall" line was right.
```

DEFINE does not break the load/store contract, because that contract is a MEMORY-tier claim.
A horizon is the AUTHORITY tier (CR3-shaped), a separate privileged axis — real ISAs carry load,
store, AND privileged control-register writes (`MOV CR3`). BOUND-define is that third class.

Two models for DEFINE, **Model A current, both recorded:**

```text
MODEL A   (current) DEFINE is a privileged authority-tier write — a syscall, on-demand, independent
          of arena lifecycle. Horizons are first-class account objects: named, attached at the
          account, reused across arenas. Kernel surface = two MEMORY crossings + one AUTHORITY
          write.
MODEL B   a named horizon is just a durable object — it crystallizes at REMEMBER by contributing
          (like types and contracts) and reloads via ACTIVATE. Kernel surface stays two. Cost: no
          durable horizon mid-arena, only its ephemeral form until the arena commits.
```

See `boundaries.md` SEAM 2 for the same split from the boundary side.

---

## 3 · The pin is identity, not elevation

Carried forward so this document stands on its own:

```text
PIN        establishes WHO YOU ARE — unlocks the account's root token. NOT a privilege mode.
ACCOUNT    the durable identity. owns a namespace and a durable graph. holds/starts many roots.
ROOT       one live arena under an account. REMEMBER commits its cone into the account's graph.
```

The pin does not grant permission to REMEMBER. It hands you the arena to name. With the token,
naming your arena is expressible; without it, still inexpressible — no permission check is added
anywhere. "Use REMEMBER whenever you want" is true while you hold the token, as a held capability,
never as a session flag. A capability travels and attenuates; a session mode does neither.

---

## 4 · Preliminary structure

```text
STATUS OF THIS SECTION: PRELIMINARY — NOT CONFIRMED. A sketch to react to, not a decision.

TOKEN
  root       anchor — which account this descends from. an identifier, not a secret. ties the
             token to that account's namespace and durable graph.
  caveats[]  ordered, append-only. each one NARROWS: arena scope · namespace subtree · expiry ·
             which primitives (may-REMEMBER / read-only) · which targets. never widens.
  chain      binds root + caveats so the list cannot be edited. append is possible; remove and
             forge are not.
```

The one property that is not negotiable, whatever the structure becomes: **narrowing is free and
widening is impossible** — and impossible by construction, not by a rule anyone enforces. That is
the horizon's narrowing-only law and the design's prefer-inexpressible-to-forbidden rule meeting
in one object.

Implementation deferred, recorded in one line so it is not lost: the party that verifies a token
(the fabric, refusing a MOVE) is untrusted and cannot hold a minting secret, which points at
public-key attenuation rather than a shared-secret chain. **Demonstrated since — see §4a.**

---

## 4a · Confirmed shape — a Biscuit-style chain (from the v8 build)

```text
STATUS OF THIS SECTION: a CONCRETE CANDIDATE for §4, demonstrated in throwaway code
(src/skyra/skyra-dna-v8), NOT yet canon. It fills the §4 sketch and confirms the public-key
direction the note above pointed at — it does not overrule §4's "not confirmed."
```

The build realized the token as a Biscuit-shaped Ed25519 chain and ran it end to end, including a
tamper test (edit one caveat → verification fails):

```text
MINT       the ROOT attests the first proof key. held ONLY by the kernel.
ATTENUATE  each caveat ROTATES the proof key and signs the next link. the bearer holds only the
           CURRENT proof, never the root minting key. append-only; cannot widen by construction.
VERIFY     anyone checks the chain from the root PUBLIC key alone (no secret).
```

This is the capability separation the §4 note required — the verifier never holds a minting secret.
Three capabilities, three holders:

```text
kernel     MINT a root horizon (the privileged authority-tier write — §2a DEFINE, Model A).
bearer     ATTENUATE a held horizon in-flight — §2a ATTENUATE, now shown FALLIBLE: (horizon, outcome).
anyone     VERIFY from the root public key.
```

And the payoff: **authority and provenance are the SAME shape** — two Ed25519 chains rooted in public
keys (the horizon chain here; the signed-provenance chain in
`../../2026-08-22/fabric/signed-provenance.md`). Signing a horizon and signing provenance are one
mechanism, not two. This is why the v8 build closed the two together.

What it does NOT settle (unchanged from §4 and Open): the caveat **meaning** stays opaque (the
constraint is an uninterpreted string — the caveat language is deliberately open), and the
**cross-machine distribution** of the root/trust anchor is undesigned (the build used one setup-time
key factory — no distribution, rotation, or revocation).

---

## Open

```text
REVOCATION        a bearer capability cannot be un-issued. the answer that keeps the offline
                  property: tokens are short-lived and the PIN re-mints. "REMEMBER whenever" =
                  your token is fresh, and the pin refreshes it silently. (access + refresh.)

CAVEAT LANGUAGE   what a caveat says, structurally. keep it tiny — scope, expiry, primitive-set,
                  target-set. a rich caveat language is a second place logic can live, and the
                  design wants only one.

DELEGATION        a caveat that another party must discharge (cross-account handoff) is deferred.
                  first-party narrowing is enough for now.

BOUND OWNER       RESOLVED via the DEFINE/ATTENUATE split (§2a), promoted into boundaries.md
                  SEAM 2. ATTENUATE is CTU-local; DEFINE is an authority-tier syscall under the
                  current Model A. What stays open is only the DEFINE model choice (A vs B).
```
