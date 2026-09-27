#!/usr/bin/env bash
# Regenerates internal/legacy/fingerprints.json from every released Kit.
# Builds each release tag, runs `kit init` in scratch repositories under the
# default and each explicit instruction scaffold version, then fingerprints the
# output. Older releases fetch rules from GitHub; set GITHUB_TOKEN to avoid
# rate limits. Usage: scripts/harvest-legacy-fingerprints.sh [work-dir]
set -euo pipefail

repo=$(git rev-parse --show-toplevel)
work=${1:-$(mktemp -d)}
mkdir -p "$work/bin" "$work/out"
export KIT_USAGE_DISABLED=1 GIT_CONFIG_GLOBAL=/dev/null

harvest() {
  local tag=$1 src="$work/src/$1" bin="$work/bin/kit-$1"
  if [ ! -x "$bin" ]; then
    mkdir -p "$src"
    git -C "$repo" archive "$tag" | tar -x -C "$src"
    (cd "$src" && go build -o "$bin" ./cmd/kit) || { echo "$tag: build failed" >&2; return 0; }
    rm -rf "$src"
  fi
  for variant in default 1 2 3; do
    local project="$work/project/$tag-$variant"
    rm -rf "$project" && mkdir -p "$project"
    git -C "$project" init -q -b main
    git -C "$project" -c user.name=kit -c user.email=kit@example.com commit -q --allow-empty -m init
    [ "$variant" = default ] || printf 'instruction_scaffold_version: %s\n' "$variant" > "$project/.kit.yaml"
    (cd "$project" && HOME="$work/home/$tag" timeout 90 "$bin" init </dev/null >/dev/null 2>&1) || true
    rm -rf "$project/.git"
    mkdir -p "$work/out/$tag/$variant"
    cp -R "$project/." "$work/out/$tag/$variant/"
  done
}
export -f harvest
export repo work

git -C "$repo" tag --list 'v*' | xargs -P 6 -I{} bash -c 'harvest {}'
(cd "$repo" && go run ./internal/legacy/fingerprintgen "$work/out") > "$repo/internal/legacy/fingerprints.json"
echo "wrote internal/legacy/fingerprints.json from $work/out"
