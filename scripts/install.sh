#!/usr/bin/env bash
# cd_project installation script
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
BINARY_NAME="cd_project"
INSTALL_DIR="${HOME}/.local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Check if fd is available
check_fd() {
    if command -v fd &>/dev/null; then
        log_info "fd found - will use for faster scanning"
    else
        log_warn "fd not found - using native Go scanner"
        echo "       Install fd for ~10x faster scanning: https://github.com/sharkdp/fd"
    fi
}

# Backup RC file
backup_rc() {
    local rc_file="$1"
    if [ -f "$rc_file" ]; then
        cp "$rc_file" "${rc_file}.backup.$(date +%Y%m%d%H%M%S)"
        log_info "Backed up $rc_file"
    fi
}

# Add source line if not present
add_source_line() {
    local rc_file="$1"
    local integration_file="$2"
    local source_line="[ -f \"$integration_file\" ] && source \"$integration_file\""

    if ! grep -qF "cd_project" "$rc_file" 2>/dev/null; then
        backup_rc "$rc_file"
        {
            echo ""
            echo "# cd_project integration"
            echo "$source_line"
        } >> "$rc_file"
        log_info "Added source line to $rc_file"
    else
        log_warn "cd_project already configured in $rc_file"
    fi
}

main() {
    echo "========================================"
    echo "     cd_project Installation"
    echo "========================================"
    echo

    # Check for Go
    if ! command -v go &>/dev/null; then
        log_error "Go is required but not installed."
        echo "       Install Go from: https://go.dev/dl/"
        exit 1
    fi
    log_info "Go found: $(go version)"

    check_fd
    echo

    # Prompt for project root
    log_step "Configure project directories"
    echo "       Enter the root directory(ies) where your Git projects are located."
    echo "       Use colon (:) to separate multiple directories."
    read -rp "Project root directories [$HOME/code]: " PROJECT_ROOTS
    PROJECT_ROOTS="${PROJECT_ROOTS:-$HOME/code}"
    echo

    # Prompt for shell selection
    log_step "Select shell integration"
    echo "  1) Bash only"
    echo "  2) Zsh only"
    echo "  3) Both"
    read -rp "Choice [3]: " SHELL_CHOICE
    SHELL_CHOICE="${SHELL_CHOICE:-3}"
    echo

    # Build binary
    log_step "Building cd_project..."
    cd "$PROJECT_ROOT"
    go build -ldflags "-X main.Version=$(git describe --tags --always --dirty 2>/dev/null || echo 'dev')" \
        -o "$BINARY_NAME" ./cmd/cd_project
    log_info "Build successful"

    # Install binary
    log_step "Installing binary..."
    mkdir -p "$INSTALL_DIR"
    cp "$BINARY_NAME" "$INSTALL_DIR/"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    log_info "Installed binary to $INSTALL_DIR/$BINARY_NAME"

    # Check if INSTALL_DIR is in PATH
    if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
        log_warn "$INSTALL_DIR is not in your PATH"
        echo "       Add this to your shell RC file:"
        echo "       export PATH=\"\$PATH:$INSTALL_DIR\""
    fi

    # Install shell integration
    log_step "Installing shell integration..."

    if [[ "$SHELL_CHOICE" == "1" || "$SHELL_CHOICE" == "3" ]]; then
        if [ -f "$HOME/.bashrc" ]; then
            cp "$PROJECT_ROOT/scripts/shell/bash_completion.sh" "$HOME/.cd_project.bash"
            add_source_line "$HOME/.bashrc" "$HOME/.cd_project.bash"

            # Add CD_PROJECT_ROOT to bashrc if not present
            if ! grep -q "CD_PROJECT_ROOT" "$HOME/.bashrc"; then
                echo "export CD_PROJECT_ROOT=\"$PROJECT_ROOTS\"" >> "$HOME/.bashrc"
                log_info "Added CD_PROJECT_ROOT to ~/.bashrc"
            fi
        else
            log_warn "~/.bashrc not found, skipping Bash integration"
        fi
    fi

    if [[ "$SHELL_CHOICE" == "2" || "$SHELL_CHOICE" == "3" ]]; then
        if [ -f "$HOME/.zshrc" ]; then
            cp "$PROJECT_ROOT/scripts/shell/zsh_completion.sh" "$HOME/.cd_project.zsh"
            add_source_line "$HOME/.zshrc" "$HOME/.cd_project.zsh"

            # Add CD_PROJECT_ROOT to zshrc if not present
            if ! grep -q "CD_PROJECT_ROOT" "$HOME/.zshrc"; then
                echo "export CD_PROJECT_ROOT=\"$PROJECT_ROOTS\"" >> "$HOME/.zshrc"
                log_info "Added CD_PROJECT_ROOT to ~/.zshrc"
            fi
        else
            log_warn "~/.zshrc not found, skipping Zsh integration"
        fi
    fi

    # Run initial refresh
    echo
    log_step "Running initial project scan..."
    export CD_PROJECT_ROOT="$PROJECT_ROOTS"
    "$INSTALL_DIR/$BINARY_NAME" --refresh

    echo
    echo "========================================"
    echo "     Installation Complete!"
    echo "========================================"
    echo
    echo "Restart your shell or run:"
    echo "  source ~/.bashrc   # for Bash"
    echo "  source ~/.zshrc    # for Zsh"
    echo
    echo "Usage:"
    echo "  cdp <project>      # cd to project directory"
    echo "  cdp <TAB>          # tab completion"
    echo "  cdp-refresh        # rescan projects"
}

main "$@"

