#!/usr/bin/env sh
set -eu

BUMP="${1:-patch}"
FIRST_VERSION="${HIDEMYENV_FIRST_VERSION:-v0.1.1}"
DRY_RUN="${HIDEMYENV_DRY_RUN:-0}"

info() {
  printf '%s\n' "$1"
}

fail() {
  printf 'hidemyenv release: %s\n' "$1" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage:
  ./scripts/release.sh [patch|minor|major]

Environment:
  HIDEMYENV_DRY_RUN=1          Print the next version without creating/pushing a tag
  HIDEMYENV_FIRST_VERSION=vX.Y.Z  First version when no previous semver tag exists

Examples:
  ./scripts/release.sh
  ./scripts/release.sh minor
  HIDEMYENV_DRY_RUN=1 ./scripts/release.sh patch
EOF
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

latest_tag() {
  git for-each-ref --sort=-v:refname --format='%(refname:short)' refs/tags/v* 2>/dev/null | while IFS= read -r tag; do
    case "$tag" in
      v[0-9]*.[0-9]*.[0-9]*)
        printf '%s' "$tag"
        break
        ;;
    esac
  done
}

next_version() {
  latest="$1"

  if [ -z "$latest" ]; then
    printf '%s' "$FIRST_VERSION"
    return 0
  fi

  raw="${latest#v}"
  major="${raw%%.*}"
  rest="${raw#*.}"
  minor="${rest%%.*}"
  patch="${rest#*.}"
  patch="${patch%%[-+]*}"

  case "$major$minor$patch" in
    *[!0-9]*) fail "latest tag is not semver-compatible: $latest" ;;
  esac

  case "$BUMP" in
    patch)
      patch=$((patch + 1))
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    major)
      major=$((major + 1))
      minor=0
      patch=0
      ;;
    -h|--help|help)
      usage
      exit 0
      ;;
    *)
      fail "unknown bump type: $BUMP"
      ;;
  esac

  printf 'v%s.%s.%s' "$major" "$minor" "$patch"
}

ensure_clean_worktree() {
  if [ -n "$(git status --porcelain)" ]; then
    fail "working tree is not clean; commit or stash changes before releasing"
  fi
}

need_cmd git
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || fail "not inside a git repository"

latest="$(latest_tag)"
next="$(next_version "$latest")"

if [ -n "$latest" ]; then
  info "latest version: $latest"
else
  info "latest version: none"
fi
info "next version: $next"

if [ "$DRY_RUN" = "1" ]; then
  exit 0
fi

ensure_clean_worktree

git tag "$next"
git push origin "$next"

info "pushed $next"
info "GitHub Actions will build and publish the release from this tag."
