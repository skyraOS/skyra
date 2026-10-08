package machine

import (
	"encoding/json"
	"errors"
	"slices"

	"skyra/controller/core/horizon"
	"skyra/controller/core/record"
)

// runOf is the Run effect for a Horizon: its surface, under the sandbox the
// Horizon renders to, with an environment that names only the sockets its
// surface was granted, and the slot it returns to.
func runOf(h horizon.Horizon, stdin []byte, floor, socket string) Run {
	// A target returns by addressing the slot its frame discharges. Being told
	// it is information, not authority: intake derives the return from the
	// emitter's own Horizon, whatever the target believes.
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
	if s := h.Authority.Scratch; s != "" {
		env = append(env, "SKYRA_SCRATCH="+s)
	}
	if dir := horizon.StateDir(h, floor); dir != "" {
		env = append(env, "SKYRA_STATE="+dir) // the actor's own state (next/32): the profile grants it
	}
	if socket != "" {
		// A connection surface reaches its edge's resident by one socket, and
		// is granted exactly that path: nothing else of the holder's directory.
		env = append(env, "SKYRA_RESIDENT="+socket)
		granted := h
		granted.Authority.Connect = append(slices.Clone(h.Authority.Connect), socket)
		return Run{Surface: *h.Surface, Stdin: stdin, Profile: horizon.Render(granted, floor), Env: env, Resident: socket}
	}
	return Run{Surface: *h.Surface, Stdin: stdin, Profile: horizon.Render(h, floor), Env: env}
}

// edgeSocket is where the resident of h's edge is reached (proposal 38). A
// connection is an edge between two actors: the adapter that invoked this
// surface, and whoever invoked the adapter. Both are names from the Horizon,
// so a program cannot pick another edge, and the floor makes the edge this
// lineage's own. No frame is named, so every turn of a conversation reaches
// the same resident.
func edgeSocket(h horizon.Horizon, floor, dir string) (string, error) {
	if h.Resume == nil || h.Resume.Name == "" {
		return "", errors.New("a resident is reached only through the adapter that invoked it")
	}
	if dir == "" {
		return "", errors.New("no directory for edge sockets")
	}
	adapter, caller := h.Resume.Name, caller(*h.Resume)
	return dir + "/" + record.Sum([]byte(floor + "\x00" + caller + "\x00" + adapter))[:24] + ".sock", nil
}
