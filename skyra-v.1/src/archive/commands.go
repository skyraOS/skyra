//go:build legacy

package reality

import (
	"fmt"
	"strings"
)

type CommandOutput struct {
	From    string
	Content string
}

// HandleCommand intercepts user commands before they become impulses.
// Returns true if the input was a command and was handled.
func HandleCommand(input string, r *Relation, outputCh chan<- CommandOutput) bool {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return false
	}

	parts := strings.Fields(input)
	cmd := parts[0]

	switch cmd {
	case "/createThread":
		return cmdCreateThread(parts[1:], r, outputCh)
	case "/threads":
		return cmdListThreads(r, outputCh)
	default:
		outputCh <- CommandOutput{From: "system", Content: "unknown command: " + cmd}
		return true
	}
}

// /createThread -skyra -builder
// Creates a thread and binds the listed beings to it.
func cmdCreateThread(args []string, r *Relation, outputCh chan<- CommandOutput) bool {
	if len(args) == 0 {
		outputCh <- CommandOutput{From: "system", Content: "usage: /createThread -being1 -being2 ..."}
		return true
	}

	var beingIDs []string
	for _, arg := range args {
		beingIDs = append(beingIDs, strings.TrimPrefix(arg, "-"))
	}

	threadID := "thread:" + strings.Join(beingIDs, "+")

	r.mu.RLock()
	_, exists := r.Threads[threadID]
	r.mu.RUnlock()
	if exists {
		outputCh <- CommandOutput{From: "system", Content: "thread already exists: " + threadID}
		return true
	}

	thread := NewThread(threadID, beingIDs[0])

	r.mu.RLock()
	for _, id := range beingIDs {
		if entity, ok := r.World[id]; ok {
			thread.Bind(entity.ID())
			if sk, ok := entity.(*Skyra); ok {
				sk.Relationships[threadID] = thread
			}
		}
	}
	r.mu.RUnlock()

	r.mu.Lock()
	r.Threads[threadID] = thread
	r.mu.Unlock()

	outputCh <- CommandOutput{From: "system", Content: "created thread: " + threadID + " with " + strings.Join(beingIDs, ", ")}
	return true
}

// /threads — list all active threads
func cmdListThreads(r *Relation, outputCh chan<- CommandOutput) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.Threads) == 0 {
		outputCh <- CommandOutput{From: "system", Content: "no active threads"}
		return true
	}

	var lines []string
	for id, thread := range r.Threads {
		thread.mu.Lock()
		count := len(thread.Members)
		thread.mu.Unlock()
		lines = append(lines, fmt.Sprintf("%s (%s, %d members)", id, thread.BeingID, count))
	}

	outputCh <- CommandOutput{From: "system", Content: strings.Join(lines, "\n")}
	return true
}
