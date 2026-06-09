# One-Way Relationships

## What This Is

Relationships become directional. A declaring B in its relationship list means A can address B. It does not mean B can address A. B can only address A if B also declares A. The system already stores relationships independently per entity — this change makes the world enforce that independence through physics.

---

## The Problem

Currently, relationships are declared per-entity in the genome but treated as symmetric by convention. If skyra lists builder and builder lists skyra, both can talk. But if skyra lists bash and bash lists skyra, the relationship carries the same weight in both directions — even though bash is a shell port that should only respond, never initiate.

There's no way to express: "this entity can reach that one, but not the other way around." Every relationship is implicitly bidirectional. That collapses a real distinction. An entity that observes is not the same as an entity that participates. An entity that responds is not the same as an entity that initiates.

---

## The Mechanism

The genome already declares relationships per-entity. Nothing changes in the declaration format:

```
grow ~name skyra ~archetype llm ~identity ... ~relationships michael,builder,skeptic,bash
grow ~name skeptic ~archetype llm ~identity ... ~relationships skyra
```

Skyra can address michael, builder, skeptic, and bash. Skeptic can address skyra. Skyra cannot address skeptic unless skyra also declares skeptic — which it does. But bash declares only skyra, so bash can respond to skyra but cannot initiate a message to michael or builder.

The change is enforcement. The world checks the sender's relationship list before routing. If the target is not in the sender's relationships, the message is dropped. This is world physics.

---

## What Changes in the Runtime

**World physics** — after resolving the target entity, check that the target is in the sender's relationship list. If not, drop the message. This is the single enforcement point, implemented as a physics rule on the world.

**Peer list in present** — already correct. It only shows the entity's own declared peers. No change needed.

**`DerivePresent`** — no change. The present already only shows the entity's own peers. An entity that can't address someone never sees them in its peer list.

**The genome** — no format change. Relationship declarations already live on each entity independently. The semantic shift is that the world now treats them as directional grants, not mutual associations.

---

## What This Means

An entity's relationship list is its reach. Not its visibility, not its awareness — its ability to initiate or respond to a specific peer. The world enforces this at routing time through physics.

This creates three possible states between any two entities:

- **Mutual** — both declare each other. Full exchange. (skyra <-> builder)
- **One-way** — A declares B, B does not declare A. A can address B. B cannot address A.
- **None** — neither declares the other. No direct exchange possible.

One-way relationships make asymmetric roles expressible. A watcher that can speak but can't be spoken to. A tool that responds but never initiates. An authority that broadcasts but doesn't take questions. The genome already has the syntax for this. The world just needs to enforce it.

---

## What This Is Not

Not access control. Not permissions. An entity that can't address another entity directly can still influence it through intermediaries — skyra can carry a message from bash to michael if both trust skyra. The relationship constraint is on direct address, not on information flow.

Not a change to the exchange model. Exchanges still track both parties symmetrically. The directionality is about who can open or continue an exchange, not how the exchange is stored once it exists.
