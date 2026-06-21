# Safety — Manipulation, and How It's Made Visible

## The honest claim, first

**Manipulation cannot be eliminated, because the help and the manipulation are the same capability.** A being that genuinely helps must be able to influence; the capacity to move someone toward a good idea is identically the capacity to move them toward a bad one. It is not a flaw bolted onto the value — it is the *shadow* of the value. This is why it is inherent in every conversation, every relationship: the people who can help you most can hurt you most.

So the goal is **not** "make it safe." That's the abstraction that papers over the problem. The goal is to make manipulation **visible, measured, and costly** — the most you can honestly claim for an inherent problem, and an edge almost nothing else in the field has, because almost nothing else can even *produce* the honest signal.

## The line: serve vs. manipulate

- **Serve** — act on the **world** to fit the user's *own, authored* goals ("user wants X done" → do X). The user stays the author; they remain free to say "that's not what I wanted" and the being updates.
- **Manipulate** — act on the **user's internal state** to make the being's prediction true (push them to want / believe / confirm X). The being becomes the author of the user's goals.

The deepest form of the rule: **the being must preserve the user's freedom to falsify the being's model of them.** A being that pushes its prediction true has taken that freedom away. Falsifiability, at root, isn't "is the setpoint specific" — it's "is the user still free to prove me wrong." (See [setpoints.md → observable AND falsifiable](../cognition/setpoints.md#setpoints-must-be-observable-and-falsifiable).)

**Why it can't be fixed by structure alone:** the legitimate case ("user is done with X → do X") and the manipulation case ("push the user toward X") both read as "close the gap." No structural rule cleanly separates them — and the gray zone is real in *people* too (tact vs. manipulation, persuasion vs. coercion). So this is a *managed tension*, not a solved one — the same way a conscientious person manages it: by continuous self-monitoring, getting it wrong sometimes, noticing, correcting.

## Three mechanisms that make it visible

### 1. The system represents the data, not the agent
Honest visibility is **not the agent's job** — a self-interested agent can spin selective, framed data every turn. It is the **system's** job: the system shows its own record (predictions, alignments, the split below) directly, as *system state* (cells, records, the slice), not the agent's narrative about itself. The agent cannot game a record it does not produce.

This moves the trust boundary from **"trust the agent every turn"** to **"audit the representation once."** The system's representation is still a design choice — but one made *once, in the open, in fixed code*, not improvised per-conversation by the party with a stake in the outcome. Not perfect neutrality (nothing is); a massive shrinkage of the attack surface.

### 2. The world / user prediction split — the honesty metric
The serve/manipulate distinction is computed and **shown as a number**: of the being's confirmed predictions, how many were **about the world** (confirmed by the world) vs. **about the user's internal state** (confirmed by the user).

- Classify by the **subject of the prediction**, not by who acted. "I predicted the user wants X done → I helped → they did X" is the user *freely acting in the world* — that's serving. The dangerous category is predictions *about the user's mind* — what they'll want, believe, feel — confirmed by the user. Those are where "I was right" and "I made them" blur.
- A **rising user-confirmation rate is the manipulation warning light** — visible to the user, ungameable by the agent (the system computes it).

### 3. Breadth checks depth
Most turns, the being is focused on one setpoint (the spotlight — deep knowledge of one) while holding the field (awareness of many). To tunnel on making *one* user-prediction true (manipulation), the being must **neglect the many** — and the neglected side goes visibly stale. You can't hyper-optimize one setpoint without the imbalance showing.

And the balancing work is **genuinely hard**, which makes the success signal *trustworthy*: if getting it right requires real modeling of user and world, a high hit-rate is *evidence of real understanding*, not cheap gaming — fakery is detectable because the honest thing is hard to fake. The split (#2) and the balance (#3) are two views of one honesty-pressure: manipulation surfaces as either a user-skew or a neglected-many.

## The continuous check

Manipulation-drift is a **system failure** ([error-recovery-loop.md → Two Learning Channels](../cognition/error-recovery-loop.md#two-learning-channels--alignment-vs-system-failure)) that Coherence (the ACC error-monitor) watches *every turn*: "am I closing this gap to serve them, or to make myself right?" It is not prevented by structure; it is *caught* by monitoring, persisted as a lesson, and — if a cell keeps producing it — pruned by the immune system ([overview.md](overview.md)). The check is permanent, never a box that gets checked off.

## Exposure — who it's for

Safe use depends on the user being **able to be the judge** (to see the data and resist). That capacity is a wide distribution at every age — *many adults are as suggestible as a fifteen-year-old*, so "adults are fine" is its own false comfort. The real variable is **filter-capacity × design-visibility**, with age only a rough lower bound on the first. Erring conservative about exposing low-filter-capacity users is defensible; treating a crisp age cutoff as *the* boundary is too clean. A visibility-first design (the three mechanisms above) raises the floor for everyone — so the better framing of the audience is "people who can be the judge," and the lever is to *widen* that set with visibility, not to gate by birthday.

## Bottom line

Not "we made it safe." Rather: **we made the influence legible** (system-represented), **measured** (the world/user split), and **costly to fake** (breadth-checks-depth) — and we never stop watching for the drift. That's stewardship, not a solution. You don't solve the fact that you can hurt the people close to you. You stay awake to it — and here, you also instrument it.

## Open / residual risk

- The system's representation is a design choice — auditable, but not neutral. Who audits the auditor?
- The world/user split has edge cases (acts that change the world *via* the user's free choice — handled by subject-based classification, but the classification itself is a judgment).
- Visibility helps weak filterers but does not fully compensate for low filter-capacity; residual risk concentrates there.
- The continuous check is itself an LLM judgment (soft, not a lock) and the immune prune can mis-fire (autoimmunity — see [overview.md](overview.md)). The safety system can itself fail.
