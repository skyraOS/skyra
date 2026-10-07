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
}

// RenderClaudeResident renders a Seatbelt profile for the claude-start
// resident. Unlike the appserver profile, Claude's CLI manages its own
// network and auth; the profile allows outbound TCP (loopback for the MCP
// bridge and HTTPS for the Anthropic API) and exec of the claude binary.
func RenderClaudeResident(h Horizon, floor string, p ClaudePaths) (string, error) {
	if h.Surface == nil || h.Surface.ResidentProfile != "darwin-claude-v1" {
		return "", fmt.Errorf("invalid claude resident profile")
	}
	for _, path := range []string{p.Home, p.Scratch, p.ClaudeBin, p.Auth, p.Cache, h.Surface.Path} {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("resident path is not absolute: %q", path)
		}
	}
	var b strings.Builder
	b.WriteString("(version 1)\n(deny default)\n")
	b.WriteString("(allow process-fork)\n")
	fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", h.Surface.Path)
	fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", p.ClaudeBin)
	b.WriteString("(allow process-exec (literal \"/usr/lib/dyld\"))\n")
	b.WriteString("(allow file-read-metadata)\n(allow file-map-executable)\n")
	b.WriteString("(allow system-socket (require-all (socket-domain AF_SYSTEM) (socket-protocol 2)))\n")
	b.WriteString("(allow ipc-posix-shm* (ipc-posix-name \"com.apple.AppleDatabaseChanged\"))\n")
	b.WriteString("(allow sysctl-read)\n")
	for _, path := range []string{"/System", "/usr/lib", "/usr/share/zoneinfo", "/usr/share/icu", "/private/etc", filepath.Join(p.Cache, "mds"), "/private/var/db/mds", "/private/var/run/com.apple.security.cryptexd", "/private/var/db/dyld", "/private/var/db/timezone", p.Home, p.Scratch, p.Auth} {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", path)
	}
	claudeDir := filepath.Dir(p.ClaudeBin)
	fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", claudeDir)
	if p.ClaudeLinkDir != "" && p.ClaudeLinkDir != claudeDir {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", p.ClaudeLinkDir)
	}
	for _, path := range []string{"/", filepath.Dir(h.Surface.Path), h.Surface.Path, p.ClaudeBin, "/dev/null", "/dev/random", "/dev/urandom"} {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", path)
	}
	for _, path := range []string{p.Home, p.Scratch} {
		fmt.Fprintf(&b, "(allow file-write* (subpath %q))\n", path)
	}
	b.WriteString("(allow file-write* (literal \"/dev/null\"))\n")
	fmt.Fprintf(&b, "(allow file-write-data (literal %q))\n", filepath.Join(p.Cache, "mds", "mds.lock"))
	for _, service := range []string{"com.apple.trustd.agent", "com.apple.trustd", "com.apple.system.notification_center", "com.apple.SystemConfiguration.DNSConfiguration", "com.apple.bsd.dirhelper", "com.apple.system.opendirectoryd.membership", "com.apple.ocspd", "com.apple.system.opendirectoryd.libinfo", "com.apple.SystemConfiguration.configd", "com.apple.networkd", "com.apple.SecurityServer"} {
		fmt.Fprintf(&b, "(allow mach-lookup (global-name %q))\n", service)
	}
	// Loopback TCP for the MCP bridge and outbound HTTPS for the Anthropic API.
	b.WriteString("(allow network-outbound (remote tcp))\n")
	b.WriteString("(allow network-inbound (local tcp))\n")
	b.WriteString("(allow network-bind (local tcp))\n")
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	return b.String(), nil
}
