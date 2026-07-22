package dnav3

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"
)

// PredictionID and DimensionID are stable identities. Attributions point to these
// names rather than slice positions, so serialization or insertion order cannot
// silently redirect evidence.
type PredictionID string
type DimensionID string

// PredictionParserID names one deterministic evaluator in the closed Phase-4F
// registry. The LLM may choose a parser and its arguments; it cannot add code to
// the grading path.
type PredictionParserID string

const (
	ParserExactAddress   PredictionParserID = "exact_address"
	ParserTokenCount     PredictionParserID = "token_count"
	ParserParagraphCount PredictionParserID = "paragraph_count"
	ParserStringMarkers  PredictionParserID = "string_markers"
)

// NumberRange is an inclusive predicted range. It stays a simple envelope while
// the comparison retains the actual number and its signed distance from the range.
type NumberRange struct {
	Low  float64
	High float64
}

// PredictionDimension is one independently inspectable coordinate of a future
// observation. Parser selects which argument family is meaningful.
type PredictionDimension struct {
	ID     DimensionID
	Parser PredictionParserID

	// CarrierDimensionID and CarrierSlot express structural ownership. A
	// one-based slot materializes only when the carrier parser observes that
	// many units. An absent slot voids this dimension rather than falsifying it.
	CarrierDimensionID DimensionID
	CarrierSlot        int

	// exact_address
	Exact Triple

	// token_count / paragraph_count
	Range NumberRange

	// string_markers
	Markers    []string
	MinMatches int

	// Calibration is frozen before the predicted observation arrives. Runtime
	// settlement copies it into the comparison; the settling impulse never
	// participates in estimating its own information value.
	Calibration BaseRateCalibration
}

// EvidenceAttribution is the LLM's frozen claim that one passage in an already
// observed raw value influenced one prediction dimension in a given direction.
// ClaimedInfluence is authored alignment, not probability and not earned credit.
type EvidenceAttribution struct {
	ObservationID    ObservationID
	StartChar        int // Unicode code-point offset, inclusive
	EndChar          int // Unicode code-point offset, exclusive
	PredictionID     PredictionID
	DimensionID      DimensionID
	ClaimedInfluence float64 // [-1,1]; negative is active counterevidence
}

// PredictionField is owned by the span/claim that authored it. AuthoredAfter is
// the observation sequence visible at birth; attribution cannot point through it
// into future or settlement evidence.
type PredictionField struct {
	ID            PredictionID
	EpisodeID     EpisodeID
	SpanID        SpanID
	OwnerKind     string
	OwnerID       string
	AuthoredAfter int

	Dimensions   []PredictionDimension
	Attributions []EvidenceAttribution
}

type DimensionSettlement string

const (
	DimensionUnknown DimensionSettlement = "unknown"
	DimensionSettled DimensionSettlement = "settled"
	DimensionVoid    DimensionSettlement = "void"
)

// ObservedDimension is the deterministic parser output. Only the field selected
// by Parser is meaningful, but the uniform record keeps comparison storage simple.
type ObservedDimension struct {
	ID       DimensionID
	Parser   PredictionParserID
	Exact    *Triple
	Number   float64
	Markers  []string
	Evidence []ObservationID
}

// BaseRateCalibration retains the positive-fire diagnosticity objective and the
// witnessed counts behind it. Ready means the caller-declared held-out support
// floor was met and the ratio was arithmetically defined; it is not promotion.
type BaseRateCalibration struct {
	Ready              bool
	WindowID           string
	CategorySamples    int
	NonCategorySamples int
	TotalSamples       int
	CategoryFires      int
	TotalFires         int
	Coverage           float64
	BaseRate           float64
	Surprise           float64
	ExpectedCredit     float64
}

// DimensionComparison is the durable learning coordinate: prediction, actual
// parser output, residual/relation, settlement state, and evidence all survive.
// Truth is a compatibility projection for the current Bernoulli Edge learner.
type DimensionComparison struct {
	Dimension   PredictionDimension
	Observed    ObservedDimension
	State       DimensionSettlement
	Truth       Truth
	Relation    string
	Residual    float64
	Calibration BaseRateCalibration
}

