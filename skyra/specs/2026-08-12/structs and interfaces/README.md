# Structs and Interfaces

**Date:** 2026-08-11 · verified against the runtime 2026-08-13
**Status: LIVE SNAPSHOT.** The current type surface of `skyra-dna-v5`, grouped so each interface
sits with the structs that implement or feed it. The source tree is authoritative; the documents in
`interface changes/` are proposals and are not descriptions of code that already exists.

Every type below carries one of three marks:

```text
SPEC        declared in 2026-08-09/interfaces.md.
SPEC(n)     NAMED in that doc's prose or a signature, never declared. the shape was implied.
ADDED       not in interfaces.md at all. these are what there is to review.
```

## The documents

| doc | holds |
|---|---|
| [01-ids](01-ids.md) | identity, and the one minting mechanism |
| [02-content](02-content.md) | the closed Value set, and the two read-time folds |
| [03-graph](03-graph.md) | the durable layer, the witness, and the credit walk |
| [04-grammar](04-grammar.md) | the two forms |
| [05-frame](05-frame.md) | what a possibility sees |
| [06-entity](06-entity.md) | the interpreter boundary, the registry, three implementations |
| [07-host](07-host.md) | capability classes and the platform layer |
| [08-kernel](08-kernel.md) | the loop, the root as transaction, staged effects |
| [09-trace](09-trace.md) | outside the design, and says so |

## Where the additions are

```text
1  A GENUINE ADDITION TO      EngramRef. it extends content.Value, which is CLOSED so that Cites
   THE SEMANTIC MODEL         can be derived exhaustively. see 02-content.

4  READ-TIME FOLD MACHINERY   Partition · unexported diff · Summary · Deref. all derived views,
                              nothing durable — and collectively a layer. see 03-graph.

6  PLUMBING THE SPEC          Graph · Minter · Registry · Index · Author · Mint. the notable one
   LEFT OPEN                  is Graph: interfaces.md says "the graph applies that record as
                              writes" and never declares it, so the witness — the only thing with
                              policy — is the biggest unnamed object in the set.

3  SCAFFOLDING                Pending · Discard · ErrNoEmission.

1  OUTSIDE THE DESIGN         trace.Run, by admission.
```

Six further types — `Live`, `View`, `view`, `Local`, `localExec`, `localStore` — are implementations
of interfaces `interfaces.md` declares, and are not counted as additions.

## What is worth cutting

Ranked by what I would defend least:

```text
THE FOLD TYPES     `Partition` and `Summary` remain. `diff` is deliberately unexported and
                   `UsedSeen` no longer exists.

Summary            pure presentation. it exists because something had to render a child set into a
                   prompt, which is a marshalling concern and belongs in the entity adapter.

Handle             already cut to `{Host, Ref}`. Process output is not part of the handle.

ErrNoEmission      an error type for one message.
```

## Related

[`../semantics/the-collapse.md`](../semantics/the-collapse.md) records the earlier surface collapse.
Its later seven-peer-method ruling is itself superseded by
[`../continuation-language/primitives.md`](../continuation-language/primitives.md). This directory
describes the runtime that exists; `interface changes/` preserves proposed replacements and marks
their current status.

[`../../2026-08-11/semantics/conversation-runtime-mismatch.md`](../../2026-08-11/semantics/conversation-runtime-mismatch.md)
records where the runtime and the conversation disagreed. Two rows concern stored error observations
and reuse-by-pointer versus reuse-by-copy.
