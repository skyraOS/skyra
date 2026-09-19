# Working context, compilation, and Remember

**Date:** 2026-09-07  
**Status:** Boundaries decided for now; concrete storage and interfaces remain open.

## Executable data and its executor

Executable data is an open-ended category: an LLM's working context, source code, bytecode, a compiled
binary, or other material interpreted by an executor. Engram should not enumerate every possible
representation. Composition belongs at the addressable boundary tied to entity and host: the selected
executor knows how to interpret the material.

The [placement decision](move-and-engram.md) holds: entity, address, and host exist on Move;
the executable thing exists on Engram. Routing and authority do not depend on teaching CTU every
executable-data format.

## Working context: history and live work

Before compilation, an addressed LLM interprets the working context; a human can participate in
changing the work. The register is the current home proposed for that context. It separates:

- **History:** invocations, opcodes, changes, work already performed, observations, errors, and their
  relationships, available to the human/LLM making further changes. A change to an instruction is not
  evidence that the instruction executed. History is not merely `[]Observation`; its stored shape is open.
- **Live work:** the current work being developed or carried out. This is the material to compile.

Changes can be recorded as line items attributable to the entities that made them. Recording a working
change does not by itself create a committed version. The earlier suggestion to treat every line item
as a separately versioned executable is not adopted.

**History informs the changer; live work is the compilation target.** The compiled executor need not
interpret the history of how its artifact was authored. If the behavior needs remembered observations,
those still need an explicit input or data representation; compilation does not erase required inputs.

[Address owns the parser/representation contract](address-and-representation.md) used to present this
material at its entity–host endpoint. There is no required universal human/LLM history format.

The current definition of live work also includes work already executing, waiting, or touching the
world. Its representation must distinguish the work being changed/compiled from an invocation already
in progress. No fields for that distinction are chosen here.

## Compilation: what is current

Compilation produces a runnable artifact for the selected executor; it need not mean native machine
code. After compilation, an addressed executor runs that artifact with the invocation's current state.
The artifact and its running state remain distinct.

**For now, the most recent successful compile is the current executable.**

- Successful compilation advances the executable selected for new invocations.
- Failed compilation leaves the previous current executable intact; failure feedback can inform
  further working changes.
- An invocation already executing keeps the executable version it started with.

Successful compilation does not by itself prove behavioral correctness. The checks required for a
compile to count as successful remain to be defined; no separate approval/version-selection workflow
is introduced here.

## Remember: what is committed

**Remember is the user-initiated commit boundary. Compilation and commit history are decoupled.**

A successful compile does not automatically commit. Remember marks a committed version separately
from whichever successful compile is current. Further edits and successful compiles can advance current
without automatically advancing the committed version.

Working history is available to the changer before Remember. Remember is not the only moment at which
observations or edits can be recorded. Exactly what it captures—working history, live work, compiled
artifact references, learned contracts, or their relationship—remains to be specified.

| Surface | What advances it |
|---|---|
| Working context | Edits, observations, and ongoing work |
| Current executable | The most recent successful compile |
| Committed version | User-initiated Remember |

These are distinct responsibilities, not a requirement for additional manual save/activate/commit
controls. No separate version-selection lever is selected for now: latest successful compile is the
default for new execution.

## What this replaces

The discussion first considered compilation as the commit boundary, with precompile changes treated
only as working history. The latest decision keeps that working-history distinction but **moves the
commit boundary to Remember**. Compile and Remember must not be implemented as one automatic event.

Older descriptions of REMEMBER compiling the trace or invariably emitting the executable at persistence
must be reconciled with this separation. Observed contract information may still inform reusable
behavior; its discovery does not imply an automatic commit or settle who emits the compiled artifact.

## Continue from here

Define how working context, current executable, and committed version fit Engram and its lifetime.
Specify what Remember captures and how identity/revisions address these forms. The existing shared
MoveID/EngramID decision remains, but its revision encoding is not settled by naming these boundaries.
No new structs, fields, automatic commit triggers, rollback policy, or garbage-collection rules are
introduced by this document.
