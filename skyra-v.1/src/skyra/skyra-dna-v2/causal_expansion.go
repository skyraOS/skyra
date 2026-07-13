package dnav2

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

type causalExpansionResult struct {
	role   string
	parsed ParsedSubgraph
	err    error
}

// expandCausalBeliefs runs Efficient and Final from the same immutable present
// snapshot. Neither result is persisted unless both validate. tr is the board
// run's content trace (nil = untraced).
func expandCausalBeliefs(d *DNA, tr *RunTrace, presentID string) (efficient, final ParsedSubgraph, err error) {
	infer := d.Infer()
	if infer == nil {
		return ParsedSubgraph{}, ParsedSubgraph{}, fmt.Errorf("causal expansion requires inference")
	}
	offered := structuralCandidateIDs(d.CausalNodes(), presentID, presentID)
	results := make(chan causalExpansionResult, 2)
	var wg sync.WaitGroup
	for _, role := range []string{CauseEfficient, CauseFinal} {
		role := role
		wg.Add(1)
		go func() {
			defer wg.Done()
			parsed, err := queryCausalBelief(d, tr, infer, role, presentID, offered)
			results <- causalExpansionResult{role: role, parsed: parsed, err: err}
		}()
	}
	wg.Wait()
	close(results)
	for result := range results {
		if result.err != nil && err == nil {
			err = fmt.Errorf("%s worker: %w", result.role, result.err)
		}
		switch result.role {
		case CauseEfficient:
			efficient = result.parsed
		case CauseFinal:
			final = result.parsed
		}
	}
	return efficient, final, err
}

func queryCausalBelief(d *DNA, tr *RunTrace, infer InferFunc, role, presentID string, offered map[string]float64) (ParsedSubgraph, error) {
	stage := "expand:" + role
	traced := tr.Wrap(infer, stage)
	// The lineage's CURRENT earned budget: the voice's specificity allowance,
	// carried across sets (edges die with their contexts; the lineage doesn't).
	budget, _ := d.QueryBudgets().Current(role)
	system := causalBeliefSystem(role, budget)
	present := causalBeliefPresent(d, role, presentID, offered)
	raw, err := traced(system, present)
	if err != nil {
		return ParsedSubgraph{}, err
	}
	parsed, perr := validateCausalBelief(d.CausalNodes(), raw, offered, budget)
	for attempt := 1; perr != nil && attempt <= parseMaxRetries; attempt++ {
		tr.Reject(stage, raw, perr)
		raw, err = traced(system, present+"\n\n---\n"+CausalBeliefCorrection(raw, perr))
		if err != nil {
			return ParsedSubgraph{}, err
		}
		parsed, perr = validateCausalBelief(d.CausalNodes(), raw, offered, budget)
	}
	if perr != nil {
		tr.Reject(stage, raw, perr)
	}
	return parsed, perr
}

// emissionTokens is the v0 token proxy — whitespace-delimited fields of the
// raw annotated emission — until a real tokenizer rides the inference seam.
func emissionTokens(raw string) int { return len(strings.Fields(raw)) }

func validateCausalBelief(ci *CausalNodes, raw string, offered map[string]float64, budget float64) (ParsedSubgraph, error) {
	// Budget is verified, not guillotined: an over-budget emission is a
	// contract failure returned for bounded correction, so a small budget
	// produces a coarser complete claim rather than a truncated sentence.
	if spent := emissionTokens(raw); float64(spent) > budget {
		return ParsedSubgraph{}, fmt.Errorf("emission spends %d tokens of a %.0f-token budget; re-emit the same belief on coarser ground", spent, budget)
	}
	parsed, err := ParseAnnotated(raw)
	if err != nil {
		return ParsedSubgraph{}, err
	}
	for _, node := range parsed.Nodes {
		if node.StableID == "" || ci.Get(node.StableID) == nil {
			continue
		}
		if _, ok := offered[node.StableID]; !ok {
			return ParsedSubgraph{}, fmt.Errorf("CausalNode %q was not offered to this worker", node.StableID)
		}
	}
	return parsed, nil
}

func causalBeliefPresent(d *DNA, role, presentID string, offered map[string]float64) string {
	var b strings.Builder
	b.WriteString("ACTIVE MATERIAL/FORMAL:\n")
	b.WriteString(renderCausalStructure(d.CausalNodes(), presentID))
	b.WriteString("\nROLE TO EMIT: ")
	b.WriteString(role)
	b.WriteString("\n\nOFFERED CAUSAL REFERENTS:\n")
	ids := make([]string, 0, len(offered))
	for id := range offered {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		b.WriteString(fmt.Sprintf("[candidate=%s amplitude=%.4f] %s\n", id, offered[id], causalSentence(d.CausalNodes().Get(id))))
	}
	return b.String()
}

func causalBeliefSystem(role string, budget float64) string {
	direction := "the inferred prior situation that brought the active present into being"
	if role == CauseFinal {
		direction = "the intended or oriented future situation toward which the active present points"
	}
	return "You emit the " + role + " causal belief for an active Material/Formal subgraph: " + direction + ". " +
		"Emit exactly one concise COMPLETE annotated sentence and nothing else. The sentence must parse into a nonempty " +
		"semantic subgraph; summary-only output is forbidden. Stand on the most specific ground firmly warranted. Offered " +
		"CausalNode IDs may be referenced, but reuse is optional and must reflect meaning rather than compression. " +
		fmt.Sprintf("Your query token budget is %.0f words. A complete claim on coarser ground always beats a truncated specific one.", budget)
}

func CausalBeliefCorrection(raw string, err error) string {
	return fmt.Sprintf("Your previous causal belief was:\n%s\n\nIt was invalid: %v. Re-emit the same belief as exactly one complete annotated sentence. Every referent and relationship must be represented; use only offered CausalNode IDs.", raw, err)
}
