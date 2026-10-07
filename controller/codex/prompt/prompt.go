// Package prompt is Codex's whole base prompt (proposal 38). It replaces the
// provider's default, which describes tools and approval modes this runtime
// never offers. A thread's prompt identifies its header and current tool roots.
package prompt

import "strings"

// Prefix is the part of every prompt that does not depend on the header.
const Prefix = "You are inside Skyra, a cognitive architecture"

// Base is the base prompt for one invocation. With no header (a direct
// conversation, until the user adapter routes by selection) it is the prefix.
func Base(header string) string {
	if header == "" {
		return Prefix + "."
	}
	return Prefix + ", fulfilling: " + header
}

func Room(header, read, scratch string, repository bool) string {
	base := Base(header)
	if read == "" && scratch == "" && !repository {
		return base
	}
	workspace, _, _ := strings.Cut(read, ":")
	if workspace != "" {
		base += "\nTools resolve relative paths against " + workspace + "."
	}
	if scratch != "" {
		base += " Scratch is " + scratch + " (use scratch: paths with read/find)."
	}
	if repository {
		base += "\nrepo.write edits repository files and supports guarded old/new edits. Git is bounded to this repository and its registered worktrees; use roots for the current permission summary and changes:true for this conversation's recorded changes. Scratch writes are saved on the separate skyra/ideas ref with agent authorship, never pushed automatically."
	} else {
		base += "\nThe workspace is read-only; write saves only in scratch. Only the offered tools are available."
	}
	return base
}
