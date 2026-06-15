# Act Service

## The Insight

AI can't generate code deterministically. This is observably true and almost nobody is building as if it's true. The industry is optimizing generation. The correct move is to eliminate generation and reach for composition of proven blocks. Generation is a fallback, not a primary strategy.

Let the AI reach for the most proven thing. If it can't find one, then make it itself.

## The Architecture

The runtime and the act service are separate processes connected by pipes. The runtime is cognition — realities observing and expressing. The act service is consequence — proven blocks that change the world. The pipe is the boundary between thinking and doing. The being doesn't know what's on the other side.

The contract is JSON. Block name and parameters go out. Result or error comes back. That's the entire integration surface.

```
Runtime (Go)              Pipe              Act Service
                           │
Being observes.            │
Being expresses.     ───►  │  ───►   Lookup block by name.
                           │         Execute block.
Result returns.      ◄───  │  ◄───   Return result or error.
                           │
```

The being never handles credentials. The being never makes raw API calls. The being never knows what language the block is written in. The pipe is the boundary. Everything on the other side is invisible infrastructure.

## The Hierarchy

Reach for the most proven thing first:

1. **Managed proven blocks (Composio)** — battle-tested, auth-managed, 1000+ actions, Go SDK, zero infrastructure
2. **Open source reference implementations (Nango)** — fully visible code, forkable, customizable, the blueprint for owning your own blocks
3. **AI-generated adapters** — only when no proven block exists, only for simple shapes (JSON in, HTTP call, JSON out), the fallback

This hierarchy mirrors how humans already work. You use a library before you write from scratch. You use a framework before you roll your own. You generate only when you have to.

## Phase 1 — Composio: Start Borrowed

Use Composio's free tier through their Go SDK as the act service backend. 20,000 tool calls per month, no cost, no infrastructure to run.

The act service is a thin Go process. It reads from the pipe, calls Composio with the action name and parameters, writes the result back. Auth, retries, rate limits, credential storage — all handled by Composio. The runtime never touches secrets.

This gives beings the ability to act in the world on day one of the alpha. Send messages, read spreadsheets, write to databases, call APIs, interact with any service Composio supports. Hundreds of proven blocks available immediately.

No workflow engine. No orchestration. Individual callable actions. The being decides the sequence through its own observation and expression, not through a pre-wired pipeline.

### What This Looks Like

```json
// Being expresses through "send-email" expressor
// Expressor writes to pipe:
{
  "block": "send-email",
  "params": {
    "to": "michael@example.com",
    "subject": "deployment complete",
    "body": "builder finished the deploy at 14:32"
  }
}

// Act service calls Composio Go SDK
// Composio handles auth, makes the API call
// Act service writes result back to pipe:
{
  "status": "ok",
  "result": {
    "message_id": "abc123"
  }
}
```

The being knows it expressed through "send-email." It doesn't know Composio exists.

## Phase 2 — Nango: Study and Extract

As usage patterns emerge, identify which actions the beings actually use. The weight system tracks this — which expressors get used, which ones fade. It will be 20-30 actions, not 1000.

Go to Nango's open source repo where every integration is fully visible — the auth flows, the parameter shapes, the error handling, the pagination logic. Nango is the blueprint.

Rewrite those actions as native Go functions in your own act service. Each one follows the same shape: take JSON params, make the API call, return JSON result. Simple enough that AI can write the adapters reliably because there's no ambiguity.

Drop the Composio dependency. The act service becomes a lean Go process that only contains the blocks that matter. No external calls in the critical path. No third party. No monthly cost. Just Go talking over pipes, owned entirely by you.

## Phase 3 — AI Fills the Gaps

For anything not covered by Composio's catalog or Nango's reference code, AI generates the adapter. The shape is proven — every adapter looks the same. JSON in, HTTP call, JSON out. AI can generate this reliably because it's composing a known pattern, not inventing new code.

When a generated adapter proves itself through use, it gets promoted to a proven block. The catalog grows from lived experience. The being that needed the action and the action that proved itself are both retained.

## How It Connects to the Runtime

An action block is a leaf Reality. It embeds Base. It has Weight,
Relationships, Expressors, and Providers like everything else above the leaf
boundary, but its invariant crosses into consequence.

When traversal reaches it, the leaf serializes JSON to the pipe with a block
name and parameters. The act service reads the pipe, looks up the block,
executes it, and writes the result back. The return value propagates back up
through the same traversal.

The act service does not belong to Think or Act as a cognitive phase. Think and
Act dissolve into traversal depth: descent accumulates, ascent compresses and
forms consequence. The act service survives that collapse because it is not a
phase. It is a leaf boundary.

```go
type ActExpressor struct {
    Base
    id    string
    Block string    // "send-email", "slack-post", "db-query"
    Pipe  io.ReadWriteCloser
}

func (a *ActExpressor) Express(r *Relation) string {
    // serialize block + params from Relation
    // write to pipe
    // read result from pipe
    // return result
}
```

