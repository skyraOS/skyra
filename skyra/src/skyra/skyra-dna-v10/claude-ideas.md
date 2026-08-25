# claude-ideas — v10 shape proposals

**Date:** 2026-08-24
Every SHAPE decision below is mine (Claude's), not the docs'. The terms are the docs'; the
*types, returns, and package split* are proposals. All are **rejectable** — reshape any and
say why in `boundary-crossings.md`. Ordered most-certain → least, with the three I'd check
first flagged **[CHECK FIRST]**.

```text
1  acquire IS A POSSIBILITY, not a set of Go calls.  [CHECK FIRST]
   the doc says discover/test/admit are children of THINK-the-possibility. i modelled the
   Acquire/Discover/Test/Admit INTERFACES for the reader, but the real thing is a Move graph
   the engine drives. seam: do the interfaces survive, or do they collapse into "assemble a
   Possibility whose moves target entities"? if the latter, the interfaces are scaffolding.

2  FAULT is a fourth Outcome value.  [CHECK FIRST]
   a name that resolves to nothing ADMITTED returns Fault, distinct from NotYet. but Fault
   carries no payload — the unadmitted Endpoint rides on the Move. seam: does the privilege
   handler have enough from (Possibility, Move) to know WHAT to grant, or does Fault need to
   be a typed value carrying the target + the grant kind it needs?

3  Horizon.AskPermission is a bool.  [CHECK FIRST]
   scarcity (D2) as one switch: default false, true only on privilege-bearing levels. seam:
   a bool says "may originate authority" but not "over WHICH resources." if a level may ask
   for X but not Y, this needs to name a set, not a flag. i bet the bool breaks the first
   time a delegate holds a scoped grant.

4  the mint lives on the KERNEL, reached by TRAP; privilege only DECIDES.
   corrected mid-design: admitting a capability is authority origination = kernel domain
   (v9's MINT). so kernel.Mint is the mechanism, privilege.AskPermission is the principal that
   decides (novelty + user policy) and traps to it. acquire holds neither. SETTLED (D3): the
   principal/mechanism split STAYS — always, single-user included; a collapsed boundary reads
   as no boundary, and it is real. STILL OPEN: is "acquire holds no kernel reference" enough,
   or must the engine refuse a Mint whose trap did not come through privilege?

5  a PARKED GRANT reuses the cut machinery (Cut / InFlight / Recognize).
   a grant that waits on a human is a Cut with a slow response leg — no new park path. seam:
   a grant and a remote-response are both "parked things"; do they want one InFlight index or
   two? and what re-drives when a human never answers (timeout ⇒ Failure, like connect)?

6  ConnectionContract.Partial + Auth.Verified=false model the deferred rung.
   test recovers auth as "required, unverified-until-exec." seam: is a bool enough, or does
   each rung need its own present/deferred/failed state?

7  novelty is Registry.Known (a read), returning NoveltyVerdict.
   endpoint-novelty and type-novelty are one gate; History is an opaque placeholder here.
   seam: where does History actually live — the Registry, or a separate REMEMBER store?

8  two grant kinds (PokeGrant, RegisterGrant).
   poke before test, register at admit — the user sees the contract before committing. seam:
   in the fast (known) path do they collapse to zero grants? and is poke ever skippable?

9  the four probe leaves are separate interface methods, not one parameterized probe.
   reach/handshake/challenge/describe each write a distinct fragment. seam: the "one act"
   (provoke-and-read) is implicit; if per-protocol dispatch dominates, one Probe(rung, proto)
   may be truer than four methods.

10 FitnessMatch is above the leaves, not a leaf.
   "does an operation serve the purpose" is a judgment; "no fit" is the diagnosis that
   refines search. seam: does the refine edge (search↔test loop) need to exist in v10 at all,
   or is one pass enough to answer the question?
```

## The one I most expect to be wrong

**#3 (the AskPermission bool).** Scarcity as a single flag is the cheapest thing that could
work, and the first scoped delegation or system-administered grant will probably force it to
name resources. If it holds through the run, that is itself a finding.
