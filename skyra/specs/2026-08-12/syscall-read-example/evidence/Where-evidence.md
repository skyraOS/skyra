# WHERE — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHERE**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/Where-outcomes.md`](../outcomes/Where-outcomes.md).

**The questions, from the engram's perspective:** *Where am I stored?* and *Where is my entity?*

**Scope:** where this engram durably lives, and where the entity that interprets it can be
reached from.

Sections are numbered within this document and cited as **Where E1**, **Where E2**, …

---

## E1 · What a process can read to identify the machine

```console
$ hostname
Michaels-MacBook-Pro-10.local
$ scutil --get ComputerName
Michael's MacBook Pro (2)
$ scutil --get LocalHostName
Michaels-MacBook-Pro-10

$ uname -a
Darwin Michaels-MacBook-Pro-10.local 24.6.0 … RELEASE_ARM64_T6000 arm64

$ ioreg -rd1 -c IOPlatformExpertDevice
"IOPlatformSerialNumber" = "QJ5YCY7YL7"
"IOPlatformUUID"         = "7F8CC55C-9AD1-51F9-9474-E91F4C0D3076"
"model"                  = <"MacBookPro18,1">

$ sysctl kern.uuid kern.bootsessionuuid hw.model hw.machine
kern.uuid:            D013020A-F19B-3730-8283-F7E47107D27A
kern.bootsessionuuid: 3CB5B5ED-961E-432E-877C-41FB1C7E2335
hw.model:             MacBookPro18,1
hw.machine:           arm64

$ ifconfig | grep ether
ether 0a:57:88:21:59:55        anpi2
ether 0a:57:88:21:59:54        anpi1
…nine interfaces, all with the 0a: prefix
```

## E2 · The UUID version nibbles

The 13th hex digit of a UUID is its version. This is a fact about the encoding, readable without
running anything twice.

```text
kern.bootsessionuuid   3CB5B5ED-961E-432E-…    v4   RANDOM
kern.uuid              D013020A-F19B-3730-…    v3   MD5 name-based
IOPlatformUUID         7F8CC55C-9AD1-51F9-…    v5   SHA1 name-based
```

```text
A v4 IS GENERATED FRESH.        nothing determines it; it is random bits.
A v3/v5 IS A HASH OF AN INPUT.  the same input always produces the same output, so the value
                                cannot change unless what is being hashed changes.
```

`kern.uuid` and `IOPlatformUUID` are both deterministic and **different from each other**, so they
hash different inputs. The nibble proves determinism; **it does not reveal what the input is.**

## E3 · Two identifiers carry visible collision scars

```text
Michaels-MacBook-Pro-**10**.local      the -10 was appended because the name was already taken
Michael's MacBook Pro **(2)**          the (2) likewise
```

Both suffixes are added when a name collides on the local network. They are evidence of collision
in the values themselves, on this machine, now.

## E4 · macOS exposes no per-install identifier

Of everything in `E1`, one value is per-boot and the rest are per-machine. There is nothing between.

```text
PER BOOT      kern.bootsessionuuid
PER MACHINE   IOPlatformSerialNumber · IOPlatformUUID · kern.uuid
```

Linux has `/etc/machine-id` for the middle tier — generated at install, surviving updates, changing
on reinstall. No equivalent appears in `E1`.

## E5 · This machine's own timeline, and what it cannot settle

```console
$ ls -ld /var/db/.AppleSetupDone      → Aug 18 2022
$ stat -f %SB /                       → 2025-07-19      (root filesystem created)
$ last reboot | head -1               → Sat Aug  8 11:02
$ uptime                              → up 4 days
```

The root filesystem was recreated in July 2025 while the setup marker is from 2022, so this machine
has been through a root-fs recreation. **The prior values of the three UUIDs were not recorded
before it**, so whether any of them survived is not answerable from this machine.

```text
WHAT WOULD SETTLE IT   record all three now; compare after the next OS update, and again after a
                       reinstall. E2 predicts the two hashes hold if their inputs are hardware.
                       UNTESTED.
```

