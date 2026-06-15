package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skyra-v1/src/inference"
)

const system = `You are the global coherence check of a cognitive runtime.

You are given a SETPOINT (a state the system is trying to reach) and a RESPONSE
(what just happened), already paired as the headline comparison. Below them is
numbered Context — cells pulled from past episodes. The VAST MAJORITY is
irrelevant noise; only a handful of cells bear on the setpoint.

Your job has two parts.

1) Headline. Fill the <alignment> for the SETPOINT vs RESPONSE comparison.
   <alignment> is a graded LEVEL from -1 to 1. USE THE WHOLE RANGE:
     +1     = the setpoint is fully met (done)
     0 to 1 = partly met / in progress (one part of a two-part setpoint ~= 0.5)
      0     = unrelated
     -1 to 0 = the response moved the world AWAY from the setpoint
   "Not yet complete" is NOT negative.

2) Context management. Group ONLY the few cells that actually bear on the setpoint.
   Ignore the noise entirely — do not group it. For each relevant bundle:
   <context-management>
     <cells>which numbered cells</cells>
     <observation>what this grouping means</observation>
     <alignment>+1 to -1 — did these cells help or hurt the setpoint</alignment>
   </context-management>
   Only cite cells from the Context list.

Output only the tags. Put a one-line reason right after the headline alignment.`

const setpoint = "the user has been greeted and has replied back"

var signalA = []string{
	`last week I greeted michael; he answered "hey — good to see you"   (episode: first call after the gap)`,
	"a greeting lands only when the other answers; an unanswered hello is half-made   (understanding)",
	`a greeting is a short opener — "hi", "hello"   (general knowledge)`,
	"two days ago I said hello and got silence   (episode: the silent morning)",
	"this is michael; we've talked many times   (identity)",
}

var signalB = append(append([]string{}, signalA...),
	`a moment ago I greeted them: "Hello! How are you today?"   (episode: this exchange)`)

var michaelDecoys = []string{
	"michael prefers his coffee black, no sugar",
	"michael's birthday is in march",
	"michael drives a blue sedan",
	"michael uses a standing desk",
	"michael's flight lands at six",
	"michael switched to a mechanical keyboard",
	"michael's desk is by the window",
	"michael takes the early train",
}

var (
	systems = []string{"staging deploy", "prod build", "ci pipeline", "database migration", "api gateway", "auth service", "cache layer", "search index", "billing job", "backup job", "cron worker", "load balancer", "cdn edge", "message queue", "webhook handler"}
	events  = []string{"failed", "passed", "restarted", "timed out", "scaled up", "rolled back", "got flagged", "recovered", "stalled", "redeployed"}
	times   = []string{"tuesday", "this morning", "overnight", "at 4am", "after lunch", "last friday", "during standup", "on the third try", "again", "twice today"}
	people  = []string{"sarah", "dave", "priya", "tom", "lena", "marco", "ana", "kev", "raj", "nina"}
	saidv   = []string{"mentioned", "noted", "flagged", "asked about", "complained that", "reminded us that"}
	topics  = []string{"the design review", "the budget meeting", "the vendor demo", "the offsite", "the retro", "the launch date", "the contractor", "the new hire", "the roadmap", "the audit"}
	moves   = []string{"got pushed to", "moved to", "is set for", "slipped to", "got rescheduled to"}
	days    = []string{"monday", "tuesday", "wednesday", "thursday", "friday", "next week", "the 9th", "end of month"}
	objects = []string{"the printer", "the espresso machine", "the projector", "the standing desk", "the whiteboard", "the door badge", "the parking pass", "the wifi", "the elevator", "the thermostat"}
	states  = []string{"is jammed again", "needs descaling", "won't connect", "is out of paper", "needs a reset", "is acting up", "got replaced", "is on order"}
	items   = []string{"eggs", "milk", "bread", "spinach", "coffee", "rice", "apples", "yogurt", "pasta", "lemons"}
)

func pick(pool []string, n int) string {
	if n < 0 {
		n = -n
	}
	return pool[n%len(pool)]
}

func genCell(i int) string {
	switch i % 6 {
	case 0:
		return fmt.Sprintf("the %s %s %s", pick(systems, i), pick(events, i/2), pick(times, i/3))
	case 1:
		return fmt.Sprintf("%s %s %s %s %s", pick(people, i), pick(saidv, i/2), pick(topics, i/3), pick(moves, i/5), pick(days, i/7))
	case 2:
		return fmt.Sprintf("%s %s", pick(objects, i), pick(states, i/2))
	case 3:
		return fmt.Sprintf("groceries: %s, %s, %s", pick(items, i), pick(items, i/2+1), pick(items, i/3+2))
	case 4:
		return fmt.Sprintf("the %s %s %s and %s %s %s", pick(systems, i), pick(events, i/2), pick(times, i/3), pick(systems, i/5+1), pick(events, i/7+1), pick(times, i/11+1))
	default:
		return fmt.Sprintf("%s is out %s; %s covers %s", pick(people, i), pick(days, i/2), pick(people, i/3+1), pick(topics, i/5))
	}
}

