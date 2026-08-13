# Semantics — 2026-08-12

**Status: THE CARRY-OVER.** What the 08-11 semantic layer says that still stands, ported here with
its status against the 08-12/13 work marked on each document.

> **Nothing in 08-11's semantics was overturned. Five documents carry whole, two carry with their
> naming superseded, and three were left behind for stated reasons.**

---

## The layer

```text
terminology.md            AUTHORITATIVE on vocabulary. Body unchanged; addendum records the
                          renames, the missing successor to `authenticate`, and one word 08-13
                          reintroduced against this document's own ruling.

the-crossing.md           UNAFFECTED.  "The kernel must HOLD the crossing. It still may not
                          READ it."                       ← was interface-revision.md

closure.md                CARRIED, RENAMED. Decided the model `primitives.md` now calls
                          `Remember`. The decision is 08-11's; only the name is new.
                                                          ← was open-until-commit.md

the-collapse.md           REASONING CARRIED, NAMING SUPERSEDED. No control plane; drive is a
                          select; accept is a commit. One structural claim did NOT carry — see
                          below.                          ← was kernel-revision.md

second-kind-of-credit.md  UNAFFECTED. Copies, exposure, and what a copy mints.

type-surface-proposals.md UNAFFECTED, and still the reading order against the code. A/B/C/D.

divergence.md             REWRITTEN for 08-13. The gap is larger and now has a second kind in it.
```

---

## The mapping, in one table

```text
08-11                      08-13                 where
────────────────────────────────────────────────────────────────────────────────────
SELECT                  →  DECIDE                ../primitives.md §3
COMMIT                  →  REMEMBER              §5 — plus Accept, plus closure
discover                →  THINK                 §2 — recall is the lookup inside it
introspect              →  ORIENT                §7 — carried verbatim
provision               →  absorbed by DECIDE    §3, the ACCESS binding
INGRESS / MOUNT         →  APPROACH              §1 — and approach has an opposite
authenticate            →  (nothing)             ← THE HOLE. see terminology addendum
────────────────────────────────────────────────────────────────────────────────────
                           REACH                 §4 — new, no 08-11 equivalent
                           RECOGNIZE             §6 — new
```

---

## Three things to know before reading any of it

**1 · `Remember` is not new.** `closure.md` decided it on 08-11 — *"closure is the only durable
write"* — and `primitives.md` renamed it and folded in `Accept`. A critique written on 08-12 called
it a change to the design; it is a change to the **code**, which still writes at every commit.

**2 · `authenticate` is the missing gate.** `terminology.md` specified it on 08-11 —
*reads the WORLD, WHO IS THIS* — `primitives.md` dropped it, and `2026-08-09/questions.md` §5 says
nothing gates entry. Those are one hole with two descriptions.

**3 · Forms versus names is unreconciled.** `the-collapse.md` insists `SELECT` and `COMMIT` are the
grammar and `discover`/`introspect` are *engrams reached by it* — *"the names are not a surface at
all."* `primitives.md` lists seven peers. If the older reading holds, `THINK` and `ORIENT` are not
primitives beside `DECIDE`, and the seven are two things wearing one list. This is
`questions.md` §16 one layer down.

---

## Left behind, and why

```text
roots-and-admission.md          ALREADY SUPERSEDED by terminology.md, by its own header, on
                                2026-08-11. Its surviving content is in terminology.md.

questions-08-11-claude.md       a conversation artifact — questions raised in one session and
                                answered in it. Open items belong in 2026-08-09/questions.md,
                                which is where §16 went.

conversation-runtime-mismatch.md  resolved by its own closing line: "the conversation was right in
                                every row." Nothing left to carry.
```

Left in `2026-08-11/`, not deleted. A document that has been superseded is still the record of how
the thing was decided.

---

## What is not here

The 08-12/13 work that is **not** semantics, and lives one level up:

```text
../primitives.md            the seven
../reach-abstraction.md     what the OS gives reach, and what it drops
../distance.md              Orient × Endpoint × Experience
../os-parallels.md          17 parallels, marked verified or proposed
../syscall-read-example/    the walkthrough, the evidence, and the decisions
```
