# WHAT — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHAT**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/What-outcomes.md`](../outcomes/What-outcomes.md).

**Scope:** the work itself — the expectation, and the observations bound to it.

Sections are numbered within this document and cited as **What E1**, **What E2**, …

---

## E1 · The fields, and what actually provides them

`struct utsname` is POSIX and the fields are present on every Unix. **How you get them is not
uniform, and on Darwin it is not a syscall.**

```console
$ grep -i uname $SDK/usr/include/sys/syscall.h
(no matches — there is no SYS_uname in Darwin's table)

$ man -w uname     /usr/share/man/man1/uname.1
$ man 3 uname      UNAME(3)   Library Functions Manual      ← SECTION 3, not 2

$ grep -E "SYS_.*sysctl" …/syscall.h
#define SYS_sysctl        202
#define SYS_sysctlbyname  274
```

```text
DARWIN   uname is a LIBC FUNCTION over sysctl. the syscalls underneath are sysctl(2), and every
         field has a sysctl equivalent — kern.ostype · kern.osrelease · hw.machine.

LINUX    uname IS a syscall — #63 on x86_64, sys_newuname. man 2 uname.
```

*An earlier version of this section said "one syscall, POSIX-mandated" and cited `uname(2)`. That is
true on Linux and false on Darwin, which is the platform everything else here was measured on.*

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

## E7 · How stable the mechanism is

**The struct is frozen. The values are not, and their format is not standardized.**

```text
THE STRUCT      unchanged since SVr4/4.4BSD. five char arrays. it does not move.

FIELD WIDTH     _SYS_NAMELEN 256 on macOS · __NEW_UTS_LEN 64 on Linux.
                same struct, different widths. harmless — the binary form is never shared.

THE VALUES      change every release. that is what a version is.

THE FORMAT      no common grammar:
                  macOS     Darwin    24.6.0              arm64
                  Ubuntu    Linux     6.8.0-45-generic    aarch64
                  FreeBSD   FreeBSD   14.0-RELEASE        arm64
```

Two consequences:

```text
arm64 AND aarch64 ARE     the same ISA under two names. `sysname` already differs, so a key does
THE SAME ISA              not collide — but any normalization needs the mapping.

EQUALITY IS SAFE,         comparing two triples for SAMENESS is reliable. extracting "is this
PARSING IS NOT            newer than" is not, across platforms.
```

**And the values are self-reported.** On Linux, `setarch --uname-2.6` and `personality()` make
`uname` return whatever they are told to, and containers do this routinely. Adequate for a key that
history hangs on; not a security claim.

*Field widths and the Darwin values verified here. The Linux and FreeBSD values, and the spoofing
mechanisms, are not.*

## E8 · Which spawn parameters overlap across every OS

The same operation, as each platform actually declares it:

```c
// POSIX
int execve(const char *path, char *const argv[], char *const envp[]);
int posix_spawn(pid_t *pid, const char *path,
                const posix_spawn_file_actions_t *file_actions,
                const posix_spawnattr_t *attrp,
                char *const argv[], char *const envp[]);

// Win32
BOOL CreateProcessW(LPCWSTR lpApplicationName, LPWSTR lpCommandLine, …,
                    LPVOID lpEnvironment, LPCWSTR lpCurrentDirectory,
                    LPSTARTUPINFOW lpStartupInfo, LPPROCESS_INFORMATION lpProcessInformation);
```

Go's `os/exec` is the same problem solved by people who had to ship it on every OS, and it draws
the line explicitly:

```go
type Cmd struct {
    Path   string        // PORTABLE
    Args   []string
    Env    []string
    Dir    string
    Stdin  io.Reader
    Stdout io.Writer
    Stderr io.Writer

    ExtraFiles  []*os.File           // "ExtraFiles is not supported on Windows."
    SysProcAttr *syscall.SysProcAttr // "This is only used on <os>." — the per-OS escape hatch
}
```

Parameter by parameter:

```text
PROGRAM       POSIX path · Win32 lpApplicationName              ✓ CLEANLY UNIVERSAL

HANDLE BACK   POSIX pid_t · Win32 PROCESS_INFORMATION           ✓ CLEANLY UNIVERSAL

ARGUMENTS     POSIX char *const argv[]  — A VECTOR              ~ concept universal,
              Win32 lpCommandLine       — ONE STRING              REPRESENTATION IS NOT.
              the callee re-parses, and the rules differ between CRT versions and
              CommandLineToArgvW. A VECTOR DOES NOT ROUND-TRIP.

ENVIRONMENT   POSIX char *const envp[]  — KEY=VALUE array       ~ concept universal.
              Win32 a double-null block                           Windows names are
                                                                  CASE-INSENSITIVE; POSIX's
                                                                  are not.

WORKING DIR   POSIX NOT A PARAMETER — chdir between fork and    ~ asymmetric. implementable
                    exec. addchdir_np is literally _np,           everywhere, native nowhere
                    "non-portable"                                on POSIX.
              Win32 lpCurrentDirectory

3 STREAMS     POSIX fds 0/1/2, inherited                        ~ concept universal,
              Win32 STARTUPINFO.hStdInput/hStdOutput/hStdError    mechanism differs

EXTRA FDs     POSIX passable, and an fd IS a capability handle  ✗ NOT UNIVERSAL. Go excludes
              Win32 no equivalent                                 it by name.
```

*Go's struct and its portability comments verified on this machine. The Win32 signatures, the
argv parsing behaviour and the case-insensitivity are NOT verified here.*

## E9 · What "program" can be

```text
POSIX      a native binary  OR  a #! script            the KERNEL handles both
WINDOWS    a native .exe                                scripts need the interpreter named
                                                        explicitly, or a registry association
```

**On POSIX the kernel resolves the shebang itself — no shell is involved.** Demonstrated by
`fork()` + `execv()` directly, with nothing else in the process:

```console
$ file hello.py
Python script text executable, ASCII text          ← NOT a binary

$ head -c 2 hello.py | xxd
00000000: 2321                                     ← "#!"

$ python3 -c "
  pid = os.fork()
  if pid == 0: os.execv('hello.py', ['hello.py'])   # raw execve. no /bin/sh.
  os.waitpid(pid, 0)"
I am a SCRIPT, exec'd with no shell involved
```

The kernel reads the first two bytes, finds `#!`, and execs the interpreter named on that line
instead. `/usr/bin/env python3` in this case.

**And what fails shows the boundary from the other side:**

```console
os.execv('cd', …)                → No such file or directory   a shell BUILTIN. not a file.
os.execv('/bin/ls | grep x', …)  → No such file or directory   a shell EXPRESSION. the pipe
                                                                is syntax, not a path.
```

So a program is **a path, not a command line.** Anything carrying pipes, redirects, globs, `&&`,
variable expansion or builtins is shell language and requires a shell to be spawned to interpret it.

```text
SAFE EVERYWHERE   program = a path to a NATIVE EXECUTABLE
POSIX ONLY        a #! script also works, and the kernel resolves the interpreter
NOWHERE           a command line
```

**One path can satisfy two `machine` values.** Also observed here:

```console
$ file /bin/sh
Mach-O universal binary with 2 architectures:
  [x86_64: Mach-O 64-bit executable x86_64]
  [arm64e: Mach-O 64-bit executable arm64e]
```

A fat binary carries several architectures and the kernel selects the slice. Bears on **decision 1**:
`machine` identifies the ABI a process will run under, and a single program file may satisfy more
than one of them.

*The shebang execve, the magic bytes, the two failures and the universal binary are all verified on
this machine. The Windows behaviour is NOT verified here.*
