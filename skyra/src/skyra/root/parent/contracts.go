// Package parent is a deterministic executable peer. It sees only serialized Moves.
package parent

// Continuation is opaque to the runtime and returned to the peer on child resume.
type Continuation struct {
	Prefix string `json:"prefix"`
}
