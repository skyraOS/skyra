package dnav3

import (
	"fmt"
	"math"
)

// CalibrationCase binds candidate-predicate observations to an independently
// settled category dimension. The category label is a durable comparison
// reference, never a free-floating model-authored boolean.
type CalibrationCase struct {
	ID                    string
	ObservationIDs        []ObservationID
	CategoryPredictionID  PredictionID
	CategoryDimensionID   DimensionID
	CategorySettledSpanID SpanID
}

// CalibrationWindow is blind evaluation input chosen outside predicate grading.
// The support floors make sparse estimates explicitly provisional; meeting them
// only makes the arithmetic Ready, not the predicate trusted or promoted.
type CalibrationWindow struct {
	ID                        string
	Cases                     []CalibrationCase
	MinimumCategorySamples    int
	MinimumNonCategorySamples int
}

type resolvedCalibrationCase struct {
	observations  []Observation
	categoryRight bool
}

// CalibratePredictionField deterministically measures every dimension against a
// raw, pre-birth window whose category labels were settled on another prediction
// path. It freezes the result into the authored field before runtime settlement.
func (b *Being) CalibratePredictionField(id PredictionID, window CalibrationWindow) (*PredictionField, error) {
	if b == nil || b.Expr == nil || b.Obs == nil {
		return nil, fmt.Errorf("prediction calibration storage is unavailable")
	}
	field, ok := b.Expr.PredictionField(id)
	if !ok {
		return nil, fmt.Errorf("prediction field %q does not exist", id)
	}
	if window.ID == "" || len(window.Cases) == 0 {
		return nil, fmt.Errorf("calibration window is unnamed or empty")
	}
	if window.MinimumCategorySamples < 1 || window.MinimumNonCategorySamples < 1 {
		return nil, fmt.Errorf("calibration window requires positive category and non-category support floors")
	}
	selectionEvidence := make(map[ObservationID]bool, len(field.Attributions))
	for _, attribution := range field.Attributions {
		selectionEvidence[attribution.ObservationID] = true
	}
	seenCases := make(map[string]bool, len(window.Cases))
	resolved := make([]resolvedCalibrationCase, 0, len(window.Cases))
	for _, candidate := range window.Cases {
		if candidate.ID == "" || seenCases[candidate.ID] {
			return nil, fmt.Errorf("calibration case has a missing or duplicate ID %q", candidate.ID)
		}
		seenCases[candidate.ID] = true
		if candidate.CategoryPredictionID == id {
			return nil, fmt.Errorf("calibration case %q lets a prediction label itself", candidate.ID)
		}
		if len(candidate.ObservationIDs) == 0 {
			return nil, fmt.Errorf("calibration case %q has no observations", candidate.ID)
		}
		observations := make([]Observation, 0, len(candidate.ObservationIDs))
		seenObservations := map[ObservationID]bool{}
		for _, observationID := range candidate.ObservationIDs {
			if seenObservations[observationID] {
				continue
			}
			seenObservations[observationID] = true
			if selectionEvidence[observationID] {
				return nil, fmt.Errorf("calibration case %q overlaps declared predicate-selection evidence %q",
					candidate.ID, observationID)
			}
			observation, exists := b.Obs.Observation(observationID)
			if !exists || observation.Provenance != ProvenanceRaw {
				return nil, fmt.Errorf("calibration case %q observation %q is missing or non-raw",
					candidate.ID, observationID)
			}
			if observation.EpisodeID != field.EpisodeID || observation.Sequence > field.AuthoredAfter {
				return nil, fmt.Errorf("calibration case %q observation %q is outside the prediction birth boundary",
					candidate.ID, observationID)
			}
			observations = append(observations, observation)
		}
		categoryRight, err := b.resolveCalibrationCategory(*field, candidate)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, resolvedCalibrationCase{observations: observations, categoryRight: categoryRight})
	}

	calibrated := clonePredictionField(*field)
	comparisons := make([]ComparisonField, len(resolved))
	for i, candidate := range resolved {
		comparisons[i] = ComparePredictionField(calibrated, candidate.observations)
	}
	for i := range calibrated.Dimensions {
		calibrated.Dimensions[i].Calibration = measureDimensionCalibration(
			calibrated.Dimensions[i].ID, window, resolved, comparisons)
	}
	if !b.Expr.ReplaceUnsettledPredictionField(calibrated) {
		return nil, fmt.Errorf("prediction field %q already settled while calibration was being frozen", id)
	}
	stored, _ := b.Expr.PredictionField(id)
	return stored, nil
}

