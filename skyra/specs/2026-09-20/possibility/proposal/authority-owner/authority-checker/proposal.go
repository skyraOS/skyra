//go:build ignore

// Package authoritychecker contains interface proposals, not runtime code.
// Move is the Skyra protocol Move; Horizon is declared in ../horizon.
// Package wiring remains provisional, as in the other dated proposals.
package authoritychecker

import "context"

// AuthorizedMove is just the checked Move together with its resolved Horizon.
// Horizon comes from the runtime's Possibility namespace, not from a claim
// supplied by the Move's sender. This wrapper is an internal runtime result.
type AuthorizedMove struct {
	Move    Move
	Horizon Horizon
}

// AuthorityChecker resolves move.PossibilityID against the Possibility
// namespace and checks whether the Move is valid under its installed Horizon.
// Success returns the original Move paired with that Horizon.
// Failure returns an error and no usable AuthorizedMove; callers must not
// dispatch the zero-value result. Failure/trap classification remains open.
type AuthorityChecker interface {
	Check(ctx context.Context, move Move) (AuthorizedMove, error)
}

// TODO: identify the owner of the Possibility namespace and define the
// checker's restricted lookup interface. The checker reads installed authority;
// it does not provision Possibilities or mint grants.
