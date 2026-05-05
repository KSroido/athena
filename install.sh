#!/usr/bin/env bash
# Install Athena CLI into a user-writable bin directory and ensure it is on PATH.
set -euo pipefail

APP_NAME="athena"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="${ATHENA_INSTALL_DIR:-$HOME/.local/bin}"
CONFIG_DIR="${ATHENA_CONFIG_DIR:-$HOME/.config/athena}"
BUILD_DIR="$REPO_ROOT/.build"
BINARY_SRC="$BUILD_DIR/$APP_NAME"
BINARY_DST="$INSTALL_DIR/$APP_NAME"

info() { printf '[athena-install] %s\n' "$*"; }
warn() { printf '[athena-install] WARNING: %s\n' "$*" >&2; }
fail() { printf '[athena-install] ERROR: %s\n' "$*" >&2; exit 1; }

usage() {
  cat <<'USAGE'
Usage: ./install.sh [--dir DIR] [--config-dir DIR] [--no-config] [--help]

Options:
  --dir DIR          Install binary to DIR. Default: $ATHENA_INSTALL_DIR or ~/.local/bin
  --config-dir DIR   Copy example config to DIR/athena.yaml if missing. Default: ~/.config/athena
  --no-config        Do not create config file
  --help             Show this help

Environment:
  ATHENA_INSTALL_DIR  Default install directory override
  ATHENA_CONFIG_DIR   Default config directory override

After install:
  athena -config ~/.config/athena/athena.yaml
USAGE
}

CREATE_CONFIG=1
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dir)
      [[ $# -ge 2 ]] || fail "--dir requires a value"
      INSTALL_DIR="$2"
      BINARY_DST="$INSTALL_DIR/$APP_NAME"
      shift 2
      ;;
    --config-dir)
      [[ $# -ge 2 ]] || fail "--config-dir requires a value"
      CONFIG_DIR="$2"
      shift 2
      ;;
    --no-config)
      CREATE_CONFIG=0
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      fail "unknown argument: $1"
      ;;
  esac
done

command -v go >/dev/null 2>&1 || fail "Go is required but not found in PATH"
command -v gcc >/dev/null 2>&1 || warn "gcc not found; go-sqlite3 CGO build may fail"

mkdir -p "$BUILD_DIR" "$INSTALL_DIR"

info "building $APP_NAME"
(
  cd "$REPO_ROOT"
  CGO_CFLAGS="${CGO_CFLAGS:--DSQLITE_ENABLE_FTS5}" \
  CGO_LDFLAGS="${CGO_LDFLAGS:--lm}" \
  GOTOOLCHAIN="${GOTOOLCHAIN:-auto}" \
  go build -o "$BINARY_SRC" ./cmd/athena
)

install -m 0755 "$BINARY_SRC" "$BINARY_DST"
info "installed binary: $BINARY_DST"

if [[ "$CREATE_CONFIG" -eq 1 ]]; then
  mkdir -p "$CONFIG_DIR"
  if [[ ! -f "$CONFIG_DIR/athena.yaml" ]]; then
    if [[ -f "$REPO_ROOT/config/athena.example.yaml" ]]; then
      cp "$REPO_ROOT/config/athena.example.yaml" "$CONFIG_DIR/athena.yaml"
      info "created config: $CONFIG_DIR/athena.yaml"
    else
      warn "example config not found: $REPO_ROOT/config/athena.example.yaml"
    fi
  else
    info "config already exists: $CONFIG_DIR/athena.yaml"
  fi
fi

add_path_line() {
  local shell_rc="$1"
  local dir="$2"
  local line="export PATH=\"$dir:\$PATH\""

  mkdir -p "$(dirname "$shell_rc")"
  touch "$shell_rc"
  if ! grep -F "$line" "$shell_rc" >/dev/null 2>&1; then
    {
      printf '\n# Athena CLI\n'
      printf '%s\n' "$line"
    } >> "$shell_rc"
    info "added PATH entry to $shell_rc"
  else
    info "PATH entry already present in $shell_rc"
  fi
}

case ":$PATH:" in
  *":$INSTALL_DIR:"*)
    info "$INSTALL_DIR is already on PATH"
    ;;
  *)
    shell_name="$(basename "${SHELL:-}")"
    case "$shell_name" in
      zsh) add_path_line "$HOME/.zshrc" "$INSTALL_DIR" ;;
      fish)
        fish_config="$HOME/.config/fish/config.fish"
        mkdir -p "$(dirname "$fish_config")"
        fish_line="fish_add_path $INSTALL_DIR"
        touch "$fish_config"
        if ! grep -F "$fish_line" "$fish_config" >/dev/null 2>&1; then
          {
            printf '\n# Athena CLI\n'
            printf '%s\n' "$fish_line"
          } >> "$fish_config"
          info "added PATH entry to $fish_config"
        else
          info "PATH entry already present in $fish_config"
        fi
        ;;
      *) add_path_line "$HOME/.bashrc" "$INSTALL_DIR" ;;
    esac
    export PATH="$INSTALL_DIR:$PATH"
    ;;
esac

if command -v "$APP_NAME" >/dev/null 2>&1; then
  resolved="$(command -v "$APP_NAME")"
else
  resolved="$BINARY_DST"
fi

info "athena command: $resolved"
info "run: athena -config $CONFIG_DIR/athena.yaml"
info "if current shell cannot find athena, run: export PATH=\"$INSTALL_DIR:\$PATH\" or open a new shell"
