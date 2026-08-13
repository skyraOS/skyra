# Terminology

**Date:** 2026-08-11
**Status: REFERENCE.** What each word means, and what it does not. Where two words were used for
one thing, or one word for two, that is recorded.

> **CARRIED FORWARD** from `specs/2026-08-11/semantics/terminology.md`, and still **AUTHORITATIVE
> on vocabulary**. The body below is unchanged. What 2026-08-12/13 renamed, and one word it
> reintroduced against this document's own ruling, is recorded in the addendum at the end.

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

SESSION       ephemeral. ALL live work under one namespace. one per mounted namespace, and it
              holds what a sign-in provisions — grants and reach, once, not per ask.
              created by Mount, ejected by Unmount. its asks are top-level possibilities,
              keyed by span, and staged effects are partitioned by the same key.

BRANCH        A POSITION, NOT A TYPE. a possibility with no parent — one ask. keyed by span,
              and the transaction boundary: staged effects are partitioned by that span and
              materialize at acceptance.

              A BRANCH IS A POSSIBILITY. what looks like a branching branch is possibilities
              nesting, which is the only nesting there is. accepting one ask must neither land
              nor discard another's writes — and that needs a KEY on the tree, not a struct.
```

```text
RETIRED   "ROOT". it implied a special node and was not one. every wrong turn — root-as-engram,
          root-as-namespace, root-as-hinge — came from that implication. what it named is a
          BRANCH, and the kernel already holds per-span state for it.

RETIRED   "CHAIN" AS A TYPE. still fine as ordinary prose for a line of work; it is not the
          name of the object.

RETIRED   "TREE" AS A TYPE. it named a shape the object stopped having the moment BRANCH
          stopped being one: its contents are a map of independent asks, and each of THOSE is
          a tree. the container holds several. it is a SESSION — what sign-in creates and
          sign-out ejects.

NOTE      FOUR WORDS ARE POSITIONS AND NONE IS A TYPE — leaf, orchestrator, head, branch.
          leaf and orchestrator derive from whether a possibility has selected; head from
          whether it is executing; branch from whether it has a parent. `kind is derived from
          composition and never declared` applies to all four.
```

## Head

```text
HEAD      a LEAF of the live tree — a possibility that is executing rather than waiting.
          DERIVED BY WALKING, never counted.

              head  ==  no live children  AND  no return under the current draft
```

Three node states, all inferable from structure, so no state field is needed:

```text
EXECUTING   inside Step, no live children.        A HEAD.
WAITING     children are live.                    not a head.
RETURNED    committed and still live, re-drivable. not a head until re-driven.
```

**A counter cannot express this.** `kernel.go` increments on entering `open()` and decrements on
leaving, so a depth-five chain with one leaf running reports five — four of which are ancestors
sitting in `runGroup`. That is not a frontier, and it makes *a stalled head* indistinguishable from
*a parent legitimately waiting*, which is the eviction decision the number exists for.

### What introspect does with them

**One descent, two filters.** It is a TREE, not a stack — cardinality > 1 means several heads at
once, so the answer is a set.

```text
STOP DESCENDING       no live children.
"WHAT AM I DOING"     those leaves with Returned == nil.   THE HEADS.
"WHAT HAS THIS DONE"  those leaves with Returned != nil.   finished, still live, revisable.
```

The second question did not exist before `open-until-commit`, because returning meant dying.

**Topology only.** The kernel may report where things are, never what they mean:

```go
type Head struct {
    Span   ids.SpanID
    Engram ids.EngramID     // AN ID. never the expectation.
    Entity ids.EntityID
    Host   ids.HostID
    Depth  int
    Since  time.Duration    // time in Step. this is what makes STALLED detectable.
}
```

The caller dereferences the engram against **their own graph**, which they may read because it is
their namespace. Same move as `Producer` being a field so `ObservationID` stays opaque: **the kernel
hands back handles and the reader resolves them.** It is also the fix for the two live `NO SEMANTICS`
breaches — `shortExpr` and `oneLine` exist because the kernel wanted to say something readable and
had no other way.

**And the subtree bound is structural, not enforced.** Descending can only reach descendants, so
`DOWN, NEVER UP` is a fact about which way the pointers point rather than a rule anyone checks.

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
CEILING       the maximum grants a namespace may ever be mounted with. DURABLE, written onto
              the namespace engram by whoever provisioned it. boot writes admin's; admin writes
              everyone else's.

MOUNT         A KERNEL METHOD. bring an entity, a namespace engram and a policy together and
              hold them — the key on the rack. it is what makes a namespace workable, and
              it is the ONLY exported kernel method.

              Mount(entity, namespace, grants) → Session   requires grants ⊆ ceiling
              Unmount(namespace)                           sign out. the key is ejected.

THE RACK      the kernel's flat map of mounted namespaces. ADMIN IS A ROW ON IT, not a level
              above it.
```

