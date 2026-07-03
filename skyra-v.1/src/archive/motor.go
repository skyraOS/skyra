//go:build legacy

package reality

import "skyra-v1/src/debug"

// Complete gets the slice and calls Fire.
func (s *Skyra) Complete(r *Relation) {
	sl := r.GetSlice(s.BeingID)
	sl.mu.Lock()
	if sl.Intent.IsEmpty() {
		sl.mu.Unlock()
		return
	}
	sl.mu.Unlock()

	s.Fire(&r.GetSlice(s.BeingID).Intent, r)
}

// Fire iterates the intent graph entries up to the first delimiter,
// calls Dispatch for each, writes outcomes to PreviousActions, removes completed entries.
func (s *Skyra) Fire(ig *IntentGraph, r *Relation) {
	phase := ig.NextPhase()
	if len(phase) == 0 {
		return
	}

	for _, key := range phase {
		entry := ig.Entries[key]
		debug.Being(s.BeingID, "motor", "dispatch:", entry.Primitive, "->", entry.Target)

		result := s.Dispatch(entry.Primitive, r, entry.Target)

		entry.Status = "success"
		if result != "" && len(result) > 10 && result[:10] == "unmatched " {
			entry.Status = "failed"
		}
		entry.Result = result

		out := s.Create(&Relation{}).(*Skyra)
		out.Content = result
		out.Description = entry.Primitive + ":" + entry.Target

		r.mu.RLock()
		if targetEntity, ok := r.World[entry.Target]; ok {
			out.Relationships[entry.Target] = targetEntity
		}
		ahID := "action:" + s.BeingID
		if ah, ok := r.ActionHistories[ahID]; ok {
			ah.Bind(out.ID())
			if targetEntity, ok := r.World[entry.Target]; ok {
				ah.Bind(targetEntity.ID())
			}
		}
		r.mu.RUnlock()

		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		sl.PreviousActions = append(sl.PreviousActions, out)
		sl.mu.Unlock()

		ig.Remove(key)
		debug.Being(s.BeingID, "motor", "completed:", entry.Primitive, "->", entry.Target, "result:", result)
	}

	// remove the delimiter after the phase completes
	if len(ig.Keys) > 0 {
		first := ig.Entries[ig.Keys[0]]
		if first.Primitive == "delimiter" {
			ig.Remove(ig.Keys[0])
		}
	}
}

// Dispatch resolves a primitive name to its function, executes it, returns a result string.
func (s *Skyra) Dispatch(name string, r *Relation, target string) string {
	fn, ok := s.Primitives[name]
	if !ok {
		return "unmatched primitive: " + name
	}
	fn(r, target)
	return "success"
}
