# 2026-07-22 — Superseded by `2026-07-24`

**The live design set moved to [`../2026-07-24/`](../2026-07-24/README.md) on 2026-07-24.**

This directory is the historical snapshot of where the v4 design stood on 07-22 through
07-23, including the reasoning that was later corrected. Read it for *how the thinking went*;
build from `2026-07-24`.

Thirteen docs were carried forward. Two were deliberately left here:

| doc | why it stayed |
|---|---|
| `traversal-spine.md` | **RETIRED.** It describes a field model that no longer exists — *entities → triples → spans* — written before `tokenizer.md` established the concrete/segmented representation. Its layer names (`LAYER 0/1/2`) collide with the live ones and will mislead. Its banner records where each surviving piece was folded. |
| `defects-found-in-review.md` | v3 correctness work order, **all in-scope items closed 2026-07-22**. Section 4 remains deliberately out of scope. Still cited from the live set. |

**Do not edit the copies here.** Divergence between two copies of the same doc is the exact
failure this project has already documented twice (`Resolve` and `episode` each naming two
things). Corrections belong in `2026-07-24`.
