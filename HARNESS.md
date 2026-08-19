# The Harness

**Date opened:** 2026-08-15
**Status: FIRST DRAFT.** A build plan for one artifact that does three jobs. Not a spec — the
semantics live in `skyra/specs/`, and this consumes them rather than deciding anything.

> **One rig: real integrations behind fully anchored addresses, instrumented end to end, with the
> kernel able to contact the user when it hits something new. It is the development instrument,
> the evidence generator, and the demo — and it must be all three or it is not worth building.**

---

## 1 · Why one artifact and not three

The temptation is a test rig, then separately a demo, then separately whatever produces numbers for
the research track. That is three codebases that drift.

```text
DEVELOPMENT INSTRUMENT   exercises the substrate against a world that pushes back. The only way
                         to find out whether the design survives contact.

EVIDENCE GENERATOR       produces the measurements: crossings, escalations, over-grant delta.
                         Numbers that no design argument can supply.

DEMO                     the same run, watched. If the instrumentation is good, the demo is a
                         view over it rather than a separate thing to build and maintain.
```

They want the same events. Build the event stream once.

---

## 2 · Every integration takes a fully anchored address

**The hard rule, and the thing that makes the confinement claim testable.**

The design already refuses to guess: `{path: "file.txt"}` names no host and no namespace, so it
does not resolve, and Horizon cannot repair a missing referent — it can bound authority to an
anchored world but cannot decide which world was meant. Guessing creates both location and
authority.

The harness enforces that at every integration boundary, with no exceptions:

```text
REFUSED     "send an email"                      to whom, from which account, in which tenant
            "post to Slack"                      which workspace, which channel
            "read the file"                      which host, which namespace
            "query the database"                 which instance, which schema

REQUIRED    a WorldRef carrying host + the name its namespace resolves.
            the scheme IS the resolution procedure.
```

```text
WHY THIS IS THE WHOLE POINT   an integration that accepts a bare name has silently chosen a
                              world on the caller's behalf. Every confinement claim downstream
                              of that is void, because the boundary was drawn after the
                              decision that mattered.

                              If an integration cannot be written this way, THAT IS THE FINDING.
                              Record it rather than adding an exception.
```

---

## 3 · What gets instrumented

The event sink already has a shape in the spec set — `Opened · Crossed · Emitted · Closed ·
Refused`, with a span on every method as the correlation key. The harness implements that sink and
derives everything else from it.

```text
CROSSINGS       one per Step. THE CLOCK, and the unit that costs money.
TOKENS          per crossing. mass shipped.
FRAME MASS      what a possibility weighs at rest, per step. the growth curve is the thing.
ESCALATIONS     failures that reached a human, and at what depth they were caught first.
FIRST-TIMES     a position or branch exercised for the first time. see §4.
GRANTED         authority provisioned at admission.
EXERCISED       authority actually spent.
                └─ the delta is the product's headline number.
DISPOSITIONS    staged · landed · discarded · failed. every effect accounted for, none silent.
```

```text
THE TWO CURVES THAT MATTER

    OVER-GRANT DELTA    granted ÷ exercised.   should fall as the runtime narrows.
    ESCALATION RATE     reached a human ÷ failures.   should fall as structure accumulates.

Neither is meaningful as a single measurement. Both are only meaningful as a slope, which means
the harness has to run the same task class REPEATEDLY. Design it for that from the start —
a rig that runs a scenario once cannot produce either number.
```

---

## 4 · First contact — and there is no tutorial subsystem

When a branch or a position is exercised for the first time, the system stops and tells the user
what it is about to do.

**Nothing about that is special machinery.** It is an entity — one that knows how the system works
and teaches the user — registered like any other, selected like any other, and accumulating
history like any other. So it gets *better at teaching* by the same mechanism everything else
improves by, instead of being a static script somebody maintains.

```text
FIRST TIME       novel. the user is consulted, shown the anchored address, the authority the
                 action requires, and the effect that would land.

THEREAFTER       silent. the decision was made and recorded; asking again is noise.
```

This is one behavior serving three purposes at once, which is why it is worth building early:

```text
ONBOARDING       the tutorial. the user learns the system by being shown its actual decisions
                 at the moment they are made, instead of by reading documentation.

CONFINEMENT      review-before-effect. the security posture from the commercial track, enforced
                 at runtime rather than promised in a report.

EVIDENCE         a first-time event is a DISCOVERY event — the constructed claim that nothing
                 known covers this. Counting them is counting novelty.
```

```text
⚠️  PERMISSION-PROMPT FATIGUE IS A KNOWN FAILURE MODE. Every OS that has tried this has trained
    its users to click through. Two things make this version survivable and both are load-bearing:

    IT ASKS ONCE, PER POSITION, NOT PER ACTION. the ask is attached to a structural position,
    so a system doing the same thing a thousand times asks once.

    THE COUNT SHOULD FALL. if first-time events are not trending toward zero on a repeated task
    class, the system is not accumulating structure and the prompt is a symptom, not a feature.
    THAT IS A MEASUREMENT, AND THE HARNESS IS WHAT TAKES IT.
```

