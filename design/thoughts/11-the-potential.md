kk# The Potential

Date: 2026-09-23  
Status: agreed 2026-09-23 (D9). Follows [10](10-the-move-is-the-state.md).
Naming revised twice later the same day (see the revisions at the end). Final:
the thing is a **Potential**, and its states are **resolve → invoke → return /
fault**.

## The question

If Invoked, Returned and Faulted are states, what is the thing that has them?
It has to be addressable. And `ls ⟨somedir: hole⟩` is not a computation yet,
because it does not have its arguments. So is it a **potential**?

## Proposal

The addressed thing is a **Potential**: an executable applied to arguments,
some of which may still be holes. It gets an address when it is written, and it
keeps that address through every stage.

```text
Executable   ls                        registered code; not a computation
Potential    ls ⟨somedir: hole⟩        addressed; cannot compute yet
Invoked      ls "somedir"              every argument is a value; committed; in an instance's custody
Returned     "a.txt b.txt"             terminal
Faulted      …                         terminal
```

| Stage | Meaning | Address | On restore |
|---|---|---|---|
| Executable | Registered program | `registry/<id>` | Kept |
| **Potential** | Applied, not yet computable: at least one argument is a hole | `hash(first record)`, its frame id | **Kept. Nothing is running, so nothing is lost** |
| Invoked | All arguments are values, handed to an instance | Same | Faulted if no live supervisor holds it |
| Returned / Faulted | Terminal, written once | Same | Kept |

This refines [10](10-the-move-is-the-state.md). There, "waiting" was Invoked
with holes. Here, **a Potential with holes is not Invoked yet.** It becomes
Invoked at the moment its last hole is filled and an instance takes custody of
it. If there are no holes, that happens immediately.

## Why the split earns its keep

1. **"Invoked" now means exactly one thing:** in a live instance's custody. That
   is the custody invariant from [02](02-custody-and-the-stamp.md), named as a
   state.
2. **Restore gets gentler, and correct.** Only Invoked things can be interrupted
   mid-execution, so only those are faulted on restore. A Potential waiting on a
   hole has no process behind it and survives a crash untouched. When its
   dependency returns (after restore, or re-run by the operator), it proceeds.
3. **Content addressing lines up.** A Potential's address is the hash of what
   was asked for. When it becomes Invoked with every argument a value,
   `hash(executable, argument values)` is the memoization key from
   [10](10-the-move-is-the-state.md).
4. **The Invoke Move creates a Potential.** A Move with holes creates something
   that waits. A Move without holes creates something that is Invoked at once.
   The wire keeps its three variants.

## Lifecycle

```text
                 last hole filled + custody
   Potential ──────────────────────────────▶ Invoked ──▶ Returned
       │                                        │
       └── a hole's dependency faults ──────────┴──────▶ Faulted
```

A Potential can fault without ever being invoked, when a dependency it waits on
faults. Fault propagation (DNA-19's recursive descendant fault) runs along the
holes, towards everything that depended on the failed result.

## Naming note

The 09-20 and 09-21 corpus used **Possibility** for the process that owns an
execution context, later renamed `ExecutionContext`. "Potential" is close in
spirit but means something different here: an addressed, not-yet-computable
application, not a process. That needs to be stated wherever both terms appear,
or one of them should be retired.

## Open

1. Is a Potential's address the hash of its first record, so holes are part of
   its identity, or the hash of the executable plus argument references? The
   first is simpler. The second lets two identical requests share one Potential.
2. How long may a Potential wait? It needs either a deadline or garbage
   collection, or a hole whose producer never runs will keep it forever.

## Naming revision (2026-09-23)

"Resolve" was considered as the name of the thing and rejected as a noun,
because it is a verb and in promise terminology "resolved" means done
(Returned). It is kept for the transition instead. That also removes an overlap
in the table above, where "Potential" named both the thing and its first state.

| | Name |
|---|---|
| The addressed thing | **Potential** |
| Its states | **Unresolved** → Invoked → Returned / Faulted |
| Filling the holes | **resolve**: intake resolves a Potential; it is resolved when its last hole fills |

Read the "Potential" row in the stage tables above as the **Unresolved** state.

## Naming revision 2 (2026-09-23): the naming test

A state name must read naturally in **"the system is in a state of ___"**:

- the system is in a state of **invoke**
- the system is in a state of **return**
- the system is in a state of **fault**
- the system is in a state of **resolve**

"Unresolved" fails the test, and so do the past-tense forms (Invoked, Returned,
Faulted). The states are the plain verbs, matching the Move variants. This
reverses the objection above: "resolved = done" belongs to the past participle,
while "a state of resolve" means resolving right now.

**resolve** covers everything before an instance takes custody: holes unfilled,
or all filled and waiting for a runner. It absorbs the proposed "Resolved"
state, so no fifth state is needed. Resolve is always safe to restart, because
nothing has run. It is the only state without a wire Move, because resolving is
the runtime's own work, not a message.

Final: Potential; **resolve → invoke → return / fault**.
