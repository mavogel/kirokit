#!/usr/bin/env bash
# Spike: the git behaviour the branch-checkout design relies on.
set -u
S="$1"; rm -rf "$S"; mkdir -p "$S"; cd "$S"
gg() { git -c user.name=t -c user.email=t@t -c commit.gpgsign=false -c tag.gpgsign=false -c core.hooksPath=/dev/null "$@"; }
echo "git: $(git --version)"
mkdir origin; cd origin; gg init -q -b main; echo 1 > f; gg add f; gg commit -qm one; gg tag v1.0.0
gg branch feature; gg checkout -q feature; echo f > f; gg commit -qam feat; gg checkout -q main
echo 2 > f; gg commit -qam two; gg tag v1.1.0; gg tag same; gg branch same; cd ..
gg clone -q origin clone; cd clone; gg checkout -q v1.1.0 2>/dev/null
echo "== 1 list refs offline"; gg for-each-ref --format='%(refname)' --sort=-v:refname refs/remotes/origin refs/tags
echo "== 2 switch to remote-only branch from detached tag"
gg show-ref --verify -q refs/remotes/origin/feature && echo has-remote
gg show-ref --verify -q refs/heads/feature || echo no-local
gg checkout -q --track -b feature origin/feature && gg symbolic-ref --short HEAD
echo "== 3 plain pull on detached tag fails today"; gg checkout -q v1.0.0; gg pull --ff-only -q 2>&1 | head -1
echo "symbolic-ref exit (1 = detached): $(gg symbolic-ref -q HEAD >/dev/null; echo $?)"
echo "== 4 non-conflicting tracked edit is carried silently"; echo 1 > f; echo extra >> f
gg checkout -q --detach refs/tags/v1.1.0 2>&1; echo "checkout exit $?"; echo "porcelain: $(gg status --porcelain --untracked-files=no)"
gg checkout -q -- f
echo "== 5 diverged local branch is detectable before switching"
gg checkout -q feature; echo local > f; gg commit -qam local
cd ../origin; gg checkout -q feature; echo up > f; gg commit -qam up; gg checkout -q main; cd ../clone
gg checkout -q main; gg fetch -q; gg merge-base --is-ancestor feature origin/feature; echo "is-ancestor exit $? (1 = not fast-forward)"
echo "== 6 a name that is both branch and tag"
gg show-ref --verify -q refs/remotes/origin/same && echo "branch same"; gg show-ref --verify -q refs/tags/same && echo "tag same"
echo "== 7 detach at a tag explicitly"; gg checkout -q --detach refs/tags/v1.0.0 && gg describe --tags --exact-match HEAD
echo "== 8 unknown name"; gg show-ref --verify -q refs/remotes/origin/nope || echo "no branch nope"; gg show-ref --verify -q refs/tags/nope || echo "no tag nope"
echo "== 9 check-ref-format"; for n in main feature/x v1.0.0 'a..b' '-x' 'a b'; do gg check-ref-format --branch "$n" >/dev/null 2>&1 && echo "ok   $n" || echo "bad  $n"; done
