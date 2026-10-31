# Ableton Live adapter research for Skyra

Researched October 31, 2026. Research and proposed design only; nothing installed or connected.

## Why

The motivating job came from a real mashup session (Kids See Ghosts "Reborn" vocals over Polyphia "LET THERE BE LIGHT", with a dubstep last drop built from Space Laces "Choppaz" and Eptic "Wall of Death"). Analysis of the exported mix found the last drop about 3–4 dB quieter than the first, about 8 dB short in the low mids, and with its bass layer roughly 40 cents flat against a guitar layer roughly 18 cents sharp. That half-semitone gap is the classic result of fitting a clip to tempo by speed change (Live's Re-Pitch warp mode) instead of a pitch-preserving warp.

Every step of that diagnosis and fix touches state that lives in the Live Set: which clips are warped, in which mode, at what `pitch_coarse` and `pitch_fine`, and what sits on each track. An agent that can read that state, and change a narrow set of clip properties within a granted boundary, would have found and fixed the problem directly instead of inferring it from a rendered file.

## Recommendation

Do not connect any community Ableton MCP server to a model directly. Follow the Blender pattern: Skyra-owned typed operations behind a private bridge, read-only first.

1. **Milestone 1, offline and read-only:** parse a *copy* of the saved `.als` file (gzip-compressed XML) for tracks, clips, warp mode, transpose and detune, and analyze an exported WAV for loudness, band balance, key and tuning per section. Neither needs Live running or any socket.
2. **Milestone 2, live and bounded:** put the typed, no-arbitrary-code server (wstierhout/ableton-live-mcp) behind a Skyra bridge, exposing only an allowlisted subset.
3. **Later:** reassess Ableton's own Extensions SDK once it leaves pre-release. It is the only route that runs inside Live without a Remote Script.

## Candidates

| Candidate | Shape | Fit | Tradeoff |
| --- | --- | --- | --- |
| [wstierhout/ableton-live-mcp](https://github.com/wstierhout/ableton-live-mcp) (`mcp-server-ableton-live` on PyPI) | Remote Script on `127.0.0.1:9877` + stdio MCP server; Live 11/12 | Preferred live candidate. States there is no arbitrary-code path; typed tools; toolsets restrictable via `ABLETON_TOOLSETS`; offline tools read `.als`/`.adg`/`.adv` | Unauthenticated loopback socket; very small community (2 stars at review); README tool count is inconsistent (104 vs 154 vs 156) |
| [ulm0/ableton-live-mcp](https://github.com/ulm0/ableton-live-mcp) | Runs inside Live via the official Ableton Extensions SDK; streamable HTTP at `127.0.0.1:8722/mcp` | Most structurally sound: no bridge process, no Remote Script | Needs Live 12.4.5+; Extensions currently require Suite and the pre-release program; 0 stars, 7 commits; no auth described; `render_track_audio` and `import_file` write files |
| [ahujasid/ableton-mcp](https://github.com/ahujasid/ableton-mcp) | Remote Script socket + `uvx ableton-mcp`; Live 10+ | Most adopted (about 3.2k stars), MIT | Anonymous telemetry on by default (`ABLETON_MCP_DISABLE_TELEMETRY=true` to disable); socket executes received commands with no described auth |
| [bschoepke/ableton-live-mcp](https://github.com/bschoepke/ableton-live-mcp) | Remote Script with Python execution | Not a fit | Advertises arbitrary Python execution inside Live and warns it can corrupt sets |
| LivePilot, Simon-Kansara (AbletonOSC), xiaolaa2 ableton-copilot | Remote Script / OSC / ableton-js | Not preferred | Larger or older surfaces; more dependencies to audit |

These were reviewed at README and listing level, not installed, source-audited or benchmarked. The ranking is an architectural judgment.

## Facts relevant to the boundary

- Live Object Model audio clips expose `warping`, `warp_mode` (0 Beats, 1 Tones, 2 Texture, 3 Re-Pitch, 4 Complex, 5 REX, 6 Complex Pro), `pitch_coarse` (−48..48 semitones) and `pitch_fine` (−50..49 cents). Re-Pitch plus a tempo change is exactly what produced the fractional detune above. ([Cycling '74 LOM Clip reference](https://docs.cycling74.com/apiref/lom/clip/))
- wstierhout's README states Live's API has no render or freeze, so export stays manual. Audio analysis therefore needs an exported file as input, which suits a read-only first milestone.
- Every Remote Script candidate listens on an unauthenticated loopback port. Loopback limits network exposure but is not authentication: any local process can drive Live. Skyra must deny other actors access to that port and grant it only to the bridge.
- Remote Scripts run inside Live's process with Live's full user permissions. Skyra's per-call sandbox does not constrain Live itself, so the boundary is the set of operations the bridge forwards, not prompt rules.

## Proposed Skyra composition

think → ableton procedure → narrowed agent → ableton.tools → private Skyra bridge → (offline parser | pinned live MCP server → Remote Script)

Briefs, reports and analysis output: `ideas/design/ableton/`. Copies of `.als` files and exported audio: a bridge-owned job directory. The original set and its samples are never written.

| Operation | Milestone | Purpose |
| --- | --- | --- |
| `context` | 1 | Report source (offline copy or live), set name, tempo, time signature, track count, and a set version for edit guards |
| `inspect_set` | 1 | Tracks, clips, devices, and per-clip warp/transpose/detune |
| `inspect_clip` | 1 | One clip by id, including warp markers |
| `analyze_audio` | 1 | Exported file → loudness per section, band energy, key and tuning per band, tempo |
| `set_clip_pitch` | 2 | `pitch_coarse` / `pitch_fine` within LOM ranges; requires `expected_set_version` |
| `set_clip_warp` | 2 | `warping` and `warp_mode`; requires `expected_set_version` |
| `set_mixer` | 2, later | Track volume/pan only |

Do not expose device insertion, file import, render-to-path, note editing, arbitrary commands or any upstream passthrough in the first live milestone. Pin the server version (for example `mcp-server-ableton-live@1.8.1`) and the Remote Script together. Mark dispatched edit timeouts as unknown and re-inspect before any retry.

## Acceptance sequence

1. Offline: parse a fixture `.als` copy and match tracks, clips, warp modes and detune against values set by hand in Live. Run `analyze_audio` on a known export and reproduce the findings above (last-drop bass about −40 cents; drop loudness 3–4 dB below the main section).
2. Live, read-only: install the pinned server, confirm `tools/list` against the allowlist, and confirm `context` and `inspect_set` match the offline parse of the same set.
3. Live edits: on a copied set, change one clip from Re-Pitch to Complex Pro and set `pitch_fine`; reopen and verify persistence; test wrong-set rejection and the stale-version guard; confirm offline record replay does not dispatch edits again.

Not yet proven: `.als` schema stability across Live 11/12 minor versions, the actual tool list and argument validation of the pinned server, Extensions SDK availability outside pre-release, or any isolation of Live itself on this Mac.
