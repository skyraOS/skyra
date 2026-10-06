// Package prompt is Codex's whole base prompt (proposal 38). It replaces the
// provider's default, which describes tools and approval modes this runtime
// never offers. A thread's prompt is the header that invoked it, nothing else.
package prompt

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
