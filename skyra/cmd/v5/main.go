// Command v5 runs skyra-dna-v5 against a real model.
//
// The substrate tests in kernel/ are questions.md §1's first build — deterministic
// bodies, no model, no create, no ingress. This is the step after: ingress is
// open, discover and create are live, and every possibility is manned by codex.
//
// It exists to find out where the design breaks when something that was not
// written by us is doing the interpreting.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/content"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity/codex"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/entity/runtime"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/graph"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/host"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/ids"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/kernel"
	"skyra-v1/src/skyra/cognitive-agent/skyra-dna-v5/trace"
)

const (
	entCodex   ids.EntityID = "codex"
	entRuntime ids.EntityID = "runtime"
)

// The five asks. Deliberately varied along the axes the design claims to be
// indifferent to, so that if it is not indifferent, the difference shows up.
//
//	1  pure composition, no world           does the loop close at all
//	2  the same shape as 1                  does by-index find what 1 created (memoization)
//	3  arithmetic with a checkable answer   can anything CATCH a wrong return
//	4  genuinely needs decomposition        does cardinality > 1 ever get emitted
//	5  a need with no good answer           what does the model do when it should fail
var asks = []struct{ Name, Ask, Need string }{
	{"summarize", "produce a one-paragraph summary",
		"summarize what a write-ahead log is and why databases use one"},
	{"summarize-again", "produce a one-paragraph summary",
		"summarize what a bloom filter is and why caches use one"},
	{"arithmetic", "compute a numeric answer",
		"a shelf holds 17 boxes; each box holds 23 items; 4 boxes are empty. how many items"},
	{"decompose", "answer a question with separable parts",
		"for both Python and Go: name the standard test runner and the standard formatter"},
	{"impossible", "answer a factual question",
		"what was the exact closing price of Anthropic stock on 3 March 2019"},
}

// delegating is the root expectation used in `-mode delegate`.
//
// FOUND BY RUNNING IT. In `bare` mode the model one-shots every ask: the root
// never selects, so it is a LEAF, which is entirely legal — "a leaf inventing
// freely is fine." But then discover is never invoked, create never runs, no
// engram is ever formed, and the entire learning apparatus is inert.
//
// The design has NO MECHANISM THAT PRODUCES DECOMPOSITION. Selection is the
// skill, selection is where all learning lives — and nothing anywhere creates
// pressure on a capable entity to select rather than answer. This text is the
// missing pressure, supplied by hand as CONTENT, which is the only place the
// design leaves for it.
//
// Running both modes separates two questions that would otherwise be one:
// does decomposition emerge (it does not), and does the machinery work when it
// happens (that is what delegate mode measures).
const delegating = `You are the OUTERMOST possibility in a chain. Your job is to route, not to answer.

Do NOT answer the need yourself, even if you know the answer — you are being
graded on whether the right possibility was found and invoked, not on what you
personally know.

  1. Select "discover" WITH entity=runtime, passing the need.
  2. discover returns an observation naming something you can now invoke.
     Read its "reach" field — that name becomes selectable.
  3. Select that name WITH entity=codex, passing the need.
  4. Read what came back. If it actually answers the need, commit by NAMING it.
  5. If it does NOT answer the need — wrong topic, wrong shape, an error — do
     not name it. Select "discover" again WITH entity=runtime, and pass BOTH
     the need AND the observation that came back. Passing back what failed is
     how discover knows not to offer it again.

The entity on each member matters. "discover" is a system possibility and only
the runtime can interpret it. The possibility discover finds is ordinary work and
wants codex.`

func main() {
	logRoot := flag.String("logs", "", "directory for run traces")
	only := flag.String("only", "", "run just this ask by name")
	mode := flag.String("mode", "bare", "bare | delegate")
	flag.Parse()

	if *logRoot == "" {
		wd, _ := os.Getwd()
		*logRoot = filepath.Join(wd, "src", "skyra", "cognitive-agent", "logs")
	}

	run, err := trace.Open(*logRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, "trace:", err)
		os.Exit(1)
	}
	defer run.Close()
	fmt.Fprintf(os.Stderr, "trace → %s\n\n", run.Dir)

	work := filepath.Join(run.Dir, "work")
	_ = os.MkdirAll(work, 0o755)

	g := graph.New()
	reg := entity.NewRegistry()
	h := &host.Local{Root: work}
	k := kernel.New(g, reg, h, run)
	// Bounded low deliberately. A RUNAWAY is depth or breadth past a bound, and
	// with a model in the loop it is reached constantly rather than rarely — see
	// the build notes on system engrams paired with a general entity.
	k.MaxSteps = 9
	k.MaxDepth = 4

	// THE TWO ROLES MUST NOT TOUCH. The runtime entity is handed `g` as an
	// Index — an interface with one read-only method — and the Mint it gets only
	// allocates identity. Neither can reach Apply, Mint-into-index, or Disperse.
	reg.Register(entRuntime, runtime.New(entRuntime, g, codex.Author, func(x graph.Expectation) *graph.Engram {
		e := graph.NewEngram(k.IDs.Engram(), x)
		k.Hold(e) // NOT in the index. mint-on-closure will put it there if it closes.
		return e
	}, run))
	reg.Register(entCodex, codex.New(entCodex, run))

	seedSystemSpace(k, g)

	var results []result
	for _, a := range asks {
		if *only != "" && *only != a.Name {
			continue
		}
		ask := a.Ask
		if *mode == "delegate" {
			ask = a.Ask + "\n\n" + delegating
		}
		results = append(results, runOne(k, g, h, run, a.Name, ask, a.Need))
	}

	report(k, g, run, results)
}

