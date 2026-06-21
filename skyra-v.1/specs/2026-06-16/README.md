# Implementation Specs — 2026-06-16

These translate the design docs (`../../design-docs/`) into concrete `.go` changes, **split by code area** — one spec per part of the tree you'd edit in one sitting. Design docs say *what and why*; these say *which files, what change, done when*.

## Specs

| Spec | Code area | Files |
|---|---|---|
| [kernel.md](kernel.md) | the wave engine + the interface | `src/skyra/space.go`, `src/skyra/relation/relation.go` |
| [world.md](world.md) | the environment / facilitator | `src/skyra/world/*.go` |
| [skyra-dna/](skyra-dna/README.md) | the being — **one spec per cell type** | `src/skyra/skyra-dna/**` |
| [inference.md](inference.md) | the generate+align call | `src/inference/*.go` |
| [boot.md](boot.md) | boot, genome, the run loop | `main.go`, `genome.skyra` |

## Build order

```
kernel  →  world  →  (inference ∥ skyra-dna)  →  boot
```

Kernel first (the interface change gates everything). World next (firing routes through it). Inference and skyra-dna can go in parallel once the interface is stable. Boot wires it together last.

## The minimal-first path (the actual MVP)

The captain's-log priority is unchanged: **port a minimal being onto the void kernel and re-run the comparator on a *traversed* frame** (not a hand-built present). The smallest slice across these specs that gets there:

1. **kernel** — remove the actor crossing (the 7-line deletion + its test + comments). The `Content`/`Executables` split and Space-owns-firing are `[after]`; a minimal being traverses in-plane without them.
2. **skyra-dna** — a membrane (sensorimotor; pass-through stub for MVP) + **one** setpoint + **Coherence** (Predict/Resolve). Nothing else.
3. **inference** — the atomic generate+align call.
4. **boot** — place the being, feed one signal, watch Coherence score a real frame.

**Defer past the MVP** (build after the loop runs): the full self/other sub-realities, assumptions/TMS, the error-recovery loop, the World facilitator / Space-owns-firing. Each is marked in its spec as `[MVP]` or `[after]`.

The MVP is the `.1004`-precondition: once the comparator runs on a traversed frame, the whole design becomes testable instead of argued.
