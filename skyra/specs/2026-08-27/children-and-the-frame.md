# Children and the frame — what the base case sees

**Date:** 2026-08-27
**Status: DECISION.** Extends and CORRECTS `horizon-and-attention.md`. Establishes that the `Context` is
the frame the base-case LLM observes, that its CHILDREN are the visible choices, and that children
therefore live on the `Context` — not only in the CTU's scheduling `Work`.

---

## The principle: Context IS the frame

Skyra runs LLMs (and the human) as the **base case**. When a possibility is invoked to decide, it
**observes its `Context` and acts**. So the `Context` is not just a register file — it is the **frame the
base case sees**. Whatever the LLM must reason over to choose has to be *on the frame*.

## Children are the visible choices

The **children** of a possibility are the **candidate choices in its frame** — the moves / sub-
possibilities the base case selects among. THINK lays them out at planning (this is "attention"). At
execute time the LLM reads them off the `Context` and picks.

## The causality — children come BEFORE work (the correction)

The earlier framing had this backwards: it treated children as something *derived from* `Work[Issued]`
after the fact. It is the reverse.

```text
Context.Children   the visible choices in the frame     ← the base case READS this to decide
      ↓ the LLM chooses
Work[Issued]       the ACTIVATED subset — in flight      (the consequence of the choice)
```

Children are the **precondition** of any work at all; `Work` is the **consequence** of choosing among
them. **You cannot derive children from `Work`** — the cause is not recoverable from the effect. Put
plainly: *you can't have work in flight if there's no visible view of children in the frame.*

If children lived only in the CTU's `Work` (a scheduling structure the LLM never sees), the base case
would stare at an **empty frame** with nothing to choose. So the choices must be on the thing the LLM
observes — the `Context`.

## The layering

```text
HORIZON (root/workflow)   the total authority the workflow may draw on         [horizon-and-attention.md]
Context.CHILDREN          the frame's visible choices at THIS node (THINK-planned)   ⊆ horizon
Work[ISSUED]              the ACTIVATED subset — choices now in flight               ⊆ children
```

Each contains the next. `Work[Issued]` is a *subset of* the frame's children, not their source.

## This refines "Context holds no move state"

The v11 cleanup put ALL move-related state in `Work`. That was too strong. Split it:

```text
SCHEDULING  (Waiting / Ready / Issued)   → Core.Work     execution state, transient, discarded at quiescence
THE FRAME   (children = the choices)      → Context       what the base case SEES — domain state, persisted
VALUES      (registers, result)           → Context       domain state
```

Scheduling stays in `Work`. The **frame the base case reads** — its children — is `Context` state. The
`Attention`-field removal (a `MoveID→Target` duplicate of `Work[Issued]`) still stands; that was the
*active subset*. This is the *candidate set*, which is a different thing and belongs on the frame.

## It is also the possibility-tree down-edge

Children are the **down-edge** of the possibility tree (parent → children). We already had the up-edge
(`Context.ReturnTo` → parent); this was the missing half. That tree is the domain Skyra exists to make
observable (observable causality) — so `Remember` persists it and the human can see it.

**Store, not derive**, for two reasons `Work` and `ReturnTo`-inversion can't cover:
```text
· CAUSALITY   Work is downstream of the choice — it cannot be the source of the choices.
· REMOTE      a child on another host is never resident here; inverting ReturnTo can't find it. the parent
              must hold the edge itself.
```

## Open — settle at encode

```text
· the SHAPE of a child in the frame:
    LIGHT  Children []MoveID     refs; the LLM resolves the candidate templates via the Body
    RICH   Children []Choice     a legible descriptor per option (target, action, contract, what it does)
  LEAN: rich enough that the frame reads as a MENU — self-describing, not handles to go resolve. the call
  decides whether the Context frame is self-contained or a set of pointers into the Body.
· the field on Context (Children []…) — name/type pinned when the shape is chosen.
```