type ComparisonField struct {
	PredictionID PredictionID
	EpisodeID    EpisodeID
	OwnerSpan    SpanID
	SettledSpan  SpanID
	Dimensions   []DimensionComparison
}

func exactPredictionField(id PredictionID, dimensionID DimensionID, episode EpisodeID, span SpanID,
	ownerKind, ownerID string, authoredAfter int, expected Triple) PredictionField {
	if dimensionID == "" {
		dimensionID = "exact"
	}
	return PredictionField{
		ID:            id,
		EpisodeID:     episode,
		SpanID:        span,
		OwnerKind:     ownerKind,
		OwnerID:       ownerID,
		AuthoredAfter: authoredAfter,
		Dimensions: []PredictionDimension{{
			ID: dimensionID, Parser: ParserExactAddress, Exact: expected,
		}},
	}
}

// prepareExpressionPredictionFields namespaces model-authored local prediction IDs
// and validates optional passage attributions against the observations visible at
// birth. Bad attribution lines are dropped without discarding an otherwise valid act.
func (b *Being) prepareExpressionPredictionFields(expressionID string, expression *Expression) []Reject {
	if b == nil || expression == nil {
		return nil
	}
	expression.PredictionAuthoredAfter = b.observationSeq
	local := map[PredictionID]*Expect{}
	duplicates := map[PredictionID]bool{}
	counts := map[PredictionID]int{}
	for i := range expression.Expects {
		x := &expression.Expects[i]
		id := x.PredictionID
		if id == "" {
			id = PredictionID(fmt.Sprintf("expect:%d", i))
		}
		instanceID := id
		if counts[id] > 0 {
			duplicates[id] = true
			instanceID = PredictionID(fmt.Sprintf("%s#%d", id, counts[id]))
		} else {
			local[id] = x
		}
		counts[id]++
		x.PredictionID = PredictionID(expressionID + "/" + string(instanceID))
		if x.DimensionID == "" {
			x.DimensionID = "exact"
		}
	}

	accepted := make([]EvidenceAttribution, 0, len(expression.PredictionAttributions))
	var rejects []Reject
	for _, attribution := range expression.PredictionAttributions {
		original := attribution
		x := local[attribution.PredictionID]
		if x == nil || duplicates[attribution.PredictionID] {
			rejects = append(rejects, Reject{Line: formatEvidenceAttribution(original),
				Reason: "evidence points to a missing or duplicate local prediction ID"})
			continue
		}
		if attribution.DimensionID != x.DimensionID {
			rejects = append(rejects, Reject{Line: formatEvidenceAttribution(original),
				Reason: fmt.Sprintf("evidence points to unknown dimension %q", attribution.DimensionID)})
			continue
		}
		attribution.PredictionID = x.PredictionID
		field := exactPredictionField(x.PredictionID, x.DimensionID, expression.EpisodeID,
			expression.SpanID, "consequence", expressionID, expression.PredictionAuthoredAfter, x.Triple)
		field.Attributions = []EvidenceAttribution{attribution}
		if err := validatePredictionField(field, b.Obs.Observation); err != nil {
			rejects = append(rejects, Reject{Line: formatEvidenceAttribution(original), Reason: err.Error()})
			continue
		}
		accepted = append(accepted, attribution)
	}
	expression.PredictionAttributions = accepted
	return rejects
}

func formatEvidenceAttribution(attribution EvidenceAttribution) string {
	return fmt.Sprintf("[%s][%d:%d]:%g->[%s.%s]", attribution.ObservationID,
		attribution.StartChar, attribution.EndChar, attribution.ClaimedInfluence,
		attribution.PredictionID, attribution.DimensionID)
}

// AdmitPredictionField validates and durably places one authored field. The
// caller supplies stable IDs; the substrate stamps ownership time when omitted.
func (b *Being) AdmitPredictionField(field PredictionField) (*PredictionField, error) {
	if b == nil || b.Expr == nil || b.Obs == nil {
		return nil, fmt.Errorf("prediction field storage is unavailable")
	}
	if field.EpisodeID == "" {
		field.EpisodeID = b.episodeID
	}
	if field.SpanID == "" {
		field.SpanID = b.ensureSpan()
	}
	if field.AuthoredAfter == 0 {
		field.AuthoredAfter = b.observationSeq
	}
	if err := validatePredictionField(field, b.Obs.Observation); err != nil {
		return nil, err
	}
	if _, exists := b.Expr.PredictionField(field.ID); exists {
		return nil, fmt.Errorf("prediction field %q already exists", field.ID)
	}
	b.Expr.PlacePredictionField(field)
	stored, _ := b.Expr.PredictionField(field.ID)
	return stored, nil
}

