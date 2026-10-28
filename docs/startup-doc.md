# Startup

Everything needed to get Skyra running on a fresh machine. **Keep this current:**
when anything about setup changes, update this file in the same commit.

**Current setup: 2026-09-28.** macOS 15.6 (Apple Silicon), Go 1.25.0,
Node 26.0.0, Python 3.9, Codex CLI **0.160.0**.

The service host launches one Codex app-server per turn under the Seatbelt
profile its program registration declares, with direct dynamic tools, private
state and an allowlisted egress proxy (proposal 49). The adapter reaches it
through `service-io` by the session's handle, and every exchange is recorded. Live conversation, runtime echo execution and replay
passed on September 28. A captured inference request contains only the direct
`functions.echo` tool. The `read` tool was added after that capture: it is
verified with the fake provider, not yet against live Codex (rerun
`tool_surface.py`).
The sidebar supports deletion of idle chats and their local history.

Validation commands (run macOS sandbox tests outside another sandbox):

```sh
skyra build
skyra doctor
skyra test
cd frontend && npm run typecheck
```

For provider-specific checks, from the repository root:

```sh
python3 controller/experiments/appserver/integration.py
python3 controller/experiments/appserver/tool_surface.py
python3 controller/experiments/appserver/integration.py --live  # real model usage
```

## 1. Prerequisites

| Need | Why | Install |
|---|---|---|
| **macOS** | the sandbox is Seatbelt (`sandbox-exec`) | — |
| **Go 1.25+** | builds the controller and interpreter; the latter downloads its CEL dependencies | `brew install go` |
| **Node 20+** | browser UI (Next.js) | `brew install node` |
| **Python 3.9+** | provider configuration and sandbox launcher | `brew install python` |
| **Codex CLI 0.160.0**, logged in | verified direct-tool configuration for real chat | `npm install -g @openai/codex@0.160.0`, then `codex login` |

After login, start `codex` once to fetch model metadata, then exit it.
`skyra doctor` checks for a file-based login and matching `~/.codex/models_cache.json`.
The real-provider launcher refuses unvalidated Codex versions; do not upgrade it
without validating the direct-tool configuration. Without Codex, use
`skyra chat --fake` for the deterministic provider.

## 2. Install (once)

```sh
git clone https://github.com/skyraOS/tony-stark-was-able-tobuild-this-in-a-cave.git skyra
cd skyra
./install.sh
```

`install.sh` installs the UI's npm dependencies, builds every binary into
`build/`, links the `skyra` command onto your `PATH` (into `~/.local/bin`, or
the first writable directory already on `PATH`), then runs `skyra doctor`. If it
prints an `export PATH=…` line, add that to `~/.zshrc` and open a new terminal.

## 3. Run

```sh
skyra
```

This builds the stateless controller, starts the browser host and the UI, and opens
http://localhost:3000. Press **+ new** to start a conversation. The terminal
shows runtime and frontend logs. Provider state and proxy journals remain on disk.
Ctrl-C stops the browser host and the UI; parked connections close with it.

| Command | Does |
|---|---|
| `skyra` | **chat with real Codex in the browser, through the stateless controller.** Every conversation is saved as its lineage and listed with its history |
| `skyra chat --fake` | the same, against a deterministic fake Codex: no model, no login |
| `skyra tui` | a fresh terminal session backed by the conversation protocol, without starting Next.js ([keys and layout](../tui/README.md)) |
| `skyra tui --fake` | terminal UI with the deterministic provider |
| `skyra app` | build and open `build/Skyra.app`, the desktop client; needs Xcode ([desktop/README.md](../desktop/README.md)) |
| `skyra serve [--fake]` | the backend alone, as the app starts it; the launcher generates `SKYRA_WS_TOKEN` unless supplied; the webhost always requires it |
| `skyra ask "TEXT"` | one question to real Codex; prints the lineage, the replay check and the answer |
| `skyra replay DIR` | re-derive a finished conversation or run from its records |
| `skyra stop` | stop actor hosts left by older Skyra versions |
| `skyra demo` | **the stateless controller**: every scenario (a continuation calling a tool, fake Codex through the actor host, a fenced native command, the sandbox probe, a misbehaving leaf, a contract violation), each ending in a replay check |
| `skyra test` | all deterministic checks: Go race/vet, interpreter, frontend lint/types/build and mock/real-host protocol, Swift, website replay/WASM and demos |
| `skyra build` | build every binary into `build/` |
| `skyra doctor` | check prerequisites and ports |
| `skyra help` | list the commands |

After updating, exit the running UI with Ctrl-C, then run `skyra` again. This
rebuilds the binaries. Start a new conversation when testing tool changes.

### One question, from the terminal

```sh
skyra ask "Call the echo tool with the text hi, then tell me what it returned."
skyra ask "What tools can you call?"                       # echo and read
skyra ask "Read README.md and summarize it."                # read, inside the workspace
```

State lives in `~/.skyra/`: `conversations/` holds chats (one `.skyra` per
conversation), `runs/` holds `ask`, and each resident has a private directory
under `providers/` with its proxy journal and diagnostics.

## 4. What to try in a conversation

- **"Call the echo tool with the text hi, then tell me what it returned."**
  Codex asks for `echo`; the runtime derives an invoke, runs echo as its own
  sandboxed leaf, and a fresh adapter step answers Codex's pending call.
- **Then: "What did the echo tool return a moment ago?"** One Codex thread per
  conversation, so it remembers.
- **"Read README.md and tell me what this project is."** Codex asks for `read`;
  the leaf runs under a sandbox that grants only the workspace. Paths outside
  it, including symlinks that point out, come back as `"error":"denied"`.
