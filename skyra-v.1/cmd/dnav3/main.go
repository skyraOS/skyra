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
	dnav3 "skyra-v1/src/skyra/skyra-dna-v3"
)

func main() {
	reps := flag.Int("reps", 3, "how many times to run the task")
	verbose := flag.Bool("v", false, "print the present and every emission")
	flag.Parse()

	root, err := os.MkdirTemp("", "skyra-world-*")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer os.RemoveAll(root)
	fmt.Printf("world: %s\n\n", root)

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

	being := dnav3.NewBeing(dnav3.NewWorld(root), infer)

	// The task. Deliberately MULTI-STEP: the script must exist, then be RUN, and the run must produce
	// an output file with the right content in it. That is three dependent states — cold, the model
	// has to discover the ordering; warm, precedent should hand it over. If passes-to-closure never
	// falls on a task this shaped, retrieval is not working.
	target := []dnav3.Triple{
		{Subject: "file:/greet.sh", Predicate: "exists", Object: "true"},
		{Subject: "dir:/", Predicate: "status", Object: "ok"},
		{Subject: "file:/out.txt", Predicate: "contains:hello", Object: "true"},
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