func validatePredictionField(field PredictionField, observation func(ObservationID) (Observation, bool)) error {
	if field.ID == "" {
		return fmt.Errorf("prediction field has no stable ID")
	}
	if field.EpisodeID == "" || field.SpanID == "" {
		return fmt.Errorf("prediction field has no episode/span owner")
	}
	if field.AuthoredAfter < 0 {
		return fmt.Errorf("prediction field has an invalid birth boundary")
	}
	if len(field.Dimensions) == 0 {
		return fmt.Errorf("prediction field has no dimensions")
	}
	dimensions := make(map[DimensionID]struct{}, len(field.Dimensions))
	for _, dimension := range field.Dimensions {
		if dimension.ID == "" {
			return fmt.Errorf("prediction field contains an unnamed dimension")
		}
		if _, duplicate := dimensions[dimension.ID]; duplicate {
			return fmt.Errorf("prediction field contains duplicate dimension %q", dimension.ID)
		}
		dimensions[dimension.ID] = struct{}{}
		if err := validatePredictionDimension(dimension); err != nil {
			return fmt.Errorf("dimension %q: %w", dimension.ID, err)
		}
	}
	for _, dimension := range field.Dimensions {
		if dimension.CarrierDimensionID == "" {
			continue
		}
		if dimension.CarrierDimensionID == dimension.ID {
			return fmt.Errorf("dimension %q cannot license itself", dimension.ID)
		}
		var carrier *PredictionDimension
		for i := range field.Dimensions {
			if field.Dimensions[i].ID == dimension.CarrierDimensionID {
				carrier = &field.Dimensions[i]
				break
			}
		}
		if carrier == nil {
			return fmt.Errorf("dimension %q names missing carrier %q", dimension.ID, dimension.CarrierDimensionID)
		}
		if carrier.Parser != ParserParagraphCount {
			return fmt.Errorf("dimension %q carrier %q is not structural", dimension.ID, dimension.CarrierDimensionID)
		}
	}

	for _, attribution := range field.Attributions {
		if attribution.PredictionID != field.ID {
			return fmt.Errorf("attribution points to prediction %q, want %q", attribution.PredictionID, field.ID)
		}
		if _, ok := dimensions[attribution.DimensionID]; !ok {
			return fmt.Errorf("attribution points to unknown dimension %q", attribution.DimensionID)
		}
		if math.IsNaN(attribution.ClaimedInfluence) || math.IsInf(attribution.ClaimedInfluence, 0) ||
			attribution.ClaimedInfluence < -1 || attribution.ClaimedInfluence > 1 {
			return fmt.Errorf("attribution influence must be finite and inside [-1,1]")
		}
		o, ok := observation(attribution.ObservationID)
		if !ok || o.Provenance != ProvenanceRaw {
			return fmt.Errorf("attribution observation %q is missing or non-raw", attribution.ObservationID)
		}
		if o.EpisodeID != field.EpisodeID || o.Sequence > field.AuthoredAfter {
			return fmt.Errorf("attribution observation %q is outside the prediction birth boundary", attribution.ObservationID)
		}
		length := utf8.RuneCountInString(o.Object)
		if attribution.StartChar < 0 || attribution.EndChar <= attribution.StartChar || attribution.EndChar > length {
			return fmt.Errorf("attribution range [%d:%d] is invalid for observation %q length %d",
				attribution.StartChar, attribution.EndChar, attribution.ObservationID, length)
		}
	}
	return nil
}

