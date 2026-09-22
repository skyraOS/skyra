# Creator and compiler — entity dissolves into two slots

**Date:** 2026-08-15
**Status: CONVERSATION RECORD. UNRATIFIED.** The session that came out of H16. Written down before
it is lost. Nothing here is checked against the code beyond the citations given.

> **`Entity` stops being one thing. An engram has a creator slot and a compiler slot, both fixed at
> solidification. Recognize runs on the return path only. Compile failure belongs to the child, not
> its parent, because the parent cannot see a body.**

---

## 0 · Where this came from

H16 asked whether an interpreter stack (`llm-python-machine`) is representable. It is not — and the
answer is that it was never a stack of entities. It is a creator, a compiler, and a host, and only
two of those are part of an engram.

```text
llm       CREATOR      authors the body.        absent at runtime.
python    COMPILER     interprets that body.    present every invocation.
machine   HOST         where that lands.        separate from the engram entirely.
```

Which resolves the fork H16 posed — three layers or three times — as **both, on different paths**:

```text
RUNTIME PATH   compiler + host. the creator is absent. two levels.
FAILURE PATH   the creator pops. reached only when compilation fails.
```

That matches the measurement already in `../../06-entity.md`: *"only [create] reaches a model.
Measured: 23 + 6 = 29 invocations with zero inference."*

---

## 1 · Three judgements, two owners

```text
CREATION        the body is authored.                  owned by the PARENT (it owns the position)
INTERPRETATION  the body is compiled and run.          owned by the CHILD
RETURN          the result is judged against contract. owned by the PARENT
```

The child owns the middle one, and therefore owns **failure to compile**. Its own mistake: it did
not think far enough ahead to handle the input arriving.

### The strongest argument is the version rule

`../../reach/age.md` already decides that a parent cannot see a body:

> *"The continuation is interior. A body rewrite does not increment."*

The four things a parent owns per position are signature, return type, entity, horizon —
*"the continuation is the part no caller can see."*

```text
PARENT OWNS   signature · return type · horizon        → judges the RETURN
CHILD OWNS    the body                                  → owns its FAILURE TO COMPILE
```

So the parent has **no standing to repair a body**. It cannot see it, cannot version it, and a
change to it is not an event at the parent's level. Assigning compile failure to the parent would
require it to fix something a decided rule says it cannot observe.

**This fills WHO COMPILES.** `../../coherence/README.md` has the slot — *"INVALID SOURCE PROGRAM,
owner: continuation compiler"* — and
`../../continuation-language/surfaces-and-solidification.md` records that no primitive corresponds
to it. The owner is the engram itself; the creator that authored it pops.

---

## 2 · Entity dissolves into two slots

**Slots on an engram, not kinds of thing.** This distinction is load-bearing:

```text
NOT   Creator and Compiler are two types of entity.        ← a capability declaration
BUT   an engram has a creator slot and a compiler slot.    ← a record of what happened
```

The second preserves the refusal `Entity` was built on:

```text
../../semantics/terminology.md   "an identity and nothing else — every declaration dissolved."
../../06-entity.md               "a declaration is made by whoever REGISTERED the entity, not by
                                 the entity."
observable-causality.md §3       "it does not ask the entity what it believes it can do."
```

One identity may fill both slots — a human creates and interprets, an LLM creates and interprets.
Python only ever appears in compiler slots, and nobody declared that. It is observed, because python
has never been recorded in a creator slot.

Same posture as leaf/orchestrator: positions derived from composition, never declared kinds.

---

## 3 · Both slots are fixed at solidification

The compiler is determined by the body's syntax. The parent does not choose it.

This contradicts `age.md`, which lists entity among the four the parent owns per position and makes
a change to it a version increment. The reconciliation is temporal:

```text
BEFORE SOLIDIFICATION   the behavior is prose. many compilers possible.
AT SOLIDIFICATION       the body acquires a syntax. THE COMPILER IS DETERMINED BY IT.
AFTER                   fixed. the parent no longer chooses.
```

Which also explains why partitioning keys on entity at all. `2026-08-09/interfaces.md`:

> *"the same expectation under Claude and under python3.11 returns different shapes."*

True of a **prose expectation**. Not true of a compiled continuation — python code has one compiler.
So entity-partitioning is the pre-solidification mechanism, and it stops mattering once the syntax
pins the reader.

