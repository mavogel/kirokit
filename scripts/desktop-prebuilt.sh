#!/usr/bin/env bash
# Stage the built desktop app for GoReleaser's prebuilt builders, for the OS
# `go env GOOS` names.
#
#   desktop-prebuilt.sh [<kirokit.app>]
#
# darwin: the Wails build is universal, and a cask only takes DMGs labelled
# amd64 or arm64, so the universal app binary is staged once per architecture at
#   desktop/build/bin/prebuilt/kirokit_darwin_{amd64,arm64}/kirokit
# and the files GoReleaser's generated bundle lacks are copied beside them
# (Info.plist, icon, the bundled CLI) for `app_bundles.extra_files`.
#
# linux: the app and its CLI from `task desktop:build` go to
#   desktop/build/bin/prebuilt/kirokit-desktop_linux_<GOARCH>/{kirokit-desktop,kirokit}
# for the kirokit-desktop nfpm. Only that architecture's folder is replaced, so
# an arm64 build copied in from another runner stays.
#
# windows: nothing; the setup programs go to the release as extra files.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
out="$root/desktop/build/bin/prebuilt"

case "$(go env GOOS)" in
darwin)
  app="${1:-$root/desktop/build/bin/kirokit.app}"
  [ -d "$app" ] || { echo "desktop-prebuilt: app folder not found: $app" >&2; exit 1; }

  rm -rf "$out"
  mkdir -p "$out/bundle/Contents/Resources/bin"

  for arch in amd64 arm64; do
    mkdir -p "$out/kirokit_darwin_$arch"
    cp "$app/Contents/MacOS/kirokit" "$out/kirokit_darwin_$arch/kirokit"
    # The build sealed the whole app, binding this binary's signature to the
    # original bundle's Info.plist and resources. Signed on its own it is valid
    # in the bundle GoReleaser assembles.
    codesign --remove-signature "$out/kirokit_darwin_$arch/kirokit"
    codesign --force --sign - "$out/kirokit_darwin_$arch/kirokit"
  done

  cp "$app/Contents/Info.plist" "$out/bundle/Contents/Info.plist"
  cp "$app/Contents/Resources/iconfile.icns" "$out/bundle/Contents/Resources/iconfile.icns"
  cp "$app/Contents/Resources/bin/kirokit" "$out/bundle/Contents/Resources/bin/kirokit"
  ;;

linux)
  bin="$root/desktop/build/bin"
  for f in kirokit-desktop kirokit; do
    [ -f "$bin/$f" ] || { echo "desktop-prebuilt: $bin/$f not found; run task desktop:build first" >&2; exit 1; }
  done
  dir="$out/kirokit-desktop_linux_${GOARCH:-$(go env GOARCH)}"
  rm -rf "$dir"
  mkdir -p "$dir"
  install -m 0755 "$bin/kirokit-desktop" "$dir/kirokit-desktop"
  install -m 0755 "$bin/kirokit" "$dir/kirokit"
  ;;

windows)
  echo "desktop-prebuilt: nothing to stage on Windows (the setup programs are release extra files)"
  exit 0
  ;;

*)
  echo "desktop-prebuilt: no desktop app for $(go env GOOS)" >&2
  exit 1
  ;;
esac
echo "staged $out"
