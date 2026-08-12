# Terminology

**Date:** 2026-08-11
**Status: REFERENCE.** What each word means, and what it does not. Where two words were used for
one thing, or one word for two, that is recorded.

---

## The deployment layer

```text
SYSTEM    the whole network of runtimes. specifically the network layer.

RUNTIME   one per device that can run one. AN AUTHORITY DOMAIN.

KERNEL    one per runtime. the authority over it. holds everything HAPPENING there.

HOST      physical hardware addressed over a network — INCLUDING devices with no OS.
          a phone with no runtime on it is a HOST, not part of a distributed runtime.
          `here` is the degenerate case: the device you are standing on.
```

**A runtime does not span devices.** It reaches them. That is transport, and transport is beneath
everything.

**Nothing crosses a runtime boundary except gossip** — which runtimes exist, which entities exist.
Not graphs, not observations, not permissions, not history. No global registry, and nothing holds a
model of the whole.

## Durable and ephemeral

Every pair in the design is the same pair.

```text
NAMESPACE     durable.   AN ENGRAM. a principal's top engram. children are what they can select.
              persists between sign-ins, and its history is the record of what they asked and
              what they kept.

ENGRAM        durable.   an expectation with a history accumulating around it. owned by exactly
              one parent.

POSSIBILITY   ephemeral. one opening of an engram.

CHAIN         ephemeral. live work under a namespace, keyed by SPAN. the transaction boundary —
              staged effects live here and materialize at acceptance.
```

```text
RETIRED   "ROOT". it implied a special node and was not one. every wrong turn — root-as-engram,
          root-as-namespace, root-as-hinge — came from that implication. what it named is a
          CHAIN, and the kernel already holds per-span state for it.
```

## Who and where

```text
ENTITY        WHO interprets. an identity and nothing else — every declaration dissolved.
              registered per runtime; known across runtimes by gossip.

PRINCIPAL     WHO ASKS. holds a namespace. an entity that can INITIATE; interpreters need
              no namespace.

HOST          WHERE it lands. see above.
```

## Authority

```text
GRANT         what may be DONE. the world boundary. not derivable from a frame.

PROVISION     the ACT of handing down access from access you hold. it can only narrow.
              boot provisions the admin namespace; the admin namespace provisions the rest;
              a select provisions its children.

SCOPE         what may be NAMED. the graph boundary. pure name resolution, and this is the
              object-capability half: you can select what you can name, you can name what you
              were given.
```

```text
RETIRED   "MOUNT". it was five existing mechanisms under one word — a child engram placed by
          provisioning · ambient names placed at boot · grants on the chain · entity reach ·
          copy. none of them needed a new type.

          And grants-on-the-reference was specifically wrong: EFFECTS HAPPEN AT LEAVES, so
          authority attached to a name at the top bounds nothing. grants bound the CHAIN,
          because the chain is what connects a leaf to an authority.
```

## The forms and the names

```text
SELECT        choose. names one or more engrams, binds observations, PROVISIONS ACCESS.
COMMIT        internalize. the value up, and the only durable write.

discover      reads the GRAPH.    what HAPPENED.     engrams · entities · observations · hosts.
introspect    reads the KERNEL.   what is HAPPENING. chains · heads · ancestry.
authenticate  reads the WORLD.    WHO IS THIS. not reflexive, so not a peer of the two above.
provision     narrows access downward. an engram, invoked from a namespace.
```

`discover` and `introspect` are paired because both are **the system asking about itself**, split on
the same durable/active line as everything else. `authenticate` asks the world and is a third kind,
closer to `create` — which is also a system engram and also not reflexive.

## Words with one meaning that carried two

```text
COMMIT        THE EMISSION, which hands an observation up and leaves the possibility live.
              distinct from CLOSURE, which is the acceptance that ends a chain and is the only
              durable write. the set used one word for both.

CROSSING      between RUNTIMES: real, and not a mechanism — two chains, a stub on each side,
              each kernel enforcing its own.
              within a runtime: does not exist. authority never moves.

ORIGINATION   producing content that was not derived. governed by "originate only where it
              could not be mistaken for composition."
              INGRESS IS NOT ORIGINATION — an inbound observation was produced by no
              possibility, so the question is a category error.
```

## Deliberately absent

```text
NO GLOBAL REGISTRY        gossip carries which runtimes and entities exist. nothing else.

NO GLOBAL GRANT ALPHABET  a grant says nothing about what you may do in another runtime.
                          that runtime's kernel decides, from its own data. two runtimes need
                          never agree on what a grant string means.

NO CONTROL PLANE          drive is a select, accept is a commit. one channel, two forms, and
                          names reached by the forms.
```
