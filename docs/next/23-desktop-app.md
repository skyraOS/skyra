# The desktop app: the user's way in

Proposal: **In the proposal.** Cut 3 needs faults that climb to the human and the room's admin facet, tracked in [next/proposal](proposal/README.md) as the rest of D1.

Date: 2026-09-28  
Status: cuts 1 and 2 built (`desktop/`, `skyra app`), with the per-launch
token. Cut 3 waits on fault routing. Follows 16,
[20](20-boundry-crossing.md) and 22. The theme is the TUI's
([`tui/README.md`](../../tui/README.md)).

## What it is

`Skyra.app` is a native macOS app, written in SwiftUI, and it is where a person
enters the system. Today there are three ways in: `skyra` in a browser, `skyra
tui` in a terminal, and `skyra ask`. Each is started from a shell in the
directory that becomes the workspace. The app replaces the shell as the front
door. It launches the backend, asks the human for the one grant the system
needs, and holds the conversations.

It is a **client of the webhost's conversation protocol**, like the TUI. It is
not a registered surface, holds no frame descriptors, does not read `.skyra/`
and holds no runtime authority. Everything [20](20-boundry-crossing.md) says
about the actor side applies to it.

## Why it is the entry point

The thesis moves approval from every action to one grant of a Horizon. Today
that grant is implicit: the workspace is wherever the human happened to type
`skyra` (22). Nobody sees it, and nobody approves it. The
app makes it the first thing the human does:

| Moment | Today | In the app |
|---|---|---|
| **The grant** | the current directory, silently | the human picks a folder and sees what is being granted: read, under this path, sent to the model provider. One approval |
| **Everything inside it** | runs with no prompt | the same. The app shows it as it happens and never asks |
| **Reaching past it** | a horizon fault, shown in the transcript | the one moment the app interrupts: a request to cross, which the human grants or refuses (needs fault routing, below) |

The app is the human's side of the approval model. It asks once, stays quiet
after that, and speaks up only at the boundary.

## Where it sits

```text
Skyra.app (SwiftUI)          actor side: renders, emits the human's input
   │  WebSocket, the browser protocol (frontend/lib/protocol/types.ts)
webhost                      ingress host (16): holds parked frames, decides nothing
   │  park / resume
skyra                        the runtime: interprets, commits, releases, renders sandboxes
```

The app starts the webhost the same way `skyra tui` does (`skyra`, `tui_chat`).
It owns that process, stops it on quit, and leaves the Codex actor host
running. The webhost's log goes to `~/.skyra/app-webhost.log`.

### The rule for a new crossing, answered

| Question (20) | For the app |
|---|---|
| Who emits it, who checks it? | the app emits the human's messages and the workspace choice. The runtime checks both: messages at intake (a typed `{"to":…}` is still text), and the workspace in the same launcher code that refuses `$HOME`, `/`, `~/.skyra` and `~/.codex` today. The app does not repeat that check; it shows the refusal |
| Is it authority? | the workspace choice becomes authority, so it enters the way it does today: as registry data the launcher writes and the runtime pins by hash. The app never builds a Horizon or a sandbox profile |
| Is it state the runtime acts on? | no. Conversations are lineages in the tree; the app renders snapshots and keeps only view preferences |
| Does it name an actor? | yes (Codex, the browser protocol), which is why it lives outside the runtime |
| Is it a descriptor? | no. The app holds a WebSocket and a child process, nothing from `.skyra/` |
| Who holds it while nothing runs? | the webhost holds parked frames; the tree holds everything else. Quitting the app loses nothing but the view |

## The theme, carried over

The TUI builds its sky from glyphs and 24-bit color. SwiftUI can draw the same
sky at full resolution and get translucency from the system instead of
imitating it. Everything [`tui/README.md`](../../tui/README.md) describes carries
over, element by element:

| TUI | SwiftUI |
|---|---|
| glyph star field, stars breathing on their own periods, the eight-second glint | one `Canvas` inside a `TimelineView(.animation)`; stars seeded once, drawn as points and soft sprites |
| depth parallax on a slow camera sway, each star stepping at its own moment | the same sway function; offset by depth, continuous instead of stepping by cells |
| the core star, three satellites on undrawn orbits, braille-dot tails | drawn in the same `Canvas`; tails as fading points along the recent path |
| satellites leave their orbit for the pointer, never more than ~20 characters from home | `onContinuousHover` feeds the pointer to the port of `satellites.go`; radii in points, scaled from the TUI's character units. No mouse-tracking toggle: selecting text just works |
| teal and violet lights lifting the background, the meteor every 90 seconds | radial gradients under the stars; the meteor as one short streak |
| translucent panels over the sky | a translucent tint, not a blur: a system material turns flat grey over a sky this dark and hides it |
| the composer's glow: teal when ready, grey while the agent works | a soft shadow and border in `accent` or `dim`, animated between them |
| `✦ skyra` as a star-chart label that recedes in a conversation and brightens when something passes | the same, over two seconds; the window uses a hidden title bar and the sky runs under the traffic lights |

