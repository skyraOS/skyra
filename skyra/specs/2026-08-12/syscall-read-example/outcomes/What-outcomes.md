# WHAT — Outcomes

**Status: MICHAEL ONLY.** Decisions about **WHAT**, taken from [`../example.md`](../example.md), on
the evidence in [`../evidence/What-evidence.md`](../evidence/What-evidence.md).

**Scope:** the work itself — the expectation, and the observations bound to it.

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
EVIDENCE   What E1   uname(2), struct utsname, and the two fields discarded
           What E2   Darwin's table does not branch on architecture
           What E3   the same release ships for two architectures, so machine is not derivable
           What E4   on Linux the set itself differs by architecture
           What E5   what differs by ABI while the contract holds
           What E6   what the triple does not certify
```
