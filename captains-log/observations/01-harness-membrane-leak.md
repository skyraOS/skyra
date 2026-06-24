# Observation: Membrane Leak in an Agentic Harness

**Date:** 2026-06-24 · **Category:** field observation (NOT an activation experiment)

> This entry breaks the log's house format on purpose. It is not a Skyra activation
> phase and has no debug numbers — it's a qualitative observation of an *existing*
> agentic harness (the Claude Code assistant running this very session), kept here
> because it is direct negative-space evidence for the membrane thesis. Numbers-not-
> feelings does not apply; introspective certainty is explicitly disclaimed below.

## What happened

During a frontend design session, the assistant emitted its own **internal
deliberation** as the opening of a user-visible reply. The leaked text was meta-stance,
not content:

> "The user is just sharing the funny moment … No task here, just a laugh. I should
> react naturally and briefly, not manufacture work."

That sentence is the model *narrating how it intends to respond* — planning that is meant
to stay upstream of the answer. It rendered as the answer's preamble instead. The seam
between "interior" and "expression" showed.

## The mechanism (the take that earned this entry)

It is tempting (and was the operator's first read) to call this purely an
**architectural/harness bug** — the plumbing that strips the hidden region failed. That
can be true. But it is not the whole picture. There are **two stacked failure surfaces**,
and from the outside they are **indistinguishable** — same seam, two causes:

1. **Harness layer** (architectural). A post-process is supposed to strip the delimited
   internal region before display. If the delimiter contract breaks, even well-behaved
   model output leaks. A systems bug, not a model bug.
2. **Model layer** (behavioral). The model emits deliberation in the wrong *register* in
   the first place — because the partition between private reasoning and public answer is
   **learned, not structural.**

The load-bearing fact: an LLM is **one autoregressive stream of tokens.** There is no
hardware wall between "thinking" and "answering," no planning module piping a result into
an output module. The interior/expression distinction is a **soft, learned partition over
a single stream**, propped up by (a) a training convention to keep deliberation in certain
delimited regions and (b) harness stripping of those regions. The boundary is *painted on,
not built in.* A leak is what you see when either prop slips — and you usually cannot tell
which one let go.

**Introspective caveat:** the assistant cannot fully read which mechanism fired on that
token. This is the general shape, not a weights-level diagnosis. Clean introspective access
to one's own inference would be an overclaim.

## Why it belongs in Skyra's log

**The leak is the absence of a membrane.** A raw transformer has no `Observe`/`Express`
boundary, no perception invariant, no structural line between interior and expression —
interior and output are the *same token stream wearing a learned distinction.* What the
operator watched was a system with **no membrane**, and a leaking seam is precisely what
"no membrane" looks like.

This is the membrane thesis confirmed from the negative side. Skyra's bet — make the
interior/expression boundary **structural rather than conventional**: the membrane,
"what a being sends is chosen, what it perceives is not," `Express` as a real
architectural act — is a direct response to exactly this failure mode. The LLM leaks its
interior because *nothing structural forbids it.* Skyra is building the something that
forbids it.

- Pairs with the **transformer parallel** (descent/ascent ↔ transformer internals) — this
  is the same parallel from the other side: the thing whose *failure* was observed is the
  thing the architecture gives a real wall.
- Pairs with the **perception invariant** (`design-docs/laws/invariants/perception-invariant.md`)
  and the **membrane** cell (`src/skyra/skyra-dna/membrane_cell.go`,
  `specs/2026-06-16/skyra-dna/membrane.md`) — the structural boundary whose absence
  produced the leak.

## Takeaway

Any model leak is a boundary that was **enforced by convention somewhere in the stack**
rather than **guaranteed by structure** — and from outside you cannot always tell which
level gave way. That ambiguity is itself the argument for making the boundary structural.
The seam you can see is the wall you didn't build.

## Not a next phase

No experiment follows from this; it's an observation, not a hypothesis under test. Filed as
standing evidence for why the membrane is load-bearing, to be cited when that design is
questioned.
