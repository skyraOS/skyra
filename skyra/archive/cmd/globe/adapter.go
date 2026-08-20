package main

import (
	"fmt"
	"time"

	dnav3 "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v3"
	"skyra-v1/src/skyra/spacetime"
	"skyra-v1/src/skyra/distributed-agent/world"
	"skyra-v1/src/skyra/distributed-agent/world3d"
)

// The bridge: dna-v3 mounted on the globe, over the time kernel.
//
// The being touches its world through dnav3.Realm — Do(Act) and Perceive(Triple). Here
// both are answered by the GLOBE:
//
//   - Do sends the act into the world as an impulse (through the being's proxy Space,
//     which is placed on the globe like any agent) and BLOCKS until the world's return
//     wave arrives. A move blocks for the whole walk — distance costs real ticks, so
//     LAW 5's `within` deadlines are physical out here.
//
//   - Perceive is the settler's targeted look: a direct instrument read. The panoramic
//     wave is for taking in a scene; settlement is checking one dial.
//
// The translation both directions is EXACT — world reports to triples, no model anywhere.

// beingProxy is the being's body on the globe: a Space that can be placed, perceived by
// others, and — the part the adapter uses — receives the world's return waves.
type beingProxy struct {
	id      string
	slice   skyra.Slice
	returns chan skyra.Impulse
}

func newBeingProxy(id string) *beingProxy {
	return &beingProxy{
		id:      id,
		slice:   world.NewSlice("proxy:"+id, id),
		returns: make(chan skyra.Impulse, 256),
	}
}

func (b *beingProxy) ID() string                            { return b.id }
func (b *beingProxy) Slice() skyra.Slice                    { return b.slice }
func (b *beingProxy) NewActivation() skyra.ActivationRecord { return proxyActivation{} }
func (b *beingProxy) Content() any                          { return nil }
func (b *beingProxy) Agents() map[string]skyra.Space        { return nil }
func (b *beingProxy) Children() map[string]skyra.Space      { return nil }
func (b *beingProxy) Express(o skyra.Space, i any) any      { return nil }

func (b *beingProxy) Observe(o skyra.Space, impulse any) {
	imp, ok := impulse.(skyra.Impulse)
	if !ok || !imp.Return {
		return // inbound engages etc. — not the adapter's lane (yet)
	}
	select {
	case b.returns <- imp:
	default: // a full buffer must not deadlock a wave; drop the oldest lane instead
	}
}

type proxyActivation struct{}

func (proxyActivation) Activate(bool, int) float64 { return 1.0 }

// globeRealm satisfies dnav3.Realm over the globe world.
type globeRealm struct {
	world   *world3d.World
	kernel  skyra.Time
	proxy   *beingProxy
	timeout time.Duration
}

func newGlobeRealm(w *world3d.World, kernel skyra.Time, proxy *beingProxy) *globeRealm {
	return &globeRealm{world: w, kernel: kernel, proxy: proxy, timeout: 60 * time.Second}
}

// await drains return impulses until the wanted primitive arrives. Unrelated returns
// (a stray perceive notification) are discarded — the being asked one question.
func (r *globeRealm) await(prim string) (skyra.Impulse, bool) {
	deadline := time.After(r.timeout)
	for {
		select {
		case imp := <-r.proxy.returns:
			if imp.Primitive == prim {
				return imp, true
			}
		case <-deadline:
			return skyra.Impulse{}, false
		}
	}
}

func (r *globeRealm) Do(a dnav3.Act) dnav3.Outcome {
	// A catch is an authoritative world event. If another seeker won while this being
	// still had an expression in flight, return that observation and supersede the
	// remaining acts. The being stops because the world told it why, not because the
	// game loop reached into cognition and canceled it invisibly.
	if by := r.world.CaughtBy(); by != "" && by != r.proxy.ID() {
		return dnav3.Outcome{Status: "superseded", Effects: []dnav3.Triple{
			{Subject: "snitch:/", Predicate: "status", Object: "caught"},
			{Subject: "snitch:/", Predicate: "caught-by", Object: by},
		}}
	}
	switch a.Primitive {
	case "move":
		return r.move(a)
	case "take":
		return r.task(world3d.PrimitiveTake, "sigil:/", "carried")
	case "activate":
		return r.task(world3d.PrimitiveActivate, "beacon:/", "status")
	case "catch":
		return r.catch()
	case "perceive":
		return r.Panorama()
	}
	return dnav3.Outcome{Status: "unavailable", Err: fmt.Sprintf("no primitive %q on the globe", a.Primitive)}
}

