#!/usr/bin/env bash
set -euo pipefail

# gaggimate-cli installer
# Downloads the pre-built binary for your OS/architecture and installs it to ~/.local/bin.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/aveseli/gaggimate-cli/main/install.sh | bash
#   Or download and run locally:
#   bash install.sh

REPO="aveseli/gaggimate-cli"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

# ─── Colors ───────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
RESET='\033[0m'

info()  { printf "${CYAN}%s${RESET}\n" "$*"; }
ok()    { printf "${GREEN}✓ %s${RESET}\n" "$*"; }
warn()  { printf "${YELLOW}⚠ %s${RESET}\n" "$*"; }
err()   { printf "${RED}Error: %s${RESET}\n" "$*" >&2; exit 1; }

# ─── Detect OS and architecture ──────────────────────────────────

detect_platform() {
    local os arch

    case "$(uname -s)" in
        Linux*)     os="linux" ;;
        Darwin*)    os="darwin" ;;
        CYGWIN*|MINGW*|MSYS*)  os="windows" ;;
        *)          err "Unsupported OS: $(uname -s)" ;;
    esac

    case "$(uname -m)" in
        x86_64|amd64)   arch="amd64" ;;
        aarch64|arm64)   arch="arm64" ;;
        *)               err "Unsupported architecture: $(uname -m)" ;;
    esac

    echo "${os}-${arch}"
}

# ─── Get latest version from GitHub API ──────────────────────────

get_latest_version() {
    local url="https://api.github.com/repos/${REPO}/releases/latest"
    local version

    if command -v curl &>/dev/null; then
        version=$(curl -fsSL "$url" | grep '"tag_name"' | head -1 | sed -E 's/.*"v([^"]+)".*/\1/')
    elif command -v wget &>/dev/null; then
        version=$(wget -qO- "$url" | grep '"tag_name"' | head -1 | sed -E 's/.*"v([^"]+)".*/\1/')
    else
        err "Neither curl nor wget found. Please install one and try again."
    fi

    if [ -z "$version" ]; then
        err "Could not determine latest version from GitHub."
    fi

    echo "$version"
}

# ─── Download binary ─────────────────────────────────────────────

download_binary() {
    local version="$1"
    local platform="$2"
    local ext=""

    if [[ "$platform" == windows-* ]]; then
        ext=".exe"
    fi

    local binary="gaggimate-${platform}${ext}"
    local url="https://github.com/${REPO}/releases/download/v${version}/${binary}"
    local tmp_dir

    tmp_dir=$(mktemp -d)
    local tmp_file="${tmp_dir}/${binary}"

    info "Downloading ${binary} v${version}..."
    if command -v curl &>/dev/null; then
        curl -fsSL -o "$tmp_file" "$url" || err "Download failed. Check your connection and try again."
    elif command -v wget &>/dev/null; then
        wget -qO "$tmp_file" "$url" || err "Download failed. Check your connection and try again."
    fi

    chmod +x "$tmp_file"
    echo "$tmp_file"
}

# ─── Main ─────────────────────────────────────────────────────────

main() {
    printf "\n${BOLD}gaggimate-cli installer${RESET}\n\n"

    local platform
    platform=$(detect_platform)
    info "Detected platform: ${platform}"

    local version
    version=$(get_latest_version)
    ok "Latest version: v${version}"

    # Download
    local tmp_binary
    tmp_binary=$(download_binary "$version" "$platform")

    # Install
    mkdir -p "$INSTALL_DIR"

    local ext=""
    if [[ "$platform" == windows-* ]]; then
        ext=".exe"
    fi

    local install_path="${INSTALL_DIR}/gaggimate${ext}"
    mv "$tmp_binary" "$install_path"
    ok "Installed to ${install_path}"

    # Clean up temp dir
    rm -rf "$(dirname "$tmp_binary")"

    # Check PATH
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*)  in_path=true ;;
        *)                     in_path=false ;;
    esac

    if [ "$in_path" = false ]; then
        printf "\n${YELLOW}⚠ ${INSTALL_DIR} is not in your PATH.${RESET}\n\n"
        printf "Add it by running one of the following:\n\n"

        local shell_name
        shell_name=$(basename "${SHELL:-/bin/bash}")

        case "$shell_name" in
            zsh)
                printf "  ${CYAN}echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc${RESET}\n"
                printf "  ${CYAN}source ~/.zshrc${RESET}\n\n"
                ;;
            fish)
                printf "  ${CYAN}fish_add_path ~/.local/bin${RESET}\n\n"
                ;;
            *)
                printf "  ${CYAN}echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc${RESET}\n"
                printf "  ${CYAN}source ~/.bashrc${RESET}\n\n"
                ;;
        esac
    else
        ok "PATH is configured correctly."
    fi

    printf "\n${GREEN}${BOLD}Done!${RESET} Restart your shell, then run:\n\n"
    printf "  ${CYAN}gaggimate version${RESET}\n\n"
}

main "$@"
