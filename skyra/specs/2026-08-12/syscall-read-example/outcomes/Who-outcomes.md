# WHO — Outcomes

**Status: MICHAEL ONLY.** Decisions about **WHO**, taken from [`../example.md`](../example.md), on
the evidence in [`../evidence/Who-evidence.md`](../evidence/Who-evidence.md).

**The question, from the engram's perspective:** *Who am I?*

**Scope:** who is interpreting me — the entity, and what makes it one identity.

---

## 1 · WHO is the interpreter's identity

**Decided 2026-08-12.**

The thing that runs is a **pairing** — an engram opened by an entity — and it speaks with the
entity's identity. Handed an expectation, Claude Opus 5 is still Claude Opus 5; running a script,
python3.14 is still python3.14.

```text
WHO    the interpreter's identity
```

**Identity, not activity.** What is being done is [WHAT](What-outcomes.md); who is doing it is the
entity, and it does not change because the expectation did.

```text
EVIDENCE   Who E1   the design's own definition, and the field that is KEYED
```

## 2 · What an entity is — the five

**Decided 2026-08-12.**

```text
1  UNIQUE                 globally.
2  INDEPENDENT EXISTENCE  it can fail without taking you with it.
3  TAKES AN EXPRESSION    polymorphic, and declares nothing about which ones it accepts.
4  RETURNS SOMETHING      evidence comes back.
5  THROUGH THE HISTORY WITH ITS OBSERVER/PARENT.
```

**Nothing in the five is declared.** Each is a property something either has or does not, found
out by trying — the same posture that dissolved language tags, capacity, modalities and host lists
off the entity.

```text
2 IS THE BOUNDARY      an entity has its OWN FAILURE DOMAIN — its own process, or its own
                       machine. It can crash while you keep running, and you learn about it as
                       an error rather than as your own death.

                       Reaching it hands back a HANDLE TO SOMETHING THAT IS NOT YOU. A linked
                       library hands back a pointer, which is an offset INTO you: same pid,
                       same address space, same crash. Not an entity.

                       This criterion also absorbs the earlier "exists somewhere" and "could
                       exist somewhere", which were one criterion counted twice.

                       It was first written as "has at least 1 unique address" and that was
                       WRONG — see Who E3. Every one of these has a name in a namespace,
                       libraries included, so naming discriminates nothing.

5 IS RELATIONAL        an entity's continuity with an observer/parent lives in the history they
                       accumulate together. It is not a private declaration carried by the entity.
```

```text
EVIDENCE   Who E2   the five against real entities, and where the boundary falls
           Who E3   why the criterion is not "has an address" — a library has one
```

## 3 · You reach an entity through an entity

**Decided 2026-08-12.**

```text
runtime ──execve──▶ curl ──HTTPS──▶ Claude
                    entity          entity
```

`curl` satisfies all five: uniquely versioned, its own process and failure domain, takes an argv it
parses and acts on, returns bytes, and is known through the history accumulated with its
observer/parent. So reaching Claude went **through** an entity.

**This makes "endpoint" recursive rather than special.** Reaching anything may itself pass through
something reachable, and it holds one level further down — `execve` is reached through the kernel,
which is the layer trusted by review rather than by history.

```text
THE COST     composition puts the middle entity in the TRUST PATH. curl can misreport what
             Claude said; the history blames the outer name either way.

WHAT IT      exec.CommandContext(ctx, "codex", args...) is ONE EntityID covering two entities —
PREDICTS     the CLI binary and the model it reaches. world.md's "start coarse and split on
             evidence" makes that the right starting state; this says what the split would be into.
```

## 4 · What first contact can establish

**Decided 2026-08-12.**

```text
1  UNIQUE                 known at first contact
3  TAKES AN EXPRESSION    known at first contact — it parses or it does not
4  RETURNS SOMETHING      known at first contact
5  THROUGH THE HISTORY WITH ITS OBSERVER/PARENT
                           established as history accumulates between them

2  INDEPENDENT EXISTENCE  requires a FAILURE, and requires surviving it.
```

You cannot establish criterion 2 by asking, or by one successful call. Something has to break while
you are watching — which is the property the whole system already runs on.

**And it is observable, which makes co-location learnable rather than declared:**

```text
two entities always fail together   →  they share a failure domain
two entities fail independently     →  they do not
```

`world.md`'s *"start coarse and **split on evidence**"* arriving from a new direction — here the
evidence is failure correlation rather than behavioural difference.

```text
CAVEAT   failure domains are NESTED, not binary — thread · process · machine · network.
         mongod is independent of your process and not of your machine. "Independent" is
         always relative to a level, and this criterion picks the process level.

         A device with no MMU has no local isolation at all, so on such a device there are
         NO LOCAL ENTITIES — everything local is a library, and the only entities are across
         the network.
```
