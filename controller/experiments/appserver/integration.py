#!/usr/bin/env python3
"""Actual Skyra adapter/Drive/replay checks; --live uses the production launcher."""
import argparse
import base64
import json
import os
from pathlib import Path
import signal
import subprocess
import sys
import tempfile
import time
from probe import Host, REPO


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--live', action='store_true')
    args = parser.parse_args()
    run = Path(tempfile.mkdtemp(prefix='skyra-e2e-', dir='/tmp')).resolve()
    os.chmod(run, 0o700)
    print(f'evidence: {run}', flush=True)
    for name in ['skyra', 'codex', 'echo', 'host']:
        subprocess.run(['go', 'build', '-o', str(run / name), './controller/' + name], cwd=REPO, check=True)
    if args.live:
        prepared = Path(subprocess.check_output([sys.executable, str(REPO / 'controller/provider/launch.py'), '--prepare', str(run / 'providers')], text=True).strip())
        actor = [sys.executable, str(REPO / 'controller/provider/launch.py'), '--launch', str(prepared)]
    else:
        subprocess.run(['go', 'build', '-o', str(run / 'fake'), './history/dna19/codex/testdata/fake-provider'], cwd=REPO, check=True)
        actor = [str(run / 'fake')]
    sock = run / 'host.sock'
    log = (run / 'host.log').open('w')
    proc = subprocess.Popen([str(run / 'host'), '-sock', str(sock), '-state', str(run / 'host-state'), '--', *actor], stdout=log, stderr=log, start_new_session=True)
    try:
        subprocess.run([sys.executable, str(REPO / 'controller/provider/launch.py'), '--ready', str(sock)], check=True, timeout=25)
        reg = {name: {'path': str(run / name), 'contract': str(REPO / f'controller/{name}/contract.json')} for name in ['codex', 'echo']}
        reg['codex']['connect'] = [str(sock)]
        registry = run / 'registry.json'; registry.write_text(json.dumps(reg))
        def turn(name, text, expected):
            root = run / name; root.mkdir()
            result = subprocess.run([str(run / 'skyra'), 'boot', '-root', str(root), '-registry', str(registry), '-to', 'codex', '-payload', json.dumps({'text': text, 'conversation': 'regression'})], capture_output=True, text=True, timeout=180)
            output = result.stdout + result.stderr
            (run / (name + '.log')).write_text(output)
            answers = [line for line in output.splitlines() if line.startswith('answer:')]
            if not answers or expected not in answers[-1] or (name not in ['native', 'unregistered'] and 'fault (' in answers[-1]):
                raise AssertionError(f'{name}: missing {expected!r}; inspect {name}.log')
            if name == 'tool' and 'invoke echo' not in output:
                raise AssertionError('tool response did not pass through a runtime echo invocation')
            replay = subprocess.run([str(run / 'skyra'), 'replay', str(root)], capture_output=True, text=True, check=True)
            print('PASS', name, replay.stdout.strip(), flush=True)
        if args.live:
            # Direct app-server endpoints test the OS independently of the
            # model's willingness to attempt a forbidden operation.
            h = Host(sock)
            canary = run / 'outside.txt'; canary.write_text('PRIVATE_CANARY')
            for method, params in [
                ('fs/readFile', {'path': str(canary)}),
                ('fs/writeFile', {'path': str(run / 'forbidden.txt'), 'dataBase64': base64.b64encode(b'NO').decode()}),
                ('process/spawn', {'command': ['/bin/sh', '-c', 'echo SHOULD_NOT_RUN'], 'cwd': str(prepared / 'work'), 'processHandle': 'os-denial'})]:
                response = h.rpc(method, params)
                assert 'error' in response and 'permi' in json.dumps(response).lower(), response
                print('PASS OS denial', method, flush=True)
            h.close()
            turn('normal', 'Reply exactly SKYRA_NORMAL_OK. Do not use tools.', 'SKYRA_NORMAL_OK')
            turn('tool', 'Call the registered echo tool with text SKYRA_TOOL_OK, then reply with its returned text.', 'SKYRA_TOOL_OK')
            # The model may refuse a native-tool prompt. Exercise the real
            # interruption protocol deterministically; the fake-provider run
            # separately verifies the adapter's native-event rejection path.
            events = [json.loads(line).get('line', {}) for line in (run / 'host-state/inbox.jsonl').read_text().splitlines()]
            thread = next(e['result']['thread']['id'] for e in events if 'thread' in e.get('result', {}))
            h = Host(sock)
            started = h.rpc('turn/start', {'threadId': thread, 'input': [{'type': 'text', 'text': 'Write a long explanation of sorting algorithms. Do not call tools.'}]})
            identifier = started['result']['turn']['id']
            answer = h.rpc('turn/interrupt', {'threadId': thread, 'turnId': identifier})
            assert 'result' in answer, answer
            deadline = time.monotonic() + 30
            while time.monotonic() < deadline:
                if any(m.get('method') == 'turn/completed' and m.get('params', {}).get('turn', {}).get('id') == identifier and m['params']['turn']['status'] == 'interrupted' for m in h.messages):
                    break
                h.pump()
            else:
                raise AssertionError('real turn interruption was not confirmed')
            h.close()
            print('PASS real running turn interruption confirmed', flush=True)
            turn('recovery', 'Reply exactly SKYRA_RECOVERY_OK. Do not use tools.', 'SKYRA_RECOVERY_OK')
        else:
            turn('normal', 'hello', 'hello')
            turn('tool', 'tool:SKYRA_TOOL_OK', 'SKYRA_TOOL_OK')
            turn('native', 'native-running', 'native.commandExecution')
            turn('unregistered', 'unregistered-tool', 'native.dynamicTool.reply')
            turn('recovery', 'after-fault', 'after-fault')
        events = [json.loads(line).get('line', {}) for line in (run / 'host-state/inbox.jsonl').read_text().splitlines()]
        interrupted = [e for e in events if e.get('method') == 'turn/completed' and e.get('params', {}).get('turn', {}).get('status') == 'interrupted']
        assert interrupted, 'no confirmed interrupted turn in provider journal'
        print('PASS interruption confirmed and next turn completed in the same conversation', flush=True)
    finally:
        if proc.poll() is None:
            os.killpg(proc.pid, signal.SIGTERM)
            try: proc.wait(timeout=5)
            except subprocess.TimeoutExpired:
                os.killpg(proc.pid, signal.SIGKILL); proc.wait()
        log.close()
    return 0


if __name__ == '__main__':
    raise SystemExit(main())
