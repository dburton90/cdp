# cd_project shell integration for Bash
# Source this file in your .bashrc

# Main cdp function - wrapper around cd
cdp() {
    if [ -z "$1" ]; then
        echo "Usage: cdp <project-name>" >&2
        return 1
    fi

    local target
    target=$(cd_project --path "$1" 2>&1)
    local exit_code=$?

    if [ $exit_code -eq 0 ] && [ -n "$target" ]; then
        cd "$target" || return 1
    else
        echo "$target" >&2
        return 1
    fi
}

# Bash completion function
_cdp_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local completions
    completions=$(cd_project --completion "$cur" 2>/dev/null)
    # shellcheck disable=SC2207
    COMPREPLY=($(compgen -W "$completions" -- "$cur"))
}

# Register completion
complete -F _cdp_completions cdp

# Refresh helper alias
alias cdp-refresh='cd_project --refresh'

