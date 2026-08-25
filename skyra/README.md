# Skyra

A universe simulator that happens to use LLMs as one of its base cases.

It's just a hashmap that calls itself.

Skyra does not make intelligence.

It gives intelligence somewhere to live, someone to learn from, and a history
that compounds.

Most agent systems treat the human as an input source and the model as the
center. Skyra treats the human as a world-shaping participant. Teaching,
correction, trust, conflict, memory, and relationship all become part of the
universe the beings inhabit.

## Reading the docs

Start with **[docs-map.md](docs-map.md)** — the curated working set of
load-bearing docs, kept current as part of landing work. The tree beneath it
splits three ways: `design-docs/` (timeless topics), `specs/` (dated build
sessions — the method's artifact trail), and `archive/` (superseded history,
with READMEs explaining why).

## Process — two lanes, two worktrees

Design and implementation run in **separate git worktrees** so a live build never shares a
tree with active spec iteration. Editing the same tree an implementer is building in races its
work — we hit this once; the worktree split is the fix.

```text
SPEC LANE   the design worktree. edits skyra/specs/. where questions are refined, interfaces
            proposed, decisions recorded. iterate freely here — no build is running.
IMP LANE    a separate worktree. runs ONE work order to completion. the implementer builds here
            and returns its trail. do NOT edit this tree from the spec lane while it runs.
```

Make an imp worktree per build: `git worktree add worktrees/<name> <branch>`.

### SPEC LANE — `skyra/specs/<YYYY-MM-DD>/`

Dated build sessions, by convention:

```text
· a reading-order README that says what changed and in what order to read.
· STATUS headers on every doc  (WORKING PROPOSAL / DECISIONS / OPEN QUESTIONS / RESEARCH).
· decisions.md — settled calls, numbered (D1, D2, …). only firm decisions live here.
· questions.md — open threads. answers graduate into a dated spec of their own.
· grouped by topic (think/ potentials/ research/) and, when a build is targeted, by
  generation (skyra-dna-vN/).
· DOCS WIN over code. if an interface and a spec disagree, the spec is right.
```

### IMP LANE — `skyra/src/skyra/skyra-dna-vN/`

One generation per build, a **FRESH** build — no imports from prior generations; re-derive from
the specs so the seams show. Each generation carries:

```text
· README.md            the central question, the one idea, the package layout, mine-vs-yours.
· overnight.md         THE WORK ORDER handed to the implementer — authority, the target trail
                       (what to prove, in order), hard lines, deliverables.
· interface-rationale  every type → its term → the known seams.
· claude-ideas.md      the SHAPE proposals — attributed, every one REJECTABLE.
· *.go                 interface + struct proposals (declarations; the implementer writes bodies).
```

The implementer returns two deliverables:

```text
· boundary-crossings.md   the narrative — where an interface didn't hold, a concept had no home,
                          a proposal was wrong, which open threads the build forced a call on.
· execution-trace.md      the receipts — the run's flow with real return values at every boundary.
```

### Roles — Claude proposes, human confirms, Codex executes

```text
CLAUDE proposes   poses the question; proposes interfaces + structs; drafts decisions. ALL REJECTABLE —
                  nothing Claude writes is settled on its own.
HUMAN confirms    decides. a proposal becomes settled design only when the human confirms it; a decision
                  (D1, D2, …) is recorded only after confirmation. this is the gate between the lanes.
CODEX executes    takes the CONFIRMED work order in the imp worktree; fills the bodies, drives the proofs,
                  RESHAPES any proposal that doesn't hold — and reports why.
LOOP              read the returned trail → human folds the calls back into the specs → next generation.
```

### The method mirrors the system

On purpose: this process is Skyra's own protocol run on ourselves. **Claude proposes → human confirms
→ Codex executes** is **THINK proposes → the user (the privilege boundary) confirms → the executor
runs**; the worktree split is non-interference; "never collapse a boundary" is D3. It is one recurring
shape — a low-authority proposer handing work to an executor across a **human-confirmed boundary**,
with no proposer holding execute or mint power. Where the process chafes, look for the same seam in the
system; where the system chafes, look here.

## North Star

The user should feel the universe changing after every conversation.

Not just "I talked to an AI," but: Skyra remembered, a relationship shifted, a
being developed, a task moved, trust changed, the map updated. The product wins
when the user can see and feel that their continued participation is growing a
world that is becoming theirs.