## E6 · What identifies, and what does not

Restating `E1` sorted by whether the value distinguishes this machine from another:

```text
DOES NOT IDENTIFY   gethostname()          user-set · DHCP-changeable · collides (E3)
                    ComputerName            same, and collides (E3)
                    LocalHostName           same
                    uname() nodename        = the hostname
                    hw.model / hw.machine   a product line — every MacBookPro18,1 shares it
                    MAC addresses           all locally-administered (0a: prefix), virtual,
                                            spoofable, nine of them

IDENTIFIES A BOOT   kern.bootsessionuuid    v4, fresh each boot (E2)

IDENTIFIES THE      IOPlatformSerialNumber  factory-issued. present before the machine had an OS.
MACHINE             IOPlatformUUID          v5, read from IOKit's platform expert
                    kern.uuid               v3, deterministic
```

## E7 · There is no single universal identifier field

The mechanism differs on every platform, and on some devices there is nothing to read.

```text
macOS      IOPlatformSerialNumber           IOKit. factory-issued.          ← verified here
           IOPlatformUUID
Linux      /sys/class/dmi/id/product_uuid   DMI/SMBIOS. ROOT-ONLY on modern kernels.
           /etc/machine-id                  systemd. generated at INSTALL, not hardware.
           /proc/cpuinfo Serial             Raspberry Pi only.
Windows    Win32_ComputerSystemProduct.UUID SMBIOS, via WMI
           HKLM\…\Cryptography\MachineGuid   generated at install
FreeBSD    kern.hostuuid
Android    no stable hardware ID available to apps
iOS        identifierForVendor — per-vendor, resets on uninstall
no-OS      whatever the chip carries, if anything
```

Five distinct reasons it does not unify:

```text
DIFFERENT MECHANISM      IOKit · DMI · registry · sysctl · WMI. no common call, so the platform
                         layer needs a per-OS implementation regardless.

OFTEN PRIVILEGED         Linux's product_uuid requires root. a non-root runtime cannot read the
                         one field that is actually hardware.

DELIBERATELY WITHHELD    Apple removed UDID access in 2013; Android's ANDROID_ID has been
                         per-app-per-install since Oreo. POLICY, not oversight.

ABSENT ON SOME HARDWARE  ARM SBCs typically have no DMI at all.

CLONES COLLIDE           a VM image or container copied N times carries N identical
                         install-time identifiers unless something regenerates them.
```

And one that is specific to this design: `terminology.md` admits *"physical hardware addressed over
a network — **including devices with no OS**."* Those have no field to read at all.

*The macOS reads are verified on this machine. Everything below that line is NOT verified here —
Docker's daemon was not running, so no Linux was available to check.*

## E8 · Reach is plural, and the design already excludes it from the entity

`world.md` decides the invocation key on exactly this:

```text
"Same question to Claude and to a human is a DIFFERENT question. SAME QUESTION TO CLAUDE
 REACHED TWO WAYS IS THE SAME QUESTION — so the host does not enter what makes two askings
 comparable."

"You invoke an entity, not a host. DELIVERY PICKS A HOST ON THE WAY OUT, and the response
 correlates by invocation identity, not by channel. A human messaged on their TV and answering
 on their phone is not two events — it is one invocation arriving back through a different door."
```

And the code implements it:

```go
type Invocation struct {
    Entity ids.EntityID   // KEYED
    Host   ids.HostID     // "RECORDED, NOT KEYED"
}
```

`Host` is excluded from the key **because** reach is plural. If there were one way to reach an
entity, the exclusion would be pointless.

**Where the mapping is said to live:**

```text
"WHICH ENTITIES ARE ON A HOST is a registry fact, held by the runtime and rebuildable from
 configuration. It is not durable structure, for the same reason addresses are not: it changes,
 and the graph is append-only."

"Addresses are not graph structure… Entity → address → network → host is real, mutable,
 rebuildable from configuration, and belongs to the runtime."
```

