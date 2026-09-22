# identity/03 · Horizon

**Package:** `identity`

Authority. It travels with the frame, is not part of it, and only ever narrows.

---

## 1 · Opaque, and one operation

```go
type Horizon interface {
    isHorizon()
}

// kernel-internal. a possibility cannot call this.
func narrow(parent Horizon, requested Horizon) (Horizon, bool)
```

Opaque for the same reason `Rendering` and `Delivery` are opaque
(`../durable/03-crossing.md`): **the incapacity should be structural rather than a policy of
restraint.** A `Horizon` with readable fields is a `Horizon` something will eventually read, compare,
and branch on as application logic.

The only operation narrows. There is no `widen`, and the absence is the enforcement of P14.

```text
CHILD HORIZON ⊆ PARENT HORIZON
```

`narrow` returning `(Horizon, bool)` rather than an error is deliberate: a request for more
authority than the parent holds is not an exceptional condition, it is `BeyondYourHorizon`, which is
one of Reach's three ordinary conclusions.

---

## 1a · What it contains — 08-15, UNRATIFIED

**DECIDED IN SESSION:** *"horizons contain endpoints allowed within a chain."*

The first concrete statement of Horizon's internals anywhere in the set. See
[`../approach-and-root.md`](../approach-and-root.md) and
[`../00-principles.md`](../00-principles.md) P18 for how it arrived — the sandbox thread put all
effect behind runtime-mediated endpoints, which makes the endpoint set the thing authority is
actually *about*.

```go
type Horizon interface {
    isHorizon()          // still opaque to a possibility
}
// its content is a MANY-KINDED SET:
//     endpoints · packages · compilers · creators
// plus what think-interface-proposal.md already lists:
//     namespaces · credentials · hosts · authority domains · resource ceilings
```

> **It is literally what the AI could possibly see. It cannot see past it.**

That is not a new claim so much as the existing one made concrete.
`think-interface-proposal.md` already says Horizon bounds what a thought *"may **consider** or
provision downward"* — consider, not merely do. Horizon was always epistemic as well as
authoritative, and `Think`'s `BEYOND YOUR HORIZON` diagnosis already reports it as a limit on what
*exists for you*, not only on what you may cause.

### It does not become a second frame

The obvious objection, and it survives:

```text
FRAME     what the possibility CURRENTLY KNOWS and can name.   holds actual observations.
HORIZON   the outer bound of what COULD EVER enter it.         holds kinds of thing, no values.
```

Potential versus actual. Horizon still *"carries no application values"* — listing that a compiler
is available is not the same as holding an observation.

### Consequences

```text
NARROWING IS SET CONTAINMENT      child's set ⊆ parent's set. checkable, cheap, no content read —
                                  which is what lets the package check run BEFORE Reach rather
                                  than at call time.

"BEYOND YOUR HORIZON" GETS A      previously abstract. now: this composition requires a member
CONCRETE SUBJECT                  that is not in the set.

CREATOR MEMBERSHIP REPLACES       ../creator-and-compiler.md and ../primitives-as-engrams.md §2
THE INJECTION STORY               explained the create/interpret split as a capability HANDED IN
                                  — python cannot create because nothing gave it an `Author`.
                                  With creators in Horizon, that is simply membership: A CHAIN
                                  WHOSE HORIZON CONTAINS NO CREATOR CANNOT AUTHOR NEW BEHAVIOR.
                                  it can only compose what already exists. one mechanism instead
                                  of two, and it still declares nothing.

COMPILERS BECOME GRANTABLE        python may be in one namespace's Horizon and absent from
                                  another's.
```

### And it sharpens H3

With this many kinds in one set, monolithic narrowing means you cannot drop compilers without also
dropping endpoints. The Linux argument in [`../approach-and-root.md`](../approach-and-root.md) §8
— uid 0 decomposed into ~40 capabilities because all-or-nothing was too coarse — applies with much
more force to a set this heterogeneous than it did when Horizon was undescribed.

