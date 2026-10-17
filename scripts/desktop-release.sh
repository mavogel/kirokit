#!/usr/bin/env bash
# Zip the desktop app and render its Homebrew cask.
#
#   desktop-release.sh --version <v> --app <kirokit.app> --out <dir> [--publish]
#
# Without --publish nothing leaves the machine: the script writes
#   <out>/kirokit-desktop_<version>_darwin_universal.zip
#   <out>/Casks/kirokit-desktop.rb
# and prints the zip's sha256. With --publish it also runs
# `gh release upload v<version>` and pushes the cask to the default branch of
# the tap repository (TAP_REPO, default mavogel/homebrew-tap). The push uses
# HOMEBREW_TAP_GITHUB_TOKEN when set.
set -euo pipefail

repo="mavogel/kiro-cli-kit"
tap="${TAP_REPO:-mavogel/homebrew-tap}"

usage() { echo "usage: desktop-release.sh --version <version> --app <kirokit.app> --out <dir> [--publish]" >&2; }
die() { echo "desktop-release: $*" >&2; usage; exit 1; }

version="" app="" out="" publish=0
while [ $# -gt 0 ]; do
  case "$1" in
    --version) [ $# -ge 2 ] || die "--version needs a value"; version="$2"; shift 2 ;;
    --app) [ $# -ge 2 ] || die "--app needs a value"; app="$2"; shift 2 ;;
    --out) [ $# -ge 2 ] || die "--out needs a value"; out="$2"; shift 2 ;;
    --publish) publish=1; shift ;;
    *) die "unknown argument: $1" ;;
  esac
done

version="${version#v}"
[ -n "$version" ] || die "--version is required"
[ -n "$app" ] || die "--app is required"
[ -n "$out" ] || die "--out is required"
[ -d "$app" ] || die "app folder not found: $app"
[ "$(basename "$app")" = "kirokit.app" ] || die "the app folder must be named kirokit.app: $app"

zip_name="kirokit-desktop_${version}_darwin_universal.zip"
zip="$out/$zip_name"
cask="$out/Casks/kirokit-desktop.rb"
url="https://github.com/${repo}/releases/download/v${version}/${zip_name}"

mkdir -p "$out/Casks"
rm -f "$zip"
# --norsrc/--noextattr/--noacl keep AppleDouble (._*) entries out of the zip:
# `unzip` extracts them as files, which breaks the app's code-signature seal.
ditto -c -k --keepParent --norsrc --noextattr --noacl "$app" "$zip"
sha="$(shasum -a 256 "$zip" | awk '{print $1}')"

cat > "$cask" <<EOF
cask "kirokit-desktop" do
  version "${version}"
  sha256 "${sha}"

  url "${url}"
  name "kirokit"
  desc "Desktop app to install and sync Kiro CLI team configs"
  homepage "https://github.com/${repo}"

  depends_on macos: ">= :ventura"

  app "kirokit.app"

  # The app is not notarized; without this macOS reports "kirokit is damaged".
  postflight do
    system_command "/usr/bin/xattr",
                   args: ["-dr", "com.apple.quarantine", "#{appdir}/kirokit.app"]
  end
end
EOF

echo "zip:    $zip"
echo "sha256: $sha"
echo "cask:   $cask (Casks/kirokit-desktop.rb)"

if [ "$publish" -eq 0 ]; then
  echo "Dry run: nothing was uploaded or pushed (pass --publish to upload and push)."
  exit 0
fi

gh release upload "v${version}" "$zip" --repo "$repo" --clobber

work="$(mktemp -d)"
trap 'rm -rf "${work:?}"' EXIT
GH_TOKEN="${HOMEBREW_TAP_GITHUB_TOKEN:-${GH_TOKEN:-}}" gh repo clone "$tap" "$work/tap" -- --depth 1
mkdir -p "$work/tap/Casks"
cp "$cask" "$work/tap/Casks/kirokit-desktop.rb"
git -C "$work/tap" add Casks/kirokit-desktop.rb
if git -C "$work/tap" diff --cached --quiet; then
  echo "cask already up to date in $tap"
else
  git -C "$work/tap" -c user.name="github-actions[bot]" -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
    commit -m "kirokit-desktop ${version}"
  GH_TOKEN="${HOMEBREW_TAP_GITHUB_TOKEN:-${GH_TOKEN:-}}" \
    git -C "$work/tap" -c credential.helper='!gh auth git-credential' push
fi
echo "Published ${zip_name} to v${version} and the cask to ${tap}."
