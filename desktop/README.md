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

The interface is organized around **Missions**. Each Mission owns a Plan
conversation, a versioned Spec, and its Launches. The left sidebar selects Missions across all
conversation procedures. The workspace has **Plan / Spec / Launches** tabs and an
explicit **Launch** button. The selected Mission's accepted execution appears in
the sky; procedure constellations and their navigation have been removed.

On the first Mission launch for a workspace, existing conversations are
archived in desktop navigation. Open **Archive** and **Restore** to bring one
back. Their runtime records, specs, and tree runs stay intact. New Missions
remain in the active list across launches; archiving never cancels work.

Use **New mission** to create an empty Mission. Its conversation, in
**Plan**, is with Plan itself: each message is a Plan turn, which can take a
while. Plan looks around the workspace, answers, and keeps the spec; a reply
that changed the spec shows **Spec rN**, which opens the spec as that reply
left it. **Spec** shows a revision, the latest by default, each part coloured
by where it came from (quoted from you, written by you, or Plan's own words),
with its cite coverage, tools and tree profile; step through revisions with
the arrows. The profile is the constants Launch's tree runs under: tools per
node (past it, a node delegates), turns per node, depth, instructions per
call, and a budget of turns for the whole tree.
**Edit** changes the latest revision: what you keep keeps its source, what you
write is yours, and the steppers set its profile. **Launch** passes the revision shown, with its tools, and opens
Launches. Spec drafts survive tab and
Mission changes during the session. Procedure, model, and workspace settings
are available from the toolbar's settings button.

| Key | Action |
|---|---|
| Return | Send to the Plan conversation; the first message creates a Mission if needed |
| Shift+Return | New line |
| ⌘N | Create a Mission |
| ⌘D | Close the selected Mission's chat |
| ⇧⌘⌫ | Archive the selected Mission |
| ⌘O | Change the workspace |
| ⇧⌘R | Restart the backend |
| ⇧⌘T | Open the detailed Trees inspector |

The sky uses Metal at 15 frames per second, with a per-Mission background
glow, a tree anchor before execution, and accepted star nodes and activity
lights during execution. Text panels quiet the sky underneath them. Window
occlusion, app inactivity, Archive, and Reduce Motion pause ambient animation.
The Trees inspector retains detailed replay. Tree execution needs a real
backend with its tree registry; fake mode supports the Plan conversation only.
Lifecycle and protocol: [conversation work](../docs/design/conversation-work.md).

The Trees window shows accepted work as motion: new branches grow from their
parents, spec cuts glow amber, passed observations travel downward, and results
travel upward in green. Tool calls and faults pulse at their node. Click an
activity chip to read a bounded payload preview and the source/destination name
references. The timeline can replay those same events. Cards, curves and transfer
labels share one layout projection as the window resizes. Live nodes waiting
on tools or children breathe softly; paused history and inactive windows stop
that motion. A parent resuming after all siblings return successfully draws their
results together, and a successful root return blooms once. Faults and exits
do not trigger a success bloom. Reduce Motion keeps static cues and readable
activity details.

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
