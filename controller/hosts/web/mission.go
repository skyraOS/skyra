package main

// Missions: each conversation is a Mission whose human holds a root above its
// trees (core/humanroot). A spec the human saves is a human-authored Spec
// observation; Plan proposes one from the conversation; Launch is the human
// passing exactly one held Spec, with its capabilities, to a tree. The bridge
// (runtime/mission) records each decision before anything runs.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	"skyra/controller/core/record"
	"skyra/controller/runtime/mission"
)

// human is the principal at every Mission's root: the person at this host.
const human = "human"

// missionSetup is the operator's catalogue, pinned once per host: the tools
// a Launch may be granted, what Plan is granted (its way around the
// workspace, the viewer and the proposer), and the kernel endpoints.
type missionSetup struct {
	catalogue, planner *mission.Catalogue
	plan, launch       mission.Endpoint
	epoch              string // names the pins: a rebuilt tool or kernel is a new epoch
	refs               map[string][]root.Capability
	names              map[root.Capability]string
}

type missions struct {
	mu      sync.Mutex
	setup   *missionSetup
	bridges map[string]*mission.Bridge // by conversation and epoch
}

// missionSetup loads the catalogue on first use. Loading hashes every
// registered binary, so it happens once, outside h.mu.
func (h *webhost) missionSetup() (*missionSetup, error) {
	h.missions.mu.Lock()
	defer h.missions.mu.Unlock()
	if h.missions.setup != nil {
		return h.missions.setup, nil
	}
	if h.trees == nil {
		return nil, errors.New("this webhost was started without trees")
	}
	if len(h.trees.targets) < len(treeTargets) {
		fmt.Fprintf(h.log, "webhost: a Launch may hold only %v: the rest of %v are not programs in this registry\n", h.trees.targets, treeTargets)
	}
	catalogue, err := mission.Load(h.trees.registry, h.trees.targets)
	if err != nil {
		return nil, err
	}
	planner, err := mission.Load(h.trees.registry, append([]string{mission.Viewer, mission.Proposer, mission.Replier}, mission.Navigation...))
	if err != nil {
		return nil, err
	}
	kernel, err := os.ReadFile(h.skyra)
	if err != nil {
		return nil, err
	}
	plan, err := mission.NewEndpoint("sha256:"+record.Sum(kernel), true)
	if err != nil {
		return nil, err
	}
	launch, err := mission.NewEndpoint(plan.Contract.Binary, false)
	if err != nil {
		return nil, err
	}
	s := &missionSetup{catalogue: catalogue, planner: planner, plan: plan, launch: launch, refs: map[string][]root.Capability{}, names: map[root.Capability]string{}}
	for _, d := range catalogue.Typed.Entries() {
		s.refs[d.Target] = append(s.refs[d.Target], d.Ref)
		s.names[d.Ref] = d.Target
	}
	pins, _ := json.Marshal([]any{plan.Ref, launch.Ref, catalogue.CapabilityRefs(), planner.CapabilityRefs()})
	sum := sha256.Sum256(pins)
	s.epoch = hex.EncodeToString(sum[:8])
	h.missions.setup = s
	return s, nil
}

// bridge is a conversation's Mission at the current epoch. Its journal lives
// beside the conversation, never where a worker's sandbox reaches.
func (h *webhost) bridge(s *missionSetup, id, home string) (*mission.Bridge, error) {
	h.missions.mu.Lock()
	defer h.missions.mu.Unlock()
	key := id + "/" + s.epoch
	if b := h.missions.bridges[key]; b != nil {
		return b, nil
	}
	b, err := mission.Open(filepath.Join(home, "mission", s.epoch), id, human, s.catalogue, s.planner, s.plan, s.launch)
	if err != nil {
		return nil, err
	}
	if h.missions.bridges == nil {
		h.missions.bridges = map[string]*mission.Bridge{}
	}
	h.missions.bridges[key] = b
	return b, nil
}

// capabilities are a tool selection's catalogue references, in catalogue order.
func (s *missionSetup) capabilities(tools []string) ([]root.Capability, error) {
	out := []root.Capability{}
	for _, t := range tools {
		refs, ok := s.refs[t]
		if !ok {
			return nil, fmt.Errorf("%s is not a tool a Mission may grant here", t)
		}
		out = append(out, refs...)
	}
	slices.Sort(out)
	return slices.Compact(out), nil
}

