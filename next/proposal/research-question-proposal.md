# Research question: proposal

Proposes changes to [research-question](../research-question.md). Ids refer
to the [list](README.md).

## What changes

The question and the thesis hold. The **system description** and **H3**
change with the new design, and both are claims a reader will check.

### "The system": two kinds of node

| It says | Now |
|---|---|
| Leaf: "a Horizon with **empty reach**: it can only return. Any invoke it attempts is refused" | a leaf runs under its caller's Horizon (B6). Leaf and continuation are behaviours, not authority. Containment is per Horizon, and a caller narrows to contain a tool (D2) |
| Intake "parses them as one Move (invoke, return or fault)" | derives the state (14): the slot is a return, a name in reach is an invoke, unaddressed output is a leaf fault |
| Release: "the parent… sends the child its Horizon, which is always a subset of the parent's… that fd is the capability" | the room derives the Horizon; the child is handed its hash and looks it up (B3, B4). Authority is derived from state plus proposal, never held |
| Execute: "a Seatbelt sandbox rendered from its Horizon" | holds, one profile per Horizon, plus the exclusions |

### H3, containment

H3 measures "how narrow a Horizon can be while H2 still holds". Under a
shared Horizon, the unit of narrowing is the **Horizon**, not the tool call.
That makes the measurement cleaner: count the narrowings (`system.narrow`
invokes) a task needs, and the authority each Horizon holds, against the
baseline's one ambient authority. Agent-built tools are contained by the
Horizon they are registered under (24, registration).

State the trade explicitly: a tool shares its caller's authority unless
narrowed. The claim is that effects outside **the Horizon** succeed zero
times, not that each tool is individually least-privileged.

### H0, interruptions

Gains a mechanism: an authority fault climbs to the human only when no
ancestor holds the authority, and the human's answer is one invoke of the
room's admin facet (23 proposal).

## Edits to the document

- Rewrite "The system" table and steps as above.
- Reword H3 to the Horizon as the unit, with the trade stated.
- Link 24 and 26 under "The system".

## Code

None.
