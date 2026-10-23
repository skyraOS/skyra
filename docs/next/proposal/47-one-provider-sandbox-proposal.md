# 47 — One sandbox path for Codex and Claude

Date: 2026-10-23

Status: **implemented** (2026-10-23), with the keychain helper chosen for
open question 1. Where the build differs from the design below, "As built"
says so. Live turns of both providers pass under the shared profile.

**Scope:** how a model provider's resident (`codex.start`, `claude.start`) is
prepared and how its Seatbelt profile is rendered. The bridge, the stream
protocol, `codexio` and the resident adapters are unchanged.

## Summary

Codex and Claude are two providers of the same thing: a model CLI the user is
already logged in to, run as a resident under a profile the runtime renders.
Today they take two different paths, and only one of them is a sandbox:

| | Codex (`darwin-appserver-v1`) | Claude (`darwin-claude-v1`) |
|---|---|---|
| Base | `(deny default)` | `(allow default)` |
| Exec | the pinned app-server only | anything |
| Network | loopback to the egress proxy, host allowlist on the surface | anything, DNS included |
| Home | private `run/home`, `CODEX_HOME` | the user's real `$HOME`, read and written |
| Auth | `~/.codex/auth.json` symlinked in, read-only | whatever the CLI finds |
| Horizon grants | image reads only (the `view_image` hole) | none rendered; nothing denied either |
| Version pin | `PinnedCodex`, checked before start | none |
| Render | `RenderResident` | `RenderClaudeResident` |
| Prepare | `prepareAppserver` | `prepareClaude` |
| Tests | yes | none |

The proposal: one profile name, one render function, one prepare function.
The Horizon decides the same things for both. What differs between providers
(which binary, how the login is reached, which hosts the proxy allows) is a
small pinned descriptor on the surface, not a second code path.

## Why Claude went to `(allow default)`

`586de43` (2026-10-07) switched the Claude profile from deny-default to
allow-default because the CLI's keychain login "uses undocumented sandbox
operations that cannot be individually enumerated". It kept an exec
allowlist. `2edcb34` then removed the exec allowlist too, leaving the floor
deny as the only rule. The doc comment still describes the exec allowlist.

The keychain operations can be enumerated. They are listed below.

## Evidence

One `claude -p` turn (haiku, no tools, no MCP, `--setting-sources user`) was
run under `sandbox-exec` with a profile built from `RenderResident`'s rules,
and every `deny` the kernel logged was examined. Three changes made it work:

