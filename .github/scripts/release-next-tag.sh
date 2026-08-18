#!/usr/bin/env bash
set -euo pipefail

head_ref="${1:-HEAD}"
head_sha="$(git rev-parse "${head_ref}^{commit}")"

head_tags="$(
  git tag --points-at "${head_sha}" --list 'v3.*' \
    | grep -E '^v3\.[0-9]+\.[0-9]+$' \
    | sort -V || true
)"

head_tag_count=0
if [[ -n "${head_tags}" ]]; then
  head_tag_count="$(printf '%s\n' "${head_tags}" | wc -l | tr -d ' ')"
fi

if (( head_tag_count > 1 )); then
  echo "Multiple v3 release tags already point at ${head_sha}: ${head_tags}" >&2
  exit 1
fi

latest_tag="$({
  git tag --list 'v3.*' \
    | grep -E '^v3\.[0-9]+\.[0-9]+$' \
    | sort -V \
    | tail -n 1
} || true)"

if (( head_tag_count == 1 )); then
  next_tag="${head_tags}"
  if [[ -n "${latest_tag}" && "${next_tag}" != "${latest_tag}" ]]; then
    echo "HEAD tag ${next_tag} is not the latest v3 release tag ${latest_tag}." >&2
    exit 1
  fi
  mode="reuse"
elif [[ -z "${latest_tag}" ]]; then
  next_tag="v3.0.0"
  mode="create"
else
  if [[ ! "${latest_tag}" =~ ^v3\.([0-9]+)\.([0-9]+)$ ]]; then
    echo "Invalid v3 semantic tag format: ${latest_tag}" >&2
    exit 1
  fi
  minor="${BASH_REMATCH[1]}"
  patch="${BASH_REMATCH[2]}"
  next_tag="v3.${minor}.$((patch + 1))"
  mode="create"
fi

if git rev-parse --verify --quiet "refs/tags/${next_tag}" >/dev/null; then
  tag_sha="$(git rev-list -n 1 "refs/tags/${next_tag}")"
  if [[ "${tag_sha}" != "${head_sha}" ]]; then
    echo "Tag ${next_tag} points to ${tag_sha}, not ${head_sha}; refusing to move it." >&2
    exit 1
  fi
  mode="reuse"
fi

printf 'next_tag=%s\n' "${next_tag}"
printf 'mode=%s\n' "${mode}"
printf 'head_sha=%s\n' "${head_sha}"
