# Publishing Skyra to Homebrew

Skyra keeps its FSL-1.1-ALv2 license. Distribution uses our own Homebrew tap,
not `homebrew/core`. The source repository stays private.

## Endpoints and status

- Tap: https://github.com/skyraOS/homebrew-tap — cask at `Casks/skyra.rb`.
- Downloads: https://github.com/skyraOS/skyra-releases — versioned GitHub Releases.
- Homepage: the downloads repository's README.
- Install after a release is published: `brew tap skyraOS/tap`, then
  `brew install --cask skyra`.

Releases can be published without a paid Apple Developer membership. The initial
release uses ad-hoc code signatures, without Apple Developer ID or notarization.
Homebrew installs it with quarantine enabled; macOS may block first launch.
Users can review it in System Settings > Privacy & Security and use Open Anyway
if they trust the download. Do not disable Gatekeeper or remove quarantine
automatically. Developer ID signing and notarization remain available for future
releases with a paid Apple Developer membership.

Initial support is Apple Silicon, macOS 14 or later. Intel downloads and a cask
architecture mapping can be added after an Intel build passes the same checks.
Homebrew installs Python 3.13 as a runtime dependency. Go and Xcode are build
requirements for maintainers; users do not need them to launch the packaged app.
Optional source-check tools may still require their toolchains.
Desktop observation additionally needs CuaDriver 0.34.0 installed at
`/Applications/CuaDriver.app` with its macOS Accessibility and Screen Recording
grants. The packaged app starts without desktop observation when the driver is
absent; it does not download or grant permissions automatically.
Codex 0.160.0 and the user's own login are required for that provider; we do not
redistribute provider credentials. Users can install the pinned version with
`npm install -g @openai/codex@0.160.0`, then `codex login`. Node is required for
that installation method. Keep this pin aligned with the provider launch policy.

## One-time maintainer setup

1. **Optional for notarized releases:** use an Apple Developer Program account with a **Developer ID Application**
   certificate and private key installed in the build machine's keychain.
   `security find-identity -v -p codesigning` must list that identity.
2. **Optional for notarized releases:** store notary credentials under a named keychain profile. For example, run
   `xcrun notarytool store-credentials skyra-notary` and follow its prompts.
   Use an app-specific password or an App Store Connect API key; never commit
   the certificate private key, password, API key or GitHub token.
3. Authenticate GitHub with `gh auth login`; the account needs write permission
   on both endpoints. Install Homebrew, full Xcode, Go 1.25+, and Python.
4. Register the tap locally with `brew tap skyraOS/tap`. Its Git checkout is
   reported by `brew --repository skyraOS/tap`; use that registered checkout for
   publishing so Homebrew can validate the cask.

The scripts do not change the source repository's visibility. They explicitly
copy runtime assets rather than uploading a private checkout or its Git history.

## Publish a version

Run from the source repository on an Apple Silicon Mac. Choose a new SemVer
version; the first intended stable version is `0.1.0`.

1. Commit the source changes, verify the relevant Go/Swift/Python suites, and
   record release notes in a file. The release metadata records the source HEAD.
2. Build the native app and all backend helpers:

   ```sh
   SKYRA_NO_OPEN=1 ./skyra app
   python3 scripts/release/package.py 0.1.0
   python3 scripts/release/smoke.py build/releases/0.1.0/arm64/Skyra.app
   ```

3. For the initial release, create ad-hoc signatures and archive the app without
   Apple credentials:

   ```sh
   python3 scripts/release/sign.py build/releases/0.1.0/arm64/Skyra.app --unsigned
   ```

   For a Developer ID signed and notarized release instead, use:

   ```sh
   python3 scripts/release/sign.py build/releases/0.1.0/arm64/Skyra.app \
     --identity 'Developer ID Application: YOUR NAME (TEAMID)' \
     --notary-profile skyra-notary
   ```

4. Unpack the final ZIP on a clean Mac, move the app away from the build folder,
   and test launch from Finder, workspace grants, a real provider turn, tools,
   restart and conversation replay. Test without Go or Xcode on PATH. Run the
   smoke script against the unpacked app as well. Verify the signature and
   stapled ticket after relocation for notarized releases; for unsigned releases,
   confirm the macOS first-launch approval flow. A passing fake smoke test alone does not
   validate provider login, TCC permissions or the signed runtime on another Mac.
5. Publish the verified ZIP and update the tap:

   ```sh
   python3 scripts/release/publish.py \
     build/releases/0.1.0/arm64/Skyra-0.1.0-macos-arm64.zip \
     --tap "$(brew --repository skyraOS/tap)" --notes /tmp/skyra-release-notes.md \
     --allow-unsigned
   ```

   Omit `--allow-unsigned` for notarized releases. All releases verify metadata
   and code signatures; notarized releases also require a stapled ticket and
   Gatekeeper acceptance. The command creates a release
   on the public downloads endpoint, generates the checksum-pinned cask from
   `packaging/homebrew/Casks/skyra.rb.in`, checks style, and commits/pushes the tap.
   It refuses to overwrite an existing release. If publishing succeeds but tap
   validation/push fails, keep the asset: fix the cask in the tap manually using
   the downloaded archive's SHA-256 and commit/push it. Do not rerun with a
   replacement asset under the same version.
