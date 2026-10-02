"""Codex provider launch grants, verified on macOS 15.6 / Codex 0.160.0 (launch.CODEX_VERSION)."""
import json
from pathlib import Path
import shutil
import subprocess

def native_binary():
    command = Path(shutil.which("codex") or "").resolve()
    if command.suffix != ".js":
        return command
    matches = list(command.parent.parent.glob("node_modules/@openai/codex-darwin-*/vendor/*/bin/codex"))
    if len(matches) != 1:
        raise RuntimeError("pass --codex with the native Codex executable")
    return matches[0].resolve()


def render(binary, home, scratch, work, auth, proxy_port=None, capture_port=None):
    # Operator bootstrap grants, not yet derived from Horizon fields. Network
    # remains broadly available for inference; egress restriction is separate.
    rules = ["(version 1)", "(deny default)", "(allow process-fork)",
             f"(allow process-exec (literal {json.dumps(str(binary))}))",
             '(allow process-exec (literal "/usr/lib/dyld"))',
             '(allow file-read-metadata)',
             '(allow file-map-executable)',
             '(allow system-socket (require-all (socket-domain AF_SYSTEM) (socket-protocol 2)))',
             '(allow ipc-posix-shm* (ipc-posix-name "com.apple.AppleDatabaseChanged"))',
             "(allow sysctl-read)"]
    cache = Path(subprocess.check_output(["/usr/bin/getconf", "DARWIN_USER_CACHE_DIR"], text=True).strip()).resolve()
    for path in ["/System", "/usr/lib", "/usr/share/zoneinfo", "/usr/share/icu", "/private/etc",
                 str(cache / "mds"),
                 "/private/var/db/mds",
                 "/private/var/run/com.apple.security.cryptexd",
                 "/private/var/db/dyld", "/private/var/db/timezone", str(home), str(scratch), str(work)]:
        rules.append(f"(allow file-read* (subpath {json.dumps(path)}))")
    for path in ["/", str(binary.parent), str(binary), str(binary.parent.parent / "codex-package.json"), str(auth), str(home.parent / "models.json"), "/dev/null", "/dev/random", "/dev/urandom"]:
        rules.append(f"(allow file-read* (literal {json.dumps(path)}))")
    for path in [home, scratch]:
        rules.append(f"(allow file-write* (subpath {json.dumps(str(path))}))")
    rules.append('(allow file-write* (literal "/dev/null"))')
    rules.append(f'(allow file-write-data (literal {json.dumps(str(cache / "mds/mds.lock"))}))')
    rules.append(f'(deny file-write* (literal {json.dumps(str(home / "config.toml"))}))')
    rules.append(f'(deny file-write* (literal {json.dumps(str(auth))}))')
    # Start with explicit services, never unrestricted mach-lookup.
    for service in ["com.apple.trustd.agent", "com.apple.trustd",
                    "com.apple.system.notification_center",
                    "com.apple.SystemConfiguration.DNSConfiguration", "com.apple.bsd.dirhelper",
                    "com.apple.system.opendirectoryd.membership", "com.apple.ocspd",
                    "com.apple.system.opendirectoryd.libinfo", "com.apple.SystemConfiguration.configd",
                    "com.apple.networkd", "com.apple.SecurityServer"]:
        rules.append(f"(allow mach-lookup (global-name {json.dumps(service)}))")
    if proxy_port is None:
        # Existing probe callers exercise the historical profile. Production
        # launches always supply a proxy port and cannot fall back to this.
        rules.extend(['(allow network-outbound (remote ip "*:*"))',
                      '(allow network-outbound (control-name "com.apple.netsrc"))',
                      '(allow network-outbound (remote unix-socket (path-literal "/private/var/run/mDNSResponder")))'])
    else:
        if not isinstance(proxy_port, int) or not 1 <= proxy_port <= 65535:
            raise ValueError('invalid egress proxy port')
        rules.append(f'(allow network-outbound (remote ip "localhost:{proxy_port}"))')
    if capture_port is not None:
        # Test only (tool_surface.py): a local server that captures the
        # inference request in place of the provider.
        if not isinstance(capture_port, int) or not 1 <= capture_port <= 65535:
            raise ValueError('invalid capture port')
        rules.append(f'(allow network-outbound (remote ip "localhost:{capture_port}"))')
    return "\n".join(rules) + "\n"
