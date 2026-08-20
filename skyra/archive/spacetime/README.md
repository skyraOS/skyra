# Spacetime

Spacetime is the neutral substrate shared by agents and worlds. `space.go` defines the core
contracts for spaces, slices, time, impulses, coordinates, activation, and traversal.

The substrate may route an impulse, preserve its origin, and maintain traversal state. It
must not interpret cognitive payloads, decide what an observation means, or contain
device-specific policy.

Subpackages:

- `relation/` — the live wave kernel over `Space` and `Slice`;
- `traverse/` — a smaller traversal prototype with its own tests.
