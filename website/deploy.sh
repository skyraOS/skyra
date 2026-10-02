#!/bin/sh
# Builds the site here and publishes it to the Mac mini: only the published
# files go over (pages, assets, runs), never the source. On the mini, Caddy
# serves them on 127.0.0.1:8080, kept running by launchd (deploy/install.sh).
#
#   website/deploy.sh            # to the host `skyra` (an ssh alias)
#   DEPLOY_HOST=other website/deploy.sh
set -eu
here=$(cd "$(dirname "$0")" && pwd)
host=${DEPLOY_HOST:-skyra}

"$here/build.sh"

ssh "$host" 'mkdir -p ~/site/www'
rsync -a --delete \
  --include='*.html' --include='assets/***' --include='runs/***' --exclude='*' \
  "$here/" "$host:site/www/"
scp -q "$here/deploy/Caddyfile" "$here/deploy/com.skyra.site.plist" "$host:site/"
ssh "$host" 'sh -s' < "$here/deploy/install.sh"