```text
WHAT THIS FIXES   an incoherent pairing — a python body bound to an LLM interpreter — becomes
                  INEXPRESSIBLE rather than merely unwise. nothing enforces it today.
```

---

## 4 · Recognize runs on the return path only. Kind 1 is gone

**DECIDED IN SESSION.** There is no down-path compile check against a candidate.

```text
REMOVED   failure kind 1, INCOMPATIBLE POTENTIAL — "the proposed engram does not satisfy the
          current child contract," caught before Reach.
```

The consequence is accepted deliberately: an incompatible composition costs a realization before
anyone notices. That matches the design's posture — *"the system records what happened when an
entity was tried."*

### The failure set is cleaner afterward

```text
BEFORE                                    AFTER
1  INCOMPATIBLE POTENTIAL  the contract   1  UNREALIZABLE DECISION   REACH
2  UNREALIZABLE DECISION   Reach          2  FAILURE TO COMPILE      THE CHILD   ← new
3  UNSATISFIED RETURN      the parent     3  UNSATISFIED RETURN      THE PARENT
```

Still three, but **each owner is now a distinct party** — boundary, child, parent — rather than two
of them being contracts caught at different times. And they map onto the escalation ladder without
overlap:

```text
1  nothing to repair here. the composition never became live.
2  the CREATOR path. the creator pops and fixes its own body.
3  the WAKE path. the parent wakes into a focused frame.
```

---

## 5 · Leaves, and the only parent

**A leaf has no creation half.** There is no interior body to author —
`surfaces-and-solidification.md`: *"LEAF has NO body. the entity's interpretation of the expectation
produces the observation."*

```text
LEAF          compiler slot only. originates at the crossing.
              runs Recognize because RECOGNIZE IS THE FIRST PART OF THE RETURN PATH.
ORCHESTRATOR  both slots.
```

### The compiler is small — it checks its child's type

**REVISED 08-15.** Michael: *"the compiler is just 'check your child's type.' That's it."*

Not a machine, and not a second interpretation layer. The compiler slot holds whatever performs the
return check for this engram's child positions — which is Recognize, §4.

**A leaf's child is the world.** It has no child position, so what it checks against is the one
thing guaranteed to differ on every invocation. Pid churn is the purest case: it changes on every
restart and nothing about the return's meaning changed with it.

```text
IF THE LEAF CAN MAP THE WORLD'S ANSWER BACK INTO ITS CONTRACT
    nothing wakes. no repair, no escalation, no creator.
IF IT CANNOT
    the parser returns 0, and only then does anything happen.
```

**The contract is a variance absorber** — it states how far the world may move before anyone has to
care. `2026-08-09/interfaces.md` had this from the start: *"{tests: 480} vs {tests: 500} —
comparable. THE WORLD MOVED."* That case never woke anything, and was never meant to.

This is why leaves carry both jobs at once: they are where novel material enters
([`primitives-as-engrams.md`](primitives-as-engrams.md) §7) **and** where world variance is
absorbed. Same boundary.

A leaf therefore cannot have failure kind 2. Every leaf failure is a return failure, parent-owned —
which agrees with the tiering in [`primitives-as-engrams.md`](primitives-as-engrams.md) §8, where a
malformed leaf return is *child performed wrong* and re-drives without waking anyone.

### Recognize is the only parent for non-namespace possibilities

```text
EVERY RETURN           judged by Recognize.
THE OUTERMOST RETURN   judged by ACCEPTANCE. no parent exists.
```

`observable-causality.md` §4: *"The outermost caller is the final authority for its branch."*

This is the same induction as `primitives-as-engrams.md` §7, and it lands on the same exception:
**the namespace**, which is also the object with no `Position` in H11. Two holes turning out to be
one fact about the outer edge.

---

## 6 · The creator is not told its behavior will be compiled

**DECIDED IN SESSION.** And it is not a new rule — it is the reception rule reaching the authoring
view.

> `../../semantics/terminology.md`: *"That structural bound preserves the rule that **a possibility
> cannot observe its own reception** or optimize directly for being cited."*

Knowing your output will be compiled is knowing something about your reception. Same family as
`age.md`'s *"age is a resolution coordinate, never a selection input"* — a signal read at the wrong
moment becomes something to optimize toward.

### And there is no choice to withhold

The creator does not pick the syntax either. `continuation-language.md`: *"The LLM does not invent
the environment in which it writes. The runtime supplies a frame in which that code is legal."*

