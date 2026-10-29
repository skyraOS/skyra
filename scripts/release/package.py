#!/usr/bin/env python3
"""Build a relocatable Skyra.app; prepare the distribution ZIP with sign.py afterwards."""
import argparse
import json
import platform
import plistlib
import re
import shutil
import subprocess
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]


def package(version, app):
    if not re.fullmatch(r'\d+\.\d+\.\d+(?:-[a-zA-Z0-9.]+)?', version):
        raise ValueError('version must be SemVer, e.g. 0.1.0')
    if not (app/'Contents/MacOS/Skyra').is_file():
        raise ValueError('build the desktop app with SKYRA_NO_OPEN=1 ./skyra app first')
    arch = {'arm64': 'arm64', 'x86_64': 'x86_64'}[platform.machine()]
    output = ROOT/'build/releases'/version/arch
    output.mkdir(parents=True, exist_ok=True)
    target = output/'Skyra.app'
    if app.resolve() == target.resolve():
        raise ValueError('input app must differ from the package output')
    if target.exists():
        shutil.rmtree(target)
    shutil.copytree(app, target)
    runtime = target/'Contents/Resources/skyra'
    runtime.mkdir(parents=True)
    for name in ('skyra', 'LICENSE'):
        shutil.copy2(ROOT/name, runtime/name)
    for name in ('docs/PUBLISHING.md',):
        if (ROOT/name).exists():
            shutil.copy2(ROOT/name, runtime/Path(name).name)
    # Explicit runtime inputs: do not archive the private checkout, history,
    # credentials, conversations, attachments, or development notes.
    for path in sorted((ROOT/'controller').rglob('*')):
        rel = path.relative_to(ROOT)
        if any(p in {'.git', '.skyra', 'node_modules', '__pycache__', 'testdata', 'bin'} for p in rel.parts):
            continue
        if path.is_symlink():
            raise ValueError(f'symlink runtime input refused: {rel}')
        runtime_json = path.suffix == '.json' and (path.name.endswith('contract.json') or path.name == 'bundle.json')
        runtime_python = path.suffix == '.py' and any(str(rel).startswith(prefix) for prefix in ('controller/codex/provider/', 'controller/blender/', 'controller/design/', 'controller/cua/')) and not path.name.startswith('test_') and not path.name.endswith('_test.py') and path.name != 'build_pilot.py'
        fixture = rel == Path('controller/blender/fixtures/core.blend')
        if path.is_file() and (runtime_json or runtime_python or fixture):
            dest = runtime/rel
            dest.parent.mkdir(parents=True, exist_ok=True)
            shutil.copy2(path, dest)
    binaries = runtime/'build/controller'
    binaries.mkdir(parents=True)
    names = 'skyra echo read find write roots agent user codexresident service-io service-host webhost fake-codex clauderesident claude-start claude-keychain desktop-observe cua-bridge cua-windows ideas-host design-tools repo-host blender-tools apprentice'.split()
    for name in names:
        source = ROOT/'build/controller'/name
        if not source.is_file():
            raise ValueError(f'missing binary: {source}; run ./skyra build')
        shutil.copy2(source, binaries/name)
        header = subprocess.check_output(['file', str(source)], text=True)
        if ('arm64' if arch == 'arm64' else 'x86_64') not in header:
            raise ValueError(f'wrong architecture: {name}')
    shutil.copy2(ROOT/'build/tui', runtime/'build/tui')
    metadata = {'version': version, 'arch': arch, 'commit': subprocess.check_output(['git', '-C', str(ROOT), 'rev-parse', 'HEAD'], text=True).strip(), 'codex_version': '0.160.0'}
    (runtime/'skyra-release.json').write_text(json.dumps(metadata, indent=2)+'\n')
    info = target/'Contents/Info.plist'
    with info.open('rb') as f:
        plist = plistlib.load(f)
    for key in ('SkyraRoot', 'SkyraPATH'):
        plist.pop(key, None)
    plist.update(CFBundleIdentifier='com.skyraos.skyra', CFBundleShortVersionString=version.split('-')[0], CFBundleVersion=version.split('-')[0])
    with info.open('wb') as f:
        plistlib.dump(plist, f)
    # The app and every helper must be signed after the bundle is complete.
    print(target)
    return target


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('version')
    parser.add_argument('--app', type=Path, default=ROOT/'build/Skyra.app')
    args = parser.parse_args()
    package(args.version, args.app)
