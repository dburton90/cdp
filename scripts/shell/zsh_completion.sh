# cd_project shell integration for Zsh
# Source this file in your .zshrc

# Main cdp function - wrapper around cd
cdp() {
    if [[ -z "$1" ]]; then
        echo "Usage: cdp <project-name>" >&2
        return 1
    fi

    local target
    target=$(cd_project --path "$1" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$target" ]]; then
        cd "$target" || return 1
    else
        echo "$target" >&2
        return 1
    fi
}

# Zsh completion function
_cdp() {
    local -a completions
    completions=(${(f)"$(cd_project --completion "${words[2]:-}" 2>/dev/null)"})
    _describe 'project' completions
}

# Register completion
compdef _cdp cdp

# Refresh helper alias
alias cdp-refresh='cd_project --refresh'

