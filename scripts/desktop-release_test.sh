#!/usr/bin/env bash
# Dry-run test for scripts/desktop-release.sh on a fake kirokit.app folder.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
script="$root/scripts/desktop-release.sh"

work="$(mktemp -d)"
trap 'rm -rf "${work:?}"' EXIT

fail() { echo "FAIL: $*" >&2; exit 1; }

app="$work/kirokit.app"
mkdir -p "$app/Contents/MacOS"
echo "binary" > "$app/Contents/MacOS/kirokit"

out="$work/out"
log="$work/log"
bash "$script" --version 1.2.3 --app "$app" --out "$out" > "$log" 2>&1 || { cat "$log" >&2; fail "dry run exited non-zero"; }

zip="$out/kirokit-desktop_1.2.3_darwin_universal.zip"
[ -f "$zip" ] || fail "zip missing: $zip"

# The app folder is the zip's top level.
first="$(unzip -Z1 "$zip" | head -n 1)"
case "$first" in kirokit.app/*) ;; *) fail "first zip entry is '$first', want kirokit.app/..." ;; esac
unzip -Z1 "$zip" | grep -qx 'kirokit.app/Contents/MacOS/kirokit' || fail "binary missing from zip"

sha="$(shasum -a 256 "$zip" | awk '{print $1}')"
cask="$out/Casks/kirokit-desktop.rb"
[ -f "$cask" ] || fail "cask missing: $cask"

check() { grep -qF -- "$1" "$cask" || { cat "$cask" >&2; fail "cask lacks: $1"; }; }
check 'cask "kirokit-desktop" do'
check 'version "1.2.3"'
check "sha256 \"$sha\""
check 'url "https://github.com/mavogel/kiro-cli-kit/releases/download/v1.2.3/kirokit-desktop_1.2.3_darwin_universal.zip"'
check 'app "kirokit.app"'
check 'postflight do'
check '"/usr/bin/xattr"'
check 'args: ["-dr", "com.apple.quarantine"'

grep -qF "$sha" "$log" || fail "output lacks the sha256"
grep -qF 'kirokit-desktop_1.2.3_darwin_universal.zip' "$log" || fail "output lacks the zip name"
grep -qF 'Casks/kirokit-desktop.rb' "$log" || fail "output lacks the cask path"
grep -qi 'nothing was uploaded or pushed' "$log" || fail "output must say nothing was uploaded or pushed"

# A leading v on the version is accepted and dropped.
bash "$script" --version v1.2.3 --app "$app" --out "$work/out-v" > /dev/null 2>&1 || fail "leading v rejected"
[ -f "$work/out-v/kirokit-desktop_1.2.3_darwin_universal.zip" ] || fail "leading v not dropped"

# Missing app folder and missing version fail with a message.
if bash "$script" --version 1.2.3 --app "$work/nope.app" --out "$work/o2" > "$work/e1" 2>&1; then fail "missing app accepted"; fi
[ -s "$work/e1" ] || fail "missing app gave no message"
if bash "$script" --app "$app" --out "$work/o3" > "$work/e2" 2>&1; then fail "missing version accepted"; fi
[ -s "$work/e2" ] || fail "missing version gave no message"

echo "desktop-release_test: ok"
