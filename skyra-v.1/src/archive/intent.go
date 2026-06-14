package reality

import (
	"fmt"
	"strings"
)

type IntentEntry struct {
	Primitive string // look, listen, engage, delimiter
	Target    string // reality ID, content, or empty
	Status    string // pending, success, failed
	Result    string // what came back — success message or error message
}

type IntentGraph struct {
	Keys    []string               // ordered action IDs — insertion order preserved
	Entries map[string]IntentEntry // keyed by action ID
}

func NewIntentGraph() IntentGraph {
	return IntentGraph{
		Keys:    []string{},
		Entries: make(map[string]IntentEntry),
	}
}

// NextPhase returns keys up to the first delimiter.
func (ig *IntentGraph) NextPhase() []string {
	var phase []string
	for _, key := range ig.Keys {
		entry := ig.Entries[key]
		if entry.Primitive == "delimiter" {
			break
		}
		phase = append(phase, key)
	}
	return phase
}

// Remove removes an entry from Keys and Entries.
func (ig *IntentGraph) Remove(key string) {
	delete(ig.Entries, key)
	for i, k := range ig.Keys {
		if k == key {
			ig.Keys = append(ig.Keys[:i], ig.Keys[i+1:]...)
			return
		}
	}
}

// Render formats the intent graph for Present() display.
func (ig *IntentGraph) Render() string {
	if len(ig.Keys) == 0 {
		return ""
	}
	var parts []string
	for _, key := range ig.Keys {
		entry := ig.Entries[key]
		if entry.Primitive == "delimiter" {
			parts = append(parts, "<wait></wait>")
			continue
		}
		if entry.Target != "" {
			parts = append(parts, fmt.Sprintf("<%s>%s</%s>", entry.Primitive, entry.Target, entry.Primitive))
		} else {
			parts = append(parts, fmt.Sprintf("<%s></%s>", entry.Primitive, entry.Primitive))
		}
	}
	return strings.Join(parts, "\n")
}

func (ig *IntentGraph) IsEmpty() bool {
	return len(ig.Keys) == 0
}

// ParseIntent parses LLM output into a fresh IntentGraph.
func ParseIntent(raw string) IntentGraph {
	ig := NewIntentGraph()
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ig
	}

	phases := strings.Split(raw, "<wait></wait>")
	for i, phase := range phases {
		phase = strings.TrimSpace(phase)
		if phase != "" {
			parsePhaseEntries(phase, &ig)
		}
		if i < len(phases)-1 {
			key := NewID()
			ig.Keys = append(ig.Keys, key)
			ig.Entries[key] = IntentEntry{Primitive: "delimiter"}
		}
	}
	return ig
}

func parsePhaseEntries(phase string, ig *IntentGraph) {
	primitives := []string{"look", "listen", "engage"}
	for _, p := range primitives {
		open := "<" + p + ">"
		close := "</" + p + ">"
		remaining := phase
		for {
			start := strings.Index(remaining, open)
			if start == -1 {
				break
			}
			start += len(open)
			end := strings.Index(remaining[start:], close)
			if end == -1 {
				break
			}
			target := strings.TrimSpace(remaining[start : start+end])
			key := NewID()
			ig.Keys = append(ig.Keys, key)
			ig.Entries[key] = IntentEntry{
				Primitive: p,
				Target:    target,
				Status:    "pending",
			}
			remaining = remaining[start+end+len(close):]
		}
	}
}
