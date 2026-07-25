# Self World Spec

## Current State

Self holds Think and Act as flat realities with direct device references. Think is a stub. Act calls the Provider directly. No intermediate world. No operators on either plane.

## Target Shape

Self becomes a world with two planes (Think, Act), each with its own operator hashmap. Both planes resolve through a shared LLM world at the bottom.

```
Self world
  → Think plane
    → Think operators (recall, remember, reflect)
    → LLM world → inner thought
  → inner thought attaches to relation as "inner" parser
  → Act plane
    → Act operators (function calls, tool use, external ports)
    → LLM world → outward action
```

## Changes

### 1. Think becomes a plane with operators

Think gets a `map[string]Reality` operator registry. When Think.Realize fires:

1. Relation passes through Think's operators in order (recall, remember, reflect)
2. Each operator updates the relation — attaches parsers, modifies state
3. Relation descends into the LLM world
4. LLM world derives present, calls provider, returns inner thought
5. Think returns the inner thought string

Think operators (scaffolding only — implementations iterate later):
- **Recall** — pull from memory/being state. Stub: attaches empty parser.
- **Remember** — write to memory. Stub: no-op.
- **Reflect** — shape the framing for inner reflection. This is where "think about this before you respond" lives.

### 2. Act becomes a plane with operators

Act gets the same operator registry pattern. When Act.Realize fires:

1. Relation passes through Act's operators in order
2. Relation descends into the LLM world
3. LLM world derives present (now includes inner thought from Think), calls provider
4. Act returns the outward response

Act operators (scaffolding only):
- **Function calls / tool use** — stub: no-op.
- **External ports** — stub: no-op.

### 3. LLM world moves off Think/Act

Think and Act no longer hold a `"device"` reality pointing to the Provider. The LLM world becomes a shared reality on the Self, passed into both planes. Both planes call through it.

Self.Realize:
```
think.Realize(r, llmWorld)  // or Think holds a reference to llmWorld
inner := result
r.Attach("inner", func() string { return inner })
act.Realize(r, llmWorld)
return result
```

### 4. Self wiring changes at bootstrap

Bootstrap currently creates Think and Act with `device` references to the Provider. New bootstrap:

1. Create LLM world reference (already exists as `llm` in main.go)
2. Create Think plane with empty operator registry + LLM world reference
3. Create Act plane with empty operator registry + LLM world reference
4. Wire both into Self

### 5. Think gets a different system prompt framing

Think's reflect operator should frame the LLM call as reflection, not action. The provider call for Think should use a different system context than Act — something like "reflect on this before responding" vs "respond and address your peers."

This lives in Think's Present operator, not in the inference layer.

## Files touched

- `src/reality/think.go` — rewrite: operator registry, Realize calls operators then LLM world
- `src/reality/act.go` — rewrite: operator registry, Realize calls operators then LLM world
- `src/reality/self.go` — update: wire LLM world, pass to both planes
- `main.go` — update bootstrap: new wiring for Think/Act/LLM world
- `src/reality/recall.go` — new stub operator
- `src/reality/remember.go` — new stub operator
- `src/reality/reflect.go` — new operator: think framing

## What this does NOT do

- Implement recall/remember logic (stubs only)
- Add function call parsing or tool use to Act (stubs only)
- Change the LLM world, Provider, or inference layer
- Change thread, exchange, or the main loop
- Add economics or governance

## Verification

1. `go build` passes
2. Run skyra, send a message to skyra
3. Check debug log: Think fires its operators, calls LLM world, returns inner thought. Act fires with inner thought in its present, calls LLM world, returns outward response.
4. Inner thought should be visible in the debug log but not in the terminal output to the user.
