package mission

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	"skyra/controller/core/record"
	"skyra/controller/runtime/boot"
)

// Plan holds ordinary leaf contracts, which make it a delegate: Viewer
// returns the contracts the Mission's human holds for Launch, and Proposer
// turns a spec naming tools from that view into a typed Spec.
const (
	Viewer   = "view-plan-contracts"
	Proposer = "propose-spec"
)

// Navigation is how Plan may look around the workspace before it proposes:
// pwd and cd return a directory observation, and cd, ls, read and find take one.
var Navigation = []string{"plan.pwd", "plan.cd", "plan.ls", "plan.find", "plan.read"}

// steps are what Plan's goal asks of each contract it holds, one line each,
// so its delegate root can cut one for the child it passes that contract.
var steps = map[string]string{
	"plan.pwd":  "Find where you are in the workspace: the directory every other step starts from.",
	"plan.cd":   "Move into a directory inside one you have found.",
	"plan.ls":   "List a directory you have found.",
	"plan.find": "Search under a directory you have found, by name or by text; matches are relative to the workspace, so read follows one from where you started.",
	"plan.read": "Read a file in a directory you have found.",
	Viewer:      "View the capabilities you may attach to this plan: those this mission's human holds, each with its tool, its input and the scope it reaches.",
	Proposer:    "Propose the spec from the viewed capabilities, passing them as its catalogue. Its goal is the complete task for Launch, which sees only that goal, the inputs it cites and the tools it proposes. Build the goal from segments: quote the request and conversation verbatim, by their ids below, wherever they say what to do, and add your own text only where they do not; every word you add lowers the spec's cite coverage. If the current spec is passed, keep what you want of it by quoting it by its id: only the words a human wrote or cited in it count as quoted, so add again as your own text any you wrote. Its note is your reply to the human's message: answer them, say what you changed, and ask what you need to know. If they have not yet said what Launch should do, propose an empty goal and no tools, and ask. Its inputs are the ids, below, of the passed observations Launch needs. Its capabilities are the tools, by target, from the viewed capabilities that Launch needs. Its profile is the constants Launch's tree runs under, all four required: act, the most tools one node may hold and invoke itself before it must delegate; turns, the most turns each node takes; depth, how deep a child may be; fanout, the most instructions in one call, at most 32; budget, the most turns the whole tree takes, every node's together. Keep the current spec's profile unless the human asks to change it; with no current spec, use act 1, turns 8, depth 3, fanout 32, budget 64.",
}

// PlanProfile is the constants Plan's own tree runs under. Its five
// navigation contracts fit one act node, an explorer that looks around in its
// own turns; viewing and proposing fit another; the root, holding all seven,
// delegates to them and goes no deeper. The budget bounds a whole answer.
var PlanProfile = typed.Profile{Act: 5, Turns: 24, Depth: 1, FanOut: 8, Budget: 60}

// ViewFile is where, in the viewer's own state, the host leaves the view.
const ViewFile = "contracts.json"

// plannerHolds requires the viewer and the proposer, and allows navigation.
func plannerHolds(planner *Catalogue) error {
	targets := planner.targets()
	if !slices.Contains(targets, Viewer) || !slices.Contains(targets, Proposer) {
		return fmt.Errorf("a planner holds the %s and %s contracts", Viewer, Proposer)
	}
	for _, t := range targets {
		if _, ok := steps[t]; !ok {
			return fmt.Errorf("%s is not a contract Plan may hold", t)
		}
	}
	return nil
}

func (c *Catalogue) targets() []string {
	var out []string
	for _, d := range c.Typed.Entries() {
		if !slices.Contains(out, d.Target) {
			out = append(out, d.Target)
		}
	}
	return out
}

// PlanGrant is what the human grants Plan: the planner's contracts.
func (b *Bridge) PlanGrant() []root.Capability {
	if b.Planner == nil {
		return []root.Capability{}
	}
	return b.Planner.CapabilityRefs()
}

// PlanRoot and LaunchRoot are where a call's kernel run lives, once it is
// claimed: the tree a viewer replays.
func (b *Bridge) PlanRoot(call root.ID) string {
	return filepath.Join(b.dir, "plans", string(call), ".skyra")
}

func (b *Bridge) LaunchRoot(call root.ID) string {
	return filepath.Join(b.dir, "launches", string(call), ".skyra")
}

