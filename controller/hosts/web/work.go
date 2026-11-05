package main

// Conversation work is operator-authored metadata, separate from the user
// continuation and the tree machine. Only an explicit run-spec dispatches a
// saved spec; neither chat messages nor edits become tree input implicitly.
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type specRevision struct {
	Revision int    `json:"revision"`
	Text     string `json:"text"`
}

type conversationWork struct {
	Conversation string         `json:"conversation"`
	TreeID       string         `json:"treeId"`
	SpecID       string         `json:"specId"`
	Revision     int            `json:"revision"`
	Spec         string         `json:"spec"`
	Revisions    []specRevision `json:"revisions"`
	LastRunID    string         `json:"lastRunId,omitempty"`
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

// Called under h.mu, and releases it on every path. The ordinary request
// ledger provides reconnect deduplication for edits and run admission alike.
func (h *webhost) handleWork(m inbound, reject func(string, string)) {
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
	if m.SpecRevision != c.work.Revision {
		refuse("spec-conflict", "the spec changed; reload it before saving or running")
		return
	}
	if m.Type == "run-spec" {
		if h.trees == nil {
			refuse("tree", "this host has no tree executor")
			return
		}
		if strings.TrimSpace(c.work.Spec) == "" {
			refuse("spec", "save a spec before running it")
			return
		}
		h.trees.mu.Lock()
		busy := false
		for _, r := range h.trees.runs {
			if r.TreeID == c.work.TreeID && r.Phase == "running" {
				busy = true
				break
			}
		}
		h.trees.mu.Unlock()
		if busy {
			refuse("busy", "this conversation's tree is already running")
			return
		}
	}
	if len(m.Text) > 256*1024 {
		refuse("spec", "the spec exceeds 256 KiB")
		return
	}
	if err := h.admit(m, c.id); err != nil {
		refuse("storage", err.Error())
		return
	}
	fail := func(code, why string) { h.finishRequest(m.RequestID, "failed", code, why); refuse(code, why) }
	next := *c.work
	var run *treeRun
	if m.Type == "save-spec" {
		if m.Text != next.Spec {
			next.Revision++
			next.Spec = m.Text
			next.Revisions = append(append([]specRevision{}, next.Revisions...), specRevision{next.Revision, m.Text})
		}
	} else {
		actor := c.model
		if actor == "" {
			actor = "codex"
		}
		kind := "delegate"
		if len(m.Reach) == 1 {
			kind = "act"
		}
		var why string
		run, why = h.prepareTree(next.Spec, kind, m.Reach, &next, actor)
		if why != "" {
			fail("tree", why)
			return
		}
		next.LastRunID = run.ID
	}
	if err := saveWork(h.homeOf(c), &next); err != nil {
		if run != nil {
			h.trees.mu.Lock()
			run.Phase, run.Error = "failed", "spec linkage could not be saved; execution did not start"
			_ = h.trees.save(run)
			h.trees.mu.Unlock()
		}
		fail("storage", fmt.Sprintf("save conversation work: %v", err))
		return
	}
	c.work = &next
	h.finishRequest(m.RequestID, "returned", "", "")
	h.mu.Unlock()
	h.send(map[string]string{"type": "accepted", "requestId": m.RequestID})
	h.broadcast()
	if run != nil {
		h.sendTrees()
		go h.runTree(run)
	}
}

// Revision history stays on disk; snapshots carry only the current spec.
func (w *conversationWork) projection() map[string]any {
	return map[string]any{"conversation": w.Conversation, "treeId": w.TreeID, "specId": w.SpecID, "revision": w.Revision, "spec": w.Spec, "lastRunId": w.LastRunID}
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
