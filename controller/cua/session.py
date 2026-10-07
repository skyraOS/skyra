#!/usr/bin/env python3
"""Operator-owned, session-scoped Cua bootstrap. Never runs inside a model tool."""
import argparse
import hashlib
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time

VERSION = "0.34.0"
APP = Path("/Applications/CuaDriver.app")


def choose_window(windows):
    if not windows:
        raise RuntimeError("no matching windows are open; open the app you want to observe and try again")
    if not sys.stdin.isatty():
        raise RuntimeError("window selection needs a terminal; use 'skyra desktop --list' and supply --app, --pid and --window-id for a scripted launch")
    windows = sorted(windows, key=lambda w: (w["app"].lower(), w["title"].lower(), w["window_id"]))
    def label(text):
        return " ".join("".join(c for c in text if c.isprintable()).split())[:120]
    print("Choose a window for Skyra to observe:", flush=True)
    for i, window in enumerate(windows, 1):
        title = label(window["title"]) or "Untitled window"
        print(f"  {i}. {label(window['app'])} — {title}", flush=True)
    while True:
        try:
            choice = input("Window number (q to cancel): ").strip()
        except EOFError:
            return None
        if choice.lower() in ("q", "quit", "cancel"):
            return None
        if choice.isdecimal() and 1 <= int(choice) <= len(windows):
            return windows[int(choice) - 1]
        print(f"Enter a number from 1 to {len(windows)}, or q.", flush=True)


def main():
    p = argparse.ArgumentParser(prog="skyra desktop", description="Choose a window, then run Skyra with read-only vision (default chat)")
    p.add_argument("--all-windows", action="store_true", help="normal app startup: provide observation of windows open at launch")
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
    if a.all_windows:
        windows = windows[:128]
        if not windows:
            raise RuntimeError("no windows are available for desktop observation")
        a.app = ",".join(sorted({w["bundle_id"] for w in windows}))
        a.pid, a.window_id = 0, 0
    elif not a.app or not a.pid or not a.window_id:
        candidates = [w for w in windows if
                      (a.app is None or w["bundle_id"] == a.app) and
                      (a.pid is None or w["pid"] == a.pid) and
                      (a.window_id is None or w["window_id"] == a.window_id)]
        selected = choose_window(candidates)
        if selected is None:
            print("Desktop session cancelled.")
            return 0
        a.app, a.pid, a.window_id = selected["bundle_id"], selected["pid"], selected["window_id"]
        # The operator may take time to choose. Recheck ownership before
        # starting either service rather than trusting the old list.
        windows = json.loads(subprocess.check_output([str(build / "cua-windows")]))
    if not a.all_windows and not any(w["bundle_id"] == a.app and w["pid"] == a.pid and w["window_id"] == a.window_id for w in windows):
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
    def interrupted(signum, frame):
        raise KeyboardInterrupt
    signal.signal(signal.SIGTERM, interrupted)
    signal.signal(signal.SIGINT, interrupted)
    # Short socket paths also work on Darwin. This directory is outside the
    # workspace, so read/write grants cannot reach upstream service state.
    with tempfile.TemporaryDirectory(prefix="skyra-cua-", dir="/private/tmp") as folder:
        folder = Path(folder)
        manifest, bridge = folder / "manifest.json", folder / "bridge.sock"
        manifest.write_bytes(subprocess.check_output([str(build / "cua-bridge"), "manifest", "-app", a.app]))
        try:
            bridge_args = []
            if a.all_windows:
                inventory = folder / "windows.json"
                inventory.write_text(json.dumps([{"bundle_id": w["bundle_id"], "pid": w["pid"],
                    "window_id": w["window_id"], "app_name": w["app"].encode()[:512].decode(errors="ignore"),
                    "window_title": w["title"].encode()[:1024].decode(errors="ignore")} for w in windows]))
                bridge_args = ["-windows", str(inventory)]
            # The bridge launches the bounded daemon from this approved
            # manifest, verifies it, and replaces it when Cua latches the
            # manifest idle or expired. It ends its daemon when it stops.
            service = subprocess.Popen([str(build / "cua-bridge"), "-app", a.app, "-pid", str(a.pid),
                "-window-id", str(a.window_id), "-driver-version", VERSION, "-manifest", str(manifest),
                "-driver-app", str(APP), "-socket", str(bridge), *bridge_args])
            children.append(service)
            deadline = time.monotonic() + 45
            while not bridge.exists():
                if service.poll() is not None or time.monotonic() > deadline:
                    raise RuntimeError("observation bridge failed to start; check Cua's macOS permission dialog")
                time.sleep(.1)
            # Stored alongside the generated procedure/contract and pinned at
            # genesis by the ordinary registry. Only this bridge is granted.
            config = folder / "desktop.json"
            config.write_text(json.dumps({"socket": str(bridge), "bundle_id": a.app, "pid": a.pid,
                "window_id": a.window_id, "driver_version": VERSION,
                "driver_binary_sha256": hashlib.sha256(binary.read_bytes()).hexdigest(),
                "manifest_sha256": hashlib.sha256(b"cua-driver-capability-manifest-v3\0" + manifest.read_bytes()).hexdigest()}))
            env = dict(os.environ, SKYRA_DESKTOP_CONFIG=str(config))
            client = subprocess.Popen([str(root / "skyra"), *command], env=env)
            children.append(client)
            while client.poll() is None:
                if service.poll() is not None:
                    raise RuntimeError("desktop service stopped; ending the scoped session")
                time.sleep(.2)
            return client.returncode
        finally:
            # A second signal must not cut the cleanup short: the wrappers
            # above forward theirs, and a terminal's reaches every layer.
            signal.signal(signal.SIGTERM, signal.SIG_IGN)
            signal.signal(signal.SIGINT, signal.SIG_IGN)
            for child in reversed(children):
                if child.poll() is None:
                    child.terminate()
                    try:
                        child.wait(timeout=10)
                    except subprocess.TimeoutExpired:
                        child.kill()
                        child.wait()
            stop_daemons(folder, binary)


def stop_daemons(folder, binary):
    """End any daemon still serving from this session's private directory.

    The bridge stops its daemon; this covers a bridge that was killed before
    it could. Only processes of the installed executable whose arguments name
    this session's directory are signalled."""
    found = subprocess.run(["pgrep", "-f", str(folder) + "/driver-"], capture_output=True, text=True).stdout.split()
    for pid in map(int, found):
        command = subprocess.run(["ps", "-p", str(pid), "-o", "comm="], capture_output=True, text=True)
        if command.stdout.strip() == str(binary):
            try:
                os.kill(pid, signal.SIGTERM)
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
