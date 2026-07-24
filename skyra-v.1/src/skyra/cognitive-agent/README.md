# Ky Agents

This directory contains Skyra agents whose defining responsibility is cognition.

An agent may retain observations, predictions, learned relations, and planning state. It
does not own the topology or authoritative live state of the world it occupies. Moving an
agent between local, spatial, or device-backed worlds should change its mount, not its
cognitive identity.

Current implementation:

- `skyra-dna-v3/` — the active v3 cognitive substrate.
