# Roots, Namespaces, and Admission

**Date:** 2026-08-11
**Status: SUPERSEDED IN PART.** Revises what `2026-08-09` and `kernel-revision.md` call a root.
Nothing here is built.

> **SUPERSEDED, 2026-08-11 later, by [`terminology.md`](terminology.md), which is authoritative on
> words.** Two things this document is built on were retired after it:
>
> ```text
> ROOT      retired. what it named is a BRANCH — one ask, keyed by span, the transaction
>           boundary. a namespace is an ENGRAM, and a TREE is all live work under one.
> MOUNT     retired AS A NOUN. it survives as a KERNEL METHOD: attach a namespace to the rack.
> ```
>
> Sections marked **RETIRED** or **REVISED** carry the correction inline. Everything else stands.

> **A namespace is durable memory; access to it is not. The confusion this document exists to fix
> was calling both of them a root.**

---

## Why this exists

The set has been using *root* for a durable memory, and the build followed: `rootFor` scans for an
engram whose expectation matches the ask, and `kernel.Root` holds a `*graph.Engram`.

Three defects trace to that one identification.

```text
SELF-REDISCOVERY      a root's expectation IS the ask, so it is the best lexical match for its
                      own chain's need. by-index returned the root as the resolution of the
                      root's own need until the depth bound caught it.

INVERTED OWNERSHIP    Engram.History is "the record of that parent invoking it." an engram is a
                      record of INVOKING, never of BEING INVOKED. an engram whose expectation is
                      the incoming ask is the agent holding a record of being asked.

AN UNSCOPED SELECT    kernel-revision.md: "opening a root is the one select with no frame
                      bounding it." which is why questions.md §5 has no gate — there is nothing
                      to check against.
```

All three dissolve if a root is a **namespace** rather than a memory.

---

## The rule

Everything below compresses to three sentences:

> **A namespace cannot be invoked.**
> **A namespace can only be provisioned from a namespace, and boot provisions the first.**
> **When you ask another entity to do something you provision their access from the access given to
> you, and nothing can widen it.**

The second is stated as a **provenance** rather than a prohibition, and that is deliberate. *A
namespace cannot be reached from inside* would be a rule; saying where namespaces come from makes
unreachability a consequence — the only way to make one is somewhere no branch can stand. Safe by
incapacity, not by restraint.

The third is `world.md`'s *you cannot invoke your own grader*, one layer up.

*(An earlier version of the middle sentence said **a root can only be provisioned from a namespace**.
Roots are retired; provisioning runs namespace → namespace, with boot as the axiom.)*

## What a namespace is

```text
AN ENGRAM     an expectation + the history of invoking it.   SELECTED.
A NAMESPACE   an engram nobody can name.                     SIGNED INTO.
```

Nothing selects a namespace, nothing interprets it, and it originates nothing. It is where naming
*starts* — and it is durable, which is the half this document originally got wrong by calling the
whole thing a root.

**The code half-implements this already** and calls it something else:

```go
// Builtin names, present in every frame. discover is nameable everywhere — a
// builtin name, not something you must discover in order to discover.
Builtin map[string]*graph.Engram
```

That is one global name table, and it should be per-namespace: **ambient names are children every
namespace is given at boot.**

### The sign-in engram

A mount points at an engram, and that engram already carries `Engrams` and `History`. So **earned
structure accumulates as its children**, by the ordinary mechanism — the namespace needs no parallel
child map and no history of its own.

> **The sign-in engram is the first engram a principal sees, and everything they build hangs off
> it.**

Which also answers the 8-out-of-8 measurement without a rule: the outermost parent is an engram like
any other, so topology accumulates on it, and it persists between sign-ins the way a home directory
does. `rootFor`'s scan-by-expression disappears — you are signed in, so it is already yours.

---

## Profiles, mounts and resolution order — RETIRED

This section proposed `Mount`, `MountSource`, `Resolver`, a three-place resolution order, and
*mounts shadow earned entries*. **All of it is gone**, and the argument is in
[`terminology.md`](terminology.md):

