package dnav3

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

func TestToy9FHeldOutCalibrationSelectsDiagnosticMiddle(t *testing.T) {
	b := NewBeing(nil, nil)
	window := CalibrationWindow{
		ID: "held-out-100", MinimumCategorySamples: 20, MinimumNonCategorySamples: 80,
	}
	var authoredAfter int
	for i := 0; i < 100; i++ {
		var markers []string
		categoryRight := i < 20
		if categoryRight && i < 6 {
			markers = append(markers, "qa")
		}
		if (categoryRight && i < 18) || (!categoryRight && i < 27) {
			markers = append(markers, "qb")
		}
		if categoryRight || (!categoryRight && i < 65) {
			markers = append(markers, "qc")
		}
		if categoryRight || (!categoryRight && i < 90) {
			markers = append(markers, "qevil")
		}
		span := b.startExternalSpan()
		observation := b.appendObservation(Triple{
			Subject: "message:/history", Predicate: "text", Object: strings.Join(markers, " "),
		}, ProvenanceRaw, 0, nil, span)
		authoredAfter = observation.Sequence
		labelID := PredictionID(fmt.Sprintf("independent-label-%d", i))
		truth := TruthFalse
		if categoryRight {
			truth = TruthTrue
		}
		b.Expr.RecordComparison(ComparisonField{
			PredictionID: labelID, EpisodeID: b.EpisodeID(), SettledSpan: span,
			Dimensions: []DimensionComparison{{
				Dimension: PredictionDimension{ID: "category"}, State: DimensionSettled, Truth: truth,
				Observed: ObservedDimension{ID: "category", Evidence: []ObservationID{observation.ID}},
			}},
		})
		window.Cases = append(window.Cases, CalibrationCase{
			ID: fmt.Sprintf("case-%d", i), ObservationIDs: []ObservationID{observation.ID},
			CategoryPredictionID: labelID, CategoryDimensionID: "category", CategorySettledSpanID: span,
		})
	}

	field := PredictionField{
		ID: "candidate-predicates", EpisodeID: b.EpisodeID(), SpanID: b.startExternalSpan(),
		AuthoredAfter: authoredAfter,
		Dimensions: []PredictionDimension{
			{ID: "qa", Parser: ParserStringMarkers, Markers: []string{"qa"}},
			{ID: "qb", Parser: ParserStringMarkers, Markers: []string{"qb"}},
			{ID: "qc", Parser: ParserStringMarkers, Markers: []string{"qc"}},
			{ID: "qevil", Parser: ParserStringMarkers, Markers: []string{"qevil"}},
		},
	}
	if _, err := b.AdmitPredictionField(field); err != nil {
		t.Fatal(err)
	}
	calibrated, err := b.CalibratePredictionField(field.ID, window)
	if err != nil {
		t.Fatal(err)
	}
	byID := map[DimensionID]BaseRateCalibration{}
	for _, dimension := range calibrated.Dimensions {
		byID[dimension.ID] = dimension.Calibration
	}
	assertCalibration(t, byID["qa"], .30, .06, math.Log(5), .30*math.Log(5))
	assertCalibration(t, byID["qb"], .90, .25, math.Log(3.6), .90*math.Log(3.6))
	assertCalibration(t, byID["qc"], 1, .65, math.Log(1/.65), math.Log(1/.65))
	assertCalibration(t, byID["qevil"], 1, .90, math.Log(1/.90), math.Log(1/.90))
	if byID["qb"].ExpectedCredit <= byID["qa"].ExpectedCredit ||
		byID["qb"].ExpectedCredit <= byID["qc"].ExpectedCredit ||
		byID["qb"].ExpectedCredit <= byID["qevil"].ExpectedCredit {
		t.Fatalf("diagnostic middle did not win: qa=%g qb=%g qc=%g evil=%g",
			byID["qa"].ExpectedCredit, byID["qb"].ExpectedCredit,
			byID["qc"].ExpectedCredit, byID["qevil"].ExpectedCredit)
	}

	settlement := b.appendObservation(Triple{
		Subject: "message:/reply", Predicate: "text", Object: "qa qb qc qevil",
	}, ProvenanceRaw, 0, nil, b.startExternalSpan())
	comparison, err := b.CompareAndRecordPrediction(field.ID, []Observation{settlement})
	if err != nil {
		t.Fatal(err)
	}
	for _, dimension := range comparison.SettledDimensions() {
		if !dimension.Calibration.Ready || dimension.Calibration.WindowID != window.ID {
			t.Fatalf("runtime comparison lost frozen calibration: %+v", dimension)
		}
	}
	if _, err := b.CalibratePredictionField(field.ID, window); err == nil {
		t.Fatal("settled prediction accepted retroactive recalibration")
	}
}

func TestToy9FCalibrationRejectsSelfLabelsAndSelectionOverlap(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	observation := b.appendObservation(Triple{
		Subject: "message:/history", Predicate: "text", Object: "traversal",
	}, ProvenanceRaw, 0, nil, span)
	labelID := PredictionID("independent-label")
	b.Expr.RecordComparison(ComparisonField{
		PredictionID: labelID, EpisodeID: b.EpisodeID(), SettledSpan: span,
		Dimensions: []DimensionComparison{{
			Dimension: PredictionDimension{ID: "category"}, State: DimensionSettled, Truth: TruthTrue,
			Observed: ObservedDimension{ID: "category", Evidence: []ObservationID{observation.ID}},
		}},
	})
	field := PredictionField{
		ID: "candidate", EpisodeID: b.EpisodeID(), SpanID: span, AuthoredAfter: observation.Sequence,
		Dimensions: []PredictionDimension{{
			ID: "topic", Parser: ParserStringMarkers, Markers: []string{"traversal"},
		}},
		Attributions: []EvidenceAttribution{{
			ObservationID: observation.ID, StartChar: 0, EndChar: 9,
			PredictionID: "candidate", DimensionID: "topic", ClaimedInfluence: 1,
		}},
	}
	if _, err := b.AdmitPredictionField(field); err != nil {
		t.Fatal(err)
	}
	base := CalibrationWindow{
		ID: "blind", MinimumCategorySamples: 1, MinimumNonCategorySamples: 1,
		Cases: []CalibrationCase{{
			ID: "overlap", ObservationIDs: []ObservationID{observation.ID},
			CategoryPredictionID: labelID, CategoryDimensionID: "category", CategorySettledSpanID: span,
		}},
	}
	if _, err := b.CalibratePredictionField(field.ID, base); err == nil || !strings.Contains(err.Error(), "selection evidence") {
		t.Fatalf("selection/evaluation overlap error=%v", err)
	}
	base.Cases[0].CategoryPredictionID = field.ID
	if _, err := b.CalibratePredictionField(field.ID, base); err == nil || !strings.Contains(err.Error(), "label itself") {
		t.Fatalf("self-label error=%v", err)
	}
}

func assertCalibration(t *testing.T, got BaseRateCalibration, coverage, baseRate, surprise, expected float64) {
	t.Helper()
	if !got.Ready || got.CategorySamples != 20 || got.NonCategorySamples != 80 || got.TotalSamples != 100 ||
		math.Abs(got.Coverage-coverage) > 1e-12 || math.Abs(got.BaseRate-baseRate) > 1e-12 ||
		math.Abs(got.Surprise-surprise) > 1e-12 || math.Abs(got.ExpectedCredit-expected) > 1e-12 {
		t.Fatalf("calibration=%+v, want coverage=%g base=%g surprise=%g expected=%g",
			got, coverage, baseRate, surprise, expected)
	}
}