func validatePredictionDimension(dimension PredictionDimension) error {
	switch dimension.Parser {
	case ParserExactAddress:
		if dimension.Exact.Subject == "" || dimension.Exact.Predicate == "" || dimension.Exact.Object == "" {
			return fmt.Errorf("exact parser requires a complete expected triple")
		}
	case ParserTokenCount, ParserParagraphCount:
		if math.IsNaN(dimension.Range.Low) || math.IsNaN(dimension.Range.High) ||
			math.IsInf(dimension.Range.Low, 0) || math.IsInf(dimension.Range.High, 0) ||
			dimension.Range.Low < 0 || dimension.Range.High < dimension.Range.Low {
			return fmt.Errorf("count parser requires a finite ordered non-negative range")
		}
	case ParserStringMarkers:
		if len(dimension.Markers) == 0 {
			return fmt.Errorf("marker parser requires at least one marker")
		}
		for _, marker := range dimension.Markers {
			if strings.TrimSpace(marker) == "" {
				return fmt.Errorf("marker parser contains an empty marker")
			}
		}
		if dimension.MinMatches < 0 || dimension.MinMatches > len(dimension.Markers) {
			return fmt.Errorf("marker parser has invalid minimum match count")
		}
	default:
		return fmt.Errorf("unknown fixed parser %q", dimension.Parser)
	}
	if dimension.CarrierDimensionID == "" && dimension.CarrierSlot != 0 {
		return fmt.Errorf("carrier slot requires a carrier dimension")
	}
	if dimension.CarrierDimensionID != "" && dimension.CarrierSlot < 1 {
		return fmt.Errorf("carrier slot must be one-based")
	}
	return nil
}

// ComparePredictionField runs only fixed deterministic parsers. It does not grade
// an Edge; callers may project settled comparisons afterward without discarding
// this record.
func ComparePredictionField(field PredictionField, observations []Observation) ComparisonField {
	comparison := ComparisonField{
		PredictionID: field.ID,
		EpisodeID:    field.EpisodeID,
		OwnerSpan:    field.SpanID,
	}
	if len(observations) > 0 {
		comparison.SettledSpan = observations[len(observations)-1].SpanID
	}
	dimensions := append([]PredictionDimension(nil), field.Dimensions...)
	sort.Slice(dimensions, func(i, j int) bool { return dimensions[i].ID < dimensions[j].ID })
	byID := make(map[DimensionID]DimensionComparison, len(dimensions))
	for _, dimension := range dimensions {
		if dimension.CarrierDimensionID != "" {
			continue
		}
		byID[dimension.ID] = comparePredictionDimension(dimension, observations)
	}
	for _, dimension := range dimensions {
		if dimension.CarrierDimensionID == "" {
			continue
		}
		carrier, ok := byID[dimension.CarrierDimensionID]
		switch {
		case !ok || carrier.State == DimensionUnknown:
			byID[dimension.ID] = unmaterializedDimension(dimension, DimensionUnknown, "carrier_unknown")
		case carrier.State == DimensionVoid:
			byID[dimension.ID] = unmaterializedDimension(dimension, DimensionVoid, "carrier_void")
		case int(carrier.Observed.Number) < dimension.CarrierSlot:
			byID[dimension.ID] = unmaterializedDimension(dimension, DimensionVoid, "carrier_absent")
		default:
			byID[dimension.ID] = comparePredictionDimension(dimension,
				paragraphSlotObservations(observations, dimension.CarrierSlot))
		}
	}
	for _, dimension := range dimensions {
		comparison.Dimensions = append(comparison.Dimensions, byID[dimension.ID])
	}
	return comparison
}

// SettledDimensions is the only generic projection surface exposed to learners.
// Unknown and void coordinates stay durable in the comparison but cannot become
// negative evidence merely because no eligible observation materialized.
func (field ComparisonField) SettledDimensions() []DimensionComparison {
	settled := make([]DimensionComparison, 0, len(field.Dimensions))
	for _, dimension := range field.Dimensions {
		if dimension.State == DimensionSettled && dimension.Truth != TruthUnknown {
			settled = append(settled, dimension)
		}
	}
	return settled
}

func unmaterializedDimension(dimension PredictionDimension, state DimensionSettlement, relation string) DimensionComparison {
	return DimensionComparison{
		Dimension:   dimension,
		Observed:    ObservedDimension{ID: dimension.ID, Parser: dimension.Parser},
		State:       state,
		Truth:       TruthUnknown,
		Relation:    relation,
		Calibration: dimension.Calibration,
	}
}

