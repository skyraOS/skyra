# Outcomes

**Status: MICHAEL ONLY.** Learned from [`example.md`](example.md).

---

## 1 · `IOPlatformSerialNumber` is fundamental in determining "where am I"

**Decided 2026-08-12.**

The walkthrough stopped at *"local" has no referent* — three candidate machines, and nothing in the
expectation, the observation, the member, the record or the registry distinguishing them. Answering
that requires a machine to be able to say which machine it is.

Of everything a process can read on macOS, the identifiers separate into two tiers, and the evidence
is the UUID version nibble rather than recall:

```text
PER BOOT      kern.bootsessionuuid     v4  RANDOM — generated fresh every boot
PER MACHINE   IOPlatformSerialNumber       factory-issued. assigned before the machine had an OS.
              IOPlatformUUID           v5  SHA1 name-based — deterministic, read from IOKit
              kern.uuid                v3  MD5 name-based — deterministic
```

Everything else a process can read fails to identify:

```text
gethostname()         Michaels-MacBook-Pro-10.local   the -10 IS a collision artifact
ComputerName          Michael's MacBook Pro (2)       the (2) IS a collision artifact
uname()               nodename = the hostname; the rest is model and kernel version
hw.model / hw.machine MacBookPro18,1 / arm64          a product line, not a machine
MAC addresses         all 0a:… — locally administered, virtual, spoofable, and nine of them
```

**And no OS update changes it.** `IOPlatformSerialNumber` is read from IOKit's platform expert —
hardware, not the operating system. `IOPlatformUUID` and `kern.uuid` are deterministic hashes, so
they cannot change unless their input does.

```text
macOS EXPOSES NO PER-INSTALL IDENTIFIER   you get THIS BOOT or THIS HARDWARE, with nothing in
                                          between. Linux has /etc/machine-id for the middle; there
                                          is no equivalent here.
```
