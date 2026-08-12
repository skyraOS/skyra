# WHAT — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHAT**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/What-outcomes.md`](../outcomes/What-outcomes.md).

**Scope:** the work itself — the expectation, and the observations bound to it.

Sections are numbered within this document and cited as **What E1**, **What E2**, …

---

## E1 · `uname(2)` is the call, and `struct utsname` is what it returns

One syscall, POSIX-mandated, present on every Unix. The syscall interface describes itself.

```c
// MacOSX.sdk/usr/include/sys/utsname.h
struct utsname {
    char sysname[_SYS_NAMELEN];   /* [XSI] Name of OS */
    char nodename[_SYS_NAMELEN];  /* [XSI] Name of this network node */
    char release[_SYS_NAMELEN];   /* [XSI] Release level */
    char version[_SYS_NAMELEN];   /* [XSI] Version level */
    char machine[_SYS_NAMELEN];   /* [XSI] Hardware type */
};
int uname(struct utsname *);
```

```console
$ uname -s   Darwin
$ uname -n   Michaels-MacBook-Pro-10.local
$ uname -r   24.6.0
$ uname -v   Darwin Kernel Version 24.6.0: Mon Jul 14 11:30:29 PDT 2025; root:xnu-11417.140.69~1/…
$ uname -m   arm64
```

Two of the five are not properties of the interface:

```text
nodename   the hostname. user-settable, DHCP-changeable, and it collides — see Where E3.
version    a build string that restates `release` plus a build timestamp. no independent axis.
```

*Verified on this machine.*

## E2 · Darwin's syscall table does not branch on architecture

```console
$ grep -c "SYS_" $SDK/usr/include/sys/syscall.h
461
$ grep -E "defined\(__i386__\)|defined\(__x86_64__\)|defined\(__arm" …/syscall.h
(no matches)
$ grep -oE "SYS_[a-z0-9_]*" …/syscall.h | grep -iE "i386|x86|arm|ldt|arch"
SYS_crossarch_trap
SYS_searchfs
```

461 numbers, zero architecture conditionals, and the only two arch-flavoured names are not
arch-conditional. **On Darwin the syscall set is the same on both architectures.**

*Verified on this machine, from the SDK header.*

## E3 · The same `release` ships for two architectures

macOS 15 / Darwin 24.x runs on both Intel and Apple Silicon Macs. So `sysname` + `release` does not
determine `machine`, and two machines reporting `Darwin 24.6.0` may have different instruction sets.

*NOT verified here — no Intel Mac available. A statement about which hardware macOS 15 supports.*

## E4 · On Linux the syscall SET itself differs by architecture

```text
arch_prctl                          x86_64 only
cacheflush                          ARM only
riscv_flush_icache · riscv_hwprobe  RISC-V only
```

So `machine` discriminates on Linux even though it does not on Darwin — which is why it cannot be
dropped from a key meant to be uniform across platforms.

*NOT verified here — no Linux machine available. Known from the Linux syscall tables.*

## E5 · What differs by ABI while the contract holds

```console
$ getconf PAGESIZE      16384        (arm64 macOS)
                         4096        (x86_64 macOS — same Darwin 24.6.0)
```

`read(fd, buf, count)` is identical on both: same signature, same return, same errno set.
`mmap`'s contract — *map at page granularity* — is also identical; the page size is a **value
queried at runtime** via `sysconf(_SC_PAGESIZE)`, not a difference in what the call does.

What differs is how the call is reached: calling convention, which registers carry arguments, and
the trap instruction (`svc` vs `syscall`).

*Page size on arm64 verified here; the x86_64 value is not.*

## E6 · What the triple does not certify

```text
SAME TRIPLE, DIFFERENT SET      a Linux kernel built without CONFIG_USERFAULTFD has no
                                userfaultfd, with the release string unchanged.

SAME TRIPLE, DIFFERENT          seccomp filters and container profiles restrict the set
AVAILABILITY                    per-process, not per-machine.
```

The triple says which interface to expect. It does not certify that every call in it will succeed.

*NOT verified here.*
