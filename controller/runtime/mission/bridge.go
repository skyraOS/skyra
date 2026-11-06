package mission

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"skyra/controller/core/contract"
	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	journal "skyra/controller/runtime/humanroot"
)

func sum(b []byte) string { s := sha256.Sum256(b); return hex.EncodeToString(s[:]) }

// Endpoint is an ordinary, operator-registered callable contract. The adapter
// pins its exact binary/input type. It is not a new worker instruction.
type Endpoint struct {
	Ref      root.Capability
	Contract typed.Contract
}

func NewEndpoint(binary string, plan bool) (Endpoint, error) {
	schema := typed.LaunchSchema()
	if plan {
		schema = typed.PlanSchema()
	}
	c, err := typed.Identify(binary, schema)
	if err != nil {
		return Endpoint{}, err
	}
	b, _ := json.Marshal(c)
	return Endpoint{root.Capability("sha256:" + sum(b)), c}, nil
}

type Bridge struct {
	Catalogue *Catalogue // what a Launch may be granted
	Planner   *Catalogue // what Plan is granted: the viewer and the proposer; nil when the Mission never plans
	Root      *journal.Journal
	Plan      Endpoint
	Launch    Endpoint
	dir       string
}

// Open pins a Mission's root. Its ceiling is the Launch catalogue and, when
// it plans, the planner's contracts, so the root records them as Plan's grant.
func Open(dir, mission, human string, catalogue, planner *Catalogue, plan, launch Endpoint) (*Bridge, error) {
	if catalogue == nil {
		return nil, fmt.Errorf("Mission needs an operator catalogue")
	}
	if plan.Ref == launch.Ref || plan.Ref == "" || launch.Ref == "" {
		return nil, fmt.Errorf("Plan and Launch need distinct pinned contracts")
	}
	expectedPlan, err := NewEndpoint(plan.Contract.Binary, true)
	if err != nil || expectedPlan != plan {
		return nil, fmt.Errorf("Plan endpoint does not match its ordinary contract")
	}
	expectedLaunch, err := NewEndpoint(launch.Contract.Binary, false)
	if err != nil || expectedLaunch != launch {
		return nil, fmt.Errorf("Launch endpoint does not match its ordinary contract")
	}
	ceiling := catalogue.CapabilityRefs()
	if planner != nil {
		if err := plannerHolds(planner); err != nil {
			return nil, err
		}
		for _, ref := range planner.CapabilityRefs() {
			if slices.Contains(ceiling, ref) {
				return nil, fmt.Errorf("a planning contract is also a Launch capability")
			}
			ceiling = append(ceiling, ref)
		}
	}
	j, err := journal.Open(dir, root.Config{Mission: mission, Human: human, Entrypoints: []root.Capability{plan.Ref, launch.Ref}, Ceiling: ceiling})
	if err != nil {
		return nil, err
	}
	return &Bridge{Catalogue: catalogue, Planner: planner, Root: j, Plan: plan, Launch: launch, dir: dir}, nil
}

// StartPlan records the human's invocation, passing only explicit observations.
// Its payload already fits the ordinary Plan contract: adapters resolve local
// observation references through Dispatch.Passed, never a global hash lookup.
func (b *Bridge) StartPlan(requestID string, request root.ID, context []root.ID, tools []root.Capability) (root.Effect, json.RawMessage, error) {
	effect, err := b.Root.Invoke(requestID, b.Plan.Ref, append([]root.ID{request}, context...), tools)
	if err != nil || effect.Kind != root.Dispatch {
		return effect, nil, err
	}
	state, err := b.Root.State()
	if err != nil {
		return effect, nil, err
	}
	requestBinding, err := b.binding(state, effect.Passed[0])
	if err != nil {
		return effect, nil, err
	}
	contextBindings := []map[string]any{}
	for _, observation := range effect.Passed[1:] {
		binding, err := b.binding(state, observation)
		if err != nil {
			return effect, nil, err
		}
		contextBindings = append(contextBindings, binding)
	}
	payload, _ := json.Marshal(map[string]any{"plan": map[string]any{"request": requestBinding, "context": contextBindings}})
	return effect, payload, nil
}

// ReceivePlan accepts a runtime-verifiable text reply or validates a Spec
// before it enters the human's scope. Plan can cite only inputs it received.
func (b *Bridge) ReceivePlan(call root.ID, receipt string, read journal.ReturnReader) (root.Effect, error) {
	if read == nil {
		return root.Effect{}, fmt.Errorf("Plan return requires a runtime receipt reader")
	}
	// Read before Receive, which holds the root while it asks its reader.
	state, err := b.Root.State()
	if err != nil {
		return root.Effect{}, err
	}
	return b.Root.Receive(call, receipt, func(inv root.Invocation, receipt string) ([]json.RawMessage, error) {
		if inv.Target != b.Plan.Ref {
			return nil, fmt.Errorf("receipt does not belong to Plan")
		}
		values, err := read(inv, receipt)
		if err != nil {
			return nil, err
		}
		if len(values) != 1 {
			return nil, fmt.Errorf("Plan must return exactly one text reply or typed Spec observation")
		}
		var reply string
		if json.Unmarshal(values[0], &reply) == nil {
			if strings.TrimSpace(reply) == "" || len(reply) > 16*1024 {
				return nil, fmt.Errorf("Plan reply must be nonempty and at most 16 KiB")
			}
			return values, nil
		}
		spec, err := typed.DecodeSpec(values[0])
		if err != nil {
			return nil, err
		}
		if !root.MayPass(inv.Inputs, spec.Inputs) {
			return nil, fmt.Errorf("Plan cited an observation it was not passed")
		}
		// Plan may quote only what it was passed, and only what it says.
		if err = spec.CheckCites(func(id root.ID) (root.Observation, bool) {
			if !slices.Contains(inv.Inputs, id) {
				return root.Observation{}, false
			}
			return state.Lookup(id)
		}); err != nil {
			return nil, err
		}
		for _, ref := range spec.Capabilities {
			if _, ok := b.Catalogue.Typed.Lookup(ref); !ok {
				return nil, fmt.Errorf("Plan proposed an unknown capability")
			}
		}
		return values, nil
	})
}

