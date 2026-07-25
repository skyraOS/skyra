package dnav3

import (
	"strings"
	"testing"
)

// The model must not know how it is scored.
//
// This test exists because the obvious "improvement" — explaining the scoring rule to the model — is
// a mistake that will be made again, probably by whoever reads plannerSystem next and thinks it is
// under-specified.
//
// Telling an LLM "you are paid for beating the base rate" does not make it calibrated. It makes it
// ROLEPLAY a score-maximizer: it deviates from history for its own sake, which is contrarianism, not
// information. And it destroys Expression.Score as a diagnostic, because a model that knows it is
// being tested on its edge over the base rate is no longer being tested on anything.
//
// A calibrated model beats the base rate automatically whenever it knows something. Asking for it is
// how you stop getting it.
func TestPromptDoesNotLeakTheScoringRule(t *testing.T) {
	banned := []string{
		"scored", "score", "punish", "reward", "earn", "paid", "pays",
		"base rate", "beat", "penal", "credit", "grade",
	}
	lower := strings.ToLower(plannerSystem)
	for _, w := range banned {
		if strings.Contains(lower, w) {
			t.Errorf("plannerSystem leaks the scoring rule (%q).\n\n"+
				"The model must not know what its confidence is USED for. Describe the JOB and the\n"+
				"FORMAT; say nothing about consequences. See the comment on plannerSystem.", w)
		}
	}
}