**The palette** is the TUI's (`tui/render.go`, `tui/wallpaper.go`): sky
`8,10,14`, accent `126,196,204`, operator `96,196,166`, ink `214,219,229`,
dim `120,128,146`, faint `60,66,80`, red `226,112,112`, and the star, core and
title ramps. It is dark only, like the TUI.

**Motion is quiet and costs little.** With Reduce Motion on, the camera and
meteors stop and the stars hold still. The sky drops its frame rate when the
window is inactive, and stops when it is occluded or minimized.

## What the app shows

1. **Launch:** the saved workspace, or a folder picker. The grant is shown in
   one line: `read · ~/code/project · contents go to the model provider`.
   Changing it restarts the backend under the new registry.
2. **Sidebar:** saved conversations, which the protocol already lists (the TUI
   does not browse them; the browser does). Delete, as in the browser.
3. **Conversation:** the transcript, the composer, phase and connection status
   in the header, faults shown inline and the conversation carrying on (16).
4. **Later, the boundary:** a request to cross, shown as a sheet over the
   conversation and delivered as a notification when the app is in the
   background. This depends on the protocol carrying such a request, which
   does not exist yet.

## Layout

```text
desktop/
  Package.swift            SwiftPM, macOS 14+, no dependencies
  Sources/Skyra/
    SkyraApp.swift         the App, the window, the backend's lifecycle
    Backend.swift          starts and stops the webhost, reads its log on failure
    Protocol.swift         Codable mirror of frontend/lib/protocol/types.ts
    Connection.swift       URLSessionWebSocketTask, reconnect, snapshot reconciliation
    Sky/                   Theme, Stars, Satellites (a port of tui/satellites.go), SkyView
    Views/                 Sidebar, Conversation, Composer, GrantSheet
  Tests/SkyraTests/        protocol decoding against webhost output; satellites against tui's cases
```

`skyra app` builds it with `swift build` and wraps the binary into
`build/Skyra.app` next to the Go binaries, so `install.sh` stays the one
setup step and Xcode is optional.

## Cuts

| Cut | Scope | Depends on |
|---|---|---|
| **1. Parity** | the sky, one window, conversations over the existing protocol, the app launching the webhost | nothing new in the runtime |
| **2. The grant** | the folder picker becomes the workspace; the launcher's refusal is shown in the app | the launcher accepting a workspace from its caller (it already reads `SKYRA_WORKSPACE`) |
| **3. The boundary** | an authority fault reaches the human as a request, the app shows it, the answer goes back as a grant or a refusal | faults climbing to the room, and the room's admin facet, which only the human holds ([24](24-the-room.md), [26](26-agents-and-facets.md)); a protocol message for it |

Cut 1 is a client. Cut 3 is where the app starts to carry the thesis, and it
cannot come before the runtime routes faults by kind.

## Limits, and what is open

- **The WebSocket is unauthenticated** (fixed for the app: `skyra serve`
  passes `SKYRA_WS_TOKEN`, and the webhost answers 401 without it. `skyra`
  and `skyra tui` still run without one). The webhost listens on loopback and
  takes messages from anyone who connects. Once the app is the human's
  entry point, a message on that socket *is* the human, and the user outranks
  the agent (16). Any local process, including one a sandbox allows loopback
  to, could speak as the human. The app launches the webhost, so it can hand
  it a per-launch token through its environment or a pipe; the webhost would
  reject connections without it. This should land with cut 1, or at the latest
  with cut 3.
- **Finding 1 of [20](20-boundry-crossing.md) is closed** (2026-09-30): the
  webhost hands back a stamp, and resume derives the instance from the tree.
  Before that, the webhost handed the runtime an unchecked `Instance` on resume. The app adds nothing to
  that, but it makes the webhost part of the product, so the fix matters more.
- **Two copies of the palette.** Go in `tui/`, Swift in `desktop/`, and CSS in
  `frontend/`. Either one file (for example `design/theme.json`) generates all
  three, or a test compares them.
- **Codex login** is still `codex login` in a terminal. The app can detect
  that it is missing and say so; doing the login itself is out of scope.
- **Signing and distribution** are out of scope. Cut 1 runs unsigned, from
  `build/`.
- **Does the browser frontend retire?** Once the app reaches parity, `frontend/`
  duplicates it. Keeping it costs a third client of the protocol.