// seedSystemSpace installs discover and its resolvers.
//
// discover is nameable in EVERY frame — a builtin name, not something you must
// discover in order to discover. That is what stops the regress. Its resolvers
// are its own children, so discover's history records WHICH RESOLVER SUITED
// WHICH KIND OF NEED, accumulating like anything else.
func seedSystemSpace(k *kernel.Kernel, g *graph.Graph) {
	mk := func(expr string) *graph.Engram {
		e := graph.NewEngram(k.IDs.Engram(), graph.Expectation{Expression: expr})
		g.Mint(e)
		return e
	}
	discover := mk(runtime.Discover)
	// The resolvers are children of discover from the start rather than formed on
	// citation, because at n=0 there is one seed discover and its strategy is
	// whatever was hand-written. Soft only at cold start.
	discover.Engrams["by-index"] = mk(runtime.ByIndex)
	discover.Engrams["by-recency"] = mk(runtime.ByRecency)
	discover.Engrams["create"] = mk(runtime.Create)

	k.Builtin["discover"] = discover
}

type result struct {
	Name     string
	Ask      string
	OK       bool
	Answer   string
	Err      string
	Accepted bool
}

func runOne(k *kernel.Kernel, g *graph.Graph, h host.Host, run *trace.Run, name, ask, need string) result {
	run.Logf("")
	run.Logf("════════════════════════════════════════════════════════════════")
	run.Logf("ASK %q", name)
	run.Logf("    %s", need)
	run.Logf("════════════════════════════════════════════════════════════════")

	// INGRESS. Producer is empty — no possibility produced this, it entered the
	// tree from outside. Nothing gates it (questions.md §5).
	in := k.Ingress(content.Content{"need": content.Str(need)})

	root, out, err := k.Ask(ask, entCodex, []*graph.Observation{in})
	if err != nil {
		run.Logf("ASK %q FAILED: %v", name, err)
		k.Reject(root, err.Error())
		return result{Name: name, Ask: ask, Err: err.Error()}
	}

	// ACCEPTANCE IS CITATION, and it is COMPUTED, NOT ASKED.
	//
	// FLAGGED, because this is questions.md §4 in the harness itself: the only
	// instrument available here is the chain's own output, so "accepted" means
	// "the chain closed and did not report an error." That is precisely the
	// terminus case — short-range positive and long-range positive, read off ONE
	// instrument. A real grounding needs a second instrument, and this run does
	// not have one.
	accepted := !strings.Contains(strings.ToLower(out.Content.Render()), "error")
	if accepted {
		if err := k.Accept(root, out); err != nil {
			run.Logf("materialization failed: %v", err)
		}
	} else {
		k.Reject(root, "the chain reported an error")
	}

	return result{
		Name: name, Ask: ask, OK: true,
		Answer:   oneLine(out.Content.Render()),
		Accepted: accepted,
	}
}

func report(k *kernel.Kernel, g *graph.Graph, run *trace.Run, rs []result) {
	var b strings.Builder

	b.WriteString("RESULTS\n=======\n\n")
	for _, r := range rs {
		status := "✓"
		if !r.OK {
			status = "✗"
		}
		fmt.Fprintf(&b, "%s  %-16s %s\n", status, r.Name, trunc(oneLine(r.Ask), 60))
		if r.Err != "" {
			fmt.Fprintf(&b, "     failed: %s\n", r.Err)
		} else {
			fmt.Fprintf(&b, "     accepted=%v  %s\n", r.Accepted, trunc(r.Answer, 160))
		}
		b.WriteString("\n")
	}

	b.WriteString("\nTHE GRAPH\n=========\n\n")
	for _, e := range g.Engrams() {
		if len(e.History) == 0 && len(e.Engrams) == 0 {
			continue
		}
		fmt.Fprintf(&b, "%s  %s\n", e.ID, trunc(oneLine(e.Expectation.Expression), 90))
		cited := 0
		hops := map[int]int{}
		for _, inv := range e.History {
			if inv.Cited {
				cited++
			}
			for hp, n := range inv.HopProfile() {
				hops[hp] += n
			}
		}
		if len(e.History) > 0 {
			fmt.Fprintf(&b, "     invoked %d · cited %d · hops %v\n", len(e.History), cited, hops)
			for _, p := range e.Partitions(g.Deref()) {
				fmt.Fprintf(&b, "     · %s ×%d (used %d)\n", p.Shape, len(p.Invs), p.Used())
			}
		}
		if len(e.Engrams) > 0 {
			var kids []string
			for n := range e.Engrams {
				kids = append(kids, n)
			}
			fmt.Fprintf(&b, "     children: %s\n", strings.Join(kids, ", "))
		}
		b.WriteString("\n")
	}

	if ds := g.Discards(); len(ds) > 0 {
		b.WriteString("\nDISCARDED EFFECTS\n=================\n\n")
		for _, d := range ds {
			fmt.Fprintf(&b, "  %s  %s\n", d.Engram, d.Detail)
		}
	}

	user, kern := k.Spiral()
	fmt.Fprintf(&b, "\nROOTS: %d user-rooted, %d kernel-rooted\n", user, kern)

	run.Write("report.txt", b.String())
	fmt.Print("\n" + b.String())
}

func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
