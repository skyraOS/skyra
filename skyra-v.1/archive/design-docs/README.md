# Archived design docs — 2026-07-12

Superseded design docs moved here after the cause-and-effect V0 (phases 0–7)
landed and the v1 `skyra-dna` cell package collapsed into `dnav2`. Kept as
history; the reasoning in them often still holds — the mechanisms don't.

| Doc | Why archived |
|---|---|
| [dna.md](dna.md) | Self-declared out of date since the v.1 kernel rewrite (stateless kernel, Perception Invariant, ObserveFn split). Its load-bearing phrase — "one type, topology differentiates" — lives on in `dnav2/memory.go`. |
| [cell-types.md](cell-types.md) | "A cell's type is fixed at birth" described the v1 cell taxonomy (`SelfCell`, `OtherCell`, …), deleted 2026-07-12. dnav2 is one `MemoryCell` plus competence namespaces; Kind is a role tag, not a behavior switch. |
| [operator-runtime-and-interrupts.md](operator-runtime-and-interrupts.md) | Operator-era runtime model; replaced by dnav2 genes, the board, and the workflow. |
| [prediction-motor-and-closure.md](prediction-motor-and-closure.md) | Extended the operator runtime. Its closure concepts were re-derived properly in `specs/2026-07-08/cause-and-effect/` (grading, Final orientation, retain-or-requery). |
| [a3-world-observe-and-result-stub.md](a3-world-observe-and-result-stub.md) | Self-declared "Superseded — done"; kept as the breadcrumb it says it is. |

Completed kernel implementation plans (`kernel-change-plan.md`,
`stateless-kernel-slice-refactor.md`) went to
[`../improving-v.1/completed-specs/`](../improving-v.1/completed-specs/),
matching the existing convention for finished plans.
