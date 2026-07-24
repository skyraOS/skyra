// Command dnav3 runs the v3 being against a real model and a real filesystem.
//
// The point is TWO NUMBERS:
//
//	passes-to-closure  ↓    retrieval is working — planning is getting cheaper
//	prediction error   ↓    the world-model is working — the bets are getting better
//
// Both down is the thesis. Either alone is a warning, and either failure LOCALIZES.
//
// Usage:
//
//	go run ./cmd/dnav3 -reps 5           the same task, five times. does planning get cheaper?
//	go run ./cmd/dnav3 -reps 5 -v         and show every emission.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"skyra-v1/src/inference"
	dnav3 "skyra-v1/src/skyra/cognitive-agent/skyra-dna-v3"
)

func main() {
	reps := flag.Int("reps", 3, "how many times to run the task")
	verbose := flag.Bool("v", false, "print the present and every emission")
	flag.Parse()

	calls := 0
	infer := func(system, present string) (string, error) {
		calls++
		if *verbose {
			fmt.Printf("\n─── PRESENT (call %d) ───\n%s\n", calls, present)
		}
		out, err := inference.CallCodex(system, present)
		if *verbose && err == nil {
			fmt.Printf("─── EMISSION ───\n%s\n", strings.TrimSpace(out))
		}
		return out, err
	}

	freshWorld := func() *dnav3.World {
		root, err := os.MkdirTemp("", "skyra-world-*")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return dnav3.NewWorld(root)
	}
	being := dnav3.NewBeing(freshWorld(), infer)

	// The task. A build with a HIDDEN PREREQUISITE.
	//
	// The target asks only that the build succeed: `build:/ status ok`. But the world's build FAILS
	// unless BOTH a source file AND a config file exist. The config is the thing a naive first
	// attempt forgets — cold, the model tends to emit "write source; build", and the build fails or
	// the plan leaves `file:/app.conf exists true` ungrounded, forcing another pass. Warm, precedent
	// shows the complete plan — write source, write config, THEN build — and it closes in one pass.
	//
	// That is the shape that makes passes-to-closure fall: a prerequisite precedent remembers and a
	// cold attempt does not.
	target := []dnav3.Triple{
		{Subject: "build:/", Predicate: "status", Object: "ok"},
	}

	fmt.Println("TARGET:")
	for _, t := range target {
		fmt.Printf("  %s %s %s\n", t.Subject, t.Predicate, t.Object)
	}
	if being.Aff.Standing(target) {
		fmt.Println("  (STANDING — no exact channel. a compass, not a bet.)")
	}
	fmt.Println()

	for i := 1; i <= *reps; i++ {
		being.NewEpisode(freshWorld()) // fresh WORLD, same MEMORY — the whole point of the experiment
		being.ObserveImpulse([]dnav3.Triple{{
			Subject: "being:operator", Predicate: "requests", Object: "build:/",
		}}, target)
		before := being.Metrics
		c, err := being.Run(target, 6)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run %d: %v\n", i, err)
			break
		}
		resolved := being.Metrics.Resolved - before.Resolved
		correct := being.Metrics.Correct - before.Correct
		edge := 0.0
		if resolved > 0 {
			edge = (being.Metrics.GradeSum - before.GradeSum) / float64(resolved)
		}
		status := "closed"
		if c.Unresolved {
			status = "UNRESOLVED (nothing emitted to the world)"
		}
		fmt.Printf("run %d  passes=%d  acts=%d  bets=%d correct=%d  edge=%+.3f  rejects=%d  %s\n",
			i, c.Passes, c.Acts, resolved, correct, edge, len(c.Rejects), status)
		for _, r := range c.Rejects {
			fmt.Printf("        reject: %s — %s\n", trunc(r.Line, 60), r.Reason)
		}
	}

	fmt.Printf("\n%s\n", being.Report())

	fmt.Println("\nEDGES LEARNED  (P = counted P(world state | expression), N = trials):")
	for _, l := range being.EdgeReport() {
		fmt.Printf("  %s\n", l)
	}
}

func trunc(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