func (r *globeRealm) move(a dnav3.Act) dnav3.Outcome {
	r.kernel.Realize(r.proxy, r.world, skyra.Impulse{
		Primitive: world3d.PrimitiveMove,
		TargetID:  a.Target,
	})
	imp, ok := r.await(world3d.PrimitiveMove)
	if !ok {
		return dnav3.Outcome{Status: "indeterminate", Err: "no arrival before timeout"}
	}
	switch res := imp.Payload.(type) {
	case world3d.MovementResult:
		return dnav3.Outcome{Status: "succeeded", Effects: []dnav3.Triple{
			{Subject: "agent:self", Predicate: "at", Object: "coord:" + res.Cell},
		}}
	case world3d.MoveBlockedResult:
		return dnav3.Outcome{Status: "blocked", Err: res.Reason}
	}
	return dnav3.Outcome{Status: "indeterminate", Err: "unrecognized movement payload"}
}

// task runs take/activate. The world reports what happened and, on failure, the first
// missing prerequisite — surfaced as `<subject> missing <name>`, the same hint channel
// as a failed build.
func (r *globeRealm) task(prim, subject, predicate string) dnav3.Outcome {
	r.kernel.Realize(r.proxy, r.world, skyra.Impulse{Primitive: prim})
	imp, ok := r.await(prim)
	if !ok {
		return dnav3.Outcome{Status: "indeterminate", Err: "no outcome before timeout"}
	}
	res, isTask := imp.Payload.(world3d.TaskResult)
	if !isTask {
		return dnav3.Outcome{Status: "indeterminate", Err: "unrecognized task payload"}
	}
	effects := []dnav3.Triple{
		{Subject: subject, Predicate: predicate, Object: r.stateWord(prim, res.OK)},
	}
	if res.Missing != "" {
		effects = append(effects, dnav3.Triple{Subject: subject, Predicate: "missing", Object: res.Missing})
	}
	return dnav3.Outcome{Status: "succeeded", Effects: effects}
}

// catch swipes at the snitch. A failed swipe is also a LOOK: the world reports where the
// snitch is right now, so every miss refreshes the intel the next plan runs on — which is
// the only way a planner this deliberate can hunt a moving target at all.
func (r *globeRealm) catch() dnav3.Outcome {
	r.kernel.Realize(r.proxy, r.world, skyra.Impulse{Primitive: world3d.PrimitiveCatch})
	imp, ok := r.await(world3d.PrimitiveCatch)
	if !ok {
		return dnav3.Outcome{Status: "indeterminate", Err: "no outcome before timeout"}
	}
	res, isTask := imp.Payload.(world3d.TaskResult)
	if !isTask {
		return dnav3.Outcome{Status: "indeterminate", Err: "unrecognized catch payload"}
	}
	if res.OK {
		return dnav3.Outcome{Status: "succeeded", Effects: []dnav3.Triple{
			{Subject: "snitch:/", Predicate: "status", Object: "caught"},
			{Subject: "snitch:/", Predicate: "caught-by", Object: r.proxy.ID()},
		}}
	}
	effects := []dnav3.Triple{
		{Subject: "snitch:/", Predicate: "status", Object: "free"},
		{Subject: "snitch:/", Predicate: "missing", Object: res.Missing},
	}
	if at, placed := r.world.SnitchAt(); placed {
		effects = append(effects, dnav3.Triple{Subject: "snitch:/", Predicate: "at", Object: "coord:" + at.ID()})
	}
	return dnav3.Outcome{Status: "succeeded", Effects: effects}
}

func (r *globeRealm) stateWord(prim string, ok bool) string {
	if prim == world3d.PrimitiveActivate {
		if ok {
			return "lit"
		}
		return "unlit"
	}
	if ok {
		return "true"
	}
	return "false"
}

