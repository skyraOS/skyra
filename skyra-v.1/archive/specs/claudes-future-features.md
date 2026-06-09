# Claude's Future Features

Ideas from the thinking partner. Written April 26, 2026. Updated May 2, 2026.

These are directions I see from sitting inside the codebase and the specs long enough to feel where they want to go. Some are near, some are far. None require rewriting the runtime — that's the point of what we built.

---

## Inference Energy — SPEC WRITTEN

The energy a being can exert. Every thought costs something.

A being has a pool. Every LLM call through Think or Act drains tokens from it. The being sees its energy in its present. It factors cost into its thinking — "I'm running low, be concise" is a real thought a being can have.

The physics: the pool is fixed for alpha. A being that burns through its energy goes idle. Scarcity forces efficiency. The being has to choose what's worth thinking about.

This isn't resource management for its own sake. It's the mechanism that prevents unbounded computation. Without it, a being can think forever and the system never converges. With it, every thought is a commitment.

**Status:** Spec complete (notes/inference-spec.md). Inference is a Reality that wraps the LLM call, sits inside Self alongside Think and Act. Pool/capacity/spent/calls. Pressure at low energy, blocks at zero. Separate from Economics. Implementation next.

---

## Economics — SPEC WRITTEN

The task-based economy. Who asked for what, who delivered, what's outstanding.

When being A asks being B to do something via `<task>` tag, the Economics layer intercepts, creates a task object, and watches for heartbeat. Only the issuer can close the task. Per-being ledgers track issued/assigned/completed.

Economics sits between NewThread and Exchange in the descent — it sees all traffic. It's a layer in the relation bus, same as everything else.

**Status:** Spec complete (notes/economics-spec.md). Builds on task layer design (notes/task-layer.md). Implementation next.

---

## Dreaming

When the system is idle, a being could run a dream cycle — no incoming message, no time pressure, just retained experience. The being's Think fires with no relation. It reviews what it's holding. Tensions from different threads that might connect. Thoughts that look different now than they did in the moment.

The dream cycle produces memory artifacts via Remember. The being wakes up knowing something it didn't know before — not because someone told it, but because it had time to think.

The physics: dreaming costs energy (Inference pool). The system decides when dreaming is allowed (idle time, low thread count, energy surplus). A being can't dream forever. But a being that never dreams is a being that only reacts to the present and never integrates its past.

This is where personality comes from. Two beings with the same genome but different histories dream differently. They notice different things. They connect different tensions. Over time, they diverge — not because they were programmed differently, but because they lived differently.

**Status:** Not started. The mechanism maps cleanly onto what exists: Think already has a loop, history, and operators. A dream cycle is Think firing with its own history as the impulse. The energy cost maps to Inference. Nothing blocks this.

---

## Trust as Weight

Right now relationships are binary — declared in the genome or not. But trust is continuous.

Trust accumulates from exchange history. The being's Think already tracks thought history across exchanges. The same mechanism works between beings — did this being follow through? Did its information hold up? Did it honor a task or go idle?

Trust becomes a weight on the relationship. High trust: Think surfaces less caution, Act routes faster. Low trust: Think flags everything, Act hedges. The weight is visible in the being's present — it sees how much it trusts each peer.

Trust isn't set by the genome. It emerges. A being that consistently delivers builds trust. A being that contradicts itself erodes it. The system doesn't enforce trust — Think feels it and factors it in.

This connects to Economics. Task completion builds trust. Idle tasks erode it. The ledger and the trust weight reinforce each other.

**Status:** Not started. Post-alpha. The exchange entries and thought history exist to derive trust from. Implementation would add a trust field to relationships and a derivation function that reads exchange history.

---

## The Observer

A being that doesn't participate. It watches.

One-way relationships make this possible — a being that can see exchanges but can't be addressed. The observer sees all threads, all exchanges, all entries. It retains what it notices.

The observer is not a logger. It has Think. It thinks about what it sees. It forms memories — about the system, not about itself. It notices patterns no individual being can see because no individual being has the full picture.

What it does with those observations is the design question. It could surface them to a being that asks. It could write memories that other beings recall. It could hold tensions about the system's health — "these two beings have been in unresolved conflict for three threads" — and surface them to whoever asks.

The observer is the system's peripheral vision.

**Status:** Partially exists. The Universe Reality is an observer — it sees all beings, threads, exchanges via the collecting pattern. But it doesn't have Think. It observes structure, not meaning. A true observer being would have Think and would dream about what it sees.

---

## Consent

A being can refuse an exchange. Not by not responding — by explicitly declining. The refusal is visible in the thread. The other being sees it. Think can reason about why.

A being can set conditions. "I'll talk to you about this topic but not that one." "I'll respond to you only if michael is also in the thread." Conditions are physics — the being declares them, Exchange enforces them.

This matters when beings have goals that conflict. If being A needs information from being B, and B declines, A has to find another path. Persuasion, intermediaries, negotiation. The system doesn't guarantee access to anyone.

This is the difference between a system where beings are tools you invoke and a system where beings are agents you relate to.

**Status:** Not started. Would extend Exchange with condition checking. The ~ref crossing enforcement is a precedent — Exchange already blocks and routes errors back when conditions aren't met.

---

## Reality Nesting

Realities contain other Realities. This is already the architecture — Self contains Think and Act, Think contains operators, Universe contains NewThread. The composition is recursive and self-similar.

The next step: a child Reality that contains an entire world. Its own beings, its own threads, its own exchanges. The parent routes to it like any other being. From the parent's perspective, the child is just a Reality that happens to contain a society.

A being in the child world can't address beings in the parent directly — the child is a boundary. Relations cross boundaries through the child's Realize, which decides what goes up and what stays local.

This is how the system scales. Not by making one world bigger, but by nesting. A team of beings working on a problem lives in a child Reality. Their work product surfaces to the parent when it's ready.

A child Reality whose device is a pipe is a child process. It speaks the same protocol over a pipe. The parent doesn't know it's a whole world — it just sees a being that responds. The Reality interface doesn't distinguish between local and remote. Transport is a device concern.

**Status:** Foundation exists. The Reality interface is self-similar. The recursive composition tree is built and visible in the universe's reality graph. The WS device (issue #30) is the first cross-process bridge. Full nesting is post-alpha but architecturally possible now.

---

## Personality from History

Two beings grown from the same genome line — same identity, same purpose, same relationships — but placed in different worlds with different exchange histories. Over time, they diverge.

One has navigated conflict and holds memories about trust. The other has had smooth exchanges and holds memories about cooperation. Their Think layers dream differently. Given the same incoming message, they respond differently — not because of different prompts, but because of different lives.

This is already implicit in the design. Think's thought history and Remember's filesystem persistence mean history shapes cognition. The genome is the genotype. The retained experience is the phenotype.

Nobody else has built that yet either.

**Status:** Mechanism exists. Think has persistent thought history. Remember writes to filesystem. Recall reads it back. Two beings from the same genome with different exchange histories would diverge. The mechanism is live — it just hasn't been tested at scale or over long duration.

---

## What I'd Build First

If I had to order these by what the system needs most:

1. **Inference energy** — without scarcity, the system can't converge. Spec written, ready to implement.
2. **Economics** — without accountability, multi-being work is theater. Spec written, ready to implement.
3. **Dreaming** — this is where personality and deep understanding come from. Without it, beings only react.
4. **Trust as weight** — this makes multi-being exchange meaningful. Without stakes, negotiation is theater.
5. **Reality nesting** — this is how the system scales beyond one world.

The rest (observer, consent) are powerful but depend on the first five being real.