```text
MOUNT-AS-A-NOUN WAS FIVE     a child engram placed by provisioning · ambient names placed at boot ·
EXISTING MECHANISMS          grants on the branch · entity reach · copy.

GRANTS-ON-THE-REFERENCE      EFFECTS HAPPEN AT LEAVES, so authority attached to a name at the top
WAS SPECIFICALLY WRONG       bounds nothing unless every intermediate re-narrows it. grants bound
                             the BRANCH, because a branch is what connects a leaf to an authority.

PATH HIJACKING DISSOLVED     one child map, no shadowing order. the collision moves to provisioning
                             time, where "who may write into my namespace" is already answered by
                             commit-is-a-request.
```

**`MOUNT` survives as a kernel method** — attach a namespace to the rack — which is a different
operation and is defined in `terminology.md`.


## Authority lives on the branch, never on the entity

An entity has **no owner**. Exactly-one-owner would make an interpreter exclusive — if I hold
authority over an AI, my sister cannot use it — which is absurd for a thing whose whole purpose is
being invoked by many parties.

> **The same entity, invoked from two branches, is two permission sets with no relationship between
> them.**

Which is the AWS shape again: a role is not owned by a user; a user *assumes* it with a policy
attached to that assumption.

**And it makes the confused deputy unrepresentable rather than fixed.** The canonical bug is C
invoking B to reach authority C does not hold. Here B has no ambient authority at all — it has
whatever the **branch that invoked it** carries. There is no "B's own authority" for C to borrow.

### What a branch carries — REVISED

This section said *a mount is three things*. With mounts retired, the three live on the branch:

```text
NAMES        resolved in frame.          what you can NAME.     the graph boundary.
GRANTS       read · write · spawn.       what you can DO.       the world boundary.
PROVISION    "you may accumulate HERE."  where writes LAND.     still has no shape — see Open.
```

The two permission categories fall on that seam and are why they are one vocabulary rather than two:

```text
DURABLE     where may you accumulate.   the graph boundary.
EPHEMERAL   read · write · spawn.       the world boundary.
```

### Commit is a request, not a bar

You may write into memory you do not own. **The authority is in the commit**, and a commit into
another entity's memory is a request to its owner.

```text
COVERED BY A GRANT   commit:<engram or subtree> was provisioned.   write.
NOT COVERED          ask the owner. deny → an ORDINARY FAILED INVOCATION, recorded on the
                     invoker's engram, and the invoker's selection learns from it.
```

The graph boundary and the world boundary now escalate through the same mechanism. They were two
authorizations with two characters; they become one with two alphabets.

**The build already holds the place for this.** `authorize.go`'s `ownership()` is a documented no-op:
*"held structurally rather than checked… if engrams ever become shareable, this stops being free and
must become real."* Shareable memory is exactly what provisioning creates.

**And denial does not stop an entity learning.** Learning *about* an entity lives on its parent's
engram, which it can never reach — `DOWN, NEVER UP`. Withholding commit authority stops an entity
accumulating its own selection history over its own children. It does not make it a pure function,
and it does not make its failures invisible. *It plays by the owner's rules.*

### Invoking a tool versus invoking a party — REVISED

This section said invoking an entity is a **crossing that switches provision**. **Within a runtime it
does not**, and that was the confused-deputy hole described as a feature.

```text
WITHIN A RUNTIME   authority NEVER MOVES. a branch runs under the grants and in the provision of
                   whatever provoked it, whether it names an engram or an entity.

ACROSS RUNTIMES    a real crossing, and not a mechanism — two branches, a local stub on each
                   side, each kernel enforcing its own. nothing is shared to widen.
```

> **Autonomy inside the provision. Accountability at the boundary.**

An entity can do what it likes within what it was provisioned, and still has to deliver back into the
shared space — where the owner's commit decides whether any of it becomes durable there.

### The index is shared; the tree is owned

Ownership happens on **selection**, not on existence. Discovery searches across what is visible;
citation is what makes a copy yours.

So a memory provision bounds **where your copies land**, not what you can see. Which leaves *here*
still needing a shape — region, quota, subtree — and it is the one object in this model that has
none.

## Revocation — REVISED

This section said *revocation is unmounting*. With mounts gone:

> **Revocation is provisioning the next tree with less.**

A name stays in your namespace — it is your memory and erasing it would be erasing what you learned —
and it simply cannot do anything you are no longer granted, because grants live on the branch and are
re-issued at every mount.

