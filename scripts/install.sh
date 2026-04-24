#!/usr/bin/env sh
set -eu

BINARY_NAME="hidemyenv"
INSTALL_DIR="${HIDEMYENV_INSTALL_DIR:-$HOME/.local/bin}"
REPO="${HIDEMYENV_REPO:-}"
VERSION="${HIDEMYENV_VERSION:-latest}"

info() {
  printf '%s\n' "$1"
}

fail() {
  printf 'hidemyenv installer: %s\n' "$1" >&2
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

source_version() {
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git describe --tags --dirty --always 2>/dev/null || printf 'dev'
  else
    printf 'dev'
  fi
}

source_commit() {
  if command -v git >/dev/null 2>&1 && git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    git rev-parse --short HEAD 2>/dev/null || printf 'local'
  else
    printf 'local'
  fi
}

build_date() {
  date -u +%Y-%m-%dT%H:%M:%SZ
}

detect_platform() {
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  arch="$(uname -m)"

  case "$os" in
    darwin) os="darwin" ;;
    linux) os="linux" ;;
    *) fail "unsupported operating system: $os" ;;
  esac

  case "$arch" in
    x86_64 | amd64) arch="amd64" ;;
    arm64 | aarch64) arch="arm64" ;;
    *) fail "unsupported architecture: $arch" ;;
  esac

  printf '%s_%s' "$os" "$arch"
}

install_binary() {
  src="$1"
  mkdir -p "$INSTALL_DIR"

  if [ -x "$INSTALL_DIR/$BINARY_NAME" ]; then
    old_version="$($INSTALL_DIR/$BINARY_NAME version 2>/dev/null || true)"
    if [ -n "$old_version" ]; then
      info "current version: $old_version"
    else
      info "current version: unknown"
    fi
  fi

  cp "$src" "$INSTALL_DIR/$BINARY_NAME"
  chmod 0755 "$INSTALL_DIR/$BINARY_NAME"
  info "installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"

  new_version="$($INSTALL_DIR/$BINARY_NAME version 2>/dev/null || true)"
  if [ -n "$new_version" ]; then
    info "installed version: $new_version"
  fi

  case ":$PATH:" in
    *":$INSTALL_DIR:"*) ;;
    *)
      info "add this to your shell profile if $BINARY_NAME is not found:"
      info "  export PATH=\"$INSTALL_DIR:\$PATH\""
      ;;
  esac
}

install_from_source() {
  need_cmd go
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT INT TERM
  info "building $BINARY_NAME from local source"
  version="$(source_version)"
  commit="$(source_commit)"
  date="$(build_date)"
  go build -ldflags="-X hidemyenv/internal/version.Version=$version -X hidemyenv/internal/version.Commit=$commit -X hidemyenv/internal/version.Date=$date" -o "$tmp/$BINARY_NAME" ./cmd/hidemyenv
  install_binary "$tmp/$BINARY_NAME"
}

download_release() {
  platform="$(detect_platform)"
  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT INT TERM

  if command -v curl >/dev/null 2>&1; then
    fetch='curl -fsSL'
  elif command -v wget >/dev/null 2>&1; then
    fetch='wget -qO-'
  else
    fail "missing curl or wget"
  fi

  if [ -z "$REPO" ]; then
    fail "HIDEMYENV_REPO is required for release install, for example: HIDEMYENV_REPO=owner/hidemyenv"
  fi

  if [ "$VERSION" = "latest" ]; then
    base_url="https://github.com/$REPO/releases/latest/download"
  else
    base_url="https://github.com/$REPO/releases/download/$VERSION"
  fi

  archive="$BINARY_NAME-$platform.tar.gz"
  url="$base_url/$archive"
  info "downloading $url"
  # shellcheck disable=SC2086
  $fetch "$url" > "$tmp/$archive"
  tar -xzf "$tmp/$archive" -C "$tmp"

  if [ ! -f "$tmp/$BINARY_NAME" ]; then
    fail "release archive did not contain $BINARY_NAME"
  fi

  install_binary "$tmp/$BINARY_NAME"
}

if [ -f "go.mod" ] && [ -d "cmd/hidemyenv" ]; then
  install_from_source
else
  download_release
fi

info "run: $BINARY_NAME --help"
