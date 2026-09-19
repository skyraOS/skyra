//go:build ignore

// Package messagebroker contains proposed contracts, not a runtime implementation.
// Referenced request, map, record, and profile types await their exact contracts.
// Profile interfaces are not empty interfaces: each exposes permitted operations.
// Surface and Component are defined in the sibling surface and TU proposals;
// package wiring remains provisional.
package messagebroker

import "github.com/google/uuid"

// StateOwner is the single owner behind the MessageBroker profiles. It serializes
// state requests; components receive restricted adapters, not general Read/Write.
// MessageBroker queues whole TU Requests with their bound routing context and
// carries acknowledgments. StateOwner validates and commits one stage at a time;
// there is no staging area or transaction spanning Interpreter and Supervisor.
type StateOwner struct {
	requests chan stateRequest
	system   systemState
	user     userState
}

type systemState struct {
	revision uint64
	profiles SystemProfiles

	continuations ContinuationMap // The map of Moves; no separate Move map.
	possibilities PossibilityMap
	custody       map[uuid.UUID]Component
	workers       map[uuid.UUID]WorkerRecord

	interpretation mailbox
	supervisor     mailbox
}

// userState holds user-registered Surfaces, such as Codex, and UserSurfaces.
// A Surface's native state remains owned by that Surface.
type userState struct {
	surfaces     map[uuid.UUID]*Surface
	userSurfaces UserSurfaces
}

// Mailbox entries reference Moves in the continuation map in FIFO order.
type mailbox struct {
	queued []uuid.UUID
}

// SystemProfiles names the narrow interfaces injected by boot when each
// component or system Surface is constructed. Profiles are fixed at boot.
// Each implementation uses a distinct private adapter type, not the full broker
// disguised behind a narrower interface. All adapters use the same state owner.
type SystemProfiles struct {
	Interpretation InterpretationState
	Supervisor     SupervisorState
	Transition     TransitionUnit
	TrapSurface    TrapSurfaceState
}

type UserSurfaces struct {
	SysCallSurface SysCallSurfaceState
}

// Kernel holds the two system entry Surfaces; profiles govern their state access.
type Kernel struct {
	TrapSurface    *Surface
	SysCallSurface *Surface
}
