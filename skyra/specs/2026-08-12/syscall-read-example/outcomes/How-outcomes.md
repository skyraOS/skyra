# HOW — Outcomes

**Status: MICHAEL ONLY.** Decisions about **HOW**, taken from [`../example.md`](../example.md), on
the evidence in [`../evidence/How-evidence.md`](../evidence/How-evidence.md).

**The question, from the engram's perspective:** *How am I executed?*

**Scope:** how am I executed — spawn, the program, and the interface available where I land.

---

## 1 · `sysname` · `release` · `machine` determine "what exists"

**Decided 2026-08-12.**

The syscall interface available on a machine is identified by three fields, and they are returned by
a single POSIX call — `uname(2)`, into `struct utsname`.

```text
sysname    Darwin      WHICH KERNEL
release    24.6.0      WHICH VERSION OF IT
machine    arm64       WHICH ABI
```

`nodename` and `version` are in the same struct and are not part of it.

```text
EVIDENCE   How E1   uname(2), struct utsname, and the two fields discarded
           How E2   Darwin's table does not branch on architecture
           How E3   the same release ships for two architectures, so machine is not derivable
           How E4   on Linux the set itself differs by architecture
           How E5   what differs by ABI while the contract holds
           How E6   what the triple does not certify
```

## 2 · What each host offers — `spawn(program) → handle`

**Decided 2026-08-12.**

Of everything a spawn call takes on any platform, exactly two elements are cleanly universal: **the
program to run**, and **a handle back** identifying what was created.

```text
spawn(program) → handle
```

Everything else is a concept that survives with a lossy or asymmetric implementation — arguments do
not round-trip as a vector on Windows, environment names differ in case sensitivity, the working
directory is not a POSIX spawn parameter at all, and extra file descriptors are not universal in any
form.

```text
EVIDENCE   How E8   the three declarations side by side, and Go's portable / not-portable line
```
