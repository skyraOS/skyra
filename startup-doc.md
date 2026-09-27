# Startup

Everything needed to get Skyra running on a fresh machine. **Keep this current:**
when anything about setup changes, update this file in the same commit.

**Last verified: 2026-09-27** on macOS 15.6 (Apple Silicon), Go 1.25.0,
Node 26.0.0, Codex CLI 0.157.1:

- `skyra` (the controller chat): a two-turn conversation with real Codex;
  Codex remembered turn one in turn two; closed, then replayed (22 records,
  byte-identical). `skyra chat --fake`: a tool call, a fenced native command
  shown as a turn with the conversation going on, then close.

- `./install.sh` and `skyra doctor` pass.
- `skyra demo`: all 7 scenarios end in `byte-identical`, including fake Codex
  through the actor host.
- `skyra ask` with real Codex: an echo tool call ran as its own sandboxed leaf
  and resumed the turn; asking it to run `ls` was derived as a horizon fault; a
  second ask reused the same actor host.
- `skyra` with real Codex: a browser message made Codex call the `echo` tool,
  which ran as a separate host process and resumed on the same turn.
- `skyra chat --fake`, `skyra demo` and `skyra test` all pass.
- `skyra live`: all 6 scenarios pass.

## 1. Prerequisites

| Need | Why | Install |
|---|---|---|
| **macOS** | the sandbox is Seatbelt (`sandbox-exec`) | — |
| **Go 1.25+** | builds everything; standard library only, no other Go dependencies | `brew install go` |
| **Node 20+** | DNA-19's browser UI (Next.js) | `brew install node` |
| **Codex CLI**, logged in | only for `skyra` with a real model | `brew install codex`, then `codex login` (a ChatGPT login works) |

Without Codex, everything else still works: use `skyra chat --fake`.

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
shows the runtime's trace: every record committed, every park and resume.
Ctrl-C stops the browser host and the UI; `skyra stop` ends the Codex host.

| Command | Does |
|---|---|
| `skyra` | **chat with real Codex in the browser, through the stateless controller.** Every conversation is saved as its lineage and listed with its history |
| `skyra chat --fake` | the same, against a deterministic fake Codex: no model, no login |
| `skyra ask "TEXT"` | one question to real Codex; prints the lineage, the replay check and the answer |
| `skyra replay DIR` | re-derive a finished conversation or run from its records |
| `skyra stop` | stop the Codex actor hosts that `skyra` and `ask` leave running |
| `skyra dna19` | DNA-19, the predecessor: its browser chat through its capability fence (`--fake`, `--echo`) |
| `skyra demo` | **the stateless controller**: every scenario (a continuation calling a tool, fake Codex through the actor host, a fenced native command, the sandbox probe, a misbehaving leaf, a contract violation), each ending in a replay check |
| `skyra test` | every Go test, for the controller and DNA-19 |
| `skyra live` | DNA-19's live suite against your installed Codex (about a minute; rewrites the evidence in `history/dna19/artifacts/`) |
| `skyra doctor` | check prerequisites and ports |
| `skyra help` | list the commands |

`skyra` always rebuilds first, so after a `git pull` just run it again.

### One question, from the terminal

```sh
skyra ask "Call the echo tool with the text hi, then tell me what it returned."
skyra ask "Run the ls command and tell me what it says."    # derived as a horizon fault
skyra stop
```

Actor hosts, logs and state live in `~/.skyra/`: `codex-real.sock` and
`codex-real-host.log` for the Codex host, `conversations/` for chats (one
`.skyra` per conversation), `runs/` for `ask`.

## 4. What to try in a conversation

- **"Call the echo tool with the text hi, then tell me what it returned."**
  Codex asks for `echo`; the runtime derives an invoke, runs echo as its own
  sandboxed leaf, and a fresh adapter step answers Codex's pending call.
- **Then: "What did the echo tool return a moment ago?"** One Codex thread per
  conversation, so it remembers.
- **"Run the ls command and tell me what it says."** Codex tries its native
  shell; the runtime derives a horizon fault, shows it as a turn, and the
  conversation goes on.
- **close**, then pick the conversation from the sidebar: its history is
  derived from its records. `skyra replay ~/.skyra/conversations/<id>/.skyra`
  re-derives every record.

**What it cannot do, by design.** Codex's threads are read-only, with no
network, `approvalPolicy=never`, and exactly the tools the user side holds
(`echo`). Anything native it tries is a horizon fault.

DNA-19's chat is still there as `skyra dna19`; its fence retires the whole
conversation on any native attempt.

## 5. Troubleshooting

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
