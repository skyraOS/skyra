package dnav3

import (
	"math"
	"sort"
)

// DensityPredicate is the first deliberately narrow pocket-name vocabulary: one
// exact raw observation triple. Richer sensors may replace it later, but only after
// the density geometry has independently earned the pocket.
type DensityPredicate struct {
	Triple        Triple
	Coverage      float64
	BaseRate      float64
	Diagnosticity float64
}

type DensityPocket struct {
	Predicate   DensityPredicate
	Occurrences int
	Cohesion    float64
}

type spanObservationLookup func(SpanID) []Observation
type settlementSpanContains func(string, Truth, SpanID) bool

type pocketSpan struct {
	features map[string]Triple
}

// differentiateSource performs Phase 3 in the required order: geometry first,
// predicate fitting second, current-span selection last. It does not write a
// category observation or grade interpretation; those are Phase 4 concerns.
func differentiateSource(
	group TargetGroup,
	settled settlementSpanContains,
	observations spanObservationLookup,
	active []SpanID,
	current SpanID,
	cfg DifferentiationConfig,
) TargetGroup {
	group.PocketMatch = -1
	group.SelectedByPocket = false
	if group.State == SourceUnresolved || len(group.Branches) < 2 || observations == nil {
		return group
	}
	if cfg.SplitValleyDepth <= 0 {
		cfg.SplitValleyDepth = DefaultActivationConfig().Differentiation.SplitValleyDepth
	}

	excluded := completionKeys(group)
	branchSpans := make([][]pocketSpan, len(group.Branches))
	minOccurrences := math.MaxInt
	for i, branch := range group.Branches {
		ids := branchOccurrenceSpans(branch, settled, active, current)
		if len(ids) < 2 {
			return group // no within-pocket pair: recurrence has not formed geometry yet
		}
		if len(ids) < minOccurrences {
			minOccurrences = len(ids)
		}
		for _, id := range ids {
			branchSpans[i] = append(branchSpans[i], pocketSpan{
				features: rawSpanFeatures(observations(id), excluded),
			})
		}
	}

	within := make([]float64, len(branchSpans))
	minWithin := math.Inf(1)
	for i, spans := range branchSpans {
		within[i] = meanWithinSimilarity(spans)
		if within[i] < minWithin {
			minWithin = within[i]
		}
	}
	maxBetween := 0.0
	for i := range branchSpans {
		for j := 0; j < i; j++ {
			if similarity := meanBetweenSimilarity(branchSpans[i], branchSpans[j]); similarity > maxBetween {
				maxBetween = similarity
			}
		}
	}
	group.RawValley = math.Max(0, minWithin-maxBetween)
	maturity := float64(minOccurrences) / (float64(minOccurrences) + confHalf)
	group.ValleyDepth = group.RawValley * maturity
	if group.ValleyDepth < cfg.SplitValleyDepth {
		return group
	}

	// The valley exists without a predicate. Only now fit a region-relative
	// diagnostic exact predicate to each already-discovered pocket.
	regionCounts, regionN := featureCounts(branchSpans)
	seenNames := map[string]bool{}
	for i, spans := range branchSpans {
		predicate, ok := fitPocketPredicate(spans, regionCounts, regionN)
		if !ok || seenNames[predicate.Triple.Key()] {
			return clearPockets(group)
		}
		seenNames[predicate.Triple.Key()] = true
		group.Branches[i].Pocket = &DensityPocket{
			Predicate: predicate, Occurrences: len(spans), Cohesion: within[i],
		}
	}
	group.Differentiated = true

	// A unique current match is retained as provisional descriptive material only.
	// Predicate selection and diagnosticity currently reuse the same historical spans;
	// the held-out current impulse proves only that the selected feature fires, not that
	// it predicts branch membership. Until a disjoint evaluation + sparse promotion gate
	// lands, a pocket match must neither create, override, nor veto an operational commit.
	if current != "" {
		currentFeatures := rawSpanFeatures(observations(current), excluded)
		matched := -1
		for i, branch := range group.Branches {
			if branch.Pocket == nil {
				continue
			}
			if _, fires := currentFeatures[branch.Pocket.Predicate.Triple.Key()]; fires {
				if matched >= 0 {
					group.PocketMatch = -1
					return group // multiple provisional names carry no operational authority
				}
				matched = i
			}
		}
		if matched >= 0 {
			group.PocketMatch = matched
		}
	}
	return group
}

