#!/bin/sh
set -eu
cd "$(dirname "$0")"
mkdir -p artifacts
go build -o artifacts/root ./root
go build -o artifacts/codex ./codex
go build -o artifacts/echo ./root/echo
go build -o artifacts/fake-provider ./codex/testdata/fake-provider
./artifacts/root '{"echo":"./artifacts/echo","observations":["hello","again"]}'
exec ./artifacts/root codex-demo ./artifacts/codex ./codex/profiles.json ./artifacts/fake-provider ./artifacts/echo