// CompareAndRecordPrediction evaluates one admitted field and retains the full
// comparison without updating any learned Edge. Grading remains an explicit later
// projection, which keeps raw dimensional evidence from being replaced by a score.
func (b *Being) CompareAndRecordPrediction(id PredictionID, observations []Observation) (ComparisonField, error) {
	if b == nil || b.Expr == nil {
		return ComparisonField{}, fmt.Errorf("prediction field storage is unavailable")
	}
	field, ok := b.Expr.PredictionField(id)
	if !ok {
		return ComparisonField{}, fmt.Errorf("prediction field %q does not exist", id)
	}
	for _, observation := range observations {
		if observation.EpisodeID != "" && observation.EpisodeID != field.EpisodeID {
			return ComparisonField{}, fmt.Errorf("observation %q belongs to another episode", observation.ID)
		}
		if observation.Sequence > 0 && observation.Sequence <= field.AuthoredAfter {
			return ComparisonField{}, fmt.Errorf("observation %q does not follow prediction birth", observation.ID)
		}
	}
	comparison := ComparePredictionField(*field, observations)
	b.Expr.RecordComparison(comparison)
	return comparison, nil
}

func (b *Being) compareExactExpectation(x *Expect, observation Observation) Truth {
	if b == nil || b.Expr == nil || x == nil {
		return TruthUnknown
	}
	comparison, err := b.CompareAndRecordPrediction(x.PredictionID, []Observation{observation})
	if err != nil {
		return TruthUnknown
	}
	for _, dimension := range comparison.SettledDimensions() {
		if dimension.Dimension.ID == x.DimensionID {
			return dimension.Truth
		}
	}
	return TruthUnknown
}

func comparePredictionDimension(dimension PredictionDimension, observations []Observation) DimensionComparison {
	result := DimensionComparison{
		Dimension:   dimension,
		Observed:    ObservedDimension{ID: dimension.ID, Parser: dimension.Parser},
		State:       DimensionUnknown,
		Truth:       TruthUnknown,
		Calibration: dimension.Calibration,
	}
	switch dimension.Parser {
	case ParserExactAddress:
		var found *Observation
		for i := range observations {
			o := &observations[i]
			if o.Fidelity() == 0 {
				continue
			}
			if sp(o.Triple) == sp(dimension.Exact) && (found == nil || o.Sequence >= found.Sequence) {
				found = o
			}
		}
		if found == nil {
			return result
		}
		actual := found.Triple
		result.Observed.Exact = &actual
		result.Observed.Evidence = []ObservationID{found.ID}
		result.State = DimensionSettled
		if actual.Object == dimension.Exact.Object {
			result.Truth, result.Relation = TruthTrue, "equal"
		} else {
			result.Truth, result.Relation = TruthFalse, "different"
		}
	case ParserTokenCount:
		if len(observations) == 0 {
			return result
		}
		for _, o := range observations {
			if o.Fidelity() == 0 {
				continue
			}
			count := o.TokenCount
			if count <= 0 {
				count = ApproxTokenCount(observationMemoryText(o))
			}
			result.Observed.Number += float64(count)
			result.Observed.Evidence = append(result.Observed.Evidence, o.ID)
		}
		if len(result.Observed.Evidence) == 0 {
			return result
		}
		result.State = DimensionSettled
		actual := result.Observed.Number
		switch {
		case actual < dimension.Range.Low:
			result.Truth, result.Relation = TruthFalse, "below_range"
			result.Residual = actual - dimension.Range.Low
		case actual > dimension.Range.High:
			result.Truth, result.Relation = TruthFalse, "above_range"
			result.Residual = actual - dimension.Range.High
		default:
			result.Truth, result.Relation = TruthTrue, "in_range"
		}
	case ParserParagraphCount:
		paragraphs := observedParagraphs(observations)
		for _, observation := range observations {
			if observation.Fidelity() != 0 {
				result.Observed.Evidence = append(result.Observed.Evidence, observation.ID)
			}
		}
		if len(result.Observed.Evidence) == 0 {
			return result
		}
		result.Observed.Number = float64(len(paragraphs))
		result.State = DimensionSettled
		actual := result.Observed.Number
		switch {
		case actual < dimension.Range.Low:
			result.Truth, result.Relation = TruthFalse, "below_range"
			result.Residual = actual - dimension.Range.Low
		case actual > dimension.Range.High:
			result.Truth, result.Relation = TruthFalse, "above_range"
			result.Residual = actual - dimension.Range.High
		default:
			result.Truth, result.Relation = TruthTrue, "in_range"
		}
	case ParserStringMarkers:
		minimum := dimension.MinMatches
		if minimum == 0 {
			minimum = 1
		}
		seen := map[string]bool{}
		for _, o := range observations {
			if o.Fidelity() == 0 {
				continue
			}
			result.Observed.Evidence = append(result.Observed.Evidence, o.ID)
			text := strings.ToLower(observationMemoryText(o))
			for _, marker := range dimension.Markers {
				canonical := strings.ToLower(strings.TrimSpace(marker))
				if strings.Contains(text, canonical) {
					seen[canonical] = true
				}
			}
		}
		if len(result.Observed.Evidence) == 0 {
			return result
		}
		for marker := range seen {
			result.Observed.Markers = append(result.Observed.Markers, marker)
		}
		sort.Strings(result.Observed.Markers)
		result.State = DimensionSettled
		result.Residual = float64(len(result.Observed.Markers) - minimum)
		if len(result.Observed.Markers) >= minimum {
			result.Truth, result.Relation = TruthTrue, "minimum_met"
		} else {
			result.Truth, result.Relation = TruthFalse, "minimum_missed"
		}
	}
	return result
}