type Launch struct {
	Dispatch root.Effect
	Resolved typed.Resolved
	Payload  json.RawMessage
}

// AuthorSpec is for an explicit human revision. Its source stays human; it is
// never relabelled as a Plan result. The same validation applies at Launch.
func (b *Bridge) AuthorSpec(requestID string, spec typed.Spec) (root.Effect, error) {
	state, err := b.Root.State()
	if err != nil {
		return root.Effect{}, err
	}
	resolved, err := b.Catalogue.Typed.Resolve(spec, state)
	if err != nil {
		return root.Effect{}, err
	}
	raw, _ := json.Marshal(resolved.Spec)
	return b.Root.Author(requestID, raw)
}

// StartLaunch commits the exact Spec reference and proposed grant together.
// A duplicate returns no Launch, so recovery cannot repeat an unknown effect.
func (b *Bridge) StartLaunch(requestID string, id root.ID) (*Launch, error) {
	state, err := b.Root.State()
	if err != nil {
		return nil, err
	}
	observation, ok := state.Lookup(id)
	if !ok {
		return nil, fmt.Errorf("Spec is not held by this Mission's human root")
	}
	source := observation.Source()
	if source.Human != state.Config().Human && source.Target != b.Plan.Ref {
		return nil, fmt.Errorf("Spec is neither a received Plan result nor a human revision")
	}
	spec, err := typed.DecodeSpec(observation.Value())
	if err != nil {
		return nil, err
	}
	if _, err = b.Catalogue.Typed.Resolve(spec, state); err != nil {
		return nil, err
	}
	ids := append([]root.ID{id}, spec.Inputs...)
	effect, err := b.Root.Invoke(requestID, b.Launch.Ref, ids, spec.Capabilities)
	if err != nil {
		return nil, err
	}
	if effect.Kind == root.Duplicate {
		return nil, nil
	}
	state, err = b.Root.State()
	if err != nil {
		return nil, err
	}
	resolved, err := b.Catalogue.Typed.ValidateHandoff(effect, state)
	if err != nil {
		return nil, err
	}
	binding, err := b.binding(state, observation)
	if err != nil {
		return nil, err
	}
	payload, _ := json.Marshal(map[string]any{"launch": map[string]any{"spec": binding}})
	return &Launch{effect, resolved, payload}, nil
}

// PrepareLaunch consumes only a fresh committed dispatch from StartLaunch.
// It never fetches extra observations from other Missions or the transcript.
func (b *Bridge) PrepareLaunch(dir, actor string, launch *Launch) (string, json.RawMessage, error) {
	if launch == nil || launch.Dispatch.Invocation == nil || launch.Dispatch.Invocation.Target != b.Launch.Ref {
		return "", nil, fmt.Errorf("missing committed Launch invocation")
	}
	state, err := b.Root.State()
	if err != nil {
		return "", nil, err
	}
	resolved, err := b.Catalogue.Typed.ValidateHandoff(launch.Dispatch, state)
	if err != nil {
		return "", nil, err
	}
	registry, payload, err := b.Catalogue.Prepare(dir, actor, resolved)
	if err != nil {
		return "", nil, err
	}
	var request map[string]map[string]any
	if err = json.Unmarshal(payload, &request); err != nil {
		return "", nil, err
	}
	imports := []map[string]any{}
	for _, observation := range launch.Dispatch.Passed {
		binding, err := b.binding(state, observation)
		if err != nil {
			return "", nil, err
		}
		imports = append(imports, binding)
	}
	request["tree"]["imports"] = imports
	payload, err = json.Marshal(request)
	return registry, payload, err
}

func (b *Bridge) binding(state root.State, observation root.Observation) (map[string]any, error) {
	source := observation.Source()
	var producer typed.Contract
	var err error
	switch {
	case source.Human == state.Config().Human:
		producer, err = typed.Identify(b.Plan.Contract.Binary, &contract.Schema{Title: "human-observation/v1", Type: "any"})
	case source.Target == b.Plan.Ref:
		producer = b.Plan.Contract
	case source.Target == b.Launch.Ref:
		producer = b.Launch.Contract
	default:
		descriptor, ok := b.Catalogue.Typed.Lookup(source.Target)
		if !ok {
			return nil, fmt.Errorf("observation has an unknown producer contract")
		}
		producer = descriptor.Contract
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"ref": observation.ID(), "binary": producer.Binary, "contract": producer.Shape, "value": observation.Value(), "source": source}, nil
}