func (s *missionSetup) tools(refs []root.Capability) []string {
	out := []string{}
	for _, ref := range refs {
		if name := s.names[ref]; name != "" && !slices.Contains(out, name) {
			out = append(out, name)
		}
	}
	slices.Sort(out)
	return out
}

// Called under h.mu, and releases it on every path. The request ledger gives
// reconnect deduplication; the human root refuses a request id reused for
// different work, so a retry never decides twice.
func (h *webhost) handleMission(m inbound, reject func(string, string)) {
	refuse := func(code, why string) { h.mu.Unlock(); reject(code, why) }
	c := h.convos[m.Conversation]
	if c == nil {
		refuse("conversation", "no such conversation")
		return
	}
	if c.work == nil {
		refuse("storage", "conversation work metadata could not be loaded")
		return
	}
	if h.trees == nil {
		refuse("tree", "this host has no tree executor")
		return
	}
	// A save builds on the latest revision; a launch names the one it runs.
	if m.Type == "save-spec" && m.SpecRevision != c.work.Revision {
		refuse("spec-conflict", "the spec changed; reload it before saving")
		return
	}
	switch m.Type {
	case "save-spec":
		if strings.TrimSpace(m.Text) == "" {
			refuse("spec", "a spec needs a goal")
			return
		}
		if len(m.Text) > 256*1024 {
			refuse("spec", "the spec exceeds 256 KiB")
			return
		}
		if latest := c.work.latest(); latest != nil && latest.Text == m.Text && slices.Equal(latest.Tools, sortedTools(m.Reach)) {
			h.mu.Unlock()
			h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
			return
		}
	case "launch-spec":
		if m.SpecRevision < 1 || m.SpecRevision > c.work.Revision {
			refuse("spec", "there is no such revision to launch")
			return
		}
		if c.work.Revisions[m.SpecRevision-1].Observation == "" {
			refuse("spec", "this spec predates Missions; save it again to launch it")
			return
		}
		if h.treeBusy(c.work.TreeID) {
			refuse("busy", "this mission is already launched")
			return
		}
	}
	if err := h.admit(m, c.id); err != nil {
		refuse("storage", err.Error())
		return
	}
	home, model := h.homeOf(c), c.model
	var prior *specRevision
	if latest := c.work.latest(); latest != nil {
		copied := *latest
		prior = &copied
	}
	h.mu.Unlock()
	if model == "" {
		model = "codex"
	}
	fail := func(code, why string) {
		h.mu.Lock()
		h.finishRequest(m.RequestID, "failed", code, why)
		h.mu.Unlock()
		reject(code, why)
	}
	s, err := h.missionSetup()
	if err != nil {
		fail("mission", err.Error())
		return
	}
	b, err := h.bridge(s, c.id, home)
	if err != nil {
		fail("mission", err.Error())
		return
	}
	switch m.Type {
	case "save-spec":
		tools := sortedTools(m.Reach)
		refs, err := s.capabilities(tools)
		if err != nil {
			fail("spec", err.Error())
			return
		}
		// An edit revises the last Spec: what the human kept keeps its source,
		// and its profile unless the human set one.
		profile := typed.DefaultProfile
		prev, _, had := priorSpec(b, s, prior)
		if had {
			profile = prev.Profile
		}
		if m.Profile != nil {
			profile = *m.Profile
		}
		spec := typed.HumanSpec(m.Text, []root.ID{}, refs, profile)
		if had {
			spec = typed.Revise(prev, m.Text, refs, profile)
		}
		authored, err := b.AuthorSpec(m.RequestID, spec)
		if err != nil {
			fail("spec", err.Error())
			return
		}
		if len(authored.Observations) != 1 {
			fail("spec", "the human root recorded no spec")
			return
		}
		h.commitRevision(m, specRevision{Text: m.Text, Tools: tools, Inputs: ids(spec.Inputs), Source: "human", Observation: string(authored.Observations[0]), Epoch: s.epoch, Segments: segmentViews(spec, priorFrom(prior)), Coverage: &spec.Coverage, Profile: &spec.Profile}, nil, fail)
	case "launch-spec":
		h.mu.Lock()
		latest := c.work.Revisions[m.SpecRevision-1]
		h.mu.Unlock()
		if latest.Epoch != s.epoch {
			fail("spec", "Skyra's tools were rebuilt since this spec was saved; save it again to launch it")
			return
		}
		launch, err := b.StartLaunch(m.RequestID, root.ID(latest.Observation))
		if err != nil {
			fail("launch", err.Error())
			return
		}
		if launch == nil {
			fail("launch", "this launch was already dispatched")
			return
		}
		h.launch(m, c, b, launch, latest, model, fail)
	}
}

