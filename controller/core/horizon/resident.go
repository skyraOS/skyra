package horizon

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// Provider is the model CLI a darwin-provider-v1 resident runs (proposal 47).
// Codex and Claude take one path: one profile, rendered by RenderResident
// from the same Horizon rules. What differs between them is this, pinned in
// the surface, so a change to it is a different surface.
type Provider struct {
	// Name selects what the runtime itself pins for the provider: the one
	// version its profile is verified for and the config it writes.
	Name string `json:"name"`
	// Program is the model CLI. When it is the surface's own path, the
	// resident is the CLI, run from where it was installed (Codex finds its
	// package beside it). Otherwise the resident is a wrapper, run from its
	// verified copy, that execs Program.
	Program       string `json:"program"`
	ProgramSHA256 string `json:"program_sha256"`
	// Reads are files of the program's own package it reads at startup.
	Reads []string `json:"reads,omitempty"`
	// Auth is how the user's existing login is reached. "file": AuthPath,
	// linked read-only into the private home. "keychain": the one
	// generic-password item Keychain, served from outside the sandbox by
	// Helper, whose copy inside the sandbox is the only `security` there.
	Auth         string `json:"auth"`
	AuthPath     string `json:"auth_path,omitempty"`
	Keychain     string `json:"keychain,omitempty"`
	Helper       string `json:"helper,omitempty"`
	HelperSHA256 string `json:"helper_sha256,omitempty"`
}

// ResidentPaths are names allocated by the runtime for one invocation. They
// are never supplied by the emitting target.
type ResidentPaths struct {
	Home, Scratch, Work, Cache string
	Config                     string // the runtime-written config in Home: readable, never writable
	Catalog                    string // a runtime-written file beside Home the provider reads, or ""
	Auth                       string // file auth: the resolved credential file
	Helper                     string // keychain auth: the verified copy of the helper, named security
	Broker                     string // keychain auth: the broker's socket
	ProxyPort                  int
	BridgePort                 int // a loopback port the resident serves its own tools on, or 0
}

// ResidentProgram is what the runtime execs for a provider resident: the
// model CLI from where it was installed when the resident is the CLI, and the
// resident's verified copy (43 D5) when it is a wrapper.
func ResidentProgram(h Horizon, floor string) string {
	if p := h.Surface.Provider; p != nil && p.Program == h.Surface.Path {
		return h.Surface.Path
	}
	return Verified(h, floor)
}

