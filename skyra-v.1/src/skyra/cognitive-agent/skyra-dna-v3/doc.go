// Package dnav3 is the being's genome, rebuilt around one claim:
//
//	The being commits to a falsifiable expected world state, acts, observes what came
//	back, and the DIFFERENCE is the grade — computed, never judged.
//
// v2's ceiling was structural. Free-text predictions could only be graded by an LLM
// judge, so the learning signal bottomed out in an opinion with no independent source.
// Causality had to be ASSERTED by the model, which made the causal graph an opinion too.
// And the goal was re-derived every turn, so the being could not want anything for longer
// than one impulse. v3 removes all three by changing the representation, not by adding
// machinery.
//
// # The cycle
//
// Written from observation, because the being does not begin by wanting something — it
// begins by being told something.
//
//	observation  ->  state change  ->  action  ->  observation
//	 (material)         (final)       (efficient)
//
// (Older docs name the middle two projection and expression; they are the state-change and
// action nodes.) The four causes are POSITIONS IN THIS LOOP, not annotations anyone emits. The
// runtime records "I observed O, I interpreted it as ΔS, I emitted action A, I then observed O′"
// and the links are simply there, because that is the order things happened. The causal graph is
// not a layer over the knowledge graph — it is the shape of the loop through it. That is why v2's
// entire causal apparatus (CausalNodes, AitiaSet, the annotation grammar, causal-belief
// expansion) has no counterpart here.
//
// Written this way the hard arrow comes first: observation -> state change is the one edge with
// the entire cognitive act inside it. The other two are cheap — state change -> action is a
// collapse, and action -> observation is the world's job.
//
// # One graph, three node types — and the loop is circular
//
// There is one durable memory graph; observation, state change, and action are node types in it,
// not separate stores. Learning LIVES ON the graded relationships between them (grades live on the
// state-change → action → observation edges). It is ONLY POSSIBLE via the observation nodes (a
// claim refers to observed state; its grade IS a diff against an observation). And which
// observation nodes ever exist is itself SHAPED BY the action nodes — you only observe what you
// acted to observe. The record is not a neutral transcript of the world; two beings in the same
// room record different observations.
//
// The dependency is circular, in the shape of the loop above. No node type is prior. None is
// disposable.
//
// # The laws
//
// LAW 1 — Observation is the only writer of world state. Expression has no write access.
// Writing a file PROJECTS "file exists"; world state does not contain it until something came
// back and said so. The being can be wrong, but it cannot believe its own output into reality.
// World state is therefore a fold over the observation nodes: deterministic, replayable,
// auditable, fiction-free.
//
// LAW 2 — Learning happens only across the boundary. MINE (projections, actions) vs THE
// WORLD'S (observations). A grade is by definition a cross-boundary diff, so grading your
// own output is a TYPE ERROR, not a rule anyone has to remember.
//
// LAW 3 — Exact first, always. The model only ever sees the residual. Node resolution, the
// diff, retrieval, and impulse handling each run an exact pass first and hand the model
// only what is left. This bounds cost and keeps the model out of the trusted path.
//
// LAW 4 — Interpretation happens at OBSERVATION time, never at GRADING time. The decomposer
// translates an impulse into triples on the way in; the grader then does a pure exact match.
// The model CANNOT TOUCH THE GRADING PATH AT ALL — there is no grading-time aligner to
// misbehave. Ambiguous references are preserved as low-weight candidates, never hard-linked:
// forcing a hard link too early corrupts the graph permanently.
//
// LAW 5 — Every act must be falsifiable, and every bet gets settled. Every act appears in at
// least one expect, and every expect carries a `within`, or the emission is rejected. At
// maturity the substrate perceives the claimed triple — the bet manufactures its own evidence,
// so absence stops being indistinguishable from "I never looked." A prediction that cannot be
// wrong is a vibe, not cognition; a bet that is never settled is the same thing.
//
// LAW 6 — Observation is produced BLIND TO DESIRE. The pass that turns an impulse into
// observation triples must not see the open bets, their confidences, or the target — only the
// entity and predicate vocabulary. Enough to speak the right language; not enough to know what
// answer was wanted. Otherwise the being resolves ambiguity toward its own bets and
// manufactures the evidence that grades it, which is the poison this whole design exists to
// keep out.
//
// # The measurement
//
// Over N repeated exposures to one task class, passes-to-closure decreases AND prediction
// error decreases. Both curves down is the thesis; either alone is a warning, and either
// failure localizes.
//
// See design-docs/skyra-dna-v3/.
package dnav3