type observedParagraph struct {
	text     string
	evidence ObservationID
}

func observedParagraphs(observations []Observation) []observedParagraph {
	var paragraphs []observedParagraph
	for _, observation := range observations {
		if observation.Fidelity() == 0 {
			continue
		}
		var current []string
		flush := func() {
			text := strings.TrimSpace(strings.Join(current, "\n"))
			if text != "" {
				paragraphs = append(paragraphs, observedParagraph{text: text, evidence: observation.ID})
			}
			current = nil
		}
		for _, line := range strings.Split(strings.ReplaceAll(observation.Object, "\r\n", "\n"), "\n") {
			if strings.TrimSpace(line) == "" {
				flush()
				continue
			}
			current = append(current, line)
		}
		flush()
	}
	return paragraphs
}

func paragraphSlotObservations(observations []Observation, slot int) []Observation {
	paragraphs := observedParagraphs(observations)
	if slot < 1 || slot > len(paragraphs) {
		return nil
	}
	paragraph := paragraphs[slot-1]
	for _, observation := range observations {
		if observation.ID != paragraph.evidence {
			continue
		}
		scoped := observation
		scoped.Object = paragraph.text
		scoped.TokenCount = ApproxTokenCount(paragraph.text)
		return []Observation{scoped}
	}
	return nil
}

func clonePredictionField(field PredictionField) PredictionField {
	out := field
	out.Dimensions = append([]PredictionDimension(nil), field.Dimensions...)
	for i := range out.Dimensions {
		out.Dimensions[i].Markers = append([]string(nil), field.Dimensions[i].Markers...)
	}
	out.Attributions = append([]EvidenceAttribution(nil), field.Attributions...)
	return out
}

func cloneComparisonField(field ComparisonField) ComparisonField {
	out := field
	out.Dimensions = append([]DimensionComparison(nil), field.Dimensions...)
	for i := range out.Dimensions {
		out.Dimensions[i].Dimension.Markers = append([]string(nil), field.Dimensions[i].Dimension.Markers...)
		out.Dimensions[i].Observed.Markers = append([]string(nil), field.Dimensions[i].Observed.Markers...)
		out.Dimensions[i].Observed.Evidence = append([]ObservationID(nil), field.Dimensions[i].Observed.Evidence...)
		if field.Dimensions[i].Observed.Exact != nil {
			exact := *field.Dimensions[i].Observed.Exact
			out.Dimensions[i].Observed.Exact = &exact
		}
	}
	return out
}