// RenderResident renders the darwin-provider-v1 profile for the Horizon's
// provider with this invocation's private paths, its one proxy endpoint and,
// for a provider that serves its own tools, its one bridge port. The network
// destinations beyond the proxy are its registered allowlist, carried by the
// Surface.
//
// The Horizon reaches every provider through the same rules: none of its
// shared grants render but images, below. A provider's native reads would be
// observations with no record. That exclusion lasts until the resident fields
// leave the surface (next/25, item 3) and a provider's Horizon is narrowed
// through the room (next/26).
func RenderResident(h Horizon, floor string, p ResidentPaths) (string, error) {
	s := h.Surface
	if s == nil || s.ResidentProfile != "darwin-provider-v1" || s.Provider == nil || p.ProxyPort < 1 || p.ProxyPort > 65535 || p.BridgePort < 0 || p.BridgePort > 65535 {
		return "", fmt.Errorf("invalid resident profile or proxy")
	}
	v := s.Provider
	exe := ResidentProgram(h, floor)
	required := []string{p.Home, p.Scratch, p.Work, p.Cache, p.Config, v.Program, exe, floor}
	switch v.Auth {
	case "file":
		required = append(required, p.Auth)
	case "keychain":
		required = append(required, p.Helper, p.Broker)
	default:
		return "", fmt.Errorf("unknown provider auth %q", v.Auth)
	}
	if p.Catalog != "" {
		required = append(required, p.Catalog)
	}
	for _, path := range append(required, v.Reads...) {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("resident path is not absolute: %q", path)
		}
	}
	var b strings.Builder
	b.WriteString("(version 1)\n(deny default)\n")
	b.WriteString("(allow process-fork)\n")
	programs := []string{exe}
	if v.Program != exe {
		programs = append(programs, v.Program)
	}
	if v.Auth == "keychain" {
		programs = append(programs, p.Helper)
	}
	for _, program := range append(programs, "/usr/lib/dyld") {
		fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", program)
	}
	b.WriteString("(allow file-read-metadata)\n(allow file-map-executable)\n")
	b.WriteString("(allow system-socket (require-all (socket-domain AF_SYSTEM) (socket-protocol 2)))\n")
	b.WriteString("(allow ipc-posix-shm* (ipc-posix-name \"com.apple.AppleDatabaseChanged\"))\n")
	b.WriteString("(allow sysctl-read)\n")
	for _, path := range []string{"/System", "/usr/lib", "/usr/share/zoneinfo", "/usr/share/icu", "/private/etc", filepath.Join(p.Cache, "mds"), "/private/var/db/mds", "/private/var/run/com.apple.security.cryptexd", "/private/var/db/dyld", "/private/var/db/timezone", p.Home, p.Scratch, p.Work} {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", path)
	}
	literal := append([]string{"/", filepath.Dir(v.Program)}, programs...)
	literal = append(literal, v.Reads...)
	if v.Auth == "file" {
		literal = append(literal, p.Auth)
	}
	if p.Catalog != "" {
		literal = append(literal, p.Catalog)
	}
	for _, path := range append(literal, "/dev/null", "/dev/random", "/dev/urandom") {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", path)
	}
	// The one hole, chosen for the experience until Skyra has its own way to
	// show an image (proposal 38): Codex's view_image may read image files,
	// and only image files, under the Horizon's read grants. What it reads is
	// not recorded in the lineage. Claude, with its native tools off, never
	// asks; the rule is the same for both.
	// Images pasted with a message are kept in the conversation's scratch,
	// so it reads images there too.
	grants := h.Authority.Read
	if h.Authority.Scratch != "" {
		grants = append(slices.Clone(grants), PathGrant{Path: h.Authority.Scratch, Scope: "subpath"})
	}
	if h.Delegated {
		grants = nil
	}
	for _, g := range grants {
		if rule := imageRead(g); rule != "" {
			b.WriteString(rule)
		}
	}
	for _, path := range []string{p.Home, p.Scratch} {
		fmt.Fprintf(&b, "(allow file-write* (subpath %q))\n", path)
	}
	b.WriteString("(allow file-write* (literal \"/dev/null\"))\n")
	fmt.Fprintf(&b, "(allow file-write-data (literal %q))\n", filepath.Join(p.Cache, "mds", "mds.lock"))
	denied := []string{p.Config}
	if v.Auth == "file" {
		denied = append(denied, p.Auth)
	}
	for _, path := range denied {
		fmt.Fprintf(&b, "(deny file-write* (literal %q))\n", path)
	}
	for _, service := range []string{"com.apple.trustd.agent", "com.apple.trustd", "com.apple.system.notification_center", "com.apple.SystemConfiguration.DNSConfiguration", "com.apple.bsd.dirhelper", "com.apple.system.opendirectoryd.membership", "com.apple.ocspd", "com.apple.system.opendirectoryd.libinfo", "com.apple.SystemConfiguration.configd", "com.apple.networkd", "com.apple.SecurityServer"} {
		fmt.Fprintf(&b, "(allow mach-lookup (global-name %q))\n", service)
	}
	fmt.Fprintf(&b, "(allow network-outbound (remote ip %q))\n", fmt.Sprintf("localhost:%d", p.ProxyPort))
	if p.BridgePort != 0 {
		bridge := fmt.Sprintf("localhost:%d", p.BridgePort)
		fmt.Fprintf(&b, "(allow network-bind network-inbound (local ip %q))\n", bridge)
		fmt.Fprintf(&b, "(allow network-outbound (remote ip %q))\n", bridge)
	}
	if v.Auth == "keychain" {
		fmt.Fprintf(&b, "(allow network-outbound (remote unix-socket (path-literal %q)))\n", p.Broker)
	}
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
	if exe != s.Path {
		// After the floor's deny, so the copy the runtime checked can be read to run.
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", exe)
	}
	return b.String(), nil
}

// imageExtensions are the files view_image may read, in any case.
var imageExtensions = []string{"png", "jpg", "jpeg", "gif", "webp", "heic", "bmp", "tif", "tiff"}

// imageRead is the rule letting the resident read images under one read
// grant: any image below a subpath grant, or a literal grant that is an image.
// A path that cannot be written safely into a rule grants nothing.
func imageRead(g PathGrant) string {
	if !filepath.IsAbs(g.Path) || strings.ContainsAny(g.Path, "\"\\\n") {
		return ""
	}
	var exts []string
	for _, e := range imageExtensions {
		var any strings.Builder
		for _, r := range e {
			fmt.Fprintf(&any, "[%c%c]", unicode.ToLower(r), unicode.ToUpper(r))
		}
		exts = append(exts, any.String())
	}
	suffix := `\.(` + strings.Join(exts, "|") + `)$`
	quoted := regexp.QuoteMeta(filepath.Clean(g.Path))
	switch g.Scope {
	case "subpath":
		return fmt.Sprintf("(allow file-read* (regex #\"^%s/.*%s\"))\n", quoted, suffix)
	case "literal":
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(g.Path)), ".")
		for _, e := range imageExtensions {
			if ext == e {
				return fmt.Sprintf("(allow file-read* (literal %q))\n", filepath.Clean(g.Path))
			}
		}
	}
	return ""
}
