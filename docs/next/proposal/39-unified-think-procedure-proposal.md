# 39 — Unified think: model as a parameter, not a fork

Date: 2026-10-07

Status: **proposed.** Builds on [38](38-think-procedure-proposal.md) (think
as the first procedure) and the Claude provider integration on
`feat/living-wallpaper-ui`.

## The problem

Think exists twice: `think` (Codex) and `think-claude` (Claude), identical
workflow bundles that differ only in actor name and reach. The user target
holds a procedure table mapping each title to its actor. Adding a model means
adding a procedure, a bundle, a table entry, and a contract key — a
combinatorial fork for what should be a parameter.

The narrow runs on every turn. `system.narrow` is a syscall that sets the
Horizon; it should run once, when the conversation's authority is established,
not on every message the human sends. Every turn today pays for a narrow that
derives the same Horizon its predecessor already derived.

History travels on the wire as a growing `context` array: the user target
serializes the full transcript and sends it as a parameter. Each adapter
deserializes it and renders it for the model. The conversation grows linearly
in wire payload. The adapters should own their own memory; the wire should
carry only the message.

## What changes

### 1. The model is a parameter

The user's selection carries a model name, not a procedure title. One
procedure, `think`, takes `(text, model)`. The user target reads the model
from the selection and passes it through:

```json
{ "think": { "text": "hello", "model": "claude" } }
```

`model` is `"claude"` or `"codex"` (the default when absent). The procedure
dispatches to the right adapter. `think-claude` as a separate procedure, its
bundle, and its entry in the user target's procedure table go away.

### 2. The procedure is a target, not a workflow

The template engine has no conditionals. A workflow bundle cannot branch on
an input parameter to choose between two actors with different reach. The
procedure becomes a compiled Go target — the same pattern as the `user`
target — that reads `(text, model)` and emits the invocation to the right
adapter.

The target is a continuation: it receives steps (the adapter's return) and
emits the next action (show the transcript, wait for the human). It replaces
the interpreter's role for the think workflow with a purpose-built binary
that knows both adapters.

### 3. Narrow creates a Horizon once, not per turn

`system.narrow` creates a new Horizon — a child of the caller's, with
authority reduced to the adapter's reach. This happens once, when the
conversation starts: the first turn narrows, the narrow creates the
Horizon, and the reference to it is stored in the procedure target's
continuation state. Every subsequent turn invokes the adapter under that
Horizon directly, without re-deriving it.

If the model changes mid-conversation (the user switches from Codex to
Claude), the procedure narrows again — a new Horizon with the new actor's
reach. This is the only case where narrow runs after the first turn.

### 4. History stays local to the adapter

The wire carries `text` and a conversation identifier. The adapters own
their conversation memory:

- **Claude:** the session ID. The Claude CLI maintains server-side
  conversation history; `--resume <session>` reconnects to it. The adapter
  persists the session ID in the conversation's state directory
  (`.skyra/<actor>/`), keyed by conversation. No transcript travels on the
  wire; the CLI replays its own context.

- **Codex:** the context observation. Codex threads are one-shot (38: "one
  thread per invocation"), so context is serialized from the lineage. This
  does not change yet — Codex's adapter already renders `context` as
  "Earlier in this conversation:" — but the serialization moves from the
  user target into the Codex adapter. The user target no longer builds or
  sends the transcript array as a parameter.

The conversation contract between the user and the procedure thins to
`(text, model, images)`. The adapter's own state carries everything else.

### 5. Models keep their return; history is the adapter's concern

The adapter returns the answer text directly, as it does today. The
kernel records it as a step in the lineage — the single source of truth.
The procedure and user target read the answer from the step value,
unchanged.

History — what the model needs to see on the next turn — is the
adapter's local concern:

- **Claude:** `--resume <session>`. The CLI maintains server-side
  conversation history. The adapter persists one session ID in
  `.skyra/claude/`, keyed by conversation. No history on the wire.

- **Codex:** the thread ID, when the app-server is alive. The adapter
  already reads `Input.Thread` and `Input.Incarnation` — when present,
  it skips `thread/start` and calls `turn/start` on the existing
  thread, which carries its full conversation history. But Codex
  threads are in-memory: they do not survive app-server restarts.
  `thread/loaded/list` lists loaded threads, not persisted ones.
  When the server restarts (idle timeout, crash, `codex.stop`), the
  thread is gone and a new one must be created.

  Two options:
  1. **Keep the server alive** for the conversation's lifetime. The
     backoff holder (proposal 38, "Connections are edges") already
     describes this: a holder pings the app-server at growing intervals,
     resets on each invocation, and stops it after a maximum. With the
     server alive, the thread persists and context travels server-side.
  2. **Fall back to context on the wire** when the thread is lost. The
     adapter persists `(thread, incarnation)` locally; on the next turn,
     it inspects (`thread/loaded/list`). If the thread is loaded, it
     resumes. If not, it starts a fresh thread and falls back to
     context from the procedure target's continuation steps.

  Option 1 is the steady state; option 2 is the graceful degradation.
  Both are needed: the server will restart eventually, and the
  conversation should not lose history when it does.

The wire thins to `(text, model, images)` for any adapter that owns its
history server-side. No local history file that duplicates the lineage.

## What stays the same

- The prompt: `"You are inside Skyra, a cognitive architecture, fulfilling:
  think"` — the same for both adapters.
- The tool set: `echo`, `read`, `find`, `roots`, `write`, plus the
  adapter's own transport operations (claude.start/exchange or
  codex.start/exchange). Attached through the narrow's reach, as 38 says.
- The Horizon: authority derived from state plus proposal at every step.
  The narrow is still `system.narrow`; it just runs once.
- The user target: still a continuation that manages the human's turns. It
  no longer builds the transcript for the adapter; it still shows the
  transcript to the browser.
- The constellation UI: stars are contracts. The model parameter changes
  which adapter lights up, not which constellation the user clicks.

## Steps

1. Add `model` to the think contract. The user target reads it from the
   selection and passes it through. Both bundles still work; the field is
   ignored until step 3.

2. Wire Claude session persistence. The claude adapter reads/writes a
   session ID in `.skyra/claude/`, keyed by conversation. `--resume`
   reconnects. Claude sees conversation history natively.

3. Replace the think workflow bundle with a compiled target. The target
   reads `model`, narrows once (stored in its continuation state), and
   dispatches to the right adapter. The think-claude bundle, its
   procedure-table entry, and the skyra script's clone of the think
   procedure go away.

4. Move transcript building out of the user target. The Codex adapter
   reads the lineage for context; the Claude adapter uses `--resume`. The
   wire payload thins to `(text, model, images)`.

5. The return becomes an ack. The adapters stamp their results into state;
   the procedure returns metadata, not content.

## What this does not do

- Change the Codex adapter's one-thread-per-invocation model.
- Introduce compaction or context windowing. That is think.md's later work.
- Add more models. But the pattern is: add an adapter, add a case in the
  procedure target. No new bundle, no new procedure.
- Change the kernel. The narrow, the Horizon, the recorded edge — all
  unchanged.
