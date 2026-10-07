package horizon

import (
	"fmt"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// ResidentPaths are names allocated by the runtime for one invocation. They
// are never supplied by the emitting target. The profile template itself is
// pinned in the released Surface.
type ResidentPaths struct {
	Home, Scratch, Work, Auth, Cache string
	ProxyPort                        int
}

// RenderResident renders the named app-server profile in the Horizon with
// this invocation's private paths and its single proxy endpoint. The network
// destinations beyond that endpoint are controlled by the registered proxy
// allowlist, which is also carried by the Surface.
//
// It renders none of the Horizon's shared grants. The app-server is the one
// program whose authority is its pinned profile template, not its Horizon:
// its native reads would be observations with no record. That exclusion lasts
// until the resident fields leave the surface (next/25, item 3) and the
// app-server's Horizon is narrowed through the room (next/26).
func RenderResident(h Horizon, floor string, p ResidentPaths) (string, error) {
	if h.Surface == nil || h.Surface.ResidentProfile != "darwin-appserver-v1" || p.ProxyPort < 1 || p.ProxyPort > 65535 {
		return "", fmt.Errorf("invalid resident profile or proxy")
	}
	for _, path := range []string{p.Home, p.Scratch, p.Work, p.Auth, p.Cache, h.Surface.Path} {
		if !filepath.IsAbs(path) {
			return "", fmt.Errorf("resident path is not absolute: %q", path)
		}
	}
	var b strings.Builder
	b.WriteString("(version 1)\n(deny default)\n")
	b.WriteString("(allow process-fork)\n")
	fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", h.Surface.Path)
	b.WriteString("(allow process-exec (literal \"/usr/lib/dyld\"))\n")
	b.WriteString("(allow file-read-metadata)\n(allow file-map-executable)\n")
	b.WriteString("(allow system-socket (require-all (socket-domain AF_SYSTEM) (socket-protocol 2)))\n")
	b.WriteString("(allow ipc-posix-shm* (ipc-posix-name \"com.apple.AppleDatabaseChanged\"))\n")
	b.WriteString("(allow sysctl-read)\n")
	for _, path := range []string{"/System", "/usr/lib", "/usr/share/zoneinfo", "/usr/share/icu", "/private/etc", filepath.Join(p.Cache, "mds"), "/private/var/db/mds", "/private/var/run/com.apple.security.cryptexd", "/private/var/db/dyld", "/private/var/db/timezone", p.Home, p.Scratch, p.Work} {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", path)
	}
	for _, path := range []string{"/", filepath.Dir(h.Surface.Path), h.Surface.Path, filepath.Join(filepath.Dir(filepath.Dir(h.Surface.Path)), "codex-package.json"), p.Auth, filepath.Join(filepath.Dir(p.Home), "models.json"), "/dev/null", "/dev/random", "/dev/urandom"} {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", path)
	}
	// The one hole, chosen for the experience until Skyra has its own way to
	// show an image (proposal 38): Codex's view_image may read image files,
	// and only image files, under the Horizon's read grants. What it reads is
	// not recorded in the lineage.
	// Images pasted with a message are kept in the conversation's scratch,
	// so it reads images there too.
	grants := h.Authority.Read
	if h.Authority.Scratch != "" {
		grants = append(slices.Clone(grants), PathGrant{Path: h.Authority.Scratch, Scope: "subpath"})
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
	for _, path := range []string{filepath.Join(p.Home, "config.toml"), p.Auth} {
		fmt.Fprintf(&b, "(deny file-write* (literal %q))\n", path)
	}
	for _, service := range []string{"com.apple.trustd.agent", "com.apple.trustd", "com.apple.system.notification_center", "com.apple.SystemConfiguration.DNSConfiguration", "com.apple.bsd.dirhelper", "com.apple.system.opendirectoryd.membership", "com.apple.ocspd", "com.apple.system.opendirectoryd.libinfo", "com.apple.SystemConfiguration.configd", "com.apple.networkd", "com.apple.SecurityServer"} {
		fmt.Fprintf(&b, "(allow mach-lookup (global-name %q))\n", service)
	}
	fmt.Fprintf(&b, "(allow network-outbound (remote ip %q))\n", fmt.Sprintf("localhost:%d", p.ProxyPort))
	fmt.Fprintf(&b, "(deny file-read* file-write* (subpath %q))\n", floor)
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