// ExecutePlan runs a committed Plan dispatch from StartPlan through the
// existing kernel, once: a delegate tree holding exactly the contracts the
// root granted Plan, passed the request and context observations, asked to
// view the human's contracts and propose one Spec from them. Its return
// enters the human's scope only through ReceivePlan.
func (b *Bridge) ExecutePlan(ctx context.Context, kernel, actor string, plan root.Effect) (root.Effect, error) {
	if plan.Kind != root.Dispatch || plan.Invocation == nil || plan.Invocation.Target != b.Plan.Ref {
		return root.Effect{}, fmt.Errorf("missing committed Plan invocation")
	}
	call := plan.Invocation.ID
	if len(call) != 64 {
		return root.Effect{}, fmt.Errorf("invalid Plan call identity")
	}
	if b.Planner == nil {
		return root.Effect{}, fmt.Errorf("this Mission has no planner")
	}
	granted := slices.Clone(plan.Invocation.Grant)
	slices.Sort(granted)
	if held := b.PlanGrant(); !slices.Equal(granted, held) {
		return root.Effect{}, fmt.Errorf("Plan runs only with the planner's contracts, as its grant")
	}
	kernelBytes, err := os.ReadFile(kernel)
	if err != nil {
		return root.Effect{}, err
	}
	if "sha256:"+record.Sum(kernelBytes) != b.Plan.Contract.Binary {
		return root.Effect{}, fmt.Errorf("kernel binary differs from the pinned Plan endpoint")
	}
	state, err := b.Root.State()
	if err != nil {
		return root.Effect{}, err
	}
	dir := filepath.Join(b.dir, "plans", string(call))
	claim, err := claimExecution(dir)
	if err != nil {
		return root.Effect{}, err
	}
	goal := b.planGoal(plan.Passed)
	reg, payload, err := b.Planner.Prepare(filepath.Join(dir, "prepared"), actor, typed.Resolved{Spec: typed.Spec{Version: 1, Goal: goal}, Contracts: b.Planner.Typed.Entries(), Kind: "delegate", Profile: PlanProfile})
	if err != nil {
		claim.Close()
		return root.Effect{}, err
	}
	// The viewer's own state, before the tree boots: what it will return,
	// the capabilities the Mission's human holds that a plan may attach.
	view, _ := json.Marshal(b.Catalogue.Typed.ViewOf(state.Config().Ceiling))
	if err = os.MkdirAll(filepath.Join(dir, ".skyra", Viewer), 0700); err == nil {
		err = os.WriteFile(filepath.Join(dir, ".skyra", Viewer, ViewFile), view, 0600)
	}
	if err != nil {
		claim.Close()
		return root.Effect{}, err
	}
	var request map[string]map[string]any
	if err = json.Unmarshal(payload, &request); err != nil {
		claim.Close()
		return root.Effect{}, err
	}
	imports := []map[string]any{}
	for _, observation := range plan.Passed {
		binding, err := b.binding(state, observation)
		if err != nil {
			claim.Close()
			return root.Effect{}, err
		}
		imports = append(imports, binding)
	}
	request["tree"]["imports"] = imports
	payload, _ = json.Marshal(request)
	binding, _ := json.Marshal(map[string]any{"call": call, "payload": json.RawMessage(payload), "view": json.RawMessage(view)})
	if err = seal(claim, dir, binding); err != nil {
		return root.Effect{}, err
	}
	runRoot := filepath.Join(dir, ".skyra")
	output, runErr := exec.CommandContext(ctx, kernel, "boot", "-root", runRoot, "-registry", reg, "-to", "apprentice", "-payload", string(payload)).CombinedOutput()
	head, err := boot.Head(runRoot)
	if err != nil {
		return root.Effect{}, fmt.Errorf("Plan outcome unavailable: %v; kernel: %v: %s", err, runErr, output)
	}
	return b.ReceivePlan(call, head, func(inv root.Invocation, receipt string) ([]json.RawMessage, error) {
		if inv.ID != call || receipt != head {
			return nil, fmt.Errorf("Plan receipt belongs to another invocation")
		}
		final, err := settled(runRoot, receipt, dir, binding)
		if err != nil {
			return nil, err
		}
		if final.State != record.Return {
			return nil, fmt.Errorf("Plan did not return a Spec: %s", final.Payload)
		}
		var answer struct {
			Returned []json.RawMessage `json:"returned"`
		}
		if err := json.Unmarshal(final.Payload, &answer); err != nil || len(answer.Returned) != 1 {
			return nil, fmt.Errorf("Plan must return exactly what %s gave back: %s", Proposer, final.Payload)
		}
		return answer.Returned, nil
	})
}

// planGoal is what the Plan tree is asked, by line, so its delegate root can
// cut each step for a child: looking around the workspace if it may, viewing
// the human's contracts, then proposing the spec from that view. The request
// and context follow as passed observations.
func (b *Bridge) planGoal(passed []root.Observation) string {
	var g strings.Builder
	g.WriteString("You are Plan, shaping this mission's spec with the human. Their message is the request passed below: answer it, and return the spec as it now stands.\n")
	held := b.Planner.targets()
	for _, t := range append(slices.Clone(Navigation), Viewer, Proposer) {
		if slices.Contains(held, t) {
			g.WriteString(steps[t] + "\n")
		}
	}
	g.WriteString("Inputs you may cite:\n")
	for _, o := range passed {
		fmt.Fprintf(&g, "  %s\n", o.ID())
	}
	return strings.TrimRight(g.String(), "\n")
}
