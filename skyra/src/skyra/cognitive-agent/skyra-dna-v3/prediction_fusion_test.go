package dnav3

import (
	"math"
	"reflect"
	"testing"
)

func TestToy9FSameClaimSensorsFuseAndDistinctClaimsSum(t *testing.T) {
	b := NewBeing(nil, nil)
	span := b.startExternalSpan()
	ready := func(pmi, expected float64) BaseRateCalibration {
		return BaseRateCalibration{
			State: CalibrationCalibrated, WindowID: "held-out", PMI: pmi, ExpectedCredit: expected,
		}
	}
	field := PredictionField{
		ID: "fusion", EpisodeID: b.EpisodeID(), SpanID: span,
		Dimensions: []PredictionDimension{
			{ID: "structure.paragraphs", ClaimID: "response-structure", Parser: ParserParagraphCount,
				Range: NumberRange{Low: 1, High: 1}, Calibration: ready(.4, .3)},
			{ID: "traversal.a", ClaimID: "traversal-topic", Parser: ParserStringMarkers,
				Markers: []string{"alpha"}, Calibration: ready(1.2, .8)},
			{ID: "traversal.duplicate", ClaimID: "traversal-topic", Parser: ParserStringMarkers,
				Markers: []string{"beta"}, Calibration: ready(1.2, .8)},
			{ID: "traversal.broad", ClaimID: "traversal-topic", Parser: ParserStringMarkers,
				Markers: []string{"gamma"}, Calibration: ready(.1, .1)},
			{ID: "paragraph.2.code", ClaimID: "paragraph-2-code", Parser: ParserStringMarkers,
				Markers: []string{"beta"}, CarrierDimensionID: "structure.paragraphs", CarrierSlot: 2,
				Calibration: ready(9, 8)},
			{ID: "unobserved.address", ClaimID: "unobserved", Parser: ParserExactAddress,
				Exact:       Triple{Subject: "missing:/address", Predicate: "state", Object: "yes"},
				Calibration: ready(9, 8)},
			{ID: "uncalibrated", ClaimID: "uncalibrated", Parser: ParserStringMarkers,
				Markers: []string{"alpha"}, Calibration: BaseRateCalibration{PMI: 9}},
			{ID: "anti-diagnostic", ClaimID: "anti-diagnostic", Parser: ParserStringMarkers,
				Markers: []string{"alpha"}, Calibration: ready(-1, -1)},
			{ID: "lottery.best", ClaimID: "sensor-lottery", Parser: ParserStringMarkers,
				Markers: []string{"does-not-fire"}, Calibration: ready(2, .9)},
			{ID: "lottery.fallback", ClaimID: "sensor-lottery", Parser: ParserStringMarkers,
				Markers: []string{"alpha"}, Calibration: ready(1, .8)},
		},
	}
	settlement := b.appendObservation(Triple{
		Subject: "message:/reply", Predicate: "text", Object: "alpha beta gamma occur in one paragraph",
	}, ProvenanceRaw, 0, nil, b.startExternalSpan())
	comparison := ComparePredictionField(field, []Observation{settlement})
	before := cloneComparisonField(comparison)
	projection, err := comparison.FusePositiveFirePMI()
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(projection.Total-1.6) > 1e-12 || len(projection.Claims) != 2 {
		t.Fatalf("projection=%+v, want traversal 1.2 plus structure 0.4", projection)
	}
	byClaim := map[ClaimID]ClaimContribution{}
	for _, claim := range projection.Claims {
		byClaim[claim.ClaimID] = claim
	}
	traversal := byClaim["traversal-topic"]
	if traversal.Winner != "traversal.a" || math.Abs(traversal.Credit-1.2) > 1e-12 ||
		len(traversal.Sensors) != 3 {
		t.Fatalf("same-claim fusion=%+v, want strongest sensor once and all provenance", traversal)
	}
	if structure := byClaim["response-structure"]; structure.Winner != "structure.paragraphs" ||
		math.Abs(structure.Credit-.4) > 1e-12 {
		t.Fatalf("distinct structure contribution=%+v", structure)
	}
	if _, exists := byClaim["sensor-lottery"]; exists {
		t.Fatal("fusion fell back to a weaker sensor after the calibration-selected winner missed")
	}
	if !reflect.DeepEqual(comparison, before) {
		t.Fatal("fusion projection mutated the durable comparison")
	}

	reordered := clonePredictionField(field)
	for i, j := 0, len(reordered.Dimensions)-1; i < j; i, j = i+1, j-1 {
		reordered.Dimensions[i], reordered.Dimensions[j] = reordered.Dimensions[j], reordered.Dimensions[i]
	}
	reorderedComparison := ComparePredictionField(reordered, []Observation{settlement})
	reorderedProjection, err := reorderedComparison.FusePositiveFirePMI()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(projection, reorderedProjection) {
		t.Fatalf("dimension order changed fusion:\n%+v\n%+v", projection, reorderedProjection)
	}
}

func TestToy9FFusionRequiresExplicitClaimIdentity(t *testing.T) {
	comparison := ComparisonField{Dimensions: []DimensionComparison{{
		Dimension: PredictionDimension{ID: "sensor"}, State: DimensionSettled, Truth: TruthTrue,
		Calibration: BaseRateCalibration{State: CalibrationCalibrated, PMI: 1},
	}}}
	if _, err := comparison.FusePositiveFirePMI(); err == nil {
		t.Fatal("fusion accepted a sensor with no explicit claim identity")
	}
}

func TestToy9FAdmissionRejectsAuthorSuppliedCalibration(t *testing.T) {
	b := NewBeing(nil, nil)
	field := PredictionField{
		ID: "forged-calibration", EpisodeID: b.EpisodeID(), SpanID: b.startExternalSpan(),
		Dimensions: []PredictionDimension{{
			ID: "sensor", ClaimID: "claim", Parser: ParserStringMarkers, Markers: []string{"anything"},
			Calibration: BaseRateCalibration{State: CalibrationCalibrated, WindowID: "invented", PMI: 100},
		}},
	}
	if _, err := b.AdmitPredictionField(field); err == nil {
		t.Fatal("prediction admission accepted author-supplied calibration")
	}
}
