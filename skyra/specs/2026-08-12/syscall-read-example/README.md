# syscall-read-example

**Status: WORKING.** A single worked example, traced against the code, used to find where the
entity/host seam actually falls.

**Coherence position:** this directory preserves the evidence and Michael's decisions which led to
the newer model. Its direct syscall and peer-method language is historical where it conflicts with
[`../continuation-language/primitives.md`](../continuation-language/primitives.md),
[`../continuation-language/think-interface-proposal.md`](../continuation-language/think-interface-proposal.md),
or [`../continuation-language/primitives-and-continuation-summary.md`](../continuation-language/primitives-and-continuation-summary.md).
Nothing in `outcomes/` is rewritten by this coherence pass.

**Everything here is asked from the ENGRAM's perspective** — not the host's, not the model's.

```text
example.md    THE WALKTHROUGH. one ask, traced step by step until the information runs out.
              poses the problem. draws no conclusions.

evidence/     OBSERVATION, one document per question. commands and their output.
              no conclusions either.

outcomes/     DECISIONS, one document per question. WRITTEN BY MICHAEL ONLY.
```

```text
WHERE   "Where am I stored?" and   the graph that holds me · the host my entity is
        "Where is my entity?"      reachable from. TWO questions.
WHO     "Who is interpreting me?"  the entity, and how it is reached
WHAT    "What am I?"               my expectation, and the observations bound to me
HOW     "How am I executed?"       spawn, the program, and the interface available where I land
WHY     "Why was I invoked?"       the parent, and the citation
WHEN    "When am I executed?"      my tick, and what I can be sequenced against
```

Each question has a **pair**: `evidence/<Q>-evidence.md` measures, `outcomes/<Q>-outcomes.md`
decides, and the outcome cites the evidence rather than restating it.

**Citations are scoped to their document.** Sections are numbered `E1`, `E2`, … within each
evidence file and cited as **Where E2**, **How E1**, so numbers never collide across questions.

> **Rule for this directory:** nothing in `outcomes/` is mine to fill in. Proposals go in the chat
> log and land there only when Michael puts them there. The example and the evidence may be
> extended; the outcomes may not be inferred from them by anyone else.

**Evidence and decisions are kept apart on purpose.** A measurement is true whoever reads it; a
decision is a choice someone made. Mixing them makes the second look like the first.

The example stops where the information stops. Every place the trace cannot continue is a place the
design has no representation for, and naming those is what the example is for. Deciding what to do
about them is not.
