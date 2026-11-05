package main

// Conversation work is the Mission's metadata, separate from the user
// continuation and the tree machine: its spec revisions, each one a Spec
// observation held at the Mission's human root (mission.go). Only an explicit
// launch-spec passes one to a tree; neither chat messages nor edits become
// tree input implicitly.
import (
	"encoding/json"
	"os"
	"path/filepath"

	typed "skyra/controller/core/mission"
)

type specRevision struct {
	Revision int      `json:"revision"`
	Text     string   `json:"text"`
	Tools    []string `json:"tools,omitempty"`  // the capabilities it proposes, by tool
	Inputs   []string `json:"inputs,omitempty"` // the observations it cites, at the root
	Source   string   `json:"source,omitempty"` // "human" or "plan"; empty before Missions
	// Observation is the Spec held at the human root, at Epoch's pins.
	Observation string `json:"observation,omitempty"`
	Epoch       string `json:"epoch,omitempty"`
	// Segments and Coverage are where the Spec's words came from, as the
	// Spec itself records them; empty before cite coverage.
	Segments []segmentView   `json:"segments,omitempty"`
	Coverage *typed.Coverage `json:"coverage,omitempty"`
	// Profile is the constants its Launch's tree runs under, as the Spec
	// states them; empty before Specs did.
	Profile *typed.Profile `json:"profile,omitempty"`
}

// segmentView is a Spec segment as the desktop shows it: From names what a
// cited segment quotes, "request" or "conversation".
type segmentView struct {
	Text   string `json:"text"`
	Source string `json:"source"`
	Of     string `json:"of,omitempty"`
	From   string `json:"from,omitempty"`
}

type conversationWork struct {
	Conversation string         `json:"conversation"`
	TreeID       string         `json:"treeId"`
	SpecID       string         `json:"specId"`
	Revision     int            `json:"revision"`
	Spec         string         `json:"spec"`
	Revisions    []specRevision `json:"revisions"`
	LastRunID    string         `json:"lastRunId,omitempty"`
	Planning     bool           `json:"planning,omitempty"`
	PlanError    string         `json:"planError,omitempty"`
	// Plan marks a conversation that is Plan's (planturn.go): its turns are
	// its transcript, and there is no chat run.
	Plan  bool       `json:"plan,omitempty"`
	Turns []planTurn `json:"turns,omitempty"`
}

// latest is the current revision, or nil before the first.
func (w *conversationWork) latest() *specRevision {
	if w.Revision == 0 {
		return nil
	}
	return &w.Revisions[w.Revision-1]
}

func newWork(id string) *conversationWork {
	return &conversationWork{Conversation: id, TreeID: "tree-" + id, SpecID: "spec-" + id, Revisions: []specRevision{}}
}

func loadWork(home, id string) *conversationWork {
	w := &conversationWork{}
	b, err := os.ReadFile(filepath.Join(home, "work.json"))
	if os.IsNotExist(err) {
		return newWork(id)
	} // older conversations acquire stable identities
	if err != nil || json.Unmarshal(b, w) != nil || w.Conversation != id || w.TreeID != "tree-"+id || w.SpecID != "spec-"+id {
		// Do not turn corrupt operator metadata into a fresh executable spec.
		return nil
	}
	if w.Revision < 0 || len(w.Revisions) != w.Revision {
		return nil
	}
	for i, revision := range w.Revisions {
		if revision.Revision != i+1 {
			return nil
		}
	}
	if w.Revision == 0 {
		if w.Spec != "" {
			return nil
		}
	} else if w.Revisions[w.Revision-1].Text != w.Spec {
		return nil
	}
	return w
}

func atomicWorkFile(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".work-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = f.Write(b); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(f.Name(), path); err != nil {
		return err
	}
	d, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	defer d.Close()
	return d.Sync()
}

func saveWork(home string, w *conversationWork) error {
	b, err := json.Marshal(w)
	if err != nil {
		return err
	}
	return atomicWorkFile(filepath.Join(home, "work.json"), b)
}

// maxShownRevisions bounds the revisions a snapshot carries; older ones
// stay on disk.
const maxShownRevisions = 50

// projection is the work a snapshot carries: the current spec, and the
// latest revisions, so a turn's reply can show the Spec as it stood then.
func (w *conversationWork) projection() map[string]any {
	p := map[string]any{"conversation": w.Conversation, "treeId": w.TreeID, "specId": w.SpecID, "revision": w.Revision, "spec": w.Spec, "lastRunId": w.LastRunID, "planning": w.Planning, "planError": w.PlanError}
	if r := w.latest(); r != nil {
		p["tools"], p["source"], p["inputs"], p["launchable"] = r.Tools, r.Source, len(r.Inputs), r.Observation != ""
		if r.Coverage != nil {
			p["segments"], p["coverage"] = r.Segments, r.Coverage
		}
		if r.Profile != nil {
			p["profile"] = r.Profile
		}
	}
	if w.Plan {
		p["plan"] = true
		shown := []map[string]any{}
		for _, r := range w.Revisions[max(0, len(w.Revisions)-maxShownRevisions):] {
			v := map[string]any{"revision": r.Revision, "text": r.Text, "tools": r.Tools, "source": r.Source, "inputs": len(r.Inputs), "launchable": r.Observation != ""}
			if r.Coverage != nil {
				v["segments"], v["coverage"] = r.Segments, r.Coverage
			}
			if r.Profile != nil {
				v["profile"] = r.Profile
			}
			shown = append(shown, v)
		}
		p["revisions"] = shown
	}
	return p
}

// Caller holds h.mu. A spec and its last result are conversation context,
// never instructions sent to the tree. A turn records the exact revision it saw.
func (h *webhost) specContext(c *conversation) map[string]any {
	if c.work == nil || c.work.Revision == 0 {
		return nil
	}
	context := c.work.projection()
	if h.trees != nil {
		h.trees.mu.Lock()
		if run := h.trees.runs[c.work.LastRunID]; run != nil && run.TreeID == c.work.TreeID && run.Conversation == c.id {
			context["execution"] = map[string]any{"revision": run.SpecRevision, "phase": run.Phase, "answer": run.Answer, "error": run.Error}
		}
		h.trees.mu.Unlock()
	}
	return context
}
