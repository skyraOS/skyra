# Roots, Namespaces, and Admission

**Date:** 2026-08-11
**Status: PROPOSAL.** Revises what `2026-08-09` and `kernel-revision.md` call a root. Nothing here is
built.

> **A root is not an engram. An engram is durable memory; a root is permissioned access.**
>
> **It is `/` — a place to resolve names from, and permission to resolve them.**

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

## What a root is

```text
AN ENGRAM   an expectation + the history of invoking it.   SELECTED.
A ROOT      a namespace + an identity.                     SIGNED INTO.
```

Nothing selects a root, nothing interprets it, and it originates nothing. It is where naming
*starts*.

**The code half-implements this already** and calls it something else:

```go
// Builtin names, present in every frame. discover is nameable everywhere — a
// builtin name, not something you must discover in order to discover.
Builtin map[string]*graph.Engram
```

That is a mount table. It exists, it is per-kernel, and it should be per-principal.

### The sign-in engram

A mount points at an engram, and that engram already carries `Engrams` and `History`. So **earned
structure accumulates as its children**, by the ordinary mechanism — the namespace needs no parallel
child map and no history of its own.

> **The sign-in engram is the first engram a principal sees, and everything they build hangs off
> it.**

Which also answers the 8-out-of-8 measurement without a rule: the outermost parent is an engram like
any other, so topology accumulates on it, and it persists between sign-ins the way a home directory
does. `rootFor`'s scan-by-expression disappears — reuse is a mount, not a match.

---

## Profiles and mounts

```text
A PROFILE   WHO IS THIS.       an identity context. an entity may hold several.
A MOUNT     WHAT THEY CAN DO.  a name bound to reach, carrying the authority to use it.
```

The AWS analogy is exact rather than loose: one identity, several named profiles, one selected per
session. The same human as themselves and as an admin. `sudo`, `assume-role`.

```go
// A Profile is a named authority context. Sign-in picks one; the ENTITY does not
// change.
type Profile struct {
    Name   string
    Mounts map[string]*Mount
}

// A Mount binds a name to reach the principal did not earn, WITH the authority to
// use it.
//
// RUNTIME CONFIGURATION, not graph structure — which is what makes unmounting
// possible at all. world.md: configuration is "volatile, mutable, rebuildable, and
// it belongs to the runtime."
type Mount struct {
    Name   string
    Engram *graph.Engram   // an ordinary engram. children and history come free.
    Grants []Grant
    Source MountSource     // Ambient · Granted · Inherited
}

// Resolver is what a frame consumes. No way to reach Mounts or Grants — the same
// shape as the Index interface the runtime entity already holds, and for the same
// reason: the write path should be ABSENT, not restrained.
type Resolver interface {
    Resolve(name string) (*graph.Engram, bool)
    Names() []string
}
```

### Grants belong on the mount, and that is a fork

```text
ACL     a name, and SEPARATELY a table of what may be done with it.
OCAP    THE REFERENCE CARRIES THE AUTHORITY. holding it is being permitted.
```

`grammar.md` already states the second nearly word for word — *capability is name resolution; you can
select what you can name; you can name what you were given* — and `kernel-revision.md` records the
convergence on object-capability explicitly.

**Splitting grants off the binding is the move back to ACLs**, and it reintroduces the question ocap
exists to delete: *given that you can name it, may you use it?* Mounting `write` with
`fs:write:/work/**` is attenuating a specific reference, which is the ocap primitive.

---

## Resolution order

```text
1  DELEGATED   this invocation only. reach handed back by something I CHOSE to invoke.
2  MOUNTS      granted.
3  ENGRAMS     earned.
```

> **Mounts shadow earned entries, never the reverse.**

If an accumulated engram named `write` could shadow a mounted `write`, a principal authors its own
`write` and everything provisioned for it silently routes through structure it wrote. That is **PATH
hijacking** — the best-known failure of exactly this design — and the ordering is the whole fix.

Delegation sitting above both is safe because delegated reach came from something the possibility
chose to invoke, and it dies with the invocation. That is `grammar.md`'s delegation-versus-injection
line, unchanged.

---

## Revocation is unmounting

`kernel-revision.md` lists revocation as the one hole with no home. The split closes most of it.

```text
GRANTED REACH   a mount. RUNTIME CONFIGURATION. removable, because the binding was never durable.
EARNED REACH    a child formed on citation. NOT removable — and revoking it would be erasing the
                principal's own memory, which is not what anyone wanted.
```

Append-only forbids deleting an **engram**. It says nothing about a **binding**.

And entity reach was already runtime — `k.Reg.Known()` flowing into `p.Reach` — so dropping an entity
from the registry revokes it for new chains while in-flight chains keep what they hold until they
end. **That is leasing, and it arrives without anyone adding a lease.**

---

## Ingress is an ordinary scoped select

```text
BEFORE   ingress + drive = an UNSCOPED select. maximum privilege, and nothing to check against.
AFTER    ingress = a select resolved against the root's namespace. the same scope check as
         every other select, no exception, no new mechanism.
```

So `questions.md §5`'s missing gate is **provisioning**, and the maximum-privilege problem becomes
*do not mount much*. That is `world.md`'s own line arriving where it was needed: *a capability
granted **once, at admission**, being exercised.*

### And it separates system space from user space

```text
SYSTEM SPACE   ambient mounts. discover · create · introspect. present in every namespace.
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
THE ADMISSION PROFILE   holds the union of what it may issue, and ONE mount: the
                        authentication engram. it can authenticate and nothing else.
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
4  the channel opens A NEW ROOT under that profile. it has no parent, because roots do not.
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

EARNED ENTRIES           PATH hijacking. a principal authors `write` and everything provisioned
SHADOWING MOUNTS         for it routes through structure it wrote.

MINIMAL-AND-GRANT        monotonic narrowing forbids widening, so an issuer must already hold what
ADMISSION                it issues. privileged-and-drop is forced, not stylistic.

A PARALLEL CHILD MAP     the mounted engram already has Engrams and History. a namespace holding
ON THE NAMESPACE         its own duplicates the engram machinery one level up.

Mount() / Unmount() AS   provisioning IS admission, and admission is the gate §5 is missing. an
NAMESPACE OPERATIONS     operation anything inside can reach is not a gate.

AN EXPECTATION FIELD     someone will read the struct and say "that is an engram with a nil
LEFT NIL ON A ROOT       field." role is the distinction; do not leave the field inviting the
                         question.

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

MAY YOU SIGN IN AT ALL   two moments that must not merge: whether you may open a root, then as
                         whom. the first cannot live in a mount, because it is what hands you the
                         mount table.

DOES AN AGENT GET        a root is what an entity has when it can INITIATE; interpreters never
A ROOT                   need one. so a scheduled agent or a daemon gets one and it should be
                         thin — which is now inspectable, because a thin profile is a small
                         mount table.

WHAT UNMOUNTING DOES     an in-flight chain holds what it holds until it ends. that is leasing,
MID-CHAIN                and the lease term is the chain — which may be an hour.
```
