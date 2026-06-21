# binding/_archive

Superseded v.1 binding stubs, kept for history. They import the deleted
`src/reality` package and do not compile — the `_archive/` prefix makes the Go
tool skip this directory (`./...` ignores `_`-prefixed dirs).

Why each was superseded (see `specs/2026-06-16/skyra-dna/`):

- `coherence.go` — became **Global** (`global.go`). Vote tally / tie-break / grouping.
- `attention.go` — folded into the **membrane's two faces** (Observe gate / Express motor).
- `prediction.go` — folded into the predict-score-close cycle (what self/other do).
- `episodic.go` — the binding-reality framing folded into the **self/other agents**.
- `prospective.go` — folded into the setpoint / grouping (future = Global's grouping index).

The live cells are `MemoryCell` (one passive node — setpoint/entity/understanding/
stress/response/action are roles), `GlobalCell`, and `MembraneCell`.