What survives unchanged is the split underneath: **granted reach is ephemeral and earned reach is
durable**, so append-only forbids deleting an engram and says nothing about what a branch may do with
it. Entity reach was already runtime state, so dropping an entity revokes it for new branches while
in-flight ones keep what they hold until they end — **leasing, arriving without anyone adding a
lease.**


## Entering is an ordinary scoped select

```text
BEFORE   ingress + drive = an UNSCOPED select. maximum privilege, and nothing to check against.
AFTER    you are MOUNTED into a namespace, and a select resolves against it. the same scope
         check as every other select, no exception, no new mechanism.
```

So `questions.md §5`'s missing gate is **provisioning**, and the maximum-privilege problem becomes
*do not provision much*. That is `world.md`'s own line arriving where it was needed: *a capability
granted **once, at admission**, being exercised.*

### And it separates system space from user space

```text
SYSTEM SPACE   ambient names. discover · create · introspect. children every namespace is
               given at boot.
USER SPACE     entries that accumulated. per principal.
```

Which turns a rule carried since `2026-08-06` into something structural:

> A system-space expectation may be entered from user space. It may never be the standing root.

**You cannot sign in to `/proc`.** You traverse into it from somewhere you can stand. The rule stops
needing to be stated.

It also gives the two ambient authorities `kernel-revision.md` flagged as an ocap smell a place to
live. `discover` nameable everywhere is `/bin` on every `PATH`; `here` is the axiom that you are
standing somewhere. Both stay ambient — but ambient **in a namespace**, which is inspectable and can
be provisioned differently per principal, rather than ambient in the language.

---

## The kernel does not sign anyone in

**Linux authentication is entirely userspace**, and the split is the design:

```text
USERSPACE   login · sshd · su · sudo · PAM · /etc/shadow
            READS THE CREDENTIAL. VERIFIES IT. DECIDES WHO YOU ARE.

KERNEL      setuid() · setgid() · struct cred · capabilities
            NEVER SEES A PASSWORD. does not know what one is.
```

`login` runs privileged, checks the hash itself, and only then calls `setuid()` to **drop** to the
authenticated user. The kernel's entire contribution is enforcing one rule:

> **Authority can only be dropped, never gained — except by something that already had it.**

That is monotonic narrowing, which is already the grant rule. The kernel is not the authenticator; it
is what makes the authenticator unable to cheat.

### So sign-in is an engram

Verifying identity is **judgement** — read a credential, compare it, decide. `NO SEMANTICS` forbids
the kernel doing that. And as an engram it competes, which is the payoff:

```text
by-password  ─┐
by-token      ├─  competing resolvers under one authentication need. history on each.
by-oauth      │   SELECTION MOVES BETWEEN THEM.
by-biometric ─┘
```

Same shape as `discover`'s resolvers and as the approval strategies in `kernel.md`. Nobody edits an
auth policy; a bad one is outcompeted.

### The kernel owns one check, and it is arithmetic

```text
ISSUING A PROFILE   the issuer must already hold AT LEAST the profile's authority.
```

**Forced rather than chosen.** Monotonic narrowing forbids widening, so whatever hands you a profile
must already contain it — which means admission is **privileged-and-drop**, exactly like `login`, and
not minimal-and-grant however much more modern that sounds. You cannot grant what you do not have.

It is also `kernel.md`'s confused-deputy fix arriving a third time: *a kernel-rooted chain runs under
the grants of the chain that provoked it, and can never exceed them.*

### The kernel never learns who anyone is

It holds a `PrincipalID` it does not interpret — **the way `uid` is a number the kernel compares.**
Root is `0` and the kernel has no idea what *root* means; it compares integers.

Stated because *the kernel handles sign-in* is the most natural wrong assumption available here.

### The admission bootstrap

The usual regress: to run the sign-in engram you need a root, and to get a root you need to sign in.
Answered the way `discover` answered it — something is given.

```text
THE ADMISSION NAMESPACE holds the union of what it may issue, and ONE child: the authentication
                        engram. it can authenticate and nothing else.
```

Deliberately the narrowest useful namespace in the system.

---

## Admission closes, and it is two chains

`getty` execs `login`, which execs your shell. The process persists and its program is replaced —
`exec`, not close-and-start. **Skyra does not need that**, and should not have it.

