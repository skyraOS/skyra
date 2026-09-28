# Startup

Everything needed to get Skyra running on a fresh machine. **Keep this current:**
when anything about setup changes, update this file in the same commit.

**Current setup: 2026-09-28.** macOS 15.6 (Apple Silicon), Go 1.25.0,
Node 26.0.0, Python 3.9, Codex CLI **0.157.1**.

The controller now launches Codex under external Seatbelt with direct dynamic
tools, isolated configuration, and no code-mode helper. Live conversation,
runtime echo execution, interruption and recovery passed on September 27.
A captured inference request contains only the direct `functions.echo` tool.
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
| **Go 1.25+** | builds everything; standard library only, no other Go dependencies | `brew install go` |
| **Node 20+** | browser UI (Next.js) | `brew install node` |
| **Python 3.9+** | provider configuration and sandbox launcher | `brew install python` |
| **Codex CLI 0.157.1**, logged in | verified direct-tool configuration for real chat | `npm install -g @openai/codex@0.157.1`, then `codex login` |

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

This builds the stateless controller, starts the Codex actor host (once; it
keeps running), starts the browser host and the UI, and opens
http://localhost:3000. Press **+ new** to start a conversation. The terminal
shows runtime, frontend and provider logs. Provider logs also remain on disk.
Ctrl-C stops the browser host and the UI; `skyra stop` ends the Codex host.

| Command | Does |
|---|---|
| `skyra` | **chat with real Codex in the browser, through the stateless controller.** Every conversation is saved as its lineage and listed with its history |
| `skyra chat --fake` | the same, against a deterministic fake Codex: no model, no login |
| `skyra tui` | a fresh terminal session backed by the conversation protocol, without starting Next.js ([keys and layout](tui/README.md)) |
| `skyra tui --fake` | terminal UI with the deterministic provider |
| `skyra ask "TEXT"` | one question to real Codex; prints the lineage, the replay check and the answer |
| `skyra replay DIR` | re-derive a finished conversation or run from its records |
| `skyra stop` | stop the Codex actor hosts that `skyra` and `ask` leave running |
| `skyra dna19` | DNA-19, the predecessor: its browser chat through its capability fence (`--fake`, `--echo`) |
| `skyra demo` | **the stateless controller**: every scenario (a continuation calling a tool, fake Codex through the actor host, a fenced native command, the sandbox probe, a misbehaving leaf, a contract violation), each ending in a replay check |
| `skyra test` | every Go test, for the controller and DNA-19 |
| `skyra live` | DNA-19's live suite against your installed Codex (about a minute; rewrites the evidence in `history/dna19/artifacts/`) |
| `skyra doctor` | check prerequisites and ports |
| `skyra help` | list the commands |

After updating, exit the running UI with Ctrl-C, run `skyra stop`, then run
`skyra` again. This rebuilds binaries and starts the provider with the current
configuration. Start a new conversation when testing tool changes.

### One question, from the terminal

```sh
skyra ask "Call the echo tool with the text hi, then tell me what it returned."
skyra ask "What tools can you call?"                       # only echo is exposed
skyra stop
```

Actor hosts, logs and state live in `~/.skyra/`: `codex-real-direct-v2.sock` and
`codex-real-direct-v2-host.log` for the Codex host, `conversations/` for chats (one
`.skyra` per conversation), `runs/` for `ask`.

## 4. What to try in a conversation

- **"Call the echo tool with the text hi, then tell me what it returned."**
  Codex asks for `echo`; the runtime derives an invoke, runs echo as its own
  sandboxed leaf, and a fresh adapter step answers Codex's pending call.
- **Then: "What did the echo tool return a moment ago?"** One Codex thread per
  conversation, so it remembers.
- **"What tools can you call?"** Only the registered `echo` tool is exposed.
  A shell request may be declined by the model. OS probes and deterministic
  native-attempt tests verify enforcement independently of model behavior.
- **close**, then pick the conversation from the sidebar: its history is
  derived from its records. `skyra replay ~/.skyra/conversations/<id>/.skyra`
  re-derives every record.

