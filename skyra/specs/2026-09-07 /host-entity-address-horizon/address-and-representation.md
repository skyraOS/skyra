# Address owns representation

**Current structure:** the address defined here is `EntityAddress`, which owns Representation.
Traversal is Entity → EntityAddress → HostAddress → Host. Earlier bidirectional pointers below
describe the conceptual relationship, not the accepted pointer layout. HostAddress carries host
contact information; representation ownership is not assigned to it. See [the proposal](proposal-entity-address-host.md).

**Date:** 2026-09-07  
**Status:** DECIDED; concrete types and encoding remain open.

EntityAddress identifies a stable entity endpoint that may be shared across multiple hosts.
**The representation contract belongs to that entity address.** Its identity and representation stay
stable while the host routes used to reach it can change. This refines the earlier one-host binding:

```text
Entity → EntityAddress → HostAddresses → Hosts
              │
              └─ representation / parser contract
```

The entity address owns how history, live work, and results are parsed and represented for its endpoint.
An LLM endpoint and a compiled-execution endpoint may require different representations. Compatibility
is defined at the address, not imposed globally on every entity, host, or engram. This supersedes the
suggestion to place representation ownership only on an executor behind the address.

History includes invocations, opcodes, observations, errors, and their relationships—not merely a list
of observations. Its presentation must preserve the relationship between an invocation and its recorded
outcome. Opcode-specific parsers/renderers may compose behind the address-owned representation boundary;
their concrete interface and the stored history shape remain open.

Representation preserves the [history versus live-work distinction](execution-history-and-live-work.md).
It must not invent an outcome for outstanding work. A timeout records expired waiting, not necessarily
that the underlying operation stopped. Exact pairing and late-result handling remain to be defined.

The proposed `History []Observation` Engram field was too narrow and is not adopted. No replacement
Engram struct is selected by this decision. Address identity, encoding, and its relationship to a
changing transport locator also remain to be specified.
