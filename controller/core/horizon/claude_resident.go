package horizon

import (
	"fmt"
	"path/filepath"
	"strings"
)

// ClaudePaths are the runtime-allocated paths for one Claude CLI invocation.
type ClaudePaths struct {
	Home, Scratch string
	ClaudeBin     string // resolved path to the claude binary
	ClaudeLinkDir string // directory containing the "claude" symlink
	Auth          string // resolved path to ~/.claude
	Cache         string // DARWIN_USER_CACHE_DIR
	UserHome      string // the user's real home directory
}

// RenderClaudeResident renders a Seatbelt profile for the claude-start
// resident. The claude CLI is the user's own authenticated application:
// its Bun runtime and macOS keychain auth require operations that the
// sandbox cannot individually enumerate, so the profile starts from
// (allow default) and restricts process-exec to only the claude binary
// and the conversation's .skyra state directory.
func RenderClaudeResident(h Horizon, floor string, p ClaudePaths) (string, error) {
	if h.Surface == nil || h.Surface.ResidentProfile != "darwin-claude-v1" {
		return "", fmt.Errorf("invalid claude resident profile")
	}
	if !filepath.IsAbs(floor) {
		return "", fmt.Errorf("resident floor is not absolute: %q", floor)
	}
	var b strings.Builder
	b.WriteString("(version 1)\n(allow default)\n")
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	if exe := Verified(h, floor); exe != "" {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", exe)
	}
	return b.String(), nil
}
