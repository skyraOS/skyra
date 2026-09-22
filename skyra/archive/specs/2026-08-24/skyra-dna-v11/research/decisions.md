# Decisions — v11

**Date:** 2026-08-24
**Status: DECISIONS.** Settled calls for the v11 line. Only firm decisions live here; open threads stay
in the research docs. (Numbering is local to v11.)

---

## D1 · Starlark is the sandbox leaf-computation language — SETTLED (direction; specifics still research)

We use Starlark for **one** thing: **leaf computation.** Starlark is the **sandbox language that gets
provisional packages to execute at the leaves** — a pure node whose only job is to transform data.

```text
USE     Role 1 only — the COMPUTE leaf. consume resolved register values → produce a derived value.
        provisional (pure, authority-free) packages execute HERE, sandboxed.
DEFER   Role 2 (Starlark as a graph-construction frontend) and graph macros → further research.
        see starlark-frontend-surface.md and starlark-computation-and-dataflow.md.
```

**Why:** it is exactly the language we need at the leaves. Hermetic by identity — *"the external world
is absent"* — so a leaf can transform data but can never cross, wait, or originate authority. That is
what lets us admit provisional **package CODE** (pure modules) without granting authority:
package-availability and effect-authority are **separate admissions**.

**Provisional:** the direction is settled, but more research is needed on the specifics — the value
contract across the register boundary, derived provenance (program digest + input cone), resource rails
(step / memory / output / integer / collection), and module admission. This decision fixes the *role*,
not those details.

**Not settled by this:** whether or how THINK authors graphs in Starlark (the frontend / Role 2) — that
stays research. This decision is the leaf only.
