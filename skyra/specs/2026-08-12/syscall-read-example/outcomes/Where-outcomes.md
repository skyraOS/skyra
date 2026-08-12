# WHERE — Outcomes

**Status: MICHAEL ONLY.** Decisions about **WHERE**, taken from [`../example.md`](../example.md), on
the evidence in [`../evidence/Where-evidence.md`](../evidence/Where-evidence.md).

**The question, from the engram's perspective:** *Where am I executed?*

**Scope:** where am I executed — which machine, and how it is identified.

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
