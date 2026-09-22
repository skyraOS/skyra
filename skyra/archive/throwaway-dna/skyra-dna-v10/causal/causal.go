// Package causal contains values that cross the v10 acquisition boundaries.
package causal

import "fmt"

type Outcome int

const (
	NotYet Outcome = iota
	Success
	Failure
	Fault
)

func (o Outcome) String() string { return [...]string{"NOTYET", "SUCCESS", "FAILURE", "FAULT"}[o] }

type Name struct{ Value string }
type Triple struct{ Domain, Type, Protocol string }
type Endpoint struct {
	Namespace, Address string
	Triple             Triple
}

func (e Endpoint) Key() Name {
	return Name{Value: fmt.Sprintf("endpoint:%s:%s/%s/%s", e.Namespace, e.Address, e.Triple.Protocol, e.Triple.Type)}
}

// There are exactly two trust blocks. Privilege is a trap, not a third block.
type Block int

const (
	Entity Block = iota
	Host
)

func (b Block) String() string {
	if b == Host {
		return "HOST/BELIEVED"
	}
	return "ENTITY/STAMPED+CHECKED"
}

type Capability struct {
	Name     Name
	Endpoint Endpoint
	Contract ConnectionContract
	Block    Block
}

// RegistrationRequest is admit's packaged endpoint+contract. It is data, not
// authority; only the kernel can turn it into a registered Capability.
type RegistrationRequest struct {
	Endpoint Endpoint
	Contract ConnectionContract
	Block    Block
}
type Candidate struct {
	Name     Name
	Endpoint Endpoint
	Novelty  NoveltyVerdict
}
type RegistrationProposal struct{ Candidates []Candidate }
type NoveltyVerdict struct {
	Known   bool
	History History
}
type History struct{ Note string }
type ResolveHints struct {
	Service string
	Triple  Triple
}
type ResolveRequest struct {
	Names []Name
	Hints ResolveHints
}

type ConnectionContract struct {
	Transport TransportContract
	Identity  IdentityContract
	Auth      AuthContract
	Fitness   FitnessContract
	Success   SuccessCriterion
	Partial   bool
}
type TransportContract struct {
	Triple Triple
	Port   int
	TLS    bool
	ALPN   string
}
type IdentityContract struct{ CertHost string }
type AuthContract struct {
	Method, CredType string
	Verified         bool
}
type FitnessContract struct {
	Operation         string
	Request, Response Contract
}
type SuccessCriterion struct{ Says string }
type Reachability struct {
	Up  bool
	Why string
}
type HandshakeFragment struct {
	Transport TransportContract
	Identity  IdentityContract
}

type GrantKind int

const (
	PokeGrant GrantKind = iota
	RegisterGrant
)

func (k GrantKind) String() string {
	if k == RegisterGrant {
		return "Register"
	}
	return "Poke"
}

// Scope makes permission resource-specific; it replaces the proposed AskPermission bool.
type Scope struct {
	Endpoint Endpoint
	Purpose  string
	Action   GrantKind
}
type Grant struct {
	Kind    GrantKind
	Subject Endpoint
	Scope   Scope
	Minted  *Capability
}

// FaultValue is the typed payload the bare FAULT outcome lacked.
type FaultValue struct {
	Target       Name
	Endpoint     Endpoint
	Kind         GrantKind
	Scope        Scope
	Registration *RegistrationRequest
}

type PossibilityID struct{ Value string }
type MoveID struct{ Value string }
type Tag struct{ Iter int }
type RegKey struct {
	Name Name
	Tag  Tag
}
type MoveState int

const (
	Waiting MoveState = iota
	Ready
	Issued
	Retired
)

// Access describes a dynamic endpoint reference.
type Access struct {
	Kind         GrantKind
	EndpointFrom RegKey
	Purpose      string
}
type Move struct {
	ID       MoveID
	Target   Name
	Reads    []RegKey
	Writes   RegKey
	Contract Contract
	State    MoveState
	Site     Site
	Access   *Access
	Payload  any
}
type Register struct {
	Contract Contract
	Value    *Observation
}
type Observation struct {
	Value   any
	Content Content
	Segment Segment
}
type Content struct{ Bytes []byte }
type Segment struct{ By Name }
type Contract struct {
	Name           Name
	RequiredFields []string
}

// Horizon carries only already-issued scoped authority. THINK has none initially.
type Horizon struct {
	Caveats []Caveat
	Grants  []Grant
}
type Caveat struct{ Constraint string }
type Pending struct {
	MoveIndex int
	GrantReg  RegKey
	Fault     FaultValue
}
type Possibility struct {
	ID        PossibilityID
	Moves     []Move
	Registers map[RegKey]Register
	InFlight  map[MoveID]Pending
	Retired   int
	Horizon   Horizon
	Return    RegKey
	Result    *Observation
}

type Site interface{ site() }
type Local struct{}

func (Local) site() {}

type Cut struct {
	ID     MoveID
	Target Name
}

func (Cut) site() {}
