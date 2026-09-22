// Package acquire assembles discover -> test -> admit as one Possibility graph.
package acquire

import (
	"fmt"

	c "skyra-v1/src/skyra/skyra-dna-v10/causal"
	"skyra-v1/src/skyra/skyra-dna-v10/engine"
)

type ReadRegistry interface {
	Known(c.Endpoint) c.NoveltyVerdict
}
type Seeder interface {
	Seed(engine.Resource)
	AttachProbe(c.Endpoint, engine.Resource)
}
type Trace interface{ Printf(string, ...any) }

type FakeNetwork struct {
	Names                                                     []c.Name
	Endpoint                                                  c.Endpoint
	Reply                                                     map[string]any
	Searches, Resolves                                        int
	ReachCalls, HandshakeCalls, ChallengeCalls, DescribeCalls int
	UnsafeCalls                                               int
}

type Kit struct {
	Registry ReadRegistry
	Fake     *FakeNetwork
	Trace    Trace
}

const Purpose = "weather in San Francisco"

var (
	rIntent     = key("intent")
	rNames      = key("names")
	rProposal   = key("proposal")
	rReach      = key("reach")
	rHandshake  = key("handshake")
	rAuth       = key("auth")
	rFitness    = key("fitness")
	rContract   = key("contract")
	rAdmission  = key("registration-request")
	rCapability = key("capability")
)

func key(v string) c.RegKey { return c.RegKey{Name: c.Name{Value: v}} }

// Install seeds only already-admitted orchestration/resolver entities and attaches
// an unreachable probe implementation. It performs no Mint.
func (k Kit) Install(dst Seeder) {
	dst.Seed(resource("search", c.Entity, k.search))
	dst.Seed(resource("resolve", c.Entity, k.resolve))
	dst.Seed(resource("test.join", c.Host, k.join))
	dst.Seed(resource("admit.package", c.Host, k.packageAdmission))
	dst.AttachProbe(k.Fake.Endpoint, resource(k.Fake.Endpoint.Key().Value, c.Entity, k.endpoint))
}
func resource(name string, block c.Block, run engine.Runner) engine.Resource {
	return engine.FunctionResource{ResourceName: c.Name{Value: name}, Trust: block, Runner: run}
}

func NewPossibility(id string, intent c.Observation) *c.Possibility {
	p := &c.Possibility{ID: c.PossibilityID{Value: id}, Registers: map[c.RegKey]c.Register{}, InFlight: map[c.MoveID]c.Pending{}, Return: rCapability,
		Horizon: c.Horizon{Caveats: []c.Caveat{{Constraint: "read-only; no ambient ask_permission; no mint"}}}}
	p.Registers[rIntent] = c.Register{Value: &intent}
	for _, r := range []c.RegKey{rNames, rProposal, rReach, rHandshake, rAuth, rFitness, rContract, rAdmission, rCapability} {
		p.Registers[r] = c.Register{}
	}
	p.Moves = []c.Move{
		move("search", "search", []c.RegKey{rIntent}, rNames, nil, nil),
		move("resolve", "resolve", []c.RegKey{rNames}, rProposal, nil, c.ResolveHints{Service: "https", Triple: c.Triple{Domain: "inet", Type: "stream", Protocol: "tcp"}}),
		move("reach", "candidate.endpoint", []c.RegKey{rProposal}, rReach, poke(), "reach"),
		move("handshake", "candidate.endpoint", []c.RegKey{rProposal, rReach}, rHandshake, poke(), "handshake"),
		move("challenge", "candidate.endpoint", []c.RegKey{rProposal, rHandshake}, rAuth, poke(), "challenge"),
		move("describe", "candidate.endpoint", []c.RegKey{rProposal, rHandshake}, rFitness, poke(), "describe"),
		move("test.join", "test.join", []c.RegKey{rProposal, rReach, rHandshake, rAuth, rFitness, rIntent}, rContract, nil, nil),
		move("admit.package", "admit.package", []c.RegKey{rProposal, rContract}, rAdmission, nil, nil),
		move("admit", "candidate.endpoint", []c.RegKey{rAdmission}, rCapability, &c.Access{Kind: c.RegisterGrant, EndpointFrom: rAdmission, Purpose: Purpose}, "admit"),
	}
	return p
}
func poke() *c.Access { return &c.Access{Kind: c.PokeGrant, EndpointFrom: rProposal, Purpose: Purpose} }
func move(id, target string, reads []c.RegKey, writes c.RegKey, access *c.Access, payload any) c.Move {
	return c.Move{ID: c.MoveID{Value: id}, Target: c.Name{Value: target}, Reads: reads, Writes: writes, Contract: c.Contract{Name: c.Name{Value: writes.Name.Value}}, Site: c.Local{}, Access: access, Payload: payload}
}

