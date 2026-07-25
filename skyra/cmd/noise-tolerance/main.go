package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"skyra-v1/src/inference"
)

// Same graded comparator prompt as comparator-mock Run 2.
const system = `You are the global coherence check of a cognitive runtime.

You are given a SETPOINT (a state the system is trying to reach) and a RESPONSE
(what just happened), already paired as the headline comparison. Below them is
numbered Context — cells pulled from past episodes. Most of the context is
irrelevant noise; only some cells bear on the setpoint.

Your job has two parts.

1) Headline. Fill the <alignment> for the SETPOINT vs RESPONSE comparison.
   <alignment> is a graded LEVEL from -1 to 1. USE THE WHOLE RANGE, not just the ends:
     +1     = the setpoint is fully met (done)
     0 to 1 = partly met / in progress — the response moved TOWARD the setpoint
              but it is not complete yet. Score proportional to how much is done.
              (a two-part setpoint with one part satisfied is about 0.5)
      0     = unrelated — the response neither advanced nor harmed the setpoint
     -1 to 0 = the response moved the world AWAY from the setpoint
   Score NEGATIVE ONLY if the response actively worked against the setpoint.
   "Not yet complete" is NOT negative.

2) Context management. Group ONLY the context cells that actually bear on the
   setpoint into bundles. Ignore irrelevant noise — do not group it. For each bundle:
   <context-management>
     <cells>which numbered cells</cells>
     <observation>what this episode means</observation>
     <alignment>+1 to -1 — did these cells help or hurt the setpoint</alignment>
   </context-management>
   Only cite cells from the Context list. Make as many as the relevant context warrants.

Output only the tags. Put a one-line reason right after the headline alignment.`

var noise = []string{
	"the staging deploy failed tuesday on the build step",
	"coffee is low; need to restock the kitchen",
	"the quarterly report is due friday",
	"it rained most of the afternoon yesterday",
	"the api rate limit reset at midnight",
	"michael prefers his coffee black, no sugar",
	"the database migration ran clean on the third try",
	"the parking garage closes at 11pm on weekends",
	"sarah mentioned the design review got pushed to thursday",
	"the new keyboard arrived but the keys feel stiff",
	"the unit tests pass but the integration suite is flaky",
	"lunch options near the office are mostly sandwiches",
	"the wifi dropped twice during the standup",
	"michael's birthday is in march",
	"the invoice from the vendor is overdue",
	"the cat knocked a plant off the windowsill",
	"the flight to the conference is booked for the 9th",
	"the linter is complaining about unused imports",
	"the espresso machine needs descaling",
	"dave is out next week on vacation",
	"the spreadsheet formula broke after the column shift",
	"the gym is closed for renovations until june",
	"the prod logs show a spike in 500s overnight",
	"the contract renewal is up for negotiation",
	"the printer on the third floor is jammed again",
	"groceries: eggs, milk, bread, spinach",
	"the meeting room projector won't connect to the mac",
	"the stock ticker dipped at open",
	"the package from amazon is delayed in transit",
	"michael drives a blue sedan",
	"the css grid layout broke on mobile",
	"the dentist appointment is at 2pm thursday",
	"the server room ac is running loud",
	"the podcast episode ran two hours long",
	"the budget meeting got rescheduled twice",
	"the umbrella is still in the car",
	"the feature flag is half rolled out to prod",
	"the neighbor's dog barks every morning",
	"the credit card expires next month",
	"the backup job finished at 4am",
	"the conference badge needs to be printed",
	"the team lunch is set for friday noon",
	"the oauth token keeps expiring early",
	"the bookshelf needs a second bracket",
	"the standup moved to 9:30 starting monday",
	"the cpu usage flatlined after the restart",
	"the rental car return is due by 6",
	"the slack channel got archived by mistake",
	"the hr portal password reset is broken",
	"the plant on the desk needs water",
	"the demo environment is down for maintenance",
	"michael uses a standing desk",
	"the tax documents are due in april",
	"the elevator inspection is scheduled friday",
	"the json parser chokes on trailing commas",
}

