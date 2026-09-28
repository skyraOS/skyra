#!/bin/bash
# Records assets/desktop.gif: Skyra.app on real Codex, reading this repo.
# From the repo root: assets/desktop-demo.sh [seconds]   (default 90)
# Needs Screen Recording and Accessibility for the terminal that runs it, a
# built build/Skyra.app (skyra app), and a logged-in Codex.
#
# The grant is a git-archive snapshot of the tracked files (no .env, no build/),
# so nothing untracked can end up in the GIF. The app's own workspace and mode
# are put back when the script exits. The raw recording is kept at
# build/desktop-demo.mov. The published GIF was then cut by hand to drop
# the wait between the last read and the answer.
set -euo pipefail

REPO="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP="$REPO/build/Skyra.app"
ID=local.skyra.desktop
SECS=${1:-90}
PROMPT=${SKYRA_DEMO_PROMPT:-"go take a look at controller/core and tell me what you think of the design."}
X=120 Y=80 W=1100 H=720      # window frame, in points
OUT="$REPO/assets/desktop.gif"
RAW="$REPO/build/desktop-demo.mov"

work=$(mktemp -d)
snap="$work/skyra"
mkdir "$snap"
git -C "$REPO" archive HEAD | tar -x -C "$snap"

old_ws=$(defaults read $ID workspace 2>/dev/null || true)
old_fake=$(defaults read $ID fake 2>/dev/null || echo 0)
restore() {
  osascript -e 'tell application "Skyra" to quit' 2>/dev/null || true
  [ -n "$old_ws" ] && defaults write $ID workspace "$old_ws"
  defaults write $ID fake -bool "$([ "$old_fake" = 1 ] && echo true || echo false)"
  rm -rf "$work"
}
trap restore EXIT

osascript -e 'tell application "Skyra" to quit' 2>/dev/null || true
sleep 2
defaults write $ID workspace "$snap"
defaults write $ID fake -bool false
open "$APP"
sleep 6   # the backend comes up

osascript <<OSA
tell application "Skyra" to activate
tell application "System Events" to tell process "Skyra"
  set position of window 1 to {$X, $Y}
  set size of window 1 to {$W, $H}
  keystroke "n" using command down
end tell
OSA
sleep 1.5

screencapture -v -x -V "$SECS" -R "$X,$Y,$W,$H" "$RAW" &
rec=$!
sleep 1.5

osascript <<OSA
tell application "System Events" to tell process "Skyra"
  repeat with c in characters of "$PROMPT"
    keystroke c
    delay 0.045
  end repeat
  delay 0.6
  key code 36
end tell
OSA

wait $rec

# The typing plays at real speed; Codex's waiting and reading run at 6x. The
# sky's twinkle defeats frame-diffing, so denoise and skip dithering to keep
# the file small.
ffmpeg -y -loglevel error -i "$RAW" -filter_complex "
  [0:v]fps=12,scale=960:-1:flags=lanczos,hqdn3d=3:3:6:6,split[a][b];
  [a]trim=0:6,setpts=PTS-STARTPTS[t];
  [b]trim=6,setpts=(PTS-STARTPTS)/6[r];
  [t][r]concat=n=2:v=1[v];
  [v]split[p][q];[p]palettegen=max_colors=128:stats_mode=diff[pal];
  [q][pal]paletteuse=dither=none:diff_mode=rectangle" "$OUT"
echo "wrote $OUT ($(du -h "$OUT" | cut -f1))"
