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
		if !categoryRight {
			markers = append(markers, "outside_only")
		}
		if (categoryRight && i < 2) || (!categoryRight && i < 68) {
			markers = append(markers, "negative_only")
		}
		span := b.startExternalSpan()
		labelID := PredictionID(fmt.Sprintf("independent-label-%d", i))
		labelTriple := Triple{Subject: fmt.Sprintf("case:/%d", i), Predicate: "category-right", Object: "true"}
		if _, err := b.AdmitPredictionField(exactPredictionField(labelID, "category", b.EpisodeID(),
			span, "calibration-label", string(labelID), b.observationSeq, labelTriple)); err != nil {
			t.Fatal(err)
		}
		observation := b.appendObservation(Triple{
			Subject: "message:/history", Predicate: "text", Object: strings.Join(markers, " "),
		}, ProvenanceRaw, 0, nil, span)
		labelOutcome := labelTriple
		if categoryRight {
			labelOutcome.Object = "true"
		} else {
			labelOutcome.Object = "false"
		}
		labelObservation := b.appendObservation(labelOutcome, ProvenanceRaw, 0, nil, span)
		authoredAfter = labelObservation.Sequence
		if _, err := b.CompareAndRecordPrediction(labelID, []Observation{labelObservation}); err != nil {
			t.Fatal(err)
		}
		window.Cases = append(window.Cases, CalibrationCase{
			ID: fmt.Sprintf("case-%d", i), ObservationIDs: []ObservationID{observation.ID},
			CategoryPredictionID: labelID, CategoryDimensionID: "category", CategorySettledSpanID: span,
		})
	}

	field := PredictionField{
		ID: "candidate-predicates", EpisodeID: b.EpisodeID(), SpanID: b.startExternalSpan(),
		AuthoredAfter: authoredAfter,
		Dimensions: []PredictionDimension{
			{ID: "qa", ClaimID: "traversal", Parser: ParserStringMarkers, Markers: []string{"qa"}},
			{ID: "qb", ClaimID: "traversal", Parser: ParserStringMarkers, Markers: []string{"qb"}},
			{ID: "qc", ClaimID: "traversal", Parser: ParserStringMarkers, Markers: []string{"qc"}},
			{ID: "qevil", ClaimID: "traversal", Parser: ParserStringMarkers, Markers: []string{"qevil"}},
			{ID: "qinert", ClaimID: "inert", Parser: ParserStringMarkers, Markers: []string{"never-fires"}},
			{ID: "qanti", ClaimID: "anti", Parser: ParserStringMarkers, Markers: []string{"outside_only"}},
			{ID: "qnegative", ClaimID: "negative", Parser: ParserStringMarkers, Markers: []string{"negative_only"}},
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
	for id, calibration := range byID {
		if math.IsNaN(calibration.PMI) || math.IsInf(calibration.PMI, 0) ||
			math.IsNaN(calibration.ExpectedCredit) || math.IsInf(calibration.ExpectedCredit, 0) {
			t.Fatalf("calibration %q stored a non-finite scalar: %+v", id, calibration)
		}
	}
	assertCalibration(t, byID["qa"], .30, .06, math.Log(5), .30*math.Log(5))
	assertCalibration(t, byID["qb"], .90, .25, math.Log(3.6), .90*math.Log(3.6))
	assertCalibration(t, byID["qc"], 1, .65, math.Log(1/.65), math.Log(1/.65))
	assertCalibration(t, byID["qevil"], 1, .90, math.Log(1/.90), math.Log(1/.90))
	if got := byID["qinert"]; got.State != CalibrationMeasuredInert || got.TotalFires != 0 ||
		got.PMI != 0 || got.ExpectedCredit != 0 {
		t.Fatalf("never-firing predicate=%+v, want measured inert with finite zero projection", got)
	}
	if got := byID["qanti"]; got.State != CalibrationMeasuredAntiDiagnostic ||
		got.CategoryFires != 0 || got.TotalFires != 80 || math.Abs(got.BaseRate-.8) > 1e-12 ||
		got.PMI != 0 || got.ExpectedCredit != 0 {
		t.Fatalf("outside-only predicate=%+v, want measured anti-diagnostic", got)
	}
	if got := byID["qnegative"]; got.State != CalibrationMeasuredAntiDiagnostic ||
		math.Abs(got.Coverage-.1) > 1e-12 || math.Abs(got.BaseRate-.5) > 1e-12 ||
		math.Abs(got.PMI-math.Log(.2)) > 1e-12 || math.Abs(got.ExpectedCredit-.1*math.Log(.2)) > 1e-12 {
		t.Fatalf("negative-PMI predicate=%+v, want finite measured anti-diagnostic", got)
	}
	if byID["qb"].ExpectedCredit <= byID["qa"].ExpectedCredit ||
		byID["qb"].ExpectedCredit <= byID["qc"].ExpectedCredit ||
		byID["qb"].ExpectedCredit <= byID["qevil"].ExpectedCredit {
		t.Fatalf("diagnostic middle did not win: qa=%g qb=%g qc=%g evil=%g",
			byID["qa"].ExpectedCredit, byID["qb"].ExpectedCredit,
			byID["qc"].ExpectedCredit, byID["qevil"].ExpectedCredit)
	}
	supportGated := PredictionField{
		ID: "support-gated", EpisodeID: b.EpisodeID(), SpanID: b.startExternalSpan(),
		AuthoredAfter: authoredAfter,
		Dimensions: []PredictionDimension{{
			ID: "qb", ClaimID: "traversal", Parser: ParserStringMarkers, Markers: []string{"qb"},
		}},
	}
	if _, err := b.AdmitPredictionField(supportGated); err != nil {
		t.Fatal(err)
	}
	thinWindow := window
	thinWindow.ID = "requires-one-more-positive"
	thinWindow.MinimumCategorySamples = 21
	provisional, err := b.CalibratePredictionField(supportGated.ID, thinWindow)
	if err != nil {
		t.Fatal(err)
	}
	if got := provisional.Dimensions[0].Calibration; got.State != CalibrationInsufficientSupport || got.Ready() ||
		got.CategorySamples != 20 ||
		got.NonCategorySamples != 80 || math.Abs(got.Coverage-.9) > 1e-12 || math.Abs(got.BaseRate-.25) > 1e-12 {
		t.Fatalf("sparse calibration did not retain counts while remaining provisional: %+v", got)
	}

	settlement := b.appendObservation(Triple{
		Subject: "message:/reply", Predicate: "text", Object: "qa qb qc qevil",
	}, ProvenanceRaw, 0, nil, b.startExternalSpan())
	comparison, err := b.CompareAndRecordPrediction(field.ID, []Observation{settlement})
	if err != nil {
		t.Fatal(err)
	}
	for _, dimension := range comparison.SettledDimensions() {
		if dimension.Calibration.State == CalibrationInsufficientSupport ||
			dimension.Calibration.WindowID != window.ID {
			t.Fatalf("runtime comparison lost frozen calibration: %+v", dimension)
		}
	}
	projection, err := comparison.FusePositiveFirePMI()
	if err != nil {
		t.Fatal(err)
	}
	if len(projection.Claims) != 1 || projection.Claims[0].ClaimID != "traversal" ||
		projection.Claims[0].Winner != "qb" || math.Abs(projection.Total-math.Log(3.6)) > 1e-12 {
		t.Fatalf("calibrated same-claim selection=%+v, want Q_B per-fire PMI once", projection)
	}
	if _, err := b.CalibratePredictionField(field.ID, window); err == nil {
		t.Fatal("settled prediction accepted retroactive recalibration")
	}
}

func TestToy9FCalibrationRejectsSelfLabelsAndSelectionOverlap(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	triple := Triple{
		Subject: "message:/history", Predicate: "text", Object: "traversal",
	}
	labelID := PredictionID("independent-label")
	if _, err := b.AdmitPredictionField(exactPredictionField(labelID, "category", b.EpisodeID(), span,
		"calibration-label", string(labelID), b.observationSeq, triple)); err != nil {
		t.Fatal(err)
	}
	observation := b.appendObservation(triple, ProvenanceRaw, 0, nil, span)
	if _, err := b.CompareAndRecordPrediction(labelID, []Observation{observation}); err != nil {
		t.Fatal(err)
	}
	field := PredictionField{
		ID: "candidate", EpisodeID: b.EpisodeID(), SpanID: span, AuthoredAfter: observation.Sequence,
		Dimensions: []PredictionDimension{{
			ID: "topic", ClaimID: "topic", Parser: ParserStringMarkers, Markers: []string{"traversal"},
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

func assertCalibration(t *testing.T, got BaseRateCalibration, coverage, baseRate, pmi, expected float64) {
	t.Helper()
	if got.State != CalibrationCalibrated || !got.Ready() || got.CategorySamples != 20 ||
		got.NonCategorySamples != 80 || got.TotalSamples != 100 ||
		math.Abs(got.Coverage-coverage) > 1e-12 || math.Abs(got.BaseRate-baseRate) > 1e-12 ||
		math.Abs(got.PMI-pmi) > 1e-12 || math.Abs(got.ExpectedCredit-expected) > 1e-12 {
		t.Fatalf("calibration=%+v, want coverage=%g base=%g PMI=%g expected=%g",
			got, coverage, baseRate, pmi, expected)
	}
}
