#!/bin/sh
# Resolve the backend build version.
#
# order:
#   1. $GIT_VERSION env (CI / explicit override), e.g. "v2.4.3"
#   2. Exact tag on HEAD (git describe --tags --exact-match)
#   3. "develop" (fallback: not on a tag, so this build is NOT a release)

set -eu

resolve() {
	if [ -n "${GIT_VERSION:-}" ]; then
		printf '%s' "$GIT_VERSION"
		return
	fi

	if tag="$(git describe --tags --exact-match 2>/dev/null || true)" && [ -n "$tag" ]; then
		printf '%s' "$tag"
		return
	fi

	printf '%s' "develop"
}

resolve
