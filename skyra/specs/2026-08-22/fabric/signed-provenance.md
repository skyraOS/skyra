# The cut on the wire — a MOVE, and signed provenance

**Date:** 2026-08-22
**Status: WORKING.** What crosses when a possibility is partitioned (`../ctu/possibility.md`). The
fabric owns transport (`2026-08-18/network/traffic-routing.md`); this doc says what a cut edge is and
how provenance survives it unforgeably. This is the concrete form of the v7 punch-list's P1.2
("provenance must survive the fabric, unforgeably").

> **A cut between partitions is a MOVE. It carries an intermediate observation from one partition to
> the next. Each CTU signs the provenance segment it executed, so a downstream CTU can RECOGNIZE a
> stamp it did not make without trusting the sender's word — the signature is the trust, not the claim.**

---

## The cut edge is a MOVE

```text
IDENTITY   (possibility identity, edge). the v7 build's changes.md calls this the MoveID; the two
           legs of the edge are its request and response directions.
CARRIES    an intermediate observation, in transit. the envelope carries the cut-edge identity and
           header for routing; the payload stays opaque (the fabric never reads it).
NOT ON IT  arenaID and PossibilityID do NOT cross — the response is routed by (Endpoint, edge), and
           the owning possibility is resolved locally at home (../ctu/possibility.md).
```

## Signed segments

Each CTU **signs the provenance segment it executed.** A slice accretes a chain of signed segments as
it crosses CTUs. Unforgeability across a trust boundary is then a signature, not an assertion:

```text
STAMP        an entity yields raw Content; its CTU produces the Observation (provenance) — never the
             entity (2026-08-18/entity-ctu-boundary/decode-and-run.md: content vs causality).
SIGN         the producing CTU signs its segment.
RECOGNIZE    a downstream CTU verifies the signature; it need not trust the sender's word.
BOUND        the horizon bounds what a hosting CTU may do to a slice merely passing through it
             (attenuated to the slice — ../ctu/possibility.md, "least-knowledge").
```

## The dual chain — authority and provenance, one shape (confirmed by the v8 build)

The v8 build (`src/skyra/skyra-dna-v8`, throwaway) demonstrated that the signed segment here and the
signed horizon of `2026-08-18/kernel-ctu-boundary/horizon-and-root.md` §4a are the **same
construction** — two Ed25519 chains rooted in public verification keys:

```text
HORIZON      root attestation → append-only caveat / proof-key links.   authority — what you may touch.
PROVENANCE   CTU identity      → signed content-addressed Segment.        custody — who produced this.
```

Same shape, different chain — and the same capability separation: the signing secret stays with its
holder (the kernel mints horizons; each CTU signs only its OWN provenance), while transport and the
gather hold **public verifiers only** — they can check a stamp, never forge one. Tamper either chain
by one field and verification refuses it (tested). This is P1.2 closed **in shape**; what remains is
distribution, below.

## Provenance is also the recovery plan

The signed `Causes` references are a **lineage**: a slice lost on a dead CTU is not a loss but a
**re-execution** from its recorded inputs. The chain that makes the cone gatherable at REMEMBER
(`../kernel/remember-gather.md`) is the chain that makes a lost partition recomputable. Provenance =
custody proof = recovery plan.

---

## Open

```text
WHERE THE SEGMENT TRAVELS   the v7 build carries the stamped observation beside the opaque MOVE
                            locally; the on-the-wire shape of the signed segment across an external
                            trust boundary is unresolved (changes/boundary-crossings.md,
                            "Observation transport has no settled layer").
THE EDGE CONTRACT           the join checks the arriving observation against a contract at the
                            correlated position. this needs the set's still-undefined CONTRACT object
                            (2026-08-18/README.md, "four things most likely to change"). the typed cut
                            is the next real interface.
TRUST-ROOT DISTRIBUTION     the SHAPE is settled (dual Ed25519 chains; public verifiers in the fabric
                            — see "The dual chain"). what stays open: how a verifier on one MACHINE
                            obtains another CTU's / root's public key. the build used ONE setup-time
                            key factory, so cross-machine key distribution, rotation, and revocation
                            are undesigned. the inter-CTU-trust core of the-quadrant.
```
