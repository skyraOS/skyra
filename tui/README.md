# Skyra TUI

`skyra tui` runs the stateless controller's conversation backend and opens a
terminal client. `skyra tui --fake` uses the deterministic provider. It uses
the same conversation directories and browser WebSocket protocol as `skyra`;
the terminal client is not a registered Skyra surface and holds no frame
descriptors or runtime authority.

```text
skyra tui             real Codex
skyra tui --fake      no model or login
```

| Key | Action |
|---|---|
| Enter | Send a message; the first message opens a fresh session |
| Ctrl+N | Clear the view so the next message starts another fresh session |
| Page Up / Page Down | Scroll the transcript |
| Ctrl+U | Clear the composer |
| Ctrl+D | Close the current session |
| Ctrl+C | Clear the composer; the transcript stays |
| Ctrl+C twice | Quit the TUI (on an empty composer) |

The TUI owns its webhost process and leaves the Codex actor host running, like
the browser command. `skyra stop` stops that actor host. Webhost and runtime
logs go to `~/.skyra/tui-webhost.log` (or `SKYRA_STATE/tui-webhost.log`) so they
cannot overwrite the terminal screen. Run one Skyra UI on a WebSocket port at
a time; set `SKYRA_WS_PORT` to use another port.

The UI displays one current session: transcript, phase and connection status.
It starts with an empty view even when the backend has saved conversations.
The first submitted message opens a new conversation in the backend; later
messages continue it. Ctrl+N clears the view for a new session without deleting
the previous session's saved lineage. A future lineage pane can show calls and observation IDs
once the webhost exposes a read-only lineage projection. The TUI does not
read `.skyra/` itself.

The dim constellation wallpaper is drawn from terminal glyphs and fills the
whole terminal; everything else sits on top of it. It is almost still: stars,
mostly cool with a few warm or violet, breathe on slow periods of their own,
and now and then one glints over eight seconds. Three satellites with short
tails circle a small core star on undrawn orbits, a single braille dot at a
time: easy to miss at a glance, easy to follow once you look. Behind them the
background rises a few color steps in two faint lights, a teal one breathing
on the core and a violet one circling it every eight minutes, and every
ninety seconds a meteor crosses the whole screen in three seconds. The name
is a faint star-chart label that brightens as the middle satellite goes by.

The camera sways slowly, a few characters at most over several minutes, and
the stars follow it by depth: faint dust barely moves, bright stars move the
most, and each star steps at its own moment, so the sky parallaxes instead of
shifting as one. The core, its satellites and the name hold still in front.

The header, chat messages and composer are translucent panels over the sky,
so stars and satellites show through them, dimmed; the footer's text sits
directly on the sky.

During an active agent turn, typed text remains a draft until the session is
ready. Closed sessions stay saved in the backend, but the TUI does not browse
them.
