package horizon

import (
	"maps"
	"slices"
	"strings"

	"skyra/controller/core/contract"
)

// Internal names are execution machinery, never selectable work tools. Agent
// invocations retain the work ceiling, including when reached indirectly.
func (h Horizon) Internal(name string) bool {
	if name == "system" || strings.HasPrefix(name, "system.") {
		return true
	}
	for actor, s := range h.Reach {
		if s.Agent != nil && (name == actor || slices.Contains(s.Agent.Operations, name)) {
			return true
		}
	}
	return false
}

// AvailableTools resolves concrete actor-owned contracts. A bounded child can
// select only bindings already held; extensions never reset this ceiling.
func (h Horizon) AvailableTools() map[string]ToolBinding {
	if h.WorkTools != nil {
		return *h.WorkTools
	}

	out := map[string]ToolBinding{}
	reserved := map[string]bool{"remember": true, "procedure": true}
	names := slices.Sorted(maps.Keys(h.Reach))
	for _, name := range names {
		s := h.Reach[name]
		if h.Internal(name) || s.Resident || s.Connection || s.Host != "" || len(s.All()) == 0 || s.All()[0].Title == "remember" || !s.WorkTool {
			continue
		}
		out[name] = ToolBinding{Target: name, Contracts: s.All()}
		reserved[strings.ReplaceAll(name, ".", "_")] = true
	}
	for _, name := range names {
		s := h.Reach[name]
		all := s.All()
		if len(all) == 0 || all[0].Title != "remember" {
			continue
		}
		out["remember"] = ToolBinding{Target: name, Contracts: all[:1]}
		if len(all) > 1 {
			out["procedure"] = ToolBinding{Target: name, Contracts: all[1:]}
		}
		for _, c := range all[1:] {
			if !toolLabel(c.Title) || reserved[c.Title] {
				continue
			}
			reserved[c.Title] = true
			out[c.Title] = ToolBinding{Target: name, Contracts: []*contract.Schema{c}}
		}
		break // the same deterministic interpreter binding as the transports
	}

	return out
}

func (h Horizon) PermitsTool(target string, schema *contract.Schema) bool {
	if h.WorkTools == nil || h.Internal(target) {
		return true
	}
	want := string(contract.Canonical(schema))
	for _, b := range *h.WorkTools {
		if b.Target != target {
			continue
		}
		for _, c := range b.Contracts {
			if string(contract.Canonical(c)) == want {
				return true
			}
		}
	}
	return false
}

func toolLabel(name string) bool {
	if len(name) == 0 || len(name) > 64 {
		return false
	}
	for _, c := range name {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// ResolveTool accepts canonical actor identifiers and their transport labels.
func ResolveTool(available map[string]ToolBinding, name string) (ToolBinding, bool) {
	if b, ok := available[name]; ok {
		return b, true
	}
	label := strings.ReplaceAll(name, ".", "_")
	var found ToolBinding
	matches := 0
	for key, b := range available {
		if strings.ReplaceAll(key, ".", "_") == label {
			found = b
			matches++
		}
	}
	return found, matches == 1
}
