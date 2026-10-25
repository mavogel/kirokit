#!/usr/bin/env bash
# Checks the showcase must pass before record.sh writes it to assets/:
#   checks.sh privacy <text-file>...  no home path, user name or git email
#   checks.sh size <gif> <mp4>        GIF at most 8 MB, MP4 at most 4 MB
#   checks.sh length <media>...       each 30 to 60 seconds long
# Each failure prints "showcase: <check>: <reason>" and exits 1.
set -euo pipefail

# The project's public GitHub slugs, shown on purpose (clone address, brew
# tap). They contain the owner's name, which can equal the local user name.
public_slugs=("mavogel/kirokit" "mavogel/tap/")

fail() { echo "showcase: $1: $2" >&2; exit 1; }

privacy() {
  local email file slug text
  email="$(git config user.email || true)"
  (($#)) || fail privacy "no files to check"
  for file; do
    [[ -f "$file" ]] || fail privacy "missing $file"
    grep -qiF -- "$HOME" "$file" && fail privacy "home path in $file"
    if [[ -n "$email" ]] && grep -qiF -- "$email" "$file"; then
      fail privacy "git email in $file"
    fi
    text="$(<"$file")"
    for slug in "${public_slugs[@]}"; do text="${text//"$slug"/}"; done
    grep -qiF -- "$USER" <<<"$text" && fail privacy "user name in $file"
  done
  return 0
}

size() {
  local gif="${1:?usage: checks.sh size <gif> <mp4>}" mp4="${2:?usage: checks.sh size <gif> <mp4>}"
  check_size "$gif" 8
  check_size "$mp4" 4
}

check_size() {
  local file="$1" limit_mb="$2" bytes
  [[ -f "$file" ]] || fail size "missing $file"
  bytes="$(stat -f%z "$file")"
  if ((bytes > limit_mb * 1024 * 1024)); then
    fail size "$file is $(awk "BEGIN { printf \"%.1f\", $bytes / 1048576 }") MB (limit $limit_mb MB)"
  fi
}

length() {
  local file seconds
  (($#)) || fail length "no files to check"
  for file; do
    [[ -f "$file" ]] || fail length "missing $file"
    seconds="$(ffprobe -v error -show_entries format=duration -of csv=p=0 "$file" 2>/dev/null)" \
      || fail length "cannot read the length of $file"
    [[ "$seconds" =~ ^[0-9]+(\.[0-9]+)?$ ]] || fail length "cannot read the length of $file"
    awk "BEGIN { exit !($seconds >= 30 && $seconds <= 60) }" \
      || fail length "$file is $seconds s (needs 30 to 60 s)"
  done
}

check="${1:-}"
shift || true
case "$check" in
  privacy | size | length) "$check" "$@" ;;
  *) echo "usage: checks.sh privacy|size|length <file>..." >&2; exit 2 ;;
esac