// Panorama runs the perceive WAVE and folds everything it returned into triples. Used to
// seed each episode: the being opens its eyes before it plans.
func (r *globeRealm) Panorama() dnav3.Outcome {
	r.kernel.Realize(r.proxy, r.world, skyra.Impulse{Primitive: world3d.PrimitivePerceive})
	imp, ok := r.await(world3d.PrimitivePerceive)
	if !ok {
		return dnav3.Outcome{Status: "indeterminate", Err: "no perception before timeout"}
	}
	res, isPerception := imp.Payload.(world3d.PerceptionResult)
	if !isPerception {
		return dnav3.Outcome{Status: "indeterminate", Err: "unrecognized perception payload"}
	}
	effects := []dnav3.Triple{
		{Subject: "agent:self", Predicate: "at", Object: "coord:" + res.Self},
		{Subject: "shrine:/", Predicate: "location", Object: "coord:" + res.Shrine},
		{Subject: "beacon:/", Predicate: "location", Object: "coord:" + res.Beacon},
		{Subject: "beacon:/", Predicate: "status", Object: litWord(res.BeaconLit)},
	}
	for id, cell := range res.Agents {
		effects = append(effects, dnav3.Triple{Subject: "agent:" + id, Predicate: "at", Object: "coord:" + cell})
	}
	if at, placed := r.world.SnitchAt(); placed {
		effects = append(effects,
			dnav3.Triple{Subject: "snitch:/", Predicate: "at", Object: "coord:" + at.ID()},
			dnav3.Triple{Subject: "snitch:/", Predicate: "status", Object: snitchWord(r.world.CaughtBy())},
		)
	}
	return dnav3.Outcome{Status: "succeeded", Effects: effects}
}

// Perceive is the settler's targeted read — one dial, answered directly by the world's
// instruments. Absence is decidable here for the same reason as the filesystem: the
// perceive is authoritative for exactly the triple that was bet on.
func (r *globeRealm) Perceive(t dnav3.Triple) dnav3.Outcome {
	report := func(object string) dnav3.Outcome {
		return dnav3.Outcome{Status: "succeeded", Effects: []dnav3.Triple{
			{Subject: t.Subject, Predicate: t.Predicate, Object: object},
		}}
	}
	switch {
	case t.Subject == "agent:self" && t.Predicate == "at":
		pos, placed := r.world.PositionOf(r.proxy.ID())
		if !placed {
			return dnav3.Outcome{Status: "unavailable", Err: "not placed"}
		}
		return report("coord:" + pos.ID())
	case t.Subject == "beacon:/" && t.Predicate == "status":
		return report(litWord(r.world.Lit()))
	case t.Subject == "beacon:/" && t.Predicate == "location":
		return report("coord:" + r.world.BeaconCell().ID())
	case t.Subject == "shrine:/" && t.Predicate == "location":
		return report("coord:" + r.world.ShrineCell().ID())
	case t.Subject == "sigil:/" && t.Predicate == "carried":
		return report(boolWord(r.world.Carrying(r.proxy.ID())))
	case t.Subject == "snitch:/" && t.Predicate == "at":
		at, placed := r.world.SnitchAt()
		if !placed {
			return dnav3.Outcome{Status: "unavailable", Err: "no snitch in this world"}
		}
		return report("coord:" + at.ID())
	case t.Subject == "snitch:/" && t.Predicate == "status":
		return report(snitchWord(r.world.CaughtBy()))
	case t.Subject == "snitch:/" && t.Predicate == "caught-by":
		if by := r.world.CaughtBy(); by != "" {
			return report(by)
		}
		return report("nobody")
	}
	return dnav3.Outcome{Status: "unavailable", Err: fmt.Sprintf("cannot perceive %s %s", t.Subject, t.Predicate)}
}

func snitchWord(caughtBy string) string {
	if caughtBy != "" {
		return "caught"
	}
	return "free"
}

func litWord(lit bool) string {
	if lit {
		return "lit"
	}
	return "unlit"
}

