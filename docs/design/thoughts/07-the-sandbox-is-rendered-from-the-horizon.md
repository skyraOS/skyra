# The sandbox is rendered from the Horizon

Date: 2026-09-23  
Status: agreed 2026-09-23 (D7). Follows [06](06-the-release-is-the-horizon.md).

## The idea

A surface no longer carries a fixed Seatbelt profile. At spawn, `skyra` renders
T's profile from the Horizon and applies it to T. The Horizon is the one source
of truth. What T may **ask for** (Moves) and what T may **touch** (the OS) are
two projections of the same object.

```text
H_{N+1} ──render──▶ SBPL profile ──apply at spawn──▶ T
```

## Why this is right

- **No drift.** A static profile per surface can grant more than the frame's
  Horizon, and nothing would notice. With rendering, narrowing the Horizon
  narrows the sandbox automatically.
- **It matches the narrowing rule.** `H_{N+1} ⊆ H_N` becomes
  `profile_{N+1} ⊆ profile_N` for free.
- **It is still computed, not stored.** `render` is pure, so the profile is
  derivable from the chain. Its hash can go in the commit record, where it acts
  as evidence rather than as state.

## Details to accept

1. **Seatbelt does not read anything. `skyra` renders.** The trusted part turns
   the Horizon into an SBPL string and applies it at spawn (`sandbox-exec -p`
   or `sandbox_init`). Seatbelt only enforces.
2. **The Horizon needs an OS section.** Logical grants ("may invoke echo") say
   nothing about files or sockets. Add `fs` (read and write subpaths), `net`
   (outbound endpoints) and `exec` (binaries T may spawn), and have `narrow`
   intersect those too.
3. **A floor that cannot be granted.** `render` always denies reads and writes
   of `.skyra/`, whatever the Horizon says. No Horizon can grant it, so it is
   not a field.
4. **The network is coarse.** Seatbelt filters by address and port, not by
   hostname. "The provider's endpoints" becomes either "outbound 443" or a local
   proxy that T is limited to. Deny by default either way.
5. **Fixed for T's lifetime.** A profile cannot be re-applied to a running
   process. If a live T's Horizon must change (for example Codex `tools-add`),
   that means a new frame and a new T. This matches today's replacement-thread
   behaviour, now as a replacement process.

## Risk to test first

**Nested Seatbelt.** On macOS, the Codex CLI sandboxes its own tool commands
with Seatbelt. If we run Codex inside our profile, its own sandbox call may be
denied, and applying a sandbox inside a sandbox may not be permitted. Before
building on this, run the real Codex app-server under a rendered profile and
confirm two things: (a) it starts and runs a turn, and (b) its own tool sandbox
still applies or fails closed. If it fails open, the design changes.

## Acceptance

- From inside T, reading or writing `.skyra/` fails.
- A child Horizon narrower than its parent's renders a profile under which a
  path allowed only to the parent is denied.
- Replaying the chain re-renders byte-identical profiles, whose hashes match the
  commit records.
