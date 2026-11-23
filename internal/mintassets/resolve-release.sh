#!/usr/bin/env bash
set -euo pipefail

tag=$(gh api --hostname github.com repos/jamesonstone/mint/releases/latest \
  --jq 'select(.draft == false and .prerelease == false and .published_at != null) | .tag_name')
if [[ ! "$tag" =~ ^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$ ]]; then
  echo 'Mint latest release is not a published stable SemVer tag.' >&2
  exit 1
fi

printf 'ref=refs/tags/%s\n' "$tag" >> "$GITHUB_OUTPUT"
