package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dbarton/cd_project/internal/cache"
	"github.com/dbarton/cd_project/internal/matcher"
	"github.com/dbarton/cd_project/internal/project"
	"github.com/dbarton/cd_project/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	pathFlag       string
	refreshFlag    bool
	completionFlag string
	versionFlag    bool
	Version        = "dev"
)

var rootCmd = &cobra.Command{
	Use:   "cd_project [name]",
	Short: "Fast project directory navigator",
	Long:  `cd_project - A fast project directory navigator that helps you jump to your git repositories quickly.`,
	Args:  cobra.ArbitraryArgs,
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) != 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}

		c := cache.NewFileCache()
		projects, err := c.Load()
		if err != nil {
			return nil, cobra.ShellCompDirectiveError
		}

		m := matcher.NewMatcher()
		matches := m.Match(toComplete, projects)

		var names []string
		for _, p := range matches {
			name := p.ResolvedName
			if name == "" {
				name = p.Name
			}
			names = append(names, name)
		}
		return names, cobra.ShellCompDirectiveNoFileComp
	},
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Skip auto-refresh for commands that don't need it or will do it themselves
		if cmd.Name() == "init" || cmd.Name() == "completion" || cmd.Name() == "help" || cmd.Name() == "refresh" || refreshFlag {
			return
		}
		
		c := cache.NewFileCache()
		if !c.Exists() {
			fmt.Fprintf(os.Stderr, "Cache missing, initializing...\n")
			PerformRefresh()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			fmt.Printf("cd_project %s\n", Version)
			return
		}

		if completionFlag != "" {
			handleCompletion(completionFlag)
			return
		}

		if pathFlag != "" {
			handlePath(pathFlag)
			return
		}

		if refreshFlag {
			PerformRefresh()
			return
		}

		// If no flags and no args, show help
		if len(args) == 0 {
			cmd.Help()
		} else {
			// Legacy behavior: first arg as name for path
			handlePath(args[0])
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&pathFlag, "path", "", "Return absolute path for project name")
	rootCmd.Flags().BoolVar(&refreshFlag, "refresh", false, "Rescan directories and rebuild cache")
	rootCmd.Flags().StringVar(&completionFlag, "completion", "", "Return matching project names for query (substring match)")
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Print version")
	
	// Ensure version info is consistent with main.go if needed
}

func handleCompletion(prefix string) {
	c := cache.NewFileCache()
	projects, err := c.Load()
	if err != nil {
		os.Exit(1)
	}

	if len(projects) == 0 {
		os.Exit(0)
	}

	m := matcher.NewMatcher()
	matches := m.Match(prefix, projects)

	seen := make(map[string]bool)
	for _, p := range matches {
		name := p.ResolvedName
		if name == "" {
			name = p.Name
		}
		if !seen[name] {
			fmt.Println(name)
			seen[name] = true
		}
	}
}

func handlePath(name string) {
	c := cache.NewFileCache()
	projects, err := c.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading cache: %v\n", err)
		os.Exit(1)
	}

	if len(projects) == 0 {
		fmt.Fprintf(os.Stderr, "No projects cached.\n")
		os.Exit(1)
	}

	m := matcher.NewMatcher()
	project, err := m.FindExact(name, projects)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Println(project.Path)
}

func PerformRefresh() {
	roots, err := getRoots()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Scanning %d root(s) using %s scanner...\n",
		len(roots), scanner.ScannerType())

	s := scanner.NewScanner()
	projects, err := s.Scan(roots)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning: %v\n", err)
		os.Exit(1)
	}

	// Resolve names
	r := project.NewResolver()
	projects = r.Resolve(projects)

	c := cache.NewFileCache()
	if err := c.Save(projects); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving cache: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Found %d projects\n", len(projects))
	fmt.Fprintf(os.Stderr, "Cache saved to: %s\n", c.Path())
}

func getRoots() ([]string, error) {
	rootEnv := os.Getenv("CD_PROJECT_ROOT")
	if rootEnv == "" {
		return nil, fmt.Errorf(
			"CD_PROJECT_ROOT not set.\n" +
				"Add to your shell RC:\n" +
				"  export CD_PROJECT_ROOT=\"$HOME/code\"\n" +
				"Use colon-separated paths for multiple roots:\n" +
				"  export CD_PROJECT_ROOT=\"$HOME/code:$HOME/work\"")
	}

	roots := strings.Split(rootEnv, ":")
	var validRoots []string

	for _, root := range roots {
		root = strings.TrimSpace(root)
		root = os.ExpandEnv(root)
		if info, err := os.Stat(root); err == nil && info.IsDir() {
			validRoots = append(validRoots, root)
		}
	}

	if len(validRoots) == 0 {
		return nil, fmt.Errorf("no valid directories in CD_PROJECT_ROOT: %s", rootEnv)
	}

	return validRoots, nil
}
