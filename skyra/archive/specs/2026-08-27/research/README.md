# Research — 2026-08-27

**Status: RAW ORE.** Captured explanations of how real systems (CPUs, OSes) actually work, dropped here to
**MINE** for Skyra. These are NOT decisions — they are source material. Each doc marks its veins with
**TO MINE** callouts: the specific thing worth stealing and where it probably lands in the design.

Process: read a research doc → pull a **TO MINE** vein → develop it into a decision in the day's spec docs
(dispatch-and-decode, system-and-semantic, think-and-act, etc.) → the vein is spent.

## Index

```text
cpu-privilege-mechanism-vs-policy.md   how x86/kernel split privilege: CPU = MECHANISM (CPL bits, faults,
                                       the numeric ≤), kernel = POLICY (tables + syscall gates). Skyra
                                       un-welds x86's fused "ring 0 = trusted+privileged+kernel" into a
                                       TRUST axis (block) ⟂ AUTHORITY (capability object), and makes the
                                       mechanism/policy split two named parts: Verify (CTU) vs Mint (kernel).
                                       7 veins: no-opinion enforcer · transfer-not-ladder · kernel-chosen
                                       landing · trust-rooted-at-boot · capability-as-object (seL4/KeyKOS) ·
                                       IDT = fault-vector table · copy_from_user = BELIEVED/CHECKED.
```
