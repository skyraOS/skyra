#!/bin/sh
set -eu
cd "$(dirname "$0")"
# The test skips unless explicit opt-in, exact version and launch policy are set.
exec go test -v -count=1 ./root -run '^TestLiveProvider$'
