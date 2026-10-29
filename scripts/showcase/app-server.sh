#!/usr/bin/env bash
# Start the desktop app for the showcase's app clip: a Wails v3 server build
# (`go build -tags server`) that serves the real page and its Go calls on
# http://localhost:34115, where app-drive.mjs drives it from headless Chrome.
#
#   scripts/showcase/app-server.sh <out-dir> <home>
#
# The server answers every bound method, the shell tab's included, so it only
# ever listens on loopback: WAILS_SERVER_HOST and WAILS_SERVER_PORT are set here
# (an inherited value never reaches it), and the run stops unless every listener
# on the port is the server's own on 127.0.0.1. The binary lives in <out-dir>;
# nothing ships it. The server runs in its own process group with HOME=<home>;
# its pid (the group id) is printed for the caller to stop it.
set -euo pipefail

[ "$#" -eq 2 ] || { echo "usage: scripts/showcase/app-server.sh <out-dir> <home>" >&2; exit 2; }
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
out="$1"
home="$2"
port=34115

fail() { echo "showcase: app-server: $*" >&2; exit 1; }

{
  (cd "$root/desktop/frontend" && npm ci --no-audit --no-fund --loglevel=error && npm run build)
  (cd "$root/desktop" && go build -tags server -o "$out/kirokit-server" .)
} >"$out/app-build.log" 2>&1 || fail "the server build failed; see $out/app-build.log"

# The working directory is desktop/, so the app's terminals use the dev CLI
# (desktop/build/dev-bin/kirokit) that record.sh builds.
set -m
(
  cd "$root/desktop"
  exec env -u KIRO_HOME -u KIROKIT_REPO -u SOPS_AGE_KEY_FILE \
    -u XDG_CONFIG_HOME -u XDG_DATA_HOME -u XDG_CACHE_HOME \
    HOME="$home" WAILS_SERVER_HOST=127.0.0.1 WAILS_SERVER_PORT="$port" "$out/kirokit-server"
) >"$out/app-server.log" 2>&1 &
pid=$!
set +m

stop() { kill -TERM -- "-$pid" 2>/dev/null || kill -TERM "$pid" 2>/dev/null || true; }

for _ in $(seq 720); do
  kill -0 "$pid" 2>/dev/null || fail "the app server exited; see $out/app-server.log"
  curl -fsS -o /dev/null "http://127.0.0.1:$port/" 2>/dev/null && break
  sleep 0.25
done
curl -fsS -o /dev/null "http://127.0.0.1:$port/" 2>/dev/null \
  || { stop; fail "the app server did not answer on http://localhost:$port within 180 s"; }

# Every listener on the port: its pid ("p...") and address ("n...").
listeners="$(lsof -nP -iTCP:"$port" -sTCP:LISTEN -Fpn 2>/dev/null || true)"
[ -n "$listeners" ] || { stop; fail "nothing listens on port $port"; }
owner=""
while IFS= read -r line; do
  case "$line" in
  p*) owner="${line#p}" ;;
  n*)
    addr="${line#n}"
    [ "$owner" = "$pid" ] || { stop; fail "port $port is held by another process (pid $owner, $addr)"; }
    case "$addr" in
    127.0.0.1:"$port") ;;
    *) stop; fail "the app server listens on $addr, not only on 127.0.0.1" ;;
    esac
    ;;
  esac
done <<<"$listeners"

echo "$pid"
