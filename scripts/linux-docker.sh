#!/usr/bin/env bash
# Run a command in a Linux container that mirrors CI's Ubuntu runner, with this
# checkout mounted at /src, for Linux checks on a Mac (Docker or OrbStack):
#
#   scripts/linux-docker.sh go test ./...
#   scripts/linux-docker.sh task desktop:test
#   scripts/linux-docker.sh env GOOS=windows task desktop:build
#
# The image (scripts/linux-docker.Dockerfile) is built on first use. Go's
# caches and the frontend's node_modules live in Docker volumes, so the Mac's
# own node_modules are never replaced by Linux builds. Exits with the
# command's exit code.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
image="kirokit-linux-check"

[ "$#" -gt 0 ] || { echo "usage: scripts/linux-docker.sh <command> [args...]" >&2; exit 2; }

docker build -q -t "$image" -f "$root/scripts/linux-docker.Dockerfile" "$root/scripts" >/dev/null

tty=()
[ -t 0 ] && [ -t 1 ] && tty=(-t)

exec docker run --rm -i ${tty[@]+"${tty[@]}"} \
  -v "$root":/src -w /src \
  -v kirokit-linux-gomod:/go/pkg/mod \
  -v kirokit-linux-gocache:/root/.cache/go-build \
  -v kirokit-linux-node-modules:/src/desktop/frontend/node_modules \
  "$image" "$@"
