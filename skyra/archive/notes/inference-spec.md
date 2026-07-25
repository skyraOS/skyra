# Inference — Energy Reality

A Reality that represents the energy a being can exert. Every thought costs something. Inference tracks, gates, and pressures.

---

## What it is

Inference wraps the LLM call. It sits between the being and the provider. Every time Think or Act fires the LLM, the call passes through Inference. Inference tracks the cost and decides whether the being has the energy to make the call.

Current flow:
```
Think → Provider.Realize (LLM call)
Act   → Provider.Realize (LLM call)
```

New flow:
```
Think → Inference.Realize → Provider.Realize (LLM call)
Act   → Inference.Realize → Provider.Realize (LLM call)
```

Inference lives inside Self, alongside Think and Act. Each being has its own Inference with its own energy pool. Skyra's energy is separate from Louise's.

---

## Energy pool

Each being starts with an energy budget. Every LLM call spends energy. When the pool is empty, the being can't think.

```
inference {
  id          string
  pool        int         // remaining energy (tokens)
  capacity    int         // max pool size
  spent       int         // total tokens spent lifetime
  calls       int         // total inference calls lifetime
}
```

### What costs energy

OpenRouter returns token usage on every call:
```json
{
  "usage": {
    "prompt_tokens": 100,
    "completion_tokens": 50,
    "total_tokens": 150
  }
}
```

We're not capturing this right now. `inference.Call` returns only the response string. It needs to return token counts too. The cost of a call is `total_tokens` — that's what drains from the pool.

### Pool size

Set per-being in the genome:
```
grow ~name skyra ~type llm ~device openrouter ~energy 100000 ...
```

Default: 100,000 tokens. Enough for a long conversation. The number is tunable — different beings can have different energy levels. A being built for quick responses gets a smaller pool than one built for deep research.

---

## Pressure

When the pool gets low, the being should feel it. Same pattern as think-time pressure — Inference attaches a parser to the relation that tells the being how much energy it has left.

```
energy: 85000/100000 tokens remaining
```

At 20% remaining:
```
energy: low. 18000/100000 tokens remaining. conserve your responses.
```

At 5% remaining:
```
energy: critical. 4200/100000 tokens remaining. be brief.
```

At 0%:
```
(call blocked — Inference returns empty, being goes idle)
```

The being doesn't manage its energy. It feels the pressure and adapts naturally. The LLM will shorten its responses when told to conserve. This is the same design as think-time — the system applies constraints, the being responds to them.

---

## Recharge

Open question. Options:

1. **Time-based** — pool refills at a rate (e.g., 1000 tokens/minute). Beings recover energy when idle.
2. **User-funded** — the user explicitly gives energy to a being. "skyra, here's 50k more tokens."
3. **Task-based** — completing tasks for other beings earns energy. This is the coupling point with Economics that we're keeping separate for now but could bridge later.
4. **No recharge** — the pool is what it is. When it's spent, the being is done until the user resets it.

For alpha: option 4. Fixed pool, no recharge. Simple. The pool is big enough for a demo. Recharge mechanics are a post-alpha concern.

---

## Interface

```go
type Inference struct {
    id       string
    Pool     int
    Capacity int
    Spent    int
    Calls    int
    Provider Reality   // the actual LLM provider
}
```

Implements Reality:
- `ID()` → "inference"
- `Create(r)` → new Inference with capacity from genome
- `Realize(r)` → check pool, attach pressure parser, fire Provider, deduct cost, return response

### Realize flow

```
1. Check pool > 0. If not, return "" (being is spent).
2. Attach energy pressure parser to relation.
3. Fire Provider.Realize(r) — the actual LLM call.
4. Read token usage from response (needs inference.Call to return this).
5. Deduct from pool. Increment spent and calls.
6. Return the LLM response.
```

### Collecting

When a collecting relation passes through:
```go
if r.Collecting {
    r.Export("inference:"+beingName, InferenceSnapshot{
        Pool:     i.Pool,
        Capacity: i.Capacity,
        Spent:    i.Spent,
        Calls:    i.Calls,
    })
    return ""
}
```

The universe sees every being's energy state.

---

## Changes needed

1. **inference.go (package)** — `Call` needs to return token counts alongside the response. Change signature to `Call(system, present string) (string, int, error)` where the int is total tokens used.
2. **llm.go** — Provider.Call signature changes. Provider.Realize captures token count and passes it up.
3. **New file: src/reality/inference.go** — the Inference Reality.
4. **self.go** — Self wires Inference between Think/Act and the Provider.
5. **genome.skyra** — add `~energy` field to grow lines.
6. **universe.go** — add InferenceSnapshot type. Self's collecting branch fires Inference to get its export.

---

## What this is not

- Not a billing system. No dollar amounts. Energy is tokens.
- Not shared between beings. Each being has its own pool.
- Not coupled to Economics. Inference is physics. Economics is trade.
- Not a rate limiter. It doesn't throttle — it pressures, then blocks at zero.