### 4a · Two entities, split on authority

The guide teaches. Something else arranges the surface. **They are two entities, and the reason is
authority rather than observed behavior** — which is worth stating, because it overrides the usual
default.

`world.md` says start coarse and split on evidence. That is right when the difference is behavior
you have not observed yet. Here it is known up front: the guide needs no effect grant ever, and a
compositor needs projection authority permanently. Merging them makes the read-only one hold power
it never uses, for its whole life, which is the thing this architecture exists to prevent.

The entity test agrees independently — *if a failure in one should not count against the other,
they are two.* Explaining badly and laying out badly are different failures and must not share a
track record.

```text
THE GUIDE       READ ONLY. no effect classes, in any grant, ever.
                reach: the user, and whatever it needs to read.
                it may Decide — it has to compose in order to look things up — but nothing
                beneath it can hold an effect grant, because it holds none to provision.

THE COMPOSITOR  holds PROJECTION authority. arranges surfaces. its own history, its own
                failures, its own record.
```

### 4b · Monotone narrowing does the work, and does it better than staging

An earlier reading of this had the guide *staging* a surface change for the user to accept. That is
weaker than what the model already gives, and it is wrong:

> **A read-only guide cannot provision projection authority to a child, because it does not hold
> any. It cannot invoke the compositor. Not "it proposes and you approve" — it structurally
> cannot, even if it is compromised.**

So the flow uses no new mechanism and no staging at all:

```text
GUIDE        proposes. the proposal is AN ORDINARY OBSERVATION — no effect, no grant, no stage.
USER         decides, in their own branch, under their own authority.
COMPOSITOR   holds projection. arranges the surface.
```

And it does not become click-fatigue, for the same reason first-contact does not: the user
provisions projection authority **to a position, once**. After that the compositor works inside
Horizon and nobody is asked again.

```text
WHAT THIS BUYS THE DEMO   the product's own front door is a live confinement proof. The agent
                          teaching you the system cannot touch anything, and that is enforced by
                          the same arithmetic as everything else rather than by a promise.
```

### 4c · The projection seam, which answers a parked question

`future-research/projection-and-representation.md` asks whether a projection is a host class or an
entity. It is both, and the seam is the one `host.go` already states — *whether anything
INTERPRETS*:

```text
WHAT THE SURFACE SHOULD CONTAIN   interpretation, authored elsewhere.   → THE COMPOSITOR, an entity
RENDERING IT                      mechanism, fixed reviewed semantics.  → the projection host class
                                                                          (still unnamed in host.go)
```

The question was only hard while it was being asked about one thing.

And the same document's third question — *what accumulates?* — answers itself once the compositor
is an ordinary entity: **the arrangements that persist are the ones the user cited.** The surface
learns by the credit mechanism, which is the difference that document was reaching for between *a
surface you author* (synthesis, learns nothing) and *a surface that accumulates* (selection). It
also satisfies the constraint stated there: the user should not have to design their own desktop.

---

## 5 · "The kernel approaches the user" — the mechanism

The instinct is right and the primitive is worth naming precisely, because the design already has
both halves and needs no new one.

```text
APPROACH   outside → inside.   should this be let in?
REACH      inside → outside.   how does an inside decision contact an outside party?
```

The kernel contacting a human is **inside → outside**, so it is `Reach`, with the user as an
entity — and a human already is one: *"delivery picks a host on the way out… a human messaged on
their TV and answering on their phone is not two events."* The user's answer arrives as an
observation with no producer, which is `Approach`.

```text
ONE CONSULTATION = ONE REACH OUT + ONE APPROACH BACK.
```

That is exactly the directional dual the primitives describe, running in the order nobody had
written down yet. Which means the guide's consultation, the escalation terminus, and ordinary user
input are **the same mechanism at three moments** — and the harness is where that gets proven or
falsified.

**Reach every time, with no human special case.** That is the stronger form of the claim and the
one worth building to: a human is an ordinary entity, so the kernel reaches one exactly as it
reaches an API, and delivery picks a host on the way out — phone, TV, email, Slack. If the two
diverge anywhere in the kernel beyond endpoint and latency, the claim is decorative.

```text
WHAT THE HARNESS MUST SHOW   that a Reach to a human and a Reach to an API differ only in
                             endpoint and latency. If the code paths diverge, the claim that a
                             human is an ordinary entity is decorative.
```

---

## 6 · Where AI codegen helps, and where it does not

```text
GOOD FOR      integration adapters. API clients, auth flows, schema mapping, pagination,
              retry-around-transport. High volume, well-specified, individually low-stakes,
              and each one is independently verifiable against a live endpoint.

              THIS IS MOST OF THE WORK BY LINE COUNT. Generate it.

BAD FOR       the substrate. The kernel, the authority model, the projection path. These are
              load-bearing, subtle, and the failure modes are silent — a capability quietly
              missing, a write outside a commit, a walk that shortens. The existing test suite
              exists because those defects are invisible without it.

THE TEST      if a defect in this code would be caught by an endpoint returning an error,
              generate it. If it would be caught only by a test that is a proof of a design
              claim, write it.
```