func sortedTools(tools []string) []string {
	out := slices.Clone(tools)
	slices.Sort(out)
	return slices.Compact(out)
}

func (h *webhost) treeBusy(treeID string) bool {
	h.trees.mu.Lock()
	defer h.trees.mu.Unlock()
	for _, r := range h.trees.runs {
		if r.TreeID == treeID && r.Phase == "running" {
			return true
		}
	}
	return false
}

// commitRevision publishes a spec revision recorded at the root, or links a
// launched run. A revision the human root holds is the only kind a Launch may
// pass. A Plan's revision arrives after its request was answered.
func (h *webhost) commitRevision(m inbound, r specRevision, run *treeRun, fail func(string, string)) {
	answer := true
	h.mu.Lock()
	c := h.convos[m.Conversation]
	if c == nil || c.work == nil {
		h.mu.Unlock()
		fail("conversation", "the conversation is gone")
		return
	}
	next := *c.work
	if run == nil {
		next.Revision++
		r.Revision = next.Revision
		next.Spec = r.Text
		next.Revisions = append(append([]specRevision{}, next.Revisions...), r)
	} else {
		next.LastRunID = run.ID
	}
	if err := saveWork(h.homeOf(c), &next); err != nil {
		h.mu.Unlock()
		fail("storage", fmt.Sprintf("save conversation work: %v", err))
		return
	}
	c.work = &next
	if answer {
		h.finishRequest(m.RequestID, "returned", "", "")
	}
	h.mu.Unlock()
	if answer {
		h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
	}
	h.broadcast()
}

// runPlan is one Plan turn: the human's message and the conversation so far
// are authored at the root, then passed to Plan with the latest Spec. It
// returns the Spec Plan returned, as the root holds it, and what each
// observation it may quote was.
func (h *webhost) runPlan(requestID, text string, s *missionSetup, b *mission.Bridge, model, history string, prior *specRevision) (typed.Spec, root.ID, map[root.ID]string, error) {
	value, _ := json.Marshal(text)
	request, err := b.Root.Author(requestID+"/request", value)
	if err != nil || len(request.Observations) != 1 {
		return typed.Spec{}, "", nil, fmt.Errorf("record your message: %v", err)
	}
	from := map[root.ID]string{request.Observations[0]: "request"}
	context := []root.ID{}
	if history != "" {
		value, _ = json.Marshal(history)
		conversation, err := b.Root.Author(requestID+"/conversation", value)
		if err != nil || len(conversation.Observations) != 1 {
			return typed.Spec{}, "", nil, fmt.Errorf("record the conversation: %v", err)
		}
		context = conversation.Observations
		from[conversation.Observations[0]] = "conversation"
	}
	// Plan refines the latest Spec: it is passed, so Plan may quote what a
	// human wrote or cited in it, and must add again any words of its own.
	if _, priorID, ok := priorSpec(b, s, prior); ok {
		context = append(context, priorID)
		from[priorID] = "spec"
	}
	dispatch, _, err := b.StartPlan(requestID+"/plan", request.Observations[0], context, b.PlanGrant())
	if err != nil {
		return typed.Spec{}, "", nil, err
	}
	if dispatch.Kind != root.Dispatch {
		return typed.Spec{}, "", nil, fmt.Errorf("Plan was not dispatched: %s", dispatch.Reason)
	}
	ctx, cancel := contextFor(30 * time.Minute)
	defer cancel()
	returned, err := b.ExecutePlan(ctx, h.skyra, model, dispatch)
	if err != nil {
		return typed.Spec{}, "", nil, err
	}
	state, err := b.Root.State()
	if err != nil || len(returned.Observations) != 1 {
		return typed.Spec{}, "", nil, fmt.Errorf("Plan returned no answer: %v", err)
	}
	observation, _ := state.Lookup(returned.Observations[0])
	var reply string
	if json.Unmarshal(observation.Value(), &reply) == nil {
		return typed.Spec{Note: reply}, observation.ID(), from, nil
	}
	spec, err := typed.DecodeSpec(observation.Value())
	if err != nil {
		return typed.Spec{}, "", nil, err
	}
	return spec, observation.ID(), from, nil
}

