package machine

import (
	"encoding/json"

	"skyra/controller/core/boundary"
)

// runOf is the Run effect for a Boundary: its surface, under the sandbox the
// Boundary renders to, with an environment that names only the sockets its
// surface was granted, and the slot it returns to.
func runOf(h boundary.Boundary, stdin []byte, floor string) Run {
	// A target returns by addressing the slot its frame discharges. Being told
	// it is information, not authority: intake derives the return from the
	// emitter's own Boundary, whatever the target believes.
	env := []string{"PATH=/usr/bin:/bin", "SKYRA_FRAME=" + h.Frame}
	for _, sock := range h.Authority.Connect {
		env = append(env, "SKYRA_CONNECT="+sock)
	}
	if len(h.Authority.Connect) > 0 {
		sockets, _ := json.Marshal(h.Authority.Connect)
		env = append(env, "SKYRA_CONNECTIONS="+string(sockets))
	}
	// A target is told where its read grants are, as PATH lists directories.
	// This is information, not authority: the rendered profile enforces them.
	if reads := h.Authority.Read; len(reads) > 0 {
		paths := reads[0].Path
		for _, grant := range reads[1:] {
			paths += ":" + grant.Path
		}
		env = append(env, "SKYRA_READ="+paths)
	}
	if w := h.Authority.Workspace; w != "" {
		env = append(env, "SKYRA_WORKSPACE="+w) // where a relative read resolves
	}
	if s := h.Authority.Scratch; s != "" {
		env = append(env, "SKYRA_SCRATCH="+s)
	}
	if dir := boundary.StateDir(h, floor); dir != "" {
		env = append(env, "SKYRA_STATE="+dir) // the actor's own state (next/32): the profile grants it
	}
	return Run{Surface: *h.Surface, Exec: boundary.Verified(h, floor), Stdin: stdin, Profile: boundary.Render(h, floor), Env: env}
}