```text
FRAME     supplies the syntax
   ↓
SYNTAX    determines the compiler
   ↓
CREATOR   fills in a body. never chooses, never needs to know.
```

Nothing is withheld, because there was never a decision point. That is the stronger form of the
same guarantee.

### Different views, which the design already does

Leaf and orchestrator are derived from whether a possibility has selected, and `terminology.md`'s
introspect already reports two different things off one descent — *"WHAT AM I DOING"* versus
*"WHAT HAS THIS DONE."* The creation view and the interpretation view are the same move applied to
the two slots.

### The failure path leaks it, and that is correct

Waking a creator to fix a compile failure necessarily reveals that compilation exists. The leak has
a shape:

```text
DURING AUTHORING   nothing. no prospective knowledge to optimize against.
ON FAILURE ONLY    and only about the specific thing that failed.
```

Which is the focused-frame discipline already in the set: retrospective, scoped, never in advance.

```text
THE LINE TO HOLD, because the two look alike and only one is permitted:

LEARNING WHAT COMPILES     learning to do the work.        ALLOWED — and continuation-language.md
                           already wants it: "the caller learns locally which candidate
                           expectation serves this engram's particular purpose."

OBSERVING RECEPTION        learning to flatter the judge.  REFUSED.
```

---

## 7 · The blocker

**The creator is not recorded, so `pop()` has nothing to pop to.**

> `../../06-entity.md`: *"`Author` is the bootstrap... It is also `questions.md §2` made concrete —
> **the authored expression appears in no citation.**"*

Every engram a creator wrote is uncited work. There is no recorded edge from a body back to whatever
authored it, so the failure path in §1 and §4 is unimplementable as things stand.

```text
questions.md §2 WAS   a curiosity about attribution.
IT IS NOW             the thing that makes the creator path impossible.
```

This converges with H16's citation question from the other direction. H16 asked *is the creator
recorded as a cause*; this says **it has to be**, and names what breaks without it.

---

## 8 · What this contradicts

So none of it is worked stale:

```text
engram-contracts-and-composition.md   failure kind 1, and acceptance test 6 — "Decide composes an
                                      executable engram with a typed reach observation," checked at
                                      composition. §4 removes that check.

age.md                                entity is one of four parent-owned per position and a change
                                      versions it. §3 makes the compiler engram-owned and fixed at
                                      solidification.

durable/01-engram.md §2               "NO Entity FIELD." an engram now carries two slots.

kernel/07-recognize-orient.md         Recognize's placement. §4 makes it return-path only.

06-entity.md                          the Registry is map[EntityID]func() Interpreter. two slots
                                      do not fit one map.
```

---

## 9 · Tools — software with no creator slot

**LEANING, NOT DECIDED.** Michael: *"its probably a tool."* Recorded because the question was
raised as needing a decision now.

### The mapping that prompted it

```text
THE LLM        the creator.                            authors the body.
THE BODY       what it writes.                          markup, script, continuation.
THE COMPILER   the named interpreter of that body.      a rendering engine, python, the runtime.
```

Consistent with §2 — a browser engine interpreting LLM-authored markup is an ordinary compiler slot.

**But "a browser someone wants to use" is a different object.** Nothing inside skyra authored it,
and there is no creator to reach. Which gives the distinction a structural form rather than a
categorical one:

> **A tool is software with no creator slot. It cannot be repaired from inside.**

Nobody can pop to the thing that wrote Firefox. That is not a policy about applications — it is the
absence of a slot, derivable the same way leaf and orchestrator are derived.

### What follows

```text
NO CREATOR SLOT     failure kind 2 (FAILURE TO COMPILE, §4) is IMPOSSIBLE for a tool.
                    there is nothing to wake and nothing to fix. its failures can only be
                    kind 1 (unreachable) or kind 3 (its return did not satisfy the parent).

NOT AN ENGRAM       no contracts, no continuation. therefore NOT COMPOSABLE — it cannot occupy a
                    child position, because a child position requires contracts to occupy it.

STILL RECORDED      it runs under endpoints like everything else, so its EFFECTS are in the causal
                    record even though its interior is not. same as any entity. P18's corollary.
```

### It is a fourth floor in a pattern that keeps recurring

```text
EXECUTION      bottoms out at a LEAF       no child return to wait on
TYPE CHECK     bottoms out at a PRIMITIVE  kernel-produced, trusted by provenance
REACH          bottoms out at a HOST       nothing beneath it to reach
COMPOSITION    bottoms out at a TOOL       nothing to compose with
```

