# The inference substrate — the LLM as a driver

**Date:** 2026-09-07
**Status:** DESIGN — decided this session. Applies [the driver boundary](drivers-and-the-boundary.md) to the
model. Names the Codex/ChatGPT integration boundary; the tool-encoding rules are decided, the adapter is unbuilt.

## Why Codex, and what it is to us

Anthropic does not currently allow subscription-backed auth for a third-party harness (inference bills to a
different budget); that route is deferred. **Codex does allow inference to draw from the ChatGPT subscription
budget** — better for users and for building this project — so it is the first inference substrate.

Codex handles the ChatGPT OAuth flow, token refresh, inference, thread lifecycle, and tool-call continuation.
The harness owns everything real: UI, memory, permissions, routing, persistence, and execution logic. Codex is
the **inference/turn substrate**, nothing more.

## The model is an entity with a driver at an address

The LLM is not a special case — it is an **entity + address + driver** like the filesystem or a remote CTU.
Calling the model is an `Exec` routed to the inference driver; the driver's inbound normalization renders the
Focus and context into the model's representation, and its **outbound normalization produces Move(s)** — the
egress invariant. A tool call becomes a `Call`; a final answer becomes a `Ret`. Parallel tool calls are several
`Call`s from one turn (1-to-N).

Its only distinguishing trait — that its output is *new Moves* rather than a return value — is not special:
it is exactly the runtime's Move-emitting loop, feeding `Call` ingress instead of `Ret` ingress.

**Consequence for the "MCP is annoying" problem:** the MCP tool shape lives entirely inside this one driver's
representation. Nothing MCP-shaped crosses inward, because the driver is the anti-corruption layer. The runtime
speaks `Move / Action / Params / Observation`; the adapter is the only place the transport encoding exists.

## The tool encoding: minimize what the model sees

Model-facing tools are `name / description / inputSchema`. Treat that triple as a **transport encoding**, not
your ontology. The mapping:

```text
name        = an ephemeral OPAQUE HANDLE (a per-turn UUID). ZERO semantics. unique within the Focus. throwaway.
description = the ACTION — the one instruction. runtime-governed length (160-token default, user-settable), UNCAPPED at transport.
inputSchema = the PARAMS.
(host-only)  opcode · EngramID = hash(Action + ordered Params) · InvocationID · Horizon/caps · tracing · callId map.
```

### Why the name is an opaque handle, not the Action

The transport `name` field is a constrained identifier (~64 chars, `[a-zA-Z0-9_-]`, no spaces). A rich Action
cannot be the name — and locking the Action's length to the transport's limit would let the transport dictate
the ontology, which is exactly the leak the driver boundary forbids. So:

- **The Action lives in `description`** (uncapped at transport), governed by a *runtime* policy — 160 tokens as
  a default the user can move. That cap is doing real work: the whole tool list is sent every turn, so it bounds
  tool-list context cost. It is runtime-enforced, not inherited from the transport.
- **The name is a throwaway handle.** It carries zero semantics deliberately: it cannot be misread as an
  instruction, it forces the model to read the real instruction (the description), and it is trivially unique.
- **Ephemeral per turn.** Focus is dynamic and Actions can be refined between turns; a stable name risks a
  "same handle, different meaning" association. Fresh-per-turn forces this turn's description to be authoritative.
  All cross-turn continuity comes from host-side logs (`name / callId → Action`), not the wire.

Rejected alternatives, for the record: a **truncated** Action as the name (a truncated imperative reads as a
different, often inverted instruction, and invites shortcutting past the description); and **numbering with
flavor** ("do this"/"do that" — opaque like an index but imperative-shaped, reintroducing the misread).

### callId ↔ MoveID

Codex assigns its own `callId`. Treat it as an opaque transport correlation ID; keep a private map
`callId → (MoveID/fd + InvocationID)` and return the `callId` on completion. A tool result arriving by `callId`
is an incoming `Ret` you resolve to its move and land into its register — the ingress→return path. Codex's
"tool-call continuation" *is* the runtime's async return.

### Canonical param ordering → the content address

`name(handle)` is throwaway, but `description(Action) + filled params(Params)` is `hash(Action + ordered Params)`
= the `EngramID/MoveID`. So the model's tool call **deterministically names the engram it invokes** — an elegant
tie to identity. The one requirement it imposes: the param schema must yield **canonically ordered** params, or
the hash is not stable. The adapter enforces the ordering.

## Multiple tools = the Focus (not a single-tool DSL)

Present the Focus as multiple tools — one per available operation — rather than collapsing everything behind one
tool and a DSL string. Focus *is* "the menu of choices presented to this move," and models are trained for
structured tool selection and arg-filling. A single-tool DSL buys no privacy (name/description/schema already
expose nothing host-side) while throwing away the substrate's one real strength. Collapse toward one tool only
if tool-list size/cost or per-turn re-registration actually bites, and prefer top-N-Focus + a fallback over a
full DSL even then.

Uniqueness note: "names unique within the active tool set" = handles unique within a Focus. With opaque per-turn
handles, uniqueness is free by construction.

## Transport vs. canonical — the layering

```text
TRANSPORT (ephemeral, boundary-only, 0 host semantics)
  name   = opaque per-turn handle        ← unique in the Focus, throwaway
  callId = Codex's call correlation       ← throwaway

CANONICAL (host-side, the runtime's ontology)
  description = Action (runtime-governed length)   inputSchema = Params
  opcode · EngramID = hash(Action + ordered Params) · InvocationID · Horizon/caps · tracing
```

The inference driver maps transport ↔ canonical; nothing transport-shaped crosses inward.

## Open

```text
· ANTHROPIC ROUTE. deferred until subscription-backed auth is available; a second inference driver later.
· ACTION LENGTH ENFORCEMENT. tokenization + the 160 default's enforcement point remain open (from authority-and-focus).
· FOCUS SIZE. large Focus × long Actions is real tool-list context cost; Focus-size management is a separate concern.
· ADAPTER SHAPE. the concrete normalize-in / normalize-out code and the Codex SDK wiring are unbuilt.
```

## Related

- [Drivers and the boundary](drivers-and-the-boundary.md) — the transducer and the egress invariant this specializes.
- [The main loop](main-loop.md) — where the `Call`/`Ret` moves this driver emits flow.
- [Authority and Focus](authority-and-focus.md) — Focus as `[]c.Capability`, Action as the sole model-visible name.