func insert(s []string, pos int, v string) []string {
	if pos < 0 {
		pos = 0
	}
	if pos > len(s) {
		pos = len(s)
	}
	s = append(s, "")
	copy(s[pos+1:], s[pos:])
	s[pos] = v
	return s
}

func signalFracs(n int) []float64 {
	if n == 6 {
		return []float64{0.04, 0.23, 0.42, 0.6, 0.8, 0.97}
	}
	return []float64{0.04, 0.27, 0.5, 0.73, 0.96}
}

// build generates noise to a token budget, scatters decoys then signal across
// the full span, and returns the numbered context block plus the index sets.
func build(targetTokens int, signal []string) (block string, sigIdx, decoyIdx []int, nCells int) {
	targetChars := targetTokens * 4
	var cells []string
	chars, i := 0, 0
	for chars < targetChars {
		c := genCell(i)
		cells = append(cells, c)
		chars += len(c) + 8
		i++
	}
	for k := 0; k < len(michaelDecoys); k++ {
		pos := (len(cells) * (k + 1)) / (len(michaelDecoys) + 1)
		cells = insert(cells, pos, michaelDecoys[k])
	}
	fracs := signalFracs(len(signal))
	for k, s := range signal {
		pos := int(float64(len(cells)) * fracs[k])
		cells = insert(cells, pos, s)
	}

	sigSet := map[string]bool{}
	for _, s := range signal {
		sigSet[s] = true
	}
	decoySet := map[string]bool{}
	for _, d := range michaelDecoys {
		decoySet[d] = true
	}
	var b strings.Builder
	b.WriteString("Context (each from a past episode):\n")
	for idx, c := range cells {
		num := idx + 1
		fmt.Fprintf(&b, "  #%d  %s\n", num, c)
		if sigSet[c] {
			sigIdx = append(sigIdx, num)
		} else if decoySet[c] {
			decoyIdx = append(decoyIdx, num)
		}
	}
	return b.String(), sigIdx, decoyIdx, len(cells)
}

func present(response, contextBlock string) string {
	return fmt.Sprintf(`SETPOINT: %s
RESPONSE: %s

<comparison>
  <expected>setpoint</expected>
  <actual>response</actual>
  <alignment> ? </alignment>
</comparison>

%s
Then group ONLY the relevant context into <context-management> blocks and score each against the setpoint. Ignore the noise.`, setpoint, response, contextBlock)
}

func model() string {
	if m := os.Getenv("DEEPSEEK_MODEL"); m != "" {
		return m
	}
	return "deepseek-v4-pro"
}

type run struct {
	id          string
	tokens      int
	response    string
	signalCells []string
}

func execute(base string, r run) error {
	dir := filepath.Join(base, r.id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	block, sigIdx, decoyIdx, nCells := build(r.tokens, r.signalCells)
	p := present(r.response, block)
	approxTokens := len(p) / 4

	meta := fmt.Sprintf("# %s\nprovider: deepseek\nmodel: %s\ntarget tokens: %d\napprox present tokens: %d\ntotal cells: %d\nsignal cells (should be found): %v\nentity-decoy cells (should be ignored): %v\n",
		r.id, model(), r.tokens, approxTokens, nCells, sigIdx, decoyIdx)
	if err := os.WriteFile(filepath.Join(dir, "call-meta.txt"), []byte(meta+"\n===== SYSTEM =====\n"+system+"\n"), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "present.txt"), []byte(p), 0o644); err != nil {
		return err
	}

	fmt.Printf("\n========== %s ==========\n", r.id)
	fmt.Printf("cells: %d   approx tokens: %d   signal at %v   decoys at %v\n\n", nCells, approxTokens, sigIdx, decoyIdx)
	out, err := inference.CallDeepSeek(system, p)
	if err != nil {
		_ = os.WriteFile(filepath.Join(dir, "response.txt"), []byte("ERROR: "+err.Error()+"\n"), 0o644)
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "response.txt"), []byte(out+"\n"), 0o644); err != nil {
		return err
	}
	fmt.Println(out)
	fmt.Printf("\n[saved → %s]\n", dir)
	return nil
}

func main() {
	base := "experiment-out"
	if len(os.Args) > 1 {
		base = os.Args[1]
	}
	respA := `"Hello! How are you today?"   (my greeting, going out now)`
	respB := `the user replied: "good thanks — you?"`

	runs := []run{
		{"a-25k", 25000, respA, signalA},
		{"a-100k", 100000, respA, signalA},
		{"a-200k", 200000, respA, signalA},
		{"b-200k", 200000, respB, signalB},
	}
	for _, r := range runs {
		if err := execute(base, r); err != nil {
			fmt.Printf("ERROR on %s: %v\n", r.id, err)
			os.Exit(1)
		}
	}
}
