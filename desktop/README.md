# Skyra.app

The desktop client and the user's way in ([next/23](../docs/next/23-desktop-app.md)).
Like the TUI, it is a client of the webhost's conversation protocol: it holds
no frame descriptors, never reads `.skyra/`, and holds no runtime authority.

```text
skyra app          build build/Skyra.app and open it (needs Xcode, not only the Command Line Tools)
swift test         in desktop/: protocol, session and sky tests
swift Icon/make-icon.swift OUT.iconset   the app icon, drawn in code; skyra app packs it
```

On first launch it asks for the **grant**: the folder the read tool may read,
approved once. It suggests `~/Skyra`, a clean folder of its own that is created
only when you grant it; Choose… picks any other. It then runs `skyra serve` in that folder, with a free loopback
port and a fresh token the webhost requires. If the launcher refuses the
folder (`$HOME`, `/`, or anything overlapping `~/.skyra` or `~/.codex`), its
reason is shown. The backend's output goes to `~/.skyra/app-webhost.log`.
Quitting the app stops the webhost; the Codex actor host keeps running
(`skyra stop`).

| Key | Action |
|---|---|
| Return | Send; the first message opens a conversation |
| Shift+Return | New line |
| ⌘N | New conversation |
| ⌘D | Close the conversation on screen |
| ⇧⌘⌫ | Delete the conversation on screen (also: the trash on a hovered row, or right-click) |
| ⌘O | Change the workspace (restarts the backend); also the folder in the header, which lists the folders granted before |
| ⇧⌘R | Restart the backend |
| Conversation drag bar and edges | Drag the bar to move; drag an edge or the visible corner to resize. Text remains selectable and scrollable. Placement is kept between launches |
| Expand / Reading width | Fill the window width for long content, or return to the readable column |
| Focus / Exit focus | Fill the space between header and composer and pause the sky |
| Context menu in header | Select procedures, switch Think models, and navigate the conversation working set, including in Focus |

The sky is the TUI's (`tui/wallpaper.go`, `tui/satellites.go`), drawn with
Metal at 15 frames a second: the same
stars in the same places (the tests pin the hash), and satellites that leave
their orbit for the pointer. Under them, a nebula of slow lights and
filaments drifting through a field the pointer stirs ([next/31](../docs/next/31-the-sky-dreams.md)).
Each frame is painted into a draw list (`Sky/SkyPaint.swift`) and drawn in a
few instanced passes (`Sky/SkyRenderer.swift`); the tests draw frames
offscreen and read them back. Over it, the conversation
on screen's Horizons, as the webhost reads them from its lineage: each a faint
region as wide as what it reaches, each surface in reach one star (its name
shows under the pointer), shared stars where Horizons overlap, and a call's
light flying to the star it invokes. Under the pointer, away from the
conversation, a Horizon draws its constellation: its lines traced one at a
time, its stars lit and named. The backdrop warps a little around the
pointer, nearer stars more. It pauses when the window is hidden, the app is inactive, history is open,
Focus is enabled, or Reduce Motion is enabled. The procedure overview shows
actors and tool counts; each procedure's inspector opens a scrollable list
of actors and their tools.

The live tests drive the app's own session against a running backend:

```sh
SKYRA_WS_PORT=18432 SKYRA_WS_TOKEN=t skyra serve --fake &
SKYRA_TEST_WS=ws://127.0.0.1:18432/ws SKYRA_TEST_TOKEN=t swift test
```

For performance diagnostics, use Instruments > Points of Interest with the
`skyra.desktop` subsystem. Intervals cover snapshot decode/apply, scene
generation, GPU admission wait, drawable acquisition, and encode/upload.
Snapshot events contain byte and entry counts, without transcript text.
`SKYRA_PROFILE_UI=1` enables the webhost's snapshot size/build-duration logs
and the opt-in desktop synthetic timing workload:

```sh
DEVELOPER_DIR=/Applications/Xcode.app/Contents/Developer SKYRA_PROFILE_UI=1 swift test --package-path desktop --filter profileThinkWorkloads
```

The desktop opts into conversation replacement deltas after receiving its
initial full snapshot. The host retains legacy full-snapshot behavior for
other clients; reconnects and baseline mismatches request a fresh full snapshot.
Unchanged conversation histories are retained locally. Frame-size truncation
continues to use the existing 16 MiB limit.
