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
owned=0 # set once this run has created $demo; until then cleanup touches nothing shared

fail() { echo "showcase: $step: $*" >&2; exit 1; }
say() { echo "showcase: $step: $*"; }

cleanup() {
  local code=$? pid
  trap - EXIT INT TERM
  ((owned)) || exit "$code"
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

if [[ -f "$demo/pid" ]] && kill -0 "$(<"$demo/pid")" 2>/dev/null; then
  fail "another showcase run (pid $(<"$demo/pid")) is using $demo"
fi
rm -rf "${demo:?}"
# Fails if anything recreated the folder in between; only this user can enter it.
mkdir -m 700 "$demo" || fail "could not create $demo"
owned=1
echo $$ > "$demo/pid"
mkdir "$demo"/{bin,home,kiro,work} "$out"

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

# --- 4. app clip ------------------------------------------------------------
step=app
say "starting the app in Wails dev mode"
# The throwaway HOME holds the app's settings, clone and Kiro home; the Go and
# npm caches stay the real ones so nothing is rebuilt or downloaded again.
gocache="$(go env GOCACHE)"
gomodcache="$(go env GOMODCACHE)"
gopath="$(go env GOPATH)"
npmcache="$(npm config get cache)"
# Link, never copy, the age key to the default SOPS path in the throwaway home.
keydir="$demo/home/Library/Application Support/sops/age"
mkdir -p "$keydir"
ln -s "$agekey" "$keydir/keys.txt"

set -m # own process group, so stopping it also stops vite and the dev app
(
  cd desktop
  exec env -u KIRO_HOME -u KIROKIT_REPO -u SOPS_AGE_KEY_FILE \
    -u XDG_CONFIG_HOME -u XDG_DATA_HOME -u XDG_CACHE_HOME \
    HOME="$demo/home" GOCACHE="$gocache" GOMODCACHE="$gomodcache" GOPATH="$gopath" \
    npm_config_cache="$npmcache" go tool wails dev
) >"$out/wails.log" 2>&1 &
wails=$!
set +m
pids+=("$wails")

for _ in $(seq 720); do
  grep -q 'Using DevServer URL: http://localhost:34115' "$out/wails.log" && break
  kill -0 "$wails" 2>/dev/null || fail "wails dev exited; see $out/wails.log (run with --keep)"
  sleep 0.25
done
grep -q 'Using DevServer URL: http://localhost:34115' "$out/wails.log" \
  || fail "wails dev did not serve http://localhost:34115 within 180 s"

say "recording the app"
node "$here/app-drive.mjs" "$out" || fail "the app driver stopped"
kill -TERM -- "-$wails" 2>/dev/null || true
wait "$wails" 2>/dev/null || true
pids=()
say "app.webm ready"

# --- 5. captions ------------------------------------------------------------
step=captions
node "$here/caption.mjs" "$out" || fail "caption rendering stopped"
# shellcheck source=/dev/null
source "$out/cues.env"

# --- 6. compose -------------------------------------------------------------
step=compose
say "composing showcase.mp4 and showcase.gif"
title=2.5 # seconds on the title card
ending=3.5 # seconds on the end card
fade=0.4 # into the CLI clip and into the end card
cross=0.5 # between the CLI and the app clip
calc() { awk "BEGIN { printf \"%.3f\", $1 }"; }
off1="$(calc "$title - $fade")"
off2="$(calc "$off1 + $CLI_DURATION - $cross")"
off3="$(calc "$off2 + $APP_DURATION - $fade")"

# Each clip sits in the 1280x800 video area on #13111f inside the rounded frame;
# its caption fills the 96 px band below. The CLI clip is padded with the
# terminal's own background to the frame size.
norm="fps=30,settb=AVTB,setsar=1,format=yuv420p"
ffmpeg -v error -y \
  -loop 1 -framerate 30 -t "$title" -i "$out/title.png" \
  -i "$out/cli.gif" \
  -i "$out/app.webm" \
  -loop 1 -framerate 30 -t "$ending" -i "$out/end.png" \
  -i "$out/frame.png" -i "$out/frame.png" \
  -i "$out/cap-1.png" -i "$out/cap-2.png" -i "$out/cap-3.png" -i "$out/cap-4.png" \
  -filter_complex "
    color=c=0x13111f:s=1280x896:r=30:d=${CLI_DURATION}[bg1];
    color=c=0x13111f:s=1280x896:r=30:d=${APP_DURATION}[bg2];
    [0:v]${norm}[title];
    [1:v]fps=30,pad=1180:760:(ow-iw)/2:(oh-ih)/2:color=0x171b21[c0];
    [bg1][c0]overlay=50:20[c1];
    [c1][4:v]overlay=50:20[c2];
    [c2][6:v]overlay=0:800:enable='lt(t,$CAP2_AT)'[c3];
    [c3][7:v]overlay=0:800:enable='gte(t,$CAP2_AT)',${norm}[cli];
    [2:v]split=2[a][b];
    [a]trim=start=$APP_A0:end=$APP_A1,setpts=PTS-STARTPTS[a1];
    [b]trim=start=$APP_B0:end=$APP_B1,setpts=PTS-STARTPTS[b1];
    [a1][b1]concat=n=2:v=1,fps=30,scale=1180:760[p0];
    [bg2][p0]overlay=50:20[p1];
    [p1][5:v]overlay=50:20[p2];
    [p2][8:v]overlay=0:800:enable='lt(t,$CAP4_AT)'[p3];
    [p3][9:v]overlay=0:800:enable='gte(t,$CAP4_AT)',${norm}[app];
    [3:v]${norm}[end];
    [title][cli]xfade=transition=fade:duration=$fade:offset=${off1}[x1];
    [x1][app]xfade=transition=fade:duration=$cross:offset=${off2}[x2];
    [x2][end]xfade=transition=fade:duration=$fade:offset=${off3}[v]" \
  -map "[v]" -an -c:v libx264 -preset slow -crf 24 -pix_fmt yuv420p -movflags +faststart \
  "$out/showcase.mp4"

ffmpeg -v error -y -i "$out/showcase.mp4" -filter_complex \
  "fps=12,scale=1000:-1:flags=lanczos,split[s0][s1];
   [s0]palettegen=max_colors=128:stats_mode=diff[p];
   [s1][p]paletteuse=dither=bayer:bayer_scale=4:diff_mode=rectangle" \
  -loop 0 "$out/showcase.gif"
say "showcase.mp4 and showcase.gif ready"

# --- 7. checks --------------------------------------------------------------
# checks.sh prints its own "showcase: <check>: <reason>".
step=checks
"$here/checks.sh" privacy "$out/cli.cast" "$out/app-dock.txt" "$out/captions.txt" || exit 1
"$here/checks.sh" size "$out/showcase.gif" "$out/showcase.mp4" || exit 1
"$here/checks.sh" length "$out/showcase.mp4" "$out/showcase.gif" || exit 1

# --- 8. publish -------------------------------------------------------------
# Copy next to the target first, then rename, so assets/ never holds half a file.
step=publish
for ext in gif mp4; do
  cp "$out/showcase.$ext" "assets/.kirokit-showcase.$ext.tmp"
  mv -f "assets/.kirokit-showcase.$ext.tmp" "assets/kirokit-showcase.$ext"
done
say "wrote assets/kirokit-showcase.gif ($(du -h assets/kirokit-showcase.gif | cut -f1)) and assets/kirokit-showcase.mp4 ($(du -h assets/kirokit-showcase.mp4 | cut -f1))"