**Why the privileged stage must not survive.** If admission stayed alive as the parent of the
signed-in chain, then the most privileged thing in the system would be a live ancestor of every user
chain. The leak is not authority — grants narrow downward, so a child holds a subset regardless. The
leak is **position**:

> **A live admission ancestor is the grader of the session. The principal is supposed to be.**

Every return would flow up to admission, and admission's commit would be the acceptance. That
inverts the one relationship the whole pressure model rests on.

**So it is two chains, and the channel is the seam:**

```text
1  the admission chain runs under the admission profile.
2  it COMMITS AN OBSERVATION CARRYING A PROFILE. an ordinary commit. nothing new.
3  the admission chain CLOSES. records written, staged effects disposed, done.
4  the kernel MOUNTS that namespace onto the rack under the issued grants.
```

Nothing inside the model produces a root. The channel does, and the channel is already the one thing
outside the two forms. **No new mechanism, and no tail call.**

### Which answers the general question, narrowly

Chains already close — commit is closing. What admission needed was not a new ability to close but
the recognition that **a hand-off across an authority boundary is not one chain.**

**And `exec` should not be added for the general case.** An orchestrator that selects one child and
names its return looks like pure overhead — a frame, a record, a hop. But that hop *is* the hop
profile:

```text
{hop1, hop2, hop3}   validated locally AND at range.
```

Collapsing intermediaries deletes the measurement that the intermediary was used. The overhead is the
instrument.

---

## Rejected

```text
ROOT AS AN ENGRAM        it is durable memory versus permissioned access. the identification
                         produced self-rediscovery, inverted ownership, and an unscoped select.

A KERNEL SIGN-IN         verifying a credential is judgement, and the kernel has no semantics.
METHOD                   linux does not do it either; authentication is userspace and the kernel
                         only enforces that authority drops.

GRANTS IN A SEPARATE     the ACL model. it reintroduces "given that you can name it, may you use
TABLE                    it," which is the question object-capability exists to delete.

EARNED ENTRIES           RETIRED WITH MOUNTS. one child map, so there is no shadowing order to
SHADOWING MOUNTS         get wrong; the collision moves to provisioning time.

MINIMAL-AND-GRANT        monotonic narrowing forbids widening, so an issuer must already hold what
ADMISSION                it issues. privileged-and-drop is forced, not stylistic.

A PARALLEL CHILD MAP     a namespace IS an engram, so it already has Engrams and History.
ON THE NAMESPACE         holding its own duplicates the engram machinery one level up.

PROVISIONING REACHABLE   provisioning IS admission, and admission is the gate §5 is missing. an
FROM INSIDE A BRANCH     operation anything inside can reach is not a gate. Mount and Unmount are
                         KERNEL methods, called from outside.

AN EXPECTATION FIELD     someone will read the struct and say "that is an engram with a nil
LEFT NIL                 field." a namespace IS an engram; it simply has no expectation anyone
                         can name.

exec / TAIL CALLS        collapsing a pass-through intermediary deletes the hop that records it
                         was used. the overhead is the instrument.

ADMISSION AS THE         a live privileged ancestor is the grader of the session. the principal
PARENT OF THE SESSION    is supposed to be.
```

## Open

```text
IS A PRINCIPAL AN        asker-identity is closer to a session or an account than to an entity,
ENTITY                   and nothing in the set has that shape yet. an entity is "the thing you
                         want a single track record of"; a principal is "the thing you want a
                         single namespace for." possibly the same, possibly not.

CROSS-PRINCIPAL          nothing is shared, so two principals asking the same thing build
LEARNING                 separate structure and neither learns from the other. correct for
                         privacy, possibly very wasteful. world.md's "duplication is nearly free"
                         was argued about EXPECTATIONS, which never improve — histories do.

MAY YOU BE MOUNTED       two moments that must not merge: whether you may be mounted at all, then
AT ALL                   under what grants. the first cannot live in a grant, because it is what
                         hands you the grants.

DOES AN AGENT GET        a namespace is what an entity has when it can INITIATE; interpreters
A NAMESPACE              never need one. so a scheduled agent or a daemon gets one and it should
                         be thin — which is inspectable, because a thin namespace is a small
                         child map.

WHAT UNMOUNTING DOES     an in-flight branch holds what it holds until it ends. that is leasing,
MID-BRANCH               and the lease term is the branch — which may be an hour.
```
