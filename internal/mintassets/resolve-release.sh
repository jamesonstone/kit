#!/usr/bin/env bash
set -euo pipefail

if ! command -v gh >/dev/null 2>&1; then
  echo 'GitHub CLI (gh) is required to resolve the latest Mint release. Install gh on the runner or add a setup step before Mint.' >&2
  exit 1
fi

tag=$(gh api --hostname github.com repos/jamesonstone/mint/releases/latest \
  --jq 'select(.draft == false and .prerelease == false and .published_at != null) | .tag_name')
if [[ ! "$tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo 'Mint latest release is not a published stable SemVer tag.' >&2
  exit 1
fi

printf 'ref=refs/tags/%s\n' "$tag" >> "$GITHUB_OUTPUT"
