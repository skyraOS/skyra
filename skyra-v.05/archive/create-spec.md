# Create — Realities Creating Realities

## Current State

Grow is a hardcoded factory on NewThread (`newthread.go:206`). It's a switch statement that knows the full internal composition of every being type. It's the one place in the system that breaks composability.

```go
switch beingType {
case "llm":
    // knows Self needs Think with recall/remember/skill
    // knows Self needs Act with plan
    // knows Think needs an LLM component
case "user":
    // knows User needs a device
}
```

This works but it means Thread knows the internals of Self, Think, Act, User, and every operator. Nothing else in the system works that way.

## The Problem

A reality needs context to create another reality:
- What devices exist?
- What components are available?
- Where does the new reality register?

That context lives above the creator in the tree. So either:
1. The creator reaches up (breaks the recursion)
2. The context comes down on the relation (stays composable)

## Target: Create Is the Factory

Every Reality already has `Create(*Relation) Reality`. If the Relation carries enough context, any Reality can create sub-realities within its own world.

- Thread creates beings — it puts available devices and components on the relation, calls Create on the being type
- Self creates operators — it puts its Think/Act layers on the relation, the operator registers itself
- MacOS creates components — it puts its component registry on the relation, the component registers itself
- A being could create another being if it has the right context on its relation

Each level creates within its own world. No central factory. The relation is the context carrier.

## The Genome Is a List of Create Calls

The verb is the type. Each line creates a reality and can only reference what's been declared above it.

```
device ~name macbook ~type macos
component ~name terminal ~type stdin ~device macbook
component ~name ws ~type websocket ~port 8080 ~device macbook
component ~name openrouter ~type llm ~model anthropic/claude-sonnet-4-5 ~device macbook
being ~name skyra ~type llm ~devices macbook
being ~name michael ~type user ~devices macbook
being ~name louise ~type llm ~devices macbook
```

The parser reads top to bottom. Each line:
1. Sees the verb (the type of reality to create)
2. Builds a Relation with the impulse and available context
3. Calls Create on the matching Reality type
4. Registers the result in the appropriate world

Bootstrap and runtime creation are the same path. The genome is the language. Create is the interpreter.

## Order Matters

A line can only reference things declared above it. If you put a being before its device, it fails. The dependency order is explicit in the file — you read it and know what exists at any point.

At runtime, the same rule applies. A being can only create something if its dependencies already exist in the tree. The Relation carries what's available. Create uses what's there. If a dependency is missing, creation fails.

No second pass. No lazy resolution. No circular dependencies. Top to bottom.

## What This Means

### Create becomes richer

Currently Create is mostly a constructor — it takes a Relation and returns a new Reality with some fields set. Under this model, Create also receives the world context it needs to wire itself in.

The Relation already has:
- `Impulse` — the configuration / declaration
- `Realities` — map of available realities (devices, components, peers)
- `Exports` — arbitrary key-value context

A Self.Create could:
1. Read its identity/purpose from the Impulse
2. Find available LLM components from Realities
3. Create its own Think and Act by calling their Create with the same context
4. Return a fully wired Self

Thread's role shrinks to: build the relation with the right context, call Create on the right type, register the result.

### Runtime creation

At runtime, a being saying `being ~name poet ~type llm ~devices macbook` would go through the same path as bootstrap — the impulse gets parsed, a Relation gets built with available context, Create gets called, the new reality registers. The genome and runtime creation are the same mechanism.

### Beings creating beings

If a being has enough context on its relation (devices, component refs, peer list), it could call Create on a new being type. This is the reproduction path from world-physics — "creating a new being costs accumulated experience." The economics layer gates it, but the mechanism is just Create with context.

## Open Questions

1. **What context does Create need?** The minimum set for each type. Self needs an LLM component. User needs a device. Operators need nothing. What's the shape of the relation that carries this?

2. **Registration.** Create returns a new Reality, but someone has to put it in the right hashmap (Thread.Beings, MacOS.Components, Think.Operators). Does Create register itself, or does the caller register the result?

3. **Alpha scope.** How much of this do we need before June? The current Grow works. The refactor is architectural purity. Worth doing if it unblocks something (like beings creating beings), not worth doing just to be clean.
