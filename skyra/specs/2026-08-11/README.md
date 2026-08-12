# 2026-08-11

**The set that revises `2026-08-09` against the first build.** Nothing here is built.

> **PRECEDENCE. Where two documents in this set disagree, the one with the later decision wins, and
> the superseded one carries a banner saying so. `terminology.md` is authoritative on words.**
>
> A document without a banner has not been checked against the ones after it — that is the failure
> this rule exists to stop, and it has already happened once.

Read `terminology.md` first. It retires two words — **ROOT** and **MOUNT-as-a-noun** — that appear
throughout `2026-08-09` and throughout the earlier documents here.

---

## semantics

| doc | owns | status |
|---|---|---|
| [terminology](semantics/terminology.md) | what each word means, and what it no longer means. | **REFERENCE** · authoritative on words |
| [open-until-commit](semantics/open-until-commit.md) | a possibility returns without closing. closure is the only durable write. | **PROPOSAL** |
| [kernel-revision](semantics/kernel-revision.md) | the surface collapse, two faces, the ocap convergence. | **PROPOSAL** · one stale line, marked |
| [interface-revision](semantics/interface-revision.md) | the kernel must hold the crossing without reading it. | **PROPOSAL** |
| [second-kind-of-credit](semantics/second-kind-of-credit.md) | credit on an observation across mounts. one decision, the rest parked. | **OPEN** |
| [roots-and-admission](semantics/roots-and-admission.md) | namespaces and admission. | **SUPERSEDED IN PART** — see its banner |
| [type-surface-proposals](semantics/type-surface-proposals.md) | the type surface cross-referenced against all of the above. | **REVIEW** |
| [divergence](semantics/divergence.md) | how far `skyra-dna-v5` is from this set. | **NOTE** · measured 08-11, pre-terminology |
| [conversation-runtime-mismatch](semantics/conversation-runtime-mismatch.md) | five mismatches, all now resolved toward the conversation. | **RESOLVED** |
| [questions-08-11-claude](semantics/questions-08-11-claude.md) | working notes. not a spec. | **NOTES** |

## structs and interfaces

The complete type surface of `skyra-dna-v5` as built, grouped by package — `01-ids` through
`09-trace`, plus a README. **Descriptive of the code, not of this set**, so it diverges from
`semantics/` by exactly the amount `divergence.md` measures.

`interface changes/` holds change proposals. **Those are the documents that get implemented**, so
they are the ones that must never be stale.

## What this set decided

```text
A POSSIBILITY RETURNS      commit hands an observation up and leaves it live. only closure
WITHOUT CLOSING            writes. mint-on-closure was named for the right event all along.

A NAMESPACE IS AN ENGRAM   durable, per principal, its children are what you can select. its
                           history is the principal's record of what they asked and kept —
                           which is where acceptance lands, and why the rootless commit closed.

MOUNT IS A KERNEL METHOD   attach a namespace to the rack. the only exported method, and it
                           replaces Ingress. the rack is FLAT and admin is a row on it.

THE CEILING IS DURABLE     written onto a namespace engram by whoever provisioned it, so mount
                           needs no live admin to authorize a sign-in.

A SELECT PROVISIONS        you hand a child access from the access you hold, and nothing widens.
                           default is nothing — least privilege by construction.

THERE IS NO CONTROL        drive is a select, accept is a commit, and seven of nine exported
PLANE                      kernel methods were already in the design under other names.

EVERY ENTITY HAS TWO       a selectable face and a privileged face, and the privileged face is
FACES                      never in any frame. this is object-capability, arrived at from a
                           credit requirement.
```

## Still open

```text
DEFAULT GRANTS ON A SELECT   nothing vs everything-I-hold. recommended: nothing.
PRODUCER ACROSS A DEEP COPY  and it is a distribution question, not only a correctness one.
WHAT A MEMORY PROVISION IS   "you may accumulate HERE" has no HERE.
WHICH OBSERVATIONS COME      discovery is "picking and choosing what carries over."
WITH A COPY
THE SECOND KIND OF CREDIT    parked deliberately. nothing needs the answer yet.
```
