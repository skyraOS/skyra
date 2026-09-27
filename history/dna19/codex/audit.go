package main

import (
	"encoding/json"
	"os"
)

// Opt-in diagnostics contain correlation metadata, never parent tokens, signing
// material, prompts or tool contents. They do not participate in Move authority.
func (a *actor) audit(kind string, s *session, oldThread string) {
	if os.Getenv("SKYRA_CODEX_AUDIT") != "1" {
		return
	}
	value := map[string]any{"codex_audit": kind, "owner": s.ref.Owner, "parent": s.ref.Sequence, "thread": s.thread, "turn": s.turn, "tools": s.tools}
	if oldThread != "" {
		value["old_thread"] = oldThread
	}
	if s.work != nil {
		value["command"] = s.work.Invoke.Frame
	}
	_ = json.NewEncoder(os.Stderr).Encode(value)
}
