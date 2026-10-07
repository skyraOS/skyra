# Cua desktop observations

Skyra can observe one operator-selected macOS window through Cua Driver
**0.34.0**. The tool returns accessibility text and a screenshot, records the
exact image bytes, and delivers the image to Codex as `inputImage`. Browser and
native desktop ledgers can display the recorded screenshot. It cannot click,
type, navigate, launch apps, write clipboard contents or dispatch arbitrary
Cua tools. Desktop access is opt-in for each launch.

## Use

Install the pinned release with the official installer:

```sh
curl -fsSL https://cua.ai/driver/install.sh -o /tmp/cua-install.sh
CUA_DRIVER_RS_VERSION=0.34.0 bash /tmp/cua-install.sh --no-modify-path
```

Grant **CuaDriver** Accessibility and Screen & System Audio Recording through
macOS when its permission dialog appears. Permissions belong to the signed
`/Applications/CuaDriver.app`; the session launcher uses LaunchServices to
preserve that identity. Then select an already-open window:

```sh
./skyra desktop --list
./skyra desktop --app com.example.App --pid 123 --window-id 456 chat
```

The same prefix works with `tui`, `serve`, or `ask "TEXT"`. The default is
`chat`. The window must currently belong to the selected bundle ID and PID.
A window chooser is operator-side; other windows are never offered to the
model. To grant a different window, end the session and start another.

`desktop_observe {}` observes the selected window. Explicit `pid` and
`window_id` must match it. Optional `include_screenshot:false` requests text
only; `query`, `max_elements` (up to 250) and `max_depth` (up to 25) bound the
accessibility walk. Queries filter accessibility text, not screenshot pixels.
Cua's accessibility result can include the selected application's menu bar.
App scope does not restrict browser tabs or web origins.

## Lifetime and authority

`session.py` is trusted operator bootstrap, outside model tool authority. It
checks the signed installed app and exact driver version, verifies the selected
window, creates a private temporary directory, and starts a separate Cua app
instance with an immutable, approved **bounded** manifest. The manifest allows
only `get_window_state`, the selected app's windows, and no full-display capture.
The bridge further restricts requests to the exact selected PID/window pair.

Bootstrap checks the socket's reported process against its PID file and the
installed executable. It records binary and manifest hashes in the tool
contract, which the ordinary registry pins at genesis. It generates an optional
`think` bundle that includes desktop observation. The ordinary launcher remains
without desktop access. The session owns and closes its bridge, driver and
client; service death ends the scoped session. Conversations retain history,
but the closed service cannot provide future observations.

Each operation verifies driver version, bounded mode and the exact manifest
hash on the same upstream connection. The generated manifest expires after
8 hours and 30 minutes idle. The adapter gets the bridge socket, not the
upstream driver socket. Skyra's shared Horizon may also include browser/ideas
service grants; the adapter selects only its designated bridge from the
complete `SKYRA_CONNECTIONS` list. No new filesystem grant is needed.

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
operator window selection and existing immutable records.