Which suggests a tool sits behind a leaf the way a host sits behind Reach — the terminal thing that
is used but is not part of the recursion.

### CONFIRMED, with a requirement attached

Michael: *"this is correct — structurally we can change it. But we can preserve the causal record
of it."*

So the absence of a creator slot means skyra cannot **repair** a tool. It does not mean tools are
static — they are replaced, updated, and versioned from outside. What must survive is the record:

> **Skyra cannot own a tool's change. It must still witness it.**

This is not optional bookkeeping. `../../06-entity.md` already makes the argument in its
statelessness rule:

> *"Anything accumulating between invocations is world state and must be carried explicitly; if
> state hid inside an interpreter, two invocations that look identical would give different answers
> and **same question becomes a lie**."*

A tool silently changing version is exactly that failure. Two invocations identical in the record
produce different results, and the same-question fold — which `Partitions()` depends on — starts
grouping incomparable things.

```text
THEREFORE   a tool needs a recorded identity AND version on every invocation that used it,
            even though nothing inside skyra can author, repair, or version it.

WITNESSED, NOT OWNED   the same posture the graph already takes toward the world: it records what
                       happened, and does not claim authorship of it.
```

Note this is a *weaker* requirement than the engram version chain in `../../reach/age.md` — there
is no `Prev` pointer to walk and no positional coordinate, because nobody provisioned the change.
It is closer to `Invocation.Host`: recorded, not keyed.

### A tool is an address, not an identity — and this generalizes

Michael: *"This is also a problem with just api endpoints in general. We can't see what changed
behind that boundary. Which is why I don't think it's an entity... it's an address we don't have
control over. The runtime can facilitate opening that process but it's not responsible for that
process. It's responsible for the communication to the process and from the process."*

The structural difference from an entity is exact:

```text
ENTITY    the runtime INVOKED it, so it knows WHICH BEHAVIOR produced the observation.
          provenance is SUBSTRATE-KNOWN.

ADDRESS   the runtime sent bytes somewhere and bytes came back.
          it knows WHERE IT SENT. it does not know WHAT PRODUCED.
```

So `Provenance.Engram` would be a **fiction** for a tool return. There is nothing on the other side
the substrate witnessed. That is the real reason it is neither entity nor engram — not the missing
creator slot alone, but the missing grounded identity underneath it.

**This is probably already typed.** `../../02-content.md`, in the closed Value set, `SPEC`:

```go
type WorldRef struct {
    Host   ids.HostID
    Handle string
}
```

An address at a host.

### But an address alone is not enough — REVISED 08-15

`Handle` as a bare address was considered and rejected in session. **It has to carry process
identity.**

```text
ADDRESS ALONE     two reaches to the same address may hit two different running things.
                  a restart is INVISIBLE, and the record claims they were the same.

ADDRESS : PID     a restart is a VISIBLE discontinuity.
```

The reason is causality, and it is the same argument `../../06-entity.md` makes for statelessness:

> *"if state hid inside an interpreter, two invocations that look identical would give different
> answers and **same question becomes a lie**."*

A silently replaced process behind a stable address is exactly that lie, and `Partitions()` would
fold the two together and diff across them.

**This refines what §9 said above.** Two different things were being collapsed under
"unknowable":

```text
WHAT IS INSIDE THE PROCESS       genuinely unknowable. do not record it, do not self-report a
                                 version. THE WORLD MOVED absorbs it.

WHETHER IT IS THE SAME PROCESS   KNOWABLE, and substrate-observed — the runtime spawned it or
                                 connected to it, so it holds the pid. NOT a self-report, so P4
                                 is satisfied.
```

So the record does not claim to know what changed. It only refuses to claim continuity it cannot
witness.

**And this granularity stays in the record — it must not reach the partition key.** If pid entered
the key, every restart would fork a new partition and leave nothing to diff against. See
[`durable/02-graph.md`](durable/02-graph.md) §4:

```text
THE RECORD asks   did continuity hold?    address:pid. never claims more than it witnessed.
THE FOLD asks     are these comparable?   address only. a restart is THE WORLD MOVING.
```

```text
OPEN, AND A REAL LIMIT   a pid exists for a process the runtime spawned. A REMOTE HTTP ENDPOINT
                         HAS NONE — you cannot observe process identity on someone else's machine.
                         What plays the pid's role there, if anything, is unresolved. Absent an
                         answer, remote-address returns carry weaker continuity than local ones,
                         and the fold has to tolerate that.
```

