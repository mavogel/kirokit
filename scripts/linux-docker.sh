#!/usr/bin/env bash
# Run a command in a Linux container that mirrors CI's Ubuntu runner, with this
# checkout mounted at /src, for Linux checks on a Mac (Docker or OrbStack):
#
#   scripts/linux-docker.sh go test ./...
#   scripts/linux-docker.sh task desktop:test
#   scripts/linux-docker.sh env GOOS=windows task desktop:build
#   LINUX_DOCKER_ARCH=amd64 scripts/linux-docker.sh task desktop:build
#
# The image (scripts/linux-docker.Dockerfile) is built on first use, for this
# machine's architecture or LINUX_DOCKER_ARCH (amd64 or arm64; amd64 is what
# ubuntu-latest runs). Go's caches and the frontend's node_modules live in
# Docker volumes per architecture, so the Mac's own node_modules are never
# replaced by Linux builds. Exits with the command's exit code.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
arch="${LINUX_DOCKER_ARCH:-}"
image="kirokit-linux-check${arch:+-$arch}"
platform=()
[ -n "$arch" ] && platform=(--platform "linux/$arch")

[ "$#" -gt 0 ] || { echo "usage: scripts/linux-docker.sh <command> [args...]" >&2; exit 2; }

docker build -q ${platform[@]+"${platform[@]}"} -t "$image" -f "$root/scripts/linux-docker.Dockerfile" "$root/scripts" >/dev/null

tty=()
[ -t 0 ] && [ -t 1 ] && tty=(-t)

exec docker run --rm -i ${tty[@]+"${tty[@]}"} ${platform[@]+"${platform[@]}"} \
  -v "$root":/src -w /src \
  -v kirokit-linux-gomod:/go/pkg/mod \
  -v "$image-gocache":/root/.cache/go-build \
  -v "$image-node-modules":/src/desktop/frontend/node_modules \
  "$image" "$@"
