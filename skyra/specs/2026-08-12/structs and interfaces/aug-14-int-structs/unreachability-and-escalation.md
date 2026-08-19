# Unreachability, and who acts on it

**Date:** 2026-08-15
**Status: CONVERSATION RECORD. UNRATIFIED.** Fourth session. Not checked against the code beyond
the citations given.

> **Failure kind 1 had a boundary as its owner and no party who could act. The namespace is that
> party. And the thinking behaviors that try alternatives live in a namespace of their own — the
> kernel's, which is what `Think` is.**

---

## 0 · Where this came from

`creator-and-compiler.md` §9 ended on address stability: a `WorldRef` is only as stable as the thing
it points at, and a URL that silently repoints is unknowability one level up. Michael: *"address
stability is the same thing the runtime already has. You try it, it sucks, you try another one."*

Correct — and most of it is already built.

---

## 1 · Retry-with-fallback is already below the Reach boundary

`../../continuation-language/primitives.md` §4:

> *"The kernel coordinates the attempt, while hidden composable behavior may interpret prior reach
> experience. **A failed endpoint and a working fallback remain causal observations from which that
> behavior can improve.**"*

And `think-interface-proposal.md` already lists `think-reach-entity` among Think's hidden behaviors.

```text
TRY IT, IT SUCKS, TRY ANOTHER   already the model. hidden, below Reach, and already the material
                                a thinking behavior learns from.

"ERRORS ROLL UP"                already true in the weak sense that matters: BEYOND CURRENT REACH
                                rises as a BOUNDED IMPOSSIBILITY OBSERVATION, not an Error object,
                                and the parent's parser judges it like any return. P5 and P15 hold.
```

So neither option raised in session is new. What is new is what happens when both are exhausted.

---

## 2 · The gap — kind 1 has no actor

From `creator-and-compiler.md` §4:

```text
1  UNREALIZABLE DECISION   owner: REACH        ← A BOUNDARY. not a party. it cannot decide.
2  FAILURE TO COMPILE      owner: THE CHILD    ← an actor. the creator pops.
3  UNSATISFIED RETURN      owner: THE PARENT   ← an actor. the parent wakes.
```

Kinds 2 and 3 name someone who can do something. Kind 1 names the boundary that **detects** it and
nobody who **decides** about it. The table looked complete because "owner: Reach" reads like the
other two rows, and it is a different kind of entry.

---

## 3 · The namespace is kind 1's actor

**DECIDED IN SESSION.** When Reach is exhausted and no thinking behavior can propose an alternative,
the namespace that opened the chain is shown that the chain cannot complete.

The argument for it being the namespace rather than a parent:

```text
AN UNREACHABLE ADDRESS IS NOT A CONTRACT FAILURE.
    no parent's contract is wrong. the composition was valid. THE WORLD MOVED.
    so no parent has standing to repair it — there is nothing at any parent to fix.

THE ONLY PARTY WITH STANDING owns the chain, because the only remaining question is
    DO YOU STILL WANT THIS, GIVEN IT CANNOT BE REACHED?
```

That question is not answerable at any interior position. It is answerable by whoever asked.

This also completes the ownership table by giving every failure a party:

```text
1  UNREALIZABLE DECISION   detected at REACH   →  acted on by THE NAMESPACE
2  FAILURE TO COMPILE      detected on compile →  acted on by THE CREATOR
3  UNSATISFIED RETURN      detected by parser  →  acted on by THE PARENT
```

---

## 4 · It is a fourth rung on the existing ladder

Not a new mechanism. The same escalation shape as `primitives-as-engrams.md` §8 and
`creator-and-compiler.md` §4, extended by one:

```text
ADDRESS FAILS            Reach tries alternatives.              hidden. nobody wakes.
ALL ALTERNATIVES FAIL    think-reach-entity proposes another.   still hidden.
NOTHING PROPOSED         BEYOND CURRENT REACH rises.            the parent's parser judges it.
NO PARENT CAN CONTINUE   the namespace is shown.                it owns the chain.
```

Each rung costs more than the one before, and the expensive rung is reached only when the cheap ones
are exhausted. Same discipline as retry → wake → escalate-to-creator.

---

## 5 · `Think` is the kernel's namespace

**DECIDED IN SESSION.** Michael: *"by kernel I mean the Think interface, because that is the kernel
namespaces."*

This answers something previously unstated: **where Think's hidden behaviors live.**

