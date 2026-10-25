#!/usr/bin/env bash
# Types the showcase CLI story into a running tmux session, one character at a
# time, and waits for each screen text before the next step:
#   kirokit install platform-team, kiro-cli chat, trust, /jir, Esc, /quit
#
# It never presses Enter on a chat message, so nothing reaches the model.
# Usage: cli-drive.sh <tmux-socket> <session>
set -euo pipefail

socket="${1:?usage: cli-drive.sh <tmux-socket> <session>}"
session="${2:?usage: cli-drive.sh <tmux-socket> <session>}"

tm() { tmux -u -L "$socket" "$@"; }

# typeit types literal text at 70 ms per character.
typeit() {
  local text="$1" i
  for ((i = 0; i < ${#text}; i++)); do
    tm send-keys -t "$session" -l "${text:i:1}"
    sleep 0.07
  done
}

key() { tm send-keys -t "$session" "$@"; }

# waitfor polls the visible pane every 0.25 s for up to 30 s.
waitfor() {
  local text="$1" deadline=$((SECONDS + 30))
  while ((SECONDS < deadline)); do
    if tm capture-pane -p -t "$session" 2>/dev/null | grep -qF -- "$text"; then
      return 0
    fi
    sleep 0.25
  done
  echo "timeout waiting for $text" >&2
  exit 1
}

# capture-pane drops trailing spaces, so the "$ " prompt reads as "$".
waitfor '$'
sleep 0.8
typeit 'kirokit install platform-team'
key Enter
waitfor 'installed platform-team'
sleep 1.5

typeit 'kiro-cli chat --agent devops --v3'
key Enter
waitfor 'Trust this workspace'
sleep 1
key Enter
waitfor 'ask a question'
sleep 1

typeit '/jir'
waitfor 'jira-search'
sleep 3.5

key Escape
sleep 0.5
typeit '/quit'
key Enter
waitfor 'Session ended'
sleep 1
typeit 'exit'
key Enter
