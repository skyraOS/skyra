# WHERE — Evidence

**Status: OBSERVATION.** Measurements bearing on **WHERE**. Commands and their output. **No
conclusions** — what these mean is [`../outcomes/Where-outcomes.md`](../outcomes/Where-outcomes.md).

**Scope:** which machine work runs on, and how a machine knows which one it is.

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
