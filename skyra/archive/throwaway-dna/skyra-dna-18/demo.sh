#!/bin/sh
set -eu
cd "$(dirname "$0")"
mkdir -p artifacts
go build -o artifacts/echo ./echo
go build -o artifacts/root ./root
exec ./artifacts/root '{"echo":"./artifacts/echo","observations":["hello","again"]}'
