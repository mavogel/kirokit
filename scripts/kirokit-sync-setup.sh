#!/usr/bin/env bash
# Set up automatic `kirokit sync` on this machine: clones the config repo and
# registers a launchd agent (macOS) or systemd user timer (Linux) that keeps
# the team's Kiro config current.
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: kirokit-sync-setup.sh --team <team> --repo-url <url> [options]
       kirokit-sync-setup.sh --team <team> --remove [options]

  --team <team>        team to install and keep in sync (required)
  --repo-url <url>     git URL of the kirokit config repository (required unless --remove)
  --dir <path>         where to keep the checkout (default: ~/.kirokit-config)
  --interval <hours>   how often to sync (default: 4)
  --latest-tag         follow the newest v* release tag instead of the branch
  --dry-run            print the scheduler definition and exit; change nothing
  --remove             unregister the scheduler (the checkout and installed config stay)

Requires git and kirokit on PATH. Logs go to ~/.kirokit-sync.log.
EOF
}

team="" repo_url="" dir="$HOME/.kirokit-config" interval=4 latest_tag="" dry_run="" remove=""
while [ $# -gt 0 ]; do
  case "$1" in
    --team) team="${2:?--team needs a value}"; shift 2 ;;
    --repo-url) repo_url="${2:?--repo-url needs a value}"; shift 2 ;;
    --dir) dir="${2:?--dir needs a value}"; shift 2 ;;
    --interval) interval="${2:?--interval needs a value}"; shift 2 ;;
    --latest-tag) latest_tag="--latest-tag"; shift ;;
    --dry-run) dry_run=1; shift ;;
    --remove) remove=1; shift ;;
    -h|--help) usage; exit 0 ;;
    *) echo "unknown option: $1" >&2; usage >&2; exit 2 ;;
  esac
done

[ -n "$team" ] || { echo "--team is required" >&2; usage >&2; exit 2; }
[[ "$team" =~ ^[A-Za-z0-9._-]+$ ]] || { echo "invalid team name: $team" >&2; exit 2; }
[[ "$interval" =~ ^[1-9][0-9]*$ ]] || { echo "--interval must be a positive integer (hours)" >&2; exit 2; }
if [ -z "$remove" ]; then
  [ -n "$repo_url" ] || { echo "--repo-url is required" >&2; usage >&2; exit 2; }
  command -v git >/dev/null || { echo "git not found on PATH" >&2; exit 1; }
  kirokit_bin="$(command -v kirokit)" || { echo "kirokit not found on PATH" >&2; exit 1; }
else
  kirokit_bin="kirokit"
fi

dir="${dir/#\~/$HOME}"
log="$HOME/.kirokit-sync.log"
label="io.github.mavogel.kirokit-sync.$team"
sync_args=(sync "$team" --repo "$dir")
[ -n "$latest_tag" ] && sync_args+=("$latest_tag")
# Scheduled jobs get a minimal PATH; give them kirokit's and git's directories.
job_path="$(dirname "$kirokit_bin"):$(command -v git >/dev/null && dirname "$(command -v git)" || true):/usr/local/bin:/usr/bin:/bin"

xml_escape() { sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' <<<"$1"; }

plist_path="$HOME/Library/LaunchAgents/$label.plist"
render_plist() {
  local args="" a
  for a in "$kirokit_bin" "${sync_args[@]}"; do
    args+="    <string>$(xml_escape "$a")</string>"$'\n'
  done
  cat <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key><string>$label</string>
  <key>ProgramArguments</key>
  <array>
$args  </array>
  <key>EnvironmentVariables</key>
  <dict><key>PATH</key><string>$(xml_escape "$job_path")</string></dict>
  <key>StartInterval</key><integer>$((interval * 3600))</integer>
  <key>RunAtLoad</key><true/>
  <key>StandardOutPath</key><string>$(xml_escape "$log")</string>
  <key>StandardErrorPath</key><string>$(xml_escape "$log")</string>
</dict>
</plist>
EOF
}

unit_dir="${XDG_CONFIG_HOME:-$HOME/.config}/systemd/user"
unit_name="kirokit-sync-$team"
render_service() {
  local cmd="" a
  for a in "$kirokit_bin" "${sync_args[@]}"; do cmd+=" $(printf '%q' "$a")"; done
  cat <<EOF
[Unit]
Description=Sync kirokit config for team $team

[Service]
Type=oneshot
Environment=PATH=$job_path
ExecStart=${cmd# }
StandardOutput=append:$log
StandardError=append:$log
EOF
}
render_timer() {
  cat <<EOF
[Unit]
Description=Periodically sync kirokit config for team $team

[Timer]
OnBootSec=2min
OnUnitActiveSec=${interval}h
Persistent=true

[Install]
WantedBy=timers.target
EOF
}

os="$(uname -s)"
case "$os" in
  Darwin|Linux) ;;
  *) echo "unsupported OS: $os (macOS and Linux only)" >&2; exit 1 ;;
esac

if [ -n "$dry_run" ]; then
  if [ "$os" = Darwin ]; then
    echo "# $plist_path"; render_plist
  else
    echo "# $unit_dir/$unit_name.service"; render_service
    echo; echo "# $unit_dir/$unit_name.timer"; render_timer
  fi
  exit 0
fi

if [ -n "$remove" ]; then
  if [ "$os" = Darwin ]; then
    launchctl bootout "gui/$(id -u)/$label" 2>/dev/null || true
    rm -f "$plist_path"
  else
    systemctl --user disable --now "$unit_name.timer" 2>/dev/null || true
    rm -f "$unit_dir/$unit_name.service" "$unit_dir/$unit_name.timer"
    systemctl --user daemon-reload
  fi
  echo "removed scheduled sync for $team (checkout $dir and installed config left in place)"
  exit 0
fi

if [ -d "$dir/.git" ]; then
  echo "using existing checkout $dir"
else
  [ ! -e "$dir" ] || { echo "$dir exists and is not a git checkout" >&2; exit 1; }
  git clone --quiet "$repo_url" "$dir"
fi

# First sync runs in the foreground so a broken setup fails here, not silently later.
"$kirokit_bin" "${sync_args[@]}"

if [ "$os" = Darwin ]; then
  mkdir -p "$(dirname "$plist_path")"
  render_plist >"$plist_path"
  launchctl bootout "gui/$(id -u)/$label" 2>/dev/null || true
  launchctl bootstrap "gui/$(id -u)" "$plist_path"
  echo "registered launchd agent $label (every ${interval}h, log: $log)"
else
  mkdir -p "$unit_dir"
  render_service >"$unit_dir/$unit_name.service"
  render_timer >"$unit_dir/$unit_name.timer"
  systemctl --user daemon-reload
  systemctl --user enable --now "$unit_name.timer"
  echo "registered systemd timer $unit_name (every ${interval}h, log: $log)"
fi