func boolWord(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// globeAffordances is what the globe can MEASURE — the instrument list that stamps every
// claim's mode. Note what is absent: no perceive for `worthy`, `close`, `safe`. The
// world's dials decide what kind of goal a target is, same as everywhere.
func globeAffordances() *dnav3.Affordances {
	return dnav3.NewAffordances(map[string][]string{
		"agent":  {"at"},
		"beacon": {"status", "location"},
		"shrine": {"location"},
		"sigil":  {"carried"},
		"snitch": {"at", "status", "caught-by"},
	})
}

// globeSystem is the planner prompt for this world: same FORMAT as everywhere, different
// instrument list. It describes the job and the dials, and — deliberately — nothing about
// scoring (see present.go for why that line is held).
const globeSystem = `You act on a small globe. You stand on a cell; cells are addressed coord:f,x,y.

For every act, state what world state you expect it to produce, and how likely you think that is.

FORMAT — reply with only this:

<framing>one sentence: what you think is going on and what you are trying to do</framing>

<act id="a1" target="coord:2,4,4" primitive="move"></act>
<act id="a2" target="sigil:/" primitive="take" requires="agent:self at coord:2,4,4"></act>

<expect of="a1">  0.95  1  agent:self  at       coord:2,4,4  </expect>
<expect of="a2">  0.80  1  sigil:/     carried  true         </expect>

PRIMITIVES:
  move      target is a coord — you walk there. Reports the cell you ARRIVED at.
  take      target sigil:/ — pick up the sigil where you are standing.
  activate  target beacon:/ — activate the beacon where you are standing.

EXPECT LINE:  <likelihood 0..1> <within> <subject> <predicate> <object>
  likelihood  your honest probability that this comes true. Not a target and not a hope — what you
              actually think the odds are, given what you know about THIS case in particular.
  within      how many ticks until it should be checked. 1 = immediately after the act.

PREDICATES YOU CAN CHECK:
  agent:self  at        coord:f,x,y
  beacon:/    status    lit | unlit     (checked live)
  beacon:/    location  coord:f,x,y
  shrine:/    location  coord:f,x,y
  sigil:/     carried   true | false

CURRENT WORLD STATE may also show  beacon:/ missing <name>  or  sigil:/ missing <name>  — a hint
the world gives on a failed act, naming a prerequisite it still needs. Read it and act on it. Do
not PREDICT it; it is not yours to claim, only the world's to report.

Predicting anything else is pointless — nothing can check it.

An act happens WHERE YOU STAND. Whether you are standing in the right place is world state —
order your acts with requires on agent:self at.

REQUIRES:  requires="<subject> <predicate> <object>"  — an act only fires once this is OBSERVED
to be true. This is how you order work: no <parallel>, no <if>. State the condition.

RULES:
- Every act must have at least one expect.
- Use your judgement. Say what you actually think will happen.`

// seekerSystem is the duel's planner prompt. The one epistemically new thing it must
// say: the snitch MOVES, so an observed position is already history. How likely a plan
// built on stale intel is to work is exactly the judgment the expects exist to hold.
const seekerSystem = `You are a seeker in a snitch hunt on a small globe. You stand on a cell; cells are
addressed coord:f,x,y. Another seeker is hunting the same snitch. First catch wins.

For every act, state what world state you expect it to produce, and how likely you think that is.

FORMAT — reply with only this:

<framing>one sentence: what you think is going on and what you are trying to do</framing>

<act id="a1" target="coord:2,4,4" primitive="move"></act>
<act id="a2" target="snitch:/" primitive="catch" requires="agent:self at coord:2,4,4"></act>

<expect of="a1">  0.95  1  agent:self  at      coord:2,4,4  </expect>
<expect of="a2">  0.40  1  snitch:/    status  caught       </expect>

PRIMITIVES:
  move   target is a coord — you walk there. Reports the cell you ARRIVED at.
  catch  target snitch:/ — swipe at the snitch. Works only within ONE CELL of where the
         snitch actually is AT THAT MOMENT. A miss still reports where the snitch is now.

THE SNITCH MOVES. Every position you have observed for it is already old. Walking to a
stale position and swiping is a bet — an honest likelihood on a catch is usually low.
Cutting it off beats chasing it: aim where it is HEADING, not where it was.

EXPECT LINE:  <likelihood 0..1> <within> <subject> <predicate> <object>
  likelihood  your honest probability that this comes true. Not a target and not a hope — what you
              actually think the odds are, given what you know about THIS case in particular.
  within      how many ticks until it should be checked. 1 = immediately after the act.

PREDICATES YOU CAN CHECK:
  agent:self  at         coord:f,x,y
  snitch:/    at         coord:f,x,y      (live — where it is when checked)
  snitch:/    status     free | caught
  snitch:/    caught-by  <agent> | nobody

CURRENT WORLD STATE may also show  snitch:/ missing <name>  — the world naming what a
failed catch lacked (reach). Read it. Do not PREDICT it; it is not yours to claim.

Predicting anything else is pointless — nothing can check it.

REQUIRES:  requires="<subject> <predicate> <object>"  — an act only fires once this is OBSERVED
to be true. This is how you order work: no <parallel>, no <if>. State the condition.

RULES:
- Every act must have at least one expect.
- Use your judgement. Say what you actually think will happen.`
