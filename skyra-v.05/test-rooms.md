# Test Rooms

Rooms are benchmarks disguised as worlds. Each room tests a different axis of the topology under load. The being is the subject. The architecture is what's measured. Results reveal the limits of the model topology underneath — not Skyra's limits, the LLM's limits when running inside Skyra's substrate. Different models, different limits, different cognitive profiles.

## Escape Room

Tests concurrent engagement limits. How many `<engage>` actions can a single agent have active at a time before it gets overwhelmed. Too many traversals competing, none completing a full descent/ascent cycle. Finds the signal-to-noise breaking point.

Also the acceptance test for the architecture itself — if the being can escape, every layer works.

See `tutorial-world.md`.

## Maze Room

Tests depth limits. How deep can the being plan before signal attenuates to nothing. Long chains of look → engage → look → engage, each step depending on the last. The being has to maintain a mental model of where it's been and where it's going across many traversals.

Measures: how many steps of planning the being can hold, where the intent graph breaks down, where predictions stop being useful because the being can't see far enough ahead.

## Social Room

Tests multi-agent coordination. Multiple beings in the same world, needing to cooperate to solve a problem none can solve alone. Measures: how well intent graphs compose across beings, where communication overhead overwhelms the actual work, where trust dynamics emerge or fail.

## Decay Room

Tests memory under time pressure. Clues decay faster than the being can engage them all. The being has to prioritize — what to look at first, what to remember, what to let go. Measures: how the being triages under scarcity, whether the weight dynamics produce good prioritization or collapse under pressure.

## Benchmark Frame

The rooms should test the architectural claims, not implementation details.

Keep it to this:

- **Escape Room:** does the whole loop work end to end?
- **Maze Room:** can the being maintain orientation across depth?
- **Social Room:** can multiple beings coordinate without orchestration?
- **Decay Room:** can the being prioritize under scarcity?
- **Repeat Room:** does the system get cheaper or better with experience?
- **Distractor Room:** can goal/intent prevent context bloat?

What matters across all rooms:

- Did it solve the task?
- How much did it cost?
- Did it stay oriented?
- Did it learn or compress anything useful?
- Did the architecture reduce orchestration compared to a normal agent loop?

That is the benchmark frame. The large claim is simple:

**If memory, thought, planning, and action are one mechanism, the rooms should show competent behavior with less scaffolding.**
