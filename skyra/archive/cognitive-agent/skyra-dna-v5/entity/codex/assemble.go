// Package codex reaches a model through the Codex CLI.
//
// Marshalling is what interpreting IS — rendering the frame into whatever the
// entity actually reads. For a model that is a prompt.
//
//	ASSEMBLE   engram state → a frame.    bounded by the ENTITY. FULL.
//	MONITOR    observation → frame'.      same bound, same owner. DELTA.
package codex

import (
	"fmt"
	"strings"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/frame"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
)

// theGrammar is the whole action vocabulary, and it is fixed text.
//
// It is not part of the frame: a frame is what a POSSIBILITY sees, and the
// grammar is what an ENTITY must know to emit at all. Shipping it separately
// keeps "what a possibility sees" honest — it would otherwise be inflated by
// boilerplate that has nothing to do with what this possibility knows.
const theGrammar = `You are manning one possibility inside a runtime whose entire action vocabulary is
two forms. Nothing else parses. There is no API, no tool call, no function.

  SELECT — invoke one or more possibilities and await all of them.

    <select>
      <name cite=obs-AAAA entity=some-entity>obs-BBBB, obs-CCCC</engram>
      <other cite=obs-AAAA entity=some-entity>obs-DDDD</engram>
    </select>

    The OPENING tag is the name you are invoking. The CLOSING tag is always
    literally </engram>. Inside the tag are the observation ids you are passing
    as arguments, comma separated (it may be empty).

    Ids written obs-AAAA above are PLACEHOLDERS. Never emit one — every id you
    emit must be one you were actually shown below.

    cite=   REQUIRED, per member. The observation that made you able to name this.
            "Why am I reaching for this?" — not what you are passing it.
    entity= REQUIRED, per member. Who should interpret it. Pick from ADDRESSABLE.

    Cardinality IS the concurrency declaration. Two members in one select means
    neither needs the other's answer to begin, and both run together. Two
    selects in sequence means the second may use the first's result. There is
    no other way to express either.

  COMMIT — close this possibility and hand one observation up.

    <commit>{"field": "value", "other": 3}</commit>     you AUTHOR the content
    <commit><obs-DDDD></commit>                          you NAME what came back

Emit EXACTLY ONE of these. Say nothing after it.

REFERENCES. Any string of the form <obs-DDDD> inside a commit body is a
reference to that observation, and it is how credit flows. If your answer rests
on what a child returned, reference it — carrying it forward is the only way to
say you used it. You cannot claim to have used something you did not carry.`

// System is the fixed half of the crossing — the grammar, which every entity
// needs and no frame contains.
func System() string { return theGrammar }

// Assemble renders the FULL frame. This is what an entity is handed at
// invocation open.
func Assemble(x graph.Expectation, f frame.Frame) string {
	var b strings.Builder

	fmt.Fprintf(&b, "YOUR EXPECTATION — what this possibility is for:\n\n    %s\n",
		indent(x.Expression, "    "))

	writeObservations(&b, "GIVEN — the observations you were passed", f.Observations())
	writeObservations(&b, "RETURNED — what your selects handed back", f.Returned())

	writeSelectable(&b, f)
	writeAddressable(&b, f)
	writeRule(&b, f)

	return b.String()
}

func writeObservations(b *strings.Builder, title string, obs []*graph.Observation) {
	b.WriteString("\n\n" + title + ":\n")
	if len(obs) == 0 {
		b.WriteString("\n    (none)\n")
		return
	}
	for _, o := range obs {
		fmt.Fprintf(b, "\n  <%s>%s\n", o.ID, indent(o.Content.Render(), "  "))
	}
}

// writeSelectable renders what a possibility SEES on its engram's children:
// the name, the stored returns, the credit records, and the cite-source of each
// selection.
//
//	here is what I have used, what it said, when it worked, and when it did not.
//
// It sees DOWN, never UP. Its own reception lives on its PARENTS' engrams, which
// it cannot reach — which is what makes the citation signal ungameable from
// inside: a thing that cannot observe whether it is being cited cannot optimise
// for being cited.
func writeSelectable(b *strings.Builder, f frame.Frame) {
	b.WriteString("\n\nSELECTABLE — names you may invoke, and how they have gone:\n")

	v := f.Engrams()
	names := v.Names()
	if lister, ok := f.(interface{ Names() []string }); ok {
		names = lister.Names()
	}
	if len(names) == 0 {
		b.WriteString("\n    (none yet — discover is always nameable)\n")
		return
	}

	for _, n := range names {
		s, ok := v.Summary(n)
		if !ok {
			// Nameable but with no history here: a builtin, or something a child
			// just delegated. Both are selectable and neither has a record on
			// THIS engram yet.
			fmt.Fprintf(b, "\n  %s\n      no history here yet\n", n)
			continue
		}
		fmt.Fprintf(b, "\n  %s\n", n)
		fmt.Fprintf(b, "      invoked %d ·  its answer was used %d\n", s.Invocations, s.CitedCount)

		if len(s.Hops) > 0 {
			var hs []string
			for _, h := range s.HopKeys() {
				hs = append(hs, fmt.Sprintf("hop%d×%d", h, s.Hops[h]))
			}
			fmt.Fprintf(b, "      accepted at range: %s\n", strings.Join(hs, " "))
		} else if s.Invocations > 0 {
			fmt.Fprintf(b, "      accepted at range: never\n")
		}

		for _, p := range s.Partitions {
			fmt.Fprintf(b, "      · shape %s — %d asking(s), %d used\n",
				p.Shape, len(p.Invs), p.Used())
			if last := p.Invs[len(p.Invs)-1]; last.Return != nil {
				fmt.Fprintf(b, "        last: %s\n", oneLine(last.Return.Content.Render()))
			}
			for _, d := range p.Diffs() {
				for k, fromTo := range d.Fields {
					fmt.Fprintf(b, "        moved: %s  %s → %s\n", k, fromTo[0], fromTo[1])
				}
			}
		}
	}
}

func writeAddressable(b *strings.Builder, f frame.Frame) {
	b.WriteString("\n\nADDRESSABLE — entities you may name in a select:\n\n")
	if len(f.Entities()) == 0 {
		b.WriteString("    (none)\n")
		return
	}
	for _, e := range f.Entities() {
		fmt.Fprintf(b, "    %s\n", e)
	}
	fmt.Fprintf(b, "\n  You are being interpreted by: %s\n", f.Entity())
}

// writeRule states the origination forfeit.
//
//	Originate only where it could not be mistaken for composition.
//
// The forfeit fires at the FIRST select, because before selecting everything a
// possibility commits is unambiguously invented — there is no child material it
// could be confused with. The instant it selects, its commit could be either.
func writeRule(b *strings.Builder, f frame.Frame) {
	b.WriteString("\n\nWHAT YOU MAY COMMIT:\n\n")
	if len(f.Returned()) == 0 {
		b.WriteString(`    You have not selected anything. You may AUTHOR content directly:
    <commit>{...}</commit> — invent the answer, or select first and compose after.
`)
		return
	}
	b.WriteString(`    You have already selected, so you have FORFEITED origination —
    permanently, for this possibility. You may only NAME what came back:

        <commit><one of the ids under RETURNED above></commit>

    You may NOT author new content. If the answer needs a field none of your
    children produced, you must select something that produces it. Selecting
    again is always available.
`)
}

func indent(s, pad string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		if i > 0 {
			lines[i] = pad + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

func oneLine(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 200 {
		s = s[:200] + "…"
	}
	return s
}
