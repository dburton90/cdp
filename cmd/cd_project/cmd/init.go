package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Output shell integration scripts",
	Long:  `Output the shell script required to initialize cd_project (cdp function and completion).`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		shell := args[0]
		switch shell {
		case "bash":
			fmt.Print(bashTemplate)
		case "zsh":
			fmt.Print(zshTemplate)
		default:
			fmt.Fprintf(os.Stderr, "Unsupported shell: %s. Supported: bash, zsh\n", shell)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

const bashTemplate = `
cdp() {
    if [ "$#" -eq 0 ]; then
        cd_project --help
        return
    fi
    local dest
    dest=$(cd_project "$1")
    if [ -n "$dest" ]; then
        cd "$dest"
    fi
}

_cdp_completion() {
    local cur
    cur="${COMP_WORDS[COMP_CWORD]}"
    # Use cobra's built-in completion mechanism
    local matches
    matches=$(cd_project __complete "$cur" 2>/dev/null | cut -f1)
    COMPREPLY=( $(compgen -W "${matches}" -- "$cur") )
}
complete -F _cdp_completion cdp
`

const zshTemplate = `
cdp() {
    if [ "$#" -eq 0 ]; then
        cd_project --help
        return
    fi
    local dest
    dest=$(cd_project "$1")
    if [ -n "$dest" ]; then
        cd "$dest"
    fi
}

_cdp_completion() {
    local -a matches
    matches=(${(f)"$(cd_project __complete "$PREFIX" 2>/dev/null | cut -f1)"})
    _describe 'projects' matches
}
compdef _cdp_completion cdp
`
