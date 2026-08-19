# WHERE — Outcomes

**Status: MICHAEL ONLY.** Decisions about **WHERE**, taken from [`../example.md`](../example.md), on
the evidence in [`../evidence/Where-evidence.md`](../evidence/Where-evidence.md).

**The questions, from the engram's perspective:** *Where am I stored?* and *Where is my entity?*

**Scope:** where this engram durably lives, and where the entity that interprets it can be
reached from. Two questions — see decision 3.

---

## 1 · `IOPlatformSerialNumber` is fundamental in determining "where am I"

**Decided 2026-08-12.**

The walkthrough stopped at *"local" has no referent* — three candidate machines, and nothing in the
expectation, the observation, the member, the record or the registry distinguishing them. Answering
that requires a machine to be able to say which machine it is.

```text
EVIDENCE   E6   what identifies and what does not
           E2   the version nibbles — why the deterministic ones cannot drift
           E3   the two identifiers with collision scars in their current values
           E1   the raw reads
```

## 2 · Device identity is DERIVED — a hash over hardware fields read at boot

**Decided 2026-08-12.**

Device recognition happens before the kernel exists, and the implementation differs per OS. That is
fine. What the method must return is **uniqueness**, and the way to get it is to hash a combination
of fields read from the device at boot — with `IOPlatformSerialNumber` among them on macOS, per
[decision 1](#1--ioplatformserialnumber-is-fundamental-in-determining-where-am-i).

```text
DERIVED, NOT MINTED
   NO STATE TO LOSE       there is no file to go missing, so identity cannot be orphaned.
   DETERMINISTIC          re-derived identically on every boot. a fresh random would change
                          each time, and persisting one reintroduces the state file.
```

This is what `IOPlatformUUID` and `kern.uuid` already are — v5 and v3, both hashes of a hardware
input. The choice is which inputs, not whether to hash.

```text
WHAT IT RESTS ON     terminology.md: "RUNTIME — ONE PER DEVICE that can run one."
                     a hardware-derived hash cannot distinguish two runtimes on one machine, and
                     does not have to, because there is only ever one. THAT LINE IS LOAD-BEARING,
                     not descriptive.

THE FALLBACK         a device with no readable hardware fact has nothing to hash, and falls back
                     to mint-and-persist with its state-loss weakness. The per-OS layer differs
                     anyway, so this is one more branch rather than a new mechanism.
```

```text
EVIDENCE   Where E7   there is no single universal field, and five reasons it does not unify
           Where E2   IOPlatformUUID and kern.uuid are already hashes — v5 and v3
           Where E6   what identifies and what does not
```

## 3 · WHERE is two questions, and they stay distinct

**Decided 2026-08-12.**

```text
WHERE AM I STORED     which graph holds this engram.        DURABLE. one answer.
WHERE IS MY ENTITY    which host can reach the entity that  VOLATILE. plural — see decision 2.
                      will interpret me.
```

**Both are fundamental, and they are independent.** The walkthrough was built on them differing —
the runtime on a laptop, the model remote, the file on a server — and it is that independence that
made *"local"* have no referent at step 4.

Neither answers the other. An engram stored on one machine may be interpreted by an entity reachable
only from another, and nothing about where it is stored constrains where it can be run.

```text
EVIDENCE   example.md, step 4   "local" has no referent — three candidate machines, and nothing
                                distinguishing them
           Where E8             reach is plural, and belongs to hosts
```

## 4 · Hosts have endpoints

**Decided 2026-08-12.** Answers the second question of [decision 3](#3--where-is-two-questions-and-they-stay-distinct) — *where is my entity?*

An entity is reachable more than one way — Claude Opus 5 through the Anthropic API, through
OpenRouter, through a CLI invocation; an open model through a local inference endpoint. **Those
endpoints are not the entity's.**

```text
AN ENTITY   an identity. globally unique. it does not carry where it can be reached.
A HOST      holds the endpoints.
```

An entity carrying its own endpoints would be a declaration — *"a declaration is made by whoever
registered the entity, not by the entity"* — and it would go stale the moment anything moved.

```text
EVIDENCE   Where E8   the design keys on the entity and excludes the host, because reach is plural
```

## 5 · Entity resolution sits BELOW the top-level engram

**Decided 2026-08-12.**

```text
THE TOP LEVEL   establishes WHERE.
THE KERNEL      resolves WHICH ADDRESS.
```

An expectation must not name an endpoint. The reason is structural rather than tidy:

> `interfaces.md` — an expectation is *"immutable. Expectations **never improve** — authored once
> and re-run."*

**Reach inside an expectation strands it the day the endpoint changes.** An engram that says
*"ask Claude via `api.anthropic.com`"* is wrong the moment a second path exists, and it cannot be
edited — only replaced. Resolution below the expectation is what keeps expectations survivable.

**Resolvers are modeled as engrams even when deterministic.** Not because they all call a model —
because an engram accumulates history, earns credit, and is selected among. A hardcoded table would
be the one component that can never improve, sitting where the world changes fastest.

**Fallback is the learning signal, not error handling.** Each fallback is an observation: *this
address, this moment, failed; that one worked.* The mechanism generates its own evidence. And it
only learns where there are ALTERNATIVES — one address can never improve, four can, which is why
a human (phone · TV · email · Slack) is the interesting case.

```text
EVIDENCE   Where E9    address and transport, measured — and both ending in an fd
           Who E3      criterion 2 needs a failure and needs surviving it
```

## 6 · The regress stops at the kernel interface

**Decided 2026-08-12; revised 2026-08-13.** *"Something resolves you"* immediately raises **which
resolver**. The regress stops because `Reach` is a fixed method on the kernel interface. Calling it
does not first require resolving a resolver.

The current runtime contains the older precursor:

```go
// Builtin names, present in every frame. discover is nameable everywhere — a
// builtin name, not something you must discover in order to discover. That is
// what stops the regress.
Builtin map[string]*graph.Engram
```

`Builtin` is read in `frame.go` and nothing populates it. It reflects the old name-based mechanism;
the decided surface moves the entry point to the kernel interface. Reach may still invoke learnable
resolver behavior after entry.

## 7 · The kernel routes; the resolver ranks

**Decided 2026-08-12.** The kernel invoking a resolver is `Ask`/`Drive`-shaped and legitimate — it
routes to something that decides. The kernel **ordering addresses itself** is what would break
*no semantics*. The resolver picks; the kernel carries the answer and records it.

## 8 · `Invocation.Host` records what the resolver chose

**Decided 2026-08-12.**

```go
Host ids.HostID   // "RECORDED, NOT KEYED"
```

Written once, at `kernel/kernel.go:369`, as `k.Host.ID()`. **Read nowhere** — grep finds no reader.

Under decision 5 it stops being dead weight: it records the resolver's choice, and the resolver
reads its own history back. The field's existing comment turns out to have been right for a reason
nobody had yet — it is not part of the question, it is **the answer to a different one**.

## 9 · Address and transport are two columns

**Decided 2026-08-12.**

```text
                ADDRESS                        TRANSPORT
curl            /usr/bin/curl                  fork(2) + execve(59)
Claude          api.anthropic.com:443          socket(97) + connect(98)
mongod          /tmp/mongodb-27017.sock        socket(97) + connect(98)
libssl          /opt/…/libssl.3.dylib          open(5) + mmap(197)
```

**A name never tells you how to reach it — the namespace does.** Two of those names live in the
same filesystem namespace and resolve to different transports.

```text
EVIDENCE   Where E9   the four rows, measured
           Who E4     a path and a network address are both names in a namespace
```

## 10 · `Reach` and `Think` are kernel methods backed by learnable behavior

**Decided 2026-08-12. Amended 2026-08-13** — the second name was wrong. See *What changed* below.

```text
THINK      what behavior might address this need?
REACH      resolve and contact the entity that will interpret it
```

Both live on the kernel interface, which stops the regress
([decision 6](#6--the-regress-stops-at-the-kernel-interface)). Each method may open an engram for
the semantic portion of its work, allowing ranking and interpretation to accumulate history and
improve without making the method itself a name that must first be resolved.

**The verb is `reach`** — not `dispatch`, not `open`, not `dial`:

```text
Skyra reached entity y
Skyra reached for entity z and couldn't        → the error message writes itself
```

*Reached for and couldn't* is a **failure with a subject**, and under decision 5 it is also the
training signal. The log line and the observation are the same event.

```text
WHY NOT `dispatch`   `dispatch` is already a MOMENT in this design — world.md: "projection happens
                     AT DISPATCH, not at return." That stays true and is not in competition:
                     dispatch is WHEN, reach is WHAT YOU DO. See Where E10.

WHY NOT `open` ·     both collide with syscalls that mean something adjacent but different.
`connect` · `bind`   naming the abstraction after a syscall it is not would mislead whoever
                     reads the platform layer.
```

### What changed, and why

The original read **`DISCOVER — find engrams · entities · observations · hosts`**, argued from a
log vocabulary that included *"Skyra discovered entity x."* **Both the definition and the argument
were wrong, and Michael retracted the argument.**

```text
THE ARGUMENT FAILS   Skyra already met the entity — that is WHY it is in the graph. Finding
                     something already known is RECALL, not discovery.

THE DEFINITION       "find engrams · entities · observations · hosts" is the LOOKUP reading, and
CONTRADICTS THE      discover.md rejects it outright:
SPEC IT NAMED
                       "Discovery is not a lookup. A flat list is raw material. 'There is nothing
                        here that does X' is a CONSTRUCTED claim — it requires knowing what was
                        sought and judging that nothing met it. A function returns a list. ONLY A
                        POSSIBILITY CAN RETURN A DIAGNOSIS."
```

So the three separate:

```text
RECALL      a flat lookup over what is known. mechanical. a HELPER.
DISCOVER    a diagnosis, and it belongs on NOVELTY — "there is nothing here that does X" is the
            constructed claim, and it is the signal that something is new.
THINK       the surface name (../../continuation-language/primitives.md): search AND reason over what is known, then rank
            candidates. Recall is a helper inside it.
```

`discover` is not renamed to `think` — it is a **different operation** that `think` may reach, and
it is much larger than either. Which decomposition survives is open:
**`questions.md` §16**.

```text
EVIDENCE   Where E10   the design already places reachability BEFORE the possibility opens
           Where E9    address and transport, measured
```
