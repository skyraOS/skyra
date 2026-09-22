# Two directions and the leaf — the producer model

**Date:** 2026-09-02
**Status: BRAINSTORM. Leaf definition + leaf-as-producer (lean: this settles the LITERAL corner toward
Option 2). Does NOT settle args — see args-decision.md.**

---

## The idea (yours)

> There are only 2 directions → forward: select a child. back: return an observation.
> A leaf is one that only does one of these things — backward.
> An LLM can freely return a summary because LLMs are good at creating summaries; it returns it directly.

```text
FORWARD   SELECT a child — an invoke of (Target, Action). carries NO data (see below).
BACK      RETURN an observation — the value flowing UP the fd chain (a Landing).

LEAF  ≝  a possibility/move that orchestrates NO observation dependencies — it only returns backward.
         (an LLM leaf that emits a summary; a driver leaf that reads the world; a const leaf.)
```

This is a crisp, better definition of "leaf" than "empty `Children`": a leaf is defined by BEHAVIOR (returns
directly, waits on nothing), and empty-children is just its structural signature.

## Leaf-as-producer settles the literal corner (Option 2)

If a value can be PRODUCED by a leaf and returned, then a literal/constant does not need a home as a
`Body.Inputs` field (operand-model.md Option 1). It is just a leaf that returns it:

```text
"hello"  is not a baked field — it is a leaf whose return IS "hello".  (an LLM emits it, or a const driver does.)
```

That is operand-model.md **Option 2 (const-leaf children)**, reached from the two-directions axioms. One
mechanism (every operand is a child's return), no new type surface, content-addressing stays fine
(`const("A") ≠ const("B")`). Cost is the known one: more nodes per literal.

## The catch that this does NOT solve: forward carries no data

The two-directions rule is strict: **forward SELECTS a child; it does not PUSH data.** Data flows UP only.
That is internally consistent with the settled model ("inputs are NOT pushed — they arrive as LANDINGS",
move-and-possibility.md). But it has a hard consequence:

```text
A leaf three levels down that needs external datum "X" cannot have "X" pushed to it — the model forbids it.
So "X" must be PRODUCED by one of that leaf's own children (a const/producer leaf).
```

That works for values BORN INSIDE the tree. It does NOT work for values ENTERING the tree from outside — the
REQUEST, or any per-invocation argument. Those have exactly two homes, and that is the args decision:

```text
(a) THINK bakes the datum into the body as a const-leaf   →  body is request-SPECIFIC → no reuse → WARM dies.
(b) the datum is bound at MATERIALIZE time (an ARG)        →  body stays generic → reuse survives → args exist.
```

So: **leaf-as-producer handles values born in the tree; it is silent on external data entering the tree,
because forward carries no data.** That silence is exactly the args question, carried in args-decision.md.
