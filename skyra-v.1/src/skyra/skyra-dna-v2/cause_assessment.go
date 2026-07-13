package dnav2

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"sync"
)

const (
	CauseEfficient = "efficient"
	CauseFinal     = "final"
)

// CauseAssessment is one judgment of an active cause against one observed
// subgraph. The cause is the target; ObservationID identifies the Evidence
// input. A grade samples the position edge's signed EMA and the emitting
// lineage's query budget immediately; no-grade samples neither.
type CauseAssessment struct {
	ObservationID string
	CauseID       string
	EdgeID        string
	Role          string
	Grade         float64
	HasGrade      bool
}

type causeTarget struct {
	id     string
	edgeID string
	role   string
}

type causeGradeResult struct {
	grade    float64
	hasGrade bool
	valid    bool
}

var causeGradeRE = regexp.MustCompile(`(?s)^\s*<grade>\s*([-+]?(?:\d+(?:\.\d*)?|\.\d+))\s*</grade>\s*$`)
var causeNoGradeRE = regexp.MustCompile(`(?is)^\s*<no-grade\s*/>\s*$`)

// AssessActiveCauses snapshots the active Aitia set from the Slice, then grades
// every populated cause position independently. Calls share immutable inputs
// and run in parallel; malformed or failed calls produce no assessment. tr is
// the board run's content trace (nil = untraced).
func (d *DNA) AssessActiveCauses(observationID string, tr *RunTrace) []CauseAssessment {
	infer := d.Infer()
	observation := d.CausalNodes().Get(observationID)
	set, active := d.ActiveSet()
	if infer == nil || observation == nil || !active {
		return nil
	}
	var targets []causeTarget
	if edge := d.CausalNodes().Edge(set.EfficientEdgeID); edge != nil {
		targets = append(targets, causeTarget{id: edge.Record.FromID, edgeID: edge.ID(), role: CauseEfficient})
	}
	if edge := d.CausalNodes().Edge(set.FinalEdgeID); edge != nil {
		targets = append(targets, causeTarget{id: edge.Record.ToID, edgeID: edge.ID(), role: CauseFinal})
	}

	observationText := causalSentence(observation)
	results := make(chan CauseAssessment, len(targets))
	var wg sync.WaitGroup
	for _, target := range targets {
		target := target
		cause := d.CausalNodes().Get(target.id)
		if cause == nil {
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			stage := "assess:" + target.role
			present := fmt.Sprintf("ACTIVE CAUSE (%s):\n%s\n\nOBSERVATION:\n%s", target.role, causalSentence(cause), observationText)
			gradeResult := inferCauseGradeResult(tr.Wrap(infer, stage), tr, stage, present)
			if !gradeResult.valid {
				return
			}
			assessment := CauseAssessment{
				ObservationID: observationID, CauseID: target.id, EdgeID: target.edgeID,
				Role: target.role, Grade: gradeResult.grade, HasGrade: gradeResult.hasGrade,
			}
			if gradeResult.hasGrade {
				d.Signed().Grade(target.edgeID, gradeResult.grade, d.Clock())
				_, _ = d.QueryBudgets().ApplyGrade(target.role, target.edgeID, gradeResult.grade)
			} else {
				_, _ = d.QueryBudgets().NoGrade(target.role, target.edgeID)
			}
			_ = d.CausalNodes().AttachEvidence(target.edgeID, observationID)
			results <- assessment
		}()
	}
	wg.Wait()
	close(results)
	var out []CauseAssessment
	for result := range results {
		out = append(out, result)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Role < out[j].Role })
	d.CausalNodes().AddAssessments(observationID, out...)
	return out
}

func inferCauseGrade(infer InferFunc, present string) (float64, bool) {
	result := inferCauseGradeResult(infer, nil, "assess", present)
	return result.grade, result.valid && result.hasGrade
}

func inferCauseGradeResult(infer InferFunc, tr *RunTrace, stage, present string) causeGradeResult {
	raw, err := infer(causeAssessmentSystem, present)
	if err != nil {
		return causeGradeResult{}
	}
	grade, hasGrade, valid := parseCauseGrade(raw)
	for attempt := 1; !valid && attempt <= parseMaxRetries; attempt++ {
		tr.Reject(stage, raw, fmt.Errorf("not <grade>N</grade> or <no-grade/>"))
		correction := fmt.Sprintf("%s\n\nYour previous response %q was invalid. Return only <grade>N</grade> with N between -1 and 1, or <no-grade/> when the observation cannot resolve the belief.", present, raw)
		raw, err = infer(causeAssessmentSystem, correction)
		if err != nil {
			return causeGradeResult{}
		}
		grade, hasGrade, valid = parseCauseGrade(raw)
	}
	if !valid {
		tr.Reject(stage, raw, fmt.Errorf("exhausted correction retries"))
	}
	return causeGradeResult{grade: grade, hasGrade: hasGrade, valid: valid}
}

func causalSentence(cell interface{ Content() any }) string {
	content, _ := cell.Content().(CausalNodeContent)
	return content.Sentence
}

func parseCauseGrade(raw string) (grade float64, hasGrade, valid bool) {
	if causeNoGradeRE.MatchString(raw) {
		return 0, false, true
	}
	match := causeGradeRE.FindStringSubmatch(raw)
	if match == nil {
		return 0, false, false
	}
	grade, err := strconv.ParseFloat(match[1], 64)
	if err != nil || grade < -1 || grade > 1 {
		return 0, false, false
	}
	return grade, true, true
}

const causeAssessmentSystem = "You grade one ACTIVE CAUSE against one observed causal subgraph. " +
	"The active cause is the object being graded; the observation is evidence. Judge whether the " +
	"observation puts the cause in resonance or tension. Return only <grade>N</grade>, where N is " +
	"between -1 and 1. Grade 0 is a real cancellation/no-charge judgment. If the observation is " +
	"absent, irrelevant, ambiguous, or insufficient to resolve the belief, return <no-grade/>."
