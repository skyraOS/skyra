# Syscalls → World Affordances

## Status: Superseded

This document captured the intermediate step — primitives moved from the being to the kernel. That was closer but not right. Primitives are the world's affordances, not the kernel's methods.

See: [world-dna.md](../world/world-dna.md)

## What Survived

- **Primitives don't live on the being.** Still true. The being can't touch the world directly.
- **The kernel mediates.** Still true. The kernel carries requests and traverses.
- **Permission levels exist.** Still true. Focus/move/engage are unprivileged. Exec/kill/mount are user-only.
- **Speaking is not a primitive.** Still true. It's the being's motor express. The kernel propagates it.

## What Changed

- Primitives are not kernel methods. They are the world's competence functions — affordances the world offers.
- The kernel doesn't dispatch syscalls. It traverses the world the same way it traverses a being — DFS, score, filter. The world's DNA defines what happens.
- Different affordances have different activation sequences. Focus spreads wide. Move is direct. Engage is targeted. The kernel doesn't know the difference — the world cells do.
- The kernel is dumb. It just traverses. The cell defines what happens. Same pattern everywhere.
