package reality

import "skyra-v1/src/debug"

// PrimitiveLook — broad activation. Signal distributed across everything
// that activates in the space. Each node gets a fraction. Light bump across many.
// Populates surroundings — entries below threshold fall off.
func PrimitiveLook(s *Skyra) func(*Relation, string) {
	return func(r *Relation, target string) {
		r.mu.RLock()
		world := r.World
		r.mu.RUnlock()

		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		for id, entity := range world {
			score := entity.Activation(r)
			if score > 0.01 {
				sl.Surroundings[id] = score
			} else {
				delete(sl.Surroundings, id)
			}
		}

		for _, rel := range s.Relationships {
			score := rel.Activation(r)
			if score > 0.01 {
				sl.Surroundings[rel.ID()] = score
			} else {
				delete(sl.Surroundings, rel.ID())
			}
		}
		sl.mu.Unlock()
	}
}

// PrimitiveEngage — the being's action on the world. Makes the LLM call.
// Branches on target: entity → fire Realize, content → SetImpulse + thread bump.
func PrimitiveEngage(s *Skyra, call func(string, string) (string, error)) func(*Relation, string) {
	return func(r *Relation, target string) {
		// stub — LLM call with thread context + compressed understanding goes here
		debug.Being(s.BeingID, "engage", "target:", target)
	}
}

// PrimitiveListen — subscribe to a space. Opens thread_alignment gate to 1.0.
// Creates or joins a thread for the target.
func PrimitiveListen(s *Skyra) func(*Relation, string) {
	return func(r *Relation, target string) {
		threadID := "thread:" + s.BeingID + ":" + target
		r.mu.RLock()
		thread, exists := r.Threads[threadID]
		r.mu.RUnlock()

		if !exists {
			thread = NewThread(threadID, s.BeingID)
			r.mu.Lock()
			r.Threads[threadID] = thread
			r.mu.Unlock()
		}

		thread.Bind(s.id)
		s.Relationships[threadID] = thread
	}
}
