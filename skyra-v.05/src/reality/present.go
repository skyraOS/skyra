package reality

import (
	"fmt"
	"sort"
	"strings"
)

// Present — assembles the being's current state into sections for inference.
// Each section is a header. Empty sections are dropped. Primitives are always present.
func Present(r *Relation, beingID string) string {
	sl := r.GetSlice(beingID)
	sl.mu.Lock()
	defer sl.mu.Unlock()

	var sections []string

	// primitives — always present
	sections = append(sections, "## Primitives\nspeak")

	// surroundings — what the being can see right now
	if len(sl.Surroundings) > 0 {
		type entry struct {
			id    string
			score float64
		}
		var entries []entry
		for id, score := range sl.Surroundings {
			entries = append(entries, entry{id, score})
		}
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].score > entries[j].score
		})
		var lines []string
		for _, e := range entries {
			lines = append(lines, fmt.Sprintf("%s: %.2f", e.id, e.score))
		}
		sections = append(sections, "## Surroundings\n"+strings.Join(lines, "\n"))
	}

	// threads — active threads the being is listening to
	r.mu.RLock()
	var threadSections []string
	for threadID, thread := range r.Threads {
		if thread.BeingID != beingID {
			continue
		}
		thread.mu.Lock()
		if len(thread.Members) > 0 {
			type entry struct {
				id       string
				strength float64
			}
			var entries []entry
			for id, strength := range thread.Members {
				entries = append(entries, entry{id, strength})
			}
			sort.Slice(entries, func(i, j int) bool {
				return entries[i].strength > entries[j].strength
			})
			var lines []string
			for _, e := range entries {
				lines = append(lines, fmt.Sprintf("%s: %.2f", e.id, e.strength))
			}
			threadSections = append(threadSections, "## Thread: "+threadID+"\n"+strings.Join(lines, "\n"))
		}
		thread.mu.Unlock()
	}
	r.mu.RUnlock()
	sections = append(sections, threadSections...)

	// thoughts — accumulated during traversal
	if len(sl.TrainOfThought) > 0 {
		var lines []string
		for _, t := range sl.TrainOfThought {
			if sk, ok := t.(*Skyra); ok && sk.Content != "" {
				lines = append(lines, sk.Content)
			}
		}
		if len(lines) > 0 {
			sections = append(sections, "## Train of Thought\n"+strings.Join(lines, "\n"))
		}
	}

	// goal — only if set
	if sl.Goal != "" {
		sections = append(sections, "## Goal\n"+sl.Goal)
	}

	// intent — only if set
	if !sl.Intent.IsEmpty() {
		sections = append(sections, "## Intent\n"+sl.Intent.Render())
	}

	return strings.Join(sections, "\n\n")
}