func contextFor(d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), d)
}

// transcript is the conversation's exchange, newest last, bounded: what Plan
// is passed as context. Tool calls stay out; they are the agent's working.
func transcript(entries []entry) string {
	var parts []string
	for _, e := range entries {
		if (e.Role == "user" || e.Role == "agent") && strings.TrimSpace(e.Text) != "" {
			parts = append(parts, e.Role+": "+e.Text)
		}
	}
	text := strings.Join(parts, "\n\n")
	const limit = 64 * 1024
	if len(text) > limit {
		text = text[len(text)-limit:]
	}
	return text
}

// launch records the run, links it to the mission, and executes the committed
// handoff once. The tree view replays the Launch's own kernel run.
func (h *webhost) launch(m inbound, c *conversation, b *mission.Bridge, launch *mission.Launch, spec specRevision, model string, fail func(string, string)) {
	id, err := runID()
	if err != nil {
		fail("launch", err.Error())
		return
	}
	call := launch.Dispatch.Invocation.ID
	h.mu.Lock()
	treeID, specID := c.work.TreeID, c.work.SpecID
	h.mu.Unlock()
	r := &treeRun{Actor: model, ID: id, Goal: spec.Text, Kind: launch.Resolved.Kind, Reach: launch.Resolved.Targets, Phase: "running", Started: time.Now().UnixMilli(),
		Conversation: c.id, TreeID: treeID, SpecID: specID, SpecRevision: spec.Revision, Root: b.LaunchRoot(call), Spec: spec.Observation, Call: string(call), root: b.LaunchRoot(call)}
	t := h.trees
	t.mu.Lock()
	t.runs[id] = r
	if err := t.save(r); err != nil {
		delete(t.runs, id)
		t.mu.Unlock()
		fail("storage", err.Error())
		return
	}
	t.mu.Unlock()
	h.commitRevision(m, specRevision{}, r, fail)
	h.sendTrees()
	go func() {
		result, err := b.ExecuteLaunch(context.Background(), h.skyra, model, launch)
		answer := ""
		if state, stateErr := b.Root.State(); stateErr == nil && len(result.Observations) == 1 {
			if o, ok := state.Lookup(result.Observations[0]); ok {
				answer = string(o.Value())
			}
		}
		t.mu.Lock()
		r.View = t.render(r.root)
		r.Answer = answer
		switch {
		case err != nil:
			r.Phase, r.Error = "failed", err.Error()
		case answer == "" || strings.Contains(answer, `"faulted"`) || strings.Contains(answer, `"fault"`):
			r.Phase = "failed"
		default:
			r.Phase = "done"
		}
		t.save(r)
		t.mu.Unlock()
		h.sendTrees()
		h.broadcast()
	}()
}

func ids(in []root.ID) []string {
	out := make([]string, len(in))
	for i, id := range in {
		out[i] = string(id)
	}
	return out
}

func segmentViews(spec typed.Spec, from map[root.ID]string) []segmentView {
	out := make([]segmentView, len(spec.Segments))
	for i, seg := range spec.Segments {
		out[i] = segmentView{Text: seg.Text, Source: seg.Source, Of: string(seg.Of), From: from[seg.Of]}
	}
	return out
}

// priorSpec is the Spec a revision builds on: the latest one, as the root
// holds it, if it was made at these pins.
func priorSpec(b *mission.Bridge, s *missionSetup, prior *specRevision) (typed.Spec, root.ID, bool) {
	if prior == nil || prior.Observation == "" || prior.Epoch != s.epoch {
		return typed.Spec{}, "", false
	}
	state, err := b.Root.State()
	if err != nil {
		return typed.Spec{}, "", false
	}
	o, ok := state.Lookup(root.ID(prior.Observation))
	if !ok {
		return typed.Spec{}, "", false
	}
	spec, err := typed.DecodeSpec(o.Value())
	return spec, o.ID(), err == nil
}

// priorFrom is what each observation the prior revision quoted was.
func priorFrom(prior *specRevision) map[root.ID]string {
	from := map[root.ID]string{}
	if prior != nil {
		for _, seg := range prior.Segments {
			if seg.Of != "" {
				from[root.ID(seg.Of)] = seg.From
			}
		}
	}
	return from
}
