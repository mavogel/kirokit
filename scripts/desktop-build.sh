#!/usr/bin/env bash
# Build the desktop app for the OS `go env GOOS` names, with its own copy of
# the kirokit CLI:
#
#   darwin   desktop/build/bin/kirokit.app: a universal Wails v3 app with a
#            universal CLI at Contents/Resources/bin/kirokit
#   linux    desktop/build/bin/kirokit-desktop (WebKitGTK 4.1) and the CLI at
#            desktop/build/bin/kirokit, for GOARCH (default: this machine's)
#   windows  desktop/build/bin/installers/kirokit-desktop_<v>_windows_<arch>-setup.exe
#            (NSIS, with bin\kirokit.exe beside the app) for each of ARCHES
#            (default "amd64 arm64"); needs makensis on PATH
#
# Wails does not cross-compile cgo, so Linux builds on Linux; Windows needs no
# cgo, so GOOS=windows also works from Linux (scripts/linux-docker.sh).
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

# Wails v3's task files (desktop/Taskfile.yml, desktop/build/*/Taskfile.yml)
# call a bare `wails3`, so the version desktop/go.mod pins goes first on PATH.
# It runs on this machine, so a GOOS=windows cross-build must not reach it.
exe=""
[ "$(env -u GOOS go env GOOS)" = windows ] && exe=".exe"
(cd desktop && env -u GOOS -u GOARCH go build -o "$work/tools/wails3$exe" github.com/wailsapp/wails/v3/cmd/wails3)
export PATH="$work/tools:$PATH"

# Wails v3 writes to desktop/bin/; every consumer (CI's smoke start,
# scripts/desktop-prebuilt.sh, .goreleaser.yaml, the release workflow) reads
# desktop/build/bin/, so the output is staged there under its v2 names.
need() {
  for f in "$@"; do
    [ -e "$f" ] || { echo "desktop-build: missing $f" >&2; exit 1; }
  done
}
rm -rf desktop/bin desktop/build/bin
mkdir -p desktop/build/bin

case "$(go env GOOS)" in
darwin)
  for arch in amd64 arm64; do
    CGO_ENABLED=0 GOOS=darwin GOARCH="$arch" \
      go build -trimpath -ldflags "$ldflags" -o "$work/kirokit-$arch" ./cmd/kirokit
  done
  lipo -create -output "$work/kirokit" "$work/kirokit-amd64" "$work/kirokit-arm64"

  # macOS 13 is the app's minimum; the process environment wins over the task file's.
  export CGO_CFLAGS="-mmacosx-version-min=13.0"
  export CGO_LDFLAGS="-mmacosx-version-min=13.0"
  export MACOSX_DEPLOYMENT_TARGET="13.0"

  (cd desktop && wails3 task darwin:package:universal)
  need desktop/bin/kirokit.app/Contents/MacOS/kirokit desktop/bin/kirokit.app/Contents/Resources/iconfile.icns
  cp -R desktop/bin/kirokit.app desktop/build/bin/

  app="desktop/build/bin/kirokit.app"
  mkdir -p "$app/Contents/Resources/bin"
  install -m 0755 "$work/kirokit" "$app/Contents/Resources/bin/kirokit"
  # Wails sealed the bundle before the CLI joined it; seal it again (ad-hoc) so the signature stays valid.
  xattr -cr "$app"
  codesign --force --deep --sign - "$app"
  echo "built $app (version $version)"
  ;;

linux)
  arch="${GOARCH:-$(go env GOARCH)}"
  CGO_ENABLED=0 GOOS=linux GOARCH="$arch" \
    go build -trimpath -ldflags "$ldflags" -o "$work/kirokit" ./cmd/kirokit
  # webkit2_41: the WebKitGTK ABI of Ubuntu 22.04+, Debian 12+ and Fedora 40+.
  (cd desktop && go tool wails build -platform "linux/$arch" -tags webkit2_41 -clean -o kirokit-desktop)
  install -m 0755 "$work/kirokit" desktop/build/bin/kirokit
  echo "built desktop/build/bin/kirokit-desktop for linux/$arch (version $version)"
  ;;

windows)
  bin="desktop/build/bin"
  rm -rf "$bin"
  mkdir -p "$bin/installers"
  for arch in ${ARCHES:-amd64 arm64}; do
    # project.nsi takes the CLI from bin/cli/windows_<arch>/kirokit.exe.
    CGO_ENABLED=0 GOOS=windows GOARCH="$arch" \
      go build -trimpath -ldflags "$ldflags" -o "$bin/cli/windows_$arch/kirokit.exe" ./cmd/kirokit
    # No -clean: it would delete the CLI staged above. The wails tool itself
    # runs on this machine, so a GOOS=windows cross-build must not reach it.
    (cd desktop && env -u GOOS -u GOARCH go tool wails build -platform "windows/$arch" -nsis -o "kirokit-desktop-$arch.exe")
    mv "$bin/kirokit-$arch-installer.exe" "$bin/installers/kirokit-desktop_${version}_windows_$arch-setup.exe"
  done
  echo "built $bin/installers (version $version):"
  ls "$bin/installers"
  ;;

*)
  echo "desktop-build: no desktop build for $(go env GOOS)" >&2
  exit 1
  ;;
esac