**Direct tool calling.** New launches expose registered Skyra tools directly
(currently `functions.echo`), without the `exec`/`wait` JavaScript wrapper.
The launcher creates a read-only model catalog selecting standard tool mode,
disables built-in goals, clock, questions and collaboration, and suppresses
skill discovery. The code-mode helper has no execute grant. The adapter also
rejects calls naming tools absent from its registered tool set.

This configuration is verified for Codex **0.157.1** and requires that version's
`~/.codex/models_cache.json`. Missing or mismatched metadata fails startup;
refresh it using the installed Codex outside Skyra. A Codex upgrade requires
rerunning `controller/experiments/appserver/tool_surface.py` and validating the
launch policy before changing the version guard. Existing conversations may
retain old descriptions in their history; start a new conversation to inspect
the new tool surface.

**Provider isolation.** New `skyra` and `skyra ask` launches require an external
macOS Seatbelt profile, with isolated configuration and scratch space. Shell
execution and file access outside the granted paths are blocked by the OS.
`HOME`, `CODEX_HOME` and `TMPDIR` point into that isolated provider run.
The adapter interrupts native attempts and confirms that the affected turn
has ended before emitting the horizon fault. An unconfirmed interruption
produces a custody fault instead. Threads retain `readOnly`/`never` policy.

The provider still has broad outbound IP access for inference, can read its
login and granted system files, and can write its temporary state and a system
certificate cache lock. This is not a network isolation guarantee. Auth refresh
is not supported by the read-only shared login grant; refresh the login outside
Skyra and restart if needed. Python 3 is required for the launcher.

Already running sessions are not migrated. Run `skyra stop`, then `skyra` to
restart with the new provider. New launches use a separate
`codex-real-direct-v2` host namespace and never reuse the old unsandboxed host.
Provider launch state is retained under `~/.skyra/providers/`; the login is a
symlink, not a credential copy. `SKYRA_STATE` can select an isolated state root.

DNA-19's chat is still there as `skyra dna19`; its fence retires the whole
conversation on any native attempt.

## 5. Troubleshooting

The sidebar's trash button deletes a chat and its local Skyra history after
confirmation. It is disabled while the agent is working. Deleting a waiting
chat releases its parked frames; a small deletion marker prevents late messages
or a restart from restoring it. This does not erase the separate Codex provider
transcript or host diagnostic logs.

| Symptom | Fix |
|---|---|
| `skyra: command not found` | open a new terminal; or add `export PATH="$HOME/.local/bin:$PATH"` to `~/.zshrc`; or run `./skyra` from the repo |
| `port 8080 is in use` / `port 3000 is in use` | stop the other process, or `SKYRA_WS_PORT=8090 SKYRA_UI_PORT=3001 skyra` |
| `codex is not logged in` | `codex login`, or use `skyra chat --fake` |
| a turn shows `⚠ the agent faulted (custody)` | the Codex adapter's last stderr line is the cause; `skyra stop`, then run `skyra` again. `… actor is not alive` means Codex exited under a host that kept running; `skyra` now restarts such a host on launch |
| `skyra dna19`: the first message fails with `contamination:` | a newer Codex emits an event DNA-19's fence doesn't know. Run `skyra live` to see which; the allowlist is in `history/dna19/codex/events.go` |
| the UI says `open` but stays on "waiting for the interaction to open…" | the UI is dropping the server's snapshots as invalid. Open the browser console and look for `schema_invalid` drops: the server and `frontend/lib/protocol/schemas.ts` disagree (fixed 2026-09-26 for empty lists sent as `null`) |
| the UI loads but never connects | the browser host isn't up: check the terminal for an error above the trace |
| a conversation stays grey after a restart | expected: its waiting frame lived in the browser host that stopped. It is viewable, and replayable, but cannot continue yet |
| the UI fails to start | read `build/ui.log`; rerun `./install.sh` to reinstall the npm dependencies |
| `the Codex actor host did not start` | read `~/.skyra/codex-real-host.log`; `skyra stop`, then try again |
| anything else | `skyra doctor` |

## 6. Keeping this current

Before telling anyone to run this:

1. `git pull`, then `skyra doctor` and `skyra test`.
2. `skyra demo`: every scenario ends in `byte-identical`.
3. `skyra chat --fake`, then `skyra`: the conversation prompts above; then `skyra ask`.
4. If Codex was upgraded: `skyra live`, then update the versions and date in
   **Last verified** above.
