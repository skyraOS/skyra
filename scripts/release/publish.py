#!/usr/bin/env python3
"""Publish an ARM64 release and update the public Homebrew tap."""
import argparse
import hashlib
import json
import re
import subprocess
import tempfile
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
ENDPOINT = 'skyraOS/skyra-releases'


def run(*args):
    subprocess.run(args, check=True)


def main():
    p = argparse.ArgumentParser(description=__doc__)
    p.add_argument('archive', type=Path)
    p.add_argument('--tap', type=Path, required=True, help='checkout of skyraOS/homebrew-tap')
    p.add_argument('--notes', type=Path, required=True)
    p.add_argument('--allow-unsigned', action='store_true', help='publish without Apple notarization; disclose this in the cask')
    args = p.parse_args()
    archive = args.archive.resolve(strict=True)
    match = re.fullmatch(r'Skyra-(\d+\.\d+\.\d+)-macos-arm64\.zip', archive.name)
    if not match:
        p.error('publishing currently supports stable SemVer ARM64 releases only')
    version = match[1]
    tap = args.tap.resolve(strict=True)
    remote = subprocess.check_output(['git', '-C', str(tap), 'remote', 'get-url', 'origin'], text=True).strip()
    if remote.removesuffix('.git') not in ('https://github.com/skyraOS/homebrew-tap', 'git@github.com:skyraOS/homebrew-tap'):
        p.error('tap origin must be skyraOS/homebrew-tap')
    if subprocess.check_output(['git', '-C', str(tap), 'status', '--porcelain'], text=True).strip():
        p.error('tap checkout must be clean')
    checksum = hashlib.sha256(archive.read_bytes()).hexdigest()
    # Verify the actual downloaded ZIP contents, not a separate development app.
    with tempfile.TemporaryDirectory(prefix='skyra-assess-') as tmp:
        run('ditto', '-x', '-k', str(archive), tmp)
        app = Path(tmp)/'Skyra.app'
        metadata = json.loads((app/'Contents/Resources/skyra/skyra-release.json').read_text())
        if metadata['version'] != version or metadata['arch'] != 'arm64':
            p.error('archive metadata does not match filename')
        run('codesign', '--verify', '--deep', '--strict', str(app))
        if not args.allow_unsigned:
            run('xcrun', 'stapler', 'validate', str(app))
            run('spctl', '--assess', '--type', 'execute', str(app))
    # Never replace a published version's assets. Corrections get a new version.
    existing = subprocess.run(['gh', 'release', 'view', 'v'+version, '--repo', ENDPOINT], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
    if existing.returncode == 0:
        p.error('release already exists; use a new version, or update the tap manually after checking the existing asset')
    checksum_file = archive.with_suffix('.zip.sha256')
    checksum_file.write_text(f'{checksum}  {archive.name}\n')

    cask = (ROOT/'packaging/homebrew/Casks/skyra.rb.in').read_text().replace('@VERSION@', version).replace('@SHA256@', checksum).replace('@SIGNING_CAVEAT@', 'This release is not signed by an Apple Developer ID or notarized.\n    macOS may block first launch; review it in System Settings > Privacy & Security.' if args.allow_unsigned else 'This release is signed with Apple Developer ID and notarized.')
    dest = tap/'Casks/skyra.rb'; dest.parent.mkdir(exist_ok=True); dest.write_text(cask)
    run('brew', 'style', '--cask', str(dest))

    run('gh', 'release', 'create', 'v'+version, str(archive), str(checksum_file), '--repo', ENDPOINT, '--title', 'Skyra '+version, '--notes-file', str(args.notes))
    readme = tap/'README.md'
    readme.write_text('# Skyra Homebrew tap\n\nOfficial distribution of Skyra '+version+'. Apple Silicon; macOS 14+.\n\n```sh\nbrew tap skyraOS/tap\nbrew install --cask skyra\n```\n\nDownloads and setup: https://github.com/skyraOS/skyra-releases\n\nUpdate with `brew update` and `brew upgrade --cask skyra`.\nUninstall with `brew uninstall --cask skyra`; conversations, workspaces and provider logins remain.\nSkyra is licensed under FSL-1.1-ALv2.\n')
    if args.allow_unsigned:
        with readme.open('a') as f:
            f.write('\nThis release is ad-hoc signed, without Apple Developer ID or notarization. macOS may block first launch; review it in System Settings > Privacy & Security.\n')
    run('git', '-C', str(tap), 'add', '--', 'Casks/skyra.rb', 'README.md')
    run('git', '-C', str(tap), 'commit', '-m', 'Release Skyra '+version)
    run('git', '-C', str(tap), 'push', 'origin', 'HEAD:main')
    print('Published: brew tap skyraOS/tap && brew install --cask skyra')


if __name__ == '__main__':
    main()