func (b *Being) resolveCalibrationCategory(field PredictionField, candidate CalibrationCase) (bool, error) {
	if candidate.CategoryPredictionID == "" || candidate.CategoryDimensionID == "" ||
		candidate.CategorySettledSpanID == "" {
		return false, fmt.Errorf("calibration case %q has an incomplete category-settlement reference", candidate.ID)
	}
	var found *DimensionComparison
	for _, comparison := range b.Expr.Comparisons(candidate.CategoryPredictionID) {
		if comparison.SettledSpan != candidate.CategorySettledSpanID {
			continue
		}
		for i := range comparison.Dimensions {
			dimension := &comparison.Dimensions[i]
			if dimension.Dimension.ID != candidate.CategoryDimensionID ||
				dimension.State != DimensionSettled || dimension.Truth == TruthUnknown {
				continue
			}
			if found != nil {
				return false, fmt.Errorf("calibration case %q category reference is ambiguous", candidate.ID)
			}
			copy := *dimension
			found = &copy
		}
	}
	if found == nil {
		return false, fmt.Errorf("calibration case %q category reference is not independently settled", candidate.ID)
	}
	for _, evidenceID := range found.Observed.Evidence {
		evidence, ok := b.Obs.Observation(evidenceID)
		if !ok || evidence.Provenance != ProvenanceRaw || evidence.EpisodeID != field.EpisodeID ||
			evidence.Sequence > field.AuthoredAfter {
			return false, fmt.Errorf("calibration case %q category settlement is not raw pre-birth evidence", candidate.ID)
		}
	}
	if len(found.Observed.Evidence) == 0 {
		return false, fmt.Errorf("calibration case %q category settlement has no evidence", candidate.ID)
	}
	return found.Truth == TruthTrue, nil
}

func measureDimensionCalibration(id DimensionID, window CalibrationWindow,
	cases []resolvedCalibrationCase, comparisons []ComparisonField) BaseRateCalibration {
	calibration := BaseRateCalibration{WindowID: window.ID}
	for i, comparison := range comparisons {
		var measured *DimensionComparison
		for j := range comparison.Dimensions {
			if comparison.Dimensions[j].Dimension.ID == id {
				measured = &comparison.Dimensions[j]
				break
			}
		}
		if measured == nil || measured.State != DimensionSettled || measured.Truth == TruthUnknown {
			continue
		}
		calibration.TotalSamples++
		if cases[i].categoryRight {
			calibration.CategorySamples++
			if measured.Truth == TruthTrue {
				calibration.CategoryFires++
			}
		} else {
			calibration.NonCategorySamples++
		}
		if measured.Truth == TruthTrue {
			calibration.TotalFires++
		}
	}
	if calibration.CategorySamples > 0 {
		calibration.Coverage = float64(calibration.CategoryFires) / float64(calibration.CategorySamples)
	}
	if calibration.TotalSamples > 0 {
		calibration.BaseRate = float64(calibration.TotalFires) / float64(calibration.TotalSamples)
	}
	if calibration.CategorySamples < window.MinimumCategorySamples ||
		calibration.NonCategorySamples < window.MinimumNonCategorySamples ||
		calibration.Coverage <= 0 || calibration.BaseRate <= 0 {
		return calibration
	}
	calibration.Surprise = math.Log(calibration.Coverage / calibration.BaseRate)
	calibration.ExpectedCredit = calibration.Coverage * calibration.Surprise
	calibration.Ready = true
	return calibration
}
