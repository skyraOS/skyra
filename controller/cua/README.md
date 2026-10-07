# Cua desktop observations

The normal Skyra app includes read-only macOS vision through Cua Driver
**0.34.0**. The tool returns accessibility text and a screenshot, records the
exact image bytes, and delivers the image to Codex as `inputImage`. Browser and
native desktop ledgers can display the recorded screenshot. It cannot click,
type, navigate, launch apps, write clipboard contents or dispatch arbitrary
Cua tools. The normal real app backend starts the vision service automatically. No extra
launch command or window picker is required. Fake backends do not start Cua.

## Use

Open Skyra normally and start a new conversation under `think`. Its
`desktop_observe` tool can list windows open at backend startup, then capture
one by its `pid` and `window_id`. Restart the app to refresh that inventory
when you need a newly opened window. Existing conversations retain their
original grants; start a new conversation after updating Skyra.

The app backend starts a private bounded Cua service and closes it when the
backend ends. It has no click, typing or arbitrary driver dispatch operations.
Cua must be installed as described below.

### Installation and optional single-window sessions

Install the pinned release with the official installer:

```sh
curl -fsSL https://cua.ai/driver/install.sh -o /tmp/cua-install.sh
CUA_DRIVER_RS_VERSION=0.34.0 bash /tmp/cua-install.sh --no-modify-path
```

Grant **CuaDriver** Accessibility and Screen & System Audio Recording through
macOS when its permission dialog appears. Permissions belong to the signed
`/Applications/CuaDriver.app`; the session launcher uses LaunchServices to
preserve that identity. Start the window picker in your terminal:

```sh
./skyra desktop
```

Choose a window by number to start chat with `desktop_observe` available inside
`think`. Enter `q` to cancel. You can also use `skyra desktop tui` or
`skyra desktop --app com.example.App` to filter the picker to one app.
The selected window's ownership is checked again before the session starts.

For scripted launches, select an already-open window explicitly:

```sh
./skyra desktop --list
./skyra desktop --app com.example.App --pid 123 --window-id 456 chat
```

The same prefix works with `tui`, `serve`, or `ask "TEXT"`. The default is
`chat`. The window must currently belong to the selected bundle ID and PID.
A window chooser is operator-side; other windows are never offered to the
model. To grant a different window, end the session and start another.

In normal app sessions, `desktop_observe {}` lists the available windows;
explicit `pid` and `window_id` select one of that launch’s pinned windows.
In optional single-window sessions, omitted IDs capture the selected window. Optional `include_screenshot:false` requests text
only; `query`, `max_elements` (up to 250) and `max_depth` (up to 25) bound the
accessibility walk. Queries filter accessibility text, not screenshot pixels.
Cua's accessibility result can include the selected application's menu bar.
App scope does not restrict browser tabs or web origins.

## Lifetime and authority

`session.py` is trusted operator bootstrap, outside model tool authority. It
checks the signed installed app and exact driver version, records the selected
window inventory, creates a private temporary directory, writes the immutable,
approved **bounded** manifest there, and starts the bridge. The manifest allows
only `get_window_state`, the inventoried apps’ windows, and no full-display capture.
The bridge further restricts captures to the exact inventoried PID/window pairs.

The bridge launches a separate Cua app instance from that manifest, in the
same private directory, and checks the socket's reported process against its
PID file and the installed executable. Bootstrap records binary and manifest
hashes in the tool contract, which the ordinary registry pins at genesis. It
generates an optional `think` bundle that includes desktop observation. Normal
real `serve` startup enables desktop observation in `think` automatically.
Other launch modes can use the optional `skyra desktop` wrapper. The session
owns and closes its bridge and client; service death ends the scoped session.
Conversations retain history, but the closed service cannot provide future
observations.

Each operation verifies driver version, bounded mode and the exact manifest
hash on the same upstream connection. The generated manifest expires after
8 hours and 30 minutes idle, and Cua latches either for the life of its
process. The bridge does not keep a daemon alive; it replaces one on demand,
as an idle Codex server is reclaimed and started again. When an observation
finds the manifest idle-expired or expired, or the daemon gone, the bridge
stops that daemon, launches and verifies a new one from the same manifest
bytes (so the pinned hash still holds), and retries the read-only observation
once. A failed launch backs off exponentially, from 2 seconds to 2 minutes;
inside the backoff a request is refused with `unavailable` at once.

The bridge stops its daemon when it stops: `shutdown_if_pid` on the daemon's
socket, then SIGTERM, then SIGKILL, each only while the PID is still the
installed executable. If the bridge is killed outright, the session's cleanup
ends any daemon still serving from its private directory. Cleanup ignores
further signals, so a second one cannot cut it short. Cua's stdin
parent-liveness mode does not end a daemon outside Cua's embedded mode, which
the bridge does not use.

The adapter gets the bridge socket, not the upstream driver socket. Skyra's
shared Horizon may also include browser/ideas service grants; the adapter
selects only its designated bridge from the complete `SKYRA_CONNECTIONS` list.
No new filesystem grant is needed.

Images are inline immutable observation data in Skyra's existing records,
not mutable scratch files. PNG/JPEG images are limited to 2 MiB decoded and
1280 pixels per edge, decoded and checked against their recorded SHA-256 and
dimensions. Responses are limited to 4 MiB, text to 64 KiB, requests to 4 KiB,
concurrency to four, and each observation to 15 seconds. Replay uses recorded
bytes and never calls Cua. Provider projection separates image content from
text. Refusals, missing captures, truncation and degradation remain visible.

## Verification

On 2026-10-07, Cua 0.34.0 and real Codex successfully observed the included
fixture through the normal `skyra desktop … ask` path. Codex recognized drawn
text and a purple triangle that were absent from the accessibility tree. The
1040×616 PNG was recorded; all 55 records replayed byte-identically. Session
cleanup removed its private daemon. A real `think` conversation also returned
“a purple triangle, with the number 42” and included the recorded image in
the authenticated UI snapshot. The finished one-shot lineage replayed again
after all Cua services and the fixture app were stopped. Direct Cua `click` was denied by the
manifest, and an observation outside its resource scope was refused.

Automated tests cover malformed/widened requests, daemon identity mismatch,
image tampering and bounds, provider image projection, sandbox socket denial,
and replay after the bridge socket has been closed:

```sh
go test -race ./controller/cua/... ./controller/codex/resident ./controller/core/machine
go test ./controller/... ./tui/...
(cd frontend && npm run typecheck)
(cd desktop && DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer swift test)
```

The upstream protocol and manifest format are pinned to
[the 0.34.0 source](https://github.com/trycua/cua/tree/cua-driver-rs-v0.34.0/libs/cua-driver/rust).
The initial broader discovery/artifact proposal remains in
[next/34](../../docs/next/34-cua-driver-observation.md); this implementation uses
a pinned startup window inventory and existing immutable records.
