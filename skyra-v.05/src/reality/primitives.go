package reality

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

// PrimitiveEngage — targeted activation. Full signal on one target.
// Crosses the boundary, opens the connection. If the target is a place,
// the engage cascades into everything bound to it.
func PrimitiveEngage(s *Skyra) func(*Relation, string) {
	return func(r *Relation, target string) {
		r.mu.RLock()
		entity, ok := r.World[target]
		r.mu.RUnlock()
		if !ok {
			for _, rel := range s.Relationships {
				if rel.ID() == target {
					entity = rel
					ok = true
					break
				}
			}
		}
		if !ok {
			return
		}

		entity.Realize(r)
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

// PrimitiveSpeak — emit into the subscribed space. Untargeted.
// Bumps thread_strength on touched nodes. Resets thread_alignment decay.
func PrimitiveSpeak(s *Skyra) func(*Relation, string) {
	return func(r *Relation, content string) {
		sl := r.GetSlice(s.BeingID)
		sl.mu.Lock()
		var touchedIDs []string
		for _, t := range sl.TrainOfThought {
			touchedIDs = append(touchedIDs, t.ID())
		}
		sl.mu.Unlock()

		r.mu.RLock()
		for _, thread := range r.Threads {
			if thread.BeingID == s.BeingID {
				thread.Speak(touchedIDs)
			}
		}
		r.mu.RUnlock()
	}
}