var signalA = []string{
	`last week I greeted michael; he answered "hey — good to see you"   (episode: first call after the gap)`,
	"a greeting lands only when the other answers; an unanswered hello is half-made   (understanding)",
	`a greeting is a short opener — "hi", "hello"   (general knowledge)`,
	"two days ago I said hello and got silence   (episode: the silent morning)",
	"this is michael; we've talked many times   (identity)",
}

var signalB = append(append([]string{}, signalA...),
	`a moment ago I greeted them: "Hello! How are you today?"   (episode: this exchange)`)

type variant struct {
	id        string
	label     string
	response  string
	signal    []string
	positions []int // 1-based display positions for signal cells
}

const setpoint = "the user has been greeted and has replied back"

// build interleaves signal cells into noise at the given 1-based positions,
// returns the numbered context block, the signal indices, and the entity-decoy
// (noise cells mentioning michael) indices.
func build(total int, signal []string, positions []int) (block string, signalIdx, decoyIdx []int) {
	cells := make([]string, total)
	sig := map[int]string{}
	for i, p := range positions {
		sig[p-1] = signal[i]
	}
	ni := 0
	for i := 0; i < total; i++ {
		if s, ok := sig[i]; ok {
			cells[i] = s
			signalIdx = append(signalIdx, i+1)
		} else {
			cells[i] = noise[ni]
			ni++
		}
	}
	var b strings.Builder
	b.WriteString("Context (each from a past episode):\n")
	for i, c := range cells {
		num := i + 1
		fmt.Fprintf(&b, "  #%d  %s\n", num, c)
		isSig := false
		for _, s := range signalIdx {
			if s == num {
				isSig = true
			}
		}
		if !isSig && strings.Contains(strings.ToLower(c), "michael") {
			decoyIdx = append(decoyIdx, num)
		}
	}
	return b.String(), signalIdx, decoyIdx
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

func run(base string, v variant, total int) error {
	dir := filepath.Join(base, v.id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	block, sigIdx, decoyIdx := build(total, v.signal, v.positions)
	p := present(v.response, block)

	call := fmt.Sprintf("# inference call — %s\nprovider: deepseek\nmodel: %s\ntemperature: 0.2\ntotal cells: %d   signal: %d   noise: %d\n\n===== SYSTEM =====\n%s\n\n===== PRESENT =====\n%s\n",
		v.label, model(), total, len(sigIdx), total-len(sigIdx), system, p)
	if err := os.WriteFile(filepath.Join(dir, "call.txt"), []byte(call), 0o644); err != nil {
		return err
	}
	key := fmt.Sprintf("SIGNAL cells (the load-bearing ones — should be grouped): %v\nENTITY DECOY noise cells (mention 'michael' but irrelevant — should be ignored): %v\n", sigIdx, decoyIdx)
	if err := os.WriteFile(filepath.Join(dir, "signal-key.txt"), []byte(key), 0o644); err != nil {
		return err
	}

	fmt.Printf("\n========== %s ==========\n", v.label)
	fmt.Printf("signal at %v   entity-decoys at %v\n\n", sigIdx, decoyIdx)
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
	variants := []struct {
		v     variant
		total int
	}{
		{variant{"a-greeting-only", "VARIANT A — greeting, no reply (expect ~0.5, find 5 signal in noise)",
			`"Hello! How are you today?"   (my greeting, going out now)`, signalA, []int{8, 21, 34, 47, 56}}, 60},
		{variant{"b-reply-arrived", "VARIANT B — reply arrived (expect ~1.0, find 6 signal in noise)",
			`the user replied: "good thanks — you?"`, signalB, []int{8, 21, 34, 47, 56, 60}}, 61},
	}
	for _, vc := range variants {
		if err := run(base, vc.v, vc.total); err != nil {
			fmt.Printf("ERROR on %s: %v\n", vc.v.id, err)
			os.Exit(1)
		}
	}
}
