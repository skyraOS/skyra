# Intent Graph Struct

## The Struct

```go
type IntentEntry struct {
    Primitive string // look, listen, engage, delimiter
    Target    string // reality ID, content, or empty
    Status    string // pending, success, failed
    Result    string // what came back — success message or error message
}

type IntentGraph struct {
    Keys    []string               // ordered action IDs — insertion order preserved
    Entries map[string]IntentEntry // keyed by action ID
}
```

## Keys

The key is a generated action ID — not primitive:target, because the same primitive:target combination could appear more than once in a graph (engage:michael twice in different phases). A short random ID keeps entries unique.

## Delimiters

A delimiter is an IntentEntry with `Primitive: "delimiter"` and no target. It sits in the Keys slice like any other entry. Motor iterates Keys, fires everything before the first delimiter concurrently, then stops. On the next dispatch, motor skips past the completed delimiter and fires the next phase.

## Result Field

Result is what the primitive reports back after execution. It's for the being to read on the next traversal — not for the user.

Examples:
- Look succeeds: `"5 entities visible"` or a summary of what populated surroundings
- Listen succeeds: `"subscribed to thread:skyra+michael"`
- Engage succeeds with entity target: `"michael responded"` or the content that came back
- Engage succeeds with content target: `"delivered to thread:skyra:main"`
- Engage fails — target not found: `"target not found: forge"`
- Malformed primitive: `"unmatched primitive: foo"`
- Engage fails — LLM error: `"provider error: timeout"`

### Should Result be richer than a string?

The question: does the being need the actual Reality that came back, or is a string summary enough?

Arguments for string:
- The being doesn't read Result directly in isolation. It reads PreviousActions, which are hot on the next traversal. The full context of what happened is already on the thread (engage writes the response to Slice.Thread keyed by reality ID). Result is a summary for quick orientation, not the full payload.
- Keeping it a string means the intent graph stays lightweight. It's a plan with annotations, not a data store.
- The being's Present already renders thread entries. The full response lives there. Result just says "it worked" or "it didn't and here's why."

Arguments for Reality:
- If the being needs to reference what came back inside the same motor dispatch (actions in the same phase that fired concurrently), a string loses structure. But this shouldn't happen — concurrent actions in the same phase don't depend on each other. If they did, they'd be in different phases separated by a delimiter.
- Future primitives might return structured data that a string can't capture. But that's a future problem — and the thread already captures the full result.

**Decision: string for now.** The full payload lives on the thread. Result is a one-line summary for the self-monitoring loop. If this proves insufficient during experiments, it can become `Result Reality` later without breaking anything.

## The Being Maintains Its Own State

The intent graph is shown to the being in its Present on every LLM call — propagate and compress. The being must re-emit the full intent graph each time. Whatever it emits IS the new graph. No system-level replace/append logic.

- If it wants to keep an action, it re-emits it.
- If it wants to drop one, it doesn't include it.
- If it wants to add new ones, it includes them.
- If it wants to reorder, it emits the new order.

The system parses whatever comes back and that's the new IntentGraph. Full agency over its own attractor.

Motor still removes completed entries from the intent graph after dispatch and moves them to PreviousActions. The being should see a clean intent graph — only what hasn't fired yet. No reason to confuse the agent with clutter from actions that already executed. PreviousActions is where outcomes live. Intent graph is where pending actions live. Clean separation.

## Parsing LLM Output Into Entries

## LLM Emission Format

```
<intent>
<engage>michael</engage>
<wait></wait>
<look>workshop</look>
</intent>
```

`<wait></wait>` is the delimiter. Everything before it fires concurrently. Motor stops. Results return, next traversal, everything after fires. The being is explicitly saying "I need results back before I continue."

Symmetrical XML — no unclosed tags to confuse the LLM.

Parsing:
1. Extract content between `<intent>` tags
2. Split on `<wait></wait>` — each segment is a phase
3. For each segment, parse primitive tags (`<look>target</look>`, `<listen>target</listen>`, `<engage>target</engage>`)
4. Build IntentGraph: entries in order, delimiter entries where `<wait></wait>` appeared

## Methods

```go
// IntentGraph helpers
func (ig *IntentGraph) Render() string               // for Present() display
func (ig *IntentGraph) IsEmpty() bool
func ParseIntent(raw string) IntentGraph             // parses LLM output into a fresh IntentGraph

// Motor — all methods on Skyra
func (s *Skyra) Complete(r *Relation)                              // gets the slice, calls Fire
func (s *Skyra) Fire(ig *IntentGraph, r *Relation)                 // iterates entries, calls Dispatch for each, stops at delimiter
func (s *Skyra) Dispatch(name string, r *Relation, target string) string  // resolves primitive name → function, executes, returns result
```

Motor is registered as an ExpressFn closure that calls Complete:

```go
s.ExpressFns["motor"] = func(r *Relation) *Relation {
    s.Complete(r)
    return r
}
```

The call chain:

```
Realize() calls s.ExpressFns["motor"](r)
  → s.Complete(r)
    → r.GetSlice(s.BeingID) — gets the private slice, intent graph is there
    → s.Fire(&sl.Intent, r)
      → iterates entries in order, calls s.Dispatch(name, r, target) for each
        → s.Dispatch looks up s.Primitives[name], calls it, returns result
      → hits delimiter, stops
      → for each completed entry: writes outcome to PreviousActions, removes from intent graph
```

Everything is already on the Relation. The being ID is on `s.BeingID`. The slice is `r.GetSlice(s.BeingID)`. The intent graph is on the slice. The primitives map is on `s.Primitives`. No extra arguments needed.

The intent graph parsing block currently in Realize (lines 244-262) moves inside Complete/Fire. Motor at line 265 stays where it is — just becomes the one-line closure above.

`Dispatch` is the resolver:
- Name found → execute the primitive function, return result string
- Name not found → return `"unmatched primitive: " + name`

## Bootstrap Registration

```go
skyra.ExpressFns["compress"] = reality.ExpressCompress(skyra, inference.CallDeepSeek)
skyra.ExpressFns["motor"] = reality.ExpressMotor(skyra)  // no provider — dispatch only

skyra.Primitives["look"] = reality.PrimitiveLook(skyra)
skyra.Primitives["listen"] = reality.PrimitiveListen(skyra)
skyra.Primitives["engage"] = reality.PrimitiveEngage(skyra, inference.CallDeepSeek)  // engage gets the provider
```

Motor loses its provider argument. Engage gains one. The LLM call moved from motor into the engage primitive. Speak is gone — dissolved into engage.

## Where It Lives

IntentGraph is a field on Slice. It persists across traversals. Every LLM call (propagate, compress) sees it in Present and emits the new version. The system parses the emission and overwrites the field. Motor reads it at dispatch time.

## Interaction With PreviousActions

Motor dispatches an entry → primitive executes → motor writes the outcome to PreviousActions as a Reality (Content = result string, Description = primitive:target) → motor removes the completed entry from the intent graph. The being sees a clean intent graph (only pending actions) and PreviousActions (what just happened) on the next traversal.

PreviousActions remains `[]Reality`. Each entry is a cell whose Content is the result string and whose Description is the primitive:target. The being reads it through normal activation — no special rendering needed.