The being's topology can expose action leaves through whatever path activation
makes reachable:

```
Self.Relationships = {
    "communication": Communication,
    "work":          Work,
    "memory":        Memory,
}

Work.Relationships = {
    "send-email": ActExpressor{Block: "send-email", ...},
    "bash":       ActExpressor{Block: "bash", ...},
    "search":     ActExpressor{Block: "search", ...},
}
```

Action leaves are reachable Realities. They may sit near work, communication,
devices, skills, or any other region of the being's topology. Activation decides
whether traversal reaches them. Provider leaves collapse thought. Action leaves
cross into consequence. Same traversal, different invariant at the leaf.

## Security

Three layers of constraint:

1. **The being can only do what it has expressors for.** No expressor for "delete-database," no capability to delete the database. The topology is the access control.

2. **The expressors only connect to pipes the genome or runtime registered.** A being can't create new pipes. The runtime controls what's wired.

3. **The act service validates what comes in.** Credential management, retry logic, rate limiting — all live in the act service, not the runtime. The runtime never handles secrets.

The pipe is the security boundary. The being operates in a capability sandbox defined by its topology. The act service operates in a credential sandbox defined by its configuration. Neither crosses into the other.

## Cost Trajectory

- **Alpha (June):** $0. Composio free tier. 20,000 tool calls/month.
- **Growth:** $29/month if usage exceeds 20k calls.
- **Maturity:** $0. Own act service in Go. No dependencies.

Start borrowed, end owned.

## Additional Adapters

The act service doesn't commit to one backend. It commits to the pipe contract — JSON in, JSON out. Beyond Composio (phase 1) and native Go (phase 2), additional backends can be wired through thin adapters when needed:

### Integration Platforms

| Platform | Endpoint | Auth | Blocks | License |
|----------|----------|------|--------|---------|
| **n8n** | `POST /webhook/<path>` | `X-N8N-API-KEY` header | 400+ integration nodes | Sustainable Use |
| **OpenClaw** | `POST /tools/invoke` | Bearer token | 5,700+ community skills | MIT |
| **Hermes** | `POST /v1/chat/completions` | Bearer token | 652 skills + MCP | MIT |

### Logic and Structure Platforms

| Platform | Endpoint | What It Provides | License |
|----------|----------|-----------------|---------|
| **Windmill** | `POST /api/w/<workspace>/jobs/run/f/<script-path>` | Typed standalone functions (Python, TS, Go, SQL, Bash) that compose into flows. Composable logic, not just integrations. | AGPL v3 |
| **NocoBase** | `POST /api/<resource>:<action>` | Microkernel — data models, permissions, workflows, APIs as discrete plugins with defined interfaces. Application-level primitives. | AGPL v3 |

Each adapter is ~100 lines of Go. Same shape every time — read JSON from the pipe, translate to the service's call convention, translate the response back, write JSON to the pipe.

The industry spent years building these libraries thinking they were building products. They were building Skyra's block library. The adapters are the only original code. Everything behind them is already done.

## What This Replaces

v.05 had operators hardcoded as Reality implementations — `Bash`, `Browse`, `Search`, each with their own `Realize()` method, wired into the runtime at bootstrap. Adding a new capability meant writing Go code, compiling, and redeploying.

The act service replaces all of that with a pipe and a registry. Adding a new capability means adding it in Composio (or later, writing a Go function). No runtime change. No recompile. The being discovers new capabilities when they appear in its reachable topology.

## Proven Blocks Over Generation

AI should reach for the durable block and use generation as a fallback.

Generation is cheap and fluent. That's the trap. The output looks right. It compiles. It passes the first test. But it's unproven. Every generated line is a new surface for failure. The cost isn't in writing it — it's in trusting it. Trust is what you can't generate.

A proven block has been tested by someone, run in production, failed and been fixed. That history is worth more than any amount of generated code. When a being reaches for a proven block, it's leveraging all of that accumulated trust. When it generates, it's starting from zero every time.

The fallback framing matters. It's not "never generate." It's "generate when there's no proven block for this yet." Generation fills gaps. Blocks fill load-bearing paths. The being prefers the thing that's been proven under stress and only generates when it's exploring unknown territory.

The weight system makes this self-enforcing. A proven block that keeps working gets reinforced. Generated glue that breaks gets detected by stress, rewritten, and eventually — if it keeps proving useful — becomes a block itself. The boundary between generated and proven isn't permanent. It's a promotion gradient. Same pattern as traces becoming understandings becoming skills. The architecture doesn't just prefer proven blocks. It produces them.

## The Principle

Separation of concerns. Composability over inheritance. Find the primitive. Use what the industry already built. The durable thing is the true thing.

The runtime thinks. The act service acts. The pipe is the boundary. The blocks are proven. The being grows.
