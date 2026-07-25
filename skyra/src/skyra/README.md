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