**Mount is not the pairing that opens a possibility.** That pairing is internal and stays
unexported. Mount attaches a *namespace*; opening attaches an *engram*.

**The rack is flat because lineage and activity are different questions.**

```text
PROVISIONING LINEAGE   admin provisioned my namespace.   DURABLE. what happened.
THE RACK               who is signed in right now.        EPHEMERAL. what is happening.
```

If the rack mirrored the lineage, my namespace being active would require admin's to be active —
admin signs out and I am logged out with them, which is wrong the same way `root` not being logged in
does not stop you using your account. **What makes admin special is not a position on the rack. It is
that its ceiling was written by boot rather than by another namespace.**

And that is why the ceiling has to be durable: `Mount` reads it and compares, so **no live admin is
needed to authorize a sign-in.**

```text
RETIRED   "MOUNT" AS A NOUN — a name binding inside a namespace. that was five existing
          mechanisms under one word: a child engram placed by provisioning · ambient names
          placed at boot · grants on the branch · entity reach · copy. none needed a new type.

          And grants-on-the-reference was specifically wrong: EFFECTS HAPPEN AT LEAVES, so
          authority attached to a name at the top bounds nothing. Grants bound the BRANCH,
          because a branch is what connects a leaf to an authority.

RETIRED   "INGRESS" AS A METHOD. it described a model where the method minted an inbound
          observation; it no longer does — the payload is an argument to the select that
          follows. MOUNT is the same method named for what it does.
```

## The forms and the names

```text
SELECT        choose. names one or more engrams, binds observations, PROVISIONS ACCESS.
COMMIT        internalize. hands an observation up. THE POSSIBILITY STAYS LIVE.
              NOT the durable write — see COMMIT vs CLOSURE below. CLOSURE is.

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


---

# Addendum — 2026-08-13

## The forms and the names, renamed

[`../primitives.md`](../primitives.md) renames five of these and adds two. The concepts are
unchanged; the words moved.

```text
08-11                    08-13            note
─────────────────────────────────────────────────────────────────────────────────────────────
SELECT                → DECIDE            the choice IS the binding
COMMIT                → REMEMBER          plus Accept, plus closure. see closure.md
discover              → THINK             search AND reason. RECALL is the flat lookup inside
                                          it; DISCOVER is reserved for novelty
introspect            → ORIENT            "reads the KERNEL, what is HAPPENING" carried verbatim
provision             → absorbed by DECIDE, which binds "whatever reach and grants are
                        explicitly provisioned downward"
"INGRESS" (retired    → APPROACH          and approach has an OPPOSITE, which ingress never had:
 here in favour of                        "<caller> was avoided" is an event with a subject
 MOUNT)
─────────────────────────────────────────────────────────────────────────────────────────────
                        REACH             NEW. no 08-11 equivalent.
                        RECOGNIZE         NEW.
```

**`discover`/`introspect` reading the graph against the kernel carried across the rename intact** —
`primitives.md` states it in the same two lines. That pairing is the most durable thing in this
section.

## `authenticate` has no successor, and `APPROACH` is where it belongs

```text
authenticate  reads the WORLD.  WHO IS THIS. not reflexive, so not a peer of the two above.
```

`primitives.md` does not mention it. But `APPROACH` asks *should this be let in at all*, and
`questions.md` §5 says nothing gates that — so the missing gate and the unplaced word are the same
hole:

```text
APPROACH        the primitive. should this be let in?
  authenticate  answers WHO — and this is the half §5 says is missing
  Mount         answers WHAT, already designed: grants ⊆ ceiling
```

The ceiling bounds what an admitted session may *do*. Nothing bounds who may open one. That is
`authenticate`, and it is specified here and built nowhere.

## `ROOT` was retired here, and 08-12/13 reintroduced it

```text
RETIRED   "ROOT". it implied a special node and was not one... what it named is a BRANCH.
```

`primitives.md` §1 says *"Approach is always at a root, and always outside-in."* By this document
that should read **branch** — *a possibility with no parent, one ask, keyed by span, and the
transaction boundary.*

```text
THE CLAIM IS UNCHANGED    approach is always at a possibility with no parent. That is exactly
                          what a BRANCH is, and an approach is what creates one.

THE WORD IS WRONG         and it is wrong in the specific way this document warned about — "every
                          wrong turn came from that implication."
```

Not corrected in place, because `2026-08-09/questions.md` §5 is also written in the old vocabulary
and the two should move together.

## Untouched and still standing

Every other section of this document. The authority model in particular —
`GRANT` · `PROVISION` · `SCOPE` · `CEILING` · `MOUNT` · `THE RACK`, and the flat-rack argument —
is not addressed anywhere in the 08-12/13 work and carries whole.
