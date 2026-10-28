package service

import (
	"fmt"
	"path/filepath"
	"strings"
)

// render is a session's Seatbelt profile, from its program's registration and
// the paths the host allocated: deny by default, exec of the program alone
// (and, for a wrapper, of its CLI and helper), its home, scratch and work. A
// provider program also gets the system reads TLS needs, its one proxy port,
// its login and, if it serves its own tools, its one bridge port. Nothing in
// it comes from a client.
func render(p *Program, exe string, x Paths) (string, error) {
	provider := p.Provider != ""
	if provider && (x.ProxyPort < 1 || x.ProxyPort > 65535) || x.BridgePort < 0 || x.BridgePort > 65535 {
		return "", fmt.Errorf("invalid proxy or bridge port")
	}
	required := []string{x.Home, x.Scratch, x.Work, exe}
	if provider {
		required = append(required, x.Cache, x.Config, p.CLI)
		switch p.Auth {
		case "file":
			required = append(required, x.Auth)
		case "keychain":
			required = append(required, x.Helper, x.Broker)
		}
		if x.Catalog != "" {
			required = append(required, x.Catalog)
		}
		required = append(required, p.Reads...)
	}
	for _, path := range required {
		if !filepath.IsAbs(path) || strings.ContainsAny(path, "\"\\\n") {
			return "", fmt.Errorf("profile path is not absolute and plain: %q", path)
		}
	}
	var b strings.Builder
	b.WriteString("(version 1)\n(deny default)\n")
	b.WriteString("(allow process-fork)\n")
	programs := []string{exe}
	if provider && p.CLI != exe {
		programs = append(programs, p.CLI)
	}
	if p.Auth == "keychain" {
		programs = append(programs, x.Helper)
	}
	for _, program := range append(programs, "/usr/lib/dyld") {
		fmt.Fprintf(&b, "(allow process-exec (literal %q))\n", program)
	}
	b.WriteString("(allow file-read-metadata)\n(allow file-map-executable)\n")
	b.WriteString("(allow sysctl-read)\n")
	reads := []string{"/System", "/usr/lib", "/private/etc"}
	if provider {
		b.WriteString("(allow system-socket (require-all (socket-domain AF_SYSTEM) (socket-protocol 2)))\n")
		b.WriteString("(allow ipc-posix-shm* (ipc-posix-name \"com.apple.AppleDatabaseChanged\"))\n")
		reads = append(reads, "/usr/share/zoneinfo", "/usr/share/icu", filepath.Join(x.Cache, "mds"), "/private/var/db/mds", "/private/var/run/com.apple.security.cryptexd", "/private/var/db/dyld", "/private/var/db/timezone")
	}
	for _, path := range append(reads, x.Home, x.Scratch, x.Work) {
		fmt.Fprintf(&b, "(allow file-read* (subpath %q))\n", path)
	}
	literal := append([]string{"/"}, programs...)
	if provider {
		literal = append(literal, filepath.Dir(p.CLI))
		literal = append(literal, p.Reads...)
		if p.Auth == "file" {
			literal = append(literal, x.Auth)
		}
		if x.Catalog != "" {
			literal = append(literal, x.Catalog)
		}
	}
	for _, path := range append(literal, "/dev/null", "/dev/random", "/dev/urandom") {
		fmt.Fprintf(&b, "(allow file-read* (literal %q))\n", path)
	}
	for _, path := range []string{x.Home, x.Scratch} {
		fmt.Fprintf(&b, "(allow file-write* (subpath %q))\n", path)
	}
	b.WriteString("(allow file-write* (literal \"/dev/null\"))\n")
	if !provider {
		return b.String(), nil
	}
	fmt.Fprintf(&b, "(allow file-write-data (literal %q))\n", filepath.Join(x.Cache, "mds", "mds.lock"))
	denied := []string{x.Config}
	if p.Auth == "file" {
		denied = append(denied, x.Auth)
	}
	for _, path := range denied {
		fmt.Fprintf(&b, "(deny file-write* (literal %q))\n", path)
	}
	for _, service := range []string{"com.apple.trustd.agent", "com.apple.trustd", "com.apple.system.notification_center", "com.apple.SystemConfiguration.DNSConfiguration", "com.apple.bsd.dirhelper", "com.apple.system.opendirectoryd.membership", "com.apple.ocspd", "com.apple.system.opendirectoryd.libinfo", "com.apple.SystemConfiguration.configd", "com.apple.networkd", "com.apple.SecurityServer"} {
		fmt.Fprintf(&b, "(allow mach-lookup (global-name %q))\n", service)
	}
	fmt.Fprintf(&b, "(allow network-outbound (remote ip %q))\n", fmt.Sprintf("localhost:%d", x.ProxyPort))
	if x.BridgePort != 0 {
		bridge := fmt.Sprintf("localhost:%d", x.BridgePort)
		fmt.Fprintf(&b, "(allow network-bind network-inbound (local ip %q))\n", bridge)
		fmt.Fprintf(&b, "(allow network-outbound (remote ip %q))\n", bridge)
	}
	if p.Auth == "keychain" {
		fmt.Fprintf(&b, "(allow network-outbound (remote unix-socket (path-literal %q)))\n", x.Broker)
	}
	return b.String(), nil
}
