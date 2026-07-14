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
//	observation  ->  projection  ->  expression  ->  observation
//	 (material)        (final)       (efficient)
//
// The four causes are POSITIONS IN THIS LOOP, not annotations anyone emits. The runtime
// records "I observed O, I projected P, I emitted E, I then observed O′" and the links are
// simply there, because that is the order things happened. The causal graph is not a layer
// over the knowledge graph — it is the shape of the loop through it. That is why v2's entire
// causal apparatus (CausalNodes, AitiaSet, the annotation grammar, causal-belief expansion)
// has no counterpart here.
//
// Written this way the hard arrow comes first: observation -> projection is the one edge with
// the entire cognitive act inside it. The other two are cheap — projection -> expression is a
// collapse, and expression -> observation is the world's job.
//
// # The two graphs are mutually dependent
//
// Learning TAKES PLACE in the expression graph (grades live on expects and framings). It is
// ONLY POSSIBLE via the observation graph (an expect's claim refers to observed state; its
// grade IS a diff against an observation). And the observation graph is itself SHAPED BY the
// expression graph — you only observe what you acted to observe. The log is not a neutral
// transcript of the world; two beings in the same room record different logs.
//
// The dependency is circular, in the shape of the loop above. Neither graph is prior. Neither
// is disposable.
//
// # The laws
//
// LAW 1 — Observation is the only writer of world state. Expression has no write access.
// Writing a file PROJECTS "file exists"; the observation graph does not contain it until
// something came back and said so. The being can be wrong, but it cannot believe its own
// output into reality. World state is therefore a fold over the observation log:
// deterministic, replayable, auditable, fiction-free.
//
// LAW 2 — Learning happens only across the boundary. MINE (projections, actions) vs THE
// WORLD'S (observations). A grade is by definition a cross-boundary diff, so grading your
// own output is a TYPE ERROR, not a rule anyone has to remember.
//
// LAW 3 — Exact first, always. The model only ever sees the residual. Node resolution, the
// diff, retrieval, and impulse handling each run an exact pass first and hand the model
// only what is left. This bounds cost and keeps the model out of the trusted path.
//
// LAW 4 — The aligner may translate; it may never judge. It runs ONCE, on the observation
// side only. Predictions are literal stated claims and are never interpreted. Ambiguous
// references are preserved as low-weight candidates, never hard-linked — forcing a hard
// link too early corrupts the graph permanently.
//
// LAW 5 — Every act must be falsifiable. Every act appears in at least one expect, or the
// emission is rejected by the parser. A prediction that cannot be wrong is a vibe, not
// cognition.
//
// # The measurement
//
// Over N repeated exposures to one task class, passes-to-closure decreases AND prediction
// error decreases. Both curves down is the thesis; either alone is a warning, and either
// failure localizes.
//
// See design-docs/skyra-dna-v3/.
package dnav3