```text
THE RECURSION IS REAL AND IS NOT THE STARTING POINT.   integration adapters are exactly the kind
of behavior the system should eventually author, compile, and version itself. That is the
end state, not the bootstrap. Generate them conventionally now; the harness is what will later
prove the system can do it better.
```

---

## 7 · Preseeded triggers — a scaffold, and a tension worth naming

Preseeding `if` conditions so behavior is observable before any learning has happened is the right
call for a demo. It also **contradicts the architecture**, and the contradiction should be marked
in the code rather than discovered later.

```text
THE ARCHITECTURE   specification by ENCOUNTER. branches exist because something happened.
                   structure is accreted, never anticipated. selection learns; synthesis does not.

PRESEEDING         specification by ANTICIPATION. branches exist because someone guessed.
```

Two concrete hazards, not just a philosophical one:

```text
1  A PRESEEDED BRANCH THAT NEVER FIRES IS THE UNREACHABLE-BUT-CONNECTED CASE.
   Decay fires at the action boundary and GC fires on namespace disconnection. Neither reaches
   a valid, connected branch that is simply never taken. Preseeding POPULATES exactly the
   category nothing cleans up.

2  IT CONTAMINATES THE MEASUREMENTS.
   First-time counts and escalation rates are only meaningful over structure the system actually
   built. A preseeded branch that happens to be right makes the system look like it learned
   something it was told.
```

```text
THE DISCIPLINE   preseeded structure is TAGGED as such, excluded from every measurement, and
                 expires if never exercised. If a preseeded branch turns out to be correct, the
                 system should have to re-derive it from an encounter before it counts.
```

---

## 8 · Build order

```text
1  THE EVENT SINK           Opened · Crossed · Emitted · Closed · Refused, span-correlated.
                            Everything else reads from it. Nothing else works without it.

2  ONE INTEGRATION, FULLY   pick the one with the coarsest upstream scope and the most sensitive
   ANCHORED                 data — probably Slack or Gmail, per the commercial track. One, done
                            properly, beats four done loosely: the point is the anchoring rule
                            surviving contact.

3  THE TWO CURVES           run one task class repeatedly and plot over-grant delta and
                            escalation rate. IF THESE DO NOT MOVE, everything downstream is
                            premature and that is worth knowing in week three.

4  THE GUIDE                Reach-out / Approach-back, with a human as an ordinary entity.
                            READ-ONLY HORIZON, and that is the point: the front door is a live
                            confinement proof. Do NOT build the compositor yet.

5  THE VIEW                 the demo, as a read over the event stream. Not a separate build,
                            and not the compositor either — a fixed surface is fine here.

6  SECOND INTEGRATION       only now, and only to test whether anything generalizes.

7  THE COMPOSITOR           last, and optional for a long time. It needs the projection host
                            class, which does not exist and is blocked on WHERE. A fixed
                            surface costs nothing and proves the same architecture.
```

---

## 9 · Open

```text
WHAT COUNTS AS "THE SAME     the curves need a repeated task class, and "the same task again"
TASK AGAIN"                  is the same-question fold. If the fold fragments, the curves are
                             noise. This is the first thing that will break.

WHO ANSWERS THE              a consultation blocks a branch. A human is slow and a branch holds
CONSULTATION, AND HOW LONG   staged effects for its whole life. A tutorial that waits on a human
DOES THE BRANCH HOLD         for an hour is a long-running transaction with world effects
                             pending, at a scale nothing has been tested at.

DOES A CONSULTATION          the user answering is the outermost parent supplying an observation.
RESTORE ORIGINATION          the analogous case is resolved — a parent revising restores the
                             forfeit, because only the party the rule protects can waive it.
                             Probably the same. NOT CONFIRMED.

WHAT THE VIEW SHOWS          the event stream is topology and counts. Rendering it usefully to a
                             human is the projection question, which is parked pending WHERE.
                             Start with a log and resist building a dashboard out of engrams.

CAN THE GUIDE READ           it is READ-ONLY, so nothing stops it reading the event stream and
EVERYTHING?                  the graph. But "read access to everything" in a real tenant means
                             the guide can see material the user may not want an agent to see,
                             and READ IS AN EXFILTRATION SURFACE even when write is closed.
                             The confinement proof cares about EXITS, not writes. Its reach is
                             the user — so whatever it reads, it can say. UNRESOLVED, and it is
                             the first place the read-only claim will be pressed by a buyer.

HOW DOES THE COMPOSITOR      arrangements persist because the user cited them, which is the
LEARN ACROSS USERS?          credit mechanism working per namespace. Whether anything
                             generalizes across users without becoming a global goodness prior
                             is the same question second-kind-of-credit.md parks. DO NOT SOLVE
                             IT HERE.
```