### Security at a provisioned endpoint — DEFERRED BY DECISION

A tool return can only be checked **structurally** — nominal typing needs substrate-known
provenance, and an address has none. So the delegation/injection guarantee from `../../02-content.md`
— *"same bytes, different provenance, and only the type keeps them apart"* — **does not hold at tool
positions.** Anything that can talk to that address can produce satisfying bytes.

Michael: *"it's not a security concern if it's provisioned by the user. That's a userspace problem
... unsecure endpoints are on the user to not provision."*

```text
THE SYSTEM PROVIDES   the mechanism — endpoints live in Horizon, Horizon is provisioned, and
                      provisioning can only narrow.
THE USER DECIDES      whether a given endpoint is worth trusting.
```

Same posture as file permissions: the OS supplies `chmod` and does not stop you choosing `777`.
Recorded as a known, accepted weakness rather than an unnoticed one — and it is the reason a parent
composing with a tool position should know it is getting weaker typing than an engram position
would give.

### Which resolves the version question by dissolving it

```text
DO NOT RECORD   "Firefox 143."   ← a self-report. P4 refuses it.
DO RECORD       what was sent · to which address · when · under what authority · what came back.
```

The change behind the boundary is **unknowable**, so the record does not attempt it. And the
same-question fold does not break, because this is the case it was built to absorb —
`2026-08-09/interfaces.md`: *"{tests: 480} vs {tests: 500} — **THE WORLD MOVED**."* A tool changing
behind its address *is* the world moving.

### And statelessness stops being required out there

`../../06-entity.md` makes statelessness structural — *"a constructor per entity, not an instance"* —
because a stateful interpreter makes *same question* a lie.

```text
AN ENTITY   claims reproducibility. must be stateless BY CONSTRUCTION.
AN ADDRESS  never claimed it. the requirement does not apply, and nothing depended on it.
```

The rule stays exactly as strict where it matters, and simply has no jurisdiction where it does not.

### Where it belongs

Spawning is a host capability — `Spawn`, `ExecutionOps` — and the result is a handle at a host.
That is **H15**'s territory: *"the host layer has no representation in this surface."* Tools are not
a gap in the entity model; they are a job for the layer already flagged as unmodelled.

```text
OPEN   whether a tool is INVOKED through a leaf engram whose terminal behavior hands off to it, or
       is reached some other way. the leaf reading costs nothing and preserves the record, but was
       not discussed.

OPEN   what plays the pid's role for a REMOTE endpoint, where process identity is unobservable.
       see the limit noted above. this is the residue of the address-stability question and the
       only part of it still unanswered.
```

---

## 10 · Open edges from this session

```text
1  WHAT FILLS A CREATOR SLOT      §7. until authorship is cited, the slot cannot be populated
                                  from the record.

2  DOES THE POP RECURSE?          if the creator's fix also fails to compile, does the creator's
                                  creator pop? everything else in these two sessions got a floor —
                                  leaves, primitives-by-provenance, hosts. this needs one.
                                  Author is already called "the bootstrap: the one place the design
                                  requires something be handed to it," which is the likely floor.

3  WHAT REPLACES PARTITION-ON-    §3 says entity-partitioning is pre-solidification machinery. what
   ENTITY AFTER SOLIDIFICATION    keys the fold once the compiler is fixed is not addressed.

4  PRINCIPAL                      terminology.md: "PRINCIPAL — WHO ASKS. an entity that can
                                  INITIATE." if Entity dissolves, what a principal is becomes
                                  undefined. NOT RAISED IN SESSION — recorded here because the
                                  authority model depends on it and H3 is already unreconciled.

5  REACH RESOLVES WHAT?           Reach resolved an entity to a host. presumably now the compiler.
                                  renaming only, but unstated.
```

---

## 11 · What is NOT from this session

```text
CITED FROM SOURCE   the reception rule · the interior-body rule in age.md · WHO COMPILES ·
                    the prose-vs-compiled shape claim · the leaf-has-no-body rule ·
                    the outermost-authority rule · questions.md §2.

DECIDED IN SESSION  kind 1 removed · Recognize is return-path only · creator/compiler as slots ·
                    both fixed at solidification · the creator is not told.

PROPOSED, NOT       the temporal reconciliation in §3 · the learning/gaming line in §6 ·
DECIDED             everything in §9.
```
