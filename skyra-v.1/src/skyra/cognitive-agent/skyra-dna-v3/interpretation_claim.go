package dnav3

import (
	"fmt"
	"strings"
)

// ConsequenceRef is the audited deduction selected by an interpretation. The
// interpretation never learns from this expect's outcome; the link records the
// forward reasoning chain while each claim settles through its own address.
type ConsequenceRef struct {
	ExpressionID string
	ExpectIndex  int
}

type InterpretationProposal struct {
	Category    string
	Confidence  float64
	Context     Triple // the interpretation's own category-discriminating prediction
	Within      int
	DerivedFrom []ObservationID
	Consequence ConsequenceRef
}

// InterpretationClaim is an abductive instance claim plus the distinct prediction
// that makes it falsifiable. Context is an Expect for settlement mechanics, but it
// is indexed and learned outside the action-consequence path.
type InterpretationClaim struct {
	ID        string
	EpisodeID EpisodeID
	SpanID    SpanID

	Category    string
	DerivedFrom []ObservationID
	Consequence ConsequenceRef
	Context     Expect

	// ArmedAfter is the observation-sequence boundary. Tick alone is insufficient:
	// several observations in one impulse share a tick. Only a later append may settle.
	ArmedAfter int
	Score      float64
}

// ProposeInterpretation admits an interpretation only when it already owns a
// separate falsification address and points to the consequence it licensed.
func (b *Being) ProposeInterpretation(proposal InterpretationProposal) (*InterpretationClaim, error) {
	if b == nil || b.Expr == nil || b.Obs == nil {
		return nil, fmt.Errorf("interpretation graph is unavailable")
	}
	category := strings.TrimSpace(proposal.Category)
	if category == "" {
		return nil, fmt.Errorf("interpretation category is empty")
	}
	if strings.ContainsRune(category, '\x00') {
		return nil, fmt.Errorf("interpretation category contains reserved separator")
	}
	if proposal.Confidence <= 0 || proposal.Confidence >= 1 {
		return nil, fmt.Errorf("interpretation confidence must be between 0 and 1")
	}
	if proposal.Within <= 0 {
		return nil, fmt.Errorf("interpretation context prediction has no maturity")
	}
	if proposal.Context.Subject == "" || proposal.Context.Predicate == "" || proposal.Context.Object == "" {
		return nil, fmt.Errorf("interpretation context prediction is not world-checkable")
	}
	expression, consequence, ok := b.Expr.Consequence(proposal.Consequence)
	if !ok {
		return nil, fmt.Errorf("interpretation names no existing consequence claim")
	}
	if len(consequence.Of) == 0 || consequence.Within <= 0 || consequence.Subject == "" ||
		consequence.Predicate == "" || consequence.Object == "" {
		return nil, fmt.Errorf("interpretation names an unfalsifiable consequence claim")
	}
	if consequence.Resolution != TruthUnknown {
		return nil, fmt.Errorf("interpretation cannot retroactively license a settled consequence")
	}
	span := b.ensureSpan()
	if expression.EpisodeID != b.episodeID || expression.SpanID != span {
		return nil, fmt.Errorf("interpretation and licensed consequence do not share the current episode/span")
	}
	if sp(proposal.Context) == sp(consequence.Triple) {
		return nil, fmt.Errorf("interpretation and consequence require distinct falsification addresses")
	}
	if len(proposal.DerivedFrom) == 0 {
		return nil, fmt.Errorf("interpretation has no raw abductive evidence")
	}
	for _, id := range proposal.DerivedFrom {
		o, found := b.Obs.Observation(id)
		if !found || o.Provenance != ProvenanceRaw || o.EpisodeID != b.episodeID || o.SpanID != span {
			return nil, fmt.Errorf("interpretation evidence %q is missing, non-raw, or outside the current span", id)
		}
	}

	context := proposal.Context
	context.Confidence = proposal.Confidence
	affordances := b.Aff
	if affordances == nil {
		affordances = DefaultAffordances()
	}
	context.Mode = affordances.Mode(context)
	claim := &InterpretationClaim{
		EpisodeID:   b.episodeID,
		SpanID:      span,
		Category:    category,
		DerivedFrom: append([]ObservationID(nil), proposal.DerivedFrom...),
		Consequence: proposal.Consequence,
		Context:     Expect{Triple: context, Within: proposal.Within},
		ArmedAfter:  b.observationSeq,
	}
	claim.Context.Arm(b.tick)
	b.interpretationSeq++
	id := fmt.Sprintf("%s/interpretation:%d", b.episodeID, b.interpretationSeq)
	b.Expr.PlaceInterpretation(id, claim)
	return claim, nil
}

func interpretationEdgeKey(category, target string) string {
	return category + "\x00interpretation\x00" + target
}