```text
primitives.md   Think's hidden behaviors "accumulate their own histories and compete."
                ↓
ACCUMULATING HISTORY means they are ENGRAMS.
                ↓
AN ENGRAM LIVES IN EXACTLY ONE PARENT'S MAP.
                ↓
THAT MAP IS A NAMESPACE — and it is the KERNEL'S.
```

```text
think                      the kernel's namespace
├── think-reach-entity
├── think-recognize-observation
├── think-discover-novelty
└── further behaviors which earn their place through use
```

`../../semantics/terminology.md` — *"NAMESPACE — durable. AN ENGRAM. a principal's top engram.
children are what they can select"* — fits exactly: the hidden behaviors are what Think can select.

### Consequence: the kernel is a principal

Not stated in session, and it follows directly:

> *"PRINCIPAL — WHO ASKS. holds a namespace. an entity that can INITIATE."*

The kernel drives, so it initiates. If it holds a namespace, it is a principal, and its namespace
has a ceiling — written by boot, exactly as admin's is.

```text
WHICH WOULD MEAN   the kernel's own thinking is BOUNDED. think-reach-entity cannot reach past what
                   the kernel's namespace permits.
```

That is a stronger claim than anything currently written and it seems right: it would be odd for
the one namespace whose behaviors run on every invocation to be the only unbounded one. But it is
this document's inference, not a session decision. See §7 edge 1.

---

## 6 · The constraint on what the namespace is shown

**Easy to violate at exactly this rung, and it would undo P12.**

`primitives.md` §4:

> *"Endpoint-specific evidence remains below the Reach boundary, while its caller receives the
> conclusion at the level where it can make another decision."*
> *"Addresses and transport failures do not escape upward as application vocabulary."*

```text
THE NAMESPACE SEES        this chain cannot complete. which composition. which need went unmet.
THE NAMESPACE DOES NOT    DNS failed for api.foo.com · connection refused · TLS expired ·
SEE                       which endpoints were attempted in what order.
```

Endpoint detail stays beneath Reach, where the learning behavior uses it. If it escapes here, the
escalation rung becomes the leak the whole address-behind-the-boundary rule exists to prevent — and
it would leak at the one place a human is looking, which is the worst place for it.

```text
TENSION, UNRESOLVED   a human shown "this cannot complete" with no detail may be unable to act,
                      which defeats the point of showing them. whether there is a diagnostic
                      channel that is NOT the causal record — reflection, per age.md's
                      "a pass nobody has specified" — is open.
```

---

## 7 · Open edges

```text
1  IS THE KERNEL A PRINCIPAL?      §5's inference. if yes, its namespace has a boot-written
                                   ceiling and its thinking is bounded. if no, Think's behaviors
                                   live in a namespace with no principal, which no current
                                   definition allows.

2  WHAT DOES "WORK WITH THE        Michael: "the namespace can work with the kernel to resolve the
   KERNEL TO RESOLVE" MEAN?        chain." whether that is an ordinary Think/Decide cycle inside
                                   the namespace, or a distinct repair interaction, is
                                   unspecified. the ordinary cycle costs nothing new and is
                                   probably it.

3  IS THE CHAIN SUSPENDED?         does the chain wait while the namespace decides, or is it dead
                                   and the namespace opens a new one? primitives-and-continuation-
                                   summary.md already lists "how suspension and upward resumption
                                   are represented causally" as open; this is the same question
                                   arriving from a new direction.

4  WHAT IF THE NAMESPACE IS NOT    a chain running unattended has a namespace that may not be
   MOUNTED?                        mounted — the rack holds who is signed in RIGHT NOW. there is
                                   then no one to show. whether such a chain blocks, dies, or
                                   parks its escalation for the next sign-in is unaddressed, and
                                   it is the case that matters most for long-running work.

5  DOES THE ADDRESS STAY KEYED?    the question that started this: a URL that silently repoints is
                                   unknowable. retry handles unreachability; it does not handle an
                                   address that resolves successfully to something DIFFERENT.
                                   NOT ANSWERED by anything here.
```

---

## 8 · Provenance

```text
CITED FROM SOURCE   Reach's fallback-and-learn model · think-reach-entity as a hidden behavior ·
                    BEYOND CURRENT REACH as a bounded observation · the endpoint-evidence rule ·
                    NAMESPACE and PRINCIPAL definitions · "hidden behaviors accumulate their own
                    histories."

DECIDED IN SESSION  the namespace is kind 1's actor · Think is the kernel's namespace.

PROPOSED, NOT       kind 1's missing actor as the gap being filled · the four-rung ladder ·
RULED               the kernel-is-a-principal consequence · the constraint in §6 ·
                    everything in §7.
```