func (k Kit) search(in []c.Observation, _ c.Move) (c.Observation, c.Outcome) {
	k.Fake.Searches++
	names := append([]c.Name(nil), k.Fake.Names...)
	k.Trace.Printf("SEARCH input=%q -> names=%v\n", string(in[0].Content.Bytes), nameValues(names))
	return c.Observation{Value: names}, c.Success
}
func (k Kit) resolve(in []c.Observation, mv c.Move) (c.Observation, c.Outcome) {
	k.Fake.Resolves++
	names, ok := in[0].Value.([]c.Name)
	if !ok {
		return c.Observation{}, c.Failure
	}
	hints := mv.Payload.(c.ResolveHints)
	cand := c.Candidate{Name: names[0], Endpoint: k.Fake.Endpoint, Novelty: k.Registry.Known(k.Fake.Endpoint)}
	proposal := c.RegistrationProposal{Candidates: []c.Candidate{cand}}
	k.Trace.Printf("RESOLVE input={names:%v hints:{service:%s triple:%s/%s/%s}} -> discover RegistrationProposal{candidates:[{name:%s endpoint:%s novelty:{known:%t history:%q}}]}\n", nameValues(names), hints.Service, hints.Triple.Domain, hints.Triple.Type, hints.Triple.Protocol, cand.Name.Value, cand.Endpoint.Address, cand.Novelty.Known, cand.Novelty.History.Note)
	return c.Observation{Value: proposal}, c.Success
}
func (k Kit) endpoint(_ []c.Observation, mv c.Move) (c.Observation, c.Outcome) {
	switch mv.Payload {
	case "reach":
		k.Fake.ReachCalls++
		v := c.Reachability{Up: true, Why: "mock connect answered"}
		k.Trace.Printf("HANDOFF discover Candidate{endpoint:%s} -> test input\nPROBE reach input=%s -> Reachability{up:%t why:%q}\n", k.Fake.Endpoint.Address, k.Fake.Endpoint.Address, v.Up, v.Why)
		return c.Observation{Value: v}, c.Success
	case "handshake":
		k.Fake.HandshakeCalls++
		v := c.HandshakeFragment{Transport: c.TransportContract{Triple: k.Fake.Endpoint.Triple, Port: 443, TLS: true, ALPN: "h2"}, Identity: c.IdentityContract{CertHost: "weather.example"}}
		k.Trace.Printf("PROBE handshake input=%s -> {transport:{port:443 tls:true alpn:h2} identity:{cert_host:%q}}\n", k.Fake.Endpoint.Address, v.Identity.CertHost)
		return c.Observation{Value: v}, c.Success
	case "challenge":
		k.Fake.ChallengeCalls++
		v := c.AuthContract{Method: "bearer", CredType: "oauth2-token", Verified: false}
		k.Trace.Printf("PROBE challenge input=safe-unauthenticated-GET -> Auth{method:%s cred_type:%s verified:%t}\n", v.Method, v.CredType, v.Verified)
		return c.Observation{Value: v}, c.Success
	case "describe":
		k.Fake.DescribeCalls++
		v := c.FitnessContract{Operation: "GET /forecast", Request: c.Contract{Name: c.Name{Value: "ForecastRequest"}, RequiredFields: []string{"city"}}, Response: c.Contract{Name: c.Name{Value: "ForecastReply"}, RequiredFields: []string{"forecast", "temperature_c"}}}
		k.Trace.Printf("PROBE describe input=safe-OPTIONS -> Fitness{operation:%q request:%s response:%s required:%v}\n", v.Operation, v.Request.Name.Value, v.Response.Name.Value, v.Response.RequiredFields)
		return c.Observation{Value: v}, c.Success
	case "use":
		return c.Observation{Value: k.Fake.Reply}, c.Success
	default:
		k.Fake.UnsafeCalls++
		return c.Observation{}, c.Failure
	}
}
func (k Kit) join(in []c.Observation, _ c.Move) (c.Observation, c.Outcome) {
	h := in[2].Value.(c.HandshakeFragment)
	a := in[3].Value.(c.AuthContract)
	f := in[4].Value.(c.FitnessContract)
	intent := string(in[5].Content.Bytes)
	if intent != Purpose || f.Operation != "GET /forecast" {
		return c.Observation{}, c.Failure
	}
	k.Trace.Printf("FITNESS MATCH intent=%q operation=%q -> selected=true\n", intent, f.Operation)
	contract := c.ConnectionContract{Transport: h.Transport, Identity: h.Identity, Auth: a, Fitness: f, Success: c.SuccessCriterion{Says: "HTTP 200 with forecast and temperature_c"}, Partial: !a.Verified}
	k.Trace.Printf("TEST JOIN fragments -> ConnectionContract{transport:{port:%d tls:%t alpn:%s} identity:{cert_host:%q} auth:{method:%s cred_type:%s verified:%t} fitness:{operation:%q response:%s required:%v} partial:%t}\nHANDOFF ConnectionContract -> admit input={endpoint:%s operation:%q}\n", h.Transport.Port, h.Transport.TLS, h.Transport.ALPN, h.Identity.CertHost, a.Method, a.CredType, a.Verified, f.Operation, f.Response.Name.Value, f.Response.RequiredFields, contract.Partial, k.Fake.Endpoint.Address, f.Operation)
	return c.Observation{Value: contract}, c.Success
}
func (k Kit) packageAdmission(in []c.Observation, _ c.Move) (c.Observation, c.Outcome) {
	proposal, ok := in[0].Value.(c.RegistrationProposal)
	if !ok || len(proposal.Candidates) == 0 {
		return c.Observation{}, c.Failure
	}
	contract, ok := in[1].Value.(c.ConnectionContract)
	if !ok {
		return c.Observation{}, c.Failure
	}
	req := c.RegistrationRequest{Endpoint: proposal.Candidates[0].Endpoint, Contract: contract, Block: c.Entity}
	k.Trace.Printf("ADMIT PACKAGE input={endpoint:%s contract.operation:%q} -> RegistrationRequest{endpoint:%s block:%s contract.operation:%q}\n", req.Endpoint.Address, req.Contract.Fitness.Operation, req.Endpoint.Address, req.Block, req.Contract.Fitness.Operation)
	return c.Observation{Value: req}, c.Success
}
func nameValues(ns []c.Name) []string {
	out := make([]string, len(ns))
	for i, n := range ns {
		out[i] = n.Value
	}
	return out
}

func UsePossibility(id string, cap c.Capability) *c.Possibility {
	out := key("use.out")
	p := &c.Possibility{ID: c.PossibilityID{Value: id}, Registers: map[c.RegKey]c.Register{out: {}}, InFlight: map[c.MoveID]c.Pending{}, Return: out}
	p.Moves = []c.Move{{ID: c.MoveID{Value: "use"}, Target: cap.Endpoint.Key(), Writes: out, Contract: cap.Contract.Fitness.Response, Site: c.Local{}, Payload: "use"}}
	return p
}
func FormatCapability(cap c.Capability) string {
	return fmt.Sprintf("Capability{name:%s endpoint:%s block:%s contract.operation:%q}", cap.Name.Value, cap.Endpoint.Address, cap.Block, cap.Contract.Fitness.Operation)
}
