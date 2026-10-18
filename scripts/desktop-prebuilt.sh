#!/usr/bin/env bash
# Stage the built desktop app for GoReleaser's prebuilt builder.
#
#   desktop-prebuilt.sh [<kirokit.app>]
#
# The Wails build is universal, and a cask only takes DMGs labelled amd64 or
# arm64, so the universal app binary is staged once per architecture at
#   desktop/build/bin/prebuilt/kirokit_darwin_{amd64,arm64}/kirokit
# and the files GoReleaser's generated bundle lacks are copied beside them
# (Info.plist, icon, the bundled CLI) for `app_bundles.extra_files`.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
app="${1:-$root/desktop/build/bin/kirokit.app}"
[ -d "$app" ] || { echo "desktop-prebuilt: app folder not found: $app" >&2; exit 1; }

out="$root/desktop/build/bin/prebuilt"
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
echo "staged $out"
