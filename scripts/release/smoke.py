#!/usr/bin/env python3
"""Prove an unpacked app starts without Go and does not modify its bundle."""
import argparse
import hashlib
import json
import os
import select
import shutil
import signal
import socket
import subprocess
import tempfile
import time
from pathlib import Path


def hashes(app):
    return {str(p.relative_to(app)): hashlib.sha256(p.read_bytes()).hexdigest() for p in app.rglob('*') if p.is_file()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('app', type=Path)
    app = parser.parse_args().app.resolve(strict=True)
    source_app = app
    original = hashes(source_app)
    with tempfile.TemporaryDirectory(prefix='skyra-release-', dir='/tmp') as tmp:
        home = Path(tmp)
        app = home/'Skyra.app'
        shutil.copytree(source_app,app)
        before = hashes(app)
        runtime = app/'Contents/Resources/skyra'
        workspace = home/'work'; workspace.mkdir()
        (workspace/'a.txt').write_text('release content\n')
        state = home/'state'; state.mkdir()
        guards = home/'bin'; guards.mkdir()
        for name in ('go', 'swift', 'clang', 'npm'):
            guard = guards/name
            guard.write_text('#!/bin/sh\necho "Release attempted to run '+name+'" >&2\nexit 99\n')
            guard.chmod(0o755)
        with socket.socket() as s:
            s.bind(('127.0.0.1', 0)); port = s.getsockname()[1]
        env = dict(os.environ, PATH=str(guards)+':'+os.environ['PATH'], SKYRA_WORKSPACE=str(workspace), SKYRA_STATE=str(state), SKYRA_WS_PORT=str(port))
        process = subprocess.Popen([str(runtime/'skyra'), 'serve', '--fake'], cwd=workspace, env=env, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, start_new_session=True)
        log = b''
        try:
            deadline = time.monotonic()+45
            while time.monotonic() < deadline:
                ready, _, _ = select.select([process.stdout], [], [], .2)
                if ready:
                    data = os.read(process.stdout.fileno(), 65536)
                    if not data: break
                    log += data
                    if b'ready ws://' in log:
                        break
                if process.poll() is not None: break
            if b'ready ws://' not in log:
                raise RuntimeError(log.decode(errors='replace'))
            if not list((state/'runtime').glob('registry-*.json')):
                raise RuntimeError('registry was not stored outside the bundle')
            with socket.create_connection(('127.0.0.1', port), timeout=2):
                pass
            registry = next(p for p in (state/'runtime').glob('registry-*.json') if 'codex.start' in json.loads(p.read_text()))
            scratch = home/'notes'; scratch.mkdir()
            for number, request in enumerate(('read:a.txt', 'write:note.txt=release note')):
                result = subprocess.run([str(runtime/'build/controller/skyra'), 'boot', '-root', str(home/f'run-{number}'/'.skyra'), '-registry', str(registry), '-scratch', str(scratch), '-to', 'codex', '-text', request], env=env, cwd=workspace, capture_output=True, text=True, timeout=45)
                if result.returncode or 'byte-identical' not in result.stdout or 'refused' in result.stdout:
                    raise RuntimeError(result.stdout+result.stderr)
                if number == 0 and 'release content' not in result.stdout:
                    raise RuntimeError('provider did not receive plain read source: '+result.stdout)
            if (scratch/'note.txt').read_text() != 'release note':
                raise RuntimeError('scratch write did not land')
        finally:
            try: os.killpg(process.pid, signal.SIGTERM)
            except ProcessLookupError: pass
            try: process.wait(timeout=8)
            except subprocess.TimeoutExpired:
                os.killpg(process.pid, signal.SIGKILL); process.wait()
        if hashes(app) != before:
            raise RuntimeError('release modified its own signed bundle')
    if hashes(source_app) != original:
        raise RuntimeError('smoke test changed the input app')
    print('PASS: relocated runtime and provider read/write/replay work without build tools; bundle unchanged')


if __name__ == '__main__':
    main()
