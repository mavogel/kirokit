#!/usr/bin/env bash
# Record the README showcase from the real CLI and desktop app:
# assets/kirokit-showcase.gif and assets/kirokit-showcase.mp4.
#
# Everything runs in the fixed, neutral folder /tmp/kirokit-demo, which the
# EXIT trap removes. --keep leaves it (and the intermediate clips) for
# inspection. Needs tmux, asciinema, agg, ffmpeg, Node, Google Chrome, a
# signed-in kiro-cli and the SOPS age key. Run it with `task showcase`.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
here="$root/scripts/showcase"
cd "$root"

keep=0
case "${1:-}" in
  "") ;;
  --keep) keep=1 ;;
  *) echo "usage: scripts/showcase/record.sh [--keep]" >&2; exit 2 ;;
esac

demo=/tmp/kirokit-demo
out="$demo/out"
sock=kirokit-showcase
step=start
pids=()

fail() { echo "showcase: $step: $*" >&2; exit 1; }
say() { echo "showcase: $step: $*"; }

cleanup() {
  local code=$? pid
  trap - EXIT INT TERM
  tmux -L "$sock" kill-server 2>/dev/null || true
  for pid in "${pids[@]+"${pids[@]}"}"; do
    # Each background job runs in its own process group (set -m), so its
    # children (vite, the dev app) stop with it.
    kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null || true
  done
  for pid in "${pids[@]+"${pids[@]}"}"; do wait "$pid" 2>/dev/null || true; done
  rm -f "$demo/home/Library/Application Support/sops/age/keys.txt"
  if ((keep)); then
    [[ -d "$demo" ]] && echo "showcase: kept $demo"
  else
    rm -rf "${demo:?}"
  fi
  exit "$code"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'fail "command failed at line $LINENO"' ERR

# --- 1. preflight -----------------------------------------------------------
step=preflight
for tool in tmux asciinema agg ffmpeg ffprobe node npm go git kiro-cli lsof; do
  command -v "$tool" >/dev/null || fail "$tool not found"
done
[[ -d "/Applications/Google Chrome.app" ]] || fail "Google Chrome not found in /Applications"
agekey="${SOPS_AGE_KEY_FILE:-$HOME/Library/Application Support/sops/age/keys.txt}"
[[ -f "$agekey" ]] || fail "SOPS age key not found at $agekey"
for port in 34115 5173; do
  if lsof -nP -iTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
    fail "port $port is busy: close the running wails dev first"
  fi
done

rm -rf "${demo:?}"
mkdir -p "$demo"/{bin,home,kiro,work} "$out"

# --- 2. build ---------------------------------------------------------------
step=build
go build -o "$demo/bin/kirokit" ./cmd/kirokit
go build -o desktop/build/dev-bin/kirokit ./cmd/kirokit
npm ci --prefix "$here" --no-audit --no-fund --loglevel=error

# --- 3. CLI clip ------------------------------------------------------------
step=cli
say "recording the terminal"
tmux -u -L "$sock" -f "$here/tmux.conf" new-session -d -s kk -x 110 -y 32 -c "$demo/work" \
  "env PATH='$demo/bin:$PATH' KIRO_HOME='$demo/kiro' KIROKIT_REPO='$root' LANG=en_US.UTF-8 \
   BASH_SILENCE_DEPRECATION_WARNING=1 PS1='\$ ' bash --norc --noprofile"
asciinema rec --quiet --headless --overwrite --window-size 110x32 \
  -c "tmux -u -L $sock attach -t kk" "$out/cli-raw.cast" &
rec=$!
for _ in $(seq 40); do
  [[ -n "$(tmux -L "$sock" list-clients 2>/dev/null)" ]] && break
  sleep 0.25
done
[[ -n "$(tmux -L "$sock" list-clients 2>/dev/null)" ]] || fail "asciinema did not attach to tmux"
"$here/cli-drive.sh" "$sock" kk || fail "the terminal driver stopped"
wait "$rec" || fail "asciinema exited with an error"

# End the cast at "Session ended": tmux prints "[exited]" after it.
awk '{ print } /Session ended/ { exit }' "$out/cli-raw.cast" > "$out/cli.cast"
grep -q 'Session ended' "$out/cli.cast" || fail "the cast has no 'Session ended'"
# caption.mjs replays the cast with the same 3 s idle limit for its cue times.
agg --quiet --idle-time-limit 3 --last-frame-duration 2 --font-size 16 --theme github-dark \
  "$out/cli.cast" "$out/cli.gif"
say "cli.gif ready"
