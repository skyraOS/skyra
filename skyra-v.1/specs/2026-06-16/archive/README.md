# Archived — 2026-07-12

Archived after the cause-and-effect V0 landed and the v1 `skyra-dna` cell
package was collapsed into `dnav2`. Kept as history; do not build from these.

| Doc | Why archived |
|---|---|
| [agent.md](agent.md) | Specs `SelfCell`/`OtherCell`, deleted with the v1 package. The predict/grade machinery it describes survives in dnav2 only as explicit migration debt (`outstanding-reform-tracks.md` §3) and is expected to be re-derived through the causal graph in the Praxis spec (a self-prediction is plausibly a Final belief). |
| [action.md](action.md) | `[after]` skeleton for the action store. That ground now belongs to the Praxis spec — actions as `CausalNode`s with birth-stamped Final causes — and a stale pre-causal shape here invites building the wrong thing. |
| [boot.md](boot.md) | Unfilled skeleton. Boot is `WireGenes` + `NewDNA` in dnav2. |
| [inference.md](inference.md) | Unfilled skeleton. Inference is the injected `InferFunc` seam. |
| [world.md](world.md) | Unfilled skeleton. The world's living docs are `design-docs/world/`. |

The rest of `specs/2026-06-16/` remains live: the setpoint/entity/coherence/
membrane/midlevel/sensorimotor design ground carried into dnav2 conceptually
and is still cited by running code, even though the cells it was first written
against are gone.