Stated host-side — *which entities are on a host* — never entity-side.

**And the entity carries none of it:**

```go
// world.md, after every declaration dissolved
type Entity struct { ID EntityID }

// GONE: Reads []Language · Capacity · Presents/Captures modalities · Hosts []HostID
```

**What no type provides today:**

```go
type Host interface {
    ID() ids.HostID
    Execution() ExecutionOps
    Storage() StorageOps
    // ← nothing asks a host what it can reach
}

type Registry struct {
    interpreters map[ids.EntityID]func() Interpreter   // no host anywhere
}
```

*Quotations verified against `specs/2026-08-09/world.md` and `skyra-dna-v5/{graph/types.go,
host/host.go, entity/entity.go}` on this machine.*

## E9 · Address and transport, measured

Every entity reached this session, as the pair actually executed:

```text
                ADDRESS                                       TRANSPORT
curl            /usr/bin/curl                                 fork(2) + execve(59)
Claude          api.anthropic.com:443                         socket(97) + connect(98)
mongod          /tmp/mongodb-27017.sock                       socket(97) + connect(98)
libssl          /opt/homebrew/…/libssl.3.dylib                open(5) + mmap(197)
```

**A raw connection to Anthropic, with no curl, no shell and no HTTP library:**

```console
socket fd     3
connected to  160.79.104.10:443
local side    10.0.0.22:54060
```

**LAN and internet are the SAME mechanism** — the calls are identical and only the bytes in the
`sockaddr` differ:

```console
LAN gateway  socket()+connect() -> OSError
internet     socket()+connect() -> fd 4  peer 160.79.104.10
```

So distance is a property of the connection, not a separate resolution path. "Wifi" is a physical
layer three levels below anything the runtime sees — the same `connect()` runs over ethernet,
cellular, or a tunnel.

**Local entities are not all reached by `execve`.** These were live on this machine:

```text
/tmp/mongodb-27017.sock                    mongod, pid 889, up since 11:07
/var/run/docker.sock
/tmp/cc-socks/61290.sock                   this session
…/com.spotify.client…/SingletonSocket
```

Each is reached with `socket` + `connect`, not `execve` — you do not spawn them, you attach to
something already running.

*All of the above measured on this machine, this session.*

## E10 · The design already places reachability before the possibility opens

Written months before resolution was discussed, and it independently fixes the same boundary:

```text
world.md    "Projection happens AT DISPATCH, not at return."

              DELIVERY   could this be presented at all?   IMMEDIATE. the possibility never opens.
              WORK       was the answer any good?          however long it takes.

            "an undeliverable member returns instantly and DOES NOT BLOCK ITS GROUP — you do not
             wait on a human reading a TV to learn the TV had no speaker."

build.md    "TV with no speaker fails AT DISPATCH."

interfaces  "PROJECTION MUST FAIL LOUDLY — audio to a device with no speaker must ERROR, never
.md          silently drop. otherwise the invocation 'succeeds', nobody perceived anything, and
             the record reads as success."
```

Three properties fall out that resolution would otherwise have had to invent:

```text
IT FAILS IMMEDIATELY   an exhausted resolver fails on the same clock as a TV with no speaker.
                       one mechanism, not two.

IT FAILS LOUDLY        silent fallback is the same defect as silent projection — the record
                       would read as success. That requirement already covers resolution.

IT IS BEFORE THE       "the possibility never opens" — which is exactly where a resolver has to
POSSIBILITY OPENS      run. The layering in decision 5 and this timing are the same claim,
                       arrived at from opposite ends.
```

```text
PROJECTION FAILURE   the TV has no speaker            → fails at dispatch
REACH FAILURE        no address for Claude answers    → fails at dispatch
```

Both are *this could not be delivered*. So decision 5 is not a new subsystem — it is the existing
dispatch step gaining a resolver.

*Quotations verified against `specs/2026-08-09/{world.md, build.md, interfaces.md}`.*
