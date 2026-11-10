#!/bin/sh
# Makes the test directories and the six cases under ./testdir, fresh.
# Run from anywhere; then: LEVEL=8 python3 -I step.py testdir/case-cat-a.json
set -e
cd "$(dirname "$0")"
T="$(pwd)/testdir"
rm -rf "$T" && mkdir -p "$T/a" "$T/b" "$T/m/src" "$T/m/dst" "$T/r"
printf 'hello from notes\nsecond line\n' > "$T/a/notes.txt"
printf '{"name": "skyra"}\n' > "$T/b/config.json"
echo q3 > "$T/m/src/report.txt"
echo old > "$T/r/old.log"
T="$T" python3 -I - <<'PY'
import json, os
T = os.environ["T"]
cases = {
    "cat-a": [f"show the contents of the only file in this directory:\n{T}/a", ["ls", "cat"]],
    "cat-b": [f"show the contents of the only file in this directory:\n{T}/b", ["ls", "cat"]],
    "wc": [f"count the lines of this file:\n{T}/a/notes.txt", ["wc"]],
    "date": ["what time is it?", ["date"]],
    "mv": [f"move the only file in the first folder into the second folder:\n{T}/m/src\n{T}/m/dst", ["ls", "mv"]],
    "rm": [f"delete the only file in this directory:\n{T}/r", ["ls", "rm"]],
}
for k, v in cases.items():
    json.dump(v, open(f"{T}/case-{k}.json", "w"))
PY
