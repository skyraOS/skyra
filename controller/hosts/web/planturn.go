package main

// A Mission's conversation is Plan. Each message the human sends is one Plan
// turn: the message and the conversation so far are authored at the Mission's
// human root and passed to Plan with the latest Spec, and Plan returns the
// Spec as it now stands with a note to the human. The note is Plan's reply;
// a Spec that asks something new of Launch is the turn's revision. There is
// no other agent in the conversation.

import (
	"fmt"
	"strings"

	root "skyra/controller/core/humanroot"
	typed "skyra/controller/core/mission"
	"skyra/controller/runtime/mission"
)

// planTurn is one exchange: what the human said, and what Plan answered.
type planTurn struct {
	RequestID string `json:"requestId"`
	Text      string `json:"text"`
	Note      string `json:"note,omitempty"`
	// Revision is the spec revision this turn made, or 0 if the Spec it
	// returned asked nothing new of Launch.
	Revision int    `json:"revision,omitempty"`
	Error    string `json:"error,omitempty"`
	Done     bool   `json:"done,omitempty"`
}

// plans reports whether a new conversation of this procedure is Plan's: the
// think root, on a host that can run Plan.
func (h *webhost) plans(procedure string) bool {
	return procedure == "think" && h.trees != nil
}

// entries is the transcript a Plan conversation shows: each message, then
// Plan's reply, which names the revision it made.
func (w *conversationWork) entries() []entry {
	out := []entry{}
	for _, t := range w.Turns {
		id := t.RequestID
		out = append(out, entry{Role: "user", Text: t.Text, TS: int64(len(out)), RequestID: &id})
		switch {
		case t.Error != "":
			out = append(out, entry{Role: "agent", Text: "⚠ Plan could not answer: " + t.Error, TS: int64(len(out))})
		case t.Done:
			out = append(out, entry{Role: "agent", Text: t.Note, TS: int64(len(out)), Revision: t.Revision})
		}
	}
	return out
}

// planMessage starts a Plan turn for a message. Caller holds h.mu; it is
// released before returning.
func (h *webhost) planMessage(c *conversation, m inbound, admit func(string) bool, reject func(string, string)) {
	refuse := func(code, why string) { h.mu.Unlock(); reject(code, why) }
	switch {
	case c.work.Planning || c.pending != "":
		refuse("busy", "Plan is still answering")
		return
	case len(m.Images) > 0:
		refuse("invalid-image", "Plan reads text; images are not passed to it yet")
		return
	case strings.TrimSpace(m.Text) == "":
		refuse("empty", "say something to Plan")
		return
	}
	history := transcript(c.entries)
	var prior *specRevision
	if latest := c.work.latest(); latest != nil {
		copied := *latest
		prior = &copied
	}
	if !admit(c.id) {
		return
	}
	next := *c.work
	next.Turns = append(append([]planTurn{}, next.Turns...), planTurn{RequestID: m.RequestID, Text: m.Text})
	next.Planning, next.PlanError = true, ""
	if err := saveWork(h.homeOf(c), &next); err != nil {
		h.finishRequest(m.RequestID, "failed", "storage", err.Error())
		refuse("storage", err.Error())
		return
	}
	c.work = &next
	c.entries = next.entries()
	c.phase, c.pending = "pending", m.RequestID
	home, model := h.homeOf(c), c.model
	h.mu.Unlock()
	h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
	h.broadcast()
	go h.planTurn(c, m, home, model, history, prior)
}

// planTurn runs Plan and records its answer, and the revision it made.
func (h *webhost) planTurn(c *conversation, m inbound, home, model, history string, prior *specRevision) {
	if model == "" {
		model = "codex"
	}
	var revision *specRevision
	note, why := "", ""
	s, err := h.missionSetup()
	var b *mission.Bridge
	if err == nil {
		b, err = h.bridge(s, c.id, home)
	}
	if err == nil {
		var spec typed.Spec
		var id root.ID
		var from map[root.ID]string
		spec, id, from, err = h.runPlan(m.RequestID, m.Text, s, b, model, history, prior)
		if err == nil {
			note = strings.TrimSpace(spec.Note)
			// A Spec that asks something new of Launch is this turn's revision.
			before, _, had := priorSpec(b, s, prior)
			if strings.TrimSpace(spec.Goal) != "" && (!had || !before.SameTask(spec)) {
				revision = &specRevision{Text: spec.Goal, Tools: s.tools(spec.Capabilities), Inputs: ids(spec.Inputs), Source: "plan", Observation: string(id), Epoch: s.epoch, Segments: segmentViews(spec, from), Coverage: &spec.Coverage, Profile: &spec.Profile}
			}
			if note == "" && revision != nil {
				note = "I updated the spec."
			}
		}
	}
	if err != nil {
		why = outOf(err.Error())
		fmt.Fprintf(h.log, "webhost: plan %s: %v\n", c.id, err)
	}
	h.mu.Lock()
	defer h.broadcast()
	defer h.mu.Unlock()
	next := *c.work
	turns := append([]planTurn{}, next.Turns...)
	turn := &turns[len(turns)-1]
	if revision != nil {
		next.Revision++
		revision.Revision = next.Revision
		next.Spec = revision.Text
		next.Revisions = append(append([]specRevision{}, next.Revisions...), *revision)
		turn.Revision = revision.Revision
	}
	turn.Note, turn.Error, turn.Done = note, why, true
	next.Turns = turns
	next.Planning, next.PlanError = false, why
	if err := saveWork(home, &next); err != nil {
		fmt.Fprintf(h.log, "webhost: plan %s: save: %v\n", c.id, err)
		why = err.Error()
	}
	c.work = &next
	c.entries = next.entries()
	c.phase, c.pending = "ready", ""
	if why != "" {
		h.finishRequest(m.RequestID, "failed", "plan", why)
	} else {
		h.finishRequest(m.RequestID, "returned", "", "")
	}
}

// becomePlan makes a Mission's chat from before Plan Plan's: each message and
// the agent's reply to it become an earlier turn, and the chat run, if one is
// waiting, is left. Caller holds h.mu.
func (h *webhost) becomePlan(c *conversation) error {
	turns := []planTurn{}
	for _, e := range c.entries {
		switch e.Role {
		case "user":
			id := ""
			if e.RequestID != nil {
				id = *e.RequestID
			}
			turns = append(turns, planTurn{RequestID: id, Text: e.Text, Done: true})
		case "agent":
			if n := len(turns); n > 0 {
				turns[n-1].Note = strings.TrimSpace(turns[n-1].Note + "\n\n" + e.Text)
			}
		}
	}
	next := *c.work
	next.Plan, next.Turns = true, turns
	if err := saveWork(h.homeOf(c), &next); err != nil {
		return err
	}
	c.work = &next
	c.entries = next.entries()
	c.parked, c.phase = nil, "ready"
	return nil
}

// outOf is a failed turn's reason as the human reads it: a Plan tree that hit
// one of its profile's bounds ran out, and a narrower or further message goes
// on from the spec as it stands. Any other failure is reported as it is.
func outOf(why string) string {
	switch {
	case strings.Contains(why, "the tree's budget"):
		return "I used all my turns for this message before I could answer. Ask about something narrower, or send another message to go on from the spec as it stands."
	case strings.Contains(why, "the turn bound"):
		return "one of my steps ran out of turns before I could answer. Ask about something narrower, or send another message to go on from the spec as it stands."
	case strings.Contains(why, "the depth bound"):
		return "I tried to split the work deeper than I may. Send the message again, or ask more directly."
	}
	return why
}
