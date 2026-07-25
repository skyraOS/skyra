package dnav3

import (
	"fmt"
	"math"
	"sort"
)

// ClaimContribution is one conservative positive-fire projection. Sensors names
// every ready candidate for the claim; Winner is selected by held-out expected
// credit before the current outcome is inspected. If that winner fired, its
// per-fire PMI supplies Credit.
type ClaimContribution struct {
	ClaimID ClaimID
	Winner  DimensionID
	Sensors []DimensionID
	Credit  float64
}

// FusionProjection is derived and reversible. The complete dimensional comparison
// remains authoritative; this scalar exists only for consumers that require one.
type FusionProjection struct {
	Claims []ClaimContribution
	Total  float64
}

// FusePositiveFirePMI implements the first same-claim fusion rule:
//
//   - each claim selects one representative by held-out ExpectedCredit alone;
//   - current settlement is inspected only after that selection;
//   - only a materialized, settled true winner with positive PMI contributes;
//   - distinct claims sum.
//
// False outcomes remain in the comparison but are excluded until non-fire scoring
// is settled. Winner-takes-claim is deliberately conservative and does not claim to
// be the eventual learned joint rule. Selecting among sensors that happened to fire
// would create an uncalibrated option-value loophole, so this function never does it.
func (field ComparisonField) FusePositiveFirePMI() (FusionProjection, error) {
	groups := map[ClaimID]*ClaimContribution{}
	winners := map[ClaimID]DimensionComparison{}
	for _, dimension := range field.Dimensions {
		if dimension.Dimension.ClaimID == "" {
			return FusionProjection{}, fmt.Errorf("dimension %q has no explicit claim identity",
				dimension.Dimension.ID)
		}
		if !dimension.Calibration.Ready() {
			continue
		}
		if math.IsNaN(dimension.Calibration.PMI) || math.IsInf(dimension.Calibration.PMI, 0) ||
			math.IsNaN(dimension.Calibration.ExpectedCredit) || math.IsInf(dimension.Calibration.ExpectedCredit, 0) {
			return FusionProjection{}, fmt.Errorf("dimension %q has non-finite calibration",
				dimension.Dimension.ID)
		}
		if dimension.Calibration.PMI <= 0 || dimension.Calibration.ExpectedCredit <= 0 {
			continue
		}
		claimID := dimension.Dimension.ClaimID
		claim := groups[claimID]
		if claim == nil {
			claim = &ClaimContribution{ClaimID: claimID}
			groups[claimID] = claim
		}
		claim.Sensors = append(claim.Sensors, dimension.Dimension.ID)
		winner, hasWinner := winners[claimID]
		if !hasWinner || dimension.Calibration.ExpectedCredit > winner.Calibration.ExpectedCredit ||
			(dimension.Calibration.ExpectedCredit == winner.Calibration.ExpectedCredit &&
				(claim.Winner == "" || dimension.Dimension.ID < claim.Winner)) {
			claim.Winner = dimension.Dimension.ID
			winners[claimID] = dimension
		}
	}

	projection := FusionProjection{Claims: make([]ClaimContribution, 0, len(groups))}
	for _, claim := range groups {
		sort.Slice(claim.Sensors, func(i, j int) bool { return claim.Sensors[i] < claim.Sensors[j] })
		winner := winners[claim.ClaimID]
		if winner.State != DimensionSettled || winner.Truth != TruthTrue {
			continue
		}
		claim.Credit = winner.Calibration.PMI
		projection.Claims = append(projection.Claims, *claim)
	}
	sort.Slice(projection.Claims, func(i, j int) bool {
		return projection.Claims[i].ClaimID < projection.Claims[j].ClaimID
	})
	for _, claim := range projection.Claims {
		projection.Total += claim.Credit
	}
	return projection, nil
}
