#!/usr/bin/env python3
"""Operator-owned, session-scoped Cua bootstrap. Never runs inside a model tool."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import socket
import subprocess
import sys
import tempfile
import time

VERSION = "0.34.0"
APP = Path("/Applications/CuaDriver.app")


def exchange(path, request):
    with socket.socket(socket.AF_UNIX) as conn:
        conn.settimeout(3)
        conn.connect(str(path))
        conn.sendall(json.dumps(request).encode() + b"\n")
        with conn.makefile("rb") as stream:
            data = stream.readline((4 << 20) + 1)
        if len(data) > 4 << 20 or not data.endswith(b"\n"):
            raise RuntimeError("invalid driver frame")
        return json.loads(data)


def main():
    p = argparse.ArgumentParser(description="Run Skyra with observation access to one selected window")
    p.add_argument("--list", action="store_true", help="list windows for operator selection")
    p.add_argument("--app", help="selected app bundle ID")
    p.add_argument("--pid", type=int)
    p.add_argument("--window-id", type=int)
    p.add_argument("command", nargs=argparse.REMAINDER, help="chat, tui, serve or ask TEXT (default chat)")
    a = p.parse_args()
    root = Path(__file__).resolve().parents[2]
    build = root / "build/controller"
    windows = json.loads(subprocess.check_output([str(build / "cua-windows")]))
    if a.list:
        print(json.dumps(windows, indent=2))
        return 0
    if not a.app or not a.pid or not a.window_id:
        p.error("select --app, --pid and --window-id from 'skyra desktop --list'")
    if not any(w["bundle_id"] == a.app and w["pid"] == a.pid and w["window_id"] == a.window_id for w in windows):
        p.error("the selected window is not live under that app and process")
    command = a.command or ["chat"]
    if command[0] == "--":
        command = command[1:]
    if not command or command[0] not in ("chat", "tui", "serve", "ask"):
        p.error("command must be chat, tui, serve or ask")
    binary = APP / "Contents/MacOS/cua-driver"
    if not binary.is_file():
        raise RuntimeError("install Cua Driver 0.34.0 first; see controller/cua/README.md")
    subprocess.run(["codesign", "--verify", "--deep", "--strict", str(APP)], check=True)
    if subprocess.check_output([str(binary), "--version"], text=True).strip() != "cua-driver " + VERSION:
        raise RuntimeError("Cua Driver version differs from the pinned release " + VERSION)
    children = []
    driver_pid = None
    def interrupted(signum, frame):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM, interrupted)
    signal.signal(signal.SIGINT, interrupted)
    # Short socket paths also work on Darwin. This directory is outside the
    # workspace, so read/write grants cannot reach upstream service state.
    with tempfile.TemporaryDirectory(prefix="skyra-cua-", dir="/private/tmp") as folder:
        folder = Path(folder)
        manifest, upstream, bridge = folder / "manifest.json", folder / "driver.sock", folder / "bridge.sock"
        pidfile = folder / "driver.pid"
        manifest.write_bytes(subprocess.check_output([str(build / "cua-bridge"), "manifest", "-app", a.app]))
        env = dict(os.environ, CUA_TELEMETRY_ENABLED="false")
        try:
            launcher = subprocess.Popen(["open", "-W", "-n", "-g", "-a", str(APP), "--args", "serve",
                "--socket", str(upstream), "--pid-file", str(pidfile), "--permission-mode", "bounded",
                "--capability-manifest", str(manifest), "--approve-capability-manifest", "--no-overlay"], env=env)
            children.append(launcher)
            deadline = time.monotonic() + 30
            while not upstream.exists():
                if launcher.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("Cua startup failed; check its macOS permission dialog")
                time.sleep(.1)
            metadata = exchange(upstream, {"method": "metadata"})
            driver_pid = metadata.get("result", {}).get("pid")
            if not metadata.get("ok") or not isinstance(driver_pid, int) or driver_pid <= 0 or not pidfile.exists() or int(pidfile.read_text()) != driver_pid:
                raise RuntimeError("driver process identity could not be verified")
            executable = subprocess.check_output(["ps", "-p", str(driver_pid), "-o", "comm="], text=True).strip()
            if Path(executable).resolve() != binary.resolve():
                driver_pid = None
                raise RuntimeError("socket peer does not identify the installed Cua app")
            service = subprocess.Popen([str(build / "cua-bridge"), "-app", a.app, "-pid", str(a.pid),
                "-window-id", str(a.window_id), "-driver-version", VERSION, "-manifest", str(manifest),
                "-driver-socket", str(upstream), "-socket", str(bridge)])
            children.append(service)
            deadline = time.monotonic() + 20
            while not bridge.exists():
                if service.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("observation bridge failed to start")
                time.sleep(.1)
            # Stored alongside the generated procedure/contract and pinned at
            # genesis by the ordinary registry. Only this bridge is granted.
            config = folder / "desktop.json"
            config.write_text(json.dumps({"socket": str(bridge), "bundle_id": a.app, "pid": a.pid,
                "window_id": a.window_id, "driver_version": VERSION,
                "driver_binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                "manifest_sha256": hashlib.sha256(b"cua-driver-capability-manifest-v3\0" + manifest.read_bytes()).hexdigest()}))
            env["SKYRA_DESKTOP_CONFIG"] = str(config)
            client = subprocess.Popen([str(root / "skyra"), *command], env=env)
            children.append(client)
            while client.poll() is None:
                if service.poll() is not None or launcher.poll() is not None:
                    raise RuntimeError("desktop service stopped; ending the scoped session")
                time.sleep(.2)
            return client.returncode
        finally:
            for child in reversed(children):
                if child.poll() is None:
                    child.terminate()
                    try:
                        child.wait(timeout=5)
                    except subprocess.TimeoutExpired:
                        child.kill()
                        child.wait()
            if driver_pid is not None:
                # Only the PID attested by this private service, while still
                # executing the installed app. Never kill a default daemon.
                command = subprocess.run(["ps", "-p", str(driver_pid), "-o", "comm="], capture_output=True, text=True)
                if command.stdout.strip() == str(binary):
                    try:
                        os.kill(driver_pid, signal.SIGTERM)
                    except ProcessLookupError:
                        pass


if __name__ == "__main__":
    try:
        sys.exit(main())
    except KeyboardInterrupt:
        sys.exit(130)
    except (RuntimeError, OSError, subprocess.CalledProcessError, ValueError) as error:
        print("skyra desktop:", error, file=sys.stderr)
        sys.exit(1)
