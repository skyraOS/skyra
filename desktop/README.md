# Skyra.app

The desktop client and the user's way in ([next/23](../next/23-desktop-app.md)).
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
| ⌘O | Change the workspace (restarts the backend) |
| ⇧⌘R | Restart the backend |

The sky is the TUI's (`tui/wallpaper.go`, `tui/satellites.go`), ported to a
`Canvas`: the same stars in the same places (the tests pin the hash), and
satellites that leave their orbit for the pointer. It stops when the window is
hidden and with Reduce Motion, and slows to ten frames a second when the app
is in the background.

The live tests drive the app's own session against a running backend:

```sh
SKYRA_WS_PORT=18432 SKYRA_WS_TOKEN=t skyra serve --fake &
SKYRA_TEST_WS=ws://127.0.0.1:18432/ws SKYRA_TEST_TOKEN=t swift test
```