It does not close H3: whether these kinds are the whole of Horizon, and whether narrowing is
per-kind or monolithic, both remain open.

---

## 2 · What it bounds

```text
HORIZON
    the outer boundary of ENTITIES, NAMESPACES, CREDENTIALS, HOSTS, AUTHORITY DOMAINS,
    EFFECTS, and RESOURCE CEILINGS this thought may consider or provision downward.
```

Note what that list makes it: not a permission bit, and not a set of grants. It bounds what may be
*considered* as well as what may be *done* — which is why Think's diagnosis includes
`BeyondYourHorizon` as distinct from `NoKnownPotential`. Relevant potential can exist and be
invisible.

## 3 · Not a frame, and not an observation

Two collapses the design refuses by name.

**Horizon is not a second frame:**

```text
HORIZON CARRIES NO APPLICATION VALUES. It can only narrow down a causal chain.
```

**Horizon is not an observation**, and this is the sharper of the two —
`think-interface-proposal.md` lists it among rejected shapes:

```text
HORIZON TRAVELS AS AN OBSERVATION
    Authority would become caller-authored data. Horizon is supplied and narrowed by the runtime,
    outside the semantic operands.
```

So `Horizon` must not be a `content.Value`, must not be reachable from `Observation`, and must not
appear in `ThinkArgs`. In the type surface it lives in exactly one place — `Frame` — and it is not a
member of any operand struct.

The paired distinction with input contracts, from `engram-contracts-and-composition.md`:

```text
INPUT CONTRACT    what observational MATERIAL may be supplied.
HORIZON           how far outward the child may SEE or ACT.

"The two bounds travel together but do not mean the same thing."
```

---

## 4 · Where narrowing happens

```text
APPROACH   establishes the initial Horizon.        kernel/02-approach.md
DECIDE     provisions a subset downward.           primitives.md §3 lists HORIZON among the six
                                                   things a decision binds.
REACH      may report BeyondYourHorizon.           it does not narrow; it checks.
FOCUS      inherits, never wider.                  contracts/04-continuation.md §4
```

Only Decide narrows, and that matches `terminology.md`'s definition of provisioning as *"the ACT of
handing down access from access you hold. It can only narrow."*

`primitives.md` records that `provision` as a separate word was **absorbed by Decide** — *"which
binds whatever reach and grants are explicitly provisioned downward."* So there is no `Provision`
method, and narrowing is not separately callable.

---

## 5 · HOLE — the representation, and the unreconciled authority model

The shape above is the whole of what the 08-12/13 documents settle. Everything about what a Horizon
*is* comes from a different layer that has not been reconciled with the seven.

`../../../semantics/terminology.md` — authoritative on vocabulary, and it says the authority model
*"is not addressed anywhere in the 08-12/13 work and carries whole"*:

```text
GRANT       what may be DONE. the world boundary. not derivable from a frame.
PROVISION   the ACT of handing down access from access you hold. it can only narrow.
SCOPE       what may be NAMED. the graph boundary. pure name resolution, and this is the
            object-capability half.
CEILING     the maximum grants a namespace may ever be mounted with. DURABLE.
MOUNT       Mount(entity, namespace, grants) → Session, requires grants ⊆ ceiling.
THE RACK    the kernel's flat map of mounted namespaces.
```

**Read together, `Horizon` looks like it should be built from `GRANT` and `SCOPE`** — what may be
done and what may be named, which is exactly the pair Horizon's own definition spans. And `CEILING`
looks like the durable bound `Approach` reads to establish the initial Horizon, the way `Mount`
reads it today to authorize a sign-in without a live admin.

That reading is not made anywhere. Neither document cites the other:

```text
terminology.md    the authority model, whole, untouched by the seven.
primitives.md     Horizon, opaque, narrowing, with no stated composition.
```

Two live models of the same thing, neither superseded. This is the largest unreconciled seam in the
set and it is **HOLE H3**, shared with `../kernel/01-protocol.md` §4.
