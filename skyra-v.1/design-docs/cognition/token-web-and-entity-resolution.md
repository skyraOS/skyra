# Token Web and Entity Resolution

> **⚠️ Edge direction superseded (2026-07-07).** This doc's `entity → interaction → token`
> (entity owns its interactions) is **flipped**: interactions are **gist-owned** —
> `grouping → gist → record → entity`, and the record points to tokens (`record → token`). The
> entity no longer owns interactions; it keeps only its *content* children (`entity → location`,
> `entity → understanding`). "Interaction record" is now an **`EngageRecord`** (a Kind of
> `AffordanceRecord`). Everything below the edge direction — the resolution pipeline, aliases as
> **grouping/gist children**, confidence/TMS, the three phases — **still stands**. Unbuilt in
> dnav2 (lives only in archived old dna) → **port fresh in the new direction.** See
> [perceive-internalization.md](perceive-internalization.md) §3/§6 and
> [affordance-record.md](../being/affordance-record.md).

**Date:** 2026-07-01

## The idea

Every impulse — perceived or emitted — gets tokenized on arrival. Each token
becomes a node in the graph. The impulse itself sits in a web of those nodes.
Traversal through the accumulated token web IS recall: no separate retrieval step,
no index scan. The kernel DFS that does everything else also does memory.

This is the persistent, explicit version of what an LLM's attention mechanism does
temporarily and implicitly at each forward pass. The LLM rebuilds the web on every
call. Skyra's web accumulates across turns, sessions, and time.

## Tokenization

On every sensory intake: split the payload on whitespace, strip punctuation,
lowercase. Each resulting word is looked up in a flat token index (`ByToken
map[string]*MemoryCell` on sensory). If found, reuse the existing node. If not,
create one (`KindToken`). The impulse record (already a child of the entity node)
gets the resolved token nodes as its own children.

```
michael (entity)
  └── interaction:1  "hello skyra"
        ├── "hello"  (KindToken, shared node)
        └── "skyra"  (KindToken, shared node)
  └── interaction:2  "help me call my mom"
        ├── "help"   (KindToken)
        ├── "me"     (KindToken)
        ├── "call"   (KindToken)
        ├── "my"     (KindToken)
        └── "mom"    (KindToken → resolved by MidLevel to "sarah", confidence 0.5)
```

Token nodes are **shared** across impulses: the second time "skyra" appears it
connects to the same node. The web forms through repeated co-occurrence. Two tokens
that always appear together accumulate short paths between them; two tokens that
never do drift apart.

Stop word filtering is deferred — the graph will self-organize (filler words
accumulate many weak edges and eventually score low under activation). Explicit
filtering can be added later without changing the architecture.

## Entity resolution

The hard problem: "mom", "relation", "runtime", and "the kernel" are all tokens
that refer to existing entities — but exact string match fails them.

### The resolution pipeline

Every token passes through this sequence. There is no "unresolved" outcome —
every token ends up placed. Confidence is the signal for uncertainty, not
a binary resolved/failed state.

```
1. String lookup (sensory)
      ↓ hit → use existing node, confidence 1.0
      ↓ miss
2. LLM resolver frame (MidLevel)
      — forced to return an answer —
      "In this context [exchange + orientation + known entities],
       what does X most likely refer to?"
      returns a canonical name (e.g. "Sarah", "kernel", "michael's mother")
      ↓
3. String lookup again on the returned name
      ↓ hit → use existing node
      ↓ miss → create new node for the returned name
4. Create alias entity: From=original token, To=resolved node,
   GroupingID=active grouping, Confidence=low (0.5 default)
   Write to durable storage.
```

The LLM resolver is a translation layer only — it suggests a canonical name,
then deterministic machinery handles the rest. The LLM never directly creates
or modifies nodes. Steps 1 and 3 are identical lookups; step 3 just operates
on the LLM's output rather than the raw token.

**Example — "mom":**
1. `ByToken["mom"]` → miss
2. LLM: "In this exchange michael is talking about calling his mother. The most
   likely referent is 'Sarah'."
3. `ByToken["sarah"]` → miss → create new node "sarah"
4. Alias: `mom → sarah`, confidence 0.5, written to durable storage

Next time "sarah" appears in a michael exchange, step 1 hits directly.
Next time "mom" appears under the same grouping, the alias lookup (before
step 1 even fires) resolves it to "sarah" at full alias confidence.

**Confidence and the truth-maintenance loop:**

MidLevel is forced to resolve — it always places every token. Low confidence
means provisional, not wrong. The alias entity exists in the graph and
accumulates activation every time it's traversed. If the resolution keeps
being consistent with what actually happens, confidence can rise. If it's
contradicted, a new alias is established with higher confidence and the old
one decays below threshold.

