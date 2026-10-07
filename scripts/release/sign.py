#!/usr/bin/env python3
"""Sign every bundled Mach-O, notarize, staple, assess, then create the public ZIP."""
import argparse
import hashlib
import subprocess
from pathlib import Path


def run(*args):
    subprocess.run(args, check=True)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('app', type=Path)
    p.add_argument('--identity', required=True, help='Developer ID Application identity')
    p.add_argument('--notary-profile', required=True, help='notarytool stored-credentials profile')
    p.add_argument('--notary-keychain', help='optional keychain containing the notary profile')
    args = p.parse_args()
    app = args.app.resolve(strict=True)
    if app.name != 'Skyra.app' or not (app/'Contents/Resources/skyra/skyra-release.json').is_file():
        p.error('use an app built by package.py')
    if not args.identity.startswith('Developer ID Application:'):
        p.error('public distribution requires a Developer ID Application identity')
    for path in sorted(app.rglob('*')):
        if not path.is_file() or path.is_symlink():
            continue
        with path.open('rb') as f:
            magic = f.read(4)
        if magic in (b'\xcf\xfa\xed\xfe', b'\xfe\xed\xfa\xcf', b'\xca\xfe\xba\xbe', b'\xbe\xba\xfe\xca'):
            run('codesign', '--force', '--options', 'runtime', '--timestamp', '--sign', args.identity, str(path))
    for bundle in sorted(app.rglob('*.bundle'), key=lambda p: len(p.parts), reverse=True):
        run('codesign', '--force', '--timestamp', '--sign', args.identity, str(bundle))
    run('codesign', '--force', '--options', 'runtime', '--timestamp', '--sign', args.identity, str(app))
    run('codesign', '--verify', '--deep', '--strict', '--verbose=2', str(app))
    upload = app.parent/'notarization.zip'
    run('ditto', '-c', '-k', '--keepParent', str(app), str(upload))
    keychain = ['--keychain', args.notary_keychain] if args.notary_keychain else []
    run('xcrun', 'notarytool', 'submit', str(upload), '--keychain-profile', args.notary_profile, '--wait', *keychain)
    run('xcrun', 'stapler', 'staple', str(app))
    run('spctl', '--assess', '--type', 'execute', '--verbose=2', str(app))
    version, arch = app.parent.parent.name, app.parent.name
    archive = app.parent/f'Skyra-{version}-macos-{arch}.zip'
    run('ditto', '-c', '-k', '--keepParent', str(app), str(archive))
    checksum = hashlib.sha256(archive.read_bytes()).hexdigest()
    archive.with_suffix('.zip.sha256').write_text(f'{checksum}  {archive.name}\n')
    upload.unlink()
    print(archive)


if __name__ == '__main__':
    main()