func clearPockets(group TargetGroup) TargetGroup {
	for i := range group.Branches {
		group.Branches[i].Pocket = nil
	}
	group.Differentiated = false
	return group
}

func completionKeys(group TargetGroup) map[string]bool {
	out := map[string]bool{}
	for _, branch := range group.Branches {
		for _, completion := range branch.Completions {
			out[completion.Triple.Key()] = true
		}
	}
	for _, completion := range group.Unresolved {
		out[completion.Triple.Key()] = true
	}
	return out
}

func branchOccurrenceSpans(branch Branch, settled settlementSpanContains, active []SpanID, exclude SpanID) []SpanID {
	if len(branch.Completions) == 0 || settled == nil {
		return nil
	}
	seen := map[SpanID]bool{}
	out := make([]SpanID, 0, len(active))
	for _, span := range active {
		if span == "" || span == exclude || seen[span] {
			continue
		}
		seen[span] = true
		complete := true
		for _, completion := range branch.Completions {
			if !settled(completion.EdgeKey, TruthTrue, span) {
				complete = false
				break
			}
		}
		if complete {
			out = append(out, span)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func rawSpanFeatures(observations []Observation, excluded map[string]bool) map[string]Triple {
	out := map[string]Triple{}
	for _, observation := range observations {
		if observation.Provenance != ProvenanceRaw || excluded[observation.Key()] {
			continue
		}
		out[observation.Key()] = observation.Triple
	}
	return out
}

func meanWithinSimilarity(spans []pocketSpan) float64 {
	var total float64
	pairs := 0
	for i := range spans {
		for j := 0; j < i; j++ {
			total += jaccardFeatures(spans[i].features, spans[j].features)
			pairs++
		}
	}
	if pairs == 0 {
		return 0
	}
	return total / float64(pairs)
}

func meanBetweenSimilarity(left, right []pocketSpan) float64 {
	var total float64
	pairs := 0
	for _, a := range left {
		for _, b := range right {
			total += jaccardFeatures(a.features, b.features)
			pairs++
		}
	}
	if pairs == 0 {
		return 0
	}
	return total / float64(pairs)
}

func jaccardFeatures(left, right map[string]Triple) float64 {
	if len(left) == 0 && len(right) == 0 {
		return 0 // shared absence is not a density observation
	}
	intersection := 0
	for key := range left {
		if _, ok := right[key]; ok {
			intersection++
		}
	}
	union := len(left) + len(right) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

func featureCounts(branches [][]pocketSpan) (map[string]int, int) {
	counts := map[string]int{}
	total := 0
	for _, spans := range branches {
		for _, span := range spans {
			total++
			for key := range span.features {
				counts[key]++
			}
		}
	}
	return counts, total
}

func fitPocketPredicate(spans []pocketSpan, regionCounts map[string]int, regionN int) (DensityPredicate, bool) {
	if len(spans) == 0 || regionN == 0 {
		return DensityPredicate{}, false
	}
	local := map[string]int{}
	triples := map[string]Triple{}
	for _, span := range spans {
		for key, triple := range span.features {
			local[key]++
			triples[key] = triple
		}
	}
	keys := make([]string, 0, len(local))
	for key := range local {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	best := DensityPredicate{}
	found := false
	for _, key := range keys {
		coverage := float64(local[key]) / float64(len(spans))
		baseRate := float64(regionCounts[key]) / float64(regionN)
		if coverage <= 0 || baseRate <= 0 {
			continue
		}
		score := coverage * math.Log(coverage/baseRate)
		if score <= 0 {
			continue
		}
		if !found || score > best.Diagnosticity {
			best = DensityPredicate{
				Triple: triples[key], Coverage: coverage, BaseRate: baseRate, Diagnosticity: score,
			}
			found = true
		}
	}
	return best, found
}
