# Skyra Source

This tree separates cognition, worlds, and the substrate that connects them.

```text
cognitive-agent/     cognition and learned behavior
distributed-agent/   topology, live environmental state, and affordances
spacetime/           neutral space, time, impulse, and traversal contracts
user-agent/          the human-facing agent boundary
```

The architectural boundary is deliberate: an agent carries cognition; a world owns its
topology and live state; spacetime moves impulses without interpreting their meaning. A
device-backed place should implement the same world-facing contracts as a local world, with
transport and platform details kept below that boundary.

See the README in each directory for its narrower responsibility.

## Implementation probes

An overnight build tests whether the declared interfaces can carry a working system. Implement
behavior within that surface and leave a short trail of the seams the build exposes.

- Do not add structs or fields, or change existing type shapes and method signatures.
- Build the main loop from methods specified in the interfaces. Keep component state behind its
  owning methods; do not bypass the interface through direct field access from the loop.
- Do not hide missing architectural state in globals, side tables, or unrelated fields.
- Implementation choices inside those boundaries are yours. If a boundary prevents progress,
  report the missing capability and continue the work that fits; propose surface changes separately.

The current [CTU engine work order](skyra-ctu-engine/overnight.md) applies these constraints to a
minimal local loop. Older overnight orders that permit adding runtime structures do not override
these rules for this probe.
