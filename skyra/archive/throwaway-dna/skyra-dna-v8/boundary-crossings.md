# Overnight v8 build — boundary crossings

**Date:** 2026-08-23
**Status: THROWAWAY IMPLEMENTATION NOTES.** The trail from making one possibility cross two CTUs
and making REMEMBER gather its signed cone home. Design authority remains
`skyra/specs/2026-08-22/` on `skyra/specs/2026-08-18/`.

## What runs

```text
root possibility @ client CTU
  → SCHEDULE chooses SendTo
  → BOUND attenuates the signed Horizon to the cut
  → MOVE(request, Slice) → GREET @ worker CTU
  → Slice becomes a non-root child (Parent = incoming MoveID; Arena = local scope)
  → entity Run → Content
  → worker CTU stamps + signs Segment
  → MOVE(response, Observation{Segment})
  → client RECOGNIZE verifies, resolves MoveID locally, folds Frontier
  → root REMEMBER starts at the kept Segment
  → Fetch each content-addressed Cause, verify + dedup
  → driver atomically persists the gathered Cone
```

## Interfaces that did not hold

### A live possibility cannot cross `Execute` by value

`Execute(Possibility)` could not leave the returned observation or changed frontier on the object
that owns that state. It is now `Execute(*Possibility)`. `Greet` correspondingly returns a pointer.
The CTU retains only a `MoveID → owning *Possibility` routing index; the computation state stays on
the possibility.

### `Parent *Possibility` cannot represent a parent across a cut

The parent possibility deliberately does not travel with a slice. A receiver therefore cannot
construct the proposed pointer without either copying forbidden state or inventing a fake parent.
`Parent` is now `*MoveID`: nil is structurally root; a remote child's incoming cut is its immediate
one-hop parent link. This also keeps Arena and PossibilityID off the response.

### The possibility carried no executable work

The proposed `Possibility` held identity, arena, frame, and frontier but no instruction, horizon, or
payload. `Execute(possibility)` therefore had nothing to execute. `Name`, `Horizon`, and `Payload`
now live on the possibility and are copied to/from the forwarded Slice at GREET.

### Signed provenance had no route to RECOGNIZE or REMEMBER

`Observation` carried an unsigned `Provenance`, while the signed `Segment` was not reachable from a
MOVE response or the kept observation. Both required consumers therefore had nothing verifiable.
`Observation` now carries `Segment` directly beside Content. The entity still returns only Content;
the CTU alone constructs and signs the Segment.

### `Custody.Sign` could not produce a complete segment

The original call returned only `Signature`, even though signing also establishes the Segment's
content-address. It now returns the completed `Segment` plus `Outcome`. This prevents callers from
pairing a valid signature with an independently invented reference.

### Cryptographic attenuation is fallible

The proposed `Bound(Horizon, Caveat) Horizon` had no way to report an invalid chain or failed proof
rotation. It now returns `(Horizon, Outcome)`. The build uses a Biscuit-shaped approximation: the
root authority attests the first proof key; every caveat rotates that key and signs the next link;
the bearer retains only the current attenuation proof, never the root minting key.

### APPROACH had lost the response address

`Fabric.Approach(Move)` did not tell normalized ingress which endpoint sent a request, so GREET
could not form the response leg without placing a home identity in the Slice. Restoring
`Approach(from, Move)` keeps the return endpoint in transport metadata and keeps Arena off the wire.

### ORIENT needs the fragment it reads

`Orient()` was ambiguous on a CTU executing more than one possibility. It is now
`Orient(*Possibility) Frontier`: explicitly local to that fragment. Whole-possibility orientation
remains a propagated cut, not a global machine view.

## Authority and provenance: the same shape, different chains

Both are Ed25519 chains rooted in public verification keys:

```text
Horizon      root attestation → append-only caveat/proof-key links
Provenance   CTU identity → signed content-addressed Segment
```

Tampering tests change one caveat and one provenance field; both are refused. This fixes v7's
unsigned `{Root,Caveats}` stub without giving the fabric a minting secret.

## New implementation names

These name mechanisms, not new machine concepts:

```text
Security        setup-only key factory that issues narrower component capabilities.
Local           the local-ingress realization already named by traffic-routing.md.
File            the concrete durable-storage driver.
RemoteResource  an eligible resource whose scheduler selection is SendTo.
Seed            throwaway setup that installs already-known Engrams in the File driver.
owners          the CTU's local MoveID routing index; it points into Possibility-owned state.
```

## Deliberately unresolved

- Caveat **unforgeability** runs; caveat **meaning** does not. `Constraint` is still an opaque string,
  because the spec explicitly leaves the language open. The fabric verifies the narrowing chain but
  does not pretend to know target/expiry/primitive semantics.
- `Security` is one setup-only in-memory key factory. Components receive narrower views: the kernel
  gets the root minter; each CTU gets only its own provenance signer; transport/gather get public
  verifiers. This exercises capability separation and signatures, not key distribution, rotation,
  revocation, or the inter-CTU trust-root design the spec leaves open.
- The cargo sum proposal was not adopted. `Move` retains Slice/Payload/Returned fields and rejects
  impossible combinations at runtime. A Go sum would make them unrepresentable, but costs more
  scaffolding than this boundary run needs.
- PUSH is deterministic when no local entity exists; PULL policy, remote liveness, resource pressure,
  retries, acknowledgements, external ingress, and consistent-cut termination remain open.
- The gathered Cone persists signed segments. As in v7, the contribution surface still does not carry
  a new compiled Engram, so installing a new name → meaning binding remains undecomposed.
- No process type, table, lifecycle, or host-process semantic was introduced.
