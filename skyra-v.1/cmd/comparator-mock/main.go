package main

import (
	"fmt"
	"os"
	"path/filepath"

	"skyra-v1/src/inference"
)

const system = `You are the global coherence check of a cognitive runtime.

You are given a SETPOINT (a state the system is trying to reach) and a RESPONSE
(what just happened), already paired as the headline comparison. Below them is
numbered Context — cells pulled from past episodes.

Your job has two parts.

1) Headline. Fill the <alignment> for the SETPOINT vs RESPONSE comparison.
   <alignment> is a graded LEVEL from -1 to 1. USE THE WHOLE RANGE, not just the ends:
     +1     = the setpoint is fully met (done)
     0 to 1 = partly met / in progress — the response moved TOWARD the setpoint
              but it is not complete yet. Score proportional to how much is done.
              (a two-part setpoint with one part satisfied is about 0.5)
      0     = unrelated — the response neither advanced nor harmed the setpoint
     -1 to 0 = the response moved the world AWAY from the setpoint — made it worse,
              undid progress, or is the opposite of what is wanted
   Score NEGATIVE ONLY if the response actively worked against the setpoint.
   "Not yet complete" is NOT negative. An unfinished-but-advancing setpoint is
   positive-and-below-1. Judge what is true right now: if the setpoint names a
   condition that has not happened yet (e.g. "and has replied back"), it is not
   FULLY met just because you acted — but the part you did counts as progress.

2) Context management. Group the context cells into bundles. For each bundle:
   <context-management>
     <cells>which numbered cells</cells>
     <observation>what this grouping means</observation>
     <alignment>+1 to -1 — did these cells help or hurt the setpoint. Use the whole
       range: strongly supportive near +1, mildly supportive around +0.3, neutral 0,
       negative only if the cells point AWAY from reaching the setpoint</alignment>
   </context-management>
   Only cite cells from the Context list. Make as many as the context warrants.

Output only the tags. Put a one-line reason right after the headline alignment.`

const presentA = `SETPOINT: the user has been greeted and has replied back
RESPONSE: "Hello! How are you today?"   (my greeting, going out now)

<comparison>
  <expected>setpoint</expected>
  <actual>response</actual>
  <alignment> ? </alignment>
</comparison>

Context (each from a past episode):
  #1  last week I greeted michael; he answered "hey — good to see you"   (episode: first call after the gap)
  #2  a greeting lands only when the other answers; an unanswered hello is half-made   (understanding)
  #3  a greeting is a short opener — "hi", "hello"   (general knowledge)
  #4  two days ago I said hello and got silence   (episode: the silent morning)
  #5  this is michael; we've talked many times   (identity)`

const presentB = `SETPOINT: the user has been greeted and has replied back
RESPONSE: the user replied: "good thanks — you?"

<comparison>
  <expected>setpoint</expected>
  <actual>response</actual>
  <alignment> ? </alignment>
</comparison>

Context (each from a past episode):
  #1  last week I greeted michael; he answered "hey — good to see you"   (episode: first call after the gap)
  #2  a greeting lands only when the other answers; an unanswered hello is half-made   (understanding)
  #3  a greeting is a short opener — "hi", "hello"   (general knowledge)
  #4  two days ago I said hello and got silence   (episode: the silent morning)
  #5  this is michael; we've talked many times   (identity)
  #6  a moment ago I greeted them: "Hello! How are you today?"   (episode: this exchange)`

type variant struct {
	id      string
	label   string
	present string
}

func model() string {
	if m := os.Getenv("DEEPSEEK_MODEL"); m != "" {
		return m
	}
	return "deepseek-v4-pro"
}

func callFile(label, present string) string {
	return fmt.Sprintf(`# inference call — %s
provider: deepseek
model: %s
temperature: 0.2

===== SYSTEM =====
%s

===== PRESENT =====
%s
`, label, model(), system, present)
}

func run(base string, v variant) error {
	dir := filepath.Join(base, v.id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "call.txt"), []byte(callFile(v.label, v.present)), 0o644); err != nil {
		return err
	}

	fmt.Printf("\n========== %s ==========\n", v.label)
	out, err := inference.CallDeepSeek(system, v.present)
	if err != nil {
		// still record the failure so the run is reproducible
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

	variants := []variant{
		{"a-greeting-only", "VARIANT A — greeting sent, no reply yet (expect partial)", presentA},
		{"b-reply-arrived", "VARIANT B — reply has arrived (expect near-complete)", presentB},
	}
	for _, v := range variants {
		if err := run(base, v); err != nil {
			fmt.Printf("ERROR on %s: %v\n", v.id, err)
			os.Exit(1)
		}
	}
}
