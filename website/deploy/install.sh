#!/bin/sh
# Runs on the mini (deploy.sh sends it over ssh): installs Caddy if it is
# missing, installs the launchd agent, and (re)starts the site.
set -eu
export PATH=/opt/homebrew/bin:$PATH
command -v caddy >/dev/null || brew install caddy

export SITE_ROOT=$HOME/site/www SITE_LOG=$HOME/site/access.log
caddy validate --config "$HOME/site/Caddyfile" --adapter caddyfile >/dev/null 2>&1 ||
  caddy validate --config "$HOME/site/Caddyfile" --adapter caddyfile

agents=$HOME/Library/LaunchAgents
mkdir -p "$agents"
sed "s#HOME/#$HOME/#g" "$HOME/site/com.skyra.site.plist" > "$agents/com.skyra.site.plist"
uid=$(id -u)
launchctl bootout "gui/$uid/com.skyra.site" 2>/dev/null || true
launchctl bootstrap "gui/$uid" "$agents/com.skyra.site.plist"

for _ in 1 2 3 4 5 6 7 8 9 10; do
  if code=$(curl -fsS -o /dev/null -w '%{http_code}' http://127.0.0.1:8080/ 2>/dev/null); then
    echo "serving: $code on 127.0.0.1:8080 ($(ls "$SITE_ROOT"/runs/*.json | wc -l | tr -d ' ') bundles)"
    exit 0
  fi
  sleep 0.5
done
echo "caddy did not come up; see ~/site/caddy.log" >&2
tail -5 "$HOME/site/caddy.log" >&2 || true
exit 1
