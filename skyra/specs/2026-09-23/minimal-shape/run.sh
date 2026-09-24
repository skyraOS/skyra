#!/bin/sh
# Boot to echo, three ways. Each scenario gets a fresh .skyra/ state directory.
set -eu
cd "$(dirname "$0")"
mkdir -p bin
go build -o bin/skyra ./skyra
go build -o bin/echo ./echo

scenario() {
  name=$1; shift
  root=$(mktemp -d)/.skyra
  echo
  echo "══ $name ══"
  ./bin/skyra boot -root "$root" -echo ./bin/echo "$@"
  echo "   state: $root"
}

scenario "echo hello"                     -text hello
scenario "echo probes .skyra/ from inside" -text hello -probe
scenario "echo (a leaf) tries to invoke"   -misbehave