Self reads the grouping's alias children alongside prediction nodes. A
low-confidence alias is visible to self — it can raise a clarifying question
if the uncertainty materially affects the move ("I want to make sure: when
you say 'mom', are you referring to Sarah?"). Self surfaces the uncertainty;
it does not resolve it directly.

### The three-phase arc

**Phase 1 — string lookup only (now).** Sensory does steps 1 and creates a
new node on miss. No LLM call. The web builds, resolution is naive.

**Phase 2 — LLM resolver active.** MidLevel runs the full pipeline above for
any token that missed step 1. Aliases accumulate. The graph starts encoding
context-scoped resolution.

**Phase 3 — co-occurrence as signal.** Tokens that always appear together
accumulate short paths between them. The graph topology supplements the alias
layer — even without an explicit alias, traversal from "relation" reaches
"kernel" through shared neighbors. The two mechanisms (aliases + topology)
reinforce each other.

## Alias entities

An alias is a **first-class node** (`KindAlias`), not a bare edge. Bare edges carry
no metadata. An alias entity carries: what surface token it resolves, what canonical
entity it points to, which grouping scoped it, when it was established, and with
what confidence.

```
AliasRecord {
    From      string   // surface token ("relation")
    To        string   // canonical entity ID ("kernel")
    GroupingID string  // the orientation that scoped this alias
    Confidence float64
    Established int    // slice timestamp
}
```

Alias entities are **children of the grouping** that established them — not children
of the token or the canonical entity. The grouping IS the scope.

```
grouping (gist: "skyra runtime development")
  ├── other-pred-root
  ├── self-pred
  ├── alias: "relation" → kernel   (KindAlias)
  ├── alias: "time"     → kernel   (KindAlias)
  └── alias: "runtime"  → kernel   (KindAlias)
```

Resolution during `ensureToken`: before creating a new token node, check the active
grouping's children for an alias where `From == word`. If found, return the
canonical entity node directly. The kernel DFS that traverses the grouping surfaces
these aliases; their edge activation scores reflect how recently they've been used.

**Traversing in from a different grouping** and finding an alias is a good problem
to have, not one to prevent. The alias edge will score lower under a foreign
grouping's activation (its edge was grooved by a different context), which is a
signal the resolution is context-borrowed. Display it differently, trust it less.
The graph expresses the difference through topology.

## Durable storage

Alias entities are written to durable storage when established. They are learned
facts about the world — they survive restarts. The in-memory graph is the live
session state; the alias store is permanent.

Format: a flat file (`aliases.json` at the project root), one record per alias. At
boot: read the file, find or create the referenced groupings, add alias entities as
children. When the LLM resolver or a human establishes a new alias: write to the
graph AND append to the file immediately.

## The gradient

This design creates a natural altitude hierarchy:

```
Sensory    activates nouns (entity nodes — who/what is here)
MidLevel   recognizes orientation (which graph region is lit)
Other      reasons over the active token web (world state)
Self       predicts a move given that world state
Motor      realizes the move as surface content
```

MidLevel's grouping is the resolution scope for everything below it. The orientation
determines which aliases are live. Entities at sensory are the traversal roots —
the kernel DFS fans out from them through the token web, surfacing what's relevant.

## What this resolves

- The entity resolution open question from `entities-and-surroundings.md` — a
  concrete three-phase path with alias entities as the resolution mechanism.
- How memory accumulates without a separate retrieval system — tokens are nodes,
  the web is the memory, traversal is recall.
- How context scopes resolution — through the grouping, which is already the
  orientation anchor.

## Open

- **LLM resolver runs every pass** — currently the resolver only fires on a
  string-lookup miss. The right behaviour is to run it on every MidLevel pass:
  existing low-confidence aliases get re-examined as context accumulates, and
  confidence can rise or a better resolution can replace the current one. The
  pipeline is the same; the trigger condition changes from "miss" to "always."

- Stop word filtering — currently deferred to self-organization.
- Embedding-based resolution for Phase 2 → Phase 3 bridge (nomic-embed-text on
  the mac mini is the right tool).
- Merging two nodes that are discovered to be the same entity (alias entity
  establishes the connection; full merge is a separate operation).
- Token nodes for skyra's own emissions (motor output) — same mechanism, not yet
  wired.

## Source

- `design-docs/cognition/entities-and-surroundings.md` — the entity graph this
  builds on; entity resolution flagged as open there.
- `design-docs/cognition/groupings-and-understandings.md` — groupings as the
  orientation anchor.
- `design-docs/kernel/activation-and-traversal.md` — the traversal mechanism that
  makes "recall = traversal" work.
