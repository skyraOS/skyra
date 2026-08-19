# Kubernetes State and Recovery

**Date:** 2026-08-12  
**Status: FUTURE RESEARCH.** Not a Skyra proposal or decision. This is a subject to investigate
before deciding whether Skyra needs any durable representation of unfinished execution.

---

## Why this is worth studying

Kubernetes survives worker failure by separating the durable description of what should exist from
the processes currently realizing it. Its control plane stores cluster state in a replicated
database and reconciliation loops replace execution that disappears.

Skyra currently takes a different position:

```text
KUBERNETES    persist desired state; reconstruct disposable execution after failure.

SKYRA         keep a branch ephemeral until closure; a crash leaves no accepted memory,
              partial record, or materialized effect.
```

The contrast is useful because `open-until-commit` makes a long branch behave like a long-running
transaction. It gives strong abandonment semantics and no durable progress before closure. A kernel
or host failure can therefore erase an hour of valid but unfinished work.

The research question is not *should Skyra become Kubernetes?* It is:

> What is the minimum state required to reconstruct unfinished execution after a kernel or host
> failure without allowing provisional work to become memory, evidence, credit, or discovery input?

---

## Kubernetes concepts to examine

```text
ETCD AND QUORUM          which system facts require consensus; what happens when a majority is
                         unavailable; snapshot and disaster-recovery semantics.

DESIRED / OBSERVED       durable declarations versus reports about what is running now.

RECONCILIATION           controllers repeatedly moving observed execution toward declared state.

LEASES AND HEARTBEATS    distinguishing a dead host from a slow or partitioned one.

OBJECT IDENTITY          replacement of an instance versus continuation of the same instance.

SERVICES AND ENDPOINTS   stable logical names over changing physical locations.

PERSISTENT VOLUMES       durable data whose lifetime is independent of the process using it.

STATEFUL WORKLOADS       what Kubernetes reconstructs, what the application must reconstruct,
                         and where the boundary lies.
```

---

## Skyra questions this may sharpen

### 1 · Is graph durability the only kind of durability?

The graph is epistemic: it records what happened and what was accepted. Recovery may require an
operational record with a different subject and different readers.

A possible distinction to investigate, not adopt:

```text
GRAPH       accepted history. selectable, comparable, and reachable by the credit walk.

JOURNAL     opaque continuation material. not selectable, not citable, not evidence, and never
            searched by discovery. exists only to reconstruct live topology after failure.

KERNEL      what is currently happening. reconstructed from the journal, if one exists.
```

Adding a journal would retract the current claim that crash semantics collapse to nothing. The
benefit must be demonstrated before introducing it.

### 2 · What identifies continued work?

If a possibility resumes on another host, is it the same possibility or a replacement?

```text
SAME IDENTITY       preserves the causal line, but requires enough state to prove continuity.

NEW IDENTITY        records a retry or replacement, but may duplicate effects or observations.
```

Span, branch, possibility, port, and entity identities may have different answers.

### 3 · What can be replayed safely?

Reassembly can reconstruct what an entity was shown. It does not prove that repeating `Step` is
safe. An entity may have contacted the world before the kernel learned whether the crossing
completed.

Research needs to distinguish:

```text
REPLAYABLE TRANSFORMATION    assemble · project · parse.
AT-MOST-ONCE EFFECT          may have happened even if acknowledgement was lost.
IDEMPOTENT EFFECT            safe to request again under the same operation identity.
COMPENSATABLE EFFECT         can be reversed by an explicit later action.
```

### 4 · How do ports survive kernel failure?

If kernels manage ports, recovery includes communication topology rather than only computation.

Questions:

- Is a port strictly local to one kernel lifetime?
- Does recovery mint a replacement port or restore the same logical port?
- How does the remote kernel distinguish resume from a duplicate connection?
- Which side owns the operation identity used for deduplication?
- What happens when each side believes the other failed?
- Can a branch resume when its remote peer has already closed its corresponding branch?

### 5 · Does recovery require consensus?

Skyra proposes many sovereign kernels connected only by explicit Reach operations. Reach can
establish recorded relationships and routing observations; it does not establish one authoritative
global ordering.

Determine which facts, if any, require stronger agreement:

```text
LIKELY LOCAL              live possibilities · branch drafts · local ports · local grants.

POSSIBLY AGREED           ownership of a resumed operation · effect deduplication · hand-off of a
                          port between hosts.

MUST NOT BECOME GLOBAL    graphs · namespaces · observations · histories · authority.
```

The answer may be per-crossing agreement between two kernels rather than system-wide consensus.

---

## Experiments worth building

1. Kill a kernel halfway through a multi-step branch and measure exactly what is lost.
2. Reassemble a possibility on a new host and compare its next emission with an uninterrupted run.
3. Kill the sender immediately before and after a remote effect, then look for duplication.
4. Partition two kernels while a port is active and let each independently decide the peer failed.
5. Restore from an intentionally stale continuation record and identify which invariants break.
6. Compare three policies on the same long task:
   - no recovery;
   - opaque operational journal;
   - durable desired-state reconciliation.
7. Verify that no recovery artifact can be selected, cited, discovered, or counted as acceptance.

---

## Admission rule

Do not add a database, consensus system, checkpoint, journal, controller loop, or durable
possibility merely because Kubernetes has one.

Admit a mechanism only after an experiment demonstrates that an important Skyra operation cannot
recover acceptably without it. Preserve the separation between:

```text
continuing unfinished work
recording accepted knowledge
materializing effects
discovering capabilities
```

Those may share implementation machinery eventually. They are not the same semantic event.