- **"What tools can you call?"** Only the registered `echo` and `read` tools are exposed.
  A shell request may be declined by the model. OS probes and deterministic
  native-attempt tests verify enforcement independently of model behavior.
- **close**, then pick the conversation from the sidebar: its history is
  derived from its records. `skyra replay ~/.skyra/conversations/<id>/.skyra`
  re-derives every record.

**Direct tool calling.** New launches expose registered Skyra tools directly
(currently `functions.echo` and `functions.read`), without the `exec`/`wait` JavaScript wrapper.
The launcher creates a read-only model catalog selecting standard tool mode,
disables built-in goals, clock, questions and collaboration, and suppresses
skill discovery. The code-mode helper has no execute grant. The adapter also
rejects calls naming tools absent from its registered tool set.

This configuration is verified for Codex **0.160.0** and requires that version's
`~/.codex/models_cache.json`. Missing or mismatched metadata fails startup;
refresh it using the installed Codex outside Skyra. A Codex upgrade requires
rerunning `controller/experiments/appserver/tool_surface.py` and validating the
launch policy before changing the version guard. Existing conversations may
retain old descriptions in their history; start a new conversation to inspect
the new tool surface.

**Provider isolation.** New `skyra` and `skyra ask` launches require an external
macOS Seatbelt profile rendered from the registered Horizon. Each Codex process
has private configuration, work and scratch paths. Shell execution and file
access outside the granted paths are blocked by the OS. `HOME`, `CODEX_HOME`
and `TMPDIR` point into that isolated provider run.
The adapter interrupts native attempts and confirms that the affected turn
has ended before emitting the horizon fault. An unconfirmed interruption
produces a custody fault instead. Threads retain `readOnly`/`never` policy.

The provider can connect only to its local egress proxy; that proxy accepts
HTTPS tunnels to the registered host list. It can read its login and granted
system files and write its private state. Auth refresh is not supported by the
read-only shared login grant; refresh the login outside Skyra and start a new
conversation if needed. Python 3 runs the egress proxy.

Already running sessions are not migrated. Exit the UI and start a new
conversation to use the new provider.
Provider launch state is retained under `~/.skyra/providers/`; the login is a
symlink, not a credential copy. `SKYRA_STATE` can select an isolated state root.

**The workspace.** `read` is granted one directory: where you started `skyra`,
or `SKYRA_WORKSPACE` if set. That grant is the Horizon you approve by starting
it; every read inside it runs with no prompt, and anything outside is denied by
the OS. The file contents a read returns go to the model. `skyra` refuses a
workspace that is your home directory, `/`, or that overlaps `~/.skyra` or
`~/.codex`, because the sandbox floor protects only the current conversation's
records. A read returns at most 2000 lines or 64 KiB; every read is recorded in
the conversation's lineage.

## 5. Troubleshooting

The sidebar's trash button deletes a chat and its local Skyra history after
confirmation. It is disabled while the agent is working. Deleting a waiting
chat releases its parked frames and resident connection; a small deletion marker
prevents late messages or a restart from restoring it. Provider diagnostics
remain under `providers/`.

| Symptom | Fix |
|---|---|
| `the workspace … overlaps …` or `cannot be the workspace` | `cd` into a project directory before `skyra`, or `SKYRA_WORKSPACE=/path/to/project skyra` |
| `skyra: command not found` | open a new terminal; or add `export PATH="$HOME/.local/bin:$PATH"` to `~/.zshrc`; or run `./skyra` from the repo |
| `port 8080 is in use` / `port 3000 is in use` | stop the other process, or `SKYRA_WS_PORT=8090 SKYRA_UI_PORT=3001 skyra` |
| `codex is not logged in` | `codex login`, or use `skyra chat --fake` |
| a turn shows `⚠ the agent faulted (custody)` | check the runtime trace and the conversation's lineage; start a new conversation after a resident connection fault |
| the UI says `open` but stays on "waiting for the interaction to open…" | the UI is dropping the server's snapshots as invalid. Open the browser console and look for `schema_invalid` drops: the server and `frontend/lib/protocol/schemas.ts` disagree (fixed 2026-09-26 for empty lists sent as `null`) |
| the UI loads but never connects | the browser host isn't up: check the terminal for an error above the trace |
| a conversation stays grey after a restart | expected: its waiting frame lived in the browser host that stopped. It is viewable, and replayable, but cannot continue yet |
| the UI fails to start | read `build/ui.log`; rerun `./install.sh` to reinstall the npm dependencies |
| anything else | `skyra doctor` |

## 6. Keeping this current

Before telling anyone to run this:

1. `git pull`, then `skyra doctor` and `skyra test`.
2. `skyra demo`: every scenario ends in `byte-identical`.
3. `skyra chat --fake`, then `skyra`: the conversation prompts above; then `skyra ask`.
4. If Codex was upgraded: `skyra live`, then update the versions and date in
   **Last verified** above.

## Ingress and recovery

Every launcher mode generates a fresh token unless `SKYRA_WS_TOKEN` is set.
The browser launch link carries it in `#skyra-token=...`; the client moves it
into session storage and removes the fragment. Use the launch link when
opening a new tab. Native clients send a Bearer header. Browser origins must
exactly match `SKYRA_WS_ORIGIN`, which `skyra chat` sets to its localhost UI.
Direct `webhost` launches require a token; unset tokens are refused.

Request admission is persisted before execution. Reusing an admitted ID with
different work is a conflict; reconnects receive status rather than replaying
commands. A host restart or runtime I/O interruption reports `delivery-unknown`.
Inspect the recorded history before sending that work again: partial external
effects may have happened. This is failure reporting, not automatic crash recovery.

New histories use genesis version 2 and store the entire emission for each
resolved round. Old histories still replay with terminal-state, uniqueness and
ownership checks, but cannot prove that every originally emitted slot is present.
