#!/usr/bin/env sh
set -eu

export GOCACHE="${GOCACHE:-/tmp/go-build}"
exec go run ./cmd/api
