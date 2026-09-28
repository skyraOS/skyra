#!/usr/bin/env python3
"""Opt-in real-Codex Seatbelt experiment through the existing Skyra host.

Uses the existing login read-only, isolated state, harmless local canaries, and
optionally real model turns. Never changes the production launcher or thread
sandbox mode. Raw evidence stays in a private temporary directory.
"""

import argparse
import base64
import hashlib
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import tempfile
import time
import uuid


REPO = Path(__file__).resolve().parents[3]


import sys
sys.path.insert(0, str(REPO / "controller/provider"))
from sandbox import native_binary, render
from launch import FEATURES, configuration, direct_catalog


class Host:
    def __init__(self, path):
        self.socket = socket.socket(socket.AF_UNIX)
        self.socket.settimeout(5)
        self.socket.connect(str(path))
        self.file = self.socket.makefile("rwb")
        self.seq = 0
        self.serial = 0
        self.prefix = uuid.uuid4().hex
        self.messages = []

    def call(self, request):
        self.file.write(json.dumps(request).encode() + b"\n")
        self.file.flush()
        line = self.file.readline()
        if not line:
            raise RuntimeError("host disconnected")
        answer = json.loads(line)
        if "error" in answer:
            raise RuntimeError(answer["error"])
        return answer

    def deliver(self, message):
        self.serial += 1
        return self.call({"op": "deliver", "id": f"probe-{self.prefix}-{self.serial}", "line": message})

    def pump(self):
        result = self.call({"op": "read", "from": self.seq, "wait_ms": 1000})
        self.seq = result["next"]
        for event in result.get("events") or []:
            if event.get("host"):
                raise RuntimeError(event["host"])
            if isinstance(event.get("line"), dict):
                self.messages.append(event["line"])

    def rpc(self, method, params, timeout=30):
        identifier = f"rpc-{self.prefix}-{self.serial + 1}"
        start = len(self.messages)
        self.deliver({"id": identifier, "method": method, "params": params})
        deadline = time.monotonic() + timeout
        while time.monotonic() < deadline:
            for message in self.messages[start:]:
                if message.get("id") == identifier and "method" not in message:
                    return message
            start = len(self.messages)
            self.pump()
        raise TimeoutError(method)

    def close(self):
        self.file.close()
        self.socket.close()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--codex", type=Path, help="native executable, not the npm shim")
    parser.add_argument("--live", action="store_true", help="also make real model requests using the existing login")
    args = parser.parse_args()
    binary = (args.codex or native_binary()).resolve()
    auth = Path.home() / ".codex" / "auth.json"
    if not auth.is_file():
        raise RuntimeError("this experiment needs an existing file-based Codex login")
    run = Path(tempfile.mkdtemp(prefix="skyra-os-", dir="/tmp")).resolve()
    os.chmod(run, 0o700)
    home, scratch, work = (run / name for name in ["codex-home", "tmp", "work"])
    for directory in [home, scratch, work]:
        directory.mkdir(mode=0o700)
    # No credential bytes are read, copied into evidence, or printed by this
    # harness. Seatbelt grants the actor read access to the existing login only.
    (home / "auth.json").symlink_to(auth)
    (run / "models.json").write_text(json.dumps(direct_catalog(binary)))
    (home / "config.toml").write_text(configuration(run))
    profile = render(binary, home, scratch, work, auth)
    profile_path = run / "actor.sb"
    profile_path.write_text(profile)
    canary = run / "outside-grant.txt"
    canary.write_text("SKYRA_OS_READ_CANARY")
    allowed = scratch / "inside-grant.txt"
    allowed.write_text("SKYRA_OS_ALLOWED_CANARY")
    host_bin = run / "host"
    subprocess.run(["go", "build", "-o", str(host_bin), "./controller/host"], cwd=REPO, check=True)
    log = (run / "host.log").open("w")
    sock = run / "host.sock"
    env = {"PATH": "/usr/bin:/bin", "HOME": str(Path.home()), "CODEX_HOME": str(home), "TMPDIR": str(scratch)}
    # Keep shell tools available for the denial probes. Disable unrelated
    # integrations so this isolated app-server cannot start personal plugins.
    flags = []
    for feature in FEATURES:
        flags.extend(["--disable", feature])
    argv = [str(host_bin), "-sock", str(sock), "-state", str(run / "host-state"), "--",
            "/usr/bin/sandbox-exec", "-f", str(profile_path), str(binary), *flags, "--enable", "skip_host_skill_discovery", "app-server", "--listen", "stdio://"]
    report = {"binary": str(binary), "version": subprocess.check_output([str(binary), "--version"], text=True).strip(),
              "profile_sha256": hashlib.sha256(profile.encode()).hexdigest(), "checks": [], "live": args.live,
              "limitations": ["experimental file/Mach grants", "no egress host allowlist", "existing auth is read-only",
                              "production launcher unchanged", "normal threads retain read-only policy",
                              "dynamic tool response supplied by probe, not Skyra Drive",
                              "system certificate cache lock and named shared memory are writable",
                              "grants are not yet minimized or tested across OS versions"]}
    proc = subprocess.Popen(argv, cwd=work, env=env, stdout=log, stderr=log, start_new_session=True)
    h = None
    print(f"evidence: {run}", flush=True)

    def check(name, passed, detail):
        report["checks"].append({"name": name, "passed": bool(passed), "detail": detail})
        print(f"{'PASS' if passed else 'FAIL'} {name}: {json.dumps(detail)}", flush=True)
        if not passed:
            raise RuntimeError(name)

    def result(method, params):
        answer = h.rpc(method, params)
        if "error" in answer:
            raise RuntimeError(f"{method}: {answer['error']}")
        return answer["result"]

    try:
        deadline = time.monotonic() + 10
        while not sock.exists():
            if proc.poll() is not None or time.monotonic() > deadline:
                raise RuntimeError("host failed to start; inspect host.log")
            time.sleep(0.05)
        h = Host(sock)
        initialized = result("initialize", {"clientInfo": {"name": "skyra-os-probe", "version": "1"},
                                             "capabilities": {"experimentalApi": True}})
        h.deliver({"method": "initialized", "params": {}})
        check("app-server initializes under Seatbelt", True, initialized)
        check("no inherited managed requirements", result("configRequirements/read", {}).get("requirements") is None, "requirements=null")
        read = result("fs/readFile", {"path": str(allowed)})
        check("in-process allowed read control", base64.b64decode(read["dataBase64"]) == b"SKYRA_OS_ALLOWED_CANARY", "allowed canary read")
        written = scratch / "write-control.txt"
        result("fs/writeFile", {"path": str(written), "dataBase64": base64.b64encode(b"WRITE_CONTROL").decode()})
        check("in-process allowed write control", written.read_bytes() == b"WRITE_CONTROL", "scratch write succeeded")
        denied = h.rpc("fs/writeFile", {"path": str(home / "config.toml"), "dataBase64": base64.b64encode((home / "config.toml").read_bytes()).decode()})
        check("config remains read-only", "error" in denied and "permi" in json.dumps(denied).lower(), denied)
        denied = h.rpc("fs/readFile", {"path": str(canary)})
        check("in-process outside read denied", "error" in denied and "permi" in json.dumps(denied).lower(), denied)
        target = work / "should-not-exist.txt"
        denied = h.rpc("fs/writeFile", {"path": str(target), "dataBase64": base64.b64encode(b"OS_PROBE").decode()})
        check("in-process outside write denied", "error" in denied and not target.exists() and "permi" in json.dumps(denied).lower(), denied)
        # process/spawn intentionally bypasses Codex's command sandbox, but it
        # still runs in the app-server's outer Seatbelt. This benign direct
        # request isolates OS enforcement without weakening any thread policy.
        denied = h.rpc("process/spawn", {"processHandle": "denied-shell", "command": ["/bin/sh", "-c", "printf SKYRA_SHELL_EXECUTED"], "cwd": str(work)})
        check("direct shell spawn denied by outer OS sandbox", "error" in denied and "permi" in json.dumps(denied).lower(), denied)
        denied = h.rpc("command/exec", {"command": ["/bin/sh", "-c", "printf SKYRA_SHELL_EXECUTED"], "cwd": str(work),
                                        "sandboxPolicy": {"type": "readOnly"}, "timeoutMs": 3000})
        check("read-only native command cannot execute", "permi" in json.dumps(denied).lower() and "SKYRA_SHELL_EXECUTED" not in denied.get("result", {}).get("stdout", ""), denied)

        if args.live:
            tools = [{"name": "echo", "description": "Return the supplied text through the external test client.",
                      "inputSchema": {"type": "object", "required": ["text"], "properties": {"text": {"type": "string"}}, "additionalProperties": False}}]
            started = result("thread/start", {"cwd": str(work), "approvalPolicy": "never", "sandbox": "read-only",
                                              "environments": [], "dynamicTools": tools, "ephemeral": True})
            check("thread retains read-only policy", started.get("sandbox", {}).get("type") == "readOnly" and started.get("approvalPolicy") == "never",
                  {key: started.get(key) for key in ["sandbox", "approvalPolicy", "model", "environments"]})
            thread = started["thread"]["id"]
            turn_start = len(h.messages)
            turn = result("turn/start", {"threadId": thread, "input": [{"type": "text", "text": "Call the echo tool exactly once with text SKYRA_DYNAMIC_OK. Then reply with its returned text. Do not use any other tools."}]})["turn"]["id"]
            deadline = time.monotonic() + 120
            count, final, complete = 0, "", None
            index = turn_start
            while time.monotonic() < deadline and complete is None:
                h.pump()
                while index < len(h.messages):
                    m = h.messages[index]
                    index += 1
                    p = m.get("params", {})
                    if p.get("threadId") != thread:
                        continue
                    if m.get("method") == "item/tool/call":
                        check("registered dynamic tool requested", p.get("tool") == "echo", {"tool": p.get("tool"), "arguments": p.get("arguments")})
                        count += 1
                        text = p["arguments"]["text"]
                        h.deliver({"id": m["id"], "result": {"success": True, "contentItems": [{"type": "inputText", "text": text}]}})
                    elif m.get("method") == "item/completed" and p.get("item", {}).get("type") == "agentMessage":
                        final = p["item"]["text"]
                    elif m.get("method") == "turn/completed" and p.get("turn", {}).get("id") == turn:
                        complete = p["turn"]["status"]
            check("real model turn with dynamic tool completes", complete == "completed" and count == 1 and "SKYRA_DYNAMIC_OK" in final,
                  {"status": complete, "tool_calls": count, "answer": final})
        report["passed"] = True
    except Exception as error:
        report["passed"] = False
        report["error"] = str(error)
        print(f"ERROR {error}; inspect {run / 'host.log'}", flush=True)
    finally:
        if h is not None:
            h.close()
        # Only this experiment's newly created process group is terminated.
        if proc.poll() is None:
            os.killpg(proc.pid, signal.SIGTERM)
            try:
                proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(proc.pid, signal.SIGKILL)
                proc.wait()
        log.close()
        (home / "auth.json").unlink()
        (run / "report.json").write_text(json.dumps(report, indent=2) + "\n")
    return 0 if report.get("passed") else 1


if __name__ == "__main__":
    raise SystemExit(main())
