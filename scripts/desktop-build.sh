#!/usr/bin/env bash
# Build desktop/build/bin/kirokit.app: a universal Wails app with a universal
# kirokit CLI at Contents/Resources/bin/kirokit.
#
# VERSION comes from the environment (the release job passes the new tag) and
# defaults to `git describe`. A leading "v" is dropped, as GoReleaser does.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

version="${VERSION:-$(git describe --tags --always --dirty)}"
version="${version#v}"
commit="$(git rev-parse HEAD)"
date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

pkg="github.com/mavogel/kirokit/cmd/kirokit/commands"
ldflags="-s -w -X ${pkg}.version=${version} -X ${pkg}.commit=${commit} -X ${pkg}.date=${date}"

work="$(mktemp -d)"
trap 'rm -rf "${work:?}"' EXIT

for arch in amd64 arm64; do
  CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" \
    go build -trimpath -ldflags "$ldflags" -o "$work/kirokit-$arch" ./cmd/kirokit
done
lipo -create -output "$work/kirokit" "$work/kirokit-amd64" "$work/kirokit-arm64"

# macOS 13 is the app's minimum; without these the linker warns about mixed targets.
export CGO_CFLAGS="-mmacosx-version-min=13.0"
export CGO_LDFLAGS="-mmacosx-version-min=13.0"

# -clean empties build/bin, so the CLI is copied in afterwards.
(cd desktop && go run github.com/wailsapp/wails/v2/cmd/wails@v2.14.0 build -platform darwin/universal -clean)

app="desktop/build/bin/kirokit.app"
mkdir -p "$app/Contents/Resources/bin"
install -m 0755 "$work/kirokit" "$app/Contents/Resources/bin/kirokit"
# Wails sealed the bundle before the CLI joined it; seal it again (ad-hoc) so the signature stays valid.
codesign --force --deep --sign - "$app"
echo "built $app (version $version)"
