# Relation Kernel

This package implements the live spacetime wave kernel. A relation realizes an impulse by
walking a space's child graph, applying edge-scoped activation, and preserving the wave's
origin and amplitude.

The kernel owns traversal mechanics and short bookkeeping locks. It does not grade claims,
interpret route data, choose actions, or know whether a space represents a grid, a person,
or a remote device.

Integration and concurrency behavior are exercised in `tests/`.
