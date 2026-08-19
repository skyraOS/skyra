// Package ids holds the identity types.
//
// Distinct named types rather than a shared ID, so the compiler refuses a host where
// an engram belongs.
//
// Global uniqueness is load-bearing and easy to misread: provenance resolves across
// trees, and the credit walk is bounded by CITES BEING EMPTY at a crossing, never by
// ID scope — so nobody should reach for scoped IDs to "fix" a boundary that is
// already held.
package ids

type (
	// EngramID names DURABLE BEHAVIOUR. What it IS; the NAME is the parent's map key.
	EngramID string

	// PossibilityID names ONE OPENING of an engram.
	//
	// RENAMED FROM SpanID. `span` is borrowed from tracing and names a DURATION; the
	// thing identified is not a duration. A branch is the possibility with no parent,
	// and a loop iteration is a possibility, so ONE TYPE IS USED AT BOTH POSITIONS —
	// there is no separate InvocationID, because an invocation IS a possibility and
	// already has a name.
	PossibilityID string

	// BranchID names ONE ATTEMPT within a possibility.
	BranchID string

	ObservationID string

	// EntityID names WHO interprets. An identity and nothing else — every declaration
	// dissolved, because a declaration is made by whoever REGISTERED the entity, not
	// by the entity.
	EntityID string

	// HostID names WHERE it lands. Physical hardware, including devices with no OS.
	HostID string

	// RuntimeID names ONE AUTHORITY DOMAIN — one per device that can run one.
	//
	// A runtime is not a host. Held by the kernel, not carried per opening.
	RuntimeID string
)

// Address is WHERE something the runtime does not own can be reached.
//
// It is not an entity: for an entity the runtime INVOKED the behaviour, so provenance
// is substrate-known; for an address it knows only WHERE IT SENT, never WHAT PRODUCED.
type Address string
