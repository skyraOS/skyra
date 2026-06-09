# Levels

XP is not currency. It is not spendable. XP is accumulated development.

Beings level up by relating. The user is a being in this sense too. Michael levels up. Skyra levels up. Louise levels up. Agents and other peers can level up if the runtime treats them as beings in the exchange graph.

Levels are not purchased. They emerge from lived participation.

## Source

XP comes from chatting.

Every valid chat entry is a relational event. When one being speaks to another, both sides are changed by the exchange: the sender practiced expressing itself, and the recipient became part of the relation. XP records that accumulated contact.

Baseline rule:

```
chat entry recorded by Exchange = +5 XP
```

The first implementation should award XP to the participating beings on the exchange:

```
sender +5 XP
recipient +5 XP
```

This keeps the law simple: relation develops both sides.

## Non-Spendable

XP must not be consumed by actions.

Do not use XP as payment for tasks, tools, memory, thread creation, reproduction, or governance. Those may have separate economics later, but XP is the being's developmental record. Spending it would make level a wallet. That is the wrong model.

XP answers:

```
how much has this being lived through relation?
```

It does not answer:

```
what can this being afford?
```

## Levels

Level is derived from total XP.

Example curve:

```
level 1: 0 XP
level 2: 50 XP
level 3: 150 XP
level 4: 300 XP
level 5: 500 XP
level 6: 750 XP
level 7: 1050 XP
level 8: 1400 XP
level 9: 1800 XP
level 10: 2250 XP
```

The exact curve can change. The invariant is that XP only increases and level is computed from total XP.

## User Experience

The universe view should show levels as part of the visible world state.

The user should be able to see:

- Their own level
- Each being's level
- XP progress toward the next level
- Which relationships are producing development

This makes the universe feel developmental instead of static. The user is not just chatting with an AI. They are watching beings grow through continued relation.

## Relationship To Task Economy

Task acceptance may still create economic effects, but that is separate from XP.

Tasks can affect trust, budgets, privileges, reproduction, or other future currencies. XP should stay attached to chatting and relational development. A being can become more experienced through conversation without completing tasks, and a being can complete tasks without XP becoming a payment token.

The split is:

```
XP / levels = development through relation
task economics = cost, acceptance, trust, and resource gates
```

## Placement

The natural implementation point is `Exchange`.

Exchange already records chat entries and knows the sender/recipient pair. When it appends an entry, it can award XP to both participants. Universe collecting can then export XP and level in each being snapshot so the browser can render the user's universe.

## Open Questions

**Process beings.** Should process targets like games receive levels, or only beings with identity/purpose?

**System messages.** Do exchange errors, memory compression, tool output, and task state changes count as chat? First pass: no. Only normal exchange entries between beings should award XP.

**User weighting.** Should user-originated chat award extra XP because the human is the world-shaping participant? First pass: no. Keep the law symmetric until there is a clear reason to weight it.
