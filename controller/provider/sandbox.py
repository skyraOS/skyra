"""Codex provider launch grants, verified on macOS 15.6 / Codex 0.157.1.

Outbound IP access remains broad; this is not an egress boundary.
"""
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


def render(binary, home, scratch, work, auth):
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
    rules.extend(['(allow network-outbound (remote ip "*:*"))',
                  '(allow network-outbound (control-name "com.apple.netsrc"))',
                  '(allow network-outbound (remote unix-socket (path-literal "/private/var/run/mDNSResponder")))'])
    return "\n".join(rules) + "\n"