1. **Auth.** The CLI does not use the keychain API in-process: it execs
   `/usr/bin/security find-generic-password -s "Claude Code-credentials"`.
   The profile needs exec of `/usr/bin/security`, mach-lookup of
   `com.apple.securityd.xpc` (beside Codex's `com.apple.SecurityServer`),
   and reads of `~/Library/Keychains`, `~/Library/Preferences/com.apple.security.plist`,
   the global preference plists and `~/.CFUserTextEncoding`.
2. **Home.** With `HOME` pointing at a private directory, `security` looks
   for the login keychain under it and the CLI reports "Not logged in".
   `HOME` must stay the real home, but nothing under it is readable except
   the keychain paths above. The CLI's own state moves with
   `CLAUDE_CONFIG_DIR=run/home/.claude`. Setting
   `CLAUDE_SECURESTORAGE_CONFIG_DIR=` (empty) keeps the keychain item name
   unsuffixed, so the existing login is found. Without it the CLI looks for
   `Claude Code-credentials-<hash of the config dir>`.
   The user's default profile reaches the CLI by copying
   `~/.claude/settings.json` into that directory before start.
3. **Network.** Under deny-default the CLI hangs on DNS
   (`/private/var/run/mDNSResponder`). With `HTTPS_PROXY` pointing at
   `controller/codex/provider/proxy.py` and outbound allowed only to that
   loopback port, it works. The proxy journal shows `api.anthropic.com`
   allowed and `http-intake.logs.us5.datadoghq.com` refused.

Result: `{"is_error": false, "result": "ok"}`. Denials left in the log are
attempts the CLI survives without: exec of `git`, `ssh`, `ps` and `/bin/sh`,
`com.apple.FSEvents`, `logd`, `analyticsd`, the notification-center shm,
`/tmp/cc-socks-502`, and a lock under `~/.local/state/claude`. They stay
denied.

With the user's real `~/.claude` as the config directory instead, the CLI
writes `projects/`, `sessions/`, `telemetry/` and plugin lock files there on
every turn. That is the reason for the private config directory, not only
tidiness: a provider turn must not change the user's own Claude Code state.

## The design

### D1. One profile, `darwin-provider-v1`

`darwin-appserver-v1` and `darwin-claude-v1` are replaced by one name. The
registry accepts only it for a provider resident. Old records that name the
two old profiles still replay: replay checks the recorded profile hash, it
does not re-render.

### D2. A provider descriptor on the surface

What differs between providers is data, pinned in the surface (so it is part
of the surface's identity and the registry refuses a change to it):

```go
type Provider struct {
    Exec       []string // programs the resident may exec, beyond its verified copy
    Version    string   // the one version the profile is verified for
    ProxyHosts []string // the egress allowlist (moves here from Surface)
    Auth       string   // "file" or "keychain"
    AuthPath   string   // file: the credential file, linked read-only into home
    Keychain   string   // keychain: the generic-password service name
}
```

Registered values:

| | Codex | Claude |
|---|---|---|
| `Exec` | — (the app-server is the verified copy) | `claude` binary (resolved), `/usr/bin/security` |
| `Version` | `0.160.0` | `2.1.309` |
| `ProxyHosts` | `chatgpt.com`, `api.openai.com`, `auth.openai.com` | `api.anthropic.com`, `claude.ai`, `platform.claude.com` |
| `Auth` | `file`, `~/.codex/auth.json` | `keychain`, `Claude Code-credentials` |

For Claude the verified copy is `claude-start`, which execs the CLI. The CLI
binary is checked the way `PinnedCodex` is: by `--version` before start.

### D3. One render: `RenderProvider(h, floor, desc, paths)`

`RenderResident` and `RenderClaudeResident` become one function. In order:

1. **Common base**, the rules both providers already need: `(deny default)`,
   fork, exec of the verified copy and `desc.Exec`, the loader and system
   reads, the Codex mach-service list, `file-read-metadata`, the private
   `home`, `scratch` and `work` directories.
2. **Horizon grants**, the same for both. Today the only one is the image-read
   hole (`imageRead` over `h.Authority.Read` and `Scratch`, none when
   `h.Delegated`). A provider's native reads are unrecorded observations, so
   no other shared grant renders for either provider until next/25 item 3
   and next/26 narrow a provider's Horizon through the room. This is the
   part that "ties grants to the render": the Horizon reaches both providers
   through the same lines of code.
3. **Auth**, by kind: `file` adds a read of `AuthPath` and a write deny on
   it; `keychain` adds the reads and the `securityd.xpc` lookup from
   Evidence 1.
4. **Network**: outbound to the proxy port only. For Claude, also the bridge
   port (see D4).
5. **Floor**: deny `.skyra`, then allow the verified copy's literal read.

### D4. One prepare: `prepareProvider`

`prepareAppserver` and `prepareClaude` share the run directory, `home`/
`tmp`/`work`, the cache lookup, the proxy start, the version check and
writing `actor.sb`. What differs:

- Auth materialization. `file` links the credential file into `home`;
  `keychain` sets `CLAUDE_SECURESTORAGE_CONFIG_DIR=` and keeps `HOME` real.
- Config. Codex writes `config.toml` with `quiet`. Claude gets
  `CLAUDE_CONFIG_DIR=home/.claude` with a copy of the user's
  `~/.claude/settings.json`, plus `CLAUDE_CODE_TMPDIR=tmp`.
- The bridge port. `claude-start` binds its MCP bridge at runtime, after the
  profile is rendered. Prepare allocates the port instead and passes it in,
  so the profile names one loopback port rather than `localhost:*`, which
  would reach any local service.

Both get `HTTPS_PROXY`. `provider.Environment` keeps refusing inherited
API credentials.

### D5. Tests

- `RenderProvider` golden tests for both descriptors.
- A property test: the Horizon-dependent lines of the profile are identical
  for both descriptors given the same Horizon.
- The darwin-only probe: one Claude turn and one Codex turn under the
  rendered profile, asserting success and the proxy journal's hosts. This is
  the "live isolation smoke test" `controller/claude/README.md` lists as
  required.

## Steps

1. Add `Provider` and `RenderProvider`; port Codex to it with byte-identical
   output (the existing tests are the check).
2. Port Claude to it; delete `RenderClaudeResident` and `prepareClaude`.
3. Merge the prepare functions; registry accepts only `darwin-provider-v1`.
4. Register `claude.start` with proxy hosts and the version pin; the darwin
   probe.
5. Update `controller/claude/README.md` ("Still required" item 1 done).

## As built

- **Login with no keychain in the sandbox.** The helper
  (`controller/claude/keychain`) goes further than Evidence 1 needed. Its
  broker runs outside the sandbox and reads the keychain there, so the
  profile grants none of Evidence 1's rules: no `/usr/bin/security`, no
  `~/Library/Keychains`, no `securityd.xpc`. Inside, the helper's copy is the
  only `security` on `PATH` and talks to the broker over one socket in
  `/private/tmp/skyra-<uid>`. It answers find and update for the pinned
  item and refuses everything else, deletion included.
- **A private `HOME` for Claude too.** With the keychain outside, Evidence 2's
  reason for a real `HOME` is gone. Claude's `HOME` is the run's private home,
  as Codex's is, with the user's `settings.json` copied into
  `home/.claude/` and denied to writes. No `CLAUDE_CONFIG_DIR` is needed.
- **The descriptor** is `horizon.Provider`: `name`, `program`, `reads`,
  `auth`, `auth_path`, `keychain`, `helper`, with hashes pinned by the
  registry. The proxy fields stay on the surface, shared by both. Version
  pins are runtime constants (`PinnedCodex`, `PinnedClaude`), checked
  against `program --version` before start, so a registry cannot relax them.
- **Bridge port**: `prepareProvider` allocates it, `claude-start` binds
  `SKYRA_CLAUDE_BRIDGE_PORT`, and the profile allows bind and connect on that
  port only.
- **Tests**: `TestProvidersShareTheHorizonsRules` (the property in D5),
  `TestProviderProfiles`, `TestOSClaudeProfileEnforced` (Seatbelt denies the
  real `security`, `/bin/sh`, any other port and any other socket), the
  registry's `TestProviderIsPinnedWithItsProfile`, and the helper's tests.

## Open questions

1. **Keychain reach.** *Resolved: the helper.* Exec of `/usr/bin/security` plus a read of
   `~/Library/Keychains` lets the resident ask for any generic password
   whose ACL trusts `security`, not only Claude's. The profile cannot filter
   `security`'s arguments. Alternatives: `claude setup-token` with the token
   given to `claude-start` through a file the runtime reads (but
   `provider.Environment` refuses `CLAUDE_CODE_OAUTH_TOKEN` today, as a
   billing guard, and that rule would need to change), or a small pinned
   helper that only returns the one item. This is the main reason to review
   before building.
2. **Plugins.** The private config directory has `settings.json` but not
   `plugins/`, so `enabledPlugins` names plugins that are not installed there.
   With native tools off, they are unused. Copying or linking them in is
   possible but widens the reads.
3. **Version pin for Claude.** The CLI updates itself. A pin means the
   provider refuses to start after an auto-update until the profile is
   re-verified, as Codex does.
4. **Profile names in old records.** D1 assumes replay never re-renders.
   Confirm against 43 D5.
