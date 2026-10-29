package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// destination is a root the human can select (proposal 38): an authored
// procedure, the actor it runs on, its contract, and its outline. Listing one
// grants nothing; intake checks every message against the conversation's
// Boundary.
type destination struct {
	Actor   string          `json:"actor"`
	Title   string          `json:"title"`
	Schema  json.RawMessage `json:"schema"`
	Outline outline         `json:"outline"`
}

// outline is a procedure's composition as it is written: each node, in its
// order, what it calls, and what each narrowing keeps. It is derived from the
// procedure's definition, never drawn by hand, so its lines are the
// procedure's real structure. Below the root everything is inspectable only.
type outline struct {
	Nodes []outlineNode `json:"nodes"`
}

type outlineNode struct {
	Name    string     `json:"name"`
	Calls   string     `json:"calls"`             // the actor it invokes: a call through a narrowing's reference names the narrowed actor
	Via     string     `json:"via,omitempty"`     // a call through a narrowing whose actor is a parameter: that narrowing's node
	After   []string   `json:"after"`             // the nodes it waits on
	Narrows *narrowing `json:"narrows,omitempty"` // a narrowing node: the actor it narrows, and the reach it keeps
}

type narrowing struct {
	Actor string   `json:"actor"`           // "" when the actor is a parameter of the procedure
	Param string   `json:"param,omitempty"` // that parameter: think narrows to its model
	Reach []string `json:"reach"`
}

// loadDestinations reads every authored procedure the operator's registry
// installs: each entry that names a bundle contributes its procedures as
// roots on that entry. Remembered procedures, whose nodes live in the
// apprentice's state, are not roots yet.
func loadDestinations(registryFile string) ([]destination, error) {
	b, err := os.ReadFile(registryFile)
	if err != nil {
		return nil, err
	}
	var registry map[string]struct {
		Bundle string   `json:"bundle"`
		Uses   []string `json:"uses"`
	}
	if err := json.Unmarshal(b, &registry); err != nil {
		return nil, err
	}
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	sort.Strings(names)
	out := []destination{}
	for _, actor := range names {
		e := registry[actor]
		if len(e.Uses) > 0 && !strings.Contains(actor, ".") {
			adapters[actor] = true // runs through other actors: its steps are its operations, not tool calls
		}
		if e.Bundle == "" {
			continue
		}
		path := e.Bundle
		if !filepath.IsAbs(path) {
			path = filepath.Join(filepath.Dir(registryFile), path)
		}
		ds, err := bundleDestinations(actor, path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", actor, err)
		}
		out = append(out, ds...)
	}
	return out, nil
}

func bundleDestinations(actor, path string) ([]destination, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var bundle struct {
		Procedures []struct {
			Contract json.RawMessage `json:"contract"`
			Order    []string        `json:"order"`
			Nodes    map[string]struct {
				Call  json.RawMessage `json:"call"`
				In    json.RawMessage `json:"in"`
				After []string        `json:"after"`
			} `json:"nodes"`
		} `json:"procedures"`
	}
	if err := json.Unmarshal(b, &bundle); err != nil {
		return nil, err
	}
	var out []destination
	for _, p := range bundle.Procedures {
		var c struct {
			Title string `json:"title"`
		}
		if json.Unmarshal(p.Contract, &c) != nil || c.Title == "" {
			continue // a root is selected by its title
		}
		// Its conversations are kept in a directory named for it.
		if !validConversationID(c.Title) || strings.HasPrefix(c.Title, "c-") {
			return nil, fmt.Errorf("procedure %q cannot name the directory its conversations are kept in", c.Title)
		}
		order := append([]string(nil), p.Order...)
		for name := range p.Nodes {
			if !contains(order, name) {
				order = append(order, name)
			}
		}
		narrows := map[string]*narrowing{}
		var o outline
		for _, name := range order {
			n, ok := p.Nodes[name]
			if !ok {
				continue
			}
			node := outlineNode{Name: name, After: append([]string{}, n.After...)}
			var target string
			if json.Unmarshal(n.Call, &target) == nil {
				node.Calls = target
			} else if from := bindingNode(n.Call); narrows[from] != nil {
				node.Calls = narrows[from].Actor // through the reference that narrowing returned
				if node.Calls == "" {
					node.Via = from // the actor is chosen by a parameter, per conversation
				}
			}
			if node.Calls == "system" {
				nr := parseNarrow(n.In)
				if nr != nil {
					node.Narrows, narrows[name] = nr, nr
				}
			}
			o.Nodes = append(o.Nodes, node)
		}
		out = append(out, destination{Actor: actor, Title: c.Title, Schema: p.Contract, Outline: o})
	}
	return out, nil
}

func parseNarrow(in json.RawMessage) *narrowing {
	var raw struct {
		Narrow map[string]json.RawMessage `json:"narrow"`
	}
	if json.Unmarshal(in, &raw) != nil || raw.Narrow == nil {
		return nil
	}
	var nr narrowing
	if actor, ok := raw.Narrow["actor"]; ok {
		if json.Unmarshal(actor, &nr.Actor) != nil {
			// An actor read from the input is a parameter of the procedure,
			// chosen per conversation: think's model (proposal 39).
			if path := bindingPath(actor); len(path) > 1 && path[0] == "input" {
				nr.Param = path[len(path)-1]
			} else {
				nr.Actor = bindingNode(actor)
			}
		}
	}
	if reach, ok := raw.Narrow["reach"]; ok {
		json.Unmarshal(reach, &nr.Reach)
	}
	if nr.Reach == nil {
		nr.Reach = []string{}
	}
	return &nr
}

// bindingNode is the node a {"$": [node, ...]} binding reads from.
func bindingNode(b json.RawMessage) string {
	if path := bindingPath(b); len(path) > 0 {
		return path[0]
	}
	return ""
}

// bindingPath is the whole path a {"$": [...]} binding reads.
func bindingPath(b json.RawMessage) []string {
	var ref struct {
		Path []string `json:"$"`
	}
	if json.Unmarshal(b, &ref) != nil {
		return nil
	}
	return ref.Path
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// destinationFor is the root a conversation opened with the given title runs:
// the first root when none is named.
func (h *webhost) destinationFor(title string) (destination, bool) {
	if title == "" && len(h.destinations) > 0 {
		return h.destinations[0], true
	}
	for _, d := range h.destinations {
		if d.Title == title {
			return d, true
		}
	}
	return destination{}, false
}

// normalizeRoot is the procedure and model a conversation runs. The model is a
// parameter of one procedure (proposal 39); conversations recorded before
// that named the Claude model as a procedure of its own, "think-claude".
func normalizeRoot(procedure, model string) (string, string) {
	if procedure == "think-claude" {
		return "think", "claude"
	}
	return procedure, model
}
