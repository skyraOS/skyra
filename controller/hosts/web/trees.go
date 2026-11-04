package main

// Trees (56): the human gives a spec, the webhost boots one tree of it on
// the apprentice, as the operator, and streams the tree as it forms. The
// view is the treeview binary's: the tree's log replayed through Step, so
// the page shows only what the machine itself made of the log.

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
)

// treeTargets are what a tree's root may be given to hold: read only, so a
// tree started from the page changes nothing.
var treeTargets = []string{"find", "read", "echo"}

type treeRun struct {
	Actor   string          `json:"actor"`
	ID      string          `json:"id"`
	Goal    string          `json:"goal"`
	Kind    string          `json:"kind"`
	Reach   []string        `json:"reach"`
	Phase   string          `json:"phase"` // running | done | failed
	Answer  string          `json:"answer,omitempty"`
	Error   string          `json:"error,omitempty"`
	Started int64           `json:"started"`
	View    json.RawMessage `json:"view,omitempty"` // treeview's frames; null until the first resume

	root string
}

type trees struct {
	mu       sync.Mutex
	runs     map[string]*treeRun
	dir      string // where every tree's run lives
	view     string // the treeview binary
	registry string // the operator's registry, at the operator's depth
}

func newTrees(dir, view, registry string) *trees {
	t := &trees{runs: map[string]*treeRun{}, dir: dir, view: view, registry: registry}
	t.load()
	return t
}

// load lists past trees from their run directories.
func (t *trees) load() {
	ds, _ := filepath.Glob(filepath.Join(t.dir, "t-*", "tree.json"))
	for _, d := range ds {
		b, err := os.ReadFile(d)
		var r treeRun
		if err != nil || json.Unmarshal(b, &r) != nil {
			continue
		}
		if r.Actor == "" {
			r.Actor = "claude"
		}
		r.root = filepath.Join(filepath.Dir(d), ".skyra")
		if r.Phase == "running" { // its host is gone
			r.Phase, r.Error = "failed", "the webhost stopped while it ran"
		}
		r.View = t.render(r.root)
		t.runs[r.ID] = &r
	}
}

func (t *trees) save(r *treeRun) {
	c := *r
	c.View = nil
	b, _ := json.Marshal(c)
	_ = os.WriteFile(filepath.Join(filepath.Dir(r.root), "tree.json"), b, 0o600)
}

// render is treeview's view of a run, or nil before it has one.
func (t *trees) render(root string) json.RawMessage {
	if _, err := os.Stat(filepath.Join(root, "apprentice", "tree", "frames")); err != nil {
		return nil
	}
	out, err := exec.Command(t.view, "-json", root).Output()
	if err != nil || !json.Valid(out) {
		return nil
	}
	return bytes.TrimSpace(out)
}

// start boots a tree; it returns the run's id, or why it was refused.
func (h *webhost) startTree(goal, kind string, reach []string, actors ...string) (string, string) {
	t := h.trees
	if t == nil {
		return "", "this webhost was started without trees"
	}
	actor := "claude"
	if len(actors) > 0 && actors[0] != "" {
		actor = actors[0]
	}
	if actor != "claude" && actor != "codex" {
		return "", "the tree actor is claude or codex"
	}
	if strings.TrimSpace(goal) == "" {
		return "", "a tree needs a spec"
	}
	if kind != "act" && kind != "delegate" {
		return "", "the root is act or delegate"
	}
	if len(reach) == 0 {
		return "", "the root holds at least one target"
	}
	if kind == "act" && len(reach) != 1 {
		return "", "an act root holds exactly one target"
	}
	if kind == "delegate" && len(reach) < 2 {
		return "", "a delegate root holds two or more targets"
	}
	for _, r := range reach {
		if !slices.Contains(treeTargets, r) {
			return "", r + " is not a target a tree may hold here"
		}
	}
	nonce := make([]byte, 8)
	if _, err := rand.Read(nonce); err != nil {
		return "", err.Error()
	}
	id := "t-" + time.Now().Format("20060102-150405") + "-" + hex.EncodeToString(nonce)
	r := &treeRun{Actor: actor, ID: id, Goal: goal, Kind: kind, Reach: reach, Phase: "running", Started: time.Now().UnixMilli(),
		root: filepath.Join(t.dir, id, ".skyra")}
	if err := os.MkdirAll(r.root, 0o700); err != nil {
		return "", err.Error()
	}
	t.mu.Lock()
	t.runs[id] = r
	t.save(r)
	t.mu.Unlock()
	go h.runTree(r)
	return id, ""
}

func (h *webhost) runTree(r *treeRun) {
	t := h.trees
	payload, _ := json.Marshal(map[string]any{"tree": map[string]any{"goal": r.Goal, "kind": r.Kind, "actor": r.Actor, "reach": r.Reach}})
	cmd := exec.Command(h.skyra, "boot", "-root", r.root, "-registry", t.registry, "-scratch", h.scratchFor(filepath.Dir(r.root)), "-to", "apprentice", "-payload", string(payload))
	var out bytes.Buffer
	cmd.Stdout = io.MultiWriter(&out, h.log)
	cmd.Stderr = h.log
	err := cmd.Run()
	answer := ""
	sc := bufio.NewScanner(&out)
	sc.Buffer(make([]byte, 1<<20), 16<<20)
	for sc.Scan() {
		if a, ok := strings.CutPrefix(sc.Text(), "answer: "); ok {
			answer = a
		}
	}
	t.mu.Lock()
	r.View = t.render(r.root)
	r.Answer = answer
	switch {
	case err != nil:
		r.Phase, r.Error = "failed", err.Error()
	case strings.Contains(answer, `"faulted"`) || answer == "":
		r.Phase = "failed"
	default:
		r.Phase = "done"
	}
	t.save(r)
	t.mu.Unlock()
	h.sendTrees()
}

// refresh re-renders every running tree; it reports whether any changed.
func (t *trees) refresh() bool {
	t.mu.Lock()
	var running []*treeRun
	for _, r := range t.runs {
		if r.Phase == "running" {
			running = append(running, r)
		}
	}
	t.mu.Unlock()
	changed := false
	for _, r := range running {
		v := t.render(r.root)
		t.mu.Lock()
		if v != nil && !bytes.Equal(v, r.View) {
			r.View, changed = v, true
		}
		t.mu.Unlock()
	}
	return changed
}

func (t *trees) watch(send func()) {
	for range time.Tick(time.Second) {
		if t.refresh() {
			send()
		}
	}
}

// sendTrees sends every tree, newest first, as one message.
func (h *webhost) sendTrees() {
	t := h.trees
	if t == nil {
		return
	}
	t.mu.Lock()
	list := make([]treeRun, 0, len(t.runs))
	for _, r := range t.runs {
		list = append(list, *r)
	}
	t.mu.Unlock()
	sort.Slice(list, func(i, j int) bool { return list[i].Started > list[j].Started })
	b, _ := json.Marshal(map[string]any{"type": "trees", "targets": treeTargets, "trees": list})
	h.sendRaw(b)
}
