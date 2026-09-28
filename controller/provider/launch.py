#!/usr/bin/env python3
"""Prepare and launch a Codex process under mandatory external Seatbelt."""
import argparse
import json
import os
from pathlib import Path
import socket
import subprocess
import sys
import tempfile
import time
from sandbox import native_binary, render

FEATURES = ['apps', 'plugins', 'hooks', 'daemon_auto_start', 'browser_use',
            'browser_use_external', 'computer_use', 'image_generation',
            'multi_agent', 'multi_agent_v2', 'shell_snapshot', 'code_mode', 'code_mode_only',
            'code_mode_host', 'goals', 'sleep_tool', 'current_time_reminder',
            'shell_tool', 'skill_search', 'view_image', 'request_permissions_tool',
            'send_message_to_user_async', 'token_budget']


def direct_catalog(binary):
    # Use the installed version's own discovered model metadata; never fall
    # back to remote defaults that could restore provider-owned tools.
    cache = json.loads((Path.home() / '.codex/models_cache.json').read_text())
    version = subprocess.check_output([str(binary), '--version'], text=True).strip().split()[-1]
    if version != '0.157.1':
        raise RuntimeError('direct tool isolation is verified for Codex 0.157.1; validate the tool-surface test before upgrading')
    if cache.get('client_version') != version or not cache.get('models'):
        raise RuntimeError('refresh Codex model metadata with this installed version before launching Skyra')
    for model in cache['models']:
        model['tool_mode'] = 'standard'
        model['multi_agent_version'] = 'disabled'
        model['experimental_supported_tools'] = []
        model['supports_search_tool'] = False
        model['include_skills_usage_instructions'] = False
        model['include_plugin_usage_instructions'] = False
        model['include_apps_usage_instructions'] = False
    return {'models': cache['models']}


def configuration(run):
    return 'model_catalog_json = ' + json.dumps(str(run / 'models.json')) + '\n' + 'cli_auth_credentials_store = "file"\nweb_search = "disabled"\nsandbox_mode = "read-only"\napproval_policy = "never"\n[agents]\nenabled = false\n[tools.experimental_request_user_input]\nenabled = false\n[tools.update_plan]\nenabled = false\n[skills.bundled]\nenabled = false\n'


def prepare(state, binary=None):
    if sys.platform != 'darwin':
        raise RuntimeError('the provider requires macOS Seatbelt')
    binary = (binary or native_binary()).resolve(strict=True)
    for executable in [binary, Path('/usr/bin/sandbox-exec')]:
        if not executable.is_file() or not os.access(executable, os.X_OK):
            raise RuntimeError(f'missing executable: {executable}')
    auth = (Path.home() / '.codex/auth.json').resolve(strict=True)
    state.mkdir(parents=True, exist_ok=True, mode=0o700)
    run = Path(tempfile.mkdtemp(prefix='provider-', dir=state)).resolve()
    home, scratch, work = [run / name for name in ['home', 'tmp', 'work']]
    for directory in [home, scratch, work]:
        directory.mkdir(mode=0o700)
    (home / 'auth.json').symlink_to(auth)
    catalog = direct_catalog(binary)
    (run / 'models.json').write_text(json.dumps(catalog))
    (home / 'config.toml').write_text(configuration(run))
    (run / 'actor.sb').write_text(render(binary, home, scratch, work, auth))
    (run / 'launch.json').write_text(json.dumps({'binary': str(binary), 'user_home': str(Path.home())}))
    # Verify the profile and native loader before starting a persistent host.
    subprocess.run(['/usr/bin/sandbox-exec', '-f', str(run / 'actor.sb'), str(binary), '--version'],
                   env=environment(run), cwd=work, stdout=subprocess.DEVNULL, check=True)
    return run


def environment(run):
    manifest = json.loads((run / 'launch.json').read_text())
    return {'PATH': '/usr/bin:/bin', 'HOME': manifest['user_home'],
            'CODEX_HOME': str(run / 'home'), 'TMPDIR': str(run / 'tmp')}


def launch(run):
    manifest = json.loads((run / 'launch.json').read_text())
    argv = ['/usr/bin/sandbox-exec', '-f', str(run / 'actor.sb'), manifest['binary']]
    for feature in FEATURES:
        argv.extend(['--disable', feature])
    argv.extend(['--enable', 'skip_host_skill_discovery', 'app-server', '--listen', 'stdio://'])
    os.chdir(run / 'work')
    # Never retry without the profile, and do not inherit personal integrations.
    os.execve(argv[0], argv, environment(run))


def ready(path):
    deadline = time.monotonic() + 20
    while not path.exists():
        if time.monotonic() >= deadline:
            raise RuntimeError('host socket did not appear')
        time.sleep(.05)
    with socket.socket(socket.AF_UNIX) as sock:
        sock.settimeout(20)
        sock.connect(str(path))
        stream = sock.makefile('rwb')
        def call(value):
            # Host deduplicates bytes; match Go encoding/json's sorted maps.
            stream.write(json.dumps(value, sort_keys=True, separators=(',', ':')).encode() + b'\n'); stream.flush()
            answer = json.loads(stream.readline())
            if 'error' in answer:
                raise RuntimeError(answer['error'])
            return answer
        # Same stable initialization ID and payload as the adapter: replayed
        # delivery is a no-op. Readiness means the provider actually answered.
        result = call({'op': 'deliver', 'id': 'initialize', 'line': {
            'id': 'initialize', 'method': 'initialize', 'params': {
                'clientInfo': {'name': 'skyra', 'version': '1'},
                'capabilities': {'experimentalApi': True}}}})
        position = result['seq']
        while time.monotonic() < deadline:
            batch = call({'op': 'read', 'from': position, 'wait_ms': 1000})
            position = batch['next']
            for event in batch.get('events') or []:
                if event.get('host'):
                    raise RuntimeError(event['host'])
                message = event.get('line', {})
                if message.get('id') == 'initialize' and 'result' in message:
                    if not call({'op': 'ping'}).get('alive'):
                        raise RuntimeError('provider exited after initialization')
                    return
                if message.get('id') == 'initialize' and 'error' in message:
                    raise RuntimeError('provider initialization failed')
        raise RuntimeError('provider initialization timed out')


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    action = parser.add_mutually_exclusive_group(required=True)
    action.add_argument('--prepare', type=Path)
    action.add_argument('--launch', type=Path)
    action.add_argument('--ready', type=Path)
    action.add_argument('--check', action='store_true')
    parser.add_argument('--codex', type=Path)
    args = parser.parse_args()
    if args.check:
        if sys.platform != 'darwin' or not Path('/usr/bin/sandbox-exec').is_file():
            raise RuntimeError('macOS Seatbelt is required')
        if not (Path.home() / '.codex/auth.json').is_file():
            raise RuntimeError('run codex login with file-based credentials first')
        direct_catalog((args.codex or native_binary()).resolve(strict=True))
        print('Codex 0.157.1 login and direct-tool model metadata ready')
    elif args.prepare:
        print(prepare(args.prepare, args.codex))
    elif args.launch:
        launch(args.launch.resolve(strict=True))
    else:
        ready(args.ready)


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print(f'codex provider: {error}', file=sys.stderr)
        sys.exit(1)