6. Update the public downloads README to announce the released version and
   remove its initial pending notice (the tap README is updated by the script).
   Keep provider/setup instructions visible. Test the actual endpoint:

   ```sh
   brew tap skyraOS/tap
   brew install --cask skyra
   brew audit --new --cask skyra
   ```

   Verify the app under `/Applications`, the `skyra` CLI link, provider startup,
   uninstall and reinstall. Treat architecture/provider/TCC failures as release
   blockers. Homebrew may ask users to trust the third-party cask.

## Publish from GitHub Actions

The private source repository has a manual `Publish desktop release` workflow
(`.github/workflows/release.yml`). Configure a GitHub environment named
`distribution`; optionally require an environment reviewer before it can publish.
The workflow runs on Apple Silicon macOS, builds/tests the package, imports a
short-lived signing keychain when requested, and updates both public endpoints.
The `signing` input defaults to `unsigned`; select `notarized` to use Apple
credentials. Unsigned CI releases only require `SKYRA_DISTRIBUTION_TOKEN`.

Environment secrets:

| Name | Value |
| --- | --- |
| `SKYRA_SIGNING_P12` | Base64-encoded Developer ID Application certificate **and private key** export. |
| `SKYRA_SIGNING_PASSWORD` | Password protecting that P12 export. |
| `SKYRA_NOTARY_KEY` | Contents of an App Store Connect API private key (`.p8`). |
| `SKYRA_NOTARY_KEY_ID` | API key ID. |
| `SKYRA_NOTARY_ISSUER` | API issuer ID. |
| `SKYRA_DISTRIBUTION_TOKEN` | GitHub token with contents write on only the tap and downloads repositories. |

Environment variable `SKYRA_SIGNING_IDENTITY`: the complete
`Developer ID Application: NAME (TEAMID)` signing identity.
Never put these values in the public repositories or source control. The workflow
fails before building if credentials required for the selected mode are missing. Ephemeral signing credentials
are deleted even after a failure. It does not upload a private checkout as a
release asset. Confirm a clean-Mac real-provider acceptance run for the exact
source commit before triggering public publication; CI smoke uses the fake provider.

Run it from the source checkout with:

```sh
gh workflow run release.yml --ref main -f version=0.1.0 -f signing=unsigned
gh run list --workflow release.yml
```

Inspect the Actions run and the public ZIP/checksum/cask after completion.
A failed run can leave a successfully uploaded release without an updated cask;
follow the manual repair procedure above. Do not overwrite the asset.

## Continue publishing and roll back

For each update, repeat the versioned build, archive, clean-Mac test and publish
steps with a new version. Users update with `brew update` and
`brew upgrade --cask skyra`. Homebrew owns upgrades; no self-updater is bundled.
Never replace a ZIP at an existing versioned URL: existing casks pin its checksum.

To roll the endpoint back, change the tap's `version`, `sha256`, and URL to the
last verified release and commit/push that change. This changes future installs;
it does not automatically downgrade installed apps. Affected users can uninstall
and reinstall after backing up their data. Prefer a fixed release with a higher
version for ordinary regressions. Mark affected releases clearly in their notes.

## Pause or tear down distribution

These are explicit maintainer actions; the release scripts never perform them.

- **Pause new installs:** add `disable! date: "YYYY-MM-DD", because: "distribution is paused"`
  to `Casks/skyra.rb`, commit and push. Use the real date and explain the reason
  in the public README. Verify with `brew update` and `brew info --cask skyra`.
- **Withdraw one broken download:** first disable or repoint the cask, then run
  `gh release delete v0.1.0 --repo skyraOS/skyra-releases` after confirming the
  intended release. Existing installations are unaffected. Do not delete tags
  automatically or replace old assets; record the withdrawal.
- **Retire the service:** disable the cask first, publish a retirement notice,
  and leave the public endpoints archived as a record. Archiving is safer than
  deleting them. Revoke any dedicated publishing token and remove CI signing
  secrets after the last release. Keep backups of approved artifacts and hashes.
- **Delete endpoints permanently:** only after updating the tap and notices,
  export anything you need and delete `skyraOS/homebrew-tap` and/or
  `skyraOS/skyra-releases` in GitHub Settings. Deletion breaks old download URLs
  and Homebrew installs. This does not remotely uninstall users' apps.

Users remove the package with `brew uninstall --cask skyra` and remove the tap
with `brew untap skyraOS/tap`. We deliberately omit a destructive cask `zap`:
conversations, pinned replay binaries and logs under `~/.skyra`, workspaces,
provider logins and preferences remain. Users can remove that data themselves
only after backing up anything they need. Removing Homebrew must never delete
workspace source files or provider credentials.

## What the release contains

`Skyra.app` contains the Swift client and a runtime under
`Contents/Resources/skyra/`: launcher, compiled Go helpers, TUI, contracts,
workflow bundles, Python bridge code and synthetic Blender fixture, plus LICENSE.
There is no developer checkout path or developer PATH in its plist. A
`skyra-release.json` marker selects prebuilt mode. Generated registry/workflow
files and content-addressed pinned binaries go to `~/.skyra/runtime`; provider
state and credentials remain outside the signed application. Fake-mode smoke
verification compares the whole bundle before and after startup.

References: [Homebrew taps](https://docs.brew.sh/How-to-Create-and-Maintain-a-Tap),
[Cask Cookbook](https://docs.brew.sh/Cask-Cookbook),
[Developer ID signing](https://developer.apple.com/developer-id/).
